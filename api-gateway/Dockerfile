# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS builder

WORKDIR /src

# Download dependencies for shared
COPY shared/go.mod shared/go.sum ./shared/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd shared && go mod download

# Download dependencies for the service
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd api-gateway && go mod download

# Copy source
COPY shared/ ./shared/
COPY api-gateway/ ./api-gateway/

# Build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd api-gateway && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/service ./cmd

# Final image
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/service /bin/service
EXPOSE 8080
ENTRYPOINT ["/bin/service"]
