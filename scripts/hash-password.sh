#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if password provided
if [ "$#" -ne 1 ]; then
    echo -e "${YELLOW}Usage: $0 <password>${NC}"
    echo ""
    echo "This script generates a bcrypt hash for use in config.yaml"
    echo ""
    echo "Example:"
    echo "  $0 mypassword"
    exit 1
fi

PASSWORD=$1

# Check if Python is available with bcrypt
if command -v python3 &> /dev/null; then
    # Try to use Python with bcrypt
    HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw('$PASSWORD'.encode(), bcrypt.gensalt()).decode())" 2>/dev/null || echo "")
    
    if [ -n "$HASH" ]; then
        echo -e "${GREEN}Password hash generated:${NC}"
        echo ""
        echo "$HASH"
        echo ""
        echo -e "${YELLOW}Add this to your config.yaml:${NC}"
        echo "  - username: yourname"
        echo "    password_hash: \"$HASH\""
        exit 0
    fi
fi

# Fallback: suggest Go program
echo -e "${YELLOW}Python with bcrypt not found. Creating a Go helper program...${NC}"

cat > /tmp/hash_password.go <<'EOF'
package main

import (
	"fmt"
	"os"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run hash_password.go <password>")
		os.Exit(1)
	}
	
	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(hash))
}
EOF

# Run the Go program
cd /tmp
go mod init hashpassword 2>/dev/null || true
go get golang.org/x/crypto/bcrypt 2>/dev/null || true
HASH=$(go run hash_password.go "$PASSWORD")

echo -e "${GREEN}Password hash generated:${NC}"
echo ""
echo "$HASH"
echo ""
echo -e "${YELLOW}Add this to your config.yaml:${NC}"
echo "  - username: yourname"
echo "    password_hash: \"$HASH\""
