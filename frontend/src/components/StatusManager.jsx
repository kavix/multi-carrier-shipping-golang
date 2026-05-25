import React, { useState, useEffect } from 'react'
import { shipments } from '../services/api'

const STATUSES = ['pending', 'created', 'picked_up', 'in_transit', 'delivered', 'failed', 'returned']

export default function StatusManager() {
    const [list, setList] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [searchQuery, setSearchQuery] = useState('')
    const [filterStatus, setFilterStatus] = useState('all')
    const [updatingIds, setUpdatingIds] = useState({}) // keeps track of loading states for updating specific shipments
    const [selectedStatuses, setSelectedStatuses] = useState({}) // local state for dropdowns
    const [successMessage, setSuccessMessage] = useState('')

    useEffect(() => {
        loadShipments()
    }, [])

    const loadShipments = async () => {
        try {
            setLoading(true)
            const data = await shipments.list()
            setList(data || [])
            
            // Initialize dropdown states with current status
            const statusMap = {}
            data?.forEach(s => {
                statusMap[s.id] = s.status
            })
            setSelectedStatuses(statusMap)
            setError(null)
        } catch (err) {
            setError(err.message)
            setList([])
        } finally {
            setLoading(false)
        }
    }

    const handleStatusChange = (id, newStatus) => {
        setSelectedStatuses(prev => ({
            ...prev,
            [id]: newStatus
        }))
    }

    const handleUpdateStatus = async (id) => {
        const newStatus = selectedStatuses[id]
        const currentShipment = list.find(s => s.id === id)
        if (currentShipment.status === newStatus) {
            return // no change
        }

        try {
            setUpdatingIds(prev => ({ ...prev, [id]: true }))
            await shipments.updateStatus(id, newStatus)
            
            // Update local list state
            setList(prev => prev.map(s => s.id === id ? { ...s, status: newStatus } : s))
            
            setSuccessMessage(`Shipment ${id.substring(0, 8)} status updated to ${newStatus}`)
            setTimeout(() => setSuccessMessage(''), 4000)
        } catch (err) {
            setError(`Failed to update ${id.substring(0, 8)}: ${err.message}`)
            setTimeout(() => setError(null), 5000)
        } finally {
            setUpdatingIds(prev => ({ ...prev, [id]: false }))
        }
    }

    const filteredList = list.filter(shipment => {
        const matchesSearch = 
            shipment.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
            shipment.sender_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
            shipment.receiver_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
            shipment.sender_email?.toLowerCase().includes(searchQuery.toLowerCase()) ||
            shipment.receiver_email?.toLowerCase().includes(searchQuery.toLowerCase())
        
        const matchesStatus = filterStatus === 'all' || shipment.status === filterStatus

        return matchesSearch && matchesStatus
    })

    const statusColor = (status) => {
        const colors = {
            pending: '#f59e0b',
            created: '#10b981',
            picked_up: '#3b82f6',
            in_transit: '#6366f1',
            delivered: '#10b981',
            failed: '#ef4444',
            returned: '#ec4899',
        }
        return colors[status] || '#6b7280'
    }

    if (loading) return <div className="loading">Loading orders for status manager...</div>

    return (
        <div className="status-manager shipment-list">
            <div className="list-header">
                <div>
                    <h1>Order Status Manager</h1>
                    <p style={{ color: '#6b7280', margin: '4px 0 0 0' }}>
                        Quickly update delivery and fulfillment statuses for all shipments
                    </p>
                </div>
                <button className="btn btn-secondary" onClick={loadShipments}>
                    🔄 Refresh
                </button>
            </div>

            {successMessage && <div className="alert alert-success">{successMessage}</div>}
            {error && <div className="alert alert-error">{error}</div>}

            <div style={{ display: 'flex', gap: '16px', marginBottom: '20px', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between' }}>
                <div style={{ flex: '1', minWidth: '300px' }}>
                    <input
                        type="text"
                        placeholder="Search by ID, name or email..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        style={{
                            width: '100%',
                            padding: '10px 14px',
                            border: '1px solid #d1d5db',
                            borderRadius: '8px',
                            fontSize: '14px',
                            background: '#white'
                        }}
                    />
                </div>

                <div className="filters" style={{ margin: '0' }}>
                    <button
                        className={`filter-btn ${filterStatus === 'all' ? 'active' : ''}`}
                        onClick={() => setFilterStatus('all')}
                    >
                        All ({list.length})
                    </button>
                    {STATUSES.map(status => {
                        const count = list.filter(s => s.status === status).length
                        return (
                            <button
                                key={status}
                                className={`filter-btn ${filterStatus === status ? 'active' : ''}`}
                                onClick={() => setFilterStatus(status)}
                            >
                                {status.replace('_', ' ').charAt(0).toUpperCase() + status.replace('_', ' ').slice(1)} ({count})
                            </button>
                        )
                    })}
                </div>
            </div>

            {filteredList.length === 0 ? (
                <div className="empty-state">
                    <p>No orders found matching the criteria</p>
                </div>
            ) : (
                <div className="table-responsive">
                    <table className="data-table">
                        <thead>
                            <tr>
                                <th>ID</th>
                                <th>Sender (Email)</th>
                                <th>Receiver (Email)</th>
                                <th>Carrier</th>
                                <th>Current Status</th>
                                <th style={{ width: '280px' }}>Update Status</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredList.map(shipment => {
                                const isUpdating = updatingIds[shipment.id]
                                const hasChanged = selectedStatuses[shipment.id] !== shipment.status

                                return (
                                    <tr key={shipment.id}>
                                        <td className="mono" title={shipment.id}>
                                            {shipment.id.substring(0, 8)}...
                                        </td>
                                        <td>
                                            <div style={{ fontWeight: '500' }}>{shipment.sender_name}</div>
                                            <div className="text-muted">{shipment.sender_email}</div>
                                        </td>
                                        <td>
                                            <div style={{ fontWeight: '500' }}>{shipment.receiver_name}</div>
                                            <div className="text-muted">{shipment.receiver_email}</div>
                                        </td>
                                        <td>
                                            <span style={{ textTransform: 'uppercase', fontWeight: '600', fontSize: '12px' }}>
                                                {shipment.carrier}
                                            </span>
                                            <div className="text-muted">{shipment.service_type}</div>
                                        </td>
                                        <td>
                                            <span className="status-badge" style={{ backgroundColor: statusColor(shipment.status) }}>
                                                {shipment.status}
                                            </span>
                                        </td>
                                        <td>
                                            <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                                                <select
                                                    value={selectedStatuses[shipment.id] || shipment.status}
                                                    onChange={(e) => handleStatusChange(shipment.id, e.target.value)}
                                                    disabled={isUpdating}
                                                    style={{
                                                        padding: '8px 12px',
                                                        borderRadius: '6px',
                                                        border: '1px solid #d1d5db',
                                                        fontSize: '13px',
                                                        flex: '1',
                                                        outline: 'none',
                                                        cursor: 'pointer'
                                                    }}
                                                >
                                                    {STATUSES.map(st => (
                                                        <option key={st} value={st}>
                                                            {st.replace('_', ' ')}
                                                        </option>
                                                    ))}
                                                </select>
                                                <button
                                                    className={`btn btn-sm ${hasChanged ? 'btn-primary' : 'btn-secondary'}`}
                                                    onClick={() => handleUpdateStatus(shipment.id)}
                                                    disabled={isUpdating || !hasChanged}
                                                    style={{ minWidth: '80px', justifyContent: 'center' }}
                                                >
                                                    {isUpdating ? '...' : 'Update'}
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                )
                            })}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}
