#!/usr/bin/env bash
set -euo pipefail

# Simple, droplet-first deployment.
#
# This script is intentionally small: deployment is executed on the droplet by
# pulling pre-built images and restarting containers.
#
# See docs:
# - docs/guides/deployment.md
# - docs/guides/publishing.md

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

exec "$SCRIPT_DIR/deploy-local.sh" "$@"
