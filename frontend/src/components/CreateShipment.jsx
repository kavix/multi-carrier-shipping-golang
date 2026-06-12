import React, { useState, useEffect } from 'react'
import { shipments, carriers, addresses } from '../services/api'

export default function CreateShipment({ onSuccess, onCancel }) {
    const [formData, setFormData] = useState({
        sender_name: '',
        sender_address: '',
        sender_email: '',
        receiver_name: '',
        receiver_address: '',
        receiver_email: '',
        weight: '',
        dimensions: '',
        description: '',
        carrier: 'dhl',
        service_type: 'standard',
        pickup_location_id: '',
        drop_location_id: '',
        is_international: false,
        customs_value: '',
        customs_currency: 'USD',
    })
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)
    const [pickupLocations, setPickupLocations] = useState([])
    const [dropLocations, setDropLocations] = useState([])
    const [searchingLocations, setSearchingLocations] = useState(false)
    const [isValidating, setIsValidating] = useState({ sender: false, receiver: false })

    const validateAddress = async (type) => {
        const addr = type === 'sender' ? formData.sender_address : formData.receiver_address
        if (!addr) return

        try {
            setIsValidating(prev => ({ ...prev, [type]: true }))
            const result = await addresses.validate({ address: addr })
            if (result.is_valid) {
                // Update address with standardized version if available
                const standardized = `${result.street}, ${result.city}, ${result.state} ${result.postal_code}, ${result.country}`
                setFormData(prev => ({
                    ...prev,
                    [type === 'sender' ? 'sender_address' : 'receiver_address']: standardized
                }))
                alert(`Address validated successfully! Standardized to: ${standardized}`)
            } else {
                alert('Address could not be validated. Please check the details.')
            }
        } catch (err) {
            alert('Validation error: ' + err.message)
        } finally {
            setIsValidating(prev => ({ ...prev, [type]: false }))
        }
    }

    useEffect(() => {
        if (formData.carrier === 'fedex') {
            setFormData(prev => {
                if (prev.service_type !== 'FEDEX_GROUND' && prev.service_type !== 'FEDEX_EXPRESS_SAVER' && prev.service_type !== 'STANDARD_OVERNIGHT' && prev.service_type !== 'INTERNATIONAL_PRIORITY') {
                    return { ...prev, service_type: 'FEDEX_GROUND' }
                }
                return prev
            })
            fetchLocations()
        } else {
            setFormData(prev => {
                if (prev.service_type === 'FEDEX_GROUND' || prev.service_type === 'FEDEX_EXPRESS_SAVER' || prev.service_type === 'STANDARD_OVERNIGHT' || prev.service_type === 'INTERNATIONAL_PRIORITY') {
                    return { ...prev, service_type: 'standard' }
                }
                return prev
            })
            setPickupLocations([])
            setDropLocations([])
        }
    }, [formData.carrier, formData.sender_address, formData.receiver_address])

    const fetchLocations = async () => {
        if (formData.carrier !== 'fedex') return

        try {
            setSearchingLocations(true)

            // Only fetch if addresses are provided
            const promises = []
            if (formData.sender_address) {
                promises.push(carriers.getPickupLocations('fedex', formData.sender_address, 5)
                    .then(locs => setPickupLocations(locs))
                    .catch(e => console.error("Pickup fetch error", e)))
            }
            if (formData.receiver_address) {
                promises.push(carriers.getDropLocations('fedex', formData.receiver_address, 5)
                    .then(locs => setDropLocations(locs))
                    .catch(e => console.error("Drop fetch error", e)))
            }

            await Promise.all(promises)
        } catch (err) {
            console.error("Error fetching Fedora locations", err)
        } finally {
            setSearchingLocations(false)
        }
    }

    const handleChange = (e) => {
        const { name, value } = e.target
        setFormData(prev => ({
            ...prev,
            [name]: value
        }))
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        try {
            setLoading(true)
            setError(null)

            const payload = {
                ...formData,
                weight: parseFloat(formData.weight),
                customs_value: formData.customs_value ? parseFloat(formData.customs_value) : 0
            }

            const result = await shipments.create(payload)
            onSuccess(result)
        } catch (err) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="create-shipment" style={{ maxWidth: '850px' }}>
            <div className="list-header">
                <div>
                    <h1>Create New Shipment</h1>
                    <p className="subtitle">Dispatch parcels through integrated global carrier gateways</p>
                </div>
                <button className="btn btn-secondary" onClick={onCancel}>× Close</button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <form onSubmit={handleSubmit} className="card" style={{ padding: '32px' }}>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '32px' }}>
                    <fieldset style={{ borderBottom: 'none', paddingBottom: 0 }}>
                        <legend style={{ fontSize: '18px', borderBottom: '2.5px solid var(--border-color)', paddingBottom: '8px', marginBottom: '20px' }}>
                            👤 Sender Details
                        </legend>
                        
                        <div className="form-group">
                            <label htmlFor="sender_name">Full Name *</label>
                            <input
                                type="text"
                                id="sender_name"
                                name="sender_name"
                                value={formData.sender_name}
                                onChange={handleChange}
                                required
                                placeholder="John Doe"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="sender_address" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                Address *
                                <button 
                                    type="button" 
                                    style={{ color: 'var(--primary)', fontSize: '12px', border: 'none', background: 'none', cursor: 'pointer', fontWeight: 600 }}
                                    onClick={() => validateAddress('sender')}
                                    disabled={isValidating.sender}
                                >
                                    {isValidating.sender ? 'Validating...' : '✨ Standardize Address'}
                                </button>
                            </label>
                            <input
                                type="text"
                                id="sender_address"
                                name="sender_address"
                                value={formData.sender_address}
                                onChange={handleChange}
                                required
                                placeholder="123 Main St, New York, NY 10001"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="sender_email">Email Address</label>
                            <input
                                type="email"
                                id="sender_email"
                                name="sender_email"
                                value={formData.sender_email}
                                onChange={handleChange}
                                placeholder="john@example.com"
                            />
                        </div>
                    </fieldset>

                    <fieldset style={{ borderBottom: 'none', paddingBottom: 0 }}>
                        <legend style={{ fontSize: '18px', borderBottom: '2.5px solid var(--border-color)', paddingBottom: '8px', marginBottom: '20px' }}>
                            📍 Receiver Details
                        </legend>
                        
                        <div className="form-group">
                            <label htmlFor="receiver_name">Full Name *</label>
                            <input
                                type="text"
                                id="receiver_name"
                                name="receiver_name"
                                value={formData.receiver_name}
                                onChange={handleChange}
                                required
                                placeholder="Jane Smith"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="receiver_address" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                Address *
                                <button 
                                    type="button" 
                                    style={{ color: 'var(--primary)', fontSize: '12px', border: 'none', background: 'none', cursor: 'pointer', fontWeight: 600 }}
                                    onClick={() => validateAddress('receiver')}
                                    disabled={isValidating.receiver}
                                >
                                    {isValidating.receiver ? 'Validating...' : '✨ Standardize Address'}
                                </button>
                            </label>
                            <input
                                type="text"
                                id="receiver_address"
                                name="receiver_address"
                                value={formData.receiver_address}
                                onChange={handleChange}
                                required
                                placeholder="456 Oak Ave, Los Angeles, CA 90001"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="receiver_email">Email Address</label>
                            <input
                                type="email"
                                id="receiver_email"
                                name="receiver_email"
                                value={formData.receiver_email}
                                onChange={handleChange}
                                placeholder="jane@example.com"
                            />
                        </div>
                    </fieldset>
                </div>

                <fieldset style={{ marginTop: '24px', borderTop: '1px solid var(--border-color)', paddingTop: '24px' }}>
                    <legend style={{ fontSize: '18px', marginBottom: '20px' }}>📦 Package & Carrier Parameters</legend>
                    
                    <div className="form-group">
                        <label htmlFor="description">Item Description *</label>
                        <input
                            type="text"
                            id="description"
                            name="description"
                            value={formData.description}
                            onChange={handleChange}
                            required
                            placeholder="e.g. Books, Electronics, Clothing"
                        />
                    </div>
                    
                    <div className="form-row">
                        <div className="form-group">
                            <label htmlFor="weight">Weight (kg) *</label>
                            <input
                                type="number"
                                id="weight"
                                name="weight"
                                value={formData.weight}
                                onChange={handleChange}
                                required
                                step="0.01"
                                placeholder="0.50"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="dimensions">Dimensions (LxWxH cm)</label>
                            <input
                                type="text"
                                id="dimensions"
                                name="dimensions"
                                value={formData.dimensions}
                                onChange={handleChange}
                                placeholder="e.g. 20x15x10"
                            />
                        </div>
                    </div>

                    <div className="form-row" style={{ marginTop: '10px' }}>
                        <div className="form-group">
                            <label htmlFor="carrier">Select Carrier *</label>
                            <select
                                id="carrier"
                                name="carrier"
                                value={formData.carrier}
                                onChange={handleChange}
                                required
                            >
                                <option value="dhl">DHL Express</option>
                                <option value="fedex">FedEx Corporation</option>
                                <option value="ups">United Parcel Service</option>
                            </select>
                        </div>
                        <div className="form-group">
                            <label htmlFor="service_type">Select Service Type *</label>
                            <select
                                id="service_type"
                                name="service_type"
                                value={formData.service_type}
                                onChange={handleChange}
                                required
                            >
                                {formData.carrier === 'fedex' ? (
                                    <>
                                        <option value="FEDEX_GROUND">FedEx Ground</option>
                                        <option value="FEDEX_EXPRESS_SAVER">FedEx Express Saver</option>
                                        <option value="STANDARD_OVERNIGHT">Standard Overnight</option>
                                        <option value="INTERNATIONAL_PRIORITY">International Priority</option>
                                    </>
                                ) : (
                                    <>
                                        <option value="standard">Standard / Ground Delivery</option>
                                        <option value="express">Express / Air Delivery</option>
                                        <option value="overnight">Overnight Delivery</option>
                                        <option value="economy">Economy Saver</option>
                                    </>
                                )}
                            </select>
                        </div>
                    </div>

                    {formData.carrier === 'fedex' && (
                        <div className="form-row" style={{ marginTop: '16px', padding: '20px', backgroundColor: '#f8fafc', borderRadius: '12px', border: '1.5px solid var(--border-color)' }}>
                            <div className="form-group">
                                <label htmlFor="account_number">FedEx Account Number</label>
                                <input
                                    type="text"
                                    id="account_number"
                                    name="account_number"
                                    value={formData.account_number || ''}
                                    onChange={handleChange}
                                    placeholder="e.g. 740561073"
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="packaging_type">Packaging Type</label>
                                <select
                                    id="packaging_type"
                                    name="packaging_type"
                                    value={formData.packaging_type || 'YOUR_PACKAGING'}
                                    onChange={handleChange}
                                >
                                    <option value="YOUR_PACKAGING">Customer Packaging</option>
                                    <option value="FEDEX_ENVELOPE">FedEx Envelope</option>
                                    <option value="FEDEX_BOX">FedEx Box</option>
                                    <option value="FEDEX_PAK">FedEx Pak</option>
                                </select>
                            </div>
                        </div>
                    )}

                    <div className="form-group" style={{ marginTop: '20px' }}>
                        <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer', fontWeight: 600 }}>
                            <input
                                type="checkbox"
                                name="is_international"
                                checked={formData.is_international}
                                onChange={e => setFormData(prev => ({ ...prev, is_international: e.target.checked }))}
                                style={{ width: '18px', height: '18px', accentColor: 'var(--primary)' }}
                            />
                            This is an International Shipment (Requires Customs Declarations)
                        </label>
                    </div>

                    {formData.is_international && (
                        <div className="form-row" style={{ marginTop: '12px', padding: '20px', backgroundColor: 'var(--info-bg)', borderRadius: '12px', border: '1px solid var(--info-border)' }}>
                            <div className="form-group">
                                <label htmlFor="customs_value">Total Customs Value *</label>
                                <input
                                    type="number"
                                    id="customs_value"
                                    name="customs_value"
                                    value={formData.customs_value}
                                    onChange={handleChange}
                                    required={formData.is_international}
                                    placeholder="100.00"
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="customs_currency">Customs Currency</label>
                                <select
                                    id="customs_currency"
                                    name="customs_currency"
                                    value={formData.customs_currency}
                                    onChange={handleChange}
                                >
                                    <option value="USD">USD ($)</option>
                                    <option value="EUR">EUR (€)</option>
                                    <option value="GBP">GBP (£)</option>
                                    <option value="LKR">LKR (₨)</option>
                                </select>
                            </div>
                        </div>
                    )}
                </fieldset>

                {formData.carrier === 'fedex' && (
                    <fieldset style={{ borderBottom: 'none', paddingBottom: 0 }}>
                        <legend style={{ fontSize: '18px', marginBottom: '20px' }}>🏢 FedEx Locations Lookup</legend>
                        <div className="form-row">
                            <div className="form-group">
                                <label htmlFor="pickup_location_id">Select Local Pickup Point (Optional)</label>
                                <select
                                    id="pickup_location_id"
                                    name="pickup_location_id"
                                    value={formData.pickup_location_id}
                                    onChange={handleChange}
                                >
                                    <option value="">-- Drop off/Dispatch at Sender Address --</option>
                                    {pickupLocations.map(loc => (
                                        <option key={loc.id} value={loc.id}>
                                            {loc.name} ({loc.address}, {loc.city})
                                        </option>
                                    ))}
                                </select>
                                {searchingLocations && <small style={{ color: 'var(--primary)', fontWeight: 600 }}>Searching local FedEx pickup terminals...</small>}
                            </div>
                            <div className="form-group">
                                <label htmlFor="drop_location_id">Select Local Dropoff Terminal (Optional)</label>
                                <select
                                    id="drop_location_id"
                                    name="drop_location_id"
                                    value={formData.drop_location_id}
                                    onChange={handleChange}
                                >
                                    <option value="">-- Deliver Directly to Recipient Address --</option>
                                    {dropLocations.map(loc => (
                                        <option key={loc.id} value={loc.id}>
                                            {loc.name} ({loc.address}, {loc.city})
                                        </option>
                                    ))}
                                </select>
                                {searchingLocations && <small style={{ color: 'var(--primary)', fontWeight: 600 }}>Searching local FedEx dropoff terminals...</small>}
                            </div>
                        </div>
                    </fieldset>
                )}

                <div className="form-actions">
                    <button type="button" className="btn btn-secondary" onClick={onCancel}>
                        Cancel
                    </button>
                    <button type="submit" className="btn btn-primary" disabled={loading}>
                        {loading ? 'Processing manifest...' : '🚀 Create Shipment'}
                    </button>
                </div>
            </form>
        </div>
    )
}
