import React, { useState } from 'react'
import { setAuthToken } from '../services/api'

export default function Settings({ baseUrl, onBaseUrlChange, token, onTokenChange }) {
    const [showToken, setShowToken] = useState(false)

    const handleTokenChange = (e) => {
        const newToken = e.target.value
        onTokenChange(newToken)
        setAuthToken(newToken)
    }

    return (
        <div className="settings">
            <h1>Settings</h1>

            <div className="settings-section">
                <h2>API Configuration</h2>

                <div className="form-group">
                    <label htmlFor="baseUrl">API Base URL</label>
                    <input
                        type="text"
                        id="baseUrl"
                        value={baseUrl}
                        onChange={(e) => onBaseUrlChange(e.target.value)}
                        placeholder="http://localhost:8080"
                    />
                    <small className="text-muted">Default: http://localhost:8080</small>
                </div>

                <div className="form-group">
                    <label htmlFor="token">
                        Authorization Token
                        <button
                            type="button"
                            className="btn-icon"
                            onClick={() => setShowToken(!showToken)}
                        >
                            {showToken ? '🙈' : '👁️'}
                        </button>
                    </label>
                    <input
                        type={showToken ? 'text' : 'password'}
                        id="token"
                        value={token}
                        onChange={handleTokenChange}
                        placeholder="Bearer test-token"
                    />
                    <small className="text-muted">Used for API authentication</small>
                </div>

                <div className="info-box">
                    <strong>ℹ️ Current Configuration</strong>
                    <p>Base URL: <code>{baseUrl}</code></p>
                    <p>Token: <code>{token ? token.substring(0, 20) + '...' : 'Not set'}</code></p>
                </div>
            </div>

            <div className="settings-section">
                <h2>About</h2>
                <div className="info-box">
                    <p><strong>Multi-Carrier Shipping Platform</strong></p>
                    <p>Version 1.0.0</p>
                    <p>Frontend built with React and Vite</p>
                    <p><a href="#" onClick={(e) => { e.preventDefault(); window.open('/docs/API-GUIDE.md') }}>API Documentation</a></p>
                </div>
            </div>
        </div>
    )
}
