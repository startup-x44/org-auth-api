#!/bin/bash

# Test API Key endpoints directly
# Replace JWT_TOKEN with a valid JWT from your browser's localStorage

JWT_TOKEN="YOUR_JWT_TOKEN_HERE"
BASE_URL="http://localhost:8080/api/v1"

echo "🧪 Testing API Key Endpoints..."
echo "================================"

# 1. Create API Key
echo -e "\n1️⃣ Creating API Key..."
curl -X POST "$BASE_URL/dev/api-keys" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "name": "Test API Key",
    "description": "Testing the fixed API key system",
    "scopes": ["read", "write"]
  }' | jq .

echo -e "\n2️⃣ Listing API Keys..."
curl -X GET "$BASE_URL/dev/api-keys" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .

echo -e "\n3️⃣ Getting specific API Key (replace KEY_ID)..."
echo "curl -X GET \"$BASE_URL/dev/api-keys/ak_YOUR_KEY_ID\" -H \"Authorization: Bearer \$JWT_TOKEN\""

echo -e "\n4️⃣ Revoking API Key (replace KEY_ID)..."
echo "curl -X DELETE \"$BASE_URL/dev/api-keys/ak_YOUR_KEY_ID\" -H \"Authorization: Bearer \$JWT_TOKEN\""

echo -e "\n📝 To use this script:"
echo "1. Get your JWT token from browser localStorage"
echo "2. Replace JWT_TOKEN variable above"
echo "3. Run: chmod +x test_api_endpoints.sh && ./test_api_endpoints.sh"