import React, { useState, useEffect } from 'react'
import { shipments, tracking } from '../services/api'

export default function ShipmentDetail({ shipmentId, onBack }) {
    const [shipment, setShipment] = useState(null)
    const [trackingHistory, setTrackingHistory] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

    useEffect(() => {
        loadDetails()
    }, [shipmentId])

    const loadDetails = async () => {
        try {
            setLoading(true)
            const shipmentData = await shipments.get(shipmentId)
            setShipment(shipmentData)

            try {
                const historyData = await tracking.getHistory(shipmentId)
                setTrackingHistory(historyData || [])
            } catch (err) {
                console.log('Tracking history not available')
            }

            setError(null)
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    if (loading) return <div className="loading">Loading shipment details...</div>
    if (error) return <div className="alert alert-error">{error}</div>
    if (!shipment) return <div className="alert alert-error">Shipment not found</div>

    return (
        <div className="shipment-detail">
            <div className="detail-header">
                <button className="btn btn-secondary" onClick={onBack}>← Back</button>
                <h1>Shipment Details</h1>
            </div>

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
                            <span className="value">{shipment.tracking_number || 'Not assigned'}</span>
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
                    </div>
                </div>
            </div>

            {trackingHistory.length > 0 && (
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

function getStatusColor(status) {
    const colors = {
        pending: '#f59e0b',
        processing: '#3b82f6',
        delivered: '#10b981',
        cancelled: '#ef4444',
    }
    return colors[status] || '#6b7280'
}
