import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Clock, Loader2, Save, Power } from 'lucide-react'
import { merchantAPI, type StoreHour } from '@/lib/api'
import { toast } from 'sonner'

const days = [
  { n: 0, label: 'الأحد', short: 'أحد' },
  { n: 1, label: 'الإثنين', short: 'إثن' },
  { n: 2, label: 'الثلاثاء', short: 'ثلا' },
  { n: 3, label: 'الأربعاء', short: 'أرب' },
  { n: 4, label: 'الخميس', short: 'خمي' },
  { n: 5, label: 'الجمعة', short: 'جمع' },
  { n: 6, label: 'السبت', short: 'سبت' },
]

export default function MerchantHoursPage() {
  const [hours, setHours] = useState<Record<number, { openTime: string; closeTime: string; isOpen: boolean }>>({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    merchantAPI.getHours().then((r) => {
      const init: Record<number, any> = {}
      days.forEach((d) => {
        init[d.n] = { openTime: '10:00', closeTime: '23:00', isOpen: true }
      })
      ;(r.hours || []).forEach((h) => {
        init[h.dayOfWeek] = {
          openTime: h.openTime || '10:00',
          closeTime: h.closeTime || '23:00',
          isOpen: h.isOpen,
        }
      })
      setHours(init)
    }).finally(() => setLoading(false))
  }, [])

  const save = async () => {
    setSaving(true)
    try {
      await merchantAPI.updateHours(
        days.map((d) => ({
          dayOfWeek: d.n,
          openTime: hours[d.n].openTime,
          closeTime: hours[d.n].closeTime,
          isOpen: hours[d.n].isOpen,
        }))
      )
      toast.success('تم حفظ ساعات العمل')
    } catch (e: any) {
      toast.error(e.message)
    } finally {
      setSaving(false)
    }
  }

  const toggleDay = (n: number) => {
    setHours({ ...hours, [n]: { ...hours[n], isOpen: !hours[n].isOpen } })
  }

  const setTime = (n: number, field: 'openTime' | 'closeTime', value: string) => {
    setHours({ ...hours, [n]: { ...hours[n], [field]: value } })
  }

  if (loading) {
    return (
      <div className="py-20 text-center">
        <Loader2 className="w-6 h-6 animate-spin mx-auto" />
      </div>
    )
  }

  const openDays = days.filter((d) => hours[d.n]?.isOpen).length

  return (
    <div dir="rtl">
      <div className="flex items-center justify-between mb-1">
        <h1 className="text-xl font-bold">ساعات العمل</h1>
        <span className="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full">
          {openDays} أيام مفتوحة
        </span>
      </div>
      <p className="text-xs text-gray-500 mb-4">حدد ساعات الفتح والإغلاق لكل يوم. يمكنك إغلاق أيام معينة.</p>

      <div className="bg-white rounded-2xl border border-gray-200 overflow-hidden shadow-fluent">
        {days.map((d, idx) => {
          const h = hours[d.n] || { openTime: '10:00', closeTime: '23:00', isOpen: true }
          return (
            <motion.div
              key={d.n}
              initial={{ opacity: 0, x: 10 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: idx * 0.04 }}
              className={`flex items-center gap-3 p-4 border-b border-gray-100 last:border-0 ${!h.isOpen ? 'bg-gray-50' : ''}`}
            >
              {/* Toggle */}
              <button
                onClick={() => toggleDay(d.n)}
                className={`w-12 h-7 rounded-full p-1 transition-fluent flex-shrink-0 ${
                  h.isOpen ? 'bg-black' : 'bg-gray-200'
                }`}
                aria-label={h.isOpen ? 'إغلاق اليوم' : 'فتح اليوم'}
              >
                <div className={`w-5 h-5 rounded-full bg-white shadow-fluent transition-fluent ${
                  h.isOpen ? 'translate-x-0' : '-translate-x-5'
                }`} />
              </button>

              {/* Day name */}
              <div className={`w-20 flex-shrink-0 ${h.isOpen ? 'text-black' : 'text-gray-400'}`}>
                <p className="font-bold text-sm">{d.label}</p>
              </div>

              {/* Time pickers */}
              <div className="flex items-center gap-2 flex-1">
                <input
                  type="time"
                  value={h.openTime}
                  disabled={!h.isOpen}
                  onChange={(e) => setTime(d.n, 'openTime', e.target.value)}
                  className="h-9 px-2 rounded-lg border border-gray-200 text-sm disabled:bg-gray-100 disabled:text-gray-400 focus:outline-none focus:border-black focus:ring-1 focus:ring-black"
                />
                <span className="text-gray-400 text-xs">-</span>
                <input
                  type="time"
                  value={h.closeTime}
                  disabled={!h.isOpen}
                  onChange={(e) => setTime(d.n, 'closeTime', e.target.value)}
                  className="h-9 px-2 rounded-lg border border-gray-200 text-sm disabled:bg-gray-100 disabled:text-gray-400 focus:outline-none focus:border-black focus:ring-1 focus:ring-black"
                />
              </div>

              {/* Status badge */}
              <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full flex-shrink-0 ${
                h.isOpen ? 'bg-black text-white' : 'bg-gray-200 text-gray-500'
              }`}>
                {h.isOpen ? 'مفتوح' : 'مغلق'}
              </span>
            </motion.div>
          )
        })}
      </div>

      {/* Quick actions */}
      <div className="flex gap-2 mt-4">
        <button
          onClick={() => {
            const next: any = {}
            days.forEach((d) => {
              next[d.n] = { openTime: '10:00', closeTime: '23:00', isOpen: true }
            })
            setHours(next)
            toast.info('تم تفعيل كل الأيام')
          }}
          className="flex-1 h-10 rounded-xl border border-gray-200 bg-white text-xs font-bold hover:bg-gray-50 transition-fluent"
        >
          تفعيل الكل
        </button>
        <button
          onClick={() => {
            const next: any = {}
            days.forEach((d) => {
              next[d.n] = { ...hours[d.n], isOpen: false }
            })
            setHours(next)
            toast.info('تم إغلاق كل الأيام')
          }}
          className="flex-1 h-10 rounded-xl border border-gray-200 bg-white text-xs font-bold hover:bg-gray-50 transition-fluent"
        >
          إغلاق الكل
        </button>
      </div>

      {/* Save button */}
      <button
        onClick={save}
        disabled={saving}
        className="w-full mt-3 h-12 rounded-xl bg-black text-white font-bold flex items-center justify-center gap-2 hover:bg-gray-800 transition-fluent disabled:opacity-50"
      >
        {saving ? <Loader2 className="w-5 h-5 animate-spin" /> : <Save className="w-5 h-5" />}
        حفظ التغييرات
      </button>
    </div>
  )
}
