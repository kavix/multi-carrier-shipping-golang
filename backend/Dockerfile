FROM golang:1.22-alpine AS builder

WORKDIR /workspace

COPY go.mod .
RUN go mod download
COPY . .

RUN go build -o /usr/local/bin/order-service ./cmd/order-service
RUN go build -o /usr/local/bin/carrier-service ./cmd/carrier-service
RUN go build -o /usr/local/bin/tracking-service ./cmd/tracking-service
RUN go build -o /usr/local/bin/api-gateway ./cmd/api-gateway

FROM alpine:3.19
RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /usr/local/bin /usr/local/bin
USER app
WORKDIR /app

EXPOSE 8080 8081 8082 8083

ENV ORDER_ADDR=:8081 \
    CARRIER_ADDR=:8082 \
    TRACKING_ADDR=:8083 \
    GATEWAY_ADDR=:8080

CMD ["/usr/local/bin/api-gateway"]
