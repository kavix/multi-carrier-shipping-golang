import React, { useState } from 'react'
import { addresses, carriers } from '../services/api'

export default function AddressTools() {
    const [validation, setValidation] = useState({
        address: '',
        result: null,
        loading: false
    })

    const [locations, setLocations] = useState({
        query: '',
        carrier: 'dhl',
        type: 'pickup',
        results: [],
        loading: false
    })

    const handleValidate = async (e) => {
        e.preventDefault()
        setValidation({ ...validation, loading: true, result: null })
        try {
            const res = await addresses.validate({ address: validation.address })
            setValidation({ ...validation, loading: false, result: res })
        } catch (err) {
            setValidation({ ...validation, loading: false, result: { error: err.message } })
        }
    }

    const handleSearchLocations = async (e) => {
        e.preventDefault()
        setLocations({ ...locations, loading: true, results: [] })
        try {
            let res
            if (locations.type === 'pickup') {
                res = await carriers.getPickupLocations(locations.carrier, locations.query)
            } else {
                res = await carriers.getDropLocations(locations.carrier, locations.query)
            }
            setLocations({ ...locations, loading: false, results: res || [] })
        } catch (err) {
            setLocations({ ...locations, loading: false, results: [] })
            alert('Location search failed: ' + err.message)
        }
    }

    return (
        <div className="address-tools">
            <h1>Address & Location Tools</h1>
            <p className="subtitle">Validate addresses and discover carrier service points</p>

            <div className="detail-grid">
                <div className="detail-section">
                    <h2>Address Validation</h2>
                    <form onSubmit={handleValidate}>
                        <div className="form-group">
                            <label>Full Address String</label>
                            <textarea 
                                value={validation.address} 
                                onChange={e => setValidation({...validation, address: e.target.value})}
                                placeholder="123 Main St, New York, NY 10001, USA"
                                required
                                style={{ height: '80px', padding: '10px' }}
                            />
                        </div>
                        <button className="btn btn-primary" type="submit" disabled={validation.loading}>
                            {validation.loading ? 'Validating...' : 'Validate Address'}
                        </button>
                    </form>

                    {validation.result && (
                        <div className="response" style={{ marginTop: '20px' }}>
                            <pre>{JSON.stringify(validation.result, null, 2)}</pre>
                        </div>
                    )}
                </div>

                <div className="detail-section">
                    <h2>Location Finder</h2>
                    <form onSubmit={handleSearchLocations}>
                        <div className="form-row">
                            <div className="form-group">
                                <label>Carrier</label>
                                <select value={locations.carrier} onChange={e => setLocations({...locations, carrier: e.target.value})}>
                                    <option value="dhl">DHL</option>
                                    <option value="ups">UPS</option>
                                    <option value="fedex">FedEx</option>
                                </select>
                            </div>
                            <div className="form-group">
                                <label>Type</label>
                                <select value={locations.type} onChange={e => setLocations({...locations, type: e.target.value})}>
                                    <option value="pickup">Pickup Points</option>
                                    <option value="drop">Drop-off Points</option>
                                </select>
                            </div>
                        </div>
                        <div className="form-group">
                            <label>Near Address/City</label>
                            <input 
                                value={locations.query} 
                                onChange={e => setLocations({...locations, query: e.target.value})}
                                placeholder="New York, NY"
                                required
                            />
                        </div>
                        <button className="btn btn-secondary" type="submit" disabled={locations.loading}>
                            {locations.loading ? 'Searching...' : 'Find Locations'}
                        </button>
                    </form>

                    <div className="location-results" style={{ marginTop: '20px', maxHeight: '300px', overflowY: 'auto' }}>
                        {locations.results.length > 0 ? (
                            <ul style={{ listStyle: 'none', padding: 0 }}>
                                {locations.results.map((loc, i) => (
                                    <li key={i} style={{ padding: '12px', borderBottom: '1px solid #f3f4f6', fontSize: '0.9rem' }}>
                                        <div style={{ fontWeight: 600 }}>{loc.name || loc.id}</div>
                                        <div className="muted">{loc.address}</div>
                                        {loc.distance && <div style={{ fontSize: '0.8rem', color: '#3b82f6' }}>{loc.distance} km away</div>}
                                    </li>
                                ))}
                            </ul>
                        ) : locations.loading ? null : (
                            <p className="muted text-center">No locations found or search not started.</p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}
