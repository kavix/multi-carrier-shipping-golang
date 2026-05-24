.PHONY: all build up down logs test migrate dev

all: build up

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down -v

logs:
	docker compose logs -f

dev:
	@echo "Starting infrastructure (Kafka + PostgreSQL)..."
	docker compose up -d zookeeper kafka postgres-shipment postgres-carrier postgres-rate postgres-label postgres-tracking postgres-address postgres-billing postgres-return
	@echo "Waiting 15s for services..."
	@sleep 15
	@echo "Run migrations, then start services with 'go run ./cmd' in each service folder"

migrate:
	@echo "Applying migrations..."
	@psql postgres://postgres:postgres@localhost:5431/shipments -f shipment-service/migrations/001_create_shipments.sql 2>/dev/null || echo "  shipment migration failed or already applied"
	@psql postgres://postgres:postgres@localhost:5432/carriers -f carrier-integration-service/migrations/001_create_carriers.sql 2>/dev/null || echo "  carrier migration failed or already applied"
	@psql postgres://postgres:postgres@localhost:5433/rates -f rate-comparison-service/migrations/001_create_rates.sql 2>/dev/null || echo "  rate migration failed or already applied"
	@psql postgres://postgres:postgres@localhost:5434/labels -f label-generation-service/migrations/001_create_labels.sql 2>/dev/null || echo "  label migration failed or already applied"
	@psql postgres://postgres:postgres@localhost:5435/tracking -f tracking-service/migrations/001_create_tracking.sql 2>/dev/null || echo "  tracking migration failed or already applied"
	@psql postgres://postgres:postgres@localhost:5436/addresses -f address-validation-service/migrations/001_create_addresses.sql 2>/dev/null || echo "  address migration failed or already applied"
	@psql postgres://postgres:postgres@localhost:5437/billing -f billing-service/migrations/001_create_billing.sql 2>/dev/null || echo "  billing migration failed or already applied"
	@psql postgres://postgres:postgres@localhost:5438/returns -f return-service/migrations/001_create_returns.sql 2>/dev/null || echo "  return migration failed or already applied"

test-flow:
	@echo "=== Step 1: Validate Address ==="
	curl -s -X POST http://localhost:8080/addresses/validate \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer user-123" \
		-d '{"address":"123 Main St, New York, NY"}' | python3 -m json.tool
	@echo ""
	@echo "=== Step 2: Create Shipment ==="
	curl -s -X POST http://localhost:8080/shipments \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer user-123" \
		-d '{"sender_name":"John","sender_address":"123 Main St, New York, NY","receiver_name":"Jane","receiver_address":"456 Oak Ave, Los Angeles, CA","weight":2.5,"dimensions":"10x10x10","carrier":"dhl","service_type":"express"}' | python3 -m json.tool
	@echo ""
	@echo "=== Step 3: Compare Rates ==="
	curl -s -X POST http://localhost:8080/rates/compare \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer user-123" \
		-d '{"shipment_id":"SHIP-123","from":"New York, NY","to":"Los Angeles, CA","weight":2.5}' | python3 -m json.tool
	@echo ""
	@echo "=== Step 4: Get Carrier Rates (Direct) ==="
	curl -s "http://localhost:8080/carriers/rates?from=New+York&to=Los+Angeles&weight=2.5" \
		-H "Authorization: Bearer user-123" | python3 -m json.tool
	@echo ""
	@echo "=== Step 5: Get Pickup Locations ==="
	curl -s "http://localhost:8080/addresses/pickup-locations?address=New+York&carrier=dhl&limit=5" \
		-H "Authorization: Bearer user-123" | python3 -m json.tool
	@echo ""
	@echo "=== Step 6: Generate Label ==="
	curl -s -X POST http://localhost:8080/labels \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer user-123" \
		-d '{"shipment_id":"SHIP-123","carrier":"dhl"}' | python3 -m json.tool

status:
	@echo "=== Service Status ==="
	@docker compose ps
