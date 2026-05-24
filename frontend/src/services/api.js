const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

let authToken = 'Bearer test-token'

export const setAuthToken = (token) => {
    authToken = token
}

const apiCall = async (method, endpoint, body = null) => {
    const url = `${API_BASE}${endpoint}`
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json',
            'Authorization': authToken,
        },
    }

    if (body) {
        options.body = JSON.stringify(body)
    }

    const response = await fetch(url, options)
    const data = await response.json()

    if (!response.ok) {
        throw new Error(data.error || `API error: ${response.status}`)
    }

    return data
}

export const shipments = {
    list: () => apiCall('GET', '/shipments'),
    create: (payload) => apiCall('POST', '/shipments', payload),
    get: (id) => apiCall('GET', `/shipments/${id}`),
    update: (id, payload) => apiCall('PUT', `/shipments/${id}`, payload),
    updateStatus: (id, status) => apiCall('PATCH', `/shipments/${id}/status`, { status }),
}

export const carriers = {
    register: (payload) => apiCall('POST', '/carriers', payload),
    getRates: (params) => {
        const query = new URLSearchParams(params).toString()
        return apiCall('GET', `/carriers/rates?${query}`)
    },
    getTracking: (carrier, trackingNumber) => {
        return apiCall('GET', `/carriers/tracking?carrier=${carrier}&tracking_number=${trackingNumber}`)
    },
    getPickupLocations: (carrier, address, limit = 10) => {
        return apiCall('GET', `/carriers/pickup-locations?carrier=${carrier}&address=${address}&limit=${limit}`)
    },
    getDropLocations: (carrier, address, limit = 10) => {
        return apiCall('GET', `/carriers/drop-locations?carrier=${carrier}&address=${address}&limit=${limit}`)
    },
}

export const rates = {
    compare: (payload) => apiCall('POST', '/rates/compare', payload),
}

export const labels = {
    generate: (payload) => apiCall('POST', '/labels', payload),
    get: (id) => apiCall('GET', `/labels/${id}`),
    download: (id) => `${API_BASE}/labels/${id}/download`,
}

export const tracking = {
    getHistory: (shipmentId) => apiCall('GET', `/tracking/${shipmentId}`),
    getInfo: (shipmentId) => apiCall('GET', `/tracking/${shipmentId}/history`),
}

export const addresses = {
    validate: (payload) => apiCall('POST', '/addresses/validate', payload),
    getPickupLocations: (address, limit = 10) => {
        return apiCall('GET', `/addresses/pickup-locations?address=${address}&limit=${limit}`)
    },
    getDropLocations: (address, limit = 10) => {
        return apiCall('GET', `/addresses/drop-locations?address=${address}&limit=${limit}`)
    },
}

export const billing = {
    createInvoice: (payload) => apiCall('POST', '/billing/invoices', payload),
    getInvoice: (id) => apiCall('GET', `/billing/invoices/${id}`),
    processPayment: (payload) => apiCall('POST', '/billing/payments', payload),
}

export const returns = {
    create: (payload) => apiCall('POST', '/returns', payload),
    get: (id) => apiCall('GET', `/returns/${id}`),
    approve: (id) => apiCall('POST', `/returns/${id}/approve`),
}
