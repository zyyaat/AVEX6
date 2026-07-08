// AVEX Admin - API client
let authToken: string | null = null
export function setAuthToken(t: string | null) {
  authToken = t
  if (typeof window !== 'undefined') {
    if (t) localStorage.setItem('avex_admin_token', t)
    else localStorage.removeItem('avex_admin_token')
  }
}
export function getAuthToken(): string | null {
  if (authToken) return authToken
  if (typeof window !== 'undefined') authToken = localStorage.getItem('avex_admin_token')
  return authToken
}
async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = getAuthToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(endpoint, { ...options, headers })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export const adminAuthAPI = {
  login: (data: { phone: string; password: string }) =>
    apiFetch<{ token: string; user: any }>('/api/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  me: () => apiFetch<any>('/api/auth/me'),
}

export const adminAPI = {
  // Dashboard
  getDashboard: () => apiFetch<any>('/api/admin/dashboard'),
  getOrders: (status?: string) => apiFetch<{ orders: any[] }>(`/api/admin/orders${status ? `?status=${status}` : ''}`),

  // Zones
  getZones: () => apiFetch<{ zones: any[] }>('/api/admin/zones'),
  createZone: (data: any) => apiFetch<{ id: string }>('/api/admin/zones', { method: 'POST', body: JSON.stringify(data) }),
  updateZone: (id: string, data: any) => apiFetch(`/api/admin/zones/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  deleteZone: (id: string) => apiFetch(`/api/admin/zones/${id}`, { method: 'DELETE' }),

  // Tiers
  getTiers: () => apiFetch<{ tiers: any[] }>('/api/admin/tiers'),
  createTier: (data: any) => apiFetch<{ id: string }>('/api/admin/tiers', { method: 'POST', body: JSON.stringify(data) }),
  updateTier: (id: string, data: any) => apiFetch(`/api/admin/tiers/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  updateTierThresholds: (id: string, data: any) => apiFetch(`/api/admin/tiers/${id}/thresholds`, { method: 'PUT', body: JSON.stringify(data) }),

  // Tier Prices
  getTierPrices: (zoneId?: string) => apiFetch<{ prices: any[] }>(`/api/admin/tier-prices${zoneId ? `?zone_id=${zoneId}` : ''}`),
  updateTierPrice: (tierId: string, zoneId: string, data: any) =>
    apiFetch(`/api/admin/tier-prices/${tierId}/${zoneId}`, { method: 'PUT', body: JSON.stringify(data) }),

  // Drivers
  getDrivers: () => apiFetch<{ drivers: any[] }>('/api/admin/drivers'),
  updateDriverStatus: (id: string, isActive: boolean) =>
    apiFetch(`/api/admin/drivers/${id}/status`, { method: 'PATCH', body: JSON.stringify({ isActive }) }),
  updateDriverTier: (id: string, tierId: string) =>
    apiFetch(`/api/admin/drivers/${id}/tier`, { method: 'PATCH', body: JSON.stringify({ tierId }) }),
  getDriverTierHistory: (id: string) => apiFetch<{ history: any[] }>(`/api/admin/drivers/${id}/tier-history`),
  createShift: (id: string, data: any) => apiFetch<{ id: string }>(`/api/admin/drivers/${id}/shifts`, { method: 'POST', body: JSON.stringify(data) }),
  getShifts: (id: string) => apiFetch<{ shifts: any[] }>(`/api/admin/drivers/${id}/shifts`),

  // Applications
  getApplications: () => apiFetch<{ applications: any[] }>('/api/admin/driver-applications'),
  createApplication: (data: any) => apiFetch<{ id: string }>('/api/admin/driver-applications', { method: 'POST', body: JSON.stringify(data) }),
  verifyApplication: (id: string) => apiFetch<{ success: boolean; driverId: string; initialPassword: string }>(`/api/admin/driver-applications/${id}/verify`, { method: 'PATCH' }),
  rejectApplication: (id: string, reason: string) => apiFetch(`/api/admin/driver-applications/${id}/reject`, { method: 'PATCH', body: JSON.stringify({ reason }) }),

  // Restaurants
  getRestaurants: () => apiFetch<{ restaurants: any[] }>('/api/admin/restaurants'),
  createRestaurant: (data: any) => apiFetch<{ id: string; merchantPhone: string; merchantPassword: string }>('/api/admin/restaurants', { method: 'POST', body: JSON.stringify(data) }),
  updateRestaurant: (id: string, data: any) => apiFetch(`/api/admin/restaurants/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  deleteRestaurant: (id: string) => apiFetch(`/api/admin/restaurants/${id}`, { method: 'DELETE' }),

  // Support
  getTickets: () => apiFetch<{ tickets: any[] }>('/api/admin/support/tickets'),
  resolveTicket: (id: string, notes: string) => apiFetch(`/api/admin/support/tickets/${id}/resolve`, { method: 'PATCH', body: JSON.stringify({ adminNotes: notes }) }),
  sendMessage: (id: string, body: string) => apiFetch<{ id: string }>(`/api/admin/support/tickets/${id}/messages`, { method: 'POST', body: JSON.stringify({ body }) }),
  cancelOrder: (id: string) => apiFetch(`/api/admin/support/tickets/${id}/cancel-order`, { method: 'POST' }),

  // Settings
  updateSetting: (key: string, value: string) => apiFetch('/api/admin/settings', { method: 'PUT', body: JSON.stringify({ key, value }) }),
}
