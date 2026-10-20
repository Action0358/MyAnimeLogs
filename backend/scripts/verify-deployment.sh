#!/bin/bash
set -e
set -o pipefail

# MyAnimeLogs Deployment Verification Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "SUCCESS" ]; then
        echo -e "${GREEN}✅ $message${NC}"
    elif [ "$status" = "ERROR" ]; then
        echo -e "${RED}❌ $message${NC}"
    elif [ "$status" = "INFO" ]; then
        echo -e "${YELLOW}ℹ️  $message${NC}"
    fi
}

# Function to test API endpoint
test_endpoint() {
    local url=$1
    local description=$2
    local expected_status=${3:-200}
    
    echo "Testing: $description"
    echo "URL: $url"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$url")
    http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')

    
    if [ "$http_code" -eq "$expected_status" ]; then
        print_status "SUCCESS" "$description (HTTP $http_code)"
        echo "Response: $(echo $body | jq . 2>/dev/null || echo $body | head -c 100)..."
    else
        print_status "ERROR" "$description (HTTP $http_code, expected $expected_status)"
        echo "Response: $body"
    fi
    echo "---"
}

# Function to verify Annict API functionality
verify_annict_api() {
    local base_url=$1
    local env_name=$2
    
    print_status "INFO" "🎬 Verifying Annict API Integration for $env_name"
    echo "=================================================="
    
    # 利用者向けの一覧は同期済みのDBから返す（Annictの中継ルートは撤去済み）
    test_endpoint "$base_url/api/animes/current?limit=5" "Current Season Animes (DB)"
    test_endpoint "$base_url/api/animes/popular?limit=5" "Popular Animes (DB)"

    # Test API response structure
    echo "Checking response structure..."
    response=$(curl -s "$base_url/api/animes/current?limit=5")
    if echo "$response" | jq -e '.animes' > /dev/null 2>&1; then
        print_status "SUCCESS" "Current season API returns expected 'animes' field"
    else
        print_status "ERROR" "Current season API response structure unexpected"
        echo "Response sample: $(echo $response | head -c 200)..."
    fi
    
    echo ""
}

# Function to verify environment
verify_environment() {
    local env_name=$1
    local base_url=$2
    local frontend_origin=$3
    
    print_status "INFO" "Verifying $env_name Environment: $base_url"
    echo "=================================================="
    
    # Basic health check
    test_endpoint "$base_url/health" "Health Check"
    
    # Annict API connection test (development endpoint)
    if curl -s "$base_url/annict/test" | grep -q "error"; then
        print_status "ERROR" "Annict API Connection Test endpoint returned error"
    else
        test_endpoint "$base_url/annict/test" "Annict API Connection Test"
    fi
    
    # Get anime list (from local DB)
    test_endpoint "$base_url/api/animes?page=1&per_page=5" "Local Anime Database"
    
    # Test CORS headers
    echo "Testing CORS headers..."
    
    # Test primary frontend origin
    cors_response=$(curl -s -I -H "Origin: $frontend_origin" "$base_url/health")
    if echo "$cors_response" | grep -q "access-control-allow-origin"; then
        print_status "SUCCESS" "CORS headers present for origin: $frontend_origin"
    else
        print_status "ERROR" "CORS headers missing for origin: $frontend_origin"
    fi
    
    # Test localhost origin (for development)
    cors_response_localhost=$(curl -s -I -H "Origin: http://localhost:3000" "$base_url/health")
    if echo "$cors_response_localhost" | grep -q "access-control-allow-origin"; then
        print_status "SUCCESS" "CORS headers present for localhost development"
    else
        print_status "INFO" "CORS headers not configured for localhost (expected in production)"
    fi
    echo "---"
    
    echo ""
    
    # Run detailed Annict API verification
    verify_annict_api "$base_url" "$env_name"
}

# Main verification
main() {
    echo "🚀 MyAnimeLogs Deployment Verification"
    echo "======================================"
    echo ""
    
    # Check if jq is available for JSON formatting
    if ! command -v jq &> /dev/null; then
        print_status "INFO" "jq not found. JSON responses will not be formatted."
    fi
    
    # Verify staging environment
    if [ "${1:-all}" = "staging" ] || [ "${1:-all}" = "all" ]; then
        verify_environment "STAGING" "https://myanimelogs-staging.fly.dev" "https://stg.myanimelogs.com"
    fi
    
    # Verify production environment
    if [ "${1:-all}" = "production" ] || [ "${1:-all}" = "all" ]; then
        verify_environment "PRODUCTION" "https://myanimelogs-production.fly.dev" "https://myanimelogs.com"
    fi
    
    print_status "INFO" "Verification completed!"
}

# Help function
show_help() {
    echo "Usage: $0 [staging|production|all]"
    echo ""
    echo "Options:"
    echo "  staging     Verify only staging environment"
    echo "  production  Verify only production environment"
    echo "  all         Verify both environments (default)"
    echo "  help        Show this help message"
    echo ""
    echo "This script verifies:"
    echo "  - API endpoint functionality"
    echo "  - Annict API integration"
    echo "  - CORS configuration for appropriate origins"
    echo "  - Health checks and response structure"
    echo ""
    echo "Examples:"
    echo "  $0                    # Verify both environments"
    echo "  $0 staging           # Verify only staging"
    echo "  $0 production        # Verify only production"
}

# Parse arguments
case "${1:-all}" in
    "help"|"-h"|"--help")
        show_help
        exit 0
        ;;
    "staging"|"production"|"all")
        main "$1"
        ;;
    *)
        echo "Invalid argument: $1"
        show_help
        exit 1
        ;;
esac