import { useState, useEffect } from 'react'
import { Award, Loader2, Save } from 'lucide-react'
import { adminAPI } from '@/lib/api'
import { toast } from 'sonner'

export default function AdminTiersPage() {
  const [tiers, setTiers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<Record<string, any>>({})

  const load = () => {
    setLoading(true)
    adminAPI.getTiers().then((r) => {
      setTiers(r.tiers || [])
      const init: Record<string, any> = {}
      ;(r.tiers || []).forEach((t) => {
        init[t.id] = {
          minAcceptanceRate: t.thresholds.minAcceptanceRate,
          minCompletionRate: t.thresholds.minCompletionRate,
          minCustomerRating: t.thresholds.minCustomerRating,
          minOnTimeRate: t.thresholds.minOnTimeRate,
          minShiftAdherence: t.thresholds.minShiftAdherence,
          minLifetimeOrders: t.thresholds.minLifetimeOrders,
        }
      })
      setEditing(init)
    }).finally(() => setLoading(false))
  }
  useEffect(() => { load() }, [])

  const save = async (id: string) => {
    try {
      await adminAPI.updateTierThresholds(id, editing[id])
      toast.success('تم تحديث شروط المستوى')
    } catch (err: any) { toast.error(err.message) }
  }

  const set = (id: string, key: string, value: any) => {
    setEditing({ ...editing, [id]: { ...editing[id], [key]: value } })
  }

  if (loading) return <div className="py-20 text-center"><Loader2 className="w-6 h-6 animate-spin mx-auto" /></div>

  return (
    <div dir="rtl">
      <h1 className="text-xl font-bold mb-4">مستويات المندوبين</h1>
      <div className="space-y-3">
        {tiers.map((t) => (
          <div key={t.id} className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-full flex items-center justify-center" style={{ backgroundColor: t.color }}>
                  <Award className="w-4 h-4 text-white" />
                </div>
                <div>
                  <p className="font-bold">{t.nameAr}</p>
                  <p className="text-xs text-gray-400">الترتيب: {t.sortOrder} • الكود: {t.code}</p>
                </div>
              </div>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2 mb-3">
              {[
                { k: 'minAcceptanceRate', label: 'نسبة القبول (%)', step: 1 },
                { k: 'minCompletionRate', label: 'نسبة الإكمال (%)', step: 1 },
                { k: 'minCustomerRating', label: 'التقييم', step: 0.1 },
                { k: 'minOnTimeRate', label: 'الالتزام بالوقت (%)', step: 1 },
                { k: 'minShiftAdherence', label: 'الحضور (%)', step: 1 },
                { k: 'minLifetimeOrders', label: 'طلبات تراكمية', step: 1 },
              ].map((f) => (
                <div key={f.k}>
                  <label className="text-[10px] text-gray-500 block mb-0.5">{f.label}</label>
                  <input type="number" step={f.step} value={editing[t.id]?.[f.k] ?? 0} onChange={(e) => set(t.id, f.k, +e.target.value)}
                    className="w-full h-9 px-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:border-black" />
                </div>
              ))}
            </div>
            <button onClick={() => save(t.id)} className="px-4 h-9 rounded-lg bg-black text-white text-sm font-medium flex items-center gap-2">
              <Save className="w-4 h-4" /> حفظ الشروط
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
