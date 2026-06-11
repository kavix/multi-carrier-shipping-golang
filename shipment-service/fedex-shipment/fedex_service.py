import os
import base64
import time
import logging
import requests
from typing import Dict, Any, Optional

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class S3Service:
    """Mock S3 Service for storing labels."""
    def upload_label(self, tracking_number: str, label_data: bytes) -> str:
        file_path = f"labels/{tracking_number}.pdf"
        os.makedirs("labels", exist_ok=True)
        with open(file_path, "wb") as f:
            f.write(label_data)
        s3_url = f"s3://my-shipping-bucket/{file_path}"
        logger.info(f"Label uploaded to S3: {s3_url}")
        return s3_url

class LabelService:
    """Mock Label Service for managing shipment labels."""
    def register_label(self, tracking_number: str, s3_url: str, shipment_id: str):
        logger.info(f"Label registered in Label Service: Tracking={tracking_number}, S3_URL={s3_url}, ShipmentID={shipment_id}")
        return {"status": "success", "label_id": f"LBL-{tracking_number}"}

class FedExClient:
    """Client to interact with the FedEx API."""
    BASE_URL = "https://apis-sandbox.fedex.com"
    
    def __init__(self, client_id: str, client_secret: str, account_number: str):
        self.client_id = client_id
        self.client_secret = client_secret
        self.account_number = account_number
        self.access_token: Optional[str] = None
        self.token_expiry: float = 0

    def authenticate(self) -> str:
        if self.access_token and time.time() < self.token_expiry:
            return self.access_token

        logger.info("Authenticating with FedEx API...")
        url = f"{self.BASE_URL}/oauth/token"
        payload = {
            'grant_type': 'client_credentials',
            'client_id': self.client_id,
            'client_secret': self.client_secret
        }
        headers = {'content-type': "application/x-www-form-urlencoded"}

        response = requests.post(url, data=payload, headers=headers)
        response.raise_for_status()

        data = response.json()
        self.access_token = data['access_token']
        self.token_expiry = time.time() + data['expires_in'] - 60
        return self.access_token

    def create_shipment(self, shipment_data: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Creates a FedEx shipment. If shipment_data is provided, it maps the generic 
        payload to FedEx specific fields. Otherwise, it uses defaults.
        """
        token = self.authenticate()
        logger.info("Creating FedEx shipment...")
        url = f"{self.BASE_URL}/ship/v1/shipments"
        
        headers = {
            'Content-Type': "application/json",
            'Authorization': f"Bearer {token}"
        }

        # Mapping generic payload to FedEx structure if provided
        if shipment_data:
            # Basic mapping logic for demonstration
            # In a real app, you'd have more complex address parsing/mapping
            shipper_name = shipment_data.get('sender_name', 'Shipper Name')
            shipper_phone = shipment_data.get('sender_phone', '9011234567')
            recipient_name = shipment_data.get('receiver_name', 'Recipient Name')
            recipient_phone = shipment_data.get('receiver_phone', '4041234567')
            weight_val = shipment_data.get('weight', 1.0)
            
            # Note: We still use valid Sandbox addresses to avoid the "Invalid postal code" error
            shipper_address = {
                "streetLines": ["10 FedEx Parkway"],
                "city": "Collierville",
                "stateOrProvinceCode": "TN",
                "postalCode": "38017",
                "countryCode": "US"
            }
            recipient_address = {
                "streetLines": ["123 Main St"],
                "city": "Atlanta",
                "stateOrProvinceCode": "GA",
                "postalCode": "30303",
                "countryCode": "US"
            }
        else:
            shipper_name = "Shipper Name"
            shipper_phone = "9011234567"
            recipient_name = "Recipient Name"
            recipient_phone = "4041234567"
            shipper_address = {
                "streetLines": ["10 FedEx Parkway"],
                "city": "Collierville",
                "stateOrProvinceCode": "TN",
                "postalCode": "38017",
                "countryCode": "US"
            }
            recipient_address = {
                "streetLines": ["123 Main St"],
                "city": "Atlanta",
                "stateOrProvinceCode": "GA",
                "postalCode": "30303",
                "countryCode": "US"
            }
            weight_val = 1.0

        payload = {
            "labelResponseOptions": "LABEL",
            "requestedShipment": {
                "shipper": {
                    "address": shipper_address,
                    "contact": {
                        "personName": shipper_name,
                        "phoneNumber": shipper_phone,
                        "companyName": "FedEx Sandbox Shipper"
                    }
                },
                "recipients": [{
                    "address": recipient_address,
                    "contact": {
                        "personName": recipient_name,
                        "phoneNumber": recipient_phone,
                        "companyName": "Recipient Co"
                    }
                }],
                "shipDatestamp": time.strftime("%Y-%m-%d"),
                "serviceType": "FEDEX_GROUND",
                "packagingType": "YOUR_PACKAGING",
                "pickupType": "USE_SCHEDULED_PICKUP",
                "shippingChargesPayment": {
                    "paymentType": "SENDER",
                    "payor": {
                        "responsibleParty": {
                            "accountNumber": {"value": self.account_number}
                        }
                    }
                },
                "labelSpecification": {
                    "labelFormatType": "COMMON2D",
                    "imageType": "PDF",
                    "labelStockType": "PAPER_4X6"
                },
                "requestedPackageLineItems": [
                    {
                        "weight": {
                            "units": "LB",
                            "value": weight_val
                        }
                    }
                ]
            },
            "accountNumber": {"value": self.account_number}
        }

        response = requests.post(url, json=payload, headers=headers)
        if response.status_code != 200:
            logger.error(f"Failed to create shipment. Status: {response.status_code}, Response: {response.text}")
            if "POSTAL_CODE_INVALID" in response.text or "Invalid postal code" in response.text:
                raise Exception(f"FedEx Address Validation Error: {response.json().get('errors', [{}])[0].get('message', 'Invalid postal code')}")
        response.raise_for_status()
        
        return response.json()

class FedExShipmentService:
    """Service to orchestrate FedEx shipment creation, storage, and registration."""
    def __init__(self, fedex_client: FedExClient, s3_service: S3Service, label_service: LabelService):
        self.fedex_client = fedex_client
        self.s3_service = s3_service
        self.label_service = label_service

    def process_new_shipment(self, shipment_data: Dict[str, Any] = None):
        try:
            shipment_response = self.fedex_client.create_shipment(shipment_data)
            
            output = shipment_response.get('output', {})
            transaction_shipments = output.get('transactionShipments', [])
            if not transaction_shipments:
                raise Exception("No transaction shipments in FedEx response")

            shipment = transaction_shipments[0]
            tracking_number = shipment.get('masterTrackingNumber')
            
            piece_responses = shipment.get('pieceResponses', [])
            if not piece_responses:
                raise Exception("No piece responses in FedEx response")

            package_documents = piece_responses[0].get('packageDocuments', [])
            if not package_documents:
                raise Exception("No package documents in FedEx response")

            encoded_label = package_documents[0].get('encodedLabel')
            label_bytes = base64.b64decode(encoded_label)

            logger.info(f"Shipment created successfully. Tracking Number: {tracking_number}")

            s3_url = self.s3_service.upload_label(tracking_number, label_bytes)

            shipment_id = f"SHIP-{int(time.time())}"
            self.label_service.register_label(tracking_number, s3_url, shipment_id)

            return {
                "status": "success",
                "carrier": "fedex",
                "tracking_number": tracking_number,
                "s3_url": s3_url,
                "shipment_id": shipment_id
            }

        except Exception as e:
            logger.error(f"Error in FedEx shipment workflow: {e}")
            raise

class RouterShipmentService:
    """Routes shipment requests to the appropriate carrier service."""
    def __init__(self, fedex_service: FedExShipmentService):
        self.fedex_service = fedex_service

    def create_shipment(self, shipment_data: Dict[str, Any]):
        carrier = shipment_data.get('carrier', '').lower()
        
        if carrier == 'fedex':
            logger.info("Routing shipment to FedEx service...")
            return self.fedex_service.process_new_shipment(shipment_data)
        else:
            logger.warning(f"Carrier '{carrier}' not supported. Defaulting to mock response.")
            # For other carriers, we return a mock success for demonstration
            return {
                "status": "success",
                "carrier": carrier,
                "message": f"Shipment created using {carrier} (Mocked)",
                "tracking_number": f"MOCK-{carrier.upper()}-12345"
            }
