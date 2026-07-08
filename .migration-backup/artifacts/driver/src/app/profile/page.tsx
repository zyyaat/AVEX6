
import { useState, useEffect } from 'react'
import { useRouter } from '@/lib/navigation'
import {
  ArrowLeft, User, Phone, Lock, Zap, LogOut, ChevronLeft, Loader2, Award, Star, Power,
} from 'lucide-react'
import { useAuth } from '@/store/auth'
import { useDriver } from '@/store/driver'
import { BottomTabBar } from '@/components/BottomTabBar'
import { TierBadge } from '@/components/TierBadge'
import { toast } from 'sonner'

export default function ProfilePage() {
  const router = useRouter()
  const { isAuthenticated, driverName, driverPhone, logout } = useAuth()
  const { driver, fetchMe, setAutoAccept, clear } = useDriver()
  const [bootChecked, setBootChecked] = useState(false)

  useEffect(() => {
    if (!isAuthenticated) { router.replace('/login'); return }
    setBootChecked(true)
    fetchMe()
  }, [isAuthenticated, router, fetchMe])

  const handleLogout = () => {
    clear()
    logout()
    router.replace('/login')
  }

  if (!bootChecked) {
    return (
      <div className="min-h-dvh bg-gray-50 flex items-center justify-center">
        <Loader2 className="w-6 h-6 animate-spin" />
      </div>
    )
  }

  return (
    <div className="min-h-dvh bg-gray-50" dir="rtl">
      <header className="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 h-14 flex items-center gap-3">
        <button onClick={() => router.back()} className="w-9 h-9 rounded-full hover:bg-gray-100 flex items-center justify-center transition-fluent" aria-label="رجوع">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <h1 className="font-bold text-lg">حسابي</h1>
      </header>

      <div className="container mx-auto px-4 py-4 max-w-md pb-20 sm:pb-4">
        {/* Profile header */}
        <div className="bg-white rounded-xl border border-gray-200 p-4 mb-4 flex items-center gap-3 shadow-fluent">
          <div className="w-14 h-14 rounded-full bg-black text-white flex items-center justify-center text-xl font-bold">
            {driverName?.charAt(0) || 'م'}
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-bold text-base truncate">{driverName}</p>
            <p className="text-xs text-gray-500" dir="ltr">{driverPhone}</p>
          </div>
          {driver?.tier && (
            <TierBadge
              nameAr={driver.tier.nameAr}
              color={driver.tier.color}
              sortOrder={driver.tier.sortOrder}
              size="md"
            />
          )}
        </div>

        {/* Stats summary */}
        {driver && (
          <div className="grid grid-cols-3 gap-2 mb-4">
            <StatCard icon={<Star className="w-4 h-4" />} value={driver.stats.rating.toFixed(1)} label="التقييم" />
            <StatCard icon={<Award className="w-4 h-4" />} value={`${driver.stats.acceptanceRate.toFixed(0)}%`} label="القبول" />
            <StatCard icon={<Zap className="w-4 h-4" />} value={driver.stats.completedOrders} label="مكتمل" />
          </div>
        )}

        {/* Auto-accept toggle */}
        <button
          onClick={() => setAutoAccept(!driver?.autoAccept)}
          className="w-full bg-white rounded-xl border border-gray-200 p-4 mb-3 flex items-center gap-3 hover:bg-gray-50 transition-fluent text-right shadow-fluent"
        >
          <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
            <Zap className="w-5 h-5 text-black" />
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-bold text-sm">القبول التلقائي</p>
            <p className="text-xs text-gray-500">قبول الطلبات فور وصولها بدون انتظار</p>
          </div>
          <div className={`w-12 h-7 rounded-full p-1 transition-fluent flex-shrink-0 ${driver?.autoAccept ? 'bg-black' : 'bg-gray-200'}`}>
            <div className={`w-5 h-5 rounded-full bg-white shadow-fluent transition-fluent ${driver?.autoAccept ? 'translate-x-0' : '-translate-x-5'}`} />
          </div>
        </button>

        {/* Change password — link to dedicated page */}
        <button
          onClick={() => router.push('/change-password')}
          className="w-full bg-white rounded-xl border border-gray-200 p-4 mb-3 flex items-center gap-3 hover:bg-gray-50 transition-fluent text-right shadow-fluent"
        >
          <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
            <Lock className="w-5 h-5 text-black" />
          </div>
          <div className="flex-1">
            <p className="font-bold text-sm">تغيير كلمة المرور</p>
            <p className="text-xs text-gray-500">حدّث كلمة مرور حسابك</p>
          </div>
          <ChevronLeft className="w-4 h-4 text-gray-400" />
        </button>

        {/* Tier progress card */}
        {driver?.nextTier && (
          <div className="bg-white rounded-xl border border-gray-200 p-4 mb-3 shadow-fluent">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <Award className="w-4 h-4 text-gray-500" />
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

        {/* Logout */}
        <button
          onClick={handleLogout}
          className="w-full bg-white rounded-xl border border-gray-200 p-4 flex items-center gap-3 hover:bg-gray-50 transition-fluent text-right shadow-fluent"
        >
          <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
            <Power className="w-5 h-5 text-black" />
          </div>
          <p className="font-bold text-sm flex-1">تسجيل الخروج</p>
        </button>
      </div>

      <BottomTabBar />
    </div>
  )
}

function StatCard({ icon, value, label }: { icon: React.ReactNode; value: React.ReactNode; label: string }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-3 text-center shadow-fluent">
      <div className="flex items-center justify-center text-gray-500 mb-1">{icon}</div>
      <p className="text-sm font-bold">{value}</p>
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
