#!/bin/bash
echo "Testing All Services..."
echo ""

# Health checks
echo "=== HEALTH CHECKS ==="
echo "API Gateway:"
curl -s -H "Authorization: Bearer test-token" http://localhost:8080/health | jq . || echo "Failed"

echo ""
echo "Shipment Service:"
curl -s http://localhost:8001/health | jq . || echo "Failed"

echo ""
echo "Carrier Service:"
curl -s http://localhost:8002/health | jq . || echo "Failed"

echo ""
echo "Rate Service:"
curl -s http://localhost:8003/health | jq . || echo "Failed"

echo ""
echo "Label Service:"
curl -s http://localhost:8004/health | jq . || echo "Failed"

echo ""
echo "Tracking Service:"
curl -s http://localhost:8005/health | jq . || echo "Failed"

echo ""
echo "Address Service:"
curl -s http://localhost:8006/health | jq . || echo "Failed"

echo ""
echo "Billing Service:"
curl -s http://localhost:8007/health | jq . || echo "Failed"

echo ""
echo "Return Service:"
curl -s http://localhost:8008/health | jq . || echo "Failed"
