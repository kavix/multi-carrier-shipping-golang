# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Download dependencies for shared
COPY shared/go.mod shared/go.sum ./shared/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd shared && go mod download

# Download dependencies for the service
COPY notification-service/go.mod notification-service/go.sum ./notification-service/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd notification-service && go mod download

# Copy source
COPY shared/ ./shared/
COPY notification-service/ ./notification-service/

# Build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd notification-service && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/service ./cmd

# Final image
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/service /bin/service
ENTRYPOINT ["/bin/service"]
