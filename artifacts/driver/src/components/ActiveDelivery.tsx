import { motion } from 'framer-motion'
import {
  Store, Package, MapPin, Phone, Navigation, ArrowRight,
  CheckCircle2, Loader2, Clock, AlertCircle, ChevronDown,
} from 'lucide-react'
import { useDriver } from '@/store/driver'
import { toast } from 'sonner'
import { useState } from 'react'

export function ActiveDelivery() {
  const { activeOrder, markPickedUp, markDelivered, refreshActiveOrder } = useDriver()
  const [busy, setBusy] = useState<string | null>(null)
  const [expanded, setExpanded] = useState(false)

  if (!activeOrder) return null

  const o = activeOrder

  const statusLabels: Record<string, string> = {
    assigned: 'تم القبول — اذهب للمطعم',
    picked_up: 'تم الاستلام — اذهب للعميل',
    on_the_way: 'في الطريق للعميل',
    delivering: 'في الطريق للعميل',
    delivered: 'تم التوصيل',
  }

  const handlePickedUp = async () => {
    setBusy('picked')
    try {
      await markPickedUp(o.id)
      toast.success('تم تأكيد الاستلام')
    } catch (err: any) {
      toast.error(err.message || 'فشل تأكيد الاستلام')
    } finally {
      setBusy(null)
    }
  }

  const handleDelivered = async () => {
    setBusy('delivered')
    try {
      await markDelivered(o.id)
      toast.success('تم التوصيل بنجاح! 🎉')
    } catch (err: any) {
      toast.error(err.message || 'فشل تأكيد التوصيل')
    } finally {
      setBusy(null)
    }
  }

  return (
    <motion.div
      initial={{ y: 100 }}
      animate={{ y: 0 }}
      className="bg-white rounded-t-2xl shadow-2xl"
      style={{ paddingBottom: 'env(safe-area-inset-bottom, 0px)' }}
      dir="rtl"
    >
      {/* Status bar */}
      <div
        className="px-5 py-2.5 flex items-center justify-between rounded-t-2xl"
        style={{ backgroundColor: '#FF6B35' }}
      >
        <span className="text-white font-medium text-sm">
          {statusLabels[o.status] || o.status}
        </span>
        <span className="text-white/80 text-xs font-mono">#{o.order_number}</span>
      </div>

      {/* Main content */}
      <div className="px-5 py-4">
        {/* Customer/Restaurant info */}
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
              {o.status === 'assigned' ? (
                <Store className="w-5 h-5 text-gray-600" />
              ) : (
                <Package className="w-5 h-5 text-gray-600" />
              )}
            </div>
            <div>
              <p className="font-bold text-gray-900 text-sm">
                {o.status === 'assigned' ? o.restaurant_name : o.customer_name}
              </p>
              <p className="text-xs text-gray-500">
                {o.status === 'assigned' ? 'المطعم' : 'العميل'}
              </p>
            </div>
          </div>

          {/* Call + Chat buttons */}
          <div className="flex items-center gap-2">
            <a
              href={`tel:${o.customer_phone}`}
              className="w-9 h-9 rounded-full flex items-center justify-center transition-colors"
              style={{ backgroundColor: '#FFF0E8' }}
            >
              <Phone className="w-4 h-4" style={{ color: '#FF6B35' }} />
            </a>
          </div>
        </div>

        {/* Address */}
        <div className="flex items-start gap-2 mb-3">
          <MapPin className="w-4 h-4 text-gray-400 mt-0.5 shrink-0" />
          <p className="text-sm text-gray-600 flex-1">{o.delivery_address}</p>
        </div>

        {/* Expandable details */}
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-1 text-xs text-gray-500 mb-2"
        >
          <span>التفاصيل</span>
          <ChevronDown className={`w-4 h-4 transition-transform ${expanded ? 'rotate-180' : ''}`} />
        </button>

        {expanded && (
          <div className="space-y-2 pt-2 border-t border-gray-100">
            {/* Phone */}
            <div className="flex items-center gap-2 text-sm">
              <Phone className="w-4 h-4 text-gray-400" />
              <span className="text-gray-600" dir="ltr">{o.customer_phone}</span>
            </div>
            {/* Order time */}
            <div className="flex items-center gap-2 text-sm">
              <Clock className="w-4 h-4 text-gray-400" />
              <span className="text-gray-600">
                {new Date(o.created_at).toLocaleTimeString('ar-EG', { hour: '2-digit', minute: '2-digit' })}
              </span>
            </div>
            {/* Total */}
            <div className="flex items-center justify-between text-sm pt-1">
              <span className="text-gray-500">الإجمالي</span>
              <span className="font-bold text-gray-900">
                {(o.total / 100).toFixed(2)} {o.currency}
              </span>
            </div>
            {/* Payment method */}
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-500">طريقة الدفع</span>
              <span className="text-gray-700">
                {o.payment_method === 'cash' ? 'نقدي' : o.payment_method === 'card' ? 'بطاقة' : 'محفظة'}
              </span>
            </div>
            {/* Items */}
            {o.items && o.items.length > 0 && (
              <div className="pt-2 border-t border-gray-100">
                <p className="text-xs text-gray-500 mb-1">الطلبات ({o.items.length})</p>
                {o.items.map((item, i) => (
                  <div key={i} className="flex justify-between text-sm py-0.5">
                    <span className="text-gray-600">{item.name_ar || item.name} ×{item.quantity}</span>
                    <span className="text-gray-500">{(item.price / 100).toFixed(2)}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Action button */}
        <div className="pt-3">
          {o.status === 'assigned' && (
            <button
              onClick={handlePickedUp}
              disabled={busy === 'picked'}
              className="w-full h-12 rounded-xl font-semibold text-white flex items-center justify-center gap-2 transition-all active:scale-[0.98] disabled:opacity-50"
              style={{ backgroundColor: '#FF6B35' }}
            >
              {busy === 'picked' ? (
                <Loader2 className="w-5 h-5 animate-spin" />
              ) : (
                <>
                  <CheckCircle2 className="w-5 h-5" />
                  تأكيد الاستلام من المطعم
                </>
              )}
            </button>
          )}

          {(o.status === 'picked_up' || o.status === 'on_the_way' || o.status === 'delivering') && (
            <button
              onClick={handleDelivered}
              disabled={busy === 'delivered'}
              className="w-full h-12 rounded-xl font-semibold text-white flex items-center justify-center gap-2 transition-all active:scale-[0.98] disabled:opacity-50"
              style={{ backgroundColor: '#10B981' }}
            >
              {busy === 'delivered' ? (
                <Loader2 className="w-5 h-5 animate-spin" />
              ) : (
                <>
                  <CheckCircle2 className="w-5 h-5" />
                  تأكيد التوصيل للعميل
                </>
              )}
            </button>
          )}
        </div>
      </div>
    </motion.div>
  )
}
