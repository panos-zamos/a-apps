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

# Check configuration
if [ -z "$SERVER_HOST" ]; then
    echo -e "${RED}Error: DEPLOY_HOST not set${NC}"
    echo -e "${YELLOW}Set it with: export DEPLOY_HOST=your-server.com${NC}"
    echo -e "${YELLOW}Or edit this script to set defaults${NC}"
    exit 1
fi

echo -e "${GREEN}Deploying to $SERVER_USER@$SERVER_HOST:$SERVER_PATH${NC}"

# Backup databases before deploying
echo -e "${YELLOW}Creating backup...${NC}"
./scripts/backup.sh

# Build locally (optional, could also build on server)
echo -e "${YELLOW}Building Docker images...${NC}"
docker-compose -f deploy/docker-compose.yml build

# Save images to tar (alternative: use registry)
# echo -e "${YELLOW}Saving images...${NC}"
# docker save ... (if using tar transfer method)

# Sync code to server
echo -e "${YELLOW}Syncing code to server...${NC}"
rsync -avz --exclude='node_modules' --exclude='data' --exclude='tmp' \
    --exclude='.git' --exclude='backups' \
    ./ $SERVER_USER@$SERVER_HOST:$SERVER_PATH/

# Execute deployment on server
echo -e "${YELLOW}Deploying on server...${NC}"
ssh $SERVER_USER@$SERVER_HOST << EOF
    cd $SERVER_PATH
    
    # Pull latest images (if using registry) or build
    docker-compose -f deploy/docker-compose.yml build
    
    # Stop old containers
    docker-compose -f deploy/docker-compose.yml down
    
    # Start new containers
    docker-compose -f deploy/docker-compose.yml up -d
    
    # Show status
    docker-compose -f deploy/docker-compose.yml ps
EOF

echo -e "${GREEN}✓ Deployment complete!${NC}"
echo -e "${YELLOW}Check logs with: ssh $SERVER_USER@$SERVER_HOST 'cd $SERVER_PATH && docker-compose -f deploy/docker-compose.yml logs -f'${NC}"
