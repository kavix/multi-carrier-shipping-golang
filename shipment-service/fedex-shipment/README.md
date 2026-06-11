# Fast FedEx Shipment CLI

This project provides a clean, modular Python CLI to handle the FedEx shipping lifecycle, including S3 storage and Label Service registration.

## Features
- **OAuth 2.0 Authentication**: Automatic token management.
- **CLI Interface**: Easy commands like `./fedex create shipment`.
- **S3 Integration**: Automatically uploads labels to a (mock) S3 bucket.
- **Label Service Integration**: Registers shipments with a central Label Service.
- **Robust Address Validation**: Uses standard sandbox addresses to avoid postal code errors.

## Setup

1. **Installation**:
   ```bash
   pip install -r requirements.txt
   chmod +x fedex
   ```

2. **Usage**:
   ```bash
   ./fedex create shipment
   ```

## Output
- **S3**: Labels are stored in the `labels/` directory (mocking S3).
- **Label Service**: Logs registration of the tracking number and S3 URL.
- **Summary**: Prints tracking number, S3 URL, and Shipment ID.
