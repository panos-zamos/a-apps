#!/usr/bin/env bash
set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOY_DIR="$REPO_ROOT/deploy"
ENV_FILE="$DEPLOY_DIR/.env"

require_cmd() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || {
    echo -e "${RED}Error: required command not found: $cmd${NC}" 1>&2
    exit 1
  }
}

require_var() {
  local name="$1"
  if [ -z "${!name:-}" ]; then
    echo -e "${RED}Error: missing required variable: $name${NC}" 1>&2
    exit 1
  fi
}

require_cmd docker

docker compose version >/dev/null 2>&1 || {
  echo -e "${RED}Error: docker compose plugin not available (try: docker compose version)${NC}" 1>&2
  exit 1
}

if [ ! -f "$ENV_FILE" ]; then
  echo -e "${RED}Error: missing $ENV_FILE${NC}" 1>&2
  echo -e "${YELLOW}Create it from: cp $DEPLOY_DIR/.env.example $ENV_FILE${NC}" 1>&2
  exit 1
fi

# Load env to validate required variables (compose will also read it implicitly from deploy/)
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

require_var DOMAIN
require_var REGISTRY
require_var IMAGE_TAG
require_var JWT_SECRET

echo -e "${GREEN}Deploying on this machine using $ENV_FILE${NC}"

cd "$DEPLOY_DIR"

echo -e "${YELLOW}Pulling images...${NC}"
docker compose pull

echo -e "${YELLOW}Starting containers...${NC}"
docker compose up -d --remove-orphans

echo -e "${YELLOW}Status:${NC}"
docker compose ps

echo -e "${GREEN}✓ Done${NC}"
echo -e "${YELLOW}Caddy logs: docker logs -f a-apps-caddy${NC}"
