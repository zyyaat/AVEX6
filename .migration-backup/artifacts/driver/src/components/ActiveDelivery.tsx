
import { motion } from 'framer-motion'
import {
  Store, Package, MapPin, User, Phone, Navigation, ArrowRight,
  CheckCircle2, Loader2, Clock, TrendingUp, AlertCircle
} from 'lucide-react'
import { useDriver } from '@/store/driver'
import { toast } from 'sonner'
import { useState } from 'react'

export function ActiveDelivery() {
  const { activeOrder, pickedUp, arrived, delivered, refreshActiveOrder } = useDriver()
  const [busy, setBusy] = useState<string | null>(null)

  if (!activeOrder) return null

  const o = activeOrder

  const statusLabels: Record<string, string> = {
    assigned: 'تم القبول — اذهب للمطعم',
    picked_up: 'تم الاستلام — اضغط عند الوصول للعميل',
    on_the_way: 'في الطريق للعميل',
    delivering: 'في الطريق للعميل',
    delivered: 'تم التوصيل',
  }

  const handlePickedUp = async () => {
    setBusy('picked')
    try {
      await pickedUp(o.id)
      toast.success('تم تأكيد الاستلام من المطعم')
    } catch (err: any) {
      toast.error(err.message || 'تعذّر الاستلام')
    } finally {
      setBusy(null)
    }
  }

  const handleArrived = async () => {
    setBusy('arrived')
    try {
      await arrived(o.id)
      toast.success('في الطريق للعميل')
    } catch (err: any) {
      toast.error(err.message || 'تعذّر التحديث')
    } finally {
      setBusy(null)
    }
  }

  const handleDelivered = async () => {
    setBusy('delivered')
    try {
      const earnings = await delivered(o.id)
      toast.success(`تم التسليم — أرباحك: ${earnings.toFixed(2)} ج.م`)
    } catch (err: any) {
      toast.error(err.message || 'تعذّر التسليم')
    } finally {
      setBusy(null)
    }
  }

  // Step indicator
  const steps = [
    { key: 'assigned', label: 'المطعم', icon: Store },
    { key: 'picked_up', label: 'الاستلام', icon: Package },
    { key: 'on_the_way', label: 'الطريق', icon: Navigation },
    { key: 'delivered', label: 'التسليم', icon: CheckCircle2 },
  ]
  const currentStepIdx = steps.findIndex(s => s.key === o.status)
  if (o.status === 'delivering') currentStepIdx === 2

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-white rounded-lg border-2 border-black p-4 mb-4 shadow-fluent"
      dir="rtl"
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div>
          <h3 className="font-bold text-sm">التوصيل الحالي</h3>
          <p className="text-xs text-gray-400" dir="ltr">{o.orderNumber}</p>
        </div>
        <span className="text-xs font-bold bg-black text-white px-2.5 py-1 rounded-full">
          {statusLabels[o.status] || o.status}
        </span>
      </div>

      {/* Steps */}
      <div className="flex items-center justify-between mb-4 px-1">
        {steps.map((s, i) => {
          const Icon = s.icon
          const done = i < currentStepIdx
          const current = i === currentStepIdx
          return (
            <div key={s.key} className="flex items-center flex-1 last:flex-none">
              <div className="flex flex-col items-center gap-1">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center transition-fluent ${
                    done ? 'bg-black text-white' : current ? 'bg-black text-white ring-4 ring-gray-200' : 'bg-gray-100 text-gray-400'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                </div>
                <span className={`text-[9px] ${current || done ? 'font-bold text-black' : 'text-gray-400'}`}>{s.label}</span>
              </div>
              {i < steps.length - 1 && (
                <div className={`flex-1 h-0.5 mx-1 ${done ? 'bg-black' : 'bg-gray-200'}`} />
              )}
            </div>
          )
        })}
      </div>

      {/* Restaurant */}
      <div className="bg-gray-50 rounded-lg p-3 mb-2 border border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Store className="w-4 h-4 text-gray-600" />
            <span className="font-bold text-sm">{o.restaurantName}</span>
          </div>
          <a
            href={`https://www.google.com/maps?q=${o.restaurantLat},${o.restaurantLng}`}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-black underline flex items-center gap-1"
          >
            <Navigation className="w-3 h-3" /> خريطة
          </a>
        </div>
      </div>

      {/* Items */}
      <div className="bg-gray-50 rounded-lg p-3 mb-2 border border-gray-200">
        <p className="text-xs text-gray-500 mb-1.5">الأصناف:</p>
        <div className="space-y-1">
          {o.items.map((it, i) => (
            <div key={i} className="flex items-center justify-between text-xs">
              <span>{it.quantity}× {it.name}</span>
              <span className="text-gray-500">{(it.price * it.quantity).toFixed(2)} ج.م</span>
            </div>
          ))}
        </div>
      </div>

      {/* Customer */}
      <div className="bg-gray-50 rounded-lg p-3 mb-3 border border-gray-200">
        <div className="flex items-center justify-between mb-1">
          <div className="flex items-center gap-2">
            <User className="w-4 h-4 text-gray-600" />
            <span className="font-bold text-sm">{o.customerName}</span>
          </div>
          <a href={`tel:${o.phone}`} className="w-8 h-8 rounded-full bg-black text-white flex items-center justify-center">
            <Phone className="w-3.5 h-3.5" />
          </a>
        </div>
        <div className="flex items-start gap-2 mb-1.5">
          <MapPin className="w-4 h-4 text-gray-400 mt-0.5" />
          <span className="text-xs text-gray-600 flex-1">{o.locationAddress}</span>
        </div>
        <a
          href={o.locationUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center justify-between gap-2 bg-white rounded-lg p-2 border border-gray-200 hover:border-gray-400 transition-fluent"
        >
          <span className="text-xs font-medium flex items-center gap-1.5">
            <Navigation className="w-3.5 h-3.5" /> موقع العميل على الخريطة
          </span>
          <ArrowRight className="w-3.5 h-3.5 text-gray-400" />
        </a>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-2 mb-3">
        <div className="text-center p-2 bg-gray-50 rounded-lg border border-gray-200">
          <p className="text-xs font-bold">{o.dispatchDistanceM} م</p>
          <p className="text-[9px] text-gray-400">للمطعم</p>
        </div>
        <div className="text-center p-2 bg-gray-50 rounded-lg border border-gray-200">
          <p className="text-xs font-bold">{o.deliveryDistanceM} م</p>
          <p className="text-[9px] text-gray-400">للعميل</p>
        </div>
        <div className="text-center p-2 bg-black text-white rounded-lg">
          <p className="text-xs font-bold">{o.driverFee.toFixed(2)}</p>
          <p className="text-[9px] text-gray-300">ج.م</p>
        </div>
      </div>

      {/* Action button */}
      {o.status === 'assigned' && (
        <button
          onClick={handlePickedUp}
          disabled={busy !== null}
          className="w-full h-12 rounded-lg bg-black hover:bg-gray-800 text-white text-sm font-bold flex items-center justify-center gap-2 transition-fluent disabled:opacity-50"
        >
          {busy === 'picked' ? <Loader2 className="w-5 h-5 animate-spin" /> : <><Package className="w-5 h-5" /> وصلت للمطعم — استلم الطلب</>}
        </button>
      )}
      {o.status === 'picked_up' && (
        <button
          onClick={handleArrived}
          disabled={busy !== null}
          className="w-full h-12 rounded-lg bg-black hover:bg-gray-800 text-white text-sm font-bold flex items-center justify-center gap-2 transition-fluent disabled:opacity-50"
        >
          {busy === 'arrived' ? <Loader2 className="w-5 h-5 animate-spin" /> : <><Navigation className="w-5 h-5" /> بدأت التوصيل للعميل</>}
        </button>
      )}
      {(o.status === 'on_the_way' || o.status === 'delivering') && (
        <button
          onClick={handleDelivered}
          disabled={busy !== null}
          className="w-full h-12 rounded-lg bg-black hover:bg-gray-800 text-white text-sm font-bold flex items-center justify-center gap-2 transition-fluent disabled:opacity-50"
        >
          {busy === 'delivered' ? <Loader2 className="w-5 h-5 animate-spin" /> : <><CheckCircle2 className="w-5 h-5" /> وصلت للعميل — تأكيد التسليم</>}
        </button>
      )}

      {/* Geofence hint */}
      <p className="text-[10px] text-gray-400 mt-2 text-center flex items-center justify-center gap-1">
        <AlertCircle className="w-3 h-3" />
        يجب أن تكون على بُعد 70م من المطعم للاستلام و50م من العميل للتسليم
      </p>
    </motion.div>
  )
}
