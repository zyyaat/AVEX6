import { useState, useEffect } from 'react'
import { useRouter } from '@/lib/navigation'
import { Inbox, AlertCircle, Clock, CheckCircle2, Loader2 } from 'lucide-react'
import { agentAPI } from '@/lib/api'

const typeLabels: Record<string, string> = { cancellation_request: 'طلب إلغاء', complaint: 'شكوى', other: 'أخرى' }

export default function AgentHome() {
  const router = useRouter()
  const [stats, setStats] = useState<any>(null)
  const [recent, setRecent] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = () => {
      Promise.all([agentAPI.getStats(), agentAPI.getTickets('open')])
        .then(([s, t]) => { setStats(s); setRecent(t.tickets || []) })
        .finally(() => setLoading(false))
    }
    load()
    const id = setInterval(load, 10000)
    return () => clearInterval(id)
  }, [])

  if (loading) return <div className="py-20 text-center"><Loader2 className="w-6 h-6 animate-spin mx-auto" /></div>

  const kpis = [
    { label: 'تذاكر مفتوحة', value: stats?.openCount ?? 0, icon: Inbox },
    { label: 'تذاكر عاجلة', value: stats?.urgentCount ?? 0, icon: AlertCircle },
    { label: 'المسندة إليّ', value: stats?.mineCount ?? 0, icon: Clock },
    { label: 'تذاكر اليوم', value: stats?.todayCount ?? 0, icon: CheckCircle2 },
  ]

  return (
    <div dir="rtl">
      <h1 className="text-xl font-bold mb-4">لوحة الدعم</h1>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
        {kpis.map((k) => {
          const Icon = k.icon
          return (
            <button key={k.label} onClick={() => router.push('/inbox')}
              className="bg-white rounded-lg border border-gray-200 p-4 text-right hover:border-black transition-fluent">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-gray-500">{k.label}</span>
                <Icon className="w-4 h-4 text-gray-400" />
              </div>
              <p className="text-2xl font-bold">{k.value}</p>
            </button>
          )
        })}
      </div>

      <div className="bg-white rounded-lg border border-gray-200 p-4 mb-4">
        <h3 className="font-bold text-sm mb-3">التذاكر حسب النوع</h3>
        {stats?.byType && Object.keys(stats.byType).length > 0 ? (
          <div className="space-y-2">
            {Object.entries(stats.byType).map(([t, c]: any) => (
              <div key={t} className="flex items-center justify-between text-sm">
                <span className="text-gray-700">{typeLabels[t] || t}</span>
                <span className="font-bold">{c}</span>
              </div>
            ))}
          </div>
        ) : <p className="text-sm text-gray-400">لا توجد تذاكر</p>}
      </div>

      <h3 className="font-bold text-sm mb-3">أحدث التذاكر المفتوحة</h3>
      {recent.length === 0 ? <p className="text-sm text-gray-400 text-center py-8">لا توجد تذاكر مفتوحة</p> :
       <div className="space-y-2">
         {recent.slice(0, 5).map((t) => (
           <button key={t.id} onClick={() => router.push(`/tickets/${t.id}`)}
             className="w-full text-right bg-white rounded-lg border border-gray-200 p-3 hover:border-black transition-fluent">
             <div className="flex items-center justify-between mb-1">
               <span className="text-xs font-bold bg-gray-100 px-2 py-0.5 rounded-full">{typeLabels[t.type] || t.type}</span>
               {t.priority === 'urgent' && <span className="text-[10px] bg-black text-white px-2 py-0.5 rounded-full">عاجل</span>}
               {t.priority === 'high' && <span className="text-[10px] bg-gray-700 text-white px-2 py-0.5 rounded-full">مرتفع</span>}
             </div>
             <p className="text-xs text-gray-700 line-clamp-2">{t.reason}</p>
             <p className="text-[10px] text-gray-400 mt-1">{new Date(t.createdAt).toLocaleString('ar-EG', { dateStyle: 'short', timeStyle: 'short' })}</p>
           </button>
         ))}
       </div>}
    </div>
  )
}
