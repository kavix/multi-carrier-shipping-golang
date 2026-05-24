#!/bin/bash
# 1. Create Shipment
echo "--- POST /shipments ---"
SHIPMENT_RESP=$(curl -s -X POST http://localhost:8080/shipments \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "sender_name": "Alice Smith",
    "sender_address": "123 Maple St, Seattle, WA 98101",
    "sender_phone": "555-0123",
    "sender_email": "alice@example.com",
    "receiver_name": "Bob Jones",
    "receiver_address": "456 Pine Rd, Miami, FL 33101",
    "receiver_phone": "555-0456",
    "receiver_email": "bob@example.com",
    "weight": 1.2,
    "dimensions": "8x6x4",
    "description": "Books",
    "carrier": "ups",
    "service_type": "ground"
  }')
echo "$SHIPMENT_RESP" | jq .
SHIPMENT_ID=$(echo "$SHIPMENT_RESP" | jq -r .id)

# 1b. Register Carrier
echo -e "\n--- POST /carriers ---"
curl -s -X POST http://localhost:8080/carriers \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DHL Express",
    "code": "dhl",
    "api_key": "test-key",
    "api_secret": "test-secret",
    "base_url": "http://simulated-dhl"
  }' | jq .

# 2. Compare Rates
echo -e "\n--- POST /rates/compare ---"
curl -s -X POST http://localhost:8080/rates/compare \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "shipment_id": "'$SHIPMENT_ID'",
    "from": "Seattle, WA",
    "to": "Miami, FL",
    "weight": 1.2
  }' | jq .

# 3. Generate Label
echo -e "\n--- POST /labels ---"
curl -s -X POST http://localhost:8080/labels \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "shipment_id": "'$SHIPMENT_ID'",
    "carrier": "ups",
    "format": "pdf"
  }' | jq .

# 4. Validate Address
echo -e "\n--- POST /addresses/validate ---"
curl -s -X POST http://localhost:8080/addresses/validate \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "address": "123 Maple St, Seattle, WA 98101"
  }' | jq .

# 5. Create Invoice
echo -e "\n--- POST /billing/invoices ---"
INV_RESP=$(curl -s -X POST http://localhost:8080/billing/invoices \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "shipment_id": "'$SHIPMENT_ID'",
    "user_id": "test-user-999",
    "amount": 24.50,
    "description": "Shipping charges"
  }')
echo "$INV_RESP" | jq .
INV_ID=$(echo "$INV_RESP" | jq -r .id)

# 6. Process Payment
echo -e "\n--- POST /billing/payments ---"
curl -s -X POST http://localhost:8080/billing/payments \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_id": "'$INV_ID'",
    "method": "credit_card"
  }' | jq .

# 7. Request Return
echo -e "\n--- POST /returns ---"
curl -s -X POST http://localhost:8080/returns \
  -H "Authorization: Bearer test-user-999" \
  -H "Content-Type: application/json" \
  -d '{
    "shipment_id": "'$SHIPMENT_ID'",
    "reason": "Damaged on arrival"
  }' | jq .
