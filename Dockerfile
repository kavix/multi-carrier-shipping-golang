# Step 1: Build the binary
FROM golang:1.22-alpine AS builder

# Install system dependencies (e.g., git/certificates if needed)
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Copy dependency files
COPY go.mod ./
# RUN go mod download # Uncomment when dependencies are added

# Copy source code
COPY . .

# Build standard statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server ./cmd/server/main.go

# Step 2: Create execution container
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

# Create a non-privileged system user for running the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the pre-built binary
COPY --from=builder /app/server .

# Use non-privileged user
USER appuser

# Expose port
EXPOSE 8080

# Run the app
ENTRYPOINT ["./server"]
