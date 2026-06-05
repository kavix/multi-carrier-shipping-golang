import React, { useState, useEffect } from 'react'
import { shipments, billing } from '../services/api'

export default function BillingManager() {
    const [invoices, setInvoices] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [activeTab, setActiveTab] = useState('unpaid')
    
    const [isRedirecting, setIsRedirecting] = useState(false)

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

    const handlePayClick = async (invoice) => {
        try {
            setError(null)
            setIsRedirecting(true)
            const session = await billing.processPayment({
                invoice_id: invoice.id,
                method: 'stripe'
            })
            if (session.checkout_url) {
                window.location.href = session.checkout_url
            } else {
                throw new Error('Checkout URL not returned from backend')
            }
        } catch (err) {
            setError('Checkout redirection failed: ' + err.message)
            setIsRedirecting(false)
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
                                                onClick={() => handlePayClick(inv)}
                                                disabled={isRedirecting}
                                            >
                                                {isRedirecting ? '🔄 Redirecting...' : '💳 Pay Now'}
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

        </div>
    )
}
