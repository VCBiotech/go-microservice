#!/bin/bash

# File Manager Service Test Script
# This script demonstrates how to use the file manager service

set -e

echo "🚀 File Manager Service Test Script"
echo "=================================="

# Configuration
SERVICE_URL="http://localhost:3000"
SERVER_ID="calculator-server"
PIN="123"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if service is running
print_status "Checking if service is running..."
if ! curl -s "$SERVICE_URL/health" > /dev/null; then
    print_error "Service is not running at $SERVICE_URL"
    print_status "Please start the service first with: make dev"
    exit 1
fi
print_success "Service is running!"

# Check if Gotenberg is running
print_status "Checking if Gotenberg PDF service is running..."
if ! curl -s "http://localhost:3001/health" > /dev/null; then
    print_warning "Gotenberg PDF service is not running on port 3001"
    print_status "Starting Gotenberg with Docker..."
    docker run -d --name gotenberg-test -p 3001:3000 gotenberg/gotenberg:7
    print_status "Waiting for Gotenberg to start..."
    sleep 5
else
    print_success "Gotenberg is running!"
fi

# Test health endpoint
print_status "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$SERVICE_URL/health")
if [ "$HEALTH_RESPONSE" = "Ok" ]; then
    print_success "Health check passed: $HEALTH_RESPONSE"
else
    print_error "Health check failed: $HEALTH_RESPONSE"
    exit 1
fi

# Test service info endpoint
print_status "Testing service info endpoint..."
INFO_RESPONSE=$(curl -s "$SERVICE_URL/")
print_success "Service info: $INFO_RESPONSE"

# Test template rendering with example files
print_status "Testing template rendering with example files..."

# Check if example files exist
if [ ! -f "examples/invoice-template.html" ] || [ ! -f "examples/invoice-data.json" ]; then
    print_error "Example files not found. Please run this script from the project root directory."
    exit 1
fi

# Read JSON data from file
JSON_DATA=$(cat examples/invoice-data.json | tr -d '\n' | tr -d '\r')

print_status "Sending template rendering request..."
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X POST "$SERVICE_URL/v1/files/render-template" \
    -H "X-Server-ID: $SERVER_ID" \
    -H "X-PIN: $PIN" \
    -F "template=@examples/invoice-template.html" \
    -F "jsonData=$JSON_DATA")

# Extract HTTP code and response body
HTTP_CODE=$(echo "$RESPONSE" | tail -n1 | sed 's/.*HTTP_CODE://')
RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    print_success "Template rendered successfully!"
    echo "Response: $RESPONSE_BODY"
    
    # Extract any useful information from the response
    if echo "$RESPONSE_BODY" | grep -q "successfully"; then
        print_success "PDF generated and uploaded to S3!"
    fi
else
    print_error "Request failed with HTTP code: $HTTP_CODE"
    print_error "Response: $RESPONSE_BODY"
    exit 1
fi

# Test with invalid authentication
print_status "Testing invalid authentication..."
INVALID_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X POST "$SERVICE_URL/v1/files/render-template" \
    -H "X-Server-ID: invalid-server" \
    -H "X-PIN: wrong-pin" \
    -F "template=@examples/invoice-template.html" \
    -F "jsonData=$JSON_DATA")

INVALID_HTTP_CODE=$(echo "$INVALID_RESPONSE" | tail -n1 | sed 's/.*HTTP_CODE://')
if [ "$INVALID_HTTP_CODE" = "401" ]; then
    print_success "Authentication test passed - invalid credentials properly rejected"
else
    print_warning "Authentication test unexpected result: $INVALID_HTTP_CODE"
fi

# Cleanup
print_status "Cleaning up..."
if docker ps | grep -q "gotenberg-test"; then
    print_status "Stopping Gotenberg test container..."
    docker stop gotenberg-test > /dev/null
    docker rm gotenberg-test > /dev/null
    print_success "Cleanup completed"
fi

echo ""
print_success "🎉 All tests completed successfully!"
echo ""
echo "Next steps:"
echo "1. Check your configured S3 bucket for the generated PDF"
echo "2. Modify examples/invoice-data.json to test with your own data"
echo "3. Create your own HTML templates for different document types"
echo "4. Integrate the API into your applications"
echo ""
echo "For more information, see the README.md file." 