#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration - edit these for your setup
SERVER_USER=${DEPLOY_USER:-"root"}
SERVER_HOST=${DEPLOY_HOST:-""}
SERVER_PATH=${DEPLOY_PATH:-"/opt/a-apps"}
REGISTRY=${REGISTRY:-""}
IMAGE_TAG=${IMAGE_TAG:-"latest"}

# Check configuration
if [ -z "$SERVER_HOST" ]; then
    echo -e "${RED}Error: DEPLOY_HOST not set${NC}"
    echo -e "${YELLOW}Set it with: export DEPLOY_HOST=your-server.com${NC}"
    echo -e "${YELLOW}Or edit this script to set defaults${NC}"
    exit 1
fi

if [ -z "$REGISTRY" ]; then
    echo -e "${RED}Error: REGISTRY not set${NC}"
    echo -e "${YELLOW}Set it with: export REGISTRY=registry.digitalocean.com/your-registry${NC}"
    exit 1
fi

echo -e "${GREEN}Deploying to $SERVER_USER@$SERVER_HOST:$SERVER_PATH${NC}"

# Backup databases before deploying
echo -e "${YELLOW}Creating backup...${NC}"
./scripts/backup.sh

# Build and push images to registry
echo -e "${YELLOW}Building images locally...${NC}"
docker compose -f deploy/docker-compose.yml build

echo -e "${YELLOW}Pushing images to registry...${NC}"
docker compose -f deploy/docker-compose.yml push

# Sync code to server
echo -e "${YELLOW}Syncing code to server...${NC}"
rsync -avz --exclude='node_modules' --exclude='data' --exclude='tmp' \
    --exclude='.git' --exclude='backups' \
    ./ $SERVER_USER@$SERVER_HOST:$SERVER_PATH/

# Execute deployment on server
echo -e "${YELLOW}Deploying on server...${NC}"
ssh $SERVER_USER@$SERVER_HOST << EOF
    cd $SERVER_PATH
    
    # Pull latest images
    docker compose -f deploy/docker-compose.yml pull
    
    # Stop old containers
    docker compose -f deploy/docker-compose.yml down
    
    # Start new containers
    docker compose -f deploy/docker-compose.yml up -d
    
    # Show status
    docker compose -f deploy/docker-compose.yml ps
EOF

# Download database backups from server
echo -e "${YELLOW}Downloading database backups from server...${NC}"
mkdir -p backups/remote
rsync -avz $SERVER_USER@$SERVER_HOST:$SERVER_PATH/apps/*/data/*.db backups/remote/ 2>/dev/null \
    && echo -e "${GREEN}✓ Backups downloaded to backups/remote/${NC}" \
    || echo -e "${YELLOW}No remote databases found to download${NC}"

echo -e "${GREEN}✓ Deployment complete!${NC}"
echo -e "${YELLOW}Check logs with: ssh $SERVER_USER@$SERVER_HOST 'cd $SERVER_PATH && docker compose -f deploy/docker-compose.yml logs -f'${NC}"
