// API client for AVEX Go backend
// Uses Next.js proxy routes (/api/*) which forward to Go backend

let authToken: string | null = null

export function setAuthToken(token: string | null) {
  authToken = token
  if (typeof window !== 'undefined') {
    if (token) localStorage.setItem('avex_token', token)
    else localStorage.removeItem('avex_token')
  }
}

export function getAuthToken(): string | null {
  if (authToken) return authToken
  if (typeof window !== 'undefined') {
    authToken = localStorage.getItem('avex_token')
  }
  return authToken
}

export interface User {
  id: string
  name: string
  phone: string
  email: string
  loyaltyPoints: number
  isAdmin: boolean
  createdAt: string
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
    const error = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export const authAPI = {
  register: (data: { name: string; phone: string; password: string; email?: string }) =>
    apiFetch<{ token: string; user: User }>('/api/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (data: { phone: string; password: string }) =>
    apiFetch<{ token: string; user: User }>('/api/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  me: () => apiFetch<User>('/api/auth/me'),
}

export const menuAPI = {
  getCategories: () => apiFetch<{ categories: any[] }>('/api/menu'),
  getSettings: () => apiFetch<{ settings: Record<string, string> }>('/api/settings'),
  getRestaurants: () => apiFetch<{ restaurants: any[] }>('/api/restaurants'),
  getRestaurant: (id: string) => apiFetch<any>(`/api/restaurants/${id}`),
}

export const ordersAPI = {
  create: (data: any) => apiFetch<{ order: any }>('/api/orders', { method: 'POST', body: JSON.stringify(data) }),
  getMyOrders: () => apiFetch<{ orders: any[] }>('/api/orders'),
  trackByNumber: (orderNumber: string) => apiFetch<{ order: any }>(`/api/orders/track?number=${encodeURIComponent(orderNumber)}`),
}

export const couponsAPI = {
  validate: (code: string, subtotal: number) =>
    apiFetch<{ valid: boolean; discount: number; code: string; descriptionAr: string }>('/api/coupons/validate', { method: 'POST', body: JSON.stringify({ code, subtotal }) }),
}

export const userAPI = {
  getAddresses: () => apiFetch<{ addresses: any[] }>('/api/addresses'),
  saveAddress: (data: any) => apiFetch<{ id: string }>('/api/addresses', { method: 'POST', body: JSON.stringify(data) }),
  deleteAddress: (id: string) => apiFetch(`/api/addresses/${id}`, { method: 'DELETE' }),
  getFavorites: () => apiFetch<{ favorites: any[] }>('/api/favorites'),
  toggleFavorite: (menuItemId: string) => apiFetch<{ favorited: boolean }>(`/api/favorites/${menuItemId}/toggle`, { method: 'POST' }),
  getCards: () => apiFetch<{ cards: any[] }>('/api/cards'),
  saveCard: (data: any) => apiFetch<{ id: string }>('/api/cards', { method: 'POST', body: JSON.stringify(data) }),
  deleteCard: (id: string) => apiFetch(`/api/cards/${id}`, { method: 'DELETE' }),
  setDefaultCard: (id: string) => apiFetch(`/api/cards/${id}/default`, { method: 'POST' }),
  createPaymentKey: (data: any) => apiFetch<{ iframeUrl: string; paymentToken: string }>('/api/paymob/payment-key', { method: 'POST', body: JSON.stringify(data) }),
}

export const adminAPI = {
  getCategories: () => apiFetch<{ categories: any[] }>('/api/admin/categories'),
  createCategory: (data: any) => apiFetch<{ id: string }>('/api/admin/categories', { method: 'POST', body: JSON.stringify(data) }),
  updateCategory: (id: string, data: any) => apiFetch(`/api/admin/categories/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  deleteCategory: (id: string) => apiFetch(`/api/admin/categories/${id}`, { method: 'DELETE' }),
  getMenuItems: () => apiFetch<{ items: any[] }>('/api/admin/menu-items'),
  createMenuItem: (data: any) => apiFetch<{ id: string }>('/api/admin/menu-items', { method: 'POST', body: JSON.stringify(data) }),
  updateMenuItem: (id: string, data: any) => apiFetch(`/api/admin/menu-items/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  deleteMenuItem: (id: string) => apiFetch(`/api/admin/menu-items/${id}`, { method: 'DELETE' }),
  getCoupons: () => apiFetch<{ coupons: any[] }>('/api/admin/coupons'),
  createCoupon: (data: any) => apiFetch<{ id: string }>('/api/admin/coupons', { method: 'POST', body: JSON.stringify(data) }),
  updateCoupon: (id: string, data: any) => apiFetch(`/api/admin/coupons/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  deleteCoupon: (id: string) => apiFetch(`/api/admin/coupons/${id}`, { method: 'DELETE' }),
  updateOrderStatus: (id: string, status: string) => apiFetch(`/api/orders/${id}`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  updateSetting: (key: string, value: string) => apiFetch('/api/admin/settings', { method: 'PUT', body: JSON.stringify({ key, value }) }),
}

export async function uploadImage(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  const token = getAuthToken()
  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch('/api/upload', { method: 'POST', headers, body: formData })
  if (!res.ok) throw new Error('Upload failed')
  const data = await res.json()
  return data.url
}
