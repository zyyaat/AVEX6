
import { useState, useEffect } from 'react'
import { useRouter } from '@/lib/navigation'
import { ArrowLeft, Package, Loader2, ChevronLeft, ChevronRight } from 'lucide-react'
import { useAuth } from '@/store/auth'
import { driverAPI } from '@/lib/api'
import { BottomTabBar } from '@/components/BottomTabBar'

const statusLabels: Record<string, string> = {
  delivered: 'مكتمل',
  on_the_way: 'في الطريق',
  delivering: 'في الطريق',
  picked_up: 'تم الاستلام',
  assigned: 'مُسند',
  cancelled: 'ملغي',
  new: 'جديد',
  accepted: 'مقبول',
  preparing: 'قيد التحضير',
  ready: 'جاهز',
}

export default function HistoryPage() {
  const router = useRouter()
  const { isAuthenticated } = useAuth()
  const [orders, setOrders] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)

  useEffect(() => {
    if (!isAuthenticated) { router.replace('/login'); return }
  }, [isAuthenticated, router])

  useEffect(() => {
    setLoading(true)
    driverAPI.getHistory(page).then((h) => {
      setOrders(h.orders || [])
      setHasMore((h.orders || []).length === 20)
    }).finally(() => setLoading(false))
  }, [page])

  return (
    <div className="min-h-dvh bg-gray-50" dir="rtl">
      <header className="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 h-14 flex items-center gap-3">
        <button onClick={() => router.back()} className="w-9 h-9 rounded-full hover:bg-gray-100 flex items-center justify-center transition-fluent" aria-label="رجوع">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <h1 className="font-bold text-lg">سجلّ الطلبات</h1>
      </header>

      <div className="container mx-auto px-4 py-4 max-w-md pb-20 sm:pb-4">
        {loading ? (
          <div className="text-center py-20"><Loader2 className="w-6 h-6 animate-spin mx-auto" /></div>
        ) : orders.length === 0 ? (
          <div className="text-center py-20">
            <div className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <Package className="w-7 h-7 text-gray-300" />
            </div>
            <p className="text-sm text-gray-500 font-medium">لا توجد طلبات في سجلّك بعد</p>
            <p className="text-xs text-gray-400 mt-1">ستظهر هنا بعد استلام وتسليم أول طلب</p>
          </div>
        ) : (
          <>
            <div className="space-y-2">
              {orders.map((o) => (
                <div key={o.id} className="bg-white rounded-xl border border-gray-200 p-3 flex items-center justify-between shadow-fluent">
                  <div className="min-w-0 flex-1">
                    <p className="font-bold text-sm truncate">{o.restaurantName || 'مطعم'}</p>
                    <p className="text-[10px] text-gray-400" dir="ltr">{o.orderNumber}</p>
                    <p className="text-[10px] text-gray-500 mt-0.5">
                      {new Date(o.createdAt).toLocaleString('ar-EG', { dateStyle: 'medium', timeStyle: 'short' })}
                    </p>
                  </div>
                  <div className="text-left flex-shrink-0 ml-2">
                    <p className="font-bold text-sm text-black">{o.earnings.toFixed(2)} ج.م</p>
                    <span className={`inline-block text-[10px] px-2 py-0.5 rounded-full mt-0.5 ${
                      o.status === 'delivered' ? 'bg-black text-white' :
                      o.status === 'cancelled' ? 'bg-gray-200 text-gray-500' : 'bg-gray-100 text-gray-600'
                    }`}>
                      {statusLabels[o.status] || o.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>

            {/* Pagination */}
            {(page > 1 || hasMore) && (
              <div className="flex items-center justify-center gap-3 mt-6">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="w-10 h-10 rounded-full border border-gray-200 bg-white flex items-center justify-center disabled:opacity-30 hover:bg-gray-50 transition-fluent"
                  aria-label="السابق"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
                <span className="text-sm text-gray-500 font-medium">صفحة {page}</span>
                <button
                  onClick={() => setPage((p) => p + 1)}
                  disabled={!hasMore}
                  className="w-10 h-10 rounded-full border border-gray-200 bg-white flex items-center justify-center disabled:opacity-30 hover:bg-gray-50 transition-fluent"
                  aria-label="التالي"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
              </div>
            )}
          </>
        )}
      </div>

      <BottomTabBar />
    </div>
  )
}
