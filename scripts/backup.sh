#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Backing up SQLite databases...${NC}"

# Create backup directory with timestamp
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Find all SQLite databases in apps
DB_COUNT=0
for db_file in apps/*/data/*.db; do
    if [ -f "$db_file" ]; then
        # Extract app name
        app_name=$(echo "$db_file" | cut -d'/' -f2)
        
        # Copy database
        cp "$db_file" "$BACKUP_DIR/${app_name}.db"
        echo -e "${YELLOW}  Backed up: $db_file${NC}"
        
        DB_COUNT=$((DB_COUNT + 1))
    fi
done

if [ $DB_COUNT -eq 0 ]; then
    echo -e "${YELLOW}No databases found to backup${NC}"
    rmdir "$BACKUP_DIR"
    exit 0
fi

echo -e "${GREEN}✓ Backed up $DB_COUNT database(s) to $BACKUP_DIR${NC}"

# Optional: Clean up old backups (keep last 30 days)
find backups/ -type d -mtime +30 -exec rm -rf {} + 2>/dev/null || true

echo -e "${YELLOW}Tip: Copy $BACKUP_DIR to safe storage (cloud, external drive)${NC}"
