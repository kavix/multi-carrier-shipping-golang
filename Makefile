.PHONY: all run build test fmt vet clean

# Variables
BINARY_NAME=server
CMD_DIR=./cmd/server
BUILD_DIR=./bin

all: fmt vet test build

run:
	@go run $(CMD_DIR)/main.go

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

test:
	@go test -v ./...

fmt:
	@go fmt ./...

vet:
	@go vet ./...

clean:
	@rm -rf $(BUILD_DIR)
	@go clean
