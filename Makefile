.PHONY: all run run-shipment run-label run-auth run-notification build test fmt vet clean

# Variables
BUILD_DIR=./bin

all: fmt vet test build

run-shipment:
	@PORT=8081 LABEL_SERVICE_URL=http://localhost:8082 AUTH_SERVICE_URL=http://localhost:8083 go run ./cmd/shipment_service/main.go

run-label:
	@PORT=8082 SHIPMENT_SERVICE_URL=http://localhost:8081 AUTH_SERVICE_URL=http://localhost:8083 go run ./cmd/label_service/main.go

run-auth:
	@PORT=8083 go run ./cmd/auth_service/main.go

run-notification:
	@PORT=8084 go run ./cmd/notification_service/main.go

run:
	@echo "Starting Shipment, Label, Auth, and Customer Notification microservices..."
	@(trap 'kill 0' SIGINT; \
	PORT=8083 go run ./cmd/auth_service/main.go & \
	PORT=8082 SHIPMENT_SERVICE_URL=http://localhost:8081 AUTH_SERVICE_URL=http://localhost:8083 go run ./cmd/label_service/main.go & \
	PORT=8084 go run ./cmd/notification_service/main.go & \
	PORT=8081 LABEL_SERVICE_URL=http://localhost:8082 AUTH_SERVICE_URL=http://localhost:8083 NOTIFICATION_SERVICE_URL=http://localhost:8084 go run ./cmd/shipment_service/main.go)

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/shipment_service ./cmd/shipment_service/main.go
	@go build -o $(BUILD_DIR)/label_service ./cmd/label_service/main.go
	@go build -o $(BUILD_DIR)/auth_service ./cmd/auth_service/main.go
	@go build -o $(BUILD_DIR)/notification_service ./cmd/notification_service/main.go

test:
	@go test -v ./...

fmt:
	@go fmt ./...

vet:
	@go fmt ./... && go vet ./...

clean:
	@rm -rf $(BUILD_DIR)
	@go clean
