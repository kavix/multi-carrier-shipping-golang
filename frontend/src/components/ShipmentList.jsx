import React, { useState, useEffect } from 'react'
import { shipments } from '../services/api'

export default function ShipmentList({ onSelectShipment }) {
    const [list, setList] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [filter, setFilter] = useState('all')

    useEffect(() => {
        loadShipments()
    }, [])

    const loadShipments = async () => {
        try {
            setLoading(true)
            const data = await shipments.list()
            setList(data || [])
            setError(null)
        } catch (err) {
            setError(err.message)
            setList([])
        } finally {
            setLoading(false)
        }
    }

    const filteredList = filter === 'all'
        ? list
        : list.filter(s => s.status === filter)

    const statusColor = (status) => {
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

    const handleDelete = async (id) => {
        if (!window.confirm('Are you sure you want to delete this shipment?')) return
        try {
            await shipments.delete(id)
            loadShipments()
        } catch (err) {
            alert('Failed to delete shipment: ' + err.message)
        }
    }

    if (loading) return <div className="loading">Loading shipments...</div>

    return (
        <div className="shipment-list">
            <div className="list-header">
                <h1>Shipments</h1>
                <div className="header-actions">
                    <button className="btn btn-secondary" onClick={loadShipments}>
                        🔄 Refresh
                    </button>
                    <button className="btn btn-primary" onClick={() => window.location.hash = '#create'}>
                        + New Shipment
                    </button>
                </div>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <div className="filters">
                {['all', 'pending', 'processing', 'delivered', 'cancelled'].map(status => (
                    <button
                        key={status}
                        className={`filter-btn ${filter === status ? 'active' : ''}`}
                        onClick={() => setFilter(status)}
                    >
                        {status.charAt(0).toUpperCase() + status.slice(1)}
                    </button>
                ))}
            </div>

            {filteredList.length === 0 ? (
                <div className="empty-state">
                    <p>No shipments found</p>
                </div>
            ) : (
                <div className="table-responsive">
                    <table className="data-table">
                        <thead>
                            <tr>
                                <th>ID</th>
                                <th>From</th>
                                <th>To</th>
                                <th>Weight</th>
                                <th>Carrier</th>
                                <th>Status</th>
                                <th>Label</th>
                                <th>Created</th>
                                <th>Action</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredList.map(shipment => (
                                <tr key={shipment.id}>
                                    <td className="mono">{shipment.id.substring(0, 8)}...</td>
                                    <td>{shipment.sender_address?.substring(0, 20)}</td>
                                    <td>{shipment.receiver_address?.substring(0, 20)}</td>
                                    <td>{shipment.weight}kg</td>
                                    <td>{shipment.carrier}</td>
                                    <td>
                                        <span className="status-badge" style={{ backgroundColor: statusColor(shipment.status) }}>
                                            {shipment.status}
                                        </span>
                                    </td>
                                    <td>
                                        {shipment.label_url ? (
                                            <a href={shipment.label_url} target="_blank" rel="noopener noreferrer" className="btn btn-sm btn-outline">
                                                View Label
                                            </a>
                                        ) : (
                                            <span className="text-muted">None</span>
                                        )}
                                    </td>
                                    <td className="text-muted">{new Date(shipment.created_at).toLocaleDateString()}</td>
                                    <td>
                                        <div style={{ display: 'flex', gap: '5px' }}>
                                            <button
                                                className="btn btn-sm btn-secondary"
                                                onClick={() => onSelectShipment(shipment.id)}
                                            >
                                                View
                                            </button>
                                            {shipment.status === 'pending' && (
                                                <button
                                                    className="btn btn-sm btn-secondary"
                                                    style={{ color: '#ef4444' }}
                                                    onClick={() => handleDelete(shipment.id)}
                                                >
                                                    Delete
                                                </button>
                                            )}
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}
