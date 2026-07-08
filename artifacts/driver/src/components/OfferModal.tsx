import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  MapPin, Package, Navigation, Loader2, Store, X,
} from 'lucide-react'
import type { DispatchOffer } from '@/lib/api'
import { useDriver } from '@/store/driver'
import { toast } from 'sonner'

interface OfferModalProps {
  offer: DispatchOffer
  onClose: () => void
}

export function OfferModal({ offer, onClose }: OfferModalProps) {
  const { acceptOffer, rejectOffer } = useDriver()
  const [accepting, setAccepting] = useState(false)
  const [rejecting, setRejecting] = useState(false)
  const [secondsLeft, setSecondsLeft] = useState(15)

  // Countdown from offer expiry
  useEffect(() => {
    const expiry = new Date(offer.expires_at).getTime()
    const tick = () => {
      const left = Math.max(0, Math.ceil((expiry - Date.now()) / 1000))
      setSecondsLeft(left)
      if (left <= 0) {
        toast.warning('انتهت صلاحية العرض')
        onClose()
      }
    }
    tick()
    const interval = setInterval(tick, 1000)
    return () => clearInterval(interval)
  }, [offer.expires_at, onClose])

  // Vibrate every 5 seconds
  useEffect(() => {
    if (secondsLeft > 0 && secondsLeft % 5 === 0 && navigator.vibrate) {
      navigator.vibrate(100)
    }
  }, [secondsLeft])

  const handleAccept = async () => {
    setAccepting(true)
    try {
      await acceptOffer(offer.id)
      toast.success('تم قبول العرض!')
      onClose()
    } catch (err: any) {
      toast.error(err.message || 'فشل قبول العرض')
      onClose()
    } finally {
      setAccepting(false)
    }
  }

  const handleReject = async () => {
    setRejecting(true)
    try {
      await rejectOffer(offer.id)
      onClose()
    } catch (err: any) {
      toast.error(err.message || 'فشل رفض العرض')
      onClose()
    } finally {
      setRejecting(false)
    }
  }

  const distanceKm = offer.est_distance_m ? (offer.est_distance_m / 1000).toFixed(1) : '—'
  const durationMin = offer.est_duration_s ? Math.ceil(offer.est_duration_s / 60) : '—'
  const fare = offer.est_fare_cents ? (offer.est_fare_cents / 100).toFixed(2) : '—'

  // Progress bar percentage
  const progress = (secondsLeft / 15) * 100

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50"
        onClick={onClose}
      >
        <motion.div
          initial={{ y: '100%', opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: '100%', opacity: 0 }}
          transition={{ type: 'spring', damping: 25, stiffness: 300 }}
          onClick={(e) => e.stopPropagation()}
          className="w-full sm:max-w-md bg-white rounded-t-3xl sm:rounded-3xl overflow-hidden"
          dir="rtl"
        >
          {/* Timer bar */}
          <div className="h-1.5 bg-gray-100 relative">
            <motion.div
              className="h-full"
              style={{
                backgroundColor: secondsLeft <= 5 ? '#EF4444' : '#FF6B35',
                width: `${progress}%`,
              }}
              animate={{ width: `${progress}%` }}
              transition={{ duration: 1, ease: 'linear' }}
            />
          </div>

          {/* Header */}
          <div className="px-5 pt-4 pb-3 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full flex items-center justify-center" style={{ backgroundColor: '#FF6B35' }}>
                <Package className="w-4 h-4 text-white" />
              </div>
              <span className="font-bold text-gray-900">عرض توصيل جديد</span>
            </div>
            <div className="flex items-center gap-1">
              <span className={`text-2xl font-bold ${secondsLeft <= 5 ? 'text-red-500' : 'text-gray-900'}`}>
                {secondsLeft}
              </span>
              <span className="text-xs text-gray-500">ثانية</span>
            </div>
          </div>

          {/* Offer details */}
          <div className="px-5 pb-4 space-y-3">
            {/* Distance + Duration + Fare */}
            <div className="grid grid-cols-3 gap-2">
              <div className="bg-gray-50 rounded-xl p-3 text-center">
                <Navigation className="w-5 h-5 mx-auto mb-1 text-gray-500" />
                <p className="text-lg font-bold text-gray-900">{distanceKm}</p>
                <p className="text-xs text-gray-500">كم</p>
              </div>
              <div className="bg-gray-50 rounded-xl p-3 text-center">
                <Store className="w-5 h-5 mx-auto mb-1 text-gray-500" />
                <p className="text-lg font-bold text-gray-900">{durationMin}</p>
                <p className="text-xs text-gray-500">دقيقة</p>
              </div>
              <div className="bg-gray-50 rounded-xl p-3 text-center">
                <MapPin className="w-5 h-5 mx-auto mb-1 text-gray-500" />
                <p className="text-lg font-bold text-gray-900">{fare}</p>
                <p className="text-xs text-gray-500">{offer.currency}</p>
              </div>
            </div>

            {/* Pickup location */}
            <div className="flex items-start gap-3 p-3 bg-gray-50 rounded-xl">
              <div className="w-8 h-8 rounded-full bg-orange-100 flex items-center justify-center shrink-0">
                <Store className="w-4 h-4 text-orange-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-xs text-gray-500">نقطة الاستلام</p>
                <p className="text-sm font-medium text-gray-900 truncate">
                  {offer.pickup_lat.toFixed(4)}, {offer.pickup_lng.toFixed(4)}
                </p>
              </div>
            </div>

            {/* Delivery location */}
            <div className="flex items-start gap-3 p-3 bg-gray-50 rounded-xl">
              <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center shrink-0">
                <MapPin className="w-4 h-4 text-blue-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-xs text-gray-500">نقطة التوصيل</p>
                <p className="text-sm font-medium text-gray-900 truncate">
                  {offer.delivery_lat.toFixed(4)}, {offer.delivery_lng.toFixed(4)}
                </p>
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3 p-4 pt-0">
            <button
              onClick={handleReject}
              disabled={rejecting || accepting}
              className="flex-1 h-12 rounded-xl font-semibold text-gray-700 bg-gray-100 hover:bg-gray-200 transition-colors disabled:opacity-50 active:scale-[0.98]"
            >
              {rejecting ? <Loader2 className="w-5 h-5 animate-spin mx-auto" /> : 'رفض'}
            </button>
            <button
              onClick={handleAccept}
              disabled={accepting || rejecting}
              className="flex-1 h-12 rounded-xl font-semibold text-white transition-colors disabled:opacity-50 active:scale-[0.98]"
              style={{ backgroundColor: '#FF6B35' }}
            >
              {accepting ? <Loader2 className="w-5 h-5 animate-spin mx-auto" /> : 'قبول'}
            </button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  )
}
