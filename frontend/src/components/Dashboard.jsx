import React, { useState, useEffect } from 'react'
import { shipments, returns, billing } from '../services/api'

export default function Dashboard() {
    const [stats, setStats] = useState({
        totalShipments: 0,
        pendingShipments: 0,
        totalReturns: 0,
        pendingInvoices: 0,
    })
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

    useEffect(() => {
        loadDashboard()
    }, [])

    const loadDashboard = async () => {
        try {
            setLoading(true)
            const shipmentsList = await shipments.list()
            const pending = shipmentsList.filter(s => s.status === 'pending' || s.status === 'processing').length

            setStats({
                totalShipments: shipmentsList.length,
                pendingShipments: pending,
                totalReturns: 0,
                pendingInvoices: 0,
            })
            setError(null)
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    if (loading) return <div className="loading">Loading dashboard...</div>

    return (
        <div className="dashboard">
            <h1>Dashboard</h1>

            {error && <div className="alert alert-error">{error}</div>}

            <div className="stats-grid">
                <div className="stat-card">
                    <div className="stat-number">{stats.totalShipments}</div>
                    <div className="stat-label">Total Shipments</div>
                </div>
                <div className="stat-card">
                    <div className="stat-number" style={{ color: '#ef4444' }}>{stats.pendingShipments}</div>
                    <div className="stat-label">Pending</div>
                </div>
                <div className="stat-card">
                    <div className="stat-number">{stats.totalReturns}</div>
                    <div className="stat-label">Return Requests</div>
                </div>
                <div className="stat-card">
                    <div className="stat-number">{stats.pendingInvoices}</div>
                    <div className="stat-label">Pending Invoices</div>
                </div>
            </div>

            <div className="dashboard-sections">
                <section>
                    <h2>Recent Activity</h2>
                    <p style={{ color: '#6b7280' }}>No recent activity yet. Create a shipment to get started.</p>
                </section>
            </div>
        </div>
    )
}
