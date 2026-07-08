import { useState, useEffect } from 'react'
import { DollarSign, Loader2, Save } from 'lucide-react'
import { adminAPI } from '@/lib/api'
import { toast } from 'sonner'

export default function AdminTierPricesPage() {
  const [tiers, setTiers] = useState<any[]>([])
  const [zones, setZones] = useState<any[]>([])
  const [prices, setPrices] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [edits, setEdits] = useState<Record<string, any>>({})

  const load = () => {
    setLoading(true)
    Promise.all([adminAPI.getTiers(), adminAPI.getZones(), adminAPI.getTierPrices()])
      .then(([t, z, p]) => {
        setTiers(t.tiers || [])
        setZones(z.zones || [])
        setPrices(p.prices || [])
        const init: Record<string, any> = {}
        ;(p.prices || []).forEach((pr) => {
          init[`${pr.tierId}_${pr.zoneId}`] = pr
        })
        setEdits(init)
      })
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, [])

  const set = (tierId: string, zoneId: string, key: string, value: any) => {
    const k = `${tierId}_${zoneId}`
    setEdits({ ...edits, [k]: { ...(edits[k] || { tierId, zoneId }), [key]: value } })
  }

  const saveCell = async (tierId: string, zoneId: string) => {
    const cell = edits[`${tierId}_${zoneId}`]
    if (!cell) return
    try {
      await adminAPI.updateTierPrice(tierId, zoneId, {
        baseFee: cell.baseFee ?? 0,
        perKmFee: cell.perKmFee ?? 0,
        minFee: cell.minFee ?? 0,
        maxFee: cell.maxFee ?? 0,
        freeAbove: cell.freeAbove ?? 0,
        estimatedMinutes: cell.estimatedMinutes ?? 30,
      })
      toast.success('تم تحديث السعر')
    } catch (err: any) { toast.error(err.message) }
  }

  if (loading) return <div className="py-20 text-center"><Loader2 className="w-6 h-6 animate-spin mx-auto" /></div>

  return (
    <div dir="rtl">
      <h1 className="text-xl font-bold mb-1">مصفوفة الأسعار</h1>
      <p className="text-xs text-gray-500 mb-4">سعر التوصيل لكل مستوى × منطقة — هذه هي عمولة المندوب</p>

      <div className="bg-white rounded-lg border border-gray-200 overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-3 py-2 text-right text-xs">المستوى</th>
              {zones.map((z) => <th key={z.id} className="px-3 py-2 text-center text-xs">{z.nameAr}</th>)}
            </tr>
          </thead>
          <tbody>
            {tiers.map((t) => (
              <tr key={t.id} className="border-t border-gray-100">
                <td className="px-3 py-3">
                  <div className="flex items-center gap-2">
                    <div className="w-5 h-5 rounded-full" style={{ backgroundColor: t.color }} />
                    <span className="font-bold text-xs">{t.nameAr}</span>
                  </div>
                </td>
                {zones.map((z) => {
                  const cell = edits[`${t.id}_${z.id}`] || {}
                  return (
                    <td key={z.id} className="px-2 py-2 align-top">
                      <div className="space-y-1 min-w-[110px]">
                        <input type="number" step="0.01" placeholder="base" value={cell.baseFee ?? ''} onChange={(e) => set(t.id, z.id, 'baseFee', +e.target.value)}
                          className="w-full h-7 px-1.5 rounded border border-gray-200 text-xs focus:outline-none focus:border-black" />
                        <input type="number" step="0.01" placeholder="/كم" value={cell.perKmFee ?? ''} onChange={(e) => set(t.id, z.id, 'perKmFee', +e.target.value)}
                          className="w-full h-7 px-1.5 rounded border border-gray-200 text-xs focus:outline-none focus:border-black" />
                        <input type="number" step="0.01" placeholder="min" value={cell.minFee ?? ''} onChange={(e) => set(t.id, z.id, 'minFee', +e.target.value)}
                          className="w-full h-7 px-1.5 rounded border border-gray-200 text-xs focus:outline-none focus:border-black" />
                        <input type="number" step="0.01" placeholder="max" value={cell.maxFee ?? ''} onChange={(e) => set(t.id, z.id, 'maxFee', +e.target.value)}
                          className="w-full h-7 px-1.5 rounded border border-gray-200 text-xs focus:outline-none focus:border-black" />
                        <input type="number" placeholder="دقائق" value={cell.estimatedMinutes ?? ''} onChange={(e) => set(t.id, z.id, 'estimatedMinutes', +e.target.value)}
                          className="w-full h-7 px-1.5 rounded border border-gray-200 text-xs focus:outline-none focus:border-black" />
                        <button onClick={() => saveCell(t.id, z.id)} className="w-full h-6 rounded bg-black text-white text-[10px] flex items-center justify-center gap-0.5">
                          <Save className="w-2.5 h-2.5" /> حفظ
                        </button>
                      </div>
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="mt-3 text-xs text-gray-500 space-y-1">
        <p><b>base</b>: رسوم أساسية ثابتة (ج.م)</p>
        <p><b>/كم</b>: رسوم إضافية لكل كيلومتر (مطعم → عميل)</p>
        <p><b>min/max</b>: الحد الأدنى/الأقصى للرسوم</p>
        <p><b>دقائق</b>: زمن التوصيل المتوقع للعرض</p>
      </div>
    </div>
  )
}
