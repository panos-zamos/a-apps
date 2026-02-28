#!/usr/bin/env bash
set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

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

REGISTRY="${REGISTRY:-}"
IMAGE_TAG="${IMAGE_TAG:-}"

require_var REGISTRY
require_var IMAGE_TAG

cd "$REPO_ROOT"

echo -e "${GREEN}Publishing images to $REGISTRY (tag: $IMAGE_TAG)${NC}"
echo -e "${YELLOW}Note: this script assumes you are already logged in (docker login) if the registry is private.${NC}"

echo -e "${YELLOW}Building images...${NC}"
docker build -t "$REGISTRY/a-apps-todo-list:$IMAGE_TAG" -f apps/todo-list/Dockerfile .
docker build -t "$REGISTRY/a-apps-projects:$IMAGE_TAG" -f apps/projects/Dockerfile .

echo -e "${YELLOW}Pushing images...${NC}"
docker push "$REGISTRY/a-apps-todo-list:$IMAGE_TAG"
docker push "$REGISTRY/a-apps-projects:$IMAGE_TAG"

echo -e "${GREEN}✓ Done${NC}"
