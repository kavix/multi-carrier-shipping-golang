import requests
import base64
import time
import logging
from typing import Optional, Dict, Any

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class FedExClient:
    """
    A client to interact with the FedEx API Sandbox environment.
    """
    BASE_URL = "https://apis-sandbox.fedex.com"
    
    def __init__(self, client_id: str, client_secret: str, account_number: str):
        self.client_id = client_id
        self.client_secret = client_secret
        self.account_number = account_number
        self.access_token: Optional[str] = None
        self.token_expiry: float = 0

    def authenticate(self) -> str:
        """
        Phase 1: OAuth 2.0 Authentication
        Exchanges Client ID and Client Secret for a bearer access_token.
        Handles token expiration tracking.
        """
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
        # Set expiry with a 60-second buffer
        self.token_expiry = time.time() + data['expires_in'] - 60
        logger.info("Authentication successful.")
        return self.access_token

    def create_shipment(self) -> Dict[str, Any]:
        """
        Phase 2: Create Test Shipment
        Creates a FEDEX_GROUND shipment from Memphis, TN to Atlanta, GA.
        """
        token = self.authenticate()
        logger.info("Creating test shipment...")
        url = f"{self.BASE_URL}/ship/v1/shipments"
        
        headers = {
            'Content-Type': "application/json",
            'Authorization': f"Bearer {token}"
        }

        # Payload based on FedEx Ship API specification
        payload = {
            "labelResponseOptions": "LABEL",
            "requestedShipment": {
                "shipper": {
                    "address": {
                        "streetLines": ["No 123, Galle Road"],
                        "city": "Colombo",
                        "postalCode": "00300",
                        "countryCode": "LK"
                    },
                    "contact": {
                        "personName": "Kavindu Sachinthe",
                        "phoneNumber": "0112345678",
                        "companyName": "Lanka Exports"
                    }
                },
                "recipients": [{
                    "address": {
                        "streetLines": ["123 Receiver Lane"],
                        "city": "Atlanta",
                        "stateOrProvinceCode": "GA",
                        "postalCode": "30303",
                        "countryCode": "US"
                    },
                    "contact": {
                        "personName": "US Receiver",
                        "phoneNumber": "0987654321",
                        "companyName": "US Import Co"
                    }
                }],
                "shipDatestamp": time.strftime("%Y-%m-%d"),
                "serviceType": "INTERNATIONAL_PRIORITY",
                "packagingType": "YOUR_PACKAGING",
                "pickupType": "USE_SCHEDULED_PICKUP",
                "customsClearanceDetail": {
                    "dutiesPayment": {
                        "paymentType": "SENDER",
                        "payor": {
                            "responsibleParty": {
                                "accountNumber": {"value": self.account_number}
                            }
                        }
                    },
                    "documentContent": "NON_DOCUMENTS",
                    "totalCustomsValue": {
                        "amount": 100.0,
                        "currency": "USD"
                    },
                    "commodities": [{
                        "description": "Sample Items from Sri Lanka",
                        "countryOfManufacture": "LK",
                        "quantity": 1,
                        "quantityUnits": "PCS",
                        "unitPrice": {"amount": 100.0, "currency": "USD"},
                        "customsValue": {"amount": 100.0, "currency": "USD"},
                        "weight": {"units": "LB", "value": 2.0}
                    }]
                },
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
                            "value": 2.0
                        }
                    }
                ]
            },
            "accountNumber": {"value": self.account_number}
        }

        response = requests.post(url, json=payload, headers=headers)
        if response.status_code != 200:
            logger.error(f"Failed to create shipment. Status: {response.status_code}, Response: {response.text}")
        response.raise_for_status()
        
        return response.json()

    def process_shipment_response(self, response_data: Dict[str, Any]):
        """
        Phase 3: Label Decoding & Local Storage
        Extracts masterTrackingNumber and encodedLabel, then saves as label.pdf.
        """
        try:
            output = response_data.get('output', {})
            transaction_shipments = output.get('transactionShipments', [])
            
            if not transaction_shipments:
                logger.error("No transaction shipments found in response.")
                return

            shipment = transaction_shipments[0]
            master_tracking_number = shipment.get('masterTrackingNumber')
            
            piece_responses = shipment.get('pieceResponses', [])
            if not piece_responses:
                logger.error("No piece responses found.")
                return
            
            # The encodedLabel is inside pieceResponses -> packageDocuments
            package_documents = piece_responses[0].get('packageDocuments', [])
            if not package_documents:
                logger.error("No package documents found.")
                return

            encoded_label = package_documents[0].get('encodedLabel')
            
            if not encoded_label:
                logger.error("Encoded label not found in response.")
                return

            logger.info(f"Master Tracking Number: {master_tracking_number}")
            
            # Decode and save to label.pdf
            label_data = base64.b64decode(encoded_label)
            with open("label.pdf", "wb") as f:
                f.write(label_data)
            
            logger.info("Label successfully decoded and saved as label.pdf")

        except Exception as e:
            logger.error(f"Error processing shipment response: {e}")
            raise

if __name__ == "__main__":
    # Credentials from requirements
    CLIENT_ID = "l7d79cfde66a33429990b1e760138fb300"
    CLIENT_SECRET = "f43a6307ecc14f57b20b373186c6f3a9"
    ACCOUNT_NUMBER = "740561073"

    client = FedExClient(CLIENT_ID, CLIENT_SECRET, ACCOUNT_NUMBER)
    
    try:
        # Run the shipping lifecycle
        shipment_data = client.create_shipment()
        client.process_shipment_response(shipment_data)
        logger.info("FedEx Shipping Lifecycle completed successfully.")
    except requests.exceptions.HTTPError as http_err:
        logger.error(f"HTTP error occurred: {http_err}")
    except Exception as err:
        logger.error(f"An unexpected error occurred: {err}")
