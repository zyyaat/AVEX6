// AVEX Support - API client
let authToken: string | null = null
export function setAuthToken(t: string | null) {
  authToken = t
  if (typeof window !== 'undefined') {
    if (t) localStorage.setItem('avex_agent_token', t)
    else localStorage.removeItem('avex_agent_token')
  }
}
export function getAuthToken(): string | null {
  if (authToken) return authToken
  if (typeof window !== 'undefined') authToken = localStorage.getItem('avex_agent_token')
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

export const agentAuthAPI = {
  login: (data: { phone: string; password: string }) =>
    apiFetch<{ token: string; mustChangePassword: boolean; agent: any }>('/api/agent/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  me: () => apiFetch<any>('/api/agent/me'),
}

export const agentAPI = {
  getStats: () => apiFetch<any>('/api/agent/stats'),
  getTickets: (filter: string = '') => apiFetch<{ tickets: any[]; agentId: string }>(`/api/agent/tickets${filter ? `?filter=${filter}` : ''}`),
  getTicket: (id: string) => apiFetch<{ ticket: any; messages: any[] }>(`/api/agent/tickets/${id}`),
  assignTicket: (id: string) => apiFetch<{ success: boolean; assignedTo: string }>(`/api/agent/tickets/${id}/assign`, { method: 'POST' }),
  setPriority: (id: string, priority: string) => apiFetch(`/api/agent/tickets/${id}/priority`, { method: 'PATCH', body: JSON.stringify({ priority }) }),
  sendMessage: (id: string, body: string, isInternal: boolean = false) =>
    apiFetch<{ id: string }>(`/api/agent/tickets/${id}/messages`, { method: 'POST', body: JSON.stringify({ body, isInternal }) }),
  resolveTicket: (id: string, notes: string) => apiFetch(`/api/agent/tickets/${id}/resolve`, { method: 'PATCH', body: JSON.stringify({ adminNotes: notes }) }),
  cancelOrder: (id: string) => apiFetch(`/api/agent/tickets/${id}/cancel-order`, { method: 'POST' }),
  search: (q: string) => apiFetch<{ customers: any[]; drivers: any[]; orders: any[] }>(`/api/agent/search?q=${encodeURIComponent(q)}`),
  getOrder: (id: string) => apiFetch<{ order: any }>(`/api/agent/orders/${id}`),
  getDriver: (id: string) => apiFetch<{ driver: any; stats: any; recentOrders: any[] }>(`/api/agent/drivers/${id}`),
}
