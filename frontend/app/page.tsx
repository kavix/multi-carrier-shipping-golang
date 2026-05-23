'use client';

import { useEffect, useMemo, useState } from 'react';

const AUTH_URL = process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8083';
const SHIPMENT_URL = process.env.NEXT_PUBLIC_SHIPMENT_URL || 'http://localhost:8081';
const LABEL_URL = process.env.NEXT_PUBLIC_LABEL_URL || 'http://localhost:8082';
const NOTIFICATION_URL = process.env.NEXT_PUBLIC_NOTIFICATION_URL || 'http://localhost:8084';
const CARRIER_STATS_URL = process.env.NEXT_PUBLIC_CARRIER_STATS_URL || 'http://localhost:8085';

type Shipment = {
    id: string;
    carrier: string;
    tracking_number: string;
    weight: number;
    origin: string;
    destination: string;
    status: string;
    username: string;
    email: string;
    created_at: string;
    updated_at: string;
};

type AuditLog = {
    id: number;
    action: string;
    created_at: string;
};

type NotificationLog = {
    id: number;
    recipient: string;
    method: string;
    subject: string;
    body: string;
    status: string;
    created_at: string;
};

type CarrierStatsLog = {
    id: string;
    endpoint: string;
    status_code: number;
    duration_ms: number;
    response_size: number;
    success: boolean;
    error?: string;
    response_preview?: string;
    created_at: string;
};

export default function Home() {
    const [isRegistering, setIsRegistering] = useState(false);
    const [token, setToken] = useState('');
    const [username, setUsername] = useState('');
    const [authMessage, setAuthMessage] = useState('');
    const [authStatus, setAuthStatus] = useState<'success' | 'error' | ''>('');
    const [shipments, setShipments] = useState<Shipment[]>([]);
    const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
    const [notificationLogs, setNotificationLogs] = useState<NotificationLog[]>([]);
    const [carrierStats, setCarrierStats] = useState<Record<string, any>>({});
    const [carrierLogs, setCarrierLogs] = useState<CarrierStatsLog[]>([]);
    const [loading, setLoading] = useState(false);
    const [shipmentForm, setShipmentForm] = useState({ carrier: 'FedEx', weight: 1.5, origin: '', destination: '', email: '' });

    useEffect(() => {
        const storedToken = window.localStorage.getItem('session_token');
        const storedUser = window.localStorage.getItem('username');
        if (storedToken && storedUser) {
            setToken(storedToken);
            setUsername(storedUser);
            loadDashboard(storedToken);
        }
    }, []);

    const authenticated = Boolean(token);
    const greeting = username ? `Logged in as ${username}` : 'Sign in to view the carrier control panel';

    const buttonLabel = useMemo(() => (isRegistering ? 'Create Account' : 'Sign In'), [isRegistering]);

    async function fetchJson(url: string, options?: RequestInit) {
        const response = await fetch(url, options);
        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'Unexpected error');
        }
        return data;
    }

    async function handleAuth(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setAuthMessage('');
        setAuthStatus('');
        const form = event.currentTarget;
        const formData = new FormData(form);
        const user = String(formData.get('username') || '').trim();
        const pass = String(formData.get('password') || '').trim();

        if (!user || !pass) {
            setAuthMessage('Username and password are required.');
            setAuthStatus('error');
            return;
        }

        try {
            const endpoint = isRegistering ? `${AUTH_URL}/api/v1/auth/register` : `${AUTH_URL}/api/v1/auth/login`;
            const body = JSON.stringify({ username: user, password: pass });
            const result = await fetchJson(endpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body,
            });

            if (!isRegistering) {
                setToken(result.token);
                setUsername(result.username || user);
                window.localStorage.setItem('session_token', result.token);
                window.localStorage.setItem('username', result.username || user);
                loadDashboard(result.token);
            }

            setAuthMessage(isRegistering ? 'Account created successfully. Please sign in.' : 'Welcome back!');
            setAuthStatus('success');
            if (isRegistering) {
                setIsRegistering(false);
            }
            form.reset();
        } catch (error) {
            setAuthMessage(String(error));
            setAuthStatus('error');
        }
    }

    function handleLogout() {
        setToken('');
        setUsername('');
        setShipments([]);
        setAuditLogs([]);
        setNotificationLogs([]);
        setCarrierStats({});
        setCarrierLogs([]);
        window.localStorage.removeItem('session_token');
        window.localStorage.removeItem('username');
    }

    async function loadDashboard(sessionToken: string) {
        setLoading(true);
        try {
            await Promise.all([
                loadShipments(sessionToken),
                loadAuditLogs(sessionToken),
                loadNotificationLogs(),
                loadCarrierStats(),
            ]);
        } finally {
            setLoading(false);
        }
    }

    async function loadShipments(sessionToken: string) {
        try {
            const list = await fetchJson(`${SHIPMENT_URL}/api/v1/shipments`, {
                headers: { Authorization: `Bearer ${sessionToken}` },
            });
            setShipments(Array.isArray(list) ? list : []);
        } catch {
            setShipments([]);
        }
    }

    async function loadAuditLogs(sessionToken: string) {
        try {
            const logs = await fetchJson(`${AUTH_URL}/api/v1/auth/logs?token=${sessionToken}`);
            setAuditLogs(Array.isArray(logs) ? logs : []);
        } catch {
            setAuditLogs([]);
        }
    }

    async function loadNotificationLogs() {
        try {
            const logs = await fetchJson(`${NOTIFICATION_URL}/api/v1/notifications`);
            setNotificationLogs(Array.isArray(logs) ? logs : []);
        } catch {
            setNotificationLogs([]);
        }
    }

    async function loadCarrierStats() {
        const endpoints = {
            portCongestion: `${CARRIER_STATS_URL}/api/v1/carrier-stats/port-congestion`,
            freightRates: `${CARRIER_STATS_URL}/api/v1/carrier-stats/freight-rates`,
            fuelPrices: `${CARRIER_STATS_URL}/api/v1/carrier-stats/fuel-prices`,
            disruptions: `${CARRIER_STATS_URL}/api/v1/carrier-stats/disruptions`,
            carriers: `${CARRIER_STATS_URL}/api/v1/carrier-stats/carriers`,
            logs: `${CARRIER_STATS_URL}/api/v1/carrier-stats/logs?limit=10`,
        };

        const results = await Promise.all(
            Object.entries(endpoints).map(async ([name, url]) => {
                try {
                    const response = await fetch(url);
                    return { name, result: await response.json() };
                } catch {
                    return { name, result: null };
                }
            })
        );

        const data = results.reduce<Record<string, any>>((acc, item) => {
            acc[item.name] = item.result;
            return acc;
        }, {});

        setCarrierStats({
            portCongestion: data.portCongestion,
            freightRates: data.freightRates,
            fuelPrices: data.fuelPrices,
            disruptions: data.disruptions,
            carriers: data.carriers,
        });
        setCarrierLogs(Array.isArray(data.logs) ? data.logs : []);
    }

    async function handleShipmentSubmit(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        if (!authenticated) return;

        try {
            await fetchJson(`${SHIPMENT_URL}/api/v1/shipments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify(shipmentForm),
            });
            setShipmentForm({ carrier: 'FedEx', weight: 1.5, origin: '', destination: '', email: '' });
            loadDashboard(token);
        } catch (error) {
            console.error(error);
        }
    }

    return (
        <main>
            <section className="section header-row">
                <div>
                    <h1>Antigravity Carrier Dashboard</h1>
                    <p className="small-text">Unified shipment control panel with FreightPulse port, rate, fuel, disruption, and carrier intelligence.</p>
                </div>
                <div>
                    <span className="badge badge-info">{greeting}</span>
                    {authenticated && (
                        <button className="button" type="button" onClick={handleLogout} style={{ marginLeft: '1rem' }}>
                            Logout
                        </button>
                    )}
                </div>
            </section>

            {!authenticated ? (
                <section className="card section">
                    <div className="header-row">
                        <div>
                            <h2>{isRegistering ? 'Register Your Account' : 'Welcome Back'}</h2>
                            <p className="small-text">Access the multi-service dashboard for shipments, notifications, and carrier analytics.</p>
                        </div>
                        <button className="button" type="button" onClick={() => setIsRegistering((current) => !current)}>
                            Switch to {isRegistering ? 'Login' : 'Register'}
                        </button>
                    </div>

                    <form className="input-group" onSubmit={handleAuth}>
                        <input className="input" name="username" placeholder="Username" autoComplete="username" required />
                        <input className="input" type="password" name="password" placeholder="Password" autoComplete="current-password" required />
                        <button className="button" type="submit">{buttonLabel}</button>
                        {authMessage && <p className={authStatus === 'error' ? 'error-state' : 'empty-state'}>{authMessage}</p>}
                    </form>
                </section>
            ) : (
                <>
                    <section className="section grid-2">
                        <div className="card">
                            <div className="header-row">
                                <div>
                                    <h2>New Shipment Request</h2>
                                    <p className="small-text">Create a shipment and generate a carrier label in one flow.</p>
                                </div>
                                <span className="badge badge-success">Shipment Service</span>
                            </div>

                            <form className="input-group" onSubmit={handleShipmentSubmit}>
                                <select className="input" value={shipmentForm.carrier} onChange={(event) => setShipmentForm({ ...shipmentForm, carrier: event.target.value })}>
                                    <option>FedEx</option>
                                    <option>DHL</option>
                                    <option>UPS</option>
                                    <option>USPS</option>
                                </select>

                                <input className="input" type="number" value={shipmentForm.weight} min="0.1" step="0.1" onChange={(event) => setShipmentForm({ ...shipmentForm, weight: Number(event.target.value) })} placeholder="Weight (lbs)" required />
                                <input className="input" value={shipmentForm.origin} onChange={(event) => setShipmentForm({ ...shipmentForm, origin: event.target.value })} placeholder="Origin" required />
                                <input className="input" value={shipmentForm.destination} onChange={(event) => setShipmentForm({ ...shipmentForm, destination: event.target.value })} placeholder="Destination" required />
                                <input className="input" type="email" value={shipmentForm.email} onChange={(event) => setShipmentForm({ ...shipmentForm, email: event.target.value })} placeholder="Recipient Email" required />
                                <button className="button" type="submit">Create Shipment</button>
                            </form>
                        </div>

                        <div className="card">
                            <div className="header-row">
                                <div>
                                    <h2>Live Signals</h2>
                                    <p className="small-text">Refresh any time to sync the carrier analytics service.</p>
                                </div>
                                <button className="button" type="button" onClick={() => loadCarrierStats()}>Refresh Intelligence</button>
                            </div>

                            <div className="card-grid">
                                <div className="card-small">
                                    <h4>Port Congestion</h4>
                                    <p className="small-text">Live global congestion from FreightPulse.</p>
                                </div>
                                <div className="card-small">
                                    <h4>Fuel Prices</h4>
                                    <p className="small-text">Fuel and bunker indices.</p>
                                </div>
                                <div className="card-small">
                                    <h4>Carrier Ratings</h4>
                                    <p className="small-text">Ocean, trucking and air reliability.</p>
                                </div>
                            </div>
                        </div>
                    </section>

                    <section className="section card">
                        <div className="header-row">
                            <div>
                                <h2>Shipment Database Explorer</h2>
                                <p className="small-text">Your latest created shipments and their current state.</p>
                            </div>
                            <span className="badge badge-info">{shipments.length} shipments</span>
                        </div>

                        {shipments.length === 0 ? (
                            <p className="empty-state">No shipments found yet.</p>
                        ) : (
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>ID</th>
                                        <th>Carrier</th>
                                        <th>Recipient</th>
                                        <th>Tracking</th>
                                        <th>Status</th>
                                        <th>Created</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {shipments.map((shipment) => (
                                        <tr key={shipment.id}>
                                            <td>{shipment.id.slice(0, 8)}...</td>
                                            <td>{shipment.carrier}</td>
                                            <td>{shipment.email}</td>
                                            <td>{shipment.tracking_number}</td>
                                            <td><span className={`status-chip status-${shipment.status.toLowerCase().replace(/ /g, '-')}`}>{shipment.status}</span></td>
                                            <td>{new Date(shipment.created_at).toLocaleString()}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
                    </section>

                    <section className="section grid-2">
                        <div className="card">
                            <div className="header-row">
                                <div>
                                    <h2>Audit Timeline</h2>
                                    <p className="small-text">Recent authentication actions from the Auth Service.</p>
                                </div>
                                <span className="badge badge-success">AUTH SERVICE</span>
                            </div>
                            {auditLogs.length === 0 ? (
                                <p className="empty-state">No audit entries yet.</p>
                            ) : (
                                <div className="list-card">
                                    {auditLogs.map((log) => (
                                        <div key={log.id} className="list-item">
                                            <div>{log.action}</div>
                                            <div className="small-text">{new Date(log.created_at).toLocaleString()}</div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        <div className="card">
                            <div className="header-row">
                                <div>
                                    <h2>Notification Hub</h2>
                                    <p className="small-text">Latest logs from the decoupled notification service.</p>
                                </div>
                                <span className="badge badge-warning">NOTIFICATION SERVICE</span>
                            </div>
                            {notificationLogs.length === 0 ? (
                                <p className="empty-state">No notification activity yet.</p>
                            ) : (
                                <div className="list-card">
                                    {notificationLogs.slice(0, 8).map((note) => (
                                        <div key={note.id} className="list-item">
                                            <div>
                                                <strong>{note.method}</strong> to {note.recipient}
                                                <div className="small-text">{note.subject || 'Message logged'}</div>
                                            </div>
                                            <div className="small-text">{new Date(note.created_at).toLocaleString()}</div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </section>

                    <section className="section card">
                        <div className="header-row">
                            <div>
                                <h2>Global Carrier Intelligence</h2>
                                <p className="small-text">Aggregated FreightPulse data from your new carrier-stats service.</p>
                            </div>
                            <span className="badge badge-info">MONGODB LOGS</span>
                        </div>

                        <div className="stats-grid">
                            <div className="card-small">
                                <h4>Port Congestion</h4>
                                <p className="small-text">{carrierStats.portCongestion?.data?.data?.global_summary?.trend || 'Loading...'}</p>
                            </div>
                            <div className="card-small">
                                <h4>Freight Rates</h4>
                                <p className="small-text">{carrierStats.freightRates?.data?.data?.market_summary?.ocean_outlook || 'Loading...'}</p>
                            </div>
                            <div className="card-small">
                                <h4>Fuel Prices</h4>
                                <p className="small-text">Diesel: ${carrierStats.fuelPrices?.data?.data?.diesel?.national_average || '...'}</p>
                            </div>
                            <div className="card-small">
                                <h4>Disruptions</h4>
                                <p className="small-text">{carrierStats.disruptions?.data?.data?.active_alerts ?? '...'} active alerts</p>
                            </div>
                            <div className="card-small">
                                <h4>Carrier Ratings</h4>
                                <p className="small-text">{carrierStats.carriers?.data?.data?.ocean?.length ?? '...'} ocean carriers available</p>
                            </div>
                        </div>

                        <div className="list-card" style={{ marginTop: '1rem' }}>
                            {carrierLogs.length === 0 ? (
                                <p className="empty-state">No carrier stats logs captured yet.</p>
                            ) : (
                                carrierLogs.map((log) => (
                                    <div key={log.id} className="list-item">
                                        <div>
                                            <strong>{log.endpoint}</strong>
                                            <div className="small-text">{log.success ? 'Success' : 'Failed'} · {log.status_code}</div>
                                        </div>
                                        <div className="small-text">{new Date(log.created_at).toLocaleString()}</div>
                                    </div>
                                ))
                            )}
                        </div>
                    </section>
                </>
            )}
        </main>
    );
}
