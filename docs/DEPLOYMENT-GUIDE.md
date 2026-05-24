# Deployment Guide

## Overview

This guide covers deploying the Multi-Carrier Shipping Platform in different environments: development, staging, and production.

## Quick Start - Development

### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+
- Git
- Bash shell

### One-Command Deployment

```bash
cd /path/to/multi-carrier-shipping
./start-all.sh
```

This script:
1. Builds all Docker images
2. Creates all containers
3. Applies database migrations
4. Starts all 10 services
5. Verifies container health

### Verify Deployment

```bash
# Check all services running
docker compose ps

# View service logs
docker compose logs -f api-gateway

# Test API Gateway
curl -H "Authorization: Bearer test" http://localhost:8080/health
```

### Clean Up

```bash
./stop-all.sh     # Stop containers
docker compose down -v  # Remove containers and volumes
```

---

## Detailed Deployment Steps

### Step 1: Environment Setup

```bash
# Clone repository
git clone <repo-url>
cd multi-carrier-shipping

# Verify Docker installation
docker --version
docker-compose --version
```

### Step 2: Configure Environment Variables

Create `.env` file in root directory:

```env
# API Gateway
PORT=8080
SHIPMENT_SERVICE_URL=http://shipment-service:8081
CARRIER_SERVICE_URL=http://carrier-integration-service:8082
RATE_SERVICE_URL=http://rate-comparison-service:8083
LABEL_SERVICE_URL=http://label-generation-service:8084
TRACKING_SERVICE_URL=http://tracking-service:8085
ADDRESS_SERVICE_URL=http://address-validation-service:8086
BILLING_SERVICE_URL=http://billing-service:8087
RETURN_SERVICE_URL=http://return-service:8088

# Services
DB_USER=postgres
DB_PASS=postgres

# Kafka
KAFKA_BROKERS=kafka:29092

# External APIs
DHL_API_KEY=your-dhl-api-key
DHL_API_SECRET=your-dhl-api-secret
FEDEX_API_KEY=your-fedex-api-key
UPS_API_KEY=your-ups-api-key

# Stripe
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLIC_KEY=pk_test_...

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
```

### Step 3: Build Docker Images

```bash
# Build all services
docker compose build

# Build specific service
docker compose build shipment-service

# View build progress
docker compose build --no-cache
```

### Step 4: Start Infrastructure

```bash
# Start only infrastructure (PostgreSQL, Kafka, Zookeeper)
make dev

# Wait for services to be ready
sleep 15
```

### Step 5: Apply Migrations

```bash
# Apply all migrations
make migrate

# Or manually:
psql postgres://postgres:postgres@localhost:5431/shipments \
  -f shipment-service/migrations/001_create_shipments.sql
```

### Step 6: Start Services

```bash
# Start all services
docker compose up -d

# Or with logs
docker compose up

# View logs for specific service
docker compose logs -f shipment-service
```

### Step 7: Verify Health

```bash
# Health check
curl -H "Authorization: Bearer test" http://localhost:8080/health

# Check all containers
docker compose ps

# Check logs for errors
docker compose logs --tail=100
```

---

## Staging Deployment

### Prerequisites
- EC2 instance (t3.large or larger)
- Ubuntu 20.04 LTS
- Docker and Docker Compose installed
- SSL certificate for HTTPS
- DNS configured

### Deployment Steps

```bash
# SSH into staging server
ssh ubuntu@staging.example.com

# Clone repository
git clone <repo-url>
cd multi-carrier-shipping

# Create production .env file
vi .env

# Pull latest code
git pull origin main

# Build images
docker compose build

# Start services
docker compose up -d

# Setup reverse proxy (Nginx)
sudo cp nginx.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/shipping /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Setup SSL
sudo certbot --nginx -d api.staging.example.com
```

### Staging nginx.conf

```nginx
upstream api_gateway {
    server api-gateway:8080;
}

server {
    listen 443 ssl http2;
    server_name api.staging.example.com;

    ssl_certificate /etc/letsencrypt/live/api.staging.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.staging.example.com/privkey.pem;

    location / {
        proxy_pass http://api_gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

server {
    listen 80;
    server_name api.staging.example.com;
    return 301 https://$server_name$request_uri;
}
```

---

## Production Deployment (Kubernetes)

### Prerequisites
- Kubernetes cluster 1.20+
- kubectl configured
- Docker registry access
- PostgreSQL managed service (AWS RDS, Google Cloud SQL, etc.)
- Managed Kafka (AWS MSK, Confluent Cloud, etc.)

### Step 1: Push Images to Registry

```bash
# Tag images
docker tag multi-carrier-shipping-api-gateway:latest \
    your-registry/api-gateway:1.0.0

# Push to registry
docker push your-registry/api-gateway:1.0.0
docker push your-registry/shipment-service:1.0.0
# ... repeat for all services
```

### Step 2: Create Kubernetes Namespaces

```bash
kubectl create namespace shipping
kubectl create namespace shipping-data
```

### Step 3: Create Secrets

```bash
# Database credentials
kubectl create secret generic db-credentials \
    --from-literal=username=postgres \
    --from-literal=password=<random-password> \
    -n shipping

# API keys
kubectl create secret generic api-keys \
    --from-literal=stripe-key=sk_prod_... \
    --from-literal=dhl-api-key=... \
    -n shipping

# Create ConfigMap for environment variables
kubectl create configmap service-config \
    --from-literal=KAFKA_BROKERS=kafka.shipping:9092 \
    --from-literal=DB_USER=postgres \
    -n shipping
```

### Step 4: Deploy Services

Create `k8s/deployment.yaml`:

```yaml
---
# API Gateway Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: shipping
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: your-registry/api-gateway:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: SHIPMENT_SERVICE_URL
          value: http://shipment-service:8081
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
        envFrom:
        - configMapRef:
            name: service-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5

---
# API Gateway Service
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
  namespace: shipping
spec:
  type: LoadBalancer
  selector:
    app: api-gateway
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP

---
# Shipment Service Deployment (similar structure)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shipment-service
  namespace: shipping
spec:
  replicas: 2
  selector:
    matchLabels:
      app: shipment-service
  template:
    metadata:
      labels:
        app: shipment-service
    spec:
      containers:
      - name: shipment-service
        image: your-registry/shipment-service:1.0.0
        ports:
        - containerPort: 8081
        env:
        - name: PORT
          value: "8081"
        - name: DB_HOST
          value: postgres-shipment.shipping-data
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

Deploy:

```bash
kubectl apply -f k8s/deployment.yaml

# Verify deployment
kubectl get pods -n shipping
kubectl get svc -n shipping
```

### Step 5: Setup Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: shipping-ingress
  namespace: shipping
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.example.com
    secretName: shipping-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              number: 80
```

---

## Database Setup

### PostgreSQL Initialization

```bash
# Connect to database
docker exec postgres-shipment psql -U postgres

# Create database
CREATE DATABASE shipments;

# Create table
CREATE TABLE shipments (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    sender_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    weight DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

# Create indexes
CREATE INDEX idx_shipments_user_id ON shipments(user_id);
CREATE INDEX idx_shipments_status ON shipments(status);
CREATE INDEX idx_shipments_created_at ON shipments(created_at);
```

### Backup & Restore

```bash
# Backup database
docker exec postgres-shipment pg_dump -U postgres shipments > backup.sql

# Restore database
docker exec -i postgres-shipment psql -U postgres shipments < backup.sql
```

---

## Kafka Setup

### Verify Kafka

```bash
# Check Kafka is running
docker exec kafka kafka-broker-api-versions --bootstrap-server kafka:9092

# List topics
docker exec kafka kafka-topics \
    --list \
    --bootstrap-server kafka:9092

# Create topic (if not auto-created)
docker exec kafka kafka-topics \
    --create \
    --topic shipment.created \
    --partitions 3 \
    --replication-factor 1 \
    --bootstrap-server kafka:9092
```

### Monitor Kafka

```bash
# Consumer lag
docker exec kafka kafka-consumer-groups \
    --describe \
    --group notification-service \
    --bootstrap-server kafka:9092

# Tail topic
docker exec kafka kafka-console-consumer \
    --topic shipment.created \
    --from-beginning \
    --bootstrap-server kafka:9092
```

---

## Monitoring & Logging

### View Logs

```bash
# Docker logs
docker compose logs -f

# Specific service
docker logs -f container-name

# Kubernetes logs
kubectl logs -f deployment/api-gateway -n shipping
kubectl logs -f pod/api-gateway-xyz -n shipping
```

### Health Checks

```bash
# API Gateway
curl -H "Authorization: Bearer test" \
    http://localhost:8080/health

# Shipment Service
curl http://localhost:8081/health

# Check service connectivity
docker exec api-gateway curl http://shipment-service:8081/health
```

### Resource Monitoring

```bash
# Docker stats
docker stats

# Kubernetes resource usage
kubectl top nodes -n shipping
kubectl top pods -n shipping
```

---

## Scaling

### Horizontal Scaling (Docker Compose)

```bash
# Scale a service to 3 replicas
docker compose up -d --scale shipment-service=3
```

### Kubernetes Scaling

```bash
# Scale deployment
kubectl scale deployment shipment-service --replicas=5 -n shipping

# Autoscale
kubectl autoscale deployment shipment-service \
    --min=2 --max=10 \
    --cpu-percent=80 \
    -n shipping
```

---

## Security

### Network Security

```bash
# Docker: Services isolated in internal network
docker network inspect multi-carrier-shipping_default

# Kubernetes: Network policies
kubectl apply -f k8s/network-policy.yaml
```

### Secret Management

```bash
# Rotate secrets
kubectl create secret generic db-credentials-v2 \
    --from-literal=password=<new-password> \
    -n shipping

kubectl patch deployment api-gateway \
    -p '{"spec":{"template":{"metadata":{"annotations":{"date":"'$(date +%s)'"}}}}}' \
    -n shipping
```

### SSL/TLS

```bash
# Verify SSL certificate
openssl s_client -connect api.example.com:443

# Check certificate expiration
echo | openssl s_client -servername api.example.com -connect api.example.com:443 2>/dev/null | \
  openssl x509 -noout -dates
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker logs container-name

# Check environment variables
docker exec container-name env | grep SERVICE

# Test database connection
docker exec service-name psql -h postgres-service -U postgres -d dbname -c "SELECT 1"

# Verify network connectivity
docker exec api-gateway curl http://shipment-service:8081/health
```

### High Latency

```bash
# Check database performance
docker exec postgres-shipment psql -U postgres -c "\dt"

# Check Kafka lag
docker exec kafka kafka-consumer-groups --describe --group notification-service --bootstrap-server kafka:9092

# View slow queries
docker logs postgres-shipment | grep "duration"
```

### Crash Loop

```bash
# Check recent logs
docker logs --tail=50 container-name

# Check resource limits
docker stats container-name

# Check port conflicts
docker ps | grep ":8081"
```

---

## Rollback Procedure

### Docker Compose Rollback

```bash
# Stop current version
docker compose down

# Checkout previous version
git checkout v1.0.0

# Rebuild and restart
docker compose build
docker compose up -d
```

### Kubernetes Rollback

```bash
# View rollout history
kubectl rollout history deployment/api-gateway -n shipping

# Rollback to previous version
kubectl rollout undo deployment/api-gateway -n shipping

# Rollback to specific revision
kubectl rollout undo deployment/api-gateway --to-revision=2 -n shipping
```

---

## Performance Tuning

### Database Performance

```sql
-- Create indexes for common queries
CREATE INDEX idx_shipments_user_id ON shipments(user_id);
CREATE INDEX idx_shipments_status ON shipments(status);
CREATE INDEX idx_shipments_created_at ON shipments(created_at DESC);

-- Connection pooling
-- Update connection string with pool parameters
postgresql://user:password@host/db?pool_size=20&max_overflow=40
```

### Kafka Performance

```bash
# Increase partitions
docker exec kafka kafka-topics \
    --alter \
    --topic shipment.created \
    --partitions 6 \
    --bootstrap-server kafka:9092

# Increase consumer instances to match partitions
docker compose up -d --scale notification-service=6
```

