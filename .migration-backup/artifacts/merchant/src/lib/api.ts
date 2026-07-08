// AVEX Merchant — API client
let authToken: string | null = null

export function setAuthToken(t: string | null) {
  authToken = t
  if (typeof window !== 'undefined') {
    if (t) localStorage.setItem('avex_merchant_token', t)
    else localStorage.removeItem('avex_merchant_token')
  }
}

export function getAuthToken(): string | null {
  if (authToken) return authToken
  if (typeof window !== 'undefined') authToken = localStorage.getItem('avex_merchant_token')
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

// ===== Types =====
export interface Merchant {
  id: string
  name: string
  phone: string
  isActive: boolean
  mustChangePassword: boolean
  restaurant: {
    id: string
    name: string
    nameAr: string
    descriptionAr: string
    lat: number
    lng: number
    rating: number
    ratingCount: number
    isActive: boolean
    isPro: boolean
    deliveryTimeMin: number
    deliveryTimeMax: number
    deliveryFee: number
    minOrder: number
  }
}

export interface MerchantOrder {
  id: string
  orderNumber: string
  customerName: string
  phone: string
  locationAddress: string
  locationLat: number
  locationLng: number
  locationUrl: string
  subtotal: number
  deliveryFee: number
  discount: number
  total: number
  paymentMethod: string
  status: string
  createdAt: string
  updatedAt: string
  driverId: string
  scheduledFor: string
  itemsSummary: string
  itemsCount: number
}

export interface OrderItem {
  id: string
  menuItemId: string
  name: string
  price: number
  quantity: number
}

export interface MenuItem {
  id: string
  name: string
  nameAr: string
  description: string
  descriptionAr: string
  price: number
  image: string
  imageUrl: string
  isPopular: boolean
  isAvailable: boolean
  rating: number
  ratingCount: number
  prepTime: number
  calories: number
  categoryId: string
}

export interface Category {
  id: string
  name: string
  nameAr: string
  icon: string
}

export interface StoreHour {
  id: string
  dayOfWeek: number
  openTime: string
  closeTime: string
  isOpen: boolean
}

export interface MerchantStats {
  todayCount: number
  activeCount: number
  completedCount: number
  todayRevenue: number
  daily: { date: string; revenue: number; count: number }[] | null
}

export const merchantAuthAPI = {
  login: (data: { phone: string; password: string }) =>
    apiFetch<{ token: string; mustChangePassword: boolean; merchant: any }>('/api/merchant/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  changePassword: (data: { oldPassword: string; newPassword: string }) =>
    apiFetch<{ success: boolean }>('/api/merchant/auth/change-password', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  me: () => apiFetch<Merchant>('/api/merchant/me'),
}

export const merchantAPI = {
  // Orders
  getOrders: (status?: string) =>
    apiFetch<{ orders: MerchantOrder[] | null }>(`/api/merchant/orders${status ? `?status=${status}` : ''}`),
  getOrderItems: (id: string) =>
    apiFetch<{ items: OrderItem[] | null }>(`/api/merchant/orders/${id}/items`),
  updateOrderStatus: (id: string, status: string) =>
    apiFetch<{ success: boolean; status: string }>(`/api/merchant/orders/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),

  // Menu
  getMenu: () =>
    apiFetch<{ items: MenuItem[] | null; categories: Category[] | null }>('/api/merchant/menu'),
  createMenuItem: (data: Partial<MenuItem>) =>
    apiFetch<{ id: string }>('/api/merchant/menu/items', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateMenuItem: (id: string, data: Partial<MenuItem>) =>
    apiFetch<{ success: boolean }>(`/api/merchant/menu/items/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),
  deleteMenuItem: (id: string) =>
    apiFetch<{ success: boolean }>(`/api/merchant/menu/items/${id}`, { method: 'DELETE' }),

  // Store
  getHours: () => apiFetch<{ hours: StoreHour[] | null }>('/api/merchant/hours'),
  updateHours: (hours: any[]) =>
    apiFetch<{ success: boolean }>('/api/merchant/hours', {
      method: 'PUT',
      body: JSON.stringify({ hours }),
    }),
  togglePause: (isActive: boolean) =>
    apiFetch<{ isActive: boolean }>('/api/merchant/pause', {
      method: 'PATCH',
      body: JSON.stringify({ isActive }),
    }),
  getStats: () => apiFetch<MerchantStats>('/api/merchant/stats'),
  getScheduledOrders: () =>
    apiFetch<{ scheduledOrders: any[] | null }>('/api/merchant/scheduled-orders'),
}
