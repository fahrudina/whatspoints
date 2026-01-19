#!/bin/bash

# Deployment script for WhatsPoints
# Usage: ./deploy.sh [server-user@server-ip] [deploy-path]

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVER=${1:-""}
DEPLOY_PATH=${2:-"/app"}
BINARY_NAME="whatspoints"

echo -e "${GREEN}=== WhatsPoints Deployment Script ===${NC}"

# Check if server is provided
if [ -z "$SERVER" ]; then
    echo -e "${RED}Error: Server not specified${NC}"
    echo "Usage: ./deploy.sh user@server-ip [deploy-path]"
    echo "Example: ./deploy.sh root@123.45.67.89 /app"
    exit 1
fi

echo -e "${YELLOW}Building application...${NC}"
go build -o $BINARY_NAME

if [ ! -f "$BINARY_NAME" ]; then
    echo -e "${RED}Error: Build failed, binary not found${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Build successful${NC}"

# Check if web directory exists
if [ ! -d "web" ]; then
    echo -e "${RED}Error: web/ directory not found${NC}"
    exit 1
fi

echo -e "${YELLOW}Creating deployment directory on server...${NC}"
ssh $SERVER "mkdir -p $DEPLOY_PATH"

echo -e "${YELLOW}Uploading files to server...${NC}"

# Upload binary
echo "  - Uploading binary..."
scp $BINARY_NAME $SERVER:$DEPLOY_PATH/

# Upload web directory
echo "  - Uploading web files..."
scp -r web/ $SERVER:$DEPLOY_PATH/

# Upload .env if it exists
if [ -f ".env" ]; then
    echo "  - Uploading .env file..."
    scp .env $SERVER:$DEPLOY_PATH/
else
    echo -e "${YELLOW}  ! Warning: .env file not found, skipping...${NC}"
fi

echo -e "${GREEN}✓ Files uploaded${NC}"

echo -e "${YELLOW}Setting permissions...${NC}"
ssh $SERVER "chmod +x $DEPLOY_PATH/$BINARY_NAME"
ssh $SERVER "chmod -R 644 $DEPLOY_PATH/web/*.html"

echo -e "${GREEN}✓ Permissions set${NC}"

echo -e "${YELLOW}Checking deployment...${NC}"
ssh $SERVER "ls -la $DEPLOY_PATH/"

echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo "Next steps:"
echo "1. SSH to server: ssh $SERVER"
echo "2. Navigate to: cd $DEPLOY_PATH"
echo "3. Add a WhatsApp sender:"
echo "   - QR code: ./$BINARY_NAME -add-sender"
echo "   - Pairing code: ./$BINARY_NAME -add-sender-code +PHONE"
echo "4. Start the server:"
echo "   - Direct: ./$BINARY_NAME"
echo "   - Or set up systemd service (see DEPLOYMENT.md)"
echo ""
echo "To restart if using systemd:"
echo "   ssh $SERVER 'sudo systemctl restart whatspoints'"
