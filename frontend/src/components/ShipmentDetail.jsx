import React, { useState, useEffect } from 'react'
import { shipments, tracking, billing, returns } from '../services/api'

export default function ShipmentDetail({ shipmentId, onBack, onNavigate }) {
    const [shipment, setShipment] = useState(null)
    const [trackingHistory, setTrackingHistory] = useState([])
    const [invoice, setInvoice] = useState(null)
    const [returnRequests, setReturnRequests] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [isRefreshing, setIsRefreshing] = useState(false)
    const [isEditing, setIsEditing] = useState(false)
    const [editForm, setEditForm] = useState({})

    useEffect(() => {
        loadDetails()

        // Auto-refresh if shipment is not in a terminal state
        const pollInterval = setInterval(() => {
            if (shipment && !isEditing && (shipment.status === 'pending' || shipment.status === 'validated' || shipment.status === 'processing' || shipment.status === 'created')) {
                refreshData()
            }
        }, 5000)

        return () => clearInterval(pollInterval)
    }, [shipmentId, shipment?.status, isEditing])

    const loadDetails = async () => {
        try {
            setLoading(true)
            const data = await refreshData()
            setEditForm(data)
            setError(null)
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    const refreshData = async () => {
        try {
            setIsRefreshing(true)
            const shipmentData = await shipments.get(shipmentId)
            setShipment(shipmentData)
            if (!isEditing) setEditForm(shipmentData)

            // Fetch related data in parallel
            const [historyData, invData, retData] = await Promise.all([
                tracking.getHistory(shipmentId).catch(() => []),
                billing.getInvoiceByShipment(shipmentId).catch(() => null),
                returns.list(shipmentId).catch(() => [])
            ])

            setTrackingHistory(historyData || [])
            setInvoice(invData)
            setReturnRequests(retData || [])

            return shipmentData
        } catch (err) {
            console.error('Failed to refresh shipment data', err)
        } finally {
            setIsRefreshing(false)
        }
    }

    const handleUpdate = async (e) => {
        e.preventDefault()
        try {
            setIsRefreshing(true)
            await shipments.update(shipmentId, editForm)
            setIsEditing(false)
            await refreshData()
        } catch (err) {
            alert('Update failed: ' + err.message)
        } finally {
            setIsRefreshing(false)
        }
    }

    const handleProcessPayment = async () => {
        if (!invoice) return
        try {
            const result = await billing.processPayment({ invoice_id: invoice.id, method: 'stripe' })
            if (result.checkout_url) {
                window.location.href = result.checkout_url
            }
        } catch (err) {
            alert('Payment failed: ' + err.message)
        }
    }

    const handleRequestReturn = async () => {
        const reason = window.prompt('Please enter the reason for return:')
        if (!reason) return
        try {
            await returns.create({ shipment_id: shipmentId, reason })
            await refreshData()
        } catch (err) {
            alert('Return request failed: ' + err.message)
        }
    }

    if (loading) return <div className="loading">Loading shipment details...</div>
    if (error) return <div className="alert alert-error">{error}</div>
    if (!shipment) return <div className="alert alert-error">Shipment not found</div>

    return (
        <div className="shipment-detail">
            <div className="detail-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                    <button className="btn btn-secondary" onClick={onBack}>← Back</button>
                    <h1>Shipment Details</h1>
                    {isRefreshing && <span className="refreshing-indicator" style={{ fontSize: '13px', color: 'var(--primary)', fontWeight: 600 }}>🔄 Updating...</span>}
                </div>
                <div style={{ display: 'flex', gap: '10px' }}>
                    {shipment.status === 'pending' && (
                        <button className="btn btn-outline" onClick={() => onNavigate && onNavigate('rate-comparison', { shipmentId: shipment.id })}>
                            📊 Compare Rates
                        </button>
                    )}
                    {!isEditing && (
                        <button className="btn btn-outline" onClick={() => setIsEditing(true)}>
                            ✏️ Edit details
                        </button>
                    )}
                    <button className="btn btn-primary" onClick={refreshData} disabled={isRefreshing || isEditing}>
                        Refresh
                    </button>
                </div>
            </div>

            {isEditing ? (
                <div className="detail-section card" style={{ padding: '32px' }}>
                    <h2>Edit Shipment Information</h2>
                    <form onSubmit={handleUpdate} style={{ padding: '0', border: 'none' }}>
                        <div className="form-row">
                            <div className="form-group">
                                <label>Sender Name</label>
                                <input value={editForm.sender_name || ''} onChange={e => setEditForm({...editForm, sender_name: e.target.value})} />
                            </div>
                            <div className="form-group">
                                <label>Receiver Name</label>
                                <input value={editForm.receiver_name || ''} onChange={e => setEditForm({...editForm, receiver_name: e.target.value})} />
                            </div>
                        </div>
                        <div className="form-group">
                            <label>Sender Address</label>
                            <input value={editForm.sender_address || ''} onChange={e => setEditForm({...editForm, sender_address: e.target.value})} />
                        </div>
                        <div className="form-group">
                            <label>Receiver Address</label>
                            <input value={editForm.receiver_address || ''} onChange={e => setEditForm({...editForm, receiver_address: e.target.value})} />
                        </div>
                        <div className="form-group">
                            <label>Item Description</label>
                            <input value={editForm.description || ''} onChange={e => setEditForm({...editForm, description: e.target.value})} />
                        </div>
                        <div className="form-row">
                            <div className="form-group">
                                <label>Weight (kg)</label>
                                <input type="number" step="0.1" value={editForm.weight || ''} onChange={e => setEditForm({...editForm, weight: parseFloat(e.target.value)})} />
                            </div>
                            <div className="form-group">
                                <label>Carrier</label>
                                <input value={editForm.carrier || ''} onChange={e => setEditForm({...editForm, carrier: e.target.value})} />
                            </div>
                        </div>
                        <div className="form-actions">
                            <button className="btn btn-secondary" type="button" onClick={() => setIsEditing(false)}>Cancel</button>
                            <button className="btn btn-primary" type="submit">Save Changes</button>
                        </div>
                    </form>
                </div>
            ) : (
                <div className="detail-grid">
                    {/* Basic Manifest Card */}
                    <div className="detail-section card">
                        <h2>📌 Manifest Details</h2>
                        <div className="info-group">
                            <div className="info-row">
                                <span className="label">Shipment ID:</span>
                                <span className="value mono" style={{ fontSize: '13px' }}>{shipment.id}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Status:</span>
                                <span className="value">
                                    <span className={`status-badge ${shipment.status}`}>
                                        {shipment.status}
                                    </span>
                                </span>
                            </div>
                            <div className="info-row">
                                <span className="label">Tracking ID:</span>
                                <span className="value">
                                    {shipment.tracking_number ? (
                                        <a 
                                            href={getTrackingUrl(shipment.carrier, shipment.tracking_number)} 
                                            target="_blank" 
                                            rel="noopener noreferrer"
                                            className="tracking-link"
                                        >
                                            {shipment.tracking_number} 🔗
                                        </a>
                                    ) : (
                                        <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>Awaiting manifest generation...</span>
                                    )}
                                </span>
                            </div>
                            <div className="info-row">
                                <span className="label">Selected Carrier:</span>
                                <span className="value" style={{ textTransform: 'uppercase', fontWeight: 'bold' }}>{shipment.carrier}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Service Type:</span>
                                <span className="value" style={{ textTransform: 'capitalize' }}>{shipment.service_type?.replace(/_/g, ' ')}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Description:</span>
                                <span className="value">{shipment.description || 'No description provided'}</span>
                            </div>
                            {shipment.is_international && (
                                <div className="info-row" style={{ backgroundColor: 'var(--info-bg)', padding: '8px 12px', borderRadius: '6px', border: '1px solid var(--info-border)' }}>
                                    <span className="label">Customs:</span>
                                    <span className="value">
                                        <span style={{ color: 'var(--info-text)', fontWeight: 'bold' }}>International</span>
                                        {` (${shipment.customs_value} ${shipment.customs_currency})`}
                                    </span>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Billing Details Card */}
                    <div className="detail-section card">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                            <h2 style={{ margin: 0 }}>💳 Financial Manifest</h2>
                            {invoice && invoice.status === 'pending' && (
                                <button className="btn btn-primary btn-sm" onClick={handleProcessPayment}>
                                    Pay Now
                                </button>
                            )}
                        </div>
                        {invoice ? (
                            <div className="info-group">
                                <div className="info-row">
                                    <span className="label">Invoice ID:</span>
                                    <span className="value mono" style={{ fontSize: '13px' }}>{invoice.id}</span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Amount Due:</span>
                                    <span className="value" style={{ fontWeight: 'bold', color: 'var(--primary)', fontSize: '18px' }}>
                                        ${invoice.amount.toFixed(2)} {invoice.currency}
                                    </span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Payment Status:</span>
                                    <span className="value">
                                        <span className={`status-badge ${invoice.status === 'paid' ? 'created' : 'pending'}`}>
                                            {invoice.status}
                                        </span>
                                    </span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Due Date:</span>
                                    <span className="value">{new Date(invoice.due_date).toLocaleDateString()}</span>
                                </div>
                            </div>
                        ) : (
                            <p style={{ color: 'var(--text-muted)', fontSize: '14px', fontStyle: 'italic' }}>No billing invoicing generated yet.</p>
                        )}
                    </div>

                    {/* Sender Details Card */}
                    <div className="detail-section card">
                        <h2>👤 Origin Address</h2>
                        <div className="info-group">
                            <div className="info-row">
                                <span className="label">Name:</span>
                                <span className="value" style={{ fontWeight: 'bold' }}>{shipment.sender_name}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Address:</span>
                                <span className="value">{shipment.sender_address}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Email:</span>
                                <span className="value">{shipment.sender_email || 'N/A'}</span>
                            </div>
                        </div>
                    </div>

                    {/* Receiver Details Card */}
                    <div className="detail-section card">
                        <h2>📍 Destination Address</h2>
                        <div className="info-group">
                            <div className="info-row">
                                <span className="label">Name:</span>
                                <span className="value" style={{ fontWeight: 'bold' }}>{shipment.receiver_name}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Address:</span>
                                <span className="value">{shipment.receiver_address}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Email:</span>
                                <span className="value">{shipment.receiver_email || 'N/A'}</span>
                            </div>
                        </div>
                    </div>

                    {/* Package Specifications Card */}
                    <div className="detail-section card">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                            <h2 style={{ margin: 0 }}>📦 Package Specifications</h2>
                            {shipment.status === 'delivered' && returnRequests.length === 0 && (
                                <button className="btn btn-secondary btn-sm" onClick={handleRequestReturn}>
                                    Request Return
                                </button>
                            )}
                        </div>
                        <div className="info-group">
                            <div className="info-row">
                                <span className="label">Weight:</span>
                                <span className="value">{shipment.weight} kg</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Dimensions:</span>
                                <span className="value">{shipment.dimensions || 'Not specified'}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Estimated Cost:</span>
                                <span className="value" style={{ fontWeight: 'bold' }}>${shipment.cost?.toFixed(2) || '0.00'}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Dispatch Date:</span>
                                <span className="value">{new Date(shipment.created_at).toLocaleString()}</span>
                            </div>
                            {shipment.pickup_location_id && (
                                <div className="info-row">
                                    <span className="label">FedEx Pickup Terminal:</span>
                                    <span className="value mono">{shipment.pickup_location_id}</span>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Label PDF Actions Card */}
                    <div className="detail-section card" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <h2>🏷️ Active Shipping Label</h2>
                        {shipment.label_url ? (
                            <div className="info-group">
                                <div className="info-row">
                                    <span className="label">Label ID:</span>
                                    <span className="value mono" style={{ fontSize: '13px' }}>{shipment.label_id}</span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Actions:</span>
                                    <div style={{ display: 'flex', gap: '8px' }}>
                                        <a href={shipment.label_url} target="_blank" rel="noopener noreferrer" className="btn btn-primary btn-sm">
                                            📄 Open PDF Label
                                        </a>
                                        <a href={shipment.label_url} download className="btn btn-outline btn-sm">
                                            📥 Download
                                        </a>
                                    </div>
                                </div>
                                <p style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '8px' }}>
                                    Label generated and verified by <strong>{shipment.carrier?.toUpperCase()}</strong> routing servers.
                                </p>
                            </div>
                        ) : (
                            <div>
                                <p style={{ color: 'var(--text-muted)', fontSize: '14px', fontStyle: 'italic' }}>Creating label manifest in background...</p>
                                <div className="loading-bar-container" style={{ height: '8px', background: '#cbd5e1', borderRadius: '4px', marginTop: '12px', overflow: 'hidden' }}>
                                    <div className="loading-bar" style={{ height: '100%', width: '70%', background: 'var(--primary)', borderRadius: '4px' }}></div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {/* Return Manifest Overlay Card */}
            {!isEditing && returnRequests.length > 0 && (
                <div className="detail-section card" style={{ marginTop: '24px' }}>
                    <h2>↩️ Return Manifest Requests</h2>
                    <div className="table-responsive">
                        <table className="data-table">
                            <thead>
                                <tr>
                                    <th>Return ID</th>
                                    <th>Reason</th>
                                    <th>Status</th>
                                    <th>Return Tracking ID</th>
                                    <th>Created At</th>
                                </tr>
                            </thead>
                            <tbody>
                                {returnRequests.map(ret => (
                                    <tr key={ret.id}>
                                        <td className="mono" style={{ fontSize: '13px' }}>{ret.id}</td>
                                        <td>{ret.reason}</td>
                                        <td>
                                            <span className={`status-badge ${ret.status}`}>
                                                {ret.status}
                                            </span>
                                        </td>
                                        <td className="mono">{ret.return_tracking_number || 'Pending Allocation'}</td>
                                        <td className="text-muted">{new Date(ret.created_at).toLocaleDateString()}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            {/* Timeline Progress Tracker */}
            {!isEditing && trackingHistory.length > 0 && (
                <div className="detail-section card" style={{ marginTop: '24px' }}>
                    <h2>📍 Carrier Tracking History</h2>
                    <div className="timeline">
                        {trackingHistory.map((event, idx) => (
                            <div key={idx} className="timeline-item">
                                <div className="timeline-dot"></div>
                                <div className="timeline-content">
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                        <strong style={{ fontSize: '14px', color: 'var(--primary)', textTransform: 'uppercase' }}>
                                            {event.status}
                                        </strong>
                                        <small className="text-muted" style={{ fontSize: '12px' }}>{new Date(event.timestamp).toLocaleString()}</small>
                                    </div>
                                    <p style={{ fontWeight: '500', marginTop: '4px', fontSize: '13.5px' }}>{event.location}</p>
                                    {event.description && (
                                        <p style={{ fontSize: '12.5px', fontStyle: 'italic', color: 'var(--text-muted)', marginTop: '2px' }}>
                                            {event.description}
                                        </p>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}

function getTrackingUrl(carrier, trackingId) {
    if (!trackingId) return '#'
    switch (carrier?.toLowerCase()) {
        case 'dhl':
            return `https://www.dhl.com/track?tracking-id=${trackingId}`
        case 'fedex':
            return `https://www.fedex.com/fedextrack/?trknbr=${trackingId}`
        case 'ups':
            return `https://www.ups.com/track?tracknum=${trackingId}`
        case 'usps':
            return `https://tools.usps.com/go/TrackConfirmAction?tLabels=${trackingId}`
        default:
            return '#'
    }
}
