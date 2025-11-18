#!/bin/bash

# Log Detect Authentication Test Script
# Usage: ./test_auth.sh [server_url]
# Default server_url: http://localhost:8006

SERVER_URL=${1:-"http://localhost:8006"}

echo "🔐 Testing Log Detect Authentication System"
echo "============================================"
echo "Server URL: $SERVER_URL"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "success" ]; then
        echo -e "${GREEN}✅ $message${NC}"
    elif [ "$status" = "error" ]; then
        echo -e "${RED}❌ $message${NC}"
    else
        echo -e "${YELLOW}⚠️  $message${NC}"
    fi
}

# Test 0: Check if tables exist
echo "0. Checking Database Tables..."
# Try to login - if it fails with table not found, tables don't exist
LOGIN_TEST=$(curl -s -w "\n%{http_code}" -X POST "$SERVER_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' 2>/dev/null)

HTTP_CODE=$(echo "$LOGIN_TEST" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_status "success" "Database tables exist"
elif [ "$HTTP_CODE" = "401" ]; then
    print_status "success" "Database tables exist (login validation working)"
else
    print_status "error" "Database tables may not exist or service not running (HTTP $HTTP_CODE)"
    echo "   Try running: go run create_tables.go"
    exit 1
fi
echo ""

# Test 1: Health Check
echo "1. Testing Health Check..."
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$SERVER_URL/healthcheck")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_status "success" "Health check passed"
else
    print_status "error" "Health check failed (HTTP $HTTP_CODE)"
fi
echo ""

# Test 2: Login with admin credentials
echo "2. Testing Admin Login..."
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$SERVER_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    # Extract token from response
    TOKEN=$(echo "$LOGIN_BODY" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    if [ -n "$TOKEN" ]; then
        print_status "success" "Admin login successful"
        echo "   Token: ${TOKEN:0:50}..."
    else
        print_status "error" "Login response missing token"
        exit 1
    fi
else
    print_status "error" "Admin login failed (HTTP $HTTP_CODE)"
    echo "   Response: $LOGIN_BODY"
    exit 1
fi
echo ""

# Test 3: Access protected endpoint with valid token
echo "3. Testing Protected Endpoint Access..."
DEVICE_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" \
    "$SERVER_URL/api/v1/Device/GetAll")

HTTP_CODE=$(echo "$DEVICE_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_status "success" "Protected endpoint access successful"
elif [ "$HTTP_CODE" = "401" ]; then
    print_status "error" "Authentication failed"
elif [ "$HTTP_CODE" = "403" ]; then
    print_status "error" "Authorization failed"
else
    print_status "error" "Unexpected response (HTTP $HTTP_CODE)"
fi
echo ""

# Test 4: Access without token
echo "4. Testing Access Without Token..."
NO_TOKEN_RESPONSE=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/v1/Device/GetAll")

HTTP_CODE=$(echo "$NO_TOKEN_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_status "success" "Correctly rejected unauthorized access"
else
    print_status "error" "Should reject unauthorized access (HTTP $HTTP_CODE)"
fi
echo ""

# Test 5: Get user profile
echo "5. Testing Get User Profile..."
PROFILE_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" \
    "$SERVER_URL/api/v1/auth/profile")

HTTP_CODE=$(echo "$PROFILE_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_status "success" "User profile access successful"
else
    print_status "error" "User profile access failed (HTTP $HTTP_CODE)"
fi
echo ""

# Test 6: Test invalid token
echo "6. Testing Invalid Token..."
INVALID_TOKEN_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer invalid.token.here" \
    "$SERVER_URL/api/v1/Device/GetAll")

HTTP_CODE=$(echo "$INVALID_TOKEN_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_status "success" "Correctly rejected invalid token"
else
    print_status "error" "Should reject invalid token (HTTP $HTTP_CODE)"
fi
echo ""

# Test 7: Test expired/invalid token format
echo "7. Testing Malformed Authorization Header..."
MALFORMED_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: InvalidFormat" \
    "$SERVER_URL/api/v1/Device/GetAll")

HTTP_CODE=$(echo "$MALFORMED_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_status "success" "Correctly rejected malformed authorization header"
else
    print_status "error" "Should reject malformed authorization header (HTTP $HTTP_CODE)"
fi
echo ""

echo "🎉 Authentication Tests Completed!"
echo "=================================="
echo ""
echo "Next Steps:"
echo "1. Try creating a regular user and test role-based permissions"
echo "2. Test different endpoints with appropriate permissions"
echo "3. Verify Swagger documentation includes auth requirements"
echo ""
echo "For more details, see README_AUTH.md"
