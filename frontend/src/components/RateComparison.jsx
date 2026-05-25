import React, { useState, useEffect } from 'react'
import { shipments, rates } from '../services/api'

export default function RateComparison() {
    const [shipmentList, setShipmentList] = useState([])
    const [selectedShipmentId, setSelectedShipmentId] = useState('')
    const [fromAddress, setFromAddress] = useState('')
    const [toAddress, setToAddress] = useState('')
    const [weight, setWeight] = useState(1.0)
    
    const [loadingShipments, setLoadingShipments] = useState(true)
    const [comparing, setComparing] = useState(false)
    const [results, setResults] = useState(null)
    const [error, setError] = useState(null)

    useEffect(() => {
        loadShipments()
    }, [])

    const loadShipments = async () => {
        try {
            setLoadingShipments(true)
            const data = await shipments.list()
            setShipmentList(data || [])
            setError(null)
        } catch (err) {
            setError('Failed to load shipments: ' + err.message)
        } finally {
            setLoadingShipments(false)
        }
    }

    const handleShipmentSelect = (e) => {
        const id = e.target.value
        setSelectedShipmentId(id)
        
        if (!id) {
            setFromAddress('')
            setToAddress('')
            setWeight(1.0)
            return
        }

        const selected = shipmentList.find(s => s.id === id)
        if (selected) {
            setFromAddress(selected.sender_address || '')
            setToAddress(selected.receiver_address || '')
            setWeight(selected.weight || 1.0)
        }
    }

    const handleCompare = async (e) => {
        e.preventDefault()
        if (!fromAddress || !toAddress || !weight) {
            setError('Please fill in all comparison fields')
            return
        }

        try {
            setComparing(true)
            setResults(null)
            setError(null)
            
            // If no shipment was selected, generate a dummy ID for the comparison record
            const idToUse = selectedShipmentId || `TEMP-${Date.now()}`

            const data = await rates.compare({
                shipment_id: idToUse,
                from: fromAddress,
                to: toAddress,
                weight: parseFloat(weight)
            })

            // Parse all_rates_json if present
            let parsedRates = []
            if (data.all_rates_json) {
                try {
                    parsedRates = JSON.parse(data.all_rates_json)
                } catch (e) {
                    console.error('Failed to parse all_rates_json', e)
                }
            }

            setResults({
                best: {
                    carrier: data.best_carrier,
                    service: data.best_service,
                    cost: data.best_cost,
                    days: data.best_days
                },
                allRates: parsedRates
            })
        } catch (err) {
            setError('Comparison failed: ' + err.message)
        } finally {
            setComparing(false)
        }
    }

    const getCarrierColor = (carrierName) => {
        const name = carrierName.toLowerCase()
        if (name.includes('dhl')) return { border: '#e5c100', bg: '#fffdeb', text: '#d32f2f' }
        if (name.includes('fedex')) return { border: '#4d148c', bg: '#fbf5ff', text: '#ff6600' }
        if (name.includes('ups')) return { border: '#351c15', bg: '#faf6f4', text: '#ffb500' }
        return { border: '#3b82f6', bg: '#f0f7ff', text: '#2563eb' }
    }

    return (
        <div className="rate-comparison" style={{ maxWidth: '1100px' }}>
            <div className="list-header" style={{ marginBottom: '24px' }}>
                <div>
                    <h1>Compare Carrier Rates</h1>
                    <p style={{ color: '#6b7280', margin: '4px 0 0 0' }}>
                        Calculate and compare shipping costs across different carriers in real-time
                    </p>
                </div>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '24px', alignItems: 'start' }}>
                {/* Form Section */}
                <div className="detail-section" style={{ backgroundColor: 'white', padding: '24px', borderRadius: '12px', border: '1px solid #e5e7eb' }}>
                    <h2 style={{ fontSize: '18px', margin: '0 0 20px 0', borderBottom: '2px solid #f3f4f6', paddingBottom: '12px' }}>
                        Shipping Parameters
                    </h2>
                    
                    <form onSubmit={handleCompare} style={{ padding: '0', border: 'none' }}>
                        <div className="form-group">
                            <label>Pre-fill from Active Shipment (Optional)</label>
                            <select 
                                value={selectedShipmentId} 
                                onChange={handleShipmentSelect}
                                disabled={loadingShipments || comparing}
                            >
                                <option value="">-- Manual Calculation --</option>
                                {shipmentList.map(s => (
                                    <option key={s.id} value={s.id}>
                                        {s.sender_name} ➔ {s.receiver_name} ({s.id.substring(0,8)}...)
                                    </option>
                                ))}
                            </select>
                            {loadingShipments && <small>Loading shipments...</small>}
                        </div>

                        <div className="form-group">
                            <label>Origin Address / Location</label>
                            <input 
                                type="text" 
                                placeholder="e.g. Seattle, WA" 
                                value={fromAddress}
                                onChange={(e) => setFromAddress(e.target.value)}
                                required
                                disabled={comparing}
                            />
                        </div>

                        <div className="form-group">
                            <label>Destination Address / Location</label>
                            <input 
                                type="text" 
                                placeholder="e.g. Miami, FL" 
                                value={toAddress}
                                onChange={(e) => setToAddress(e.target.value)}
                                required
                                disabled={comparing}
                            />
                        </div>

                        <div className="form-group">
                            <label>Weight (kg)</label>
                            <input 
                                type="number" 
                                step="0.01" 
                                min="0.05"
                                placeholder="e.g. 1.2" 
                                value={weight}
                                onChange={(e) => setWeight(e.target.value)}
                                required
                                disabled={comparing}
                            />
                        </div>

                        <button 
                            type="submit" 
                            className="btn btn-primary" 
                            style={{ width: '100%', justifyContent: 'center', marginTop: '16px' }}
                            disabled={comparing}
                        >
                            {comparing ? 'Calculating Rates...' : '🔍 Compare Rates'}
                        </button>
                    </form>
                </div>

                {/* Results Section */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    {comparing && (
                        <div className="empty-state" style={{ padding: '64px 32px' }}>
                            <div className="loading" style={{ padding: '0' }}>
                                🌀 Fetching live carrier rates...
                            </div>
                        </div>
                    )}

                    {!comparing && !results && (
                        <div className="empty-state" style={{ padding: '64px 32px' }}>
                            <p style={{ fontSize: '16px', fontWeight: '500', margin: '0 0 8px 0' }}>No comparison calculated yet</p>
                            <p style={{ fontSize: '13px', color: '#9ca3af', margin: '0' }}>
                                Enter details or select an existing shipment, then click Compare Rates to fetch carrier options.
                            </p>
                        </div>
                    )}

                    {!comparing && results && (
                        <>
                            {/* Best Rate Highlight */}
                            <div 
                                className="stat-card" 
                                style={{ 
                                    background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)', 
                                    color: 'white', 
                                    border: 'none',
                                    boxShadow: '0 8px 20px rgba(16, 185, 129, 0.25)' 
                                }}
                            >
                                <span style={{ fontSize: '12px', textTransform: 'uppercase', fontWeight: '700', letterSpacing: '0.05em', backgroundColor: 'rgba(255, 255, 255, 0.2)', padding: '4px 10px', borderRadius: '12px', display: 'inline-block', marginBottom: '12px' }}>
                                    Cheapest & Recommended Option
                                </span>
                                <div style={{ fontSize: '36px', fontWeight: '800', marginBottom: '4px' }}>
                                    ${results.best.cost.toFixed(2)}
                                </div>
                                <div style={{ fontSize: '18px', fontWeight: '600', textTransform: 'capitalize' }}>
                                    {results.best.carrier} — {results.best.service}
                                </div>
                                <p style={{ fontSize: '14px', margin: '8px 0 0 0', opacity: '0.9' }}>
                                    Estimated Delivery: <strong>{results.best.days} {results.best.days === 1 ? 'day' : 'days'}</strong>
                                </p>
                            </div>

                            {/* Comparison Grid */}
                            <div className="detail-section" style={{ backgroundColor: 'white', padding: '24px', borderRadius: '12px', border: '1px solid #e5e7eb' }}>
                                <h2 style={{ fontSize: '16px', margin: '0 0 16px 0', borderBottom: '2px solid #f3f4f6', paddingBottom: '12px' }}>
                                    All Carrier Quotes ({results.allRates.length})
                                </h2>
                                
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                                    {results.allRates.map((rate, idx) => {
                                        const brand = getCarrierColor(rate.carrier_name)
                                        const isBest = rate.carrier_name === results.best.carrier && rate.service_type === results.best.service

                                        return (
                                            <div 
                                                key={idx}
                                                style={{
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'space-between',
                                                    padding: '16px',
                                                    borderRadius: '8px',
                                                    border: `2px solid ${isBest ? '#10b981' : '#e5e7eb'}`,
                                                    background: isBest ? '#f0fdf4' : '#fff',
                                                    transition: 'all 0.2s',
                                                    position: 'relative'
                                                }}
                                            >
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                                                    {/* Custom Carrier Logo Simulation */}
                                                    <div 
                                                        style={{
                                                            width: '56px',
                                                            height: '40px',
                                                            borderRadius: '6px',
                                                            border: `1px solid ${brand.border}`,
                                                            backgroundColor: brand.bg,
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            justifyContent: 'center',
                                                            fontWeight: '800',
                                                            fontSize: '11px',
                                                            color: brand.text,
                                                            textTransform: 'uppercase'
                                                        }}
                                                    >
                                                        {rate.carrier_name.substring(0, 4)}
                                                    </div>
                                                    
                                                    <div>
                                                        <div style={{ fontWeight: '700', color: '#1f2937', display: 'flex', alignItems: 'center', gap: '6px' }}>
                                                            {rate.carrier_name}
                                                            {isBest && (
                                                                <span style={{ fontSize: '10px', color: '#10b981', background: '#dcfce7', padding: '2px 6px', borderRadius: '10px', fontWeight: '600' }}>
                                                                    Cheapest
                                                                </span>
                                                            )}
                                                        </div>
                                                        <div style={{ fontSize: '13px', color: '#6b7280', textTransform: 'capitalize' }}>
                                                            {rate.service_type} • {rate.estimated_days} {rate.estimated_days === 1 ? 'day' : 'days'}
                                                        </div>
                                                    </div>
                                                </div>

                                                <div style={{ textAlign: 'right' }}>
                                                    <div style={{ fontSize: '20px', fontWeight: '800', color: isBest ? '#15803d' : '#1f2937' }}>
                                                        ${rate.cost.toFixed(2)}
                                                    </div>
                                                    <div style={{ fontSize: '11px', color: '#f59e0b' }}>
                                                        {'★'.repeat(Math.round(rate.rating || 5)) + '☆'.repeat(5 - Math.round(rate.rating || 5))}
                                                    </div>
                                                </div>
                                            </div>
                                        )
                                    })}
                                </div>
                            </div>
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}
