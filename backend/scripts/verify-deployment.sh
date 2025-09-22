#!/bin/bash

# MyAnimeLogs Deployment Verification Script

set -e

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
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')
    
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
    
    # Test Annict API endpoints with various parameters
    echo "Testing Annict API endpoints..."
    
    # Test current season
    test_endpoint "$base_url/api/annict/works" "Annict: Current Season Works"
    
    # Test popular works
    test_endpoint "$base_url/api/annict/popular-works" "Annict: Popular Works"
    
    # Test with specific season
    test_endpoint "$base_url/api/annict/works?season=2024-winter" "Annict: Winter 2024 Season"
    test_endpoint "$base_url/api/annict/works?season=2024-spring" "Annict: Spring 2024 Season"
    
    # Test with filters
    test_endpoint "$base_url/api/annict/works?filter_ids=1,2,3" "Annict: Works with Filter IDs"
    
    # Test API response structure
    echo "Checking Annict API response structure..."
    response=$(curl -s "$base_url/api/annict/works")
    if echo "$response" | jq -e '.works' > /dev/null 2>&1; then
        print_status "SUCCESS" "Annict API returns expected 'works' field"
    elif echo "$response" | jq -e '.data' > /dev/null 2>&1; then
        print_status "SUCCESS" "Annict API returns expected 'data' field"
    else
        print_status "ERROR" "Annict API response structure unexpected"
        echo "Response sample: $(echo $response | head -c 200)..."
    fi
    
    echo ""
}

# Function to verify environment
verify_environment() {
    local env_name=$1
    local base_url=$2
    
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
    
    # Get current season anime (from Annict)
    test_endpoint "$base_url/api/annict/works" "Current Season Anime (Annict)"
    
    # Get popular anime (from Annict)
    test_endpoint "$base_url/api/annict/popular-works" "Popular Anime (Annict)"
    
    # Test Annict API with specific parameters
    test_endpoint "$base_url/api/annict/works?season=2024-winter" "Annict API with Season Parameter"
    
    # Get anime list (from local DB)
    test_endpoint "$base_url/api/animes?page=1&per_page=5" "Local Anime Database"
    
    # Test CORS headers
    echo "Testing CORS headers..."
    cors_response=$(curl -s -I -H "Origin: https://stg.myanimelogs.com" "$base_url/health")
    if echo "$cors_response" | grep -q "access-control-allow-origin"; then
        print_status "SUCCESS" "CORS headers present"
    else
        print_status "ERROR" "CORS headers missing"
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
        verify_environment "STAGING" "https://myanimelogs-staging.fly.dev"
    fi
    
    # Verify production environment
    if [ "${1:-all}" = "production" ] || [ "${1:-all}" = "all" ]; then
        verify_environment "PRODUCTION" "https://myanimelogs-production.fly.dev"
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