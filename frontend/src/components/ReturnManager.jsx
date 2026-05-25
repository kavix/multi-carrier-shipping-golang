import React, { useState } from 'react'
import { returns } from '../services/api'

export default function ReturnManager() {
    const [shipmentId, setShipmentId] = useState('')
    const [reason, setReason] = useState('')
    const [returnList, setReturnList] = useState([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)
    const [success, setSuccess] = useState(null)

    const handleSearch = async (e) => {
        e.preventDefault()
        if (!shipmentId) return

        setLoading(true)
        setError(null)
        try {
            const data = await returns.list(shipmentId)
            setReturnList(data || [])
        } catch (err) {
            setError(err.message)
            setReturnList([])
        } finally {
            setLoading(false)
        }
    }

    const handleRequestReturn = async (e) => {
        e.preventDefault()
        setLoading(true)
        setError(null)
        setSuccess(null)
        try {
            const data = await returns.create({ shipment_id: shipmentId, reason })
            setSuccess(`Return requested successfully! ID: ${data.id}`)
            setReason('')
            // Refresh list
            const updatedList = await returns.list(shipmentId)
            setReturnList(updatedList || [])
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    const handleApprove = async (id) => {
        const carrier = prompt("Enter carrier (e.g., fedex, ups):", "fedex")
        if (!carrier) return

        setLoading(true)
        try {
            await returns.approve(id, carrier)
            setSuccess("Return approved!")
            const updatedList = await returns.list(shipmentId)
            setReturnList(updatedList || [])
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    const handleRefund = async (id) => {
        const amount = prompt("Enter refund amount:", "10.00")
        if (!amount) return

        setLoading(true)
        try {
            await returns.refund(id, parseFloat(amount))
            setSuccess("Refund processed!")
            const updatedList = await returns.list(shipmentId)
            setReturnList(updatedList || [])
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="return-manager">
            <h1>Return Management</h1>

            <div className="card">
                <h3>Find or Request Returns</h3>
                <form onSubmit={handleSearch} className="form-group">
                    <label>Shipment ID</label>
                    <div style={{ display: 'flex', gap: '10px' }}>
                        <input
                            type="text"
                            value={shipmentId}
                            onChange={(e) => setShipmentId(e.target.value)}
                            placeholder="Enter Shipment ID"
                            required
                        />
                        <button type="submit" className="btn btn-secondary" disabled={loading}>
                            Search
                        </button>
                    </div>
                </form>

                {shipmentId && (
                    <form onSubmit={handleRequestReturn} className="form-group" style={{ marginTop: '20px', borderTop: '1px solid #eee', paddingTop: '20px' }}>
                        <h4>Request New Return</h4>
                        <label>Reason</label>
                        <textarea
                            value={reason}
                            onChange={(e) => setReason(e.target.value)}
                            placeholder="Reason for return"
                            required
                        />
                        <button type="submit" className="btn btn-primary" disabled={loading} style={{ marginTop: '10px' }}>
                            Request Return
                        </button>
                    </form>
                )}
            </div>

            {error && <div className="alert alert-error" style={{ marginTop: '20px' }}>{error}</div>}
            {success && <div className="alert alert-success" style={{ marginTop: '20px' }}>{success}</div>}

            {loading && <div className="loading">Processing...</div>}

            {returnList.length > 0 && (
                <div className="return-list" style={{ marginTop: '30px' }}>
                    <h3>Existing Returns for {shipmentId}</h3>
                    <div className="table-responsive">
                        <table className="data-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Status</th>
                                    <th>Reason</th>
                                    <th>Refund</th>
                                    <th>Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {returnList.map(ret => (
                                    <tr key={ret.id}>
                                        <td className="mono">{ret.id.substring(0, 8)}...</td>
                                        <td>
                                            <span className={`status-badge status-${ret.status.toLowerCase()}`}>
                                                {ret.status}
                                            </span>
                                        </td>
                                        <td>{ret.reason}</td>
                                        <td>
                                            {ret.refund_amount > 0 ? `$${ret.refund_amount.toFixed(2)} (${ret.refund_status})` : 'N/A'}
                                        </td>
                                        <td>
                                            <div style={{ display: 'flex', gap: '5px' }}>
                                                {ret.status === 'requested' && (
                                                    <button onClick={() => handleApprove(ret.id)} className="btn btn-sm btn-success">
                                                        Approve
                                                    </button>
                                                )}
                                                {ret.status === 'approved' && ret.refund_status !== 'processed' && (
                                                    <button onClick={() => handleRefund(ret.id)} className="btn btn-sm btn-primary">
                                                        Refund
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
            )}
        </div>
    )
}
