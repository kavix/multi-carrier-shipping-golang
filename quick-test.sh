#!/bin/bash
echo "Testing All Services..."
echo ""

# Health checks
echo "=== HEALTH CHECKS ==="
echo "API Gateway:"
curl -s -H "Authorization: Bearer test-token" http://localhost:8080/health | jq . || echo "Failed"

echo ""
echo "Shipment Service:"
curl -s http://localhost:8081/health | jq . || echo "Failed"

echo ""
echo "Carrier Service:"
curl -s http://localhost:8082/health | jq . || echo "Failed"

echo ""
echo "Rate Service:"
curl -s http://localhost:8083/health | jq . || echo "Failed"

echo ""
echo "Label Service:"
curl -s http://localhost:8084/health | jq . || echo "Failed"

echo ""
echo "Tracking Service:"
curl -s http://localhost:8085/health | jq . || echo "Failed"

echo ""
echo "Address Service:"
curl -s http://localhost:8086/health | jq . || echo "Failed"

echo ""
echo "Billing Service:"
curl -s http://localhost:8087/health | jq . || echo "Failed"

echo ""
echo "Return Service:"
curl -s http://localhost:8088/health | jq . || echo "Failed"
