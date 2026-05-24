#!/bin/bash

# Multi-Carrier Shipping Platform - Development Launcher
# Opens each service in a separate macOS Terminal tab

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting Multi-Carrier Shipping Platform in Development Mode${NC}"
echo ""

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Check if Docker infrastructure is running
echo -e "${GREEN}Checking Docker infrastructure...${NC}"
if ! docker compose ps | grep -q "kafka"; then
    echo "Starting infrastructure (Kafka + PostgreSQL)..."
    docker compose up -d zookeeper kafka postgres-shipment postgres-carrier postgres-rate postgres-label postgres-tracking postgres-address postgres-billing postgres-return
    echo "Waiting 15 seconds for services to initialize..."
    sleep 15
else
    echo "Infrastructure already running ✓"
fi

echo ""
echo -e "${GREEN}Launching services in separate terminal tabs...${NC}"
echo ""

# Function to open a new terminal tab and run a command
open_tab() {
    local title="$1"
    local dir="$2"
    local cmd="$3"
    local color="$4"

    osascript <<EOF
        tell application "Terminal"
            activate
            tell application "System Events" to keystroke "t" using command down
            delay 0.2
            do script "cd $SCRIPT_DIR/$dir && echo \"\033[${color}m>>> $title\033[0m\" && $cmd" in front window
            do script "clear" in front window
        end tell
EOF
    echo "  ✓ $title"
    sleep 0.5
}

# Launch all services
echo "Launching services..."

# Infrastructure is already running via Docker
# Now launch Go services

open_tab "🔵 API Gateway (Port 8080)" "api-gateway" "go mod tidy && PORT=8080 SHIPMENT_SERVICE_URL=http://localhost:8081 CARRIER_SERVICE_URL=http://localhost:8082 RATE_SERVICE_URL=http://localhost:8083 LABEL_SERVICE_URL=http://localhost:8084 TRACKING_SERVICE_URL=http://localhost:8085 ADDRESS_SERVICE_URL=http://localhost:8086 BILLING_SERVICE_URL=http://localhost:8087 RETURN_SERVICE_URL=http://localhost:8088 go run ./cmd" "34"

open_tab "📦 Shipment Service (Port 8081)" "shipment-service" "go mod tidy && PORT=8081 DB_HOST=localhost DB_PORT=5431 DB_USER=postgres DB_PASS=postgres DB_NAME=shipments KAFKA_BROKERS=localhost:9092 go run ./cmd" "32"

open_tab "🚚 Carrier Integration (Port 8082)" "carrier-integration-service" "go mod tidy && PORT=8082 DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASS=postgres DB_NAME=carriers go run ./cmd" "33"

open_tab "💰 Rate Comparison (Port 8083)" "rate-comparison-service" "go mod tidy && PORT=8083 DB_HOST=localhost DB_PORT=5433 DB_USER=postgres DB_PASS=postgres DB_NAME=rates KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082 go run ./cmd" "35"

open_tab "🏷️  Label Generation (Port 8084)" "label-generation-service" "go mod tidy && PORT=8084 DB_HOST=localhost DB_PORT=5434 DB_USER=postgres DB_PASS=postgres DB_NAME=labels KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082 go run ./cmd" "36"

open_tab "📍 Tracking Service (Port 8085)" "tracking-service" "go mod tidy && PORT=8085 DB_HOST=localhost DB_PORT=5435 DB_USER=postgres DB_PASS=postgres DB_NAME=tracking KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082 go run ./cmd" "31"

open_tab "📮 Address Validation (Port 8086)" "address-validation-service" "go mod tidy && PORT=8086 DB_HOST=localhost DB_PORT=5436 DB_USER=postgres DB_PASS=postgres DB_NAME=addresses KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082 go run ./cmd" "34"

open_tab "💳 Billing Service (Port 8087)" "billing-service" "go mod tidy && PORT=8087 DB_HOST=localhost DB_PORT=5437 DB_USER=postgres DB_PASS=postgres DB_NAME=billing KAFKA_BROKERS=localhost:9092 go run ./cmd" "32"

open_tab "🔄 Return Service (Port 8088)" "return-service" "go mod tidy && PORT=8088 DB_HOST=localhost DB_PORT=5438 DB_USER=postgres DB_PASS=postgres DB_NAME=returns KAFKA_BROKERS=localhost:9092 LABEL_SERVICE_URL=http://localhost:8084 go run ./cmd" "33"

open_tab "📧 Notification Service (Kafka Consumer)" "notification-service" "go mod tidy && KAFKA_BROKERS=localhost:9092 go run ./cmd" "35"

echo ""
echo -e "${GREEN}All services launched!${NC}"
echo ""
echo "API Gateway: http://localhost:8080"
echo ""
echo "Test commands:"
echo "  make test-flow"
echo ""
echo "View logs:"
echo "  docker compose logs -f kafka"
echo ""
echo "Stop all:"
echo "  docker compose down -v"
echo "  # Then close terminal tabs manually"
