#!/bin/bash

# Multi-Carrier Shipping Platform - All-in-One Startup Script
# Starts all services, infrastructure, and applies migrations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

print_header "Multi-Carrier Shipping Platform Startup"

# Check if Docker is running
print_info "Checking Docker..."
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker Desktop."
    exit 1
fi
print_success "Docker is running"

# Check if services are already running
if docker compose ps 2>/dev/null | grep -q "Up"; then
    print_info "Some containers are already running. Stopping them..."
    docker compose down 2>/dev/null
fi

# Start all services
print_info "Starting all services, databases, and infrastructure..."
docker compose up -d

print_success "All containers started"

# Wait for services to be ready
print_info "Waiting for services to initialize (30 seconds)..."
for i in {1..30}; do
    echo -ne "\r  Waiting... ${i}s/30s"
    sleep 1
done
echo -e "\r  Ready!                    \n"

# Apply migrations
print_header "Applying Database Migrations"

# Function to run migration
run_migration() {
    local port=$1
    local db_name=$2
    local migration_file=$3
    local service_name=$4
    
    print_info "Applying $service_name migration..."
    
    # Check if migration file exists
    if [ ! -f "$migration_file" ]; then
        print_error "Migration file not found: $migration_file"
        return 1
    fi
    
    # Run migration with retry
    max_retries=5
    retry_count=0
    while [ $retry_count -lt $max_retries ]; do
        if psql "postgres://postgres:postgres@localhost:$port/$db_name" -f "$migration_file" > /dev/null 2>&1; then
            print_success "$service_name migration completed"
            return 0
        fi
        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            print_info "Retrying... ($retry_count/$max_retries)"
            sleep 2
        fi
    done
    
    print_error "$service_name migration failed after $max_retries attempts"
    return 1
}

# Apply all migrations
run_migration 5431 shipments shipment-service/migrations/001_create_shipments.sql "Shipment Service" || true
run_migration 5432 carriers carrier-integration-service/migrations/001_create_carriers.sql "Carrier Integration" || true
run_migration 5433 rates rate-comparison-service/migrations/001_create_rates.sql "Rate Comparison" || true
run_migration 5434 labels label-generation-service/migrations/001_create_labels.sql "Label Generation" || true
run_migration 5435 tracking tracking-service/migrations/001_create_tracking.sql "Tracking Service" || true
run_migration 5436 addresses address-validation-service/migrations/001_create_addresses.sql "Address Validation" || true
run_migration 5437 billing billing-service/migrations/001_create_billing.sql "Billing Service" || true
run_migration 5438 returns return-service/migrations/001_create_returns.sql "Return Service" || true

# Verify services are running
print_header "Service Status"

# Check API Gateway health
if curl -s -H "Authorization: Bearer health-check" http://localhost:8080/health > /dev/null 2>&1; then
    print_success "API Gateway (8080) is responding"
else
    print_info "Waiting for API Gateway to respond..."
    sleep 5
    if curl -s -H "Authorization: Bearer health-check" http://localhost:8080/health > /dev/null 2>&1; then
        print_success "API Gateway (8080) is responding"
    else
        print_error "API Gateway not responding"
    fi
fi

# Show all container status
echo ""
print_info "Container Status:"
docker compose ps --format "table {{.Name}}\t{{.Status}}" | tail -n +2 | while read name status; do
    if [[ $status == *"Up"* ]]; then
        echo -e "  ${GREEN}✓${NC} $(basename $name): $status"
    else
        echo -e "  ${RED}✗${NC} $(basename $name): $status"
    fi
done

print_header "Platform Ready! 🚀"

echo -e "${GREEN}Services running on:${NC}"
echo "  • API Gateway:          http://localhost:8080"
echo "  • Shipment Service:     http://localhost:8081"
echo "  • Carrier Integration:  http://localhost:8082"
echo "  • Rate Comparison:      http://localhost:8083"
echo "  • Label Generation:     http://localhost:8084"
echo "  • Tracking Service:     http://localhost:8085"
echo "  • Address Validation:   http://localhost:8086"
echo "  • Billing Service:      http://localhost:8087"
echo "  • Return Service:       http://localhost:8088"
echo "  • Kafka:                localhost:9092"
echo ""
echo -e "${YELLOW}Test the API:${NC}"
echo '  curl -H "Authorization: Bearer test-token" http://localhost:8080/health'
echo ""
echo -e "${YELLOW}View logs:${NC}"
echo "  docker compose logs -f"
echo ""
echo -e "${YELLOW}Stop all services:${NC}"
echo "  docker compose down -v"
echo ""
