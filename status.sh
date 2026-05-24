#!/bin/bash

# Multi-Carrier Shipping Platform - Status Script
# Check the health of all services

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
}

print_running() {
    echo -e "  ${GREEN}✓ $1${NC} (running)"
}

print_stopped() {
    echo -e "  ${RED}✗ $1${NC} (stopped)"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

print_header "Multi-Carrier Shipping Platform Status"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Docker is not running${NC}"
    exit 1
fi

# Get container status
echo -e "${YELLOW}Infrastructure:${NC}"
docker compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null | grep -E "(zookeeper|kafka|postgres)" | while read name status; do
    if [[ $status == *"Up"* ]]; then
        print_running "$(basename $name)"
    else
        print_stopped "$(basename $name)"
    fi
done

echo -e "\n${YELLOW}Microservices:${NC}"
docker compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null | grep -E "(shipment|carrier|rate|label|tracking|address|billing|return|notification|api-gateway)" | while read name status; do
    if [[ $status == *"Up"* ]]; then
        print_running "$(basename $name)"
    else
        print_stopped "$(basename $name)"
    fi
done

# Check API Gateway health
echo -e "\n${YELLOW}API Gateway Health:${NC}"
if curl -s -H "Authorization: Bearer health-check" http://localhost:8080/health 2>/dev/null | grep -q "ok"; then
    echo -e "  ${GREEN}✓ Responding${NC}"
    echo "  Response:"
    curl -s -H "Authorization: Bearer health-check" http://localhost:8080/health | sed 's/^/    /'
else
    echo -e "  ${RED}✗ Not responding${NC}"
fi

echo ""
