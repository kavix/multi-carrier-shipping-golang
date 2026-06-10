import React, { useState } from 'react'
import { shipments, carriers, rates, labels, addresses, billing, returns, health } from '../services/api'

const SERVICES = [
    { name: 'API Gateway', url: 'http://localhost:8080' },
    { name: 'Shipment Service', url: 'http://localhost:8001' },
    { name: 'Carrier Service', url: 'http://localhost:8002' },
    { name: 'Rate Service', url: 'http://localhost:8003' },
    { name: 'Label Service', url: 'http://localhost:8004' },
    { name: 'Tracking Service', url: 'http://localhost:8005' },
    { name: 'Address Service', url: 'http://localhost:8006' },
    { name: 'Billing Service', url: 'http://localhost:8007' },
    { name: 'Return Service', url: 'http://localhost:8008' },
]

export default function TestRunner() {
    const [logs, setLogs] = useState([])
    const [loading, setLoading] = useState(false)
    const [healthStatus, setHealthStatus] = useState({})
    const [stressCount, setStressCount] = useState(5)

    const addLog = (message, type = 'info') => {
        setLogs(prev => [{ id: Date.now() + Math.random(), message, type, time: new Date().toLocaleTimeString() }, ...prev])
    }

    const clearLogs = () => setLogs([])

    const runHealthChecks = async () => {
        setLoading(true)
        addLog('Starting health checks...')
        const newStatus = {}
        
        for (const service of SERVICES) {
            try {
                const res = await health.check(service.url)
                newStatus[service.name] = { status: 'ok', data: res }
                addLog(`✓ ${service.name} is healthy`, 'success')
            } catch (err) {
                newStatus[service.name] = { status: 'error', error: err.message }
                addLog(`✗ ${service.name} failed: ${err.message}`, 'error')
            }
        }
        setHealthStatus(newStatus)
        setLoading(false)
    }

    const runDemoWorkflow = async () => {
        setLoading(true)
        clearLogs()
        addLog('🚀 Starting Demo Workflow...')

        try {
            // 1. Create Shipment
            addLog('Step 1: Creating shipment...')
            const shipmentResp = await shipments.create({
                sender_name: "Alice Smith",
                sender_address: "123 Maple St, Seattle, WA 98101",
                sender_phone: "555-0123",
                sender_email: "alice@example.com",
                receiver_name: "Bob Jones",
                receiver_address: "456 Pine Rd, Miami, FL 33101",
                receiver_phone: "555-0456",
                receiver_email: "bob@example.com",
                weight: 1.2,
                dimensions: "8x6x4",
                description: "Books",
                carrier: "ups",
                service_type: "ground"
            })
            const shipmentId = shipmentResp.id
            addLog(`✓ Shipment created: ${shipmentId}`, 'success')

            // 1b. Register Carrier
            addLog('Step 1b: Registering carrier...')
            await carriers.register({
                name: "DHL Express",
                code: "dhl",
                api_key: "test-key",
                api_secret: "test-secret",
                base_url: "http://simulated-dhl"
            })
            addLog('✓ Carrier registered', 'success')

            // 2. Compare Rates
            addLog('Step 2: Comparing rates...')
            const ratesResp = await rates.compare({
                shipment_id: shipmentId,
                from: "Seattle, WA",
                to: "Miami, FL",
                weight: 1.2
            })
            addLog(`✓ Found ${ratesResp.length || 0} rates`, 'success')

            // 3. Generate Label
            addLog('Step 3: Generating label...')
            const labelResp = await labels.generate({
                shipment_id: shipmentId,
                carrier: "ups",
                format: "pdf"
            })
            addLog(`✓ Label generated: ${labelResp.id}`, 'success')

            // 4. Validate Address
            addLog('Step 4: Validating address...')
            await addresses.validate({
                address: "123 Maple St, Seattle, WA 98101"
            })
            addLog('✓ Address validated', 'success')

            // 5. Create Invoice
            addLog('Step 5: Creating invoice...')
            const invResp = await billing.createInvoice({
                shipment_id: shipmentId,
                user_id: "test-user-999",
                amount: 24.50,
                description: "Shipping charges"
            })
            const invId = invResp.id
            addLog(`✓ Invoice created: ${invId}`, 'success')

            // 6. Process Payment
            addLog('Step 6: Processing payment...')
            await billing.processPayment({
                invoice_id: invId,
                method: "credit_card"
            })
            addLog('✓ Payment processed', 'success')

            // 7. Request Return
            addLog('Step 7: Requesting return...')
            await returns.create({
                shipment_id: shipmentId,
                reason: "Damaged on arrival"
            })
            addLog('✓ Return requested', 'success')

            addLog('✨ Demo workflow completed successfully!', 'success')
        } catch (err) {
            addLog(`❌ Workflow failed: ${err.message}`, 'error')
        }
        setLoading(false)
    }

    const runStressTest = async () => {
        setLoading(true)
        clearLogs()
        addLog(`🏃 Starting Stress Test (\${stressCount} shipments)...`)

        for (let i = 1; i <= stressCount; i++) {
            try {
                const weight = (Math.random() * 50 + 1).toFixed(2)
                
                await shipments.create({
                    sender_name: `Sender_\${Math.floor(Math.random() * 100 + 1)}`,
                    sender_address: "123 Main St, Anytown, USA",
                    sender_email: `sender\${i}@example.com`,
                    receiver_name: `Receiver_\${Math.floor(Math.random() * 100 + 1)}`,
                    receiver_address: "456 Oak Ave, Otherville, USA",
                    receiver_email: `receiver\${i}@example.com`,
                    weight: parseFloat(weight),
                    dimensions: "10x10x10",
                    carrier: "UPS",
                    service_type: "Express"
                })
                addLog(`✓ Added shipment \${i}/\${stressCount}`, 'success')
            } catch (err) {
                addLog(`✗ Failed to add shipment \${i}: \${err.message}`, 'error')
            }
        }
        addLog('🏁 Stress test completed', 'info')
        setLoading(false)
    }

    const runFullServiceTest = async () => {
        setLoading(true)
        clearLogs()
        addLog('🏁 Starting Comprehensive Service Test...')

        try {
            // Health Checks
            addLog('Section: Health Checks')
            await runHealthChecks()

            // Shipment Service
            addLog('Section: Shipment Service')
            const shipmentResp = await shipments.create({
                sender_name: "John Doe",
                sender_address: "123 Main St, New York, NY 10001",
                sender_phone: "+1-555-0100",
                sender_email: "john@example.com",
                receiver_name: "Jane Smith",
                receiver_address: "456 Oak Ave, Los Angeles, CA 90001",
                receiver_phone: "+1-555-0200",
                receiver_email: "jane@example.com",
                weight: 2.5,
                dimensions: "10x10x10",
                description: "Electronics package",
                carrier: "dhl",
                service_type: "express"
            })
            addLog(`✓ Create Shipment: \${shipmentResp.id}`, 'success')
            
            await shipments.get(shipmentResp.id)
            addLog('✓ Get Shipment: success', 'success')
            
            await shipments.list()
            addLog('✓ List Shipments: success', 'success')

            // Carrier Service
            addLog('Section: Carrier Service')
            await carriers.register({
                name: "DHL Express",
                code: "dhl",
                api_key: "test-api-key-123",
                api_secret: "test-api-secret-456",
                base_url: "https://api.dhl.com/v1"
            })
            addLog('✓ Register Carrier: success', 'success')
            
            await carriers.getRates({ from: "New York", to: "Los Angeles", weight: 2.5 })
            addLog('✓ Get Carrier Rates: success', 'success')

            // Rate Service
            addLog('Section: Rate Service')
            await rates.compare({
                from_address: "New York, NY",
                to_address: "Los Angeles, CA",
                weight: 2.5,
                filter_by: "cost"
            })
            addLog('✓ Compare Rates: success', 'success')

            // Label Service
            addLog('Section: Label Service')
            const labelResp = await labels.generate({
                shipment_id: shipmentResp.id,
                tracking_number: "TRACK123456789",
                carrier: "dhl"
            })
            addLog(`✓ Generate Label: \${labelResp.id}`, 'success')

            // Tracking Service
            addLog('Section: Tracking Service')
            await tracking.getInfo(shipmentResp.id)
            addLog('✓ Get Tracking: success', 'success')

            // Address Service
            addLog('Section: Address Service')
            await addresses.validate({
                street: "123 Main St",
                city: "New York",
                state: "NY",
                postal_code: "10001",
                country: "USA"
            })
            addLog('✓ Validate Address: success', 'success')

            // Billing Service
            addLog('Section: Billing Service')
            const invResp = await billing.createInvoice({
                shipment_id: shipmentResp.id,
                user_id: "test-user-001",
                amount: 45.99,
                carrier: "dhl"
            })
            addLog(`✓ Create Invoice: \${invResp.id || invResp.invoice_id}`, 'success')

            // Return Service
            addLog('Section: Return Service')
            await returns.create({
                shipment_id: shipmentResp.id,
                user_id: "test-user-001",
                reason: "product_defective",
                description: "Product arrived damaged",
                return_method: "mail"
            })
            addLog('✓ Create Return Request: success', 'success')

            addLog('🏆 ALL TESTS PASSED SUCCESSFULLY!', 'success')
        } catch (err) {
            addLog(`❌ Test Failed: \${err.message}`, 'error')
        }
        setLoading(false)
    }

    return (
        <div className="test-runner">
            <h1>System Testing & Development Tools</h1>
            <p className="subtitle">Execute system-wide test scripts and manage development data</p>

            <div className="test-actions">
                <div className="card">
                    <h3>Health Monitoring</h3>
                    <p>Check the status of all microservices</p>
                    <button className="btn btn-primary" onClick={runHealthChecks} disabled={loading}>
                        {loading ? 'Running...' : 'Run Health Checks'}
                    </button>
                    
                    <div className="health-grid">
                        {SERVICES.map(s => (
                            <div key={s.name} className={`health-item ${healthStatus[s.name]?.status || 'pending'}`}>
                                <span className="dot"></span>
                                <span className="name">{s.name}</span>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="card">
                    <h3>Demo Workflow</h3>
                    <p>Execute the full end-to-end shipping process</p>
                    <button className="btn btn-secondary" onClick={runDemoWorkflow} disabled={loading}>
                        {loading ? 'Executing...' : 'Run Demo Workflow'}
                    </button>
                </div>

                <div className="card">
                    <h3>Comprehensive Test</h3>
                    <p>Run full suite of service tests (mirrors test-all-services.sh)</p>
                    <button className="btn btn-primary" onClick={runFullServiceTest} disabled={loading} style={{ backgroundColor: '#4f46e5' }}>
                        {loading ? 'Testing...' : 'Run Full Service Test'}
                    </button>
                </div>

                <div className="card">
                    <h3>Data Generation</h3>
                    <p>Stress test the system with random shipments</p>
                    <div className="input-group" style={{ marginBottom: '10px' }}>
                        <label>Count:</label>
                        <input 
                            type="number" 
                            value={stressCount} 
                            onChange={(e) => setStressCount(parseInt(e.target.value))}
                            min="1"
                            max="100"
                        />
                    </div>
                    <button className="btn btn-outline" onClick={runStressTest} disabled={loading}>
                        {loading ? 'Generating...' : 'Start Stress Test'}
                    </button>
                </div>
            </div>

            <div className="log-console">
                <div className="log-header">
                    <h3>Execution Logs</h3>
                    <button className="btn-text" onClick={clearLogs}>Clear</button>
                </div>
                <div className="log-body">
                    {logs.length === 0 ? (
                        <div className="empty-logs">No logs to display. Run a test to see output.</div>
                    ) : (
                        logs.map(log => (
                            <div key={log.id} className={`log-entry ${log.type}`}>
                                <span className="log-time">[{log.time}]</span>
                                <span className="log-message">{log.message}</span>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </div>
    )
}
