import React, { useState, useEffect } from 'react'
import { shipments, billing } from '../services/api'

export default function BillingManager() {
    const [invoices, setInvoices] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [activeTab, setActiveTab] = useState('unpaid')
    
    // Checkout Modal State
    const [showCheckout, setShowCheckout] = useState(false)
    const [selectedInvoice, setSelectedInvoice] = useState(null)
    const [cardHolder, setCardHolder] = useState('')
    const [cardNumber, setCardNumber] = useState('')
    const [cardExpiry, setCardExpiry] = useState('')
    const [cardCvv, setCardCvv] = useState('')
    const [isPaying, setIsPaying] = useState(false)
    const [paymentSuccess, setPaymentSuccess] = useState(false)
    const [paymentError, setPaymentError] = useState(null)

    useEffect(() => {
        loadData()
    }, [])

    const loadData = async () => {
        try {
            setLoading(true)
            const list = await shipments.list()
            
            // For each shipment, load its invoice details in parallel
            const invoicePromises = list.map(async (ship) => {
                try {
                    const inv = await billing.getInvoiceByShipment(ship.id)
                    return { ...inv, shipment: ship }
                } catch (e) {
                    // 404 or missing invoice: return null
                    return null
                }
            })
            
            const results = await Promise.all(invoicePromises)
            const activeInvoices = results.filter(r => r !== null)
            setInvoices(activeInvoices)
            setError(null)
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    const openCheckout = (invoice) => {
        setSelectedInvoice(invoice)
        setCardHolder('')
        setCardNumber('')
        setCardExpiry('')
        setCardCvv('')
        setPaymentError(null)
        setPaymentSuccess(false)
        setShowCheckout(true)
    }

    const handlePaymentSubmit = async (e) => {
        e.preventDefault()
        if (!selectedInvoice) return

        try {
            setIsPaying(true)
            setPaymentError(null)

            // Submit payment to Stripe endpoint
            const payment = await billing.processPayment({
                invoice_id: selectedInvoice.id,
                method: 'stripe'
            })

            setPaymentSuccess(true)
            
            // Wait 1.5 seconds to show visual checkmark, then refresh lists and close modal
            setTimeout(() => {
                setShowCheckout(false)
                loadData()
            }, 1500)
        } catch (err) {
            setPaymentError(err.message || 'Payment failed')
        } finally {
            setIsPaying(false)
        }
    }

    const unpaidInvoices = invoices.filter(inv => inv.status === 'pending')
    const paidInvoices = invoices.filter(inv => inv.status === 'paid')

    // Calculated Statistics
    const totalInvoiced = invoices.reduce((sum, inv) => sum + inv.amount, 0)
    const totalPending = unpaidInvoices.reduce((sum, inv) => sum + inv.amount, 0)
    const totalPaid = paidInvoices.reduce((sum, inv) => sum + inv.amount, 0)

    if (loading) return <div className="loading">Loading invoices and billing records...</div>

    return (
        <div className="billing-manager">
            <div className="list-header">
                <h1>Invoices & Payments</h1>
                <button className="btn btn-primary" onClick={loadData}>Refresh</button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            {/* Billing Summary Cards */}
            <div className="billing-summary">
                <div className="billing-stat-card">
                    <div className="billing-stat-val" style={{ color: '#0f172a' }}>${totalInvoiced.toFixed(2)}</div>
                    <div className="billing-stat-lbl">Total Invoiced</div>
                </div>
                <div className="billing-stat-card">
                    <div className="billing-stat-val" style={{ color: '#f59e0b' }}>${totalPending.toFixed(2)}</div>
                    <div className="billing-stat-lbl">Pending Payments</div>
                </div>
                <div className="billing-stat-card">
                    <div className="billing-stat-val" style={{ color: '#10b981' }}>${totalPaid.toFixed(2)}</div>
                    <div className="billing-stat-lbl">Successful Payments</div>
                </div>
            </div>

            {/* Navigation Tabs */}
            <div className="tabs-container">
                <div className="filters">
                    <button 
                        className={`filter-btn ${activeTab === 'unpaid' ? 'active' : ''}`}
                        onClick={() => setActiveTab('unpaid')}
                    >
                        ⏳ Unpaid Invoices ({unpaidInvoices.length})
                    </button>
                    <button 
                        className={`filter-btn ${activeTab === 'paid' ? 'active' : ''}`}
                        onClick={() => setActiveTab('paid')}
                    >
                        ✅ Payment History ({paidInvoices.length})
                    </button>
                </div>
            </div>

            {/* Tab Panels */}
            {activeTab === 'unpaid' ? (
                unpaidInvoices.length === 0 ? (
                    <div className="empty-state">
                        <p>🎉 All invoices are fully paid! No outstanding balances found.</p>
                    </div>
                ) : (
                    <div className="table-responsive">
                        <table className="data-table">
                            <thead>
                                <tr>
                                    <th>Invoice ID</th>
                                    <th>Shipment ID</th>
                                    <th>Carrier</th>
                                    <th>Description</th>
                                    <th>Amount</th>
                                    <th>Status</th>
                                    <th>Action</th>
                                </tr>
                            </thead>
                            <tbody>
                                {unpaidInvoices.map((inv) => (
                                    <tr key={inv.id}>
                                        <td className="mono" style={{ fontSize: '13px' }}>{inv.id.substring(0, 8)}...</td>
                                        <td className="mono" style={{ fontSize: '13px' }}>{inv.shipment_id.substring(0, 8)}...</td>
                                        <td style={{ textTransform: 'uppercase', fontWeight: 600 }}>{inv.shipment?.carrier || 'DHL'}</td>
                                        <td>{inv.description || 'Shipping Fee'}</td>
                                        <td style={{ fontWeight: 600 }}>${inv.amount.toFixed(2)}</td>
                                        <td>
                                            <span className="status-badge" style={{ backgroundColor: '#f59e0b' }}>
                                                {inv.status}
                                            </span>
                                        </td>
                                        <td>
                                            <button 
                                                className="btn btn-primary btn-sm"
                                                onClick={() => openCheckout(inv)}
                                            >
                                                💳 Pay Now
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )
            ) : (
                paidInvoices.length === 0 ? (
                    <div className="empty-state">
                        <p>No processed payment transactions found.</p>
                    </div>
                ) : (
                    <div className="table-responsive">
                        <table className="data-table">
                            <thead>
                                <tr>
                                    <th>Invoice ID</th>
                                    <th>Shipment ID</th>
                                    <th>Description</th>
                                    <th>Paid Amount</th>
                                    <th>Stripe Reference ID</th>
                                    <th>Status</th>
                                    <th>Paid At</th>
                                </tr>
                            </thead>
                            <tbody>
                                {paidInvoices.map((inv) => (
                                    <tr key={inv.id}>
                                        <td className="mono" style={{ fontSize: '13px' }}>{inv.id.substring(0, 8)}...</td>
                                        <td className="mono" style={{ fontSize: '13px' }}>{inv.shipment_id.substring(0, 8)}...</td>
                                        <td>{inv.description || 'Shipping Fee'}</td>
                                        <td style={{ fontWeight: 600, color: '#10b981' }}>${inv.amount.toFixed(2)}</td>
                                        <td className="mono" style={{ color: '#2563eb', fontWeight: 500 }}>
                                            {inv.stripe_id || 'stripe_sim_ref'}
                                        </td>
                                        <td>
                                            <span className="status-badge" style={{ backgroundColor: '#10b981' }}>
                                                {inv.status}
                                            </span>
                                        </td>
                                        <td>{new Date(inv.created_at).toLocaleString()}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )
            )}

            {/* Payment Modal (Checkout Drawer) */}
            {showCheckout && selectedInvoice && (
                <div className="payment-modal-overlay">
                    <div className="payment-modal">
                        <div className="payment-modal-header">
                            <h3>Secure Checkout</h3>
                            <button className="close-modal-btn" onClick={() => setShowCheckout(false)}>✕</button>
                        </div>
                        
                        <div className="payment-modal-body">
                            {/* Card Display Preview */}
                            <div className="credit-card-mockup">
                                <div className="card-logo">
                                    <span>VISA</span>
                                    <span style={{ fontSize: '12px', opacity: 0.8 }}>Stripe Gateway</span>
                                </div>
                                <div className="card-chip"></div>
                                <div className="card-number-display">
                                    {cardNumber ? cardNumber.replace(/(\d{4})/g, '$1 ').trim() : '•••• •••• •••• ••••'}
                                </div>
                                <div className="card-details-display">
                                    <div className="card-holder-display">
                                        Cardholder
                                        <strong>{cardHolder || 'JOHN DOE'}</strong>
                                    </div>
                                    <div className="card-expiry-display">
                                        Expires
                                        <strong>{cardExpiry || 'MM/YY'}</strong>
                                    </div>
                                </div>
                            </div>

                            {/* Summary Rows */}
                            <div style={{ marginBottom: '20px', backgroundColor: '#f8fafc', padding: '16px', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                                <div className="checkout-summary-row">
                                    <span>Invoice Reference:</span>
                                    <span className="mono">{selectedInvoice.id.substring(0, 8)}...</span>
                                </div>
                                <div className="checkout-summary-row">
                                    <span>Service Fee:</span>
                                    <span>${selectedInvoice.amount.toFixed(2)}</span>
                                </div>
                                <div className="checkout-summary-row total">
                                    <span>Amount Due:</span>
                                    <span>${selectedInvoice.amount.toFixed(2)}</span>
                                </div>
                            </div>

                            {/* Form Inputs */}
                            {paymentSuccess ? (
                                <div className="alert alert-success" style={{ textAlign: 'center', padding: '24px 16px' }}>
                                    <h3 style={{ margin: '0 0 8px 0' }}>✅ Payment Successful</h3>
                                    <p style={{ margin: 0, fontSize: '13px' }}>Your shipment is now scheduled for carrier handover.</p>
                                </div>
                            ) : (
                                <form onSubmit={handlePaymentSubmit} style={{ padding: 0, border: 'none' }}>
                                    {paymentError && <div className="alert alert-error">{paymentError}</div>}
                                    
                                    <div className="form-group">
                                        <label>Cardholder Name</label>
                                        <input 
                                            type="text" 
                                            placeholder="John Doe" 
                                            value={cardHolder}
                                            onChange={(e) => setCardHolder(e.target.value)}
                                            required
                                            disabled={isPaying}
                                        />
                                    </div>

                                    <div className="form-group">
                                        <label>Card Number</label>
                                        <input 
                                            type="text" 
                                            placeholder="4111 1111 1111 1111" 
                                            maxLength="16"
                                            value={cardNumber}
                                            onChange={(e) => setCardNumber(e.target.value.replace(/\D/g, ''))}
                                            required
                                            disabled={isPaying}
                                        />
                                    </div>

                                    <div className="form-row">
                                        <div className="form-group">
                                            <label>Expiry Date</label>
                                            <input 
                                                type="text" 
                                                placeholder="MM/YY" 
                                                maxLength="5"
                                                value={cardExpiry}
                                                onChange={(e) => setCardExpiry(e.target.value)}
                                                required
                                                disabled={isPaying}
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>CVV / CVC</label>
                                            <input 
                                                type="password" 
                                                placeholder="•••" 
                                                maxLength="3"
                                                value={cardCvv}
                                                onChange={(e) => setCardCvv(e.target.value.replace(/\D/g, ''))}
                                                required
                                                disabled={isPaying}
                                            />
                                        </div>
                                    </div>

                                    <button 
                                        type="submit" 
                                        className="btn btn-primary" 
                                        style={{ width: '100%', justifyContent: 'center', padding: '12px 16px', marginTop: '16px' }}
                                        disabled={isPaying}
                                    >
                                        {isPaying ? '🔒 Processing...' : `Pay $${selectedInvoice.amount.toFixed(2)}`}
                                    </button>
                                </form>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
