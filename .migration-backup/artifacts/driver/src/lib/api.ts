// AVEX Driver - API client for Go backend
// Uses Next.js proxy routes (/api/*) which forward to Go backend

let authToken: string | null = null

export function setAuthToken(token: string | null) {
  authToken = token
  if (typeof window !== 'undefined') {
    if (token) localStorage.setItem('avex_driver_token', token)
    else localStorage.removeItem('avex_driver_token')
  }
}

export function getAuthToken(): string | null {
  if (authToken) return authToken
  if (typeof window !== 'undefined') {
    authToken = localStorage.getItem('avex_driver_token')
  }
  return authToken
}

// ===== TYPES =====
export interface Driver {
  id: string
  name: string
  phone: string
  tier: {
    id: string
    nameAr: string
    color: string
    sortOrder: number
  }
  isOnline: boolean
  isActive: boolean
  isVerified: boolean
  autoAccept: boolean
  mustChangePassword: boolean
  lat: number
  lng: number
  createdAt: string
  lastSeen: string
  locationUpdatedAt: string
  stats: {
    acceptedOrders: number
    rejectedOrders: number
    completedOrders: number
    ratingCount: number
    rating: number
    onTimeRate: number
    acceptanceRate: number
    completionRate: number
    shiftAdherence: number
    totalEarnings: number
    lifetimeOrders: number
  }
  nextTier: {
    id: string
    nameAr: string
    sortOrder: number
    minAcceptanceRate: number
    minCompletionRate: number
    minCustomerRating: number
    minOnTimeRate: number
    minShiftAdherence: number
    minLifetimeOrders: number
  } | null
}

export interface Offer {
  offerId: string
  orderId: string
  orderNumber: string
  customerName: string
  phone: string
  locationLat: number
  locationLng: number
  locationUrl: string
  locationAddress: string
  subtotal: number
  deliveryFee: number
  total: number
  paymentMethod: string
  status: string
  restaurantName: string
  restaurantLat: number
  restaurantLng: number
  zoneName: string
  itemsSummary: string
  offeredAt: string
  expiresAt: string
  distanceM: number
  driverFee: number
  estimatedDeliveryDistanceM: number
}

export interface ActiveOrder {
  id: string
  orderNumber: string
  customerName: string
  phone: string
  locationUrl: string
  locationAddress: string
  locationLat: number
  locationLng: number
  paymentMethod: string
  status: string
  subtotal: number
  deliveryFee: number
  total: number
  driverFee: number
  dispatchDistanceM: number
  deliveryDistanceM: number
  createdAt: string
  restaurantName: string
  restaurantLat: number
  restaurantLng: number
  items: { name: string; price: number; quantity: number }[]
}

export interface Shift {
  id: string
  zoneId: string
  zoneName: string
  date: string
  startTime: string
  endTime: string
  status: string
  isCheckedIn: boolean
  isLate: boolean
  lateMinutes: number
}

export interface SupportTicket {
  id: string
  orderId?: string
  type: string
  reason: string
  status: string
  createdAt: string
  resolvedAt?: string
}

export interface SupportMessage {
  id: string
  sender: string
  body: string
  createdAt: string
}

// ===== CORE FETCH =====
async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = getAuthToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(endpoint, { ...options, headers })
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || `HTTP ${res.status}`)
  }
  return res.json()
}

// ===== API OBJECTS =====
export const driverAuthAPI = {
  login: (data: { phone: string; password: string }) =>
    apiFetch<{ token: string; mustChangePassword: boolean; driver: { id: string; name: string; phone: string } }>(
      '/api/driver/auth/login', { method: 'POST', body: JSON.stringify(data) }
    ),
  changePassword: (data: { oldPassword: string; newPassword: string }) =>
    apiFetch<{ success: boolean }>('/api/driver/auth/change-password', { method: 'POST', body: JSON.stringify(data) }),
}

export const driverAPI = {
  me: () => apiFetch<Driver>('/api/driver/me'),
  toggleOnline: (online: boolean) =>
    apiFetch<{ online: boolean }>('/api/driver/online', { method: 'PATCH', body: JSON.stringify({ online }) }),
  updateLocation: (lat: number, lng: number) =>
    apiFetch<{ success: boolean }>('/api/driver/location', { method: 'PATCH', body: JSON.stringify({ lat, lng }) }),
  toggleAutoAccept: (autoAccept: boolean) =>
    apiFetch<{ autoAccept: boolean }>('/api/driver/auto-accept', { method: 'PATCH', body: JSON.stringify({ autoAccept }) }),
  getShift: () => apiFetch<{ shift: Shift | null }>('/api/driver/shift'),

  getOffers: () => apiFetch<{ offers: Offer[] | null }>('/api/driver/offers'),
  acceptOffer: (offerId: string) =>
    apiFetch<{ success: boolean; orderId: string }>(`/api/driver/offers/${offerId}/accept`, { method: 'POST' }),
  rejectOffer: (offerId: string) =>
    apiFetch<{ success: boolean }>(`/api/driver/offers/${offerId}/reject`, { method: 'POST'}),
  getActiveOrder: () => apiFetch<{ order: ActiveOrder | null }>('/api/driver/active-order'),
  pickedUp: (orderId: string) =>
    apiFetch<{ success: boolean; status: string; distance: number }>(`/api/driver/orders/${orderId}/picked-up`, { method: 'POST' }),
  arrived: (orderId: string) =>
    apiFetch<{ success: boolean; status: string }>(`/api/driver/orders/${orderId}/arrived`, { method: 'POST' }),
  delivered: (orderId: string) =>
    apiFetch<{ success: boolean; status: string; earnings: number }>(`/api/driver/orders/${orderId}/delivered`, { method: 'POST' }),

  getEarnings: (period: 'today' | 'week' | 'month' = 'today') =>
    apiFetch<{ period: string; totalEarnings: number; completedOrders: number }>(`/api/driver/earnings?period=${period}`),
  getHistory: (page = 1) =>
    apiFetch<{ orders: any[]; page: number }>(`/api/driver/history?page=${page}`),

  getTickets: () => apiFetch<{ tickets: SupportTicket[] | null }>('/api/driver/support/tickets'),
  createTicket: (data: { orderId?: string; type: string; reason: string }) =>
    apiFetch<{ id: string }>('/api/driver/support/tickets', { method: 'POST', body: JSON.stringify(data) }),
  getTicket: (id: string) =>
    apiFetch<{ ticket: SupportTicket; messages: SupportMessage[] | null }>(`/api/driver/support/tickets/${id}`),
  sendMessage: (id: string, body: string) =>
    apiFetch<{ id: string }>(`/api/driver/support/tickets/${id}/messages`, { method: 'POST', body: JSON.stringify({ body }) }),
}
