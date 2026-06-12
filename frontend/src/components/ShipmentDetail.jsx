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
                <div className="header-left" style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                    <button className="btn btn-secondary" onClick={onBack}>← Back</button>
                    <h1>Shipment Details</h1>
                    {isRefreshing && <span className="refreshing-indicator" style={{ fontSize: '12px', color: '#3b82f6' }}>🔄 Updating...</span>}
                </div>
                <div style={{ display: 'flex', gap: '10px' }}>
                    {shipment.status === 'pending' && (
                        <button className="btn btn-outline" onClick={() => onNavigate && onNavigate('rate-comparison', { shipmentId: shipment.id })}>
                            📊 Compare Rates
                        </button>
                    )}
                    {!isEditing && (
                        <button className="btn btn-outline" onClick={() => setIsEditing(true)}>
                            ✏️ Edit
                        </button>
                    )}
                    <button className="btn btn-primary" onClick={refreshData} disabled={isRefreshing || isEditing}>
                        Refresh
                    </button>
                </div>
            </div>

            {isEditing ? (
                <div className="detail-section">
                    <h2>Edit Shipment Information</h2>
                    <form onSubmit={handleUpdate}>
                        <div className="form-row">
                            <div className="form-group">
                                <label>Sender Name</label>
                                <input value={editForm.sender_name} onChange={e => setEditForm({...editForm, sender_name: e.target.value})} />
                            </div>
                            <div className="form-group">
                                <label>Receiver Name</label>
                                <input value={editForm.receiver_name} onChange={e => setEditForm({...editForm, receiver_name: e.target.value})} />
                            </div>
                        </div>
                        <div className="form-group">
                            <label>Sender Address</label>
                            <input value={editForm.sender_address} onChange={e => setEditForm({...editForm, sender_address: e.target.value})} />
                        </div>
                        <div className="form-group">
                            <label>Receiver Address</label>
                            <input value={editForm.receiver_address} onChange={e => setEditForm({...editForm, receiver_address: e.target.value})} />
                        </div>
                        <div className="form-group">
                            <label>Item Description</label>
                            <input value={editForm.description} onChange={e => setEditForm({...editForm, description: e.target.value})} />
                        </div>
                        <div className="form-row">
                            <div className="form-group">
                                <label>Weight (kg)</label>
                                <input type="number" step="0.1" value={editForm.weight} onChange={e => setEditForm({...editForm, weight: parseFloat(e.target.value)})} />
                            </div>
                            <div className="form-group">
                                <label>Carrier</label>
                                <input value={editForm.carrier} onChange={e => setEditForm({...editForm, carrier: e.target.value})} />
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
                    <div className="detail-section">
                        <h2>Basic Information</h2>
                        <div className="info-group">
                            <div className="info-row">
                                <span className="label">Shipment ID:</span>
                                <span className="value mono">{shipment.id}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Status:</span>
                                <span className="value">
                                    <span className="status-badge" style={{ backgroundColor: getStatusColor(shipment.status) }}>
                                        {shipment.status}
                                    </span>
                                </span>
                            </div>
                            <div className="info-row">
                                <span className="label">Tracking Number:</span>
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
                                        <span style={{ color: '#9ca3af', fontStyle: 'italic' }}>Pending label generation...</span>
                                    )}
                                </span>
                            </div>
                            <div className="info-row">
                                <span className="label">Carrier:</span>
                                <span className="value">{shipment.carrier}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Service Type:</span>
                                <span className="value">{shipment.service_type}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Description:</span>
                                <span className="value">{shipment.description || 'No description'}</span>
                            </div>
                            {shipment.is_international && (
                                <div className="info-row">
                                    <span className="label">Customs:</span>
                                    <span className="value">
                                        <span style={{ color: '#059669', fontWeight: 600 }}>International</span> 
                                        ({shipment.customs_value} {shipment.customs_currency})
                                    </span>
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="detail-section">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                            <h2 style={{ margin: 0 }}>Billing & Payment</h2>
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
                                    <span className="value mono">{invoice.id}</span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Amount:</span>
                                    <span className="value" style={{ fontWeight: 'bold' }}>
                                        {invoice.amount} {invoice.currency}
                                    </span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Payment Status:</span>
                                    <span className="value">
                                        <span className="status-badge" style={{ backgroundColor: invoice.status === 'paid' ? '#10b981' : '#f59e0b' }}>
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
                            <p style={{ color: '#6b7280', fontSize: '14px' }}>No invoice generated yet.</p>
                        )}
                    </div>

                    <div className="detail-section">
                        <h2>Sender Information</h2>
                        <div className="info-group">
                            <div className="info-row">
                                <span className="label">Name:</span>
                                <span className="value">{shipment.sender_name}</span>
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

                    <div className="detail-section">
                        <h2>Receiver Information</h2>
                        <div className="info-group">
                            <div className="info-row">
                                <span className="label">Name:</span>
                                <span className="value">{shipment.receiver_name}</span>
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

                    <div className="detail-section">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                            <h2 style={{ margin: 0 }}>Package Details</h2>
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
                                <span className="value">${shipment.cost?.toFixed(2) || '0.00'}</span>
                            </div>
                            <div className="info-row">
                                <span className="label">Created:</span>
                                <span className="value">{new Date(shipment.created_at).toLocaleString()}</span>
                            </div>
                            {shipment.pickup_location_id && (
                                <div className="info-row">
                                    <span className="label">Pickup Location:</span>
                                    <span className="value mono">{shipment.pickup_location_id}</span>
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="detail-section">
                        <h2>Shipping Label & Center</h2>
                        {shipment.label_url ? (
                            <div className="info-group">
                                <div className="info-row">
                                    <span className="label">Label ID:</span>
                                    <span className="value mono">{shipment.label_id}</span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Action:</span>
                                    <div style={{ display: 'flex', gap: '8px' }}>
                                        <a href={shipment.label_url} target="_blank" rel="noopener noreferrer" className="btn btn-primary btn-sm">
                                            📄 View PDF
                                        </a>
                                        <a href={shipment.label_url} download className="btn btn-outline btn-sm">
                                            📥 Download
                                        </a>
                                    </div>
                                </div>
                                <p style={{ fontSize: '12px', color: '#6b7280', marginTop: '10px' }}>
                                    Label generated by <strong>{shipment.carrier}</strong> systems.
                                </p>
                            </div>
                        ) : (
                            <div>
                                <p style={{ color: '#6b7280', fontSize: '14px' }}>Label generation is in progress...</p>
                                <div className="loading-bar-container" style={{ height: '8px', background: '#e5e7eb', borderRadius: '4px', marginTop: '10px', overflow: 'hidden' }}>
                                    <div className="loading-bar" style={{ height: '100%', width: '60%', background: '#3b82f6', borderRadius: '4px' }}></div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {!isEditing && returnRequests.length > 0 && (
                <div className="detail-section" style={{ marginTop: '20px' }}>
                    <h2>Return Requests</h2>
                    <div className="table-responsive">
                        <table className="data-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Reason</th>
                                    <th>Status</th>
                                    <th>Tracking</th>
                                    <th>Created</th>
                                </tr>
                            </thead>
                            <tbody>
                                {returnRequests.map(ret => (
                                    <tr key={ret.id}>
                                        <td className="mono">{ret.id}</td>
                                        <td>{ret.reason}</td>
                                        <td>
                                            <span className="status-badge" style={{ backgroundColor: getStatusColor(ret.status) }}>
                                                {ret.status}
                                            </span>
                                        </td>
                                        <td className="mono">{ret.return_tracking_number || 'Pending'}</td>
                                        <td>{new Date(ret.created_at).toLocaleDateString()}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            {!isEditing && trackingHistory.length > 0 && (
                <div className="detail-section" style={{ marginTop: '20px' }}>
                    <h2>Tracking Timeline</h2>
                    <div className="timeline">
                        {trackingHistory.map((event, idx) => (
                            <div key={idx} className="timeline-item">
                                <div className="timeline-dot"></div>
                                <div className="timeline-content">
                                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                        <strong>{event.status.toUpperCase()}</strong>
                                        <small className="text-muted">{new Date(event.timestamp).toLocaleString()}</small>
                                    </div>
                                    <p>{event.location}</p>
                                    {event.description && <p style={{ fontSize: '13px', fontStyle: 'italic' }}>{event.description}</p>}
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

function getStatusColor(status) {
    const colors = {
        pending: '#f59e0b',
        validated: '#3b82f6',
        created: '#10b981',
        processing: '#3b82f6',
        in_transit: '#3b82f6',
        delivered: '#10b981',
        cancelled: '#ef4444',
        failed: '#ef4444',
        approved: '#10b981',
    }
    return colors[status] || '#6b7280'
}

