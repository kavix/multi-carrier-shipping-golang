import React, { useState, useEffect } from 'react'
import { shipments, tracking } from '../services/api'

export default function ShipmentDetail({ shipmentId, onBack }) {
    const [shipment, setShipment] = useState(null)
    const [trackingHistory, setTrackingHistory] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [isRefreshing, setIsRefreshing] = useState(false)
    const [isEditing, setIsEditing] = useState(false)
    const [editForm, setEditForm] = useState({})

    useEffect(() => {
        loadDetails()

        // Auto-refresh if shipment is not in a terminal state
        const pollInterval = setInterval(() => {
            if (shipment && !isEditing && (shipment.status === 'pending' || shipment.status === 'validated' || shipment.status === 'processing')) {
                refreshData()
            }
        }, 3000)

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

            try {
                const historyData = await tracking.getHistory(shipmentId)
                setTrackingHistory(historyData || [])
            } catch (err) {
                // Tracking might not be available yet
            }
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

    if (loading) return <div className="loading">Loading shipment details...</div>
    if (error) return <div className="alert alert-error">{error}</div>
    if (!shipment) return <div className="alert alert-error">Shipment not found</div>

    return (
        <div className="shipment-detail">
            <div className="detail-header">
                <div className="header-left">
                    <button className="btn btn-secondary" onClick={onBack}>← Back</button>
                    <h1>Shipment Details</h1>
                    {isRefreshing && <span className="refreshing-indicator">🔄 Updating...</span>}
                </div>
                <div style={{ display: 'flex', gap: '10px' }}>
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
                                    ) : 'Not assigned'}
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
                        </div>
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
                        <h2>Package Details</h2>
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
                                <span className="label">Cost:</span>
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
                            {shipment.drop_location_id && (
                                <div className="info-row">
                                    <span className="label">Drop Location:</span>
                                    <span className="value mono">{shipment.drop_location_id}</span>
                                </div>
                            )}
                        </div>
                    </div>

                    {shipment.label_url && (
                        <div className="detail-section">
                            <h2>Shipping Label</h2>
                            <div className="info-group">
                                <div className="info-row">
                                    <span className="label">Label ID:</span>
                                    <span className="value mono">{shipment.label_id}</span>
                                </div>
                                <div className="info-row">
                                    <span className="label">Label URL:</span>
                                    <span className="value">
                                        <a href={shipment.label_url} target="_blank" rel="noopener noreferrer" className="btn btn-primary btn-sm">
                                            Open PDF Label
                                        </a>
                                    </span>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            )}

            {!isEditing && trackingHistory.length > 0 && (
                <div className="detail-section">
                    <h2>Tracking History</h2>
                    <div className="timeline">
                        {trackingHistory.map((event, idx) => (
                            <div key={idx} className="timeline-item">
                                <div className="timeline-dot"></div>
                                <div className="timeline-content">
                                    <strong>{event.status}</strong>
                                    <p>{event.location}</p>
                                    <small className="text-muted">{new Date(event.timestamp).toLocaleString()}</small>
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
        case 'ems':
        case 'slpost':
            return `https://www.slpost.lk/track?tracking-id=${trackingId}`
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
        delivered: '#10b981',
        cancelled: '#ef4444',
    }
    return colors[status] || '#6b7280'
}
