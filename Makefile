.PHONY: all run run-shipment run-label run-auth run-notification run-carrier-stats build test fmt vet clean

# Variables
BUILD_DIR=./bin
KAFKA_BROKERS?=localhost:9092
FREIGHTPULSE_API_KEY?=

all: fmt vet test build

run-shipment:
	@PORT=8081 LABEL_SERVICE_URL=http://localhost:8082 AUTH_SERVICE_URL=http://localhost:8083 KAFKA_BROKERS=$(KAFKA_BROKERS) go run ./cmd/shipment_service/main.go

run-label:
	@PORT=8082 SHIPMENT_SERVICE_URL=http://localhost:8081 AUTH_SERVICE_URL=http://localhost:8083 go run ./cmd/label_service/main.go

run-auth:
	@PORT=8083 go run ./cmd/auth_service/main.go

run-notification:
	@PORT=8084 KAFKA_BROKERS=$(KAFKA_BROKERS) go run ./cmd/notification_service/main.go

run-carrier-stats:
	@PORT=8085 FREIGHTPULSE_BASE_URL=https://freightpulsehq.com/api/v1 FREIGHTPULSE_API_KEY=$(FREIGHTPULSE_API_KEY) MONGO_URI=mongodb://localhost:27017 MONGO_DB=carrier_stats_logs go run ./cmd/carrier_stats_service/main.go

run:
	@echo "Starting Shipment, Label, Auth, and Customer Notification microservices..."
	@(trap 'kill 0' SIGINT; \
	PORT=8083 go run ./cmd/auth_service/main.go & \
	PORT=8082 SHIPMENT_SERVICE_URL=http://localhost:8081 AUTH_SERVICE_URL=http://localhost:8083 go run ./cmd/label_service/main.go & \
	PORT=8084 KAFKA_BROKERS=$(KAFKA_BROKERS) go run ./cmd/notification_service/main.go & \
	PORT=8081 LABEL_SERVICE_URL=http://localhost:8082 AUTH_SERVICE_URL=http://localhost:8083 NOTIFICATION_SERVICE_URL=http://localhost:8084 KAFKA_BROKERS=$(KAFKA_BROKERS) go run ./cmd/shipment_service/main.go)

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/shipment_service ./cmd/shipment_service/main.go
	@go build -o $(BUILD_DIR)/label_service ./cmd/label_service/main.go
	@go build -o $(BUILD_DIR)/auth_service ./cmd/auth_service/main.go
	@go build -o $(BUILD_DIR)/notification_service ./cmd/notification_service/main.go
	@go build -o $(BUILD_DIR)/carrier_stats_service ./cmd/carrier_stats_service/main.go

test:
	@go test -v ./...

fmt:
	@go fmt ./...

vet:
	@go fmt ./... && go vet ./...

clean:
	@rm -rf $(BUILD_DIR)
	@go clean
