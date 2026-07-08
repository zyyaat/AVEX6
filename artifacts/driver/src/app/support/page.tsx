
import { useState, useEffect } from 'react'
import { useRouter } from '@/lib/navigation'
import { ArrowLeft, LifeBuoy, Plus, MessageSquare, Loader2, X } from 'lucide-react'
import { useAuth } from '@/store/auth'
import { driverAPI } from '@/lib/api'
import { BottomTabBar } from '@/components/BottomTabBar'
import { toast } from 'sonner'

export default function SupportPage() {
  const router = useRouter()
  const { isAuthenticated } = useAuth()
  const [tickets, setTickets] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [type, setType] = useState('cancellation_request')
  const [orderId, setOrderId] = useState('')
  const [reason, setReason] = useState('')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    if (!isAuthenticated) { router.replace('/login'); return }
  }, [isAuthenticated, router])

  const load = () => {
    setLoading(true)
    driverAPI.getTickets().then((t) => {
      setTickets(t.tickets || [])
    }).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!reason) { toast.error('السبب مطلوب'); return }
    setCreating(true)
    try {
      await driverAPI.createTicket({ orderId: orderId || undefined, type, reason })
      toast.success('تم فتح التذكرة')
      setShowCreate(false)
      setReason('')
      setOrderId('')
      load()
    } catch (err: any) {
      toast.error(err.message || 'فشل الإنشاء')
    } finally {
      setCreating(false)
    }
  }

  const typeLabels: Record<string, string> = {
    cancellation_request: 'طلب إلغاء طلب',
    complaint: 'شكوى',
    other: 'أخرى',
  }

  const statusLabels: Record<string, string> = {
    open: 'مفتوحة',
    resolved: 'تم الحل',
    rejected: 'مرفوضة',
  }

  return (
    <div className="min-h-dvh bg-gray-50" dir="rtl">
      <header className="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 h-14 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button onClick={() => router.back()} className="w-9 h-9 rounded-full hover:bg-gray-100 flex items-center justify-center">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <h1 className="font-bold text-lg">الدعم</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="w-9 h-9 rounded-full bg-black text-white flex items-center justify-center"
        >
          <Plus className="w-5 h-5" />
        </button>
      </header>

      <div className="container mx-auto px-4 py-4 max-w-md pb-20 sm:pb-4">
        {/* Info banner */}
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-3 mb-4 text-xs text-gray-600">
          <p className="font-bold mb-1">ملاحظة هامة:</p>
          <p>لا يمكن إلغاء طلب بعد قبوله مباشرة. لإلغاء طلب في حالة خاصة جداً (مع عذر مقنع)، افتح تذكرة "طلب إلغاء طلب" وسيقوم الدعم بمراجعتها.</p>
        </div>

        {/* Tickets list */}
        {loading ? (
          <div className="text-center py-8"><Loader2 className="w-5 h-5 animate-spin mx-auto" /></div>
        ) : tickets.length === 0 ? (
          <div className="text-center py-16">
            <div className="w-14 h-14 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
              <LifeBuoy className="w-6 h-6 text-gray-300" />
            </div>
            <p className="text-sm text-gray-500">لا توجد تذاكر دعم</p>
            <p className="text-xs text-gray-400 mt-1">اضغط + لفتح تذكرة جديدة</p>
          </div>
        ) : (
          <div className="space-y-2">
            {tickets.map((t) => (
              <button
                key={t.id}
                onClick={() => router.push(`/support/${t.id}`)}
                className="w-full text-right bg-white rounded-lg border border-gray-200 p-3 hover:bg-gray-50 transition-fluent"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-bold bg-gray-100 px-2 py-0.5 rounded-full">
                    {typeLabels[t.type] || t.type}
                  </span>
                  <span className={`text-[10px] px-2 py-0.5 rounded-full ${
                    t.status === 'open' ? 'bg-black text-white' : 'bg-gray-100 text-gray-500'
                  }`}>
                    {statusLabels[t.status] || t.status}
                  </span>
                </div>
                <p className="text-xs text-gray-700 line-clamp-2">{t.reason}</p>
                <p className="text-[10px] text-gray-400 mt-1">
                  {new Date(t.createdAt).toLocaleString('ar-EG', { dateStyle: 'medium', timeStyle: 'short' })}
                </p>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Create modal */}
      {showCreate && (
        <div
          className="fixed inset-0 z-50 bg-black/50 flex items-end sm:items-center justify-center p-0 sm:p-4"
          onClick={(e) => e.target === e.currentTarget && setShowCreate(false)}
        >
          <div className="bg-white w-full sm:max-w-md rounded-t-2xl sm:rounded-2xl p-5" dir="rtl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-bold text-base">تذكرة دعم جديدة</h3>
              <button onClick={() => setShowCreate(false)} className="w-8 h-8 rounded-full hover:bg-gray-100 flex items-center justify-center">
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleCreate} className="space-y-3">
              <div>
                <label className="text-xs text-gray-500 mb-1 block">النوع</label>
                <select
                  value={type}
                  onChange={(e) => setType(e.target.value)}
                  className="w-full h-11 px-3 rounded-lg border border-gray-200 bg-white focus:outline-none focus:border-black"
                >
                  <option value="cancellation_request">طلب إلغاء طلب (حالة خاصة)</option>
                  <option value="complaint">شكوى</option>
                  <option value="other">أخرى</option>
                </select>
              </div>
              {type === 'cancellation_request' && (
                <div>
                  <label className="text-xs text-gray-500 mb-1 block">رقم الطلب (اختياري)</label>
                  <input
                    value={orderId}
                    onChange={(e) => setOrderId(e.target.value)}
                    placeholder="AV..."
                    className="w-full h-11 px-3 rounded-lg border border-gray-200 focus:outline-none focus:border-black"
                  />
                </div>
              )}
              <div>
                <label className="text-xs text-gray-500 mb-1 block">السبب (يجب أن يكون مقنعاً)</label>
                <textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  rows={4}
                  placeholder="اشرح السبب بالتفصيل..."
                  className="w-full p-3 rounded-lg border border-gray-200 focus:outline-none focus:border-black resize-none"
                />
              </div>
              <button
                type="submit"
                disabled={creating}
                className="w-full h-11 rounded-lg bg-black text-white text-sm font-bold hover:bg-gray-800 transition-fluent disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {creating ? <Loader2 className="w-4 h-4 animate-spin" /> : 'إرسال التذكرة'}
              </button>
            </form>
          </div>
        </div>
      )}

      <BottomTabBar />
    </div>
  )
}
