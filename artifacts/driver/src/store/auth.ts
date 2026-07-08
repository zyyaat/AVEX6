
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { driverAuthAPI, setAuthToken } from '@/lib/api'

interface AuthState {
  token: string | null
  driverId: string | null
  driverName: string | null
  driverPhone: string | null
  mustChangePassword: boolean
  isLoading: boolean
  isAuthenticated: boolean

  login: (phone: string, password: string) => Promise<{ mustChangePassword: boolean }>
  logout: () => void
  setMustChangePassword: (v: boolean) => void
  initialize: () => Promise<void>
}

export const useAuth = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      driverId: null,
      driverName: null,
      driverPhone: null,
      mustChangePassword: false,
      isLoading: false,
      isAuthenticated: false,

      login: async (phone, password) => {
        set({ isLoading: true })
        try {
          const { token, mustChangePassword, driver } = await driverAuthAPI.login({ phone, password })
          setAuthToken(token)
          set({
            token, driverId: driver.id, driverName: driver.name, driverPhone: driver.phone,
            mustChangePassword, isAuthenticated: true, isLoading: false,
          })
          return { mustChangePassword }
        } catch (err) {
          set({ isLoading: false })
          throw err
        }
      },

      logout: () => {
        setAuthToken(null)
        set({
          token: null, driverId: null, driverName: null, driverPhone: null,
          mustChangePassword: false, isAuthenticated: false,
        })
      },

      setMustChangePassword: (v) => set({ mustChangePassword: v }),

      initialize: async () => {
        const token = get().token
        if (token) {
          setAuthToken(token)
        }
      },
    }),
    {
      name: 'avex-driver-auth',
      partialize: (state) => ({
        token: state.token,
        driverId: state.driverId,
        driverName: state.driverName,
        driverPhone: state.driverPhone,
        mustChangePassword: state.mustChangePassword,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)
