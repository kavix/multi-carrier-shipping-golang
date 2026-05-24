# Label Generation Service

**Port**: 8084  
**Database**: PostgreSQL (labels)  
**Role**: Shipping Label Generation  
**Kafka**: Producer (label.generated)

## Overview

The Label Generation Service creates printable shipping labels in PDF format. Labels contain barcode, tracking information, and carrier-specific formatting.

## Responsibilities

1. **Label Creation**
   - Generate PDF labels
   - Include barcodes
   - Format per carrier requirements

2. **Label Storage**
   - Store generated labels
   - Enable label retrieval
   - Support label reprinting

3. **Label Delivery**
   - Download labels
   - Email labels to users
   - Batch label generation

## API Endpoints

### POST /labels - Generate Label

**Request**:
```json
{
  "shipment_id": "SHIP-001",
  "tracking_number": "1234567890",
  "carrier": "fedex",
  "from_address": "123 Main St, New York, NY",
  "to_address": "456 Oak Ave, Los Angeles, CA",
  "weight": 2.5,
  "reference_number": "ORD-001"
}
```

**Response**: 201 Created
```json
{
  "label_id": "LABEL-001",
  "shipment_id": "SHIP-001",
  "url": "/labels/LABEL-001/download",
  "format": "pdf",
  "size": "4x6",
  "created_at": "2026-05-24T10:30:00Z"
}
```

### GET /labels/:id - Get Label Details

**Response**:
```json
{
  "label_id": "LABEL-001",
  "shipment_id": "SHIP-001",
  "carrier": "fedex",
  "tracking_number": "1234567890",
  "size": "4x6",
  "format": "pdf",
  "status": "ready",
  "created_at": "2026-05-24T10:30:00Z",
  "pdf_url": "/labels/LABEL-001/download"
}
```

### GET /labels/:id/download - Download PDF

**Response**: Binary PDF file
```
Content-Type: application/pdf
Content-Disposition: attachment; filename=LABEL-001.pdf
```

### POST /labels/batch - Batch Generate

**Request**:
```json
{
  "shipment_ids": ["SHIP-001", "SHIP-002", "SHIP-003"]
}
```

**Response**:
```json
{
  "batch_id": "BATCH-001",
  "total": 3,
  "generated": 3,
  "failed": 0,
  "labels": [
    {"shipment_id": "SHIP-001", "label_id": "LABEL-001"},
    {"shipment_id": "SHIP-002", "label_id": "LABEL-002"},
    {"shipment_id": "SHIP-003", "label_id": "LABEL-003"}
  ],
  "url": "/labels/batch/BATCH-001/download"
}
```

## Data Model

### Label Entity

```go
type Label struct {
    ID              string
    ShipmentID      string
    TrackingNumber  string
    Carrier         string
    FromAddress     Address
    ToAddress       Address
    Weight          float64
    ReferenceNumber string
    Size            string    // 4x6, 6x8, etc.
    Format          string    // pdf, zpl, etc.
    Status          string    // ready, processing, failed
    PDFPath         string
    CreatedAt       time.Time
}

type LabelBatch struct {
    ID         string
    Labels     []string  // Label IDs
    Status     string
    TotalCount int
    Created    time.Time
}
```

### Database Schema

```sql
CREATE TABLE labels (
    id VARCHAR(50) PRIMARY KEY,
    shipment_id VARCHAR(50) NOT NULL,
    tracking_number VARCHAR(100) NOT NULL,
    carrier VARCHAR(50) NOT NULL,
    from_address VARCHAR(500),
    to_address VARCHAR(500),
    weight DECIMAL(10, 2),
    reference_number VARCHAR(100),
    size VARCHAR(50),
    format VARCHAR(50),
    status VARCHAR(50),
    pdf_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE label_batches (
    id VARCHAR(50) PRIMARY KEY,
    label_ids TEXT,  -- JSON array
    status VARCHAR(50),
    total_count INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_labels_shipment_id ON labels(shipment_id);
CREATE INDEX idx_labels_tracking_number ON labels(tracking_number);
```

## Label Generation Process

### Flowchart

```
Generate Label Request
    ↓
1. Validate Input
   ├─ Check shipment exists
   ├─ Verify tracking number
   └─ Validate addresses
    ↓
2. Get Carrier Requirements
   ├─ Label size format
   ├─ Barcode type
   ├─ Custom fields
    ↓
3. Create Label Template
   ├─ Set up page size
   ├─ Add logo
   ├─ Add addresses
   ├─ Add shipping info
    ↓
4. Generate Barcode
   ├─ Generate barcode image
   ├─ Embed in template
    ↓
5. Generate PDF
   ├─ Render template
   ├─ Convert to PDF
   ├─ Save to storage
    ↓
6. Publish Event
   └─ label.generated
```

## Label Formats

### 4x6 Label (Thermal Printer)

```
┌─────────────────────────┐
│    FEDEX GROUND         │
│                         │
│ TO: Jane Smith          │
│ 456 Oak Ave             │
│ Los Angeles, CA         │
│                         │
│ [||||||||||||] ← Barcode
│ 1234567890              │
│                         │
│ FROM: John Doe          │
│ 123 Main St             │
│ New York, NY            │
│                         │
│ Weight: 2.5 kg          │
│ Reference: ORD-001      │
└─────────────────────────┘
```

### A4 Label (Standard Printer)

- Full page with addresses
- Large barcode
- Additional information
- Multiple labels per page

## PDF Library

```go
import "github.com/johnnytest-code/gofpdf"

func GenerateLabel(label Label) ([]byte, error) {
    pdf := gofpdf.New("L", "mm", "4x6", "")
    
    // Add carrier logo
    pdf.Image("logos/fedex.png", 10, 10, 30, 0, "", "", "")
    
    // Add addresses
    pdf.SetFont("Arial", "B", 14)
    pdf.Text(50, 30, "TO: "+label.ToAddress.Name)
    pdf.SetFont("Arial", "", 12)
    pdf.Text(50, 40, label.ToAddress.Street)
    pdf.Text(50, 50, label.ToAddress.City+", "+label.ToAddress.State)
    
    // Add barcode
    barcode := generateBarcode(label.TrackingNumber)
    pdf.Image(barcode, 50, 60, 80, 0, "", "", "")
    
    // Generate PDF
    return pdf.Output(io.Writer), nil
}
```

## Kafka Events

### label.generated

**Event**:
```json
{
  "event_type": "label.generated",
  "label_id": "LABEL-001",
  "shipment_id": "SHIP-001",
  "tracking_number": "1234567890",
  "carrier": "fedex",
  "url": "/labels/LABEL-001/download",
  "timestamp": "2026-05-24T10:30:00Z"
}
```

## Error Handling

### Invalid Shipment

```json
{
  "error": "shipment not found",
  "shipment_id": "SHIP-999"
}
```

### Label Generation Failure

```json
{
  "error": "label generation failed",
  "reason": "Invalid barcode format"
}
```

## Configuration

**Environment Variables**:
```
PORT=8084
CARRIER_SERVICE_URL=http://carrier-service:8082
LABEL_STORAGE=s3  # or local, gcs

# S3 Configuration
AWS_S3_BUCKET=shipping-labels
AWS_REGION=us-east-1

# Label Defaults
DEFAULT_LABEL_SIZE=4x6
DEFAULT_FORMAT=pdf
LABEL_RETENTION_DAYS=30
```

## Performance

### Batch Processing

```go
// Generate multiple labels efficiently
func GenerateBatch(shipmentIDs []string) {
    for i := 0; i < len(shipmentIDs); i += 10 {
        batch := shipmentIDs[i : i+10]
        for _, id := range batch {
            go generateLabel(id)  // Parallel
        }
    }
}
```

### Caching

```go
// Cache generated labels for 24 hours
if cached, ok := labelCache[shipmentID]; ok {
    return cached
}

label := generateLabel(shipment)
labelCache[shipmentID] = label
return label
```

## Storage Options

### Local Storage

```
/storage/labels/
├── LABEL-001.pdf
├── LABEL-002.pdf
└── LABEL-003.pdf
```

### S3 Storage

```go
s3Client.PutObject(ctx, 
    bucket: "shipping-labels",
    key: "LABEL-001.pdf",
    body: pdfBytes,
)
```

## Monitoring

### Key Metrics

- Labels generated per day
- Average generation time
- Batch generation statistics
- Success rate
- Storage usage

### Logs

```bash
# View label generation
docker logs label-service | grep "generating"

# View PDF generation
docker logs label-service | grep "pdf"

# View errors
docker logs label-service | grep ERROR
```

## Troubleshooting

### PDFs Not Generating

```bash
# Check font files
docker exec label-service ls /usr/share/fonts/

# Check temp storage
docker exec label-service df -h /tmp

# View errors
docker logs label-service | grep "generation failed"
```

### Barcode Issues

```bash
# Test barcode generation
docker exec label-service /generate-barcode 1234567890

# Check barcode library
docker exec label-service dpkg -l | grep barcode
```

## Future Enhancements

1. **Thermal Printer Integration**: Send labels directly to ZPL printer
2. **Mobile Preview**: QR code for mobile label preview
3. **Address Labels**: Print separate address labels
4. **Return Labels**: Generate return shipping labels
5. **Customs Forms**: Generate customs declaration for international
6. **Label Templates**: Custom label templates per carrier
