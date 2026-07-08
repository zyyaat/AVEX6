
import { useState, useEffect } from 'react'
import { useRouter } from '@/lib/navigation'
import { ArrowLeft, TrendingUp, Loader2, Package, Wallet, Star, Zap } from 'lucide-react'
import { useAuth } from '@/store/auth'
import { useDriver } from '@/store/driver'
import { driverAPI } from '@/lib/api'
import { BottomTabBar } from '@/components/BottomTabBar'

type Period = 'today' | 'week' | 'month'

export default function EarningsPage() {
  const router = useRouter()
  const { isAuthenticated } = useAuth()
  const { driver, fetchMe } = useDriver()
  const [period, setPeriod] = useState<Period>('today')
  const [data, setData] = useState<{ totalEarnings: number; completedOrders: number } | null>(null)
  const [loading, setLoading] = useState(true)
  const [history, setHistory] = useState<any[]>([])

  useEffect(() => {
    if (!isAuthenticated) { router.replace('/login'); return }
    fetchMe()
  }, [isAuthenticated, router, fetchMe])

  useEffect(() => {
    setLoading(true)
    Promise.all([
      driverAPI.getEarnings(period),
      driverAPI.getHistory(1),
    ]).then(([e, h]) => {
      setData({ totalEarnings: e.totalEarnings, completedOrders: e.completedOrders })
      setHistory(h.orders || [])
    }).finally(() => setLoading(false))
  }, [period])

  const periodLabels: Record<Period, string> = { today: 'اليوم', week: 'الأسبوع', month: 'الشهر' }

  // Build chart data from history (last 7 orders as bars)
  const chartData = history.slice(0, 7).reverse().map((o) => ({
    label: new Date(o.createdAt).toLocaleTimeString('ar-EG', { hour: '2-digit', minute: '2-digit' }),
    value: o.earnings,
  }))
  const maxChart = Math.max(...chartData.map((d) => d.value), 1)

  return (
    <div className="min-h-dvh bg-gray-50" dir="rtl">
      <header className="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 h-14 flex items-center gap-3">
        <button onClick={() => router.back()} className="w-9 h-9 rounded-full hover:bg-gray-100 flex items-center justify-center transition-fluent" aria-label="رجوع">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <h1 className="font-bold text-lg">أرباحي</h1>
      </header>

      <div className="container mx-auto px-4 py-4 max-w-md pb-20 sm:pb-4">
        {/* Period tabs */}
        <div className="grid grid-cols-3 gap-1 bg-gray-100 rounded-xl p-1 mb-4">
          {(['today', 'week', 'month'] as Period[]).map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`py-2.5 rounded-lg text-xs font-bold transition-fluent ${
                period === p ? 'bg-white text-black shadow-fluent' : 'text-gray-500'
              }`}
            >
              {periodLabels[p]}
            </button>
          ))}
        </div>

        {/* Total earnings card */}
        <div className="bg-black text-white rounded-2xl p-5 mb-4 shadow-fluent-lg">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Wallet className="w-5 h-5" />
              <span className="text-sm text-gray-300">إجمالي الأرباح — {periodLabels[period]}</span>
            </div>
            <TrendingUp className="w-4 h-4 text-gray-400" />
          </div>
          {loading ? (
            <Loader2 className="w-8 h-8 animate-spin" />
          ) : (
            <>
              <p className="text-4xl font-bold mb-1">
                {data?.totalEarnings.toFixed(2)}
                <span className="text-lg mr-1">ج.م</span>
              </p>
              <p className="text-sm text-gray-300">{data?.completedOrders ?? 0} طلب مكتمل</p>
            </>
          )}
        </div>

        {/* Lifetime stats */}
        <div className="grid grid-cols-3 gap-2 mb-4">
          <LifetimeStat
            icon={<Wallet className="w-4 h-4" />}
            value={`${driver?.stats.totalEarnings.toFixed(0) ?? 0}`}
            label="إجمالي تراكمي"
            suffix="ج.م"
          />
          <LifetimeStat
            icon={<Package className="w-4 h-4" />}
            value={driver?.stats.completedOrders ?? 0}
            label="طلبات مكتملة"
          />
          <LifetimeStat
            icon={<Star className="w-4 h-4" />}
            value={driver?.stats.rating.toFixed(1) ?? '0.0'}
            label="تقييمك"
          />
        </div>

        {/* Chart */}
        {!loading && chartData.length > 0 && (
          <div className="bg-white rounded-xl border border-gray-200 p-4 mb-4 shadow-fluent">
            <h3 className="text-sm font-bold mb-3 flex items-center gap-2">
              <TrendingUp className="w-4 h-4 text-gray-500" />
              آخر الطلبات
            </h3>
            <div className="flex items-end justify-between gap-1.5 h-32">
              {chartData.map((d, i) => {
                const h = (d.value / maxChart) * 100
                return (
                  <div key={i} className="flex-1 flex flex-col items-center gap-1">
                    <div className="w-full bg-gray-100 rounded-t-md flex items-end" style={{ height: '100px' }}>
                      <div
                        className="w-full bg-black rounded-t-md transition-all duration-500"
                        style={{ height: `${h}%` }}
                      />
                    </div>
                    <span className="text-[8px] text-gray-400 truncate w-full text-center">{d.label}</span>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Recent orders list */}
        <h3 className="text-sm font-bold text-black mb-3 flex items-center gap-2">
          <Package className="w-4 h-4" /> تفاصيل الطلبات
        </h3>
        {loading ? (
          <div className="text-center py-8"><Loader2 className="w-5 h-5 animate-spin mx-auto" /></div>
        ) : history.length === 0 ? (
          <div className="text-center py-12">
            <div className="w-14 h-14 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
              <Package className="w-6 h-6 text-gray-300" />
            </div>
            <p className="text-sm text-gray-400">لا توجد طلبات بعد</p>
          </div>
        ) : (
          <div className="space-y-2">
            {history.map((o) => (
              <div key={o.id} className="bg-white rounded-xl border border-gray-200 p-3 flex items-center justify-between shadow-fluent">
                <div className="min-w-0">
                  <p className="font-bold text-sm truncate">{o.restaurantName || 'مطعم'}</p>
                  <p className="text-[10px] text-gray-400" dir="ltr">{o.orderNumber}</p>
                  <p className="text-[10px] text-gray-400">{new Date(o.createdAt).toLocaleString('ar-EG', { dateStyle: 'short', timeStyle: 'short' })}</p>
                </div>
                <div className="text-left flex-shrink-0">
                  <p className="font-bold text-sm text-black">{o.earnings.toFixed(2)} ج.م</p>
                  <p className={`text-[10px] ${o.status === 'delivered' ? 'text-black' : 'text-gray-400'}`}>
                    {o.status === 'delivered' ? '✓ مكتمل' : o.status}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <BottomTabBar />
    </div>
  )
}

function LifetimeStat({ icon, value, label, suffix }: { icon: React.ReactNode; value: React.ReactNode; label: string; suffix?: string }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-2.5 text-center shadow-fluent">
      <div className="flex items-center justify-center text-gray-500 mb-1">{icon}</div>
      <p className="text-sm font-bold">
        {value}{suffix && <span className="text-[10px] mr-0.5">{suffix}</span>}
      </p>
      <p className="text-[9px] text-gray-400">{label}</p>
    </div>
  )
}
