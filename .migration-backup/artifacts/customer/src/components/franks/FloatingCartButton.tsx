
import { motion, AnimatePresence } from 'framer-motion'
import { ShoppingBag, ChevronLeft } from 'lucide-react'
import { useCart } from '@/store/cart'

export function FloatingCartButton() {
  const { getTotalItems, getTotal, setOpen } = useCart()
  const count = getTotalItems()
  const total = getTotal()

  return (
    <AnimatePresence>
      {count > 0 && (
        <motion.div
          initial={{ y: 100, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: 100, opacity: 0 }}
          className="fixed left-3 right-3 z-50 sm:hidden"
          style={{ bottom: 'calc(60px + env(safe-area-inset-bottom, 0px))' }}
        >
          <button onClick={() => setOpen(true)} className="w-full bg-black text-white rounded-lg shadow-fluent-lg px-4 py-3 flex items-center justify-between gap-3 hover:bg-gray-800 active:scale-[0.98] transition-fluent">
            <div className="flex items-center gap-3">
              <div className="relative">
                <div className="w-9 h-9 rounded-md bg-gray-800 flex items-center justify-center">
                  <ShoppingBag className="w-4 h-4" />
                </div>
                <span className="absolute -top-1 -right-1 min-w-4 h-4 px-1 rounded-full bg-white text-black text-[10px] font-bold flex items-center justify-center border border-black">{count}</span>
              </div>
              <div className="text-right">
                <p className="text-[10px] text-gray-400 leading-none">عرض السلة</p>
                <p className="font-medium text-sm leading-none mt-0.5">{total.toFixed(2)} ج.م</p>
              </div>
            </div>
            <div className="flex items-center gap-1 text-xs font-medium">
              متابعة
              <ChevronLeft className="w-3.5 h-3.5" />
            </div>
          </button>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
