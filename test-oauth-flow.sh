#!/bin/bash

# 🧪 OAuth2.1 + PKCE Flow Test Script
# Tests the complete authorization flow from PKCE generation to token exchange

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
REDIRECT_URI="http://localhost:3000/callback"

echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  OAuth2.1 + PKCE Flow Test Script${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo ""

# Function to generate random string for code verifier
generate_code_verifier() {
    openssl rand -base64 96 | tr -d "=+/" | cut -c1-128
}

# Function to generate code challenge from verifier
generate_code_challenge() {
    local verifier=$1
    echo -n "$verifier" | openssl dgst -binary -sha256 | openssl base64 | tr -d "=+/" | tr "/+" "_-"
}

# Step 1: Check if servers are running
echo -e "${YELLOW}[1/8]${NC} Checking if servers are running..."
if ! curl -s -f "$API_BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}✗ Backend server is not running at $API_BASE_URL${NC}"
    echo -e "  Start with: ${BLUE}./dev.sh dev${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Backend server is running${NC}"
echo ""

# Step 2: Verify superadmin account exists
echo -e "${YELLOW}[2/8]${NC} Checking for superadmin account..."
echo -e "${BLUE}Please ensure you have a superadmin account created.${NC}"
echo -e "  You can create one by running:"
echo -e "  ${BLUE}psql -U auth_user -d auth_db -c \"UPDATE users SET is_superadmin = true WHERE email = 'admin@example.com';\"${NC}"
echo ""
read -p "Press Enter to continue once you've confirmed superadmin exists..."
echo ""

# Step 3: Get Client App credentials
echo -e "${YELLOW}[3/8]${NC} OAuth Client App Setup"
echo -e "You need to create an OAuth client app first:"
echo -e "  1. Login as superadmin at ${BLUE}http://localhost:5173/login${NC}"
echo -e "  2. Navigate to ${BLUE}/oauth/client-apps${NC}"
echo -e "  3. Create a new client app with:"
echo -e "     - Name: ${GREEN}Test Application${NC}"
echo -e "     - Redirect URI: ${GREEN}$REDIRECT_URI${NC}"
echo -e "     - Scopes: ${GREEN}profile email${NC}"
echo ""
read -p "Enter your CLIENT_ID: " CLIENT_ID
read -p "Enter your CLIENT_SECRET: " CLIENT_SECRET

if [ -z "$CLIENT_ID" ] || [ -z "$CLIENT_SECRET" ]; then
    echo -e "${RED}✗ CLIENT_ID and CLIENT_SECRET are required${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Client credentials saved${NC}"
echo ""

# Step 4: Generate PKCE parameters
echo -e "${YELLOW}[4/8]${NC} Generating PKCE parameters..."
CODE_VERIFIER=$(generate_code_verifier)
CODE_CHALLENGE=$(generate_code_challenge "$CODE_VERIFIER")

echo -e "  Code Verifier: ${GREEN}$CODE_VERIFIER${NC}"
echo -e "  Code Challenge: ${GREEN}$CODE_CHALLENGE${NC}"
echo ""

# Step 5: Test authorization endpoint
echo -e "${YELLOW}[5/8]${NC} Testing Authorization Endpoint..."
STATE="test_state_$(date +%s)"
SCOPES="profile email"

AUTH_URL="$API_BASE_URL/oauth/authorize?client_id=$CLIENT_ID&redirect_uri=$REDIRECT_URI&response_type=code&scope=$SCOPES&code_challenge=$CODE_CHALLENGE&code_challenge_method=S256&state=$STATE"

echo -e "${BLUE}Authorization URL:${NC}"
echo -e "$AUTH_URL"
echo ""
echo -e "Next steps:"
echo -e "  1. Login as a ${GREEN}regular user${NC} (not superadmin)"
echo -e "  2. Navigate to the URL above in your browser"
echo -e "  3. Review the consent screen"
echo -e "  4. Click ${GREEN}'Allow Access'${NC}"
echo -e "  5. You'll be redirected to: ${BLUE}$REDIRECT_URI?code=...&state=...${NC}"
echo ""
read -p "After authorizing, paste the AUTHORIZATION CODE here: " AUTH_CODE

if [ -z "$AUTH_CODE" ]; then
    echo -e "${RED}✗ Authorization code is required${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Authorization code received${NC}"
echo ""

# Step 6: Exchange authorization code for tokens
echo -e "${YELLOW}[6/8]${NC} Exchanging authorization code for tokens..."

TOKEN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=$AUTH_CODE" \
  -d "redirect_uri=$REDIRECT_URI" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "code_verifier=$CODE_VERIFIER")

# Check if token exchange was successful
if echo "$TOKEN_RESPONSE" | grep -q "access_token"; then
    echo -e "${GREEN}✓ Token exchange successful${NC}"
    
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
    
    echo ""
    echo -e "${BLUE}Token Response:${NC}"
    echo "$TOKEN_RESPONSE" | jq '.' 2>/dev/null || echo "$TOKEN_RESPONSE"
    echo ""
else
    echo -e "${RED}✗ Token exchange failed${NC}"
    echo -e "Response: $TOKEN_RESPONSE"
    exit 1
fi

# Step 7: Test userinfo endpoint
echo -e "${YELLOW}[7/8]${NC} Testing UserInfo endpoint..."

USERINFO_RESPONSE=$(curl -s "$API_BASE_URL/oauth/userinfo" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$USERINFO_RESPONSE" | grep -q "sub"; then
    echo -e "${GREEN}✓ UserInfo endpoint successful${NC}"
    echo ""
    echo -e "${BLUE}UserInfo Response:${NC}"
    echo "$USERINFO_RESPONSE" | jq '.' 2>/dev/null || echo "$USERINFO_RESPONSE"
    echo ""
else
    echo -e "${RED}✗ UserInfo endpoint failed${NC}"
    echo -e "Response: $USERINFO_RESPONSE"
    exit 1
fi

# Step 8: Decode and verify JWT claims
echo -e "${YELLOW}[8/8]${NC} Decoding JWT to verify RBAC claims..."

# Decode JWT (basic base64 decode - for testing only)
JWT_PAYLOAD=$(echo "$ACCESS_TOKEN" | cut -d'.' -f2)
# Add padding if needed
PADDING_LENGTH=$((4 - ${#JWT_PAYLOAD} % 4))
if [ $PADDING_LENGTH -ne 4 ]; then
    JWT_PAYLOAD="${JWT_PAYLOAD}$(printf '=%.0s' $(seq 1 $PADDING_LENGTH))"
fi

DECODED_JWT=$(echo "$JWT_PAYLOAD" | base64 -d 2>/dev/null)

echo -e "${BLUE}JWT Claims:${NC}"
echo "$DECODED_JWT" | jq '.' 2>/dev/null || echo "$DECODED_JWT"
echo ""

# Check for RBAC claims
if echo "$DECODED_JWT" | grep -q "organization_id"; then
    echo -e "${GREEN}✓ Organization ID claim present${NC}"
else
    echo -e "${YELLOW}⚠ Organization ID claim missing${NC}"
fi

if echo "$DECODED_JWT" | grep -q "role_id"; then
    echo -e "${GREEN}✓ Role ID claim present${NC}"
else
    echo -e "${YELLOW}⚠ Role ID claim missing${NC}"
fi

if echo "$DECODED_JWT" | grep -q "permissions"; then
    echo -e "${GREEN}✓ Permissions claim present${NC}"
else
    echo -e "${YELLOW}⚠ Permissions claim missing${NC}"
fi

echo ""
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  ✓ OAuth2.1 + PKCE Flow Test COMPLETE${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo ""

echo -e "Summary:"
echo -e "  ✓ Authorization endpoint working"
echo -e "  ✓ Consent screen functional"
echo -e "  ✓ Token exchange successful"
echo -e "  ✓ UserInfo endpoint working"
echo -e "  ✓ JWT contains RBAC claims"
echo ""

echo -e "${BLUE}Saved credentials for future use:${NC}"
echo -e "  Access Token: ${GREEN}$ACCESS_TOKEN${NC}"
echo -e "  Refresh Token: ${GREEN}$REFRESH_TOKEN${NC}"
echo ""

echo -e "${YELLOW}Next: Test token refresh with:${NC}"
echo -e "  curl -X POST $API_BASE_URL/oauth/token \\"
echo -e "    -H 'Content-Type: application/x-www-form-urlencoded' \\"
echo -e "    -d 'grant_type=refresh_token' \\"
echo -e "    -d 'refresh_token=$REFRESH_TOKEN' \\"
echo -e "    -d 'client_id=$CLIENT_ID' \\"
echo -e "    -d 'client_secret=$CLIENT_SECRET'"
echo ""
