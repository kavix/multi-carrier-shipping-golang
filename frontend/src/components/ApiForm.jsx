import React, { useState } from 'react'

export default function ApiForm({ title, method = 'GET', path = '/', baseUrl, token, defaultBody }) {
  const [response, setResponse] = useState(null)
  const [loading, setLoading] = useState(false)
  const [bodyText, setBodyText] = useState(defaultBody ? JSON.stringify(defaultBody, null, 2) : '')

  const send = async () => {
    setLoading(true)
    setResponse(null)
    try {
      const url = baseUrl.replace(/\/$/, '') + path
      const opts = { method }
      const headers = {}
      if (token) headers['Authorization'] = token
      if (method !== 'GET' && bodyText) {
        headers['Content-Type'] = 'application/json'
        opts.body = bodyText
      }
      opts.headers = headers

      const res = await fetch(url, opts)
      const contentType = res.headers.get('content-type') || ''
      let data
      if (contentType.includes('application/json')) {
        data = await res.json()
      } else {
        data = await res.text()
      }
      setResponse({ status: res.status, ok: res.ok, data })
    } catch (err) {
      setResponse({ error: err.message })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="api-form">
      <div className="api-row">
        <div className="meta">
          <strong>{title}</strong>
          <div className="muted">{method} {path}</div>
        </div>
        <div className="actions">
          <button onClick={send} disabled={loading}>{loading ? 'Sending...' : 'Send'}</button>
        </div>
      </div>
      {(method !== 'GET') && (
        <div>
          <textarea value={bodyText} onChange={e => setBodyText(e.target.value)} rows={6} />
        </div>
      )}

      <div className="response">
        {response ? (
          <pre>{JSON.stringify(response, null, 2)}</pre>
        ) : (
          <div className="muted">No response yet</div>
        )}
      </div>
    </div>
  )
}

