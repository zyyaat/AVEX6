
import { create } from 'zustand'
import { driverAPI, type Driver, type Offer, type ActiveOrder } from '@/lib/api'

interface DriverState {
  driver: Driver | null
  offers: Offer[]
  activeOrder: ActiveOrder | null
  isLoading: boolean
  error: string | null

  fetchMe: () => Promise<void>
  setOnline: (online: boolean) => Promise<void>
  updateLocation: (lat: number, lng: number) => Promise<void>
  setAutoAccept: (v: boolean) => Promise<void>
  refreshOffers: () => Promise<void>
  refreshActiveOrder: () => Promise<void>
  acceptOffer: (offerId: string) => Promise<string>
  rejectOffer: (offerId: string) => Promise<void>
  pickedUp: (orderId: string) => Promise<void>
  arrived: (orderId: string) => Promise<void>
  delivered: (orderId: string) => Promise<number>
  clear: () => void
}

export const useDriver = create<DriverState>((set, get) => ({
  driver: null,
  offers: [],
  activeOrder: null,
  isLoading: false,
  error: null,

  fetchMe: async () => {
    try {
      const driver = await driverAPI.me()
      set({ driver, error: null })
    } catch (err: any) {
      set({ error: err.message })
    }
  },

  setOnline: async (online) => {
    await driverAPI.toggleOnline(online)
    const driver = get().driver
    if (driver) set({ driver: { ...driver, isOnline: online } })
    if (!online) set({ offers: [] })
  },

  updateLocation: async (lat, lng) => {
    try {
      await driverAPI.updateLocation(lat, lng)
      const driver = get().driver
      if (driver) set({ driver: { ...driver, lat, lng } })
    } catch {}
  },

  setAutoAccept: async (v) => {
    await driverAPI.toggleAutoAccept(v)
    const driver = get().driver
    if (driver) set({ driver: { ...driver, autoAccept: v } })
  },

  refreshOffers: async () => {
    try {
      const { offers } = await driverAPI.getOffers()
      set({ offers: offers || [] })
    } catch {}
  },

  refreshActiveOrder: async () => {
    try {
      const { order } = await driverAPI.getActiveOrder()
      set({ activeOrder: order })
    } catch {}
  },

  acceptOffer: async (offerId) => {
    const { orderId } = await driverAPI.acceptOffer(offerId)
    set({ offers: [] })
    await get().refreshActiveOrder()
    return orderId
  },

  rejectOffer: async (offerId) => {
    await driverAPI.rejectOffer(offerId)
    set((s) => ({ offers: s.offers.filter((o) => o.offerId !== offerId) }))
  },

  pickedUp: async (orderId) => {
    await driverAPI.pickedUp(orderId)
    await get().refreshActiveOrder()
  },

  arrived: async (orderId) => {
    await driverAPI.arrived(orderId)
    await get().refreshActiveOrder()
  },

  delivered: async (orderId) => {
    const { earnings } = await driverAPI.delivered(orderId)
    set({ activeOrder: null })
    await get().fetchMe()
    return earnings
  },

  clear: () => set({ driver: null, offers: [], activeOrder: null }),
}))
