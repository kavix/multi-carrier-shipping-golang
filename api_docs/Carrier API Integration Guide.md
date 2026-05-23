# Carrier API Integration Guide

This document explains how the repository integrates with DHL and FedEx sandbox APIs through the `label` microservice.

## FedEx integration

### Key endpoints used
- `POST https://apis-sandbox.fedex.com/oauth/token`
  - obtains OAuth bearer token
  - request uses `client_id` and `client_secret`
- `POST https://apis-sandbox.fedex.com/location/v1/locations`
  - searches for nearby FedEx locations by address
- `POST https://apis-sandbox.fedex.com/country/v1/postal/validate`
  - validates postal codes and returns cleaned postal details
- `POST https://apis-sandbox.fedex.com/rate/v1/rates/quotes`
  - retrieves rate quotes and transit time estimates

### Code locations
- `internal/label/fedex_client.go`
  - `SearchLocations()`
  - `getAccessToken()`
  - `ValidatePostalCode()`
  - `GetRatesAndTransitTimes()`
  - `GenerateLabel()` (simulated)
- `cmd/label_service/main.go`
  - initializes FedEx client with sandbox credentials
- `internal/label/service.go`
  - enriches label creation response with FedEx postal validation, rate quotes, and drop-off locations

## DHL integration

### Key endpoint used
- `GET https://api-sandbox.dhl.com/location-finder/v1/find-by-address`
  - searches DHL drop-off/pick-up points by address
  - requires `DHL-API-Key` header

### Code locations
- `internal/label/dhl_client.go`
  - `SearchLocations()`
  - `GenerateLabel()` (simulated)
- `cmd/label_service/main.go`
  - initializes DHL client using environment variable `DHL_API_KEY`
- `internal/label/service.go`
  - uses DHL location search as drop-off location enrichment for DHL label creation

## Service wiring

- `cmd/label_service/main.go`
  - configures `LabelService` on port `8082`
  - uses `MultiCarrierClient` to dispatch FedEx or DHL requests
- `cmd/shipment_service/main.go`
  - configured to call the label service at `http://localhost:8082`
  - creates shipments and requests labels from the label service

## Response formats

### Label Service `POST /api/v1/labels`

Response structure:
```json
{
  "label": { ... },
  "carrier": "FEDEX",
  "quote_date": "...",
  "origin_postal": { ... },
  "destination_postal": { ... },
  "rate_quotes": [ ... ],
  "origin_drop_off_locations": [ ... ],
  "destination_drop_off_locations": [ ... ]
}
```

### Shipment Service `POST /api/v1/shipments`

Response structure:
```json
{
  "shipment": { ... },
  "label": { ... }
}
```

## Notes

- DHL/FedEx location search is live through sandbox endpoints.
- Label generation is currently simulated for both carriers.
- Ensure `DHL_API_KEY` is set for DHL location search to work.
- The FedEx sandbox client uses hard-coded sandbox credentials in `cmd/label_service/main.go`.


Inspecting the shipment service startup and handler/service flow to connect carrier responses end-to-end.

Searched for regex `CreateLabel\(|LabelCreateResponse|origin_drop_off_locations|destination_drop_off_locations|RateQuotes|PostalInfo|SearchLocations\(` (`**/*.go`), 34 results

Read [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/cmd/shipment_service/main.go)

Read [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/handler.go#1-1), lines 1 to 260

Read [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/service.go#1-1), lines 1 to 320

Searched for regex `type Shipment struct|Shipment struct \{|type Label struct` (`**/internal/shipment/*.go`), 2 results

Read [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/domain.go)

Searched for regex `func ConfigureRouter|ConfigureRouter\(` (`**/internal/shipment/*.go`), 1 result

Created 2 todos

Read [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/handler.go#260-260), lines 260 to 310

Updating label creation enrichments and shipment-service parsing so the services can correctly exchange carrier responses.

Replacing 77 lines with 82 lines in [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/service.go), [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/label/service.go)


```

```

```

```


Read [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/service.go#130-130), lines 130 to 215

Replacing 9 lines with 10 lines in [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/service.go)


```

```


Read [](file:///Users/kavindus/Projects/multi-carrier-shipping-golang/internal/shipment/service.go#215-215), lines 215 to 320

Validating the integration changes by compiling the impacted packages and tests.

Ran terminal command:  go test ./internal/shipment ./internal/label

## ✅ Integration update completed

I integrated the carrier API response flow between the label service and shipment service and validated the code.

### What changed

- service.go
  - `CreateLabel()` now populates `origin_drop_off_locations` and `destination_drop_off_locations` for all supported carriers, including DHL.
  - FedEx still receives rich extra enrichment:
    - `origin_postal`
    - `destination_postal`
    - `rate_quotes`
    - `origin_drop_off_locations`
    - `destination_drop_off_locations`

- service.go
  - `CreateShipment()` now decodes the Label Service response as:
    - `{ "label": { ... } }`
  - This fixes the previous mismatch where Shipment Service expected a flat label object.

### Validation

- `go test ./internal/shipment ./internal/label` ✅

---

## Expected response shapes

### Label Service
`POST /api/v1/labels`

Example response:

```json
{
  "label": {
    "id": "lbl-123456789",
    "shipment_id": "shipment-123",
    "tracking_number": "FTX012345678",
    "label_url": "https://fedex-sandbox-labels.s3.amazonaws.com/labels/FTX012345678.pdf",
    "status": "ACTIVE",
    "created_at": "2026-05-23T12:34:56Z"
  },
  "carrier": "FEDEX",
  "quote_date": "2026-05-23T12:00:00Z",
  "origin_postal": {
    "postal_code": "90001",
    "city_first_initials": "L",
    "state_or_province_code": "CA",
    "country_code": "US"
  },
  "destination_postal": {
    "postal_code": "10001",
    "city_first_initials": "N",
    "state_or_province_code": "NY",
    "country_code": "US"
  },
  "rate_quotes": [
    {
      "service_type": "FEDEX_GROUND",
      "service_name": "FedEx Ground",
      "currency": "USD",
      "base_charge": 23.50,
      "total_net_charge": 27.79,
      "fuel_surcharge_percent": 0,
      "total_surcharges": 4.29,
      "transit_time": "TWO_DAYS",
      "delivery_day": "FRI",
      "commit_date_time": "2026-05-25T18:00:00",
      "delivery_postal_code": "10001",
      "rate_zone": "8"
    }
  ],
  "origin_drop_off_locations": [
    {
      "carrier": "FedEx",
      "distance": 12.5,
      "units": "MI",
      "name": "FedEx Location",
      "street_lines": ["100 FedEx Pkwy"],
      "city": "Memphis",
      "state_or_province_code": "TN",
      "postal_code": "38116",
      "country_code": "US"
    }
  ],
  "destination_drop_off_locations": [
    {
      "carrier": "FedEx",
      "distance": 8.2,
      "units": "MI",
      "name": "FedEx Location",
      "street_lines": ["200 Main St"],
      "city": "Nashville",
      "state_or_province_code": "TN",
      "postal_code": "37214",
      "country_code": "US"
    }
  ]
}
```

For DHL, you should see at least the DHL label plus drop-off locations:

```json
{
  "label": { ... },
  "carrier": "DHL",
  "origin_drop_off_locations": [
    {
      "carrier": "DHL",
      "location_type": "servicepoint",
      "distance": 3.4,
      "units": "km",
      "name": "DHL Service Point",
      "street_lines": ["12 Queen St"],
      "city": "London",
      "postal_code": "SW1A 1AA",
      "country_code": "GB",
      "opening_hours": [
        { "day_of_week": "Monday", "opens": "09:00", "closes": "18:00" }
      ]
    }
  ]
}
```

### Shipment Service
`POST /api/v1/shipments`

Example response:

```json
{
  "shipment": {
    "id": "UUID",
    "carrier": "FedEx",
    "tracking_number": "FTX012345678",
    "weight": 4.5,
    "origin": "Los Angeles, CA",
    "destination": "New York, NY",
    "status": "CREATED",
    "username": "user@example.com",
    "email": "recipient@example.com",
    "created_at": "2026-05-23T12:34:56Z",
    "updated_at": "2026-05-23T12:35:02Z"
  },
  "label": {
    "id": "lbl-123456789",
    "shipment_id": "UUID",
    "tracking_number": "FTX012345678",
    "label_url": "https://fedex-sandbox-labels.s3.amazonaws.com/labels/FTX012345678.pdf",
    "status": "ACTIVE",
    "created_at": "2026-05-23T12:34:56Z"
  }
}
```

---

## Summary

- Label Service now returns rich carrier response data for FedEx and DHL.
- Shipment Service now correctly consumes that response.
- The integration is validated by tests.

If you want, I can also add a small example `curl` command to call each endpoint directly.

Completed: *Integrate DHL and FedEx responses* (1/2)

Made changes.