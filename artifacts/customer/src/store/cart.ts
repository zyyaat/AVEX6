
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface CartItem {
  id: string
  name: string
  nameAr: string
  price: number
  image: string
  quantity: number
}

interface CartState {
  items: CartItem[]
  isOpen: boolean
  addItem: (item: Omit<CartItem, 'quantity'>) => void
  removeItem: (id: string) => void
  updateQuantity: (id: string, quantity: number) => void
  clearCart: () => void
  setOpen: (open: boolean) => void
  getTotalItems: () => number
  getSubtotal: () => number
  getDeliveryFee: () => number
  getTotal: () => number
}

export const useCart = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],
      isOpen: false,

      addItem: (item) => {
        const existing = get().items.find((i) => i.id === item.id)
        if (existing) {
          set({
            items: get().items.map((i) =>
              i.id === item.id ? { ...i, quantity: i.quantity + 1 } : i
            ),
          })
        } else {
          set({ items: [...get().items, { ...item, quantity: 1 }] })
        }
      },

      removeItem: (id) => {
        set({ items: get().items.filter((i) => i.id !== id) })
      },

      updateQuantity: (id, quantity) => {
        // منع الحذف بالخطأ عبر زر الناقص - أقل كمية هي 1
        // الحذف يتم فقط عبر زر الحذف (removeItem)
        if (quantity < 1) {
          return // لا تفعل شيئاً، الكمية تبقى 1
        }
        set({
          items: get().items.map((i) =>
            i.id === id ? { ...i, quantity } : i
          ),
        })
      },

      clearCart: () => set({ items: [] }),
      setOpen: (open) => set({ isOpen: open }),

      getTotalItems: () =>
        get().items.reduce((sum, i) => sum + i.quantity, 0),

      getSubtotal: () =>
        get().items.reduce((sum, i) => sum + i.price * i.quantity, 0),

      getDeliveryFee: () => {
        const subtotal = get().getSubtotal()
        if (subtotal === 0) return 0
        return subtotal >= 30 ? 0 : 3.99
      },

      getTotal: () => get().getSubtotal() + get().getDeliveryFee(),
    }),
    {
      name: 'avex-cart',
      // Only persist items, not isOpen (cart should be closed on page load)
      partialize: (state) => ({ items: state.items }),
    }
  )
)
