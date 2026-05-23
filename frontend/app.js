// Configuration endpoints
const AUTH_SVC_URL = 'http://localhost:8083';
const SHIPMENT_SVC_URL = 'http://localhost:8081';
const LABEL_SVC_URL = 'http://localhost:8082';
const NOTIFICATION_SVC_URL = 'http://localhost:8084';
const CARRIER_STATS_SVC_URL = 'http://localhost:8085';

// State Management
let sessionToken = localStorage.getItem('session_token') || '';
let currentUsername = localStorage.getItem('username') || '';

// DOM Elements
const authSection = document.getElementById('authSection');
const dashboardContainer = document.getElementById('dashboardContainer');
const userProfile = document.getElementById('userProfile');
const usernameDisplay = document.getElementById('usernameDisplay');
const logoutBtn = document.getElementById('logoutBtn');

const loginForm = document.getElementById('loginForm');
const registerForm = document.getElementById('registerForm');
const toggleLogin = document.getElementById('toggleLogin');
const toggleRegister = document.getElementById('toggleRegister');
const authMessage = document.getElementById('authMessage');

const createShipmentForm = document.getElementById('createShipmentForm');
const createMessage = document.getElementById('createMessage');
const shipmentsTableBody = document.getElementById('shipmentsTableBody');
const refreshShipmentsBtn = document.getElementById('refreshShipmentsBtn');
const auditTimeline = document.getElementById('auditTimeline');

const refreshCarrierStatsBtn = document.getElementById('refreshCarrierStatsBtn');
const portCongestionCard = document.getElementById('portCongestionCard');
const freightRatesCard = document.getElementById('freightRatesCard');
const fuelPricesCard = document.getElementById('fuelPricesCard');
const disruptionsCard = document.getElementById('disruptionsCard');
const carriersCard = document.getElementById('carriersCard');
const carrierStatsLogTimeline = document.getElementById('carrierStatsLogTimeline');

// Modals
const editModal = document.getElementById('editModal');
const editShipmentForm = document.getElementById('editShipmentForm');
const closeEditModal = document.getElementById('closeEditModal');
const editMessage = document.getElementById('editMessage');

const labelModal = document.getElementById('labelModal');
const closeLabelModal = document.getElementById('closeLabelModal');
const labelDetailsContainer = document.getElementById('labelDetailsContainer');
const cancelLabelBtn = document.getElementById('cancelLabelBtn');
const labelMessage = document.getElementById('labelMessage');

let selectedTrackingNumber = ''; // For canceling label

// App Init
document.addEventListener('DOMContentLoaded', () => {
  setupEventListeners();
  checkAuthentication();
});

// Setup Event Listeners
function setupEventListeners() {
  // Tabs toggle
  toggleLogin.addEventListener('click', () => {
    toggleLogin.classList.add('active');
    toggleRegister.classList.remove('active');
    loginForm.classList.add('active');
    registerForm.classList.remove('active');
    hideAlert(authMessage);
  });

  toggleRegister.addEventListener('click', () => {
    toggleRegister.classList.add('active');
    toggleLogin.classList.remove('active');
    registerForm.classList.add('active');
    loginForm.classList.remove('active');
    hideAlert(authMessage);
  });

  // Auth Forms Submission
  loginForm.addEventListener('submit', handleLogin);
  registerForm.addEventListener('submit', handleRegister);
  logoutBtn.addEventListener('click', handleLogout);

  // Shipment Operations
  createShipmentForm.addEventListener('submit', handleCreateShipment);
  refreshShipmentsBtn.addEventListener('click', loadShipmentsAndLogs);
  if (refreshCarrierStatsBtn) {
    refreshCarrierStatsBtn.addEventListener('click', loadCarrierStats);
  }

  // Modals Actions
  closeEditModal.addEventListener('click', () => editModal.style.display = 'none');
  editShipmentForm.addEventListener('submit', handleUpdateShipment);

  closeLabelModal.addEventListener('click', () => labelModal.style.display = 'none');
  cancelLabelBtn.addEventListener('click', handleCancelLabel);

  // Close modals when clicking outside
  window.addEventListener('click', (e) => {
    if (e.target === editModal) editModal.style.display = 'none';
    if (e.target === labelModal) labelModal.style.display = 'none';
  });
}

// Check Authentication Session
async function checkAuthentication() {
  if (!sessionToken) {
    showAuthPanel();
    return;
  }

  try {
    const response = await fetch(`${AUTH_SVC_URL}/api/v1/auth/verify?token=${sessionToken}`);
    if (response.ok) {
      const data = await response.json();
      currentUsername = data.username;
      localStorage.setItem('username', currentUsername);
      showDashboard();
    } else {
      // Token expired or invalid
      handleLogout();
    }
  } catch (err) {
    console.error('Auth check offline, falling back to local storage session', err);
    if (currentUsername) {
      showDashboard();
    } else {
      showAuthPanel();
    }
  }
}

// View Switches
function showAuthPanel() {
  authSection.style.display = 'block';
  dashboardContainer.style.display = 'none';
  userProfile.style.display = 'none';
}

function showDashboard() {
  authSection.style.display = 'none';
  dashboardContainer.style.display = 'block';
  userProfile.style.display = 'flex';
  usernameDisplay.textContent = `Logged in as: ${currentUsername}`;
  loadShipmentsAndLogs();
}

// Authentication Logic
async function handleRegister(e) {
  e.preventDefault();
  const username = document.getElementById('registerUsername').value;
  const password = document.getElementById('registerPassword').value;

  try {
    const response = await fetch(`${AUTH_SVC_URL}/api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const data = await response.json();

    if (response.ok) {
      showAlert(authMessage, 'Account successfully created! Please login.', 'success');
      registerForm.reset();
      // Auto toggle to login tab
      toggleLogin.click();
    } else {
      showAlert(authMessage, data.error || 'Registration failed', 'error');
    }
  } catch (err) {
    showAlert(authMessage, 'Could not reach Authentication Service', 'error');
  }
}

async function handleLogin(e) {
  e.preventDefault();
  const username = document.getElementById('loginUsername').value;
  const password = document.getElementById('loginPassword').value;

  try {
    const response = await fetch(`${AUTH_SVC_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const data = await response.json();

    if (response.ok) {
      sessionToken = data.token;
      currentUsername = data.username;
      localStorage.setItem('session_token', sessionToken);
      localStorage.setItem('username', currentUsername);
      loginForm.reset();
      showDashboard();
    } else {
      showAlert(authMessage, data.error || 'Invalid credentials', 'error');
    }
  } catch (err) {
    showAlert(authMessage, 'Could not reach Authentication Service', 'error');
  }
}

function handleLogout() {
  sessionToken = '';
  currentUsername = '';
  localStorage.removeItem('session_token');
  localStorage.removeItem('username');
  showAuthPanel();
}

// load data from Shipment, Auth, and Notification databases
async function loadShipmentsAndLogs() {
  loadShipments();
  loadAuditLogs();
  loadNotificationLogs();
  loadCarrierStats();
}

// CRUD Shipment - Create
async function handleCreateShipment(e) {
  e.preventDefault();
  const carrier = document.getElementById('carrier').value;
  const weight = parseFloat(document.getElementById('weight').value);
  const origin = document.getElementById('origin').value;
  const destination = document.getElementById('destination').value;
  const email = document.getElementById('email').value;

  try {
    const response = await fetch(`${SHIPMENT_SVC_URL}/api/v1/shipments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${sessionToken}`
      },
      body: JSON.stringify({ carrier, weight, origin, destination, email })
    });

    const data = await response.json();

    if (response.ok) {
      showAlert(createMessage, `Shipment successfully dispatched! Label generated: ${data.label.tracking_number}`, 'success');
      createShipmentForm.reset();
      loadShipmentsAndLogs();
    } else {
      showAlert(createMessage, data.error || 'Failed to dispatch shipment', 'error');
    }
  } catch (err) {
    showAlert(createMessage, 'Failed to connect to Shipment Service', 'error');
  }
}

// CRUD Shipment - Read (List)
async function loadShipments() {
  try {
    const response = await fetch(`${SHIPMENT_SVC_URL}/api/v1/shipments`, {
      headers: { 'Authorization': `Bearer ${sessionToken}` }
    });
    if (!response.ok) throw new Error('Failed to retrieve shipments');
    const shipments = await response.json();

    if (!shipments || shipments.length === 0) {
      shipmentsTableBody.innerHTML = `<tr><td colspan="9" class="placeholder-text">No active shipments in the database. Add one above.</td></tr>`;
      return;
    }

    shipmentsTableBody.innerHTML = shipments.map(s => {
      const createdDate = new Date(s.created_at).toLocaleString();
      const updatedDate = new Date(s.updated_at).toLocaleString();

      let statusClass = 'status-pending';
      if (s.status === 'CREATED') statusClass = 'status-created';
      if (s.status === 'CANCELLED') statusClass = 'status-cancelled';
      if (s.status === 'IN_TRANSIT') statusClass = 'status-transit';
      if (s.status === 'OUT_FOR_DELIVERY') statusClass = 'status-delivery';
      if (s.status === 'DELIVERED') statusClass = 'status-delivered';
      if (s.status === 'RETURNED') statusClass = 'status-returned';

      return `
        <tr>
          <td><code style="font-size: 0.75rem;">${s.id.substring(0, 8)}...</code></td>
          <td><span style="font-size: 0.75rem; font-weight: 700; padding: 2px 8px; border-radius: 9999px; background: rgba(99, 102, 241, 0.15); color: #818cf8; border: 1px solid rgba(99, 102, 241, 0.3);">${s.username}</span></td>
          <td><strong>${s.carrier}</strong></td>
          <td><span style="font-size: 0.85rem; color: var(--text-secondary);">${s.email || 'N/A'}</span></td>
          <td><code style="font-size: 0.85rem;">${s.tracking_number}</code></td>
          <td>${s.weight} lbs</td>
          <td><span class="status-badge ${statusClass}">${s.status}</span></td>
          <td>
            <div style="font-size: 0.75rem;">Created: ${createdDate}</div>
            <div style="font-size: 0.7rem; color: var(--text-secondary);">Updated: ${updatedDate}</div>
          </td>
          <td>
            <div class="action-links">
              <button class="link-btn link-sky" onclick="viewLabelDetails('${s.tracking_number}')">View Label</button>
              <button class="link-btn link-indigo" onclick="openEditShipment('${s.id}', '${s.carrier}', ${s.weight}, '${s.origin}', '${s.destination}', '${s.status}')">Edit</button>
              <button class="link-btn link-danger" onclick="deleteShipment('${s.id}')">Delete</button>
            </div>
          </td>
        </tr>
      `;
    }).join('');
  } catch (err) {
    shipmentsTableBody.innerHTML = `<tr><td colspan="9" class="placeholder-text alert-error" style="background: none;">Failed to sync with Shipment database.</td></tr>`;
  }
}

// CRUD Shipment - Update
function openEditShipment(id, carrier, weight, origin, destination, status) {
  document.getElementById('editShipmentId').value = id;
  document.getElementById('editCarrier').value = carrier;
  document.getElementById('editWeight').value = weight;
  document.getElementById('editOrigin').value = origin;
  document.getElementById('editDestination').value = destination;

  // Show transit status dropdown only for admin
  const statusGroup = document.getElementById('editStatusGroup');
  if (currentUsername === 'admin') {
    statusGroup.style.display = 'block';
    document.getElementById('editStatus').value = status || 'CREATED';
  } else {
    statusGroup.style.display = 'none';
  }

  hideAlert(editMessage);
  editModal.style.display = 'flex';
}

async function handleUpdateShipment(e) {
  e.preventDefault();
  const id = document.getElementById('editShipmentId').value;
  const carrier = document.getElementById('editCarrier').value;
  const weight = parseFloat(document.getElementById('editWeight').value);
  const origin = document.getElementById('editOrigin').value;
  const destination = document.getElementById('editDestination').value;

  // Capture status change if admin
  const status = currentUsername === 'admin' ? document.getElementById('editStatus').value : '';

  try {
    const response = await fetch(`${SHIPMENT_SVC_URL}/api/v1/shipments/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${sessionToken}`
      },
      body: JSON.stringify({ carrier, weight, origin, destination, status })
    });

    const data = await response.json();

    if (response.ok) {
      showAlert(editMessage, 'Shipment updated successfully!', 'success');
      setTimeout(() => {
        editModal.style.display = 'none';
        loadShipmentsAndLogs();
      }, 1000);
    } else {
      showAlert(editMessage, data.error || 'Failed to update shipment', 'error');
    }
  } catch (err) {
    showAlert(editMessage, 'Connection error updating shipment', 'error');
  }
}

// CRUD Shipment - Delete
async function deleteShipment(id) {
  if (!confirm('Are you absolutely sure you want to delete this shipment? This cannot be undone.')) return;

  try {
    const response = await fetch(`${SHIPMENT_SVC_URL}/api/v1/shipments/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${sessionToken}` }
    });

    const data = await response.json();

    if (response.ok) {
      loadShipmentsAndLogs();
    } else {
      alert(data.error || 'Failed to delete shipment');
    }
  } catch (err) {
    alert('Connection error deleting shipment');
  }
}

// Label Service - View Label & Drop-off locations
async function viewLabelDetails(trackingNumber) {
  if (trackingNumber === 'PENDING') {
    alert('Label is currently pending generation.');
    return;
  }

  hideAlert(labelMessage);
  selectedTrackingNumber = trackingNumber;

  try {
    // 1. Fetch label details from Label Service
    const labelResp = await fetch(`${LABEL_SVC_URL}/api/v1/labels/${trackingNumber}`);
    if (!labelResp.ok) throw new Error('Label not found');
    const label = await labelResp.json();

    // 2. Fetch shipment details from Shipment Service
    const shipmentResp = await fetch(`${SHIPMENT_SVC_URL}/api/v1/shipments/${label.shipment_id}`);
    const shipment = shipmentResp.ok ? await shipmentResp.json() : null;

    // Show location details under simulated Paper Label
    const originStr = shipment ? shipment.origin : 'N/A';
    const destStr = shipment ? shipment.destination : 'N/A';
    const carrier = shipment ? shipment.carrier : 'Carrier';
    const weight = shipment ? `${shipment.weight} LBS` : 'N/A';

    // Show cancel button only if active
    if (label.status === 'CANCELLED') {
      cancelLabelBtn.style.display = 'none';
    } else {
      cancelLabelBtn.style.display = 'block';
    }

    labelDetailsContainer.innerHTML = `
      <div class="label-mock-paper">
        <div class="label-mock-header">
          <span>${carrier.toUpperCase()} PRIORITY</span>
          <span>SANDBOX</span>
        </div>
        <div class="label-mock-addresses">
          <div class="label-mock-address-box">
            <strong>FROM (SHIPPER):</strong>
            ${originStr}
          </div>
          <div class="label-mock-address-box">
            <strong>TO (RECIPIENT):</strong>
            ${destStr}
          </div>
        </div>
        <div style="border-top: 1px solid #0f172a; padding-top: 0.5rem; font-size: 0.75rem; font-weight: 700;">
          SHIPMENT ID: <span style="font-family: monospace;">${label.shipment_id.substring(0, 14)}...</span><br>
          WEIGHT: ${weight}
        </div>
        <div class="barcode-simulated"></div>
        <div class="tracking-text-simulated">TRK# ${label.tracking_number}</div>
      </div>
      
      <div class="loc-row">
        <strong>Label Status:</strong>
        <div>
          <span class="status-badge ${label.status === 'ACTIVE' ? 'status-created' : 'status-cancelled'}">
            ${label.status}
          </span>
        </div>
      </div>

      <div class="loc-grid">
        <div class="loc-box">
          <h4>Origin FedEx Dropoffs</h4>
          <div style="font-size: 0.75rem; line-height: 1.3;">
            <strong>FedEx Office Center</strong><br>
            439 N Beverly Dr, Beverly Hills<br>
            <span style="color: var(--accent-emerald);">Distance: 1.20 MI (Nearest)</span>
          </div>
        </div>
        <div class="loc-box">
          <h4>Destination FedEx Dropoffs</h4>
          <div style="font-size: 0.75rem; line-height: 1.3;">
            <strong>FedEx Drop Box</strong><br>
            11 San Luis Ct, Walnut Creek<br>
            <span style="color: var(--accent-emerald);">Distance: 2.40 MI (Nearest)</span>
          </div>
        </div>
      </div>
    `;

    labelModal.style.display = 'flex';
  } catch (err) {
    alert('Failed to retrieve shipping label details.');
  }
}

// Label Service - Cancel Label
async function handleCancelLabel() {
  if (!confirm('Are you sure you want to cancel this shipping label? This will void the shipment.')) return;

  try {
    const response = await fetch(`${LABEL_SVC_URL}/api/v1/labels/${selectedTrackingNumber}/cancel`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${sessionToken}` }
    });

    const data = await response.json();

    if (response.ok) {
      showAlert(labelMessage, 'Shipping Label successfully voided and cancelled!', 'success');
      setTimeout(() => {
        labelModal.style.display = 'none';
        loadShipmentsAndLogs();
      }, 1000);
    } else {
      showAlert(labelMessage, data.error || 'Failed to cancel label', 'error');
    }
  } catch (err) {
    showAlert(labelMessage, 'Connection error voiding label', 'error');
  }
}

// Auth Service - Load live audit logs from DB
async function loadAuditLogs() {
  try {
    const response = await fetch(`${AUTH_SVC_URL}/api/v1/auth/logs?token=${sessionToken}`);
    if (!response.ok) throw new Error('Unauthenticated');
    const logs = await response.json();

    if (!logs || logs.length === 0) {
      auditTimeline.innerHTML = `<p class="placeholder-text">No logged actions recorded yet.</p>`;
      return;
    }

    auditTimeline.innerHTML = logs.map(l => {
      const logDate = new Date(l.created_at).toLocaleString();
      const isLogin = l.action.toLowerCase() === 'login';
      return `
        <div class="timeline-item ${isLogin ? 'login' : ''}">
          <div class="timeline-title">${l.action}</div>
          <div class="timeline-meta">${logDate}</div>
        </div>
      `;
    }).join('');
  } catch (err) {
    auditTimeline.innerHTML = `<p class="placeholder-text alert-error" style="background: none;">Failed to fetch live audit history.</p>`;
  }
}

// Helper Alert controls
function showAlert(element, message, type) {
  element.textContent = message;
  element.className = `alert alert-${type}`;
  element.style.display = 'block';

  // Auto fade out create alert after 6s
  if (element === createMessage) {
    setTimeout(() => hideAlert(element), 6000);
  }
}

function hideAlert(element) {
  element.style.display = 'none';
  element.textContent = '';
}

// Customer Notification Hub Timeline Loader
async function loadNotificationLogs() {
  const timelineEl = document.getElementById('notificationHubTimeline');
  if (!timelineEl) return;

  try {
    const response = await fetch(`${NOTIFICATION_SVC_URL}/api/v1/notifications`);
    if (!response.ok) throw new Error('Failed to retrieve notification logs');
    const logs = await response.json();

    if (!logs || logs.length === 0) {
      timelineEl.innerHTML = `<p class="placeholder-text">No notification logs recorded yet.</p>`;
      return;
    }

    timelineEl.innerHTML = logs.map(l => {
      const logDate = new Date(l.created_at).toLocaleString();
      const isEmail = l.method === 'EMAIL';
      const icon = isEmail ? '✉️' : '💬';
      const badgeClass = isEmail ? 'badge-email' : 'badge-telegram';

      let bodyText = l.body;
      if (isEmail) {
        try {
          const parser = new DOMParser();
          const doc = parser.parseFromString(l.body, 'text/html');
          const textContent = doc.body.textContent || "";
          bodyText = textContent.replace(/\s+/g, ' ').substring(0, 150).trim() + "...";
        } catch (e) {
          bodyText = l.body.substring(0, 150) + "...";
        }
      }

      return `
        <div class="timeline-item ${isEmail ? 'email' : 'telegram'}">
          <div class="timeline-title">
            <span style="font-size: 1.1rem; margin-right: 4px;">${icon}</span>
            <span class="badge ${badgeClass}" style="font-size: 0.7rem; padding: 2px 6px; border-radius: 4px; font-weight: 700; margin-right: 6px;">${l.method}</span>
            <strong style="font-size: 0.85rem;">${l.recipient}</strong>
          </div>
          <div style="font-size: 0.8rem; font-weight: 600; color: var(--accent-sky); margin-top: 0.25rem;">
            ${l.subject ? l.subject : 'Simulated Telegram Alert'}
          </div>
          <div class="timeline-body">${bodyText}</div>
          <div class="timeline-meta" style="font-size: 0.7rem; color: var(--text-secondary);">
            Status: <strong style="color: ${l.status === 'SENT' ? 'var(--accent-emerald)' : 'var(--accent-danger)'}">${l.status}</strong> | ${logDate}
          </div>
        </div>
      `;
    }).join('');
  } catch (err) {
    console.error('Failed to load notification logs:', err);
    timelineEl.innerHTML = `<p class="placeholder-text alert-error" style="background: none;">Failed to sync with Notification Hub.</p>`;
  }
}

async function loadCarrierStats() {
  const requests = [
    { key: 'portCongestion', url: `${CARRIER_STATS_SVC_URL}/api/v1/carrier-stats/port-congestion` },
    { key: 'freightRates', url: `${CARRIER_STATS_SVC_URL}/api/v1/carrier-stats/freight-rates` },
    { key: 'fuelPrices', url: `${CARRIER_STATS_SVC_URL}/api/v1/carrier-stats/fuel-prices` },
    { key: 'disruptions', url: `${CARRIER_STATS_SVC_URL}/api/v1/carrier-stats/disruptions` },
    { key: 'carriers', url: `${CARRIER_STATS_SVC_URL}/api/v1/carrier-stats/carriers` },
    { key: 'logs', url: `${CARRIER_STATS_SVC_URL}/api/v1/carrier-stats/logs?limit=12` },
  ];

  const results = await Promise.all(requests.map(async (request) => {
    try {
      const response = await fetch(request.url);
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }
      return { key: request.key, data: await response.json() };
    } catch (error) {
      return { key: request.key, error };
    }
  }));

  const resultMap = Object.fromEntries(results.map((result) => [result.key, result]));

  renderCarrierStatCard(portCongestionCard, renderPortCongestionCard(resultMap.portCongestion));
  renderCarrierStatCard(freightRatesCard, renderFreightRatesCard(resultMap.freightRates));
  renderCarrierStatCard(fuelPricesCard, renderFuelPricesCard(resultMap.fuelPrices));
  renderCarrierStatCard(disruptionsCard, renderDisruptionsCard(resultMap.disruptions));
  renderCarrierStatCard(carriersCard, renderCarriersCard(resultMap.carriers));
  renderCarrierStatsLogs(resultMap.logs);
}

function renderCarrierStatCard(element, content) {
  if (!element) return;
  element.innerHTML = content;
}

function renderCarrierStatError(title, message) {
  return `
    <div class="carrier-stat-card-header">
      <span class="carrier-stat-kicker">${title}</span>
    </div>
    <p class="placeholder-text alert-error" style="background: none; margin: auto 0;">${message}</p>
  `;
}

function formatNumber(value, digits = 0) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) return 'N/A';
  return Number(value).toLocaleString(undefined, {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  });
}

function formatCurrency(value, digits = 0) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) return 'N/A';
  return `$${Number(value).toLocaleString(undefined, {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  })}`;
}

function formatMaybeDate(value) {
  if (!value) return 'N/A';
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? 'N/A' : parsed.toLocaleString();
}

function renderChips(items, severityClass = '') {
  if (!Array.isArray(items) || items.length === 0) {
    return '<span class="carrier-stat-note">No active alerts</span>';
  }

  return `
    <div class="carrier-chip-row">
      ${items.slice(0, 3).map((item) => `<span class="carrier-chip ${severityClass}">${item}</span>`).join('')}
    </div>
  `;
}

function renderPortCongestionCard(result) {
  if (!result || result.error) {
    return renderCarrierStatError('PORT CONGESTION', 'Unable to load live port congestion data.');
  }

  const payload = result.data || {};
  const summary = payload?.data?.data?.global_summary || {};
  const ports = [...(payload?.data?.data?.ports || [])].sort((left, right) => (right.congestion_index || 0) - (left.congestion_index || 0)).slice(0, 4);

  return `
    <div class="carrier-stat-card-header">
      <span class="carrier-stat-kicker">PORT CONGESTION</span>
      <span class="carrier-stat-pill">${formatNumber(payload?.data?.total_ports || ports.length)} ports tracked</span>
    </div>
    <div class="carrier-stat-metrics">
      <div><strong>${formatNumber(summary.avg_congestion_index, 0)}</strong><span>Avg index</span></div>
      <div><strong>${formatNumber(summary.ports_high_congestion, 0)}</strong><span>High</span></div>
      <div><strong>${formatNumber(summary.ports_moderate, 0)}</strong><span>Moderate</span></div>
    </div>
    <div class="carrier-stat-list">
      ${ports.map((port) => `
        <div class="carrier-stat-row">
          <span>${port.port || 'Unknown Port'}</span>
          <span>${formatNumber(port.congestion_index, 0)} | ${String(port.congestion_level || 'n/a').toUpperCase()}</span>
        </div>
      `).join('')}
    </div>
    <div class="carrier-stat-footer">
      ${renderChips(summary.active_disruptions || [])}
      <div class="carrier-stat-note" style="margin-top: 0.5rem;">Updated ${formatMaybeDate(payload?.data?.timestamp)}</div>
    </div>
  `;
}

function renderFreightRatesCard(result) {
  if (!result || result.error) {
    return renderCarrierStatError('FREIGHT RATES', 'Unable to load live freight rates data.');
  }

  const payload = result.data || {};
  const ocean = payload?.data?.data?.ocean || {};
  const indices = ocean.indices || {};
  const routes = [...(ocean.container_rates || [])].sort((left, right) => (right.rate_40hc || 0) - (left.rate_40hc || 0)).slice(0, 3);
  const marketSummary = payload?.data?.data?.market_summary || {};

  return `
    <div class="carrier-stat-card-header">
      <span class="carrier-stat-kicker">FREIGHT RATES</span>
      <span class="carrier-stat-pill">${payload?.data?.currency || 'USD'}</span>
    </div>
    <div class="carrier-stat-metrics">
      <div><strong>${formatNumber(indices.fbx_global, 0)}</strong><span>FBX Global</span></div>
      <div><strong>${formatNumber(indices.scfi, 0)}</strong><span>SCFI</span></div>
      <div><strong>${formatNumber(indices.wci, 0)}</strong><span>WCI</span></div>
    </div>
    <div class="carrier-stat-list">
      ${routes.map((route) => `
        <div class="carrier-stat-row">
          <span>${route.route || 'Route'}</span>
          <span>${formatCurrency(route.rate_40hc, 0)} | ${formatNumber(route.transit_days, 0)}d</span>
        </div>
      `).join('')}
    </div>
    <div class="carrier-stat-footer carrier-stat-note">${marketSummary.ocean_outlook || 'Ocean outlook unavailable.'}</div>
  `;
}

function renderFuelPricesCard(result) {
  if (!result || result.error) {
    return renderCarrierStatError('FUEL PRICES', 'Unable to load live fuel price data.');
  }

  const payload = result.data || {};
  const data = payload?.data?.data || {};
  const diesel = data.diesel || {};
  const gasoline = data.gasoline || {};
  const bunker = data.bunker_fuel || {};
  const historical = data.historical || {};

  return `
    <div class="carrier-stat-card-header">
      <span class="carrier-stat-kicker">FUEL PRICES</span>
      <span class="carrier-stat-pill">${payload?.data?.unit || 'gallon'}</span>
    </div>
    <div class="carrier-stat-metrics">
      <div><strong>${formatCurrency(diesel.national_average, 2)}</strong><span>Diesel avg</span></div>
      <div><strong>${formatCurrency(gasoline.national_average, 2)}</strong><span>Gasoline avg</span></div>
      <div><strong>${formatCurrency(bunker.vlsfo, 0)}</strong><span>VLSFO</span></div>
    </div>
    <div class="carrier-stat-list">
      <div class="carrier-stat-row"><span>Diesel 30d avg</span><span>${formatCurrency(historical.diesel_30d_avg, 2)}</span></div>
      <div class="carrier-stat-row"><span>Diesel 90d avg</span><span>${formatCurrency(historical.diesel_90d_avg, 2)}</span></div>
      <div class="carrier-stat-row"><span>Diesel YoY change</span><span>${formatNumber(historical.diesel_yoy_change, 1)}%</span></div>
    </div>
    <div class="carrier-stat-footer carrier-stat-note">Updated ${formatMaybeDate(diesel.updated_at || gasoline.updated_at || payload?.data?.timestamp)}</div>
  `;
}

function renderDisruptionsCard(result) {
  if (!result || result.error) {
    return renderCarrierStatError('DISRUPTIONS', 'Unable to load live disruption data.');
  }

  const payload = result.data || {};
  const data = payload?.data?.data || {};
  const alerts = data.alerts || [];
  const topAlerts = alerts.slice(0, 3);
  const riskForecast = data.risk_forecast || {};

  return `
    <div class="carrier-stat-card-header">
      <span class="carrier-stat-kicker">DISRUPTIONS</span>
      <span class="carrier-stat-pill">${formatNumber(data.active_alerts, 0)} active</span>
    </div>
    <div class="carrier-stat-metrics">
      <div><strong>${formatNumber(data.active_alerts, 0)}</strong><span>Alerts</span></div>
      <div><strong>${(riskForecast.next_7_days || 'n/a').toUpperCase()}</strong><span>7d outlook</span></div>
      <div><strong>${formatNumber((data.resolved_recently || []).length, 0)}</strong><span>Resolved</span></div>
    </div>
    <div class="carrier-stat-list">
      ${topAlerts.map((alert) => `
        <div class="carrier-stat-row">
          <span>${alert.title || 'Alert'}</span>
          <span>${String(alert.severity || 'low').toUpperCase()}</span>
        </div>
      `).join('')}
    </div>
    <div class="carrier-stat-footer">
      ${renderChips(riskForecast.factors || [], 'severity-low')}
      <div class="carrier-stat-note" style="margin-top: 0.5rem;">${riskForecast.next_7_days ? `Forecast: ${riskForecast.next_7_days}` : 'Risk forecast unavailable.'}</div>
    </div>
  `;
}

function renderCarriersCard(result) {
  if (!result || result.error) {
    return renderCarrierStatError('CARRIERS', 'Unable to load live carrier reliability data.');
  }

  const payload = result.data || {};
  const data = payload?.data?.data || {};
  const ocean = data.ocean || [];
  const trucking = data.trucking || [];
  const air = data.air || [];

  const topOcean = [...ocean].sort((left, right) => (right.reliability_score || 0) - (left.reliability_score || 0))[0];
  const topTrucking = [...trucking].sort((left, right) => (right.reliability_score || 0) - (left.reliability_score || 0))[0];
  const topAir = [...air].sort((left, right) => (right.reliability_score || 0) - (left.reliability_score || 0))[0];

  return `
    <div class="carrier-stat-card-header">
      <span class="carrier-stat-kicker">CARRIER RELIABILITY</span>
      <span class="carrier-stat-pill">${formatNumber(ocean.length + trucking.length + air.length, 0)} carriers</span>
    </div>
    <div class="carrier-stat-metrics">
      <div><strong>${formatNumber(ocean.length, 0)}</strong><span>Ocean</span></div>
      <div><strong>${formatNumber(trucking.length, 0)}</strong><span>Trucking</span></div>
      <div><strong>${formatNumber(air.length, 0)}</strong><span>Air</span></div>
    </div>
    <div class="carrier-stat-list">
      <div class="carrier-stat-row"><span>${topOcean?.name || 'Ocean carrier'}</span><span>${formatNumber(topOcean?.reliability_score, 0)}</span></div>
      <div class="carrier-stat-row"><span>${topTrucking?.name || 'Trucking carrier'}</span><span>${formatNumber(topTrucking?.reliability_score, 0)}</span></div>
      <div class="carrier-stat-row"><span>${topAir?.name || 'Air carrier'}</span><span>${formatNumber(topAir?.reliability_score, 0)}</span></div>
    </div>
    <div class="carrier-stat-footer carrier-stat-note">Top reliability snapshot across ocean, trucking, and air carriers.</div>
  `;
}

function renderCarrierStatsLogs(result) {
  if (!carrierStatsLogTimeline) return;

  if (!result || result.error) {
    carrierStatsLogTimeline.innerHTML = `<p class="placeholder-text alert-error" style="background: none;">Failed to load carrier stats logs.</p>`;
    return;
  }

  const logs = result.data || [];
  if (!Array.isArray(logs) || logs.length === 0) {
    carrierStatsLogTimeline.innerHTML = `<p class="placeholder-text">No carrier stats logs recorded yet.</p>`;
    return;
  }

  carrierStatsLogTimeline.innerHTML = logs.map((log) => {
    const logDate = formatMaybeDate(log.created_at);
    return `
      <div class="timeline-item">
        <div class="timeline-title">${log.endpoint || 'Unknown endpoint'}</div>
        <div class="timeline-meta">${logDate}</div>
        <div class="timeline-body" style="margin-top: 0.35rem;">Status ${log.status_code || 'N/A'} | ${log.duration_ms || 0} ms | ${log.response_size || 0} bytes</div>
        <div class="timeline-meta" style="font-size: 0.7rem; color: var(--text-secondary); margin-top: 0.25rem;">
          Success: <strong style="color: ${log.success ? 'var(--accent-emerald)' : 'var(--accent-danger)'}">${log.success ? 'yes' : 'no'}</strong>
        </div>
        ${log.error ? `<div class="timeline-body" style="color: var(--accent-danger); margin-top: 0.25rem;">${log.error}</div>` : ''}
        ${log.response_preview ? `<div class="timeline-body" style="margin-top: 0.25rem; font-size: 0.72rem; opacity: 0.75;">${log.response_preview}</div>` : ''}
      </div>
    `;
  }).join('');
}
