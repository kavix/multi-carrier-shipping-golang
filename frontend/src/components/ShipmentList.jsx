import React, { useState, useEffect } from 'react'
import { shipments } from '../services/api'

export default function ShipmentList({ onSelectShipment, onNavigate }) {
    const [list, setList] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [filter, setFilter] = useState('all')
    const [searchQuery, setSearchQuery] = useState('')

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

    const handleDelete = async (id) => {
        if (!window.confirm('Are you sure you want to delete this shipment?')) return
        try {
            await shipments.delete(id)
            loadShipments()
        } catch (err) {
            alert('Failed to delete shipment: ' + err.message)
        }
    }

    // Apply status filter first
    const statusFilteredList = filter === 'all'
        ? list
        : list.filter(s => s.status === filter)

    // Apply search query filter
    const filteredList = statusFilteredList.filter(s => {
        const query = searchQuery.toLowerCase().trim()
        if (!query) return true
        
        return (
            s.id?.toLowerCase().includes(query) ||
            s.carrier?.toLowerCase().includes(query) ||
            s.receiver_name?.toLowerCase().includes(query) ||
            s.sender_name?.toLowerCase().includes(query) ||
            s.receiver_address?.toLowerCase().includes(query) ||
            s.sender_address?.toLowerCase().includes(query)
        )
    })

    if (loading) return <div className="loading">Loading shipments...</div>

    return (
        <div className="shipment-list">
            <div className="list-header">
                <div>
                    <h1>Shipments</h1>
                    <p className="subtitle">Track, update, and manage all your shipping manifests</p>
                </div>
                <div style={{ display: 'flex', gap: '10px' }}>
                    <button className="btn btn-secondary" onClick={loadShipments}>
                        🔄 Refresh
                    </button>
                    <button className="btn btn-primary" onClick={() => onNavigate && onNavigate('create')}>
                        + New Shipment
                    </button>
                </div>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '20px', flexWrap: 'wrap', marginBottom: '20px' }}>
                <div className="filters" style={{ marginBottom: 0 }}>
                    {['all', 'pending', 'validated', 'processing', 'delivered', 'cancelled'].map(status => (
                        <button
                            key={status}
                            className={`filter-btn ${filter === status ? 'active' : ''}`}
                            onClick={() => setFilter(status)}
                        >
                            {status.charAt(0).toUpperCase() + status.slice(1)}
                        </button>
                    ))}
                </div>

                <div className="search-container" style={{ margin: 0 }}>
                    <span className="search-icon">🔍</span>
                    <input
                        type="text"
                        className="search-input"
                        placeholder="Search by ID, carrier, recipient..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
            </div>

            {filteredList.length === 0 ? (
                <div className="empty-state">
                    <p>No shipments found matching the filters.</p>
                    {searchQuery && (
                        <button className="btn btn-secondary btn-sm" onClick={() => setSearchQuery('')}>Clear Search</button>
                    )}
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
                                    <td className="mono" style={{ fontSize: '13px' }}>{shipment.id.substring(0, 8)}...</td>
                                    <td style={{ fontSize: '13.5px' }}>
                                        <div><strong>{shipment.sender_name}</strong></div>
                                        <div className="text-muted" style={{ fontSize: '12px' }}>{shipment.sender_address?.substring(0, 24)}...</div>
                                    </td>
                                    <td style={{ fontSize: '13.5px' }}>
                                        <div><strong>{shipment.receiver_name}</strong></div>
                                        <div className="text-muted" style={{ fontSize: '12px' }}>{shipment.receiver_address?.substring(0, 24)}...</div>
                                    </td>
                                    <td>{shipment.weight} kg</td>
                                    <td style={{ textTransform: 'uppercase', fontWeight: 'bold', fontSize: '13px' }}>{shipment.carrier}</td>
                                    <td>
                                        <span className={`status-badge ${shipment.status}`}>
                                            {shipment.status}
                                        </span>
                                    </td>
                                    <td>
                                        {shipment.label_url ? (
                                            <a href={shipment.label_url} target="_blank" rel="noopener noreferrer" className="btn btn-sm btn-outline" style={{ padding: '4px 10px', fontSize: '12px' }}>
                                                View PDF Label
                                            </a>
                                        ) : (
                                            <span className="text-muted" style={{ fontStyle: 'italic' }}>None</span>
                                        )}
                                    </td>
                                    <td className="text-muted">{new Date(shipment.created_at).toLocaleDateString()}</td>
                                    <td>
                                        <div style={{ display: 'flex', gap: '8px' }}>
                                            <button
                                                className="btn btn-sm btn-secondary"
                                                onClick={() => onSelectShipment(shipment.id)}
                                            >
                                                View Details
                                            </button>
                                            {shipment.status === 'pending' && (
                                                <button
                                                    className="btn btn-sm btn-secondary"
                                                    style={{ color: 'var(--danger)' }}
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
