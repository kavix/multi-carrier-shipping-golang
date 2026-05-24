import React, { useState } from 'react'
import ApiForm from './components/ApiForm'

const DEFAULT_API = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export default function App() {
  const [baseUrl, setBaseUrl] = useState(DEFAULT_API)
  const [token, setToken] = useState('')

  return (
    <div className="container">
      <header>
        <h1>Multi-Carrier Shipping — Frontend</h1>
        <div className="controls">
          <label>
            API Base URL:
            <input value={baseUrl} onChange={e => setBaseUrl(e.target.value)} />
          </label>
          <label>
            Authorization Token (optional):
            <input value={token} onChange={e => setToken(e.target.value)} placeholder="Bearer ..." />
          </label>
        </div>
      </header>

      <main>
        <section>
          <h2>Shipments</h2>
          <ApiForm
            title="List Shipments"
            method="GET"
            path="/shipments"
            baseUrl={baseUrl}
            token={token}
          />
          <ApiForm
            title="Create Shipment"
            method="POST"
            path="/shipments"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{
              sender_name: 'John Doe',
              sender_address: '123 Main St, New York, NY 10001',
              sender_phone: '+1-555-0100',
              sender_email: 'john@example.com',
              receiver_name: 'Jane Smith',
              receiver_address: '456 Oak Ave, Los Angeles, CA 90001',
              receiver_phone: '+1-555-0200',
              receiver_email: 'jane@example.com',
              weight: 2.5,
              dimensions: '10x10x10',
              description: 'Electronics package',
              carrier: 'dhl',
              service_type: 'express',
            }}
          />
        </section>

        <section>
          <h2>Rates</h2>
          <ApiForm
            title="Compare Rates"
            method="POST"
            path="/rates/compare"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ shipment_id: 'SHIP-001', from: 'New York, NY 10001', to: 'Los Angeles, CA 90001', weight: 2.5 }}
          />
          <ApiForm
            title="Get Carrier Rates (Query)"
            method="GET"
            path="/carriers/rates?from=New+York&to=Los+Angeles&weight=2.5"
            baseUrl={baseUrl}
            token={token}
          />
        </section>

        <section>
          <h2>Labels</h2>
          <ApiForm
            title="Generate Label"
            method="POST"
            path="/labels"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ shipment_id: 'SHIP-001', carrier: 'dhl', format: 'pdf' }}
          />
        </section>

        <section>
          <h2>Tracking</h2>
          <ApiForm
            title="Get Tracking History"
            method="GET"
            path="/tracking/SHIP-001"
            baseUrl={baseUrl}
            token={token}
          />
          <ApiForm
            title="Carrier Tracking (Query)"
            method="GET"
            path="/carriers/tracking?carrier=dhl&tracking_number=1234567890"
            baseUrl={baseUrl}
            token={token}
          />
        </section>

        <section>
          <h2>Address</h2>
          <ApiForm
            title="Validate Address"
            method="POST"
            path="/addresses/validate"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ address: '123 Main St, New York, NY' }}
          />
        </section>

        <section>
          <h2>Billing</h2>
          <ApiForm
            title="Create Invoice"
            method="POST"
            path="/billing/invoices"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ shipment_id: 'SHIP-001', amount: 45.99, currency: 'USD', due_date: '2026-06-24' }}
          />
        </section>

        <section>
          <h2>Returns</h2>
          <ApiForm
            title="Request Return"
            method="POST"
            path="/returns"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ shipment_id: 'SHIP-001', reason: 'Product damaged' }}
          />
        </section>

      </main>

      <footer>
        <small>Frontend for local development. Uses API Gateway by default at http://localhost:8080</small>
      </footer>
    </div>
  )
}

