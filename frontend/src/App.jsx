import React, { useState, useEffect } from 'react'
import { billing } from './services/api'
import Dashboard from './components/Dashboard'
import ShipmentList from './components/ShipmentList'
import ShipmentDetail from './components/ShipmentDetail'
import CreateShipment from './components/CreateShipment'
import Settings from './components/Settings'
import ApiForm from './components/ApiForm'
import StatusManager from './components/StatusManager'
import RateComparison from './components/RateComparison'
import ReturnManager from './components/ReturnManager'
import BillingManager from './components/BillingManager'
import TestRunner from './components/TestRunner'
import CarrierManager from './components/CarrierManager'
import AddressTools from './components/AddressTools'
import LabelManager from './components/LabelManager'

const DEFAULT_API = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const DEFAULT_TOKEN = 'Bearer test-token'

export default function App() {
  const [view, setView] = useState('dashboard')
  const [selectedShipmentId, setSelectedShipmentId] = useState(null)
  const [baseUrl, setBaseUrl] = useState(DEFAULT_API)
  const [token, setToken] = useState(DEFAULT_TOKEN)
  const [paymentAlert, setPaymentAlert] = useState(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const sessionId = params.get('session_id')
    const paymentStatus = params.get('payment_status')

    if (sessionId) {
      if (paymentStatus === 'success') {
        confirmPayment(sessionId)
      } else {
        setPaymentAlert({ type: 'error', message: 'Stripe Checkout was cancelled or payment declined.' })
        clearUrlParams()
      }
    } else if (paymentStatus === 'failed') {
      setPaymentAlert({ type: 'error', message: 'Stripe Checkout payment failed.' })
      clearUrlParams()
    }
  }, [])

  const confirmPayment = async (sessionId) => {
    try {
      setPaymentAlert({ type: 'info', message: 'Verifying payment with Stripe...' })
      const payment = await billing.confirmPayment({ session_id: sessionId })
      if (payment.status === 'completed') {
        setPaymentAlert({ type: 'success', message: `Payment Succeeded! Reference: ${payment.id.substring(0, 8)}...` })
        setView('billing') // switch view to payments to see it in history
      } else {
        setPaymentAlert({ type: 'error', message: `Payment failed or declined. Status: ${payment.status}` })
      }
    } catch (err) {
      setPaymentAlert({ type: 'error', message: `Failed to confirm payment: ${err.message}` })
    } finally {
      clearUrlParams()
      setTimeout(() => {
        setPaymentAlert(null)
      }, 8000)
    }
  }

  const clearUrlParams = () => {
    const url = new URL(window.location.href)
    url.search = ''
    window.history.replaceState({}, document.title, url.pathname)
  }

  const handleSelectShipment = (id) => {
    setSelectedShipmentId(id)
    setView('detail')
  }

  const handleCreateSuccess = (shipment) => {
    setView('list')
    // Optionally show a success message
  }
const [rateShipmentId, setRateShipmentId] = useState(null)

const handleNavigate = (newView, params = {}) => {
  if (newView === 'rate-comparison' && params.shipmentId) {
    setRateShipmentId(params.shipmentId)
  }
  setView(newView)
}

const renderContent = () => {
  switch (view) {
    case 'dashboard':
      return <Dashboard onSelectShipment={handleSelectShipment} />
    case 'list':
      return <ShipmentList onSelectShipment={handleSelectShipment} />
    case 'detail':
      return <ShipmentDetail shipmentId={selectedShipmentId} onBack={() => setView('list')} onNavigate={handleNavigate} />
    case 'create':
      return <CreateShipment onSuccess={handleCreateSuccess} onCancel={() => setView('list')} />
    case 'status-manager':
      return <StatusManager />
    case 'rate-comparison':
      return <RateComparison initialShipmentId={rateShipmentId} />
    case 'returns':
      return <ReturnManager />
    case 'billing':
      return <BillingManager />
    case 'carriers':
      return <CarrierManager />
    case 'address-tools':
      return <AddressTools />
    case 'labels':
      return <LabelManager />
    case 'test-runner':
      return <TestRunner />
    case 'settings':
      return <Settings baseUrl={baseUrl} onBaseUrlChange={setBaseUrl} token={token} onTokenChange={setToken} />
    case 'api-test':
      return <ApiTestView baseUrl={baseUrl} token={token} />
    default:
      return <Dashboard />
    }
  }

  return (
    <div className="app">
      <nav className="sidebar">
        <div className="logo">
          <h2>📦 Shipping</h2>
        </div>

        <ul className="nav-menu">
          <li>
            <button
              className={`nav-item ${view === 'dashboard' ? 'active' : ''}`}
              onClick={() => setView('dashboard')}
            >
              📊 Dashboard
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'list' ? 'active' : ''}`}
              onClick={() => setView('list')}
            >
              📋 Shipments
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'create' ? 'active' : ''}`}
              onClick={() => setView('create')}
            >
              ➕ Create Shipment
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'status-manager' ? 'active' : ''}`}
              onClick={() => setView('status-manager')}
            >
              🛠️ Status Manager
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'rate-comparison' ? 'active' : ''}`}
              onClick={() => setView('rate-comparison')}
            >
              📊 Rate Comparison
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'returns' ? 'active' : ''}`}
              onClick={() => setView('returns')}
            >
              ↩️ Returns
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'billing' ? 'active' : ''}`}
              onClick={() => setView('billing')}
            >
              💳 Invoices & Payments
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'carriers' ? 'active' : ''}`}
              onClick={() => setView('carriers')}
            >
              🚢 Carrier Manager
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'address-tools' ? 'active' : ''}`}
              onClick={() => setView('address-tools')}
            >
              📍 Address Tools
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'labels' ? 'active' : ''}`}
              onClick={() => setView('labels')}
            >
              🏷️ Label Center
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'test-runner' ? 'active' : ''}`}
              onClick={() => setView('test-runner')}
            >
              🚀 System Runner
            </button>
          </li>
          <li>
            <button
              className={`nav-item ${view === 'api-test' ? 'active' : ''}`}
              onClick={() => setView('api-test')}
            >
              🧪 API Test
            </button>
          </li>
        </ul>

        <ul className="nav-menu nav-bottom">
          <li>
            <button
              className={`nav-item ${view === 'settings' ? 'active' : ''}`}
              onClick={() => setView('settings')}
            >
              ⚙️ Settings
            </button>
          </li>
        </ul>
      </nav>

      <div className="main-content">
        {paymentAlert && (
          <div className={`alert alert-${paymentAlert.type}`} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
            <span>{paymentAlert.message}</span>
            <button className="btn-icon" onClick={() => setPaymentAlert(null)}>✕</button>
          </div>
        )}
        {renderContent()}
      </div>
    </div>
  )
}

function ApiTestView({ baseUrl, token }) {
  const [showForm, setShowForm] = useState(false)

  return (
    <div className="api-test-view">
      <h1>API Testing Console</h1>
      <p style={{ color: '#6b7280', marginBottom: '20px' }}>
        Test individual API endpoints manually
      </p>

      <div style={{ marginBottom: '20px' }}>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          {showForm ? '✕ Hide' : '+ New Request'}
        </button>
      </div>

      {showForm && (
        <div style={{ marginBottom: '30px', padding: '20px', backgroundColor: '#f9fafb', borderRadius: '8px' }}>
          <h3>Shipments</h3>
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
              sender_email: 'john@example.com',
              receiver_name: 'Jane Smith',
              receiver_address: '456 Oak Ave, Los Angeles, CA 90001',
              receiver_email: 'jane@example.com',
              weight: 2.5,
              dimensions: '10x10x10',
              carrier: 'dhl',
              service_type: 'express',
            }}
          />

          <h3>Carriers</h3>
          <ApiForm
            title="Get Carrier Rates"
            method="GET"
            path="/carriers/rates?from=New+York&to=Los+Angeles&weight=2.5"
            baseUrl={baseUrl}
            token={token}
          />
          <ApiForm
            title="Carrier Tracking"
            method="GET"
            path="/carriers/tracking?carrier=dhl&tracking_number=1234567890"
            baseUrl={baseUrl}
            token={token}
          />

          <h3>Rates</h3>
          <ApiForm
            title="Compare Rates"
            method="POST"
            path="/rates/compare"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ shipment_id: 'SHIP-001', from: 'New York', to: 'Los Angeles', weight: 2.5 }}
          />

          <h3>Address</h3>
          <ApiForm
            title="Validate Address"
            method="POST"
            path="/addresses/validate"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ address: '123 Main St, New York, NY' }}
          />

          <h3>Billing</h3>
          <ApiForm
            title="Create Invoice"
            method="POST"
            path="/billing/invoices"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ shipment_id: 'SHIP-001', amount: 45.99, currency: 'USD', due_date: '2026-06-24' }}
          />

          <h3>Returns</h3>
          <ApiForm
            title="Request Return"
            method="POST"
            path="/returns"
            baseUrl={baseUrl}
            token={token}
            defaultBody={{ shipment_id: 'SHIP-001', reason: 'Product damaged' }}
          />
        </div>
      )}
    </div>
  )
}

