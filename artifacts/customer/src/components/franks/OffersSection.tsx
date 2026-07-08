
import { motion } from 'framer-motion'
import { Gift, Percent, Truck, Clock } from 'lucide-react'

const OFFERS = [
  { icon: Truck, title: 'توصيل مجاني', desc: 'للطلبات فوق 30 ج.م', code: 'FREEDEL' },
  { icon: Percent, title: 'خصم 30%', desc: 'على طلبك الأول', code: 'AVEX30' },
  { icon: Gift, title: 'وجبة عائلية', desc: '4 برغر + 4 مشروبات بـ 99 ج.م', code: 'FAMILY99' },
  { icon: Clock, title: 'ساعة الغداء', desc: 'خصم 15% من 12ظ - 3م', code: 'LUNCH15' },
]

export function OffersSection() {
  return (
    <section id="offers" className="py-8 md:py-10 bg-gray-50 border-y border-gray-100">
      <div className="container mx-auto px-4">
        <h3 className="text-base font-bold text-black mb-4">عروض</h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          {OFFERS.map((o, i) => {
            const Icon = o.icon
            return (
              <div key={i} className="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 hover:shadow-fluent transition-fluent cursor-pointer">
                <div className="flex items-start justify-between mb-2">
                  <div className="w-9 h-9 rounded-lg bg-gray-50 flex items-center justify-center">
                    <Icon className="w-4 h-4 text-black" />
                  </div>
                  <span className="text-[10px] font-mono text-gray-400 bg-gray-50 rounded px-1.5 py-0.5">{o.code}</span>
                </div>
                <h4 className="font-medium text-sm text-black">{o.title}</h4>
                <p className="text-xs text-gray-400 mt-0.5">{o.desc}</p>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
