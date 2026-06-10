import React, { useState } from 'react'
import { carriers } from '../services/api'

export default function CarrierManager() {
    const [formData, setFormData] = useState({
        name: '',
        code: '',
        api_key: '',
        api_secret: '',
        base_url: ''
    })
    const [loading, setLoading] = useState(false)
    const [message, setMessage] = useState(null)
    const [postalValidation, setPostalValidation] = useState({
        carrier: '',
        country: '',
        postalCode: '',
        result: null
    })

    const handleChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value })
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        setLoading(true)
        setMessage(null)
        try {
            await carriers.register(formData)
            setMessage({ type: 'success', text: 'Carrier registered successfully!' })
            setFormData({ name: '', code: '', api_key: '', api_secret: '', base_url: '' })
        } catch (err) {
            setMessage({ type: 'error', text: err.message })
        } finally {
            setLoading(false)
        }
    }

    const handlePostalCheck = async (e) => {
        e.preventDefault()
        try {
            const res = await carriers.validatePostalCode(
                postalValidation.carrier,
                postalValidation.country,
                postalValidation.postalCode
            )
            setPostalValidation({ ...postalValidation, result: res.valid ? 'Valid' : 'Invalid' })
        } catch (err) {
            setPostalValidation({ ...postalValidation, result: 'Error: ' + err.message })
        }
    }

    return (
        <div className="carrier-manager">
            <h1>Carrier Management</h1>
            <p className="subtitle">Register and configure shipping carriers</p>

            <div className="detail-grid">
                <div className="detail-section">
                    <h2>Register New Carrier</h2>
                    {message && (
                        <div className={`alert alert-\${message.type === 'success' ? 'success' : 'error'}`}>
                            {message.text}
                        </div>
                    )}
                    <form onSubmit={handleSubmit}>
                        <div className="form-group">
                            <label>Carrier Name</label>
                            <input name="name" value={formData.name} onChange={handleChange} placeholder="e.g. DHL Express" required />
                        </div>
                        <div className="form-group">
                            <label>Carrier Code</label>
                            <input name="code" value={formData.code} onChange={handleChange} placeholder="e.g. dhl" required />
                        </div>
                        <div className="form-group">
                            <label>API Key</label>
                            <input name="api_key" value={formData.api_key} onChange={handleChange} required />
                        </div>
                        <div className="form-group">
                            <label>API Secret</label>
                            <input name="api_secret" value={formData.api_secret} onChange={handleChange} required />
                        </div>
                        <div className="form-group">
                            <label>Base URL</label>
                            <input name="base_url" value={formData.base_url} onChange={handleChange} placeholder="https://api.carrier.com" required />
                        </div>
                        <button className="btn btn-primary" type="submit" disabled={loading}>
                            {loading ? 'Registering...' : 'Register Carrier'}
                        </button>
                    </form>
                </div>

                <div className="detail-section">
                    <h2>Postal Code Validation</h2>
                    <p className="muted" style={{ marginBottom: '16px' }}>Verify if a postal code is supported by a specific carrier</p>
                    <form onSubmit={handlePostalCheck}>
                        <div className="form-group">
                            <label>Carrier Code</label>
                            <input 
                                value={postalValidation.carrier} 
                                onChange={e => setPostalValidation({...postalValidation, carrier: e.target.value})} 
                                placeholder="dhl" required 
                            />
                        </div>
                        <div className="form-group">
                            <label>Country Code (ISO)</label>
                            <input 
                                value={postalValidation.country} 
                                onChange={e => setPostalValidation({...postalValidation, country: e.target.value})} 
                                placeholder="US" required 
                            />
                        </div>
                        <div className="form-group">
                            <label>Postal Code</label>
                            <input 
                                value={postalValidation.postalCode} 
                                onChange={e => setPostalValidation({...postalValidation, postalCode: e.target.value})} 
                                placeholder="90210" required 
                            />
                        </div>
                        <button className="btn btn-secondary" type="submit">Check Support</button>
                    </form>
                    {postalValidation.result && (
                        <div className="info-box" style={{ marginTop: '20px' }}>
                            <strong>Result:</strong> {postalValidation.result}
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
