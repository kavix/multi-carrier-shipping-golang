import React, { useState, useEffect } from 'react'
import { shipments, returns, billing, health } from '../services/api'

export default function Dashboard({ onSelectShipment }) {
    const [stats, setStats] = useState({
        totalShipments: 0,
        pendingShipments: 0,
        deliveredToday: 0,
        totalReturns: 0,
        pendingInvoices: 0,
    })
    const [recentShipments, setRecentShipments] = useState([])
    const [serviceHealth, setServiceHealth] = useState({})
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

    useEffect(() => {
        loadDashboard()
    }, [])

    const loadDashboard = async () => {
        try {
            setLoading(true)
            const shipmentsList = await shipments.list()
            const sorted = [...shipmentsList].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
            setRecentShipments(sorted.slice(0, 5))

            const pending = shipmentsList.filter(s => s.status === 'pending' || s.status === 'processing' || s.status === 'created').length
            const delivered = shipmentsList.filter(s => s.status === 'delivered').length
            const returned = shipmentsList.filter(s => s.status === 'returned').length

            // Load invoices for each shipment to count pending ones
            const invoicePromises = shipmentsList.slice(0, 20).map(async (ship) => {
                try {
                    const inv = await billing.getInvoiceByShipment(ship.id)
                    return inv
                } catch (e) {
                    return null
                }
            })
            const results = await Promise.all(invoicePromises)
            const activeInvoices = results.filter(r => r !== null)
            const pendingInvoices = activeInvoices.filter(inv => inv.status === 'pending').length

            setStats({
                totalShipments: shipmentsList.length,
                pendingShipments: pending,
                deliveredToday: delivered,
                totalReturns: returned,
                pendingInvoices: pendingInvoices,
            })

            // Check health of core services
            checkServices()
            setError(null)
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    const checkServices = async () => {
        const services = [
            { name: 'Gateway', url: 'http://localhost:8080' },
            { name: 'Shipment', url: 'http://localhost:8001' },
            { name: 'Carrier', url: 'http://localhost:8002' },
            { name: 'Rate', url: 'http://localhost:8003' },
            { name: 'Label', url: 'http://localhost:8004' },
            { name: 'Tracking', url: 'http://localhost:8005' },
            { name: 'Address', url: 'http://localhost:8006' },
            { name: 'Billing', url: 'http://localhost:8007' },
            { name: 'Return', url: 'http://localhost:8008' },
            { name: 'Notification', url: 'http://localhost:8009' },
        ]

        const healthStatus = {}
        for (const s of services) {
            try {
                const res = await health.check(s.url)
                healthStatus[s.name] = res.status === 'ok'
            } catch (e) {
                healthStatus[s.name] = false
            }
        }
        setServiceHealth(healthStatus)
    }

    if (loading) return <div className="loading">Loading dashboard...</div>

    return (
        <div className="dashboard">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '32px' }}>
                <h1>Platform Dashboard</h1>
                <button className="btn btn-outline btn-sm" onClick={loadDashboard}>🔄 Refresh Data</button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <div className="stats-grid">
                <div className="stat-card">
                    <div className="stat-number">{stats.totalShipments}</div>
                    <div className="stat-label">Total Shipments</div>
                </div>
                <div className="stat-card" style={{ borderLeft: '4px solid #f59e0b' }}>
                    <div className="stat-number" style={{ color: '#f59e0b' }}>{stats.pendingShipments}</div>
                    <div className="stat-label">Active / Pending</div>
                </div>
                <div className="stat-card" style={{ borderLeft: '4px solid #10b981' }}>
                    <div className="stat-number" style={{ color: '#10b981' }}>{stats.deliveredToday}</div>
                    <div className="stat-label">Delivered</div>
                </div>
                <div className="stat-card" style={{ borderLeft: '4px solid #ef4444' }}>
                    <div className="stat-number" style={{ color: '#ef4444' }}>{stats.pendingInvoices}</div>
                    <div className="stat-label">Unpaid Invoices</div>
                </div>
            </div>

            <div className="dashboard-grid" style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '24px' }}>
                <section className="dashboard-section" style={{ background: 'white', padding: '24px', borderRadius: '12px', border: '1px solid #e5e7eb' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '16px' }}>
                        <h2 style={{ margin: 0, fontSize: '18px' }}>Recent Shipments</h2>
                    </div>
                    {recentShipments.length > 0 ? (
                        <div className="table-responsive">
                            <table className="data-table">
                                <thead>
                                    <tr>
                                        <th>ID</th>
                                        <th>Recipient</th>
                                        <th>Status</th>
                                        <th>Carrier</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {recentShipments.map(s => (
                                        <tr key={s.id} onClick={() => onSelectShipment(s.id)} style={{ cursor: 'pointer' }}>
                                            <td className="mono" style={{ fontSize: '12px' }}>{s.id.substring(0, 8)}...</td>
                                            <td>{s.receiver_name}</td>
                                            <td>
                                                <span className="status-badge" style={{ backgroundColor: getStatusColor(s.status), fontSize: '10px' }}>
                                                    {s.status}
                                                </span>
                                            </td>
                                            <td>{s.carrier.toUpperCase()}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <p className="text-muted">No recent shipments found.</p>
                    )}
                </section>

                <section className="dashboard-section" style={{ background: 'white', padding: '24px', borderRadius: '12px', border: '1px solid #e5e7eb' }}>
                    <h2 style={{ margin: 0, fontSize: '18px', marginBottom: '16px' }}>System Health</h2>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: '10px' }}>
                        {Object.entries(serviceHealth).map(([name, isOk]) => (
                            <div key={name} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px', background: '#f9fafb', borderRadius: '6px' }}>
                                <span style={{ fontSize: '14px', fontWeight: 500 }}>{name} Service</span>
                                <span style={{ 
                                    width: '10px', 
                                    height: '10px', 
                                    borderRadius: '50%', 
                                    background: isOk ? '#10b981' : '#ef4444',
                                    boxShadow: isOk ? '0 0 8px #10b981' : '0 0 8px #ef4444'
                                }}></span>
                            </div>
                        ))}
                        {Object.keys(serviceHealth).length === 0 && <p className="text-muted">Checking services...</p>}
                    </div>
                </section>
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

