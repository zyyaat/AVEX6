
import { useState, useEffect, useRef } from 'react'
import { useRouter } from '@/lib/navigation'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Bike, Power, MapPin, Clock, TrendingUp, Star, Package,
  Zap, AlertCircle, Loader2, Navigation, Store, User,
} from 'lucide-react'
import { useAuth } from '@/store/auth'
import { useDriver } from '@/store/driver'
import { BottomTabBar } from '@/components/BottomTabBar'
import { TierBadge } from '@/components/TierBadge'
import { OfferModal } from '@/components/OfferModal'
import { ActiveDelivery } from '@/components/ActiveDelivery'
import { toast } from 'sonner'

export default function DriverHome() {
  const router = useRouter()
  const { isAuthenticated, mustChangePassword, logout } = useAuth()
  const {
    driver, offers, activeOrder,
    fetchMe, setOnline, updateLocation, refreshOffers, refreshActiveOrder, clear,
  } = useDriver()
  const [bootChecked, setBootChecked] = useState(false)
  const [togglingOnline, setTogglingOnline] = useState(false)
  const [activeOffer, setActiveOffer] = useState<string | null>(null)
  const locationIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const offersIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const activeIntervalRef = useRef<NodeJS.Timeout | null>(null)

  // Boot: check auth + load driver
  useEffect(() => {
    if (!isAuthenticated) {
      router.replace('/login')
      return
    }
    if (mustChangePassword) {
      router.replace('/change-password')
      return
    }
    setBootChecked(true)
    fetchMe()
  }, [isAuthenticated, mustChangePassword, router, fetchMe])

  // Watch GPS + send to backend when online
  useEffect(() => {
    if (!driver?.isOnline) {
      if (locationIntervalRef.current) {
        clearInterval(locationIntervalRef.current)
        locationIntervalRef.current = null
      }
      return
    }
    if (!navigator.geolocation) {
      toast.error('المتصفح لا يدعم تحديد الموقع')
      return
    }
    const watchId = navigator.geolocation.watchPosition(
      (pos) => updateLocation(pos.coords.latitude, pos.coords.longitude),
      () => {},
      { enableHighAccuracy: true, maximumAge: 5000, timeout: 10000 }
    )
    locationIntervalRef.current = setInterval(() => {
      navigator.geolocation.getCurrentPosition(
        (pos) => updateLocation(pos.coords.latitude, pos.coords.longitude),
        () => {},
        { enableHighAccuracy: true, maximumAge: 5000, timeout: 5000 }
      )
    }, 5000)
    return () => {
      navigator.geolocation.clearWatch(watchId)
      if (locationIntervalRef.current) clearInterval(locationIntervalRef.current)
    }
  }, [driver?.isOnline, updateLocation])

  // Poll offers when online and no active order
  useEffect(() => {
    if (!driver?.isOnline || activeOrder) {
      if (offersIntervalRef.current) {
        clearInterval(offersIntervalRef.current)
        offersIntervalRef.current = null
      }
      return
    }
    refreshOffers()
    offersIntervalRef.current = setInterval(refreshOffers, 3000)
    return () => {
      if (offersIntervalRef.current) clearInterval(offersIntervalRef.current)
    }
  }, [driver?.isOnline, activeOrder, refreshOffers])

  // Poll active order
  useEffect(() => {
    if (!activeOrder) {
      if (activeIntervalRef.current) {
        clearInterval(activeIntervalRef.current)
        activeIntervalRef.current = null
      }
      return
    }
    activeIntervalRef.current = setInterval(refreshActiveOrder, 5000)
    return () => {
      if (activeIntervalRef.current) clearInterval(activeIntervalRef.current)
    }
  }, [activeOrder, refreshActiveOrder])

  const handleToggleOnline = async () => {
    setTogglingOnline(true)
    try {
      const next = !driver?.isOnline
      await setOnline(next)
      toast.success(next ? 'أنت الآن متصل — استقبال الطلبات مفعّل' : 'تم إيقاف الاستقبال')
    } catch (err: any) {
      toast.error(err.message || 'تعذّر التبديل')
    } finally {
      setTogglingOnline(false)
    }
  }

  const handleLogout = () => {
    clear()
    logout()
    router.replace('/login')
  }

  // Pick the most recent offer to show in modal
  const currentOffer = activeOffer ? offers.find((o) => o.offerId === activeOffer) : offers[0]
  useEffect(() => {
    if (offers.length > 0 && !activeOffer) {
      setActiveOffer(offers[0].offerId)
    }
    if (offers.length === 0) setActiveOffer(null)
  }, [offers, activeOffer])

  if (!bootChecked) {
    return (
      <div className="min-h-dvh bg-white flex items-center justify-center">
        <Loader2 className="w-6 h-6 animate-spin" />
      </div>
    )
  }

  return (
    <div className="min-h-dvh bg-gray-50" dir="rtl">
      {/* Header */}
      <header className="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 h-14 flex items-center justify-between">
        <div className="flex items-center gap-2 min-w-0">
          <Bike className="w-5 h-5 text-black flex-shrink-0" />
          <span className="font-bold text-lg hidden sm:inline">AVEX Driver</span>
          {driver?.tier && (
            <TierBadge
              nameAr={driver.tier.nameAr}
              color={driver.tier.color}
              sortOrder={driver.tier.sortOrder}
              size="sm"
            />
          )}
        </div>
        <button
          onClick={handleToggleOnline}
          disabled={togglingOnline}
          className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-fluent disabled:opacity-50 ${
            driver?.isOnline ? 'bg-black text-white' : 'bg-gray-100 text-gray-500'
          }`}
        >
          {togglingOnline ? (
            <Loader2 className="w-3.5 h-3.5 animate-spin" />
          ) : (
            <div className={`w-2 h-2 rounded-full ${driver?.isOnline ? 'bg-white animate-pulse' : 'bg-gray-400'}`} />
          )}
          <span className="hidden sm:inline">{driver?.isOnline ? 'متصل' : 'غير متصل'}</span>
          <Power className="w-3.5 h-3.5" />
        </button>
      </header>

      <div className="container mx-auto px-4 py-4 max-w-md pb-20 sm:pb-4">
        {/* Online status banner (when online) */}
        <AnimatePresence>
          {driver?.isOnline && !activeOrder && (
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              className="bg-black text-white rounded-xl p-3 mb-4 flex items-center gap-2"
            >
              <div className="w-2 h-2 rounded-full bg-white animate-pulse" />
              <span className="text-sm font-medium">أنت متصل الآن — في انتظار الطلبات</span>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Active delivery (top priority) */}
        {activeOrder && <ActiveDelivery />}

        {/* Stats bar */}
        {!activeOrder && (
          <div className="grid grid-cols-4 gap-2 mb-4">
            <StatTile
              icon={<Package className="w-4 h-4" />}
              value={driver?.stats.completedOrders ?? 0}
              label="مكتمل"
            />
            <StatTile
              icon={<TrendingUp className="w-4 h-4" />}
              value={`${driver?.stats.acceptanceRate.toFixed(0) ?? 0}%`}
              label="قبول"
            />
            <StatTile
              icon={<Star className="w-4 h-4" />}
              value={driver?.stats.rating.toFixed(1) ?? '0.0'}
              label="تقييم"
            />
            <StatTile
              icon={<Zap className="w-4 h-4" />}
              value={driver?.stats.totalEarnings.toFixed(0) ?? 0}
              label="ج.م"
            />
          </div>
        )}

        {/* Available offers (only when online + no active order) */}
        {!activeOrder && driver?.isOnline && (
          <div>
            <h3 className="text-sm font-bold text-black mb-3 flex items-center gap-2">
              <Zap className="w-4 h-4" />
              طلبات متاحة ({offers.length})
            </h3>
            {offers.length === 0 ? (
              <EmptyOffers />
            ) : (
              <div className="space-y-3">
                {offers.map((offer, idx) => (
                  <motion.div
                    key={offer.offerId}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: idx * 0.05 }}
                  >
                    <OfferCard offer={offer} onClick={() => setActiveOffer(offer.offerId)} />
                  </motion.div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Offline state */}
        {!driver?.isOnline && !activeOrder && (
          <OfflineState
            onToggle={handleToggleOnline}
            toggling={togglingOnline}
          />
        )}

        {/* Tier progress card */}
        {!activeOrder && driver?.nextTier && (
          <div className="mt-6 bg-white rounded-xl border border-gray-200 p-4 shadow-fluent">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-gray-500" />
                <span className="text-sm font-bold">تقدّم المستوى</span>
              </div>
              {driver.tier && (
                <TierBadge
                  nameAr={driver.tier.nameAr}
                  color={driver.tier.color}
                  sortOrder={driver.tier.sortOrder}
                  size="sm"
                />
              )}
            </div>
            <p className="text-xs text-gray-600 mb-3">
              المستوى التالي: <b>{driver.nextTier.nameAr}</b>
            </p>
            <div className="space-y-2.5">
              <ProgressRow
                label="الطلبات المكتملة"
                current={driver.stats.completedOrders}
                target={driver.nextTier.minLifetimeOrders}
              />
              <ProgressRow
                label="نسبة القبول"
                current={driver.stats.acceptanceRate}
                target={driver.nextTier.minAcceptanceRate}
                suffix="%"
              />
              <ProgressRow
                label="التقييم"
                current={driver.stats.rating}
                target={driver.nextTier.minCustomerRating}
                step={0.1}
              />
              <ProgressRow
                label="الالتزام بالوقت"
                current={driver.stats.onTimeRate}
                target={driver.nextTier.minOnTimeRate}
                suffix="%"
              />
              <ProgressRow
                label="الحضور للشيفت"
                current={driver.stats.shiftAdherence}
                target={driver.nextTier.minShiftAdherence}
                suffix="%"
              />
            </div>
          </div>
        )}
      </div>

      {/* Offer Modal */}
      {currentOffer && (
        <OfferModal offer={currentOffer} onClose={() => setActiveOffer(null)} />
      )}

      <BottomTabBar />
    </div>
  )
}

// ===== Sub-components =====

function StatTile({ icon, value, label }: { icon: React.ReactNode; value: React.ReactNode; label: string }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-2.5 text-center shadow-fluent">
      <div className="flex items-center justify-center text-gray-500 mb-1">{icon}</div>
      <p className="text-base font-bold text-black">{value}</p>
      <p className="text-[10px] text-gray-400">{label}</p>
    </div>
  )
}

function ProgressRow({
  label, current, target, suffix = '', step = 1,
}: {
  label: string
  current: number
  target: number
  suffix?: string
  step?: number
}) {
  const pct = target > 0 ? Math.min(100, (current / target) * 100) : 100
  const reached = current >= target
  return (
    <div>
      <div className="flex items-center justify-between text-xs mb-1">
        <span className="text-gray-600">{label}</span>
        <span className={`font-bold ${reached ? 'text-black' : 'text-gray-500'}`}>
          {current.toFixed(step < 1 ? 1 : 0)}{suffix} / {target}{suffix}
        </span>
      </div>
      <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
        <div
          className={`h-full transition-all duration-500 ${reached ? 'bg-black' : 'bg-gray-400'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}

function EmptyOffers() {
  return (
    <div className="text-center py-16">
      <motion.div
        animate={{ scale: [1, 1.05, 1] }}
        transition={{ duration: 2, repeat: Infinity }}
        className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3"
      >
        <Package className="w-7 h-7 text-gray-300" />
      </motion.div>
      <p className="text-sm text-gray-500 font-medium">في انتظار الطلبات...</p>
      <p className="text-xs text-gray-400 mt-1">سيظهر أي طلب جديد هنا فور قبول المطعم</p>
    </div>
  )
}

function OfferCard({ offer, onClick }: { offer: any; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="w-full text-right bg-white rounded-xl border border-gray-200 p-4 hover:border-gray-400 transition-fluent shadow-fluent"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2 min-w-0">
          <div className="w-9 h-9 rounded-lg bg-gray-100 flex items-center justify-center flex-shrink-0">
            <Store className="w-4 h-4 text-black" />
          </div>
          <div className="min-w-0">
            <p className="font-bold text-sm truncate">{offer.restaurantName}</p>
            <p className="text-[10px] text-gray-400" dir="ltr">{offer.orderNumber}</p>
          </div>
        </div>
        <div className="text-left flex-shrink-0">
          <p className="font-bold text-sm">{offer.driverFee.toFixed(2)} ج.م</p>
          <p className="text-[10px] text-gray-400">أرباحك</p>
        </div>
      </div>

      <div className="space-y-1.5 text-xs text-gray-600 mb-3">
        <div className="flex items-center gap-1.5">
          <MapPin className="w-3.5 h-3.5 text-gray-400" />
          <span>{offer.zoneName}</span>
          <span className="text-gray-300">•</span>
          <span>{Math.round(offer.distanceM)} م منك</span>
        </div>
        <div className="flex items-start gap-1.5">
          <Clock className="w-3.5 h-3.5 text-gray-400 mt-0.5" />
          <span className="flex-1 line-clamp-2">{offer.itemsSummary}</span>
        </div>
      </div>

      <div className="bg-black text-white text-center py-2.5 rounded-lg text-xs font-bold">
        اضغط لعرض التفاصيل والقبول
      </div>
    </button>
  )
}

function OfflineState({ onToggle, toggling }: { onToggle: () => void; toggling: boolean }) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="text-center py-16"
    >
      <motion.div
        animate={{ scale: [1, 1.05, 1] }}
        transition={{ duration: 2.5, repeat: Infinity }}
        className="w-24 h-24 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-5"
      >
        <Power className="w-10 h-10 text-gray-300" />
      </motion.div>
      <h3 className="font-bold text-black mb-1 text-lg">أنت غير متصل</h3>
      <p className="text-sm text-gray-500 mb-6">اضغط للاتصال وبدء استقبال الطلبات</p>
      <button
        onClick={onToggle}
        disabled={toggling}
        className="px-8 h-12 rounded-xl bg-black text-white text-sm font-bold hover:bg-gray-800 active:bg-gray-900 transition-fluent inline-flex items-center gap-2 disabled:opacity-50"
      >
        {toggling ? <Loader2 className="w-5 h-5 animate-spin" /> : <Power className="w-5 h-5" />}
        ابدأ العمل
      </button>
    </motion.div>
  )
}
