#!/usr/bin/env bash
set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/deploy.sh

Modes:
  - Local mode (run on the droplet): if DEPLOY_HOST is NOT set
  - Remote mode (run from your laptop): if DEPLOY_HOST is set

Environment / config:
  - The script can read variables from ENV_FILE (default: deploy/.env) when running locally.
  - In remote mode, the droplet must already have deploy/.env (created manually).

Required vars (deployment):
  DOMAIN, REGISTRY, JWT_SECRET

Required vars (private registries, unless SKIP_REGISTRY_LOGIN=1):
  REGISTRY_USER, REGISTRY_TOKEN

Remote mode additionally requires:
  DEPLOY_HOST (and optionally DEPLOY_USER, DEPLOY_PATH)

Optional vars:
  IMAGE_TAG (default: latest)
  REGISTRY_HOST (derived from REGISTRY if empty)
  SKIP_REGISTRY_LOGIN=1
USAGE
}

require_var() {
  local name="$1"
  if [ -z "${!name:-}" ]; then
    echo -e "${RED}Error: $name is not set${NC}" 1>&2
    exit 1
  fi
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo -e "${RED}Error: required command not found: $cmd${NC}" 1>&2
    exit 1
  fi
}

derive_registry_host() {
  if [ -z "${REGISTRY_HOST:-}" ]; then
    REGISTRY_HOST="${REGISTRY%%/*}"
  fi
}

load_env_file_if_present() {
  local env_file="$1"
  if [ -f "$env_file" ]; then
    # shellcheck disable=SC1090
    set -a
    source "$env_file"
    set +a
  fi
}

sanity_check_common() {
  IMAGE_TAG="${IMAGE_TAG:-latest}"

  require_var DOMAIN
  require_var REGISTRY
  require_var JWT_SECRET

  derive_registry_host

  if [ "${SKIP_REGISTRY_LOGIN:-0}" != "1" ]; then
    require_var REGISTRY_USER
    require_var REGISTRY_TOKEN
  fi
}

registry_login_if_needed() {
  if [ "${SKIP_REGISTRY_LOGIN:-0}" = "1" ]; then
    echo -e "${YELLOW}Skipping registry login (SKIP_REGISTRY_LOGIN=1)${NC}"
    return 0
  fi

  echo -e "${YELLOW}Logging into registry (${REGISTRY_HOST})...${NC}"
  echo "$REGISTRY_TOKEN" | docker login "$REGISTRY_HOST" -u "$REGISTRY_USER" --password-stdin
}

compose_cmd() {
  # We always provide --env-file to avoid relying on the caller's current directory.
  local env_file="$1"
  echo "docker compose --env-file $env_file -f deploy/docker-compose.yml"
}

run_local() {
  local env_file="$1"

  cd "$REPO_ROOT"

  require_cmd docker
  docker compose version >/dev/null

  if [ ! -f "$env_file" ]; then
    echo -e "${RED}Error: env file not found: $env_file${NC}" 1>&2
    echo -e "${YELLOW}Create it from: cp deploy/.env.example $env_file${NC}" 1>&2
    exit 1
  fi

  load_env_file_if_present "$env_file"
  sanity_check_common

  echo -e "${GREEN}Deploying locally in $REPO_ROOT${NC}"

  # Backup (best-effort)
  if [ -x "$REPO_ROOT/scripts/backup.sh" ]; then
    echo -e "${YELLOW}Creating local backup...${NC}"
    "$REPO_ROOT/scripts/backup.sh" || true
  fi

  registry_login_if_needed

  local COMPOSE
  COMPOSE="$(compose_cmd "$env_file")"

  echo -e "${YELLOW}Pulling images...${NC}"
  $COMPOSE pull

  echo -e "${YELLOW}Starting containers...${NC}"
  $COMPOSE up -d --remove-orphans

  echo -e "${YELLOW}Status:${NC}"
  $COMPOSE ps

  echo -e "${GREEN}✓ Done${NC}"
  echo -e "${YELLOW}Caddy logs: docker logs -f a-apps-caddy${NC}"
}

run_remote() {
  local env_file="$1"

  DEPLOY_USER="${DEPLOY_USER:-root}"
  DEPLOY_PATH="${DEPLOY_PATH:-/opt/a-apps}"

  require_var DEPLOY_HOST

  echo -e "${YELLOW}Reading deploy configuration from server env file (${env_file})...${NC}"
  mapfile -t REMOTE_ENV < <(
    ssh "$DEPLOY_USER@$DEPLOY_HOST" \
      "cd '$DEPLOY_PATH' && test -f '$env_file' && set -a && . '$env_file' && set +a && echo \"\${DOMAIN}\" && echo \"\${REGISTRY}\" && echo \"\${IMAGE_TAG:-latest}\" && echo \"\${SKIP_REGISTRY_LOGIN:-0}\""
  )

  local REMOTE_DOMAIN="${REMOTE_ENV[0]:-}"
  local REMOTE_REGISTRY="${REMOTE_ENV[1]:-}"
  local REMOTE_IMAGE_TAG="${REMOTE_ENV[2]:-latest}"
  local REMOTE_SKIP_REGISTRY_LOGIN="${REMOTE_ENV[3]:-0}"

  if [ -z "${SKIP_REGISTRY_LOGIN:-}" ]; then
    SKIP_REGISTRY_LOGIN="$REMOTE_SKIP_REGISTRY_LOGIN"
  fi

  if [ -z "$REMOTE_DOMAIN" ] || [ -z "$REMOTE_REGISTRY" ]; then
    echo -e "${RED}Error: $env_file on the server is missing DOMAIN and/or REGISTRY${NC}" 1>&2
    echo -e "${YELLOW}Fix it using deploy/.env.example (see docs/deployment.md)${NC}" 1>&2
    exit 1
  fi

  # Use server values as source-of-truth to avoid tag/registry mismatches.
  if [ -n "${REGISTRY:-}" ] && [ "$REGISTRY" != "$REMOTE_REGISTRY" ]; then
    echo -e "${RED}Error: REGISTRY mismatch${NC}" 1>&2
    echo -e "${YELLOW}Local REGISTRY=$REGISTRY${NC}" 1>&2
    echo -e "${YELLOW}Server REGISTRY=$REMOTE_REGISTRY (from $env_file)${NC}" 1>&2
    exit 1
  fi
  REGISTRY="$REMOTE_REGISTRY"

  if [ -n "${IMAGE_TAG:-}" ] && [ "$IMAGE_TAG" != "$REMOTE_IMAGE_TAG" ]; then
    echo -e "${RED}Error: IMAGE_TAG mismatch${NC}" 1>&2
    echo -e "${YELLOW}Local IMAGE_TAG=$IMAGE_TAG${NC}" 1>&2
    echo -e "${YELLOW}Server IMAGE_TAG=$REMOTE_IMAGE_TAG (from $env_file)${NC}" 1>&2
    exit 1
  fi
  IMAGE_TAG="$REMOTE_IMAGE_TAG"

  derive_registry_host

  if [ "${SKIP_REGISTRY_LOGIN:-0}" != "1" ]; then
    require_var REGISTRY_USER
    require_var REGISTRY_TOKEN
  fi

  echo -e "${GREEN}Remote deploy to ${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}${NC}"

  cd "$REPO_ROOT"

  require_cmd docker
  docker compose version >/dev/null

  registry_login_if_needed

  echo -e "${YELLOW}Building images locally...${NC}"
  docker build -t "$REGISTRY/a-apps-todo-list:$IMAGE_TAG" -f apps/todo-list/Dockerfile .
  docker build -t "$REGISTRY/a-apps-projects:$IMAGE_TAG" -f apps/projects/Dockerfile .

  echo -e "${YELLOW}Pushing images...${NC}"
  docker push "$REGISTRY/a-apps-todo-list:$IMAGE_TAG"
  docker push "$REGISTRY/a-apps-projects:$IMAGE_TAG"

  echo -e "${YELLOW}Syncing code to server (excluding deploy/.env and data)...${NC}"
  rsync -avz \
    --exclude='node_modules' --exclude='data' --exclude='tmp' \
    --exclude='.git' --exclude='backups' \
    --exclude='deploy/.env' \
    ./ "$DEPLOY_USER@$DEPLOY_HOST:$DEPLOY_PATH/"

  echo -e "${YELLOW}Deploying on server...${NC}"
  ssh "$DEPLOY_USER@$DEPLOY_HOST" << EOF
set -euo pipefail
cd "$DEPLOY_PATH"

if [ ! -f "$env_file" ]; then
  echo "ERROR: Missing $env_file on server. Create it from deploy/.env.example (see docs/deployment.md)." 1>&2
  exit 1
fi

# Load env file to sanity check values that compose depends on.
set -a
. "$env_file"
set +a

# Sanity checks (server-side)
if [ -z "\${DOMAIN:-}" ]; then
  echo 'ERROR: DOMAIN missing in deploy/.env' 1>&2
  exit 1
fi

if [ -z "\${REGISTRY:-}" ]; then
  echo 'ERROR: REGISTRY missing in deploy/.env' 1>&2
  exit 1
fi

if [ -z "\${JWT_SECRET:-}" ]; then
  echo 'ERROR: JWT_SECRET missing in deploy/.env' 1>&2
  exit 1
fi

if [ -z "\${REGISTRY_HOST:-}" ]; then
  REGISTRY_HOST="\${REGISTRY%%/*}"
fi

if [ "\${SKIP_REGISTRY_LOGIN:-0}" != "1" ]; then
  test -n "\${REGISTRY_USER:-}" || (echo 'ERROR: REGISTRY_USER missing in deploy/.env' 1>&2; exit 1)
  test -n "\${REGISTRY_TOKEN:-}" || (echo 'ERROR: REGISTRY_TOKEN missing in deploy/.env' 1>&2; exit 1)
  echo "\$REGISTRY_TOKEN" | docker login "\$REGISTRY_HOST" -u "\$REGISTRY_USER" --password-stdin
else
  echo "Skipping registry login (SKIP_REGISTRY_LOGIN=1)"
fi

# Backup (best-effort)
if [ -x scripts/backup.sh ]; then
  ./scripts/backup.sh || true
fi

COMPOSE="docker compose --env-file $env_file -f deploy/docker-compose.yml"
\$COMPOSE pull
\$COMPOSE up -d --remove-orphans
\$COMPOSE ps
EOF

  echo -e "${GREEN}✓ Deployment complete!${NC}"
  echo -e "${YELLOW}Server logs: ssh $DEPLOY_USER@$DEPLOY_HOST 'docker logs -f a-apps-caddy'${NC}"
}

main() {
  if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
    usage
    exit 0
  fi

  local env_file
  env_file="${ENV_FILE:-deploy/.env}"

  # Local mode if DEPLOY_HOST not set
  if [ -z "${DEPLOY_HOST:-}" ]; then
    run_local "$env_file"
  else
    run_remote "$env_file"
  fi
}

main "$@"
