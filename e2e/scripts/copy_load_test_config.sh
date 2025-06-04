#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
WHITE='\033[0;37m'
NC='\033[0m' # No Color

# File paths
TEMPLATE_FILE="./e2e/config/e2e_load_test.config.tmpl.yaml"
CONFIG_FILE="./e2e/config/.e2e_load_test.config.yaml"

echo -e "${BLUE}🚀 Setting up E2E Load Test Configuration${NC}"
echo ""

# Check if config file already exists
if [ -f "$CONFIG_FILE" ]; then
    echo -e "${YELLOW}⚠️ Configuration file already exists at $CONFIG_FILE${NC}"
    echo -e " 👀 You can verify the existing config by running: \n     cat ./e2e/config/.e2e_load_test.config.yaml"
    echo ""
    read -p "🤔 Do you want to overwrite it? (y/N): " OVERWRITE
    if [[ ! "$OVERWRITE" =~ ^[Yy]$ ]]; then
        echo -e "ℹ️  Keeping existing configuration file"
        echo -e "${WHITE}💡 If you want to reconfigure, delete the file and run this command again${NC}"
        exit 0
    fi
fi

# Check if template file exists
if [ ! -f "$TEMPLATE_FILE" ]; then
    echo -e "${RED}❌ Error: Template file not found at $TEMPLATE_FILE${NC}"
    exit 1
fi

# Check if yq is installed
if ! command -v yq &>/dev/null; then
    echo -e "${RED}❌ Error: yq is not installed${NC}"
    echo -e "${YELLOW}💡 Please install yq to continue:${NC}"
    echo -e "${WHITE}   • macOS: brew install yq${NC}"
    echo -e "${WHITE}   • Linux: sudo snap install yq${NC}"
    echo -e "${WHITE}   • Or visit: https://github.com/mikefarah/yq#install${NC}"
    exit 1
fi

# Step 1: Prompt for Portal Application ID
echo -e "${BLUE}🔑 Portal Configuration Setup${NC}"
echo ""
echo -e "${BLUE}📝 Step 1: Portal Application ID${NC}"
echo -e "${WHITE}   This is REQUIRED if you're testing against the Grove Portal.${NC}"
echo -e "${WHITE}   If you don't have one, get it at: https://www.portal.grove.city${NC}"
echo ""
read -p "🆔 Enter your Portal Application ID (or press Enter to skip): " PORTAL_APP_ID

# Step 2: Prompt for Portal API Key
echo ""
echo -e "${BLUE}📝 Step 2: Portal API Key${NC}"
echo -e "${WHITE}   This is REQUIRED if your Portal Application ID requires an API key.${NC}"
echo -e "${WHITE}   You can find this in your Grove Portal dashboard: https://www.portal.grove.city${NC}"
echo ""
read -p "🔐 Enter your Portal API Key (or press Enter to skip): " PORTAL_API_KEY

echo ""

# Step 3: Copy the template file (only after prompts are complete)
echo "📁 Copying e2e_load_test.config.tmpl.yaml to .e2e_load_test.config.yaml"
echo "\n 👀 You can verify the new config by running:\n    cat ./e2e/config/.e2e_load_test.config.yaml\n"
cp "$TEMPLATE_FILE" "$CONFIG_FILE"
echo -e "${GREEN}✅ Successfully copied template to config file${NC}"

# Step 4: Update the config file with yq
echo "⚙️  Updating configuration file..."

if [ -n "$PORTAL_APP_ID" ]; then
    yq eval '.e2e_load_test_config.load_test_config.portal_application_id = "'"$PORTAL_APP_ID"'"' -i "$CONFIG_FILE"
    echo -e "${GREEN}✅ Portal Application ID set${NC}"
fi

if [ -n "$PORTAL_API_KEY" ]; then
    yq eval '.e2e_load_test_config.load_test_config.portal_api_key = "'"$PORTAL_API_KEY"'"' -i "$CONFIG_FILE"
    echo -e "${GREEN}✅ Portal API Key set${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Configuration setup complete!${NC}"
echo ""
echo -e "${WHITE}💡 To customize the load test config further, edit: $CONFIG_FILE${NC}"
echo ""
echo -e "${BLUE}🚀 You can now run load tests with:${NC}"
echo -e "${WHITE}   • make load_test${NC}"
echo -e "${WHITE}   • make load_test eth,anvil${NC}"
echo ""
echo -e "${WHITE} For a full list of all available services to run load tests on, see: ./config/service_qos_config.go"
echo ""
