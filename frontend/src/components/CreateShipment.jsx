import React, { useState, useEffect } from 'react'
import { shipments, carriers } from '../services/api'

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
        carrier: 'dhl',
        service_type: 'standard',
        pickup_location_id: '',
        drop_location_id: '',
    })
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)
    const [pickupLocations, setPickupLocations] = useState([])
    const [dropLocations, setDropLocations] = useState([])
    const [searchingLocations, setSearchingLocations] = useState(false)

    useEffect(() => {
        if (formData.carrier === 'fedex') {
            fetchLocations()
        } else {
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
                weight: parseFloat(formData.weight)
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
        <div className="create-shipment">
            <div className="form-header">
                <h1>Create New Shipment</h1>
                <button className="btn btn-secondary" onClick={onCancel}>× Close</button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <form onSubmit={handleSubmit}>
                <fieldset>
                    <legend>Sender Information</legend>
                    <div className="form-group">
                        <label htmlFor="sender_name">Name *</label>
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
                        <label htmlFor="sender_address">Address *</label>
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
                        <label htmlFor="sender_email">Email</label>
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

                <fieldset>
                    <legend>Receiver Information</legend>
                    <div className="form-group">
                        <label htmlFor="receiver_name">Name *</label>
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
                        <label htmlFor="receiver_address">Address *</label>
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
                        <label htmlFor="receiver_email">Email</label>
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

                <fieldset>
                    <legend>Package Details</legend>
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
                                step="0.1"
                                min="0"
                                placeholder="2.5"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="dimensions">Dimensions</label>
                            <input
                                type="text"
                                id="dimensions"
                                name="dimensions"
                                value={formData.dimensions}
                                onChange={handleChange}
                                placeholder="10x10x10 cm"
                            />
                        </div>
                    </div>
                    <div className="form-row">
                        <div className="form-group">
                            <label htmlFor="carrier">Carrier *</label>
                            <select
                                id="carrier"
                                name="carrier"
                                value={formData.carrier}
                                onChange={handleChange}
                                required
                            >
                                <option value="dhl">DHL</option>
                                <option value="fedex">FedEx</option>
                                <option value="ups">UPS</option>
                                <option value="usps">USPS</option>
                            </select>
                        </div>
                        <div className="form-group">
                            <label htmlFor="service_type">Service Type *</label>
                            <select
                                id="service_type"
                                name="service_type"
                                value={formData.service_type}
                                onChange={handleChange}
                                required
                            >
                                <option value="standard">Standard</option>
                                <option value="express">Express</option>
                                <option value="overnight">Overnight</option>
                                <option value="economy">Economy</option>
                            </select>
                        </div>
                    </div>
                </fieldset>

                {formData.carrier === 'fedex' && (
                    <fieldset className="fedex-locations">
                        <legend>FedEx Locations</legend>
                        <div className="form-row">
                            <div className="form-group">
                                <label htmlFor="pickup_location_id">Select Pickup Location (Optional)</label>
                                <select
                                    id="pickup_location_id"
                                    name="pickup_location_id"
                                    value={formData.pickup_location_id}
                                    onChange={handleChange}
                                >
                                    <option value="">-- Use Sender Address --</option>
                                    {pickupLocations.map(loc => (
                                        <option key={loc.id} value={loc.id}>
                                            {loc.name} ({loc.address}, {loc.city})
                                        </option>
                                    ))}
                                </select>
                                {searchingLocations && <small>Searching...</small>}
                            </div>
                            <div className="form-group">
                                <label htmlFor="drop_location_id">Select Drop Location (Optional)</label>
                                <select
                                    id="drop_location_id"
                                    name="drop_location_id"
                                    value={formData.drop_location_id}
                                    onChange={handleChange}
                                >
                                    <option value="">-- Deliver to Address --</option>
                                    {dropLocations.map(loc => (
                                        <option key={loc.id} value={loc.id}>
                                            {loc.name} ({loc.address}, {loc.city})
                                        </option>
                                    ))}
                                </select>
                                {searchingLocations && <small>Searching...</small>}
                            </div>
                        </div>
                    </fieldset>
                )}

                <div className="form-actions">
                    <button type="button" className="btn btn-secondary" onClick={onCancel}>
                        Cancel
                    </button>
                    <button type="submit" className="btn btn-primary" disabled={loading}>
                        {loading ? 'Creating...' : 'Create Shipment'}
                    </button>
                </div>
            </form>
        </div>
    )
}
