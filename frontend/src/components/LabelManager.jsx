import React, { useState, useEffect } from 'react'
import { shipments, labels } from '../services/api'

export default function LabelManager() {
    const [list, setList] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [generatingId, setGeneratingId] = useState(null)

    useEffect(() => {
        loadData()
    }, [])

    const loadData = async () => {
        try {
            setLoading(true)
            const data = await shipments.list()
            setList(data || [])
            setError(null)
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    const handleGenerateLabel = async (shipment) => {
        try {
            setGeneratingId(shipment.id)
            await labels.generate({
                shipment_id: shipment.id,
                carrier: shipment.carrier,
                format: 'pdf'
            })
            // Wait a bit for the async consumer to update the shipment
            setTimeout(loadData, 2000)
        } catch (err) {
            setError('Failed to generate label: ' + err.message)
        } finally {
            setGeneratingId(null)
        }
    }

    if (loading) return <div className="loading">Loading labels...</div>

    return (
        <div className="label-manager">
            <div className="list-header">
                <h1>Label Management Center</h1>
                <button className="btn btn-secondary" onClick={loadData}>🔄 Refresh</button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <div className="info-box" style={{ marginBottom: '24px' }}>
                <strong>🏷️ Shipping Label Service</strong>
                <p>Generate, view, and manage shipping labels for your shipments. Labels are automatically generated upon address validation, but can be manually triggered here if needed.</p>
            </div>

            <div className="table-responsive">
                <table className="data-table">
                    <thead>
                        <tr>
                            <th>Shipment ID</th>
                            <th>Recipient</th>
                            <th>Description</th>
                            <th>Carrier</th>
                            <th>Status</th>
                            <th>Tracking #</th>
                            <th>Label Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {list.map(shipment => (
                            <tr key={shipment.id}>
                                <td className="mono">{shipment.id.substring(0, 8)}...</td>
                                <td>{shipment.receiver_name}</td>
                                <td>{shipment.description || '---'}</td>
                                <td style={{ textTransform: 'uppercase', fontWeight: 600 }}>{shipment.carrier}</td>
                                <td>
                                    <span className="status-badge" style={{ backgroundColor: getStatusColor(shipment.status) }}>
                                        {shipment.status}
                                    </span>
                                </td>
                                <td className="mono">{shipment.tracking_number || 'N/A'}</td>
                                <td>
                                    {shipment.label_url ? (
                                        <span className="alert-success" style={{ padding: '2px 8px', borderRadius: '4px', fontSize: '12px' }}>
                                            Ready
                                        </span>
                                    ) : (
                                        <span className="alert-info" style={{ padding: '2px 8px', borderRadius: '4px', fontSize: '12px' }}>
                                            Missing
                                        </span>
                                    )}
                                </td>
                                <td>
                                    <div style={{ display: 'flex', gap: '8px' }}>
                                        {shipment.label_url ? (
                                            <a 
                                                href={shipment.label_url} 
                                                target="_blank" 
                                                rel="noopener noreferrer" 
                                                className="btn btn-sm btn-outline"
                                            >
                                                📄 Download
                                            </a>
                                        ) : (
                                            <button 
                                                className="btn btn-sm btn-primary"
                                                onClick={() => handleGenerateLabel(shipment)}
                                                disabled={generatingId === shipment.id}
                                            >
                                                {generatingId === shipment.id ? '...' : '⚡ Generate'}
                                            </button>
                                        )}
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    )
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
