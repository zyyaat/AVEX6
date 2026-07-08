
import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  MapPin, User, Package, Navigation, TrendingUp, Loader2, Store, Clock, AlertCircle,
} from 'lucide-react'
import type { Offer } from '@/lib/api'
import { useDriver } from '@/store/driver'
import { toast } from 'sonner'

interface OfferModalProps {
  offer: Offer
  onClose: () => void
}

const TOTAL_SECONDS = 15

export function OfferModal({ offer, onClose }: OfferModalProps) {
  const { acceptOffer, rejectOffer } = useDriver()
  const [accepting, setAccepting] = useState(false)
  const [rejecting, setRejecting] = useState(false)
  const [secondsLeft, setSecondsLeft] = useState(TOTAL_SECONDS)

  useEffect(() => {
    const expiry = new Date(offer.expiresAt).getTime()
    const tick = () => {
      const left = Math.max(0, Math.ceil((expiry - Date.now()) / 1000))
      setSecondsLeft(left)
      if (left <= 0) {
        toast.warning('انتهت صلاحية العرض')
        onClose()
      }
    }
    tick()
    const id = setInterval(tick, 200)
    return () => clearInterval(id)
  }, [offer.expiresAt, onClose])

  const handleAccept = async () => {
    setAccepting(true)
    try {
      await acceptOffer(offer.offerId)
      toast.success('تم قبول الطلب — اذهب للمطعم')
      onClose()
    } catch (err: any) {
      toast.error(err.message || 'فشل القبول')
      onClose()
    } finally {
      setAccepting(false)
    }
  }

  const handleReject = async () => {
    setRejecting(true)
    try {
      await rejectOffer(offer.offerId)
      toast.info('تم رفض الطلب')
      onClose()
    } catch (err: any) {
      toast.error(err.message || 'فشل الرفض')
      onClose()
    } finally {
      setRejecting(false)
    }
  }

  const isUrgent = secondsLeft <= 5
  const progress = secondsLeft / TOTAL_SECONDS
  const circumference = 2 * Math.PI * 28
  const offset = circumference * (1 - progress)

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 bg-black/60 flex items-end sm:items-center justify-center p-0 sm:p-4"
        onClick={(e) => e.target === e.currentTarget && onClose()}
      >
        <motion.div
          initial={{ y: '100%', opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: '100%', opacity: 0 }}
          transition={{ type: 'spring', damping: 28, stiffness: 320 }}
          className="bg-white w-full sm:max-w-md rounded-t-2xl sm:rounded-2xl shadow-fluent-lg overflow-hidden"
          dir="rtl"
        >
          {/* Header with countdown */}
          <div className={`px-5 py-4 flex items-center justify-between transition-colors ${
            isUrgent ? 'bg-gray-900' : 'bg-black'
          }`}>
            <div>
              <p className="text-xs text-gray-300">طلب جديد</p>
              <p className="font-bold text-lg text-white" dir="ltr">{offer.orderNumber}</p>
            </div>
            <motion.div
              animate={isUrgent ? { scale: [1, 1.08, 1] } : {}}
              transition={{ duration: 0.6, repeat: Infinity }}
              className="relative w-16 h-16"
            >
              <svg className="w-16 h-16 -rotate-90" viewBox="0 0 64 64">
                <circle cx="32" cy="32" r="28" stroke="rgba(255,255,255,0.15)" strokeWidth="3" fill="none" />
                <circle
                  cx="32" cy="32" r="28"
                  stroke={isUrgent ? '#fff' : '#fff'}
                  strokeWidth="3" fill="none"
                  strokeLinecap="round"
                  strokeDasharray={circumference}
                  strokeDashoffset={offset}
                  className="transition-all duration-200"
                />
              </svg>
              <span className={`absolute inset-0 flex items-center justify-center text-xl font-bold ${
                isUrgent ? 'animate-pulse' : ''
              }`}>
                {secondsLeft}
              </span>
            </motion.div>
          </div>

          {/* Body */}
          <div className="p-5 space-y-4">
            {/* Restaurant + Zone */}
            <div className="bg-gray-50 rounded-xl p-3 border border-gray-200">
              <div className="flex items-center gap-2 mb-1">
                <div className="w-8 h-8 rounded-lg bg-white border border-gray-200 flex items-center justify-center">
                  <Store className="w-4 h-4 text-black" />
                </div>
                <div>
                  <p className="font-bold text-sm leading-tight">{offer.restaurantName}</p>
                  <p className="text-xs text-gray-500 flex items-center gap-1">
                    <MapPin className="w-3 h-3" />
                    {offer.zoneName}
                  </p>
                </div>
              </div>
            </div>

            {/* Items summary */}
            <div className="flex items-start gap-2">
              <div className="w-8 h-8 rounded-lg bg-gray-100 flex items-center justify-center flex-shrink-0">
                <Package className="w-4 h-4 text-gray-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[10px] text-gray-400 mb-0.5">الأصناف</p>
                <p className="text-sm text-gray-700 line-clamp-2">{offer.itemsSummary}</p>
              </div>
            </div>

            {/* Customer + addresses */}
            <div className="bg-gray-50 rounded-xl p-3 border border-gray-200 space-y-2">
              <div className="flex items-center gap-2 text-sm">
                <User className="w-4 h-4 text-gray-500 flex-shrink-0" />
                <span className="font-medium">{offer.customerName}</span>
              </div>
              <div className="flex items-start gap-2 text-xs">
                <MapPin className="w-4 h-4 text-gray-400 mt-0.5 flex-shrink-0" />
                <span className="text-gray-600 flex-1">{offer.locationAddress}</span>
              </div>
            </div>

            {/* Distances */}
            <div className="grid grid-cols-2 gap-2">
              <div className="bg-white rounded-lg p-2.5 border border-gray-200 text-center">
                <Navigation className="w-4 h-4 mx-auto mb-1 text-gray-500" />
                <p className="text-sm font-bold">{Math.round(offer.distanceM)} م</p>
                <p className="text-[10px] text-gray-400">منك للمطعم</p>
              </div>
              <div className="bg-white rounded-lg p-2.5 border border-gray-200 text-center">
                <Navigation className="w-4 h-4 mx-auto mb-1 text-gray-500" />
                <p className="text-sm font-bold">{Math.round(offer.estimatedDeliveryDistanceM)} م</p>
                <p className="text-[10px] text-gray-400">للعميل</p>
              </div>
            </div>

            {/* Earnings */}
            <div className="bg-black text-white rounded-xl p-4 text-center">
              <TrendingUp className="w-5 h-5 mx-auto mb-1" />
              <p className="text-2xl font-bold">{offer.driverFee.toFixed(2)} <span className="text-sm">ج.م</span></p>
              <p className="text-[10px] text-gray-300">أرباحك من هذا الطلب</p>
            </div>

            {/* Geofence hint */}
            <div className="flex items-start gap-1.5 text-[10px] text-gray-500 bg-gray-50 rounded-lg p-2">
              <AlertCircle className="w-3 h-3 mt-0.5 flex-shrink-0" />
              <p>يجب أن تكون على بُعد 70م من المطعم للاستلام و50م من العميل للتسليم</p>
            </div>
          </div>

          {/* Actions */}
          <div className="px-5 pb-5 pt-1 grid grid-cols-2 gap-3">
            <button
              onClick={handleReject}
              disabled={accepting || rejecting}
              className="h-12 rounded-xl border border-gray-300 bg-white text-gray-700 font-medium hover:bg-gray-50 active:bg-gray-100 transition-fluent disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
            >
              {rejecting ? <Loader2 className="w-4 h-4 animate-spin" /> : 'رفض'}
            </button>
            <button
              onClick={handleAccept}
              disabled={accepting || rejecting}
              className="h-12 rounded-xl bg-black text-white font-bold hover:bg-gray-800 active:bg-gray-900 transition-fluent disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {accepting ? <Loader2 className="w-4 h-4 animate-spin" /> : 'قبول الطلب'}
            </button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  )
}
