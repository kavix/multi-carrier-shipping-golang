# API Gateway

**Port**: 8080  
**Role**: Central Request Router, Authentication, Load Balancing

## Overview

The API Gateway is the single entry point for all client requests. It handles authentication, request routing, logging, and provides a unified API surface for the entire platform.

## Responsibilities

1. **Request Routing**
   - Route requests to appropriate services
   - Maintain service directory
   - Handle service discovery

2. **Authentication**
   - Validate authorization tokens
   - Extract user context
   - Enforce authorization

3. **Cross-Cutting Concerns**
   - Request logging
   - Response formatting
   - Error handling
   - CORS

4. **Load Balancing**
   - Distribute traffic
   - Handle service failures
   - Retry logic

## Architecture

```
Client Request
    ↓
Gateway Handler
    ├─ Parse Request
    ├─ Extract Auth Token
    ├─ Validate Token
    ├─ Determine Service
    └─ Proxy Request
        ↓
    Target Service
        ├─ Process Request
        └─ Return Response
        ↓
    Gateway Handler
    ├─ Log Response
    ├─ Format Response
    └─ Return to Client
```

## Request Flow

```
1. Client sends request
   POST /shipments HTTP/1.1
   Authorization: Bearer <token>
   
2. API Gateway
   ├─ Receives request
   ├─ Extracts Authorization header
   ├─ Validates token
   ├─ Extracts user_id from token
   ├─ Adds to request context
   └─ Routes to Shipment Service (8081)
   
3. Shipment Service
   ├─ Receives forwarded request
   ├─ Reads user_id from context
   ├─ Processes request
   └─ Returns response
   
4. API Gateway
   ├─ Receives response
   ├─ Logs transaction
   └─ Returns to client
```

## Service Routing

### Service Registry

```go
services := map[string]string{
    "shipment":    os.Getenv("SHIPMENT_SERVICE_URL"),
    "carrier":     os.Getenv("CARRIER_SERVICE_URL"),
    "rate":        os.Getenv("RATE_SERVICE_URL"),
    "label":       os.Getenv("LABEL_SERVICE_URL"),
    "tracking":    os.Getenv("TRACKING_SERVICE_URL"),
    "address":     os.Getenv("ADDRESS_SERVICE_URL"),
    "billing":     os.Getenv("BILLING_SERVICE_URL"),
    "return":      os.Getenv("RETURN_SERVICE_URL"),
}
```

### Route Mapping

```
Gateway Path           Target Service    Port
──────────────────────────────────────────────────────
/shipments/*     →     Shipment Service  8081
/carriers/*      →     Carrier Service   8082
/rates/*         →     Rate Service      8083
/labels/*        →     Label Service     8084
/tracking/*      →     Tracking Service  8085
/addresses/*     →     Address Service   8086
/billing/*       →     Billing Service   8087
/returns/*       →     Return Service    8088
```

## Authentication

### Token Validation

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "missing authorization"})
            c.Abort()
            return
        }
        
        // Validate token format: Bearer <token>
        if !strings.HasPrefix(token, "Bearer ") {
            c.JSON(401, gin.H{"error": "invalid token format"})
            c.Abort()
            return
        }
        
        // Extract token
        token = token[7:]
        
        // Validate token (implement your logic)
        userID := ValidateToken(token)
        if userID == "" {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        // Store in context for downstream services
        c.Set("user_id", userID)
        c.Next()
    }
}
```

### Token Format

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### User Context Propagation

```
API Gateway receives token
    ↓
Extracts user_id: "user-123"
    ↓
Adds to request context
    ↓
Forwards to downstream service
    ↓
Downstream service reads user_id
    └─ Associates operations with user
```

## Proxy Mechanism

### HTTP Proxying

```go
func (h *GatewayHandler) proxy(c *gin.Context, service string) {
    // Get service URL
    baseURL := h.services[service]
    
    // Build target URL
    targetURL := baseURL + c.Request.URL.Path
    
    // Read request body
    body, _ := io.ReadAll(c.Request.Body)
    
    // Create new request
    req, _ := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(body))
    
    // Copy headers
    for k, v := range c.Request.Header {
        for _, val := range v {
            req.Header.Add(k, val)
        }
    }
    
    // Execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    
    // Forward response
    respBody, _ := io.ReadAll(resp.Body)
    c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}
```

## Logging

### Request Logger Middleware

```
Log Format:
┌─ Timestamp
├─ HTTP Method
├─ Path
├─ Status Code
├─ Response Time
├─ User ID
└─ Error (if any)

Example:
2026-05-24 10:30:00 POST /shipments 201 145ms user-123
```

## Configuration

**Environment Variables**:
```
PORT=8080

# Service URLs
SHIPMENT_SERVICE_URL=http://shipment-service:8081
CARRIER_SERVICE_URL=http://carrier-integration-service:8082
RATE_SERVICE_URL=http://rate-comparison-service:8083
LABEL_SERVICE_URL=http://label-generation-service:8084
TRACKING_SERVICE_URL=http://tracking-service:8085
ADDRESS_SERVICE_URL=http://address-validation-service:8086
BILLING_SERVICE_URL=http://billing-service:8087
RETURN_SERVICE_URL=http://return-service:8088
```

## Error Handling

### Service Unavailable

```json
{
  "error": "service unavailable",
  "message": "Could not reach shipment-service"
}
```

### Gateway Error

```json
{
  "error": "gateway error",
  "message": "Could not process request"
}
```

## Performance

### Caching

- Don't cache requests (let services handle)
- Cache service health for 10 seconds

### Timeouts

```go
client := &http.Client{
    Timeout: 30 * time.Second,
}
```

### Load Balancing

```
Multiple API Gateway Instances
    ↓
Load Balancer (Nginx/HAProxy)
    ├─ Round Robin
    ├─ Least Connections
    └─ Health Check
    ↓
Services
```

## Monitoring

### Health Check

```bash
curl -H "Authorization: Bearer test" http://localhost:8080/health
```

### Key Metrics

- Requests per second
- Average response time
- Error rate
- Service availability

## Security Considerations

1. **HTTPS Only**: Enforce HTTPS in production
2. **Token Validation**: Always validate tokens
3. **Rate Limiting**: Limit requests per user
4. **CORS**: Configure CORS policies
5. **Input Validation**: Let services validate

## Development

### Testing Gateway

```bash
# Health check
curl -H "Authorization: Bearer test" http://localhost:8080/health

# Create shipment through gateway
curl -X POST http://localhost:8080/shipments \
  -H "Authorization: Bearer test" \
  -H "Content-Type: application/json" \
  -d '{"sender_name":"John","receiver_name":"Jane","weight":2.5}'

# Get rates through gateway
curl "http://localhost:8080/rates/compare" \
  -H "Authorization: Bearer test"
```

## Troubleshooting

### Gateway Won't Start

```bash
# Check port is available
lsof -i :8080

# Check service URLs
docker logs api-gateway | grep SERVICE_URL

# Verify all services running
docker compose ps
```

### Service Routing Issues

```bash
# Verify gateway can reach service
docker exec api-gateway curl http://shipment-service:8081/health

# Check logs
docker logs api-gateway | grep "service unavailable"

# Test direct service access
curl http://localhost:8081/health
```

## Future Enhancements

1. **Rate Limiting**: Per-user request limits
2. **API Versioning**: Multiple API versions
3. **GraphQL**: Alternative query language
4. **Caching**: Response caching layer
5. **Analytics**: Request analytics and reporting
6. **OpenAPI**: Generate API documentation
