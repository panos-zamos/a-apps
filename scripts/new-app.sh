#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check arguments
if [ "$#" -ne 2 ]; then
    echo -e "${RED}Usage: $0 <app-name> <port>${NC}"
    echo -e "${YELLOW}Example: $0 todo-list 3001${NC}"
    exit 1
fi

APP_NAME=$1
APP_PORT=$2

# Validate app name (lowercase, hyphens only)
if ! [[ "$APP_NAME" =~ ^[a-z][a-z0-9-]*$ ]]; then
    echo -e "${RED}Error: App name must start with a letter and contain only lowercase letters, numbers, and hyphens${NC}"
    exit 1
fi

# Validate port number
if ! [[ "$APP_PORT" =~ ^[0-9]+$ ]] || [ "$APP_PORT" -lt 1024 ] || [ "$APP_PORT" -gt 65535 ]; then
    echo -e "${RED}Error: Port must be a number between 1024 and 65535${NC}"
    exit 1
fi

# Check if app already exists
if [ -d "apps/$APP_NAME" ]; then
    echo -e "${RED}Error: App 'apps/$APP_NAME' already exists${NC}"
    exit 1
fi

echo -e "${GREEN}Creating new app: $APP_NAME on port $APP_PORT${NC}"

# Create app directory
mkdir -p "apps/$APP_NAME"
mkdir -p "apps/$APP_NAME/handlers"
mkdir -p "apps/$APP_NAME/data"

# Copy template files and replace placeholders
for template_file in templates/app-template/*.tmpl; do
    if [ -f "$template_file" ]; then
        filename=$(basename "$template_file" .tmpl)
        target="apps/$APP_NAME/$filename"
        
        # Replace placeholders
        sed -e "s/{{APP_NAME}}/$APP_NAME/g" \
            -e "s/{{APP_PORT}}/$APP_PORT/g" \
            "$template_file" > "$target"
        
        echo -e "${YELLOW}  Created: $target${NC}"
    fi
done

# Copy handler templates
for template_file in templates/app-template/handlers/*.tmpl; do
    if [ -f "$template_file" ]; then
        filename=$(basename "$template_file" .tmpl)
        target="apps/$APP_NAME/handlers/$filename"
        
        # Replace placeholders
        sed -e "s/{{APP_NAME}}/$APP_NAME/g" \
            -e "s/{{APP_PORT}}/$APP_PORT/g" \
            "$template_file" > "$target"
        
        echo -e "${YELLOW}  Created: $target${NC}"
    fi
done

# Initialize Go module
cd "apps/$APP_NAME"
go mod tidy

echo -e "${GREEN}✓ App created successfully!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. cd apps/$APP_NAME"
echo -e "  2. Edit config.yaml to add users"
echo -e "  3. go run main.go"
echo -e "  4. Visit http://localhost:$APP_PORT"
echo ""
echo -e "${YELLOW}Development:${NC}"
echo -e "  make dev-$APP_NAME  # Run with hot-reload"
echo ""
echo -e "${YELLOW}Default credentials:${NC}"
echo -e "  Username: panos"
echo -e "  Password: demo123"
