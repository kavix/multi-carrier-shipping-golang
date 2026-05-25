#!/bin/bash

# Default count to 5 if not provided
COUNT=${1:-5}

for i in $(seq 1 $COUNT);
do
  SHIPMENT_ID=$(uuidgen)
  USER_ID="test-user-$(date +%s)"
  SENDER_NAME="Sender_$((RANDOM % 100 + 1))"
  SENDER_ADDRESS="123 Main St, Anytown, USA"
  SENDER_EMAIL="sender$i@example.com"
  RECEIVER_NAME="Receiver_$((RANDOM % 100 + 1))"
  RECEIVER_ADDRESS="456 Oak Ave, Otherville, USA"
  RECEIVER_EMAIL="receiver$i@example.com"
  
  WEIGHT_INT=$((RANDOM % 100 + 1))
  WEIGHT_DEC=$((RANDOM % 100))
  WEIGHT=$(printf "%.2f" "$WEIGHT_INT.$WEIGHT_DEC")

  DIMENSIONS="10x10x10"
  CARRIER="UPS"
  SERVICE_TYPE="Express"
  STATUS="pending"
  TRACKING_NUMBER="TN-$SHIPMENT_ID"

  COST_INT=$((RANDOM % 50 + 1))
  COST_DEC=$((RANDOM % 100))
  COST=$(printf "%.2f" "$COST_INT.$COST_DEC")

  JSON_PAYLOAD=$(cat <<EOF
{
  "sender_name": "$SENDER_NAME",
  "sender_address": "$SENDER_ADDRESS",
  "sender_email": "$SENDER_EMAIL",
  "receiver_name": "$RECEIVER_NAME",
  "receiver_address": "$RECEIVER_ADDRESS",
  "receiver_email": "$RECEIVER_EMAIL",
  "weight": $WEIGHT,
  "dimensions": "$DIMENSIONS",
  "carrier": "$CARRIER",
  "service_type": "$SERVICE_TYPE"
}
EOF
)

  echo "Adding shipment $i..."
  curl -sS -X POST \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer test-token" \
       -d "$JSON_PAYLOAD" \
       http://localhost:8080/shipments | grep -o '"id":"[^"]*"'
  echo -e "\n"
done
