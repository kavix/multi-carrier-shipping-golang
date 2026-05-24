#!/bin/bash

# Multi-Carrier Shipping Platform - Background Launcher
# Runs all services in background with log files

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"
mkdir -p logs

# Start infrastructure
echo "Starting Docker infrastructure..."
docker compose up -d zookeeper kafka postgres-shipment postgres-carrier postgres-rate postgres-label postgres-tracking postgres-address postgres-billing postgres-return

echo "Waiting 15 seconds..."
sleep 15

# Function to start a service in background
start_service() {
    local name="$1"
    local dir="$2"
    local env_vars="$3"

    cd "$SCRIPT_DIR/$dir"
    go mod tidy >/dev/null 2>&1
    eval "$env_vars go run ./cmd > "$SCRIPT_DIR/logs/$name.log" 2>&1 &"
    echo "$!" > "$SCRIPT_DIR/logs/$name.pid"
    echo "Started $name (PID: $!, log: logs/$name.log)"
}

echo "Starting services..."

start_service "gateway" "api-gateway" "PORT=8080 SHIPMENT_SERVICE_URL=http://localhost:8081 CARRIER_SERVICE_URL=http://localhost:8082 RATE_SERVICE_URL=http://localhost:8083 LABEL_SERVICE_URL=http://localhost:8084 TRACKING_SERVICE_URL=http://localhost:8085 ADDRESS_SERVICE_URL=http://localhost:8086 BILLING_SERVICE_URL=http://localhost:8087 RETURN_SERVICE_URL=http://localhost:8088"

start_service "shipment" "shipment-service" "PORT=8081 DB_HOST=localhost DB_PORT=5431 DB_USER=postgres DB_PASS=postgres DB_NAME=shipments KAFKA_BROKERS=localhost:9092"

start_service "carrier" "carrier-integration-service" "PORT=8082 DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASS=postgres DB_NAME=carriers"

start_service "rate" "rate-comparison-service" "PORT=8083 DB_HOST=localhost DB_PORT=5433 DB_USER=postgres DB_PASS=postgres DB_NAME=rates KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082"

start_service "label" "label-generation-service" "PORT=8084 DB_HOST=localhost DB_PORT=5434 DB_USER=postgres DB_PASS=postgres DB_NAME=labels KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082"

start_service "tracking" "tracking-service" "PORT=8085 DB_HOST=localhost DB_PORT=5435 DB_USER=postgres DB_PASS=postgres DB_NAME=tracking KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082"

start_service "address" "address-validation-service" "PORT=8086 DB_HOST=localhost DB_PORT=5436 DB_USER=postgres DB_PASS=postgres DB_NAME=addresses KAFKA_BROKERS=localhost:9092 CARRIER_SERVICE_URL=http://localhost:8082"

start_service "billing" "billing-service" "PORT=8087 DB_HOST=localhost DB_PORT=5437 DB_USER=postgres DB_PASS=postgres DB_NAME=billing KAFKA_BROKERS=localhost:9092"

start_service "return" "return-service" "PORT=8088 DB_HOST=localhost DB_PORT=5438 DB_USER=postgres DB_PASS=postgres DB_NAME=returns KAFKA_BROKERS=localhost:9092 LABEL_SERVICE_URL=http://localhost:8084"

start_service "notification" "notification-service" "KAFKA_BROKERS=localhost:9092"

echo ""
echo "All services started in background!"
echo ""
echo "API Gateway: http://localhost:8080"
echo ""
echo "View logs:"
echo "  tail -f logs/gateway.log"
echo "  tail -f logs/shipment.log"
echo ""
echo "Stop all services:"
echo "  ./stop-dev.sh"
echo ""
echo "Or manually:"
echo "  for pid in logs/*.pid; do kill \$(cat \$pid); done"
