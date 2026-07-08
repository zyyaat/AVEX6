
import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { CheckCircle2, Copy, Package, ChefHat, Bike, Home, X, Clock, MapPin, ExternalLink, ArrowRight, Search, User, Phone } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import { ordersAPI } from '@/lib/api'

const STEPS = [
  { id: 'new', label: 'تم استلام الطلب', icon: Package, description: 'وصلنا طلبك بنجاح' },
  { id: 'accepted', label: 'تم القبول', icon: CheckCircle2, description: 'تم قبول الطلب' },
  { id: 'preparing', label: 'قيد التحضير', icon: ChefHat, description: 'الشيف يحضّر طلبك الآن' },
  { id: 'ready', label: 'جاهز', icon: Package, description: 'طلبك جاهز للاستلام' },
  { id: 'delivering', label: 'في الطريق إليك', icon: Bike, description: 'المندوب في طريقه إليك' },
  { id: 'delivered', label: 'تم التوصيل', icon: Home, description: 'وصل طلبك! بالهناء والشفاء' },
]
const STATUS_TO_STEP: Record<string, number> = { new: 0, accepted: 1, preparing: 2, ready: 3, picked_up: 4, delivering: 4, delivered: 5 }

interface OrderTrackingProps { initialOrderNumber?: string; onBack: () => void }

export function OrderTracking({ initialOrderNumber, onBack }: OrderTrackingProps) {
  const [searchOrderNumber, setSearchOrderNumber] = useState(initialOrderNumber || '')
  const [order, setOrder] = useState<any | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const fetchOrder = async (num: string) => {
    if (!num.trim()) { setError('يرجى إدخال رقم الطلب'); return }
    setLoading(true); setError('')
    try {
      const data = await ordersAPI.trackByNumber(num.trim())
      if (!data.order) { setError('لم يتم العثور على الطلب'); setOrder(null) }
      else setOrder(data.order)
    } catch { setError('فشل تحميل الطلب') } finally { setLoading(false) }
  }

  useEffect(() => { if (initialOrderNumber) fetchOrder(initialOrderNumber) }, [initialOrderNumber])
  useEffect(() => {
    if (!order) return
    const interval = setInterval(async () => { try { const data = await ordersAPI.trackByNumber(order.orderNumber); if (data.order) setOrder(data.order) } catch {} }, 10000)
    return () => clearInterval(interval)
  }, [order])

  const currentStep = order ? (STATUS_TO_STEP[order.status] ?? 0) : 0
  const isCancelled = order?.status === 'cancelled' || order?.status === 'rejected'

  return (
    <div className="min-h-dvh bg-gray-50" dir="rtl">
      <header className="sticky top-0 z-30 bg-white border-b border-gray-200">
        <div className="container mx-auto px-4 h-14 flex items-center justify-between">
          <button onClick={onBack} className="flex items-center gap-2 text-sm font-medium text-gray-700"><ArrowRight className="w-4 h-4" /> العودة للمتجر</button>
          <h1 className="font-bold text-base text-gray-900">تتبع الطلب</h1>
        </div>
      </header>
      <div className="container mx-auto px-4 py-6 max-w-2xl">
        {!order ? (
          <div className="bg-white rounded-lg border p-6 mb-6">
            <div className="text-center mb-6">
              <div className="w-16 h-16 mx-auto rounded-lg bg-gray-50 flex items-center justify-center mb-3"><Package className="w-8 h-8 text-black" /></div>
              <h2 className="text-xl font-bold mb-1">تتبع طلبك</h2>
              <p className="text-sm text-gray-500">أدخل رقم الطلب لمعرفة حالته</p>
            </div>
            <div className="space-y-3">
              <Input value={searchOrderNumber} onChange={(e) => setSearchOrderNumber(e.target.value)} placeholder="FK123456789" className="h-12 rounded-xl text-center font-mono text-lg" dir="ltr" onKeyDown={(e) => e.key === 'Enter' && fetchOrder(searchOrderNumber)} />
              {error && <p className="text-sm text-red-600 text-center bg-red-50 rounded-lg p-2">{error}</p>}
              <Button onClick={() => fetchOrder(searchOrderNumber)} disabled={loading} className="w-full h-12 rounded-xl font-bold bg-black hover:bg-gray-800">{loading ? 'جاري البحث...' : 'تتبع الطلب'}{!loading && <Search className="w-4 h-4 mr-1" />}</Button>
            </div>
          </div>
        ) : (
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="space-y-4">
            <div className="bg-white rounded-lg border p-5">
              <div className="flex items-center justify-between mb-3">
                <div><p className="text-xs text-gray-500 mb-1">رقم الطلب</p><p className="font-bold text-lg" dir="ltr">{order.orderNumber}</p></div>
                <Button variant="outline" size="sm" onClick={() => { navigator.clipboard.writeText(order.orderNumber); toast.success('تم نسخ رقم الطلب') }} className="rounded-lg"><Copy className="w-4 h-4 ml-1" /> نسخ</Button>
              </div>
              <div className="flex items-center justify-between pt-3 border-t">
                <span className="text-sm text-gray-500">{new Date(order.createdAt).toLocaleString('ar', { day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' })}</span>
                <span className={`text-sm font-bold px-3 py-1 rounded-full ${isCancelled ? 'bg-red-50 text-red-700' : 'bg-gray-100 text-gray-700'}`}>{order.status === 'cancelled' ? 'ملغى' : order.status === 'rejected' ? 'مرفوض' : STEPS[currentStep]?.label || order.status}</span>
              </div>
            </div>
            {!isCancelled ? (
              <div className="bg-white rounded-lg border p-5">
                <h3 className="font-bold text-base mb-5">حالة الطلب</h3>
                <div className="relative">
                  {STEPS.map((step, idx) => {
                    const Icon = step.icon
                    const isDone = idx <= currentStep
                    const isCurrent = idx === currentStep
                    return (
                      <div key={step.id} className="flex gap-3 pb-6 last:pb-0 relative">
                        {idx < STEPS.length - 1 && <div className={`absolute right-5 top-10 w-0.5 h-[calc(100%-2rem)] ${idx < currentStep ? 'bg-black' : 'bg-gray-200'}`} />}
                        <motion.div animate={{ scale: isCurrent ? [1, 1.1, 1] : 1, backgroundColor: isDone ? '#7c3aed' : '#f3f4f6' }} transition={{ duration: 0.4, repeat: isCurrent ? Infinity : 0, repeatDelay: 1 }} className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 z-10 ${isDone ? 'text-white' : 'text-gray-400'}`}><Icon className="w-5 h-5" /></motion.div>
                        <div className="flex-1 pt-1.5"><p className={`font-bold text-sm ${isDone ? 'text-gray-900' : 'text-gray-500'}`}>{step.label}</p><p className="text-xs text-gray-500 mt-0.5">{step.description}</p></div>
                      </div>
                    )
                  })}
                </div>
              </div>
            ) : (
              <div className="bg-red-50 rounded-lg border border-red-200 p-5 text-center"><X className="w-10 h-10 text-red-500 mx-auto mb-2" /><h3 className="font-bold text-red-800">{order.status === 'cancelled' ? 'تم إلغاء الطلب' : 'تم رفض الطلب'}</h3></div>
            )}
            {order.locationUrl && <div className="bg-gray-50 border border-gray-200 rounded-lg p-4"><h4 className="font-bold text-sm mb-2 text-black flex items-center gap-2"><MapPin className="w-4 h-4" /> موقع التوصيل</h4><a href={order.locationUrl} target="_blank" rel="noopener noreferrer" className="flex items-center justify-between gap-2 bg-white rounded-lg p-2.5 border border-gray-200 hover:border-gray-400"><span className="text-sm font-medium text-black">فتح على خرائط جوجل</span><ExternalLink className="w-4 h-4 text-gray-500" /></a></div>}
            <div className="bg-white rounded-lg border p-5">
              <h3 className="font-bold text-base mb-4">عناصر الطلب</h3>
              <div className="space-y-3">{order.items?.map((item: any) => <div key={item.id} className="flex items-center justify-between text-sm"><div className="flex items-center gap-2"><span className="bg-gray-100 rounded px-2 py-0.5 text-xs font-bold">{item.quantity}×</span><span className="font-medium">{item.name}</span></div><span className="font-bold">{(item.price * item.quantity).toFixed(2)} ج.م</span></div>)}</div>
              <div className="border-t mt-4 pt-4 space-y-2 text-sm">
                <div className="flex justify-between text-gray-500"><span>المجموع الفرعي</span><span>{order.subtotal?.toFixed(2)} ج.م</span></div>
                <div className="flex justify-between text-gray-500"><span>رسوم التوصيل</span><span>{order.deliveryFee === 0 ? <span className="text-gray-500 font-bold">مجاني</span> : `${order.deliveryFee?.toFixed(2)} ج.م`}</span></div>
                <div className="flex justify-between font-bold text-base pt-1"><span>الإجمالي</span><span className="text-black">{order.total?.toFixed(2)} ج.م</span></div>
              </div>
            </div>
            <Button onClick={() => { setOrder(null); setSearchOrderNumber('') }} variant="outline" className="w-full rounded-xl">تتبع طلب آخر</Button>
          </motion.div>
        )}
      </div>
    </div>
  )
}
