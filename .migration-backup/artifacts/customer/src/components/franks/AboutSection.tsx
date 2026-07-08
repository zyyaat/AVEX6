
import { Zap, Award, Heart, Users } from 'lucide-react'

const VALUES = [
  { icon: Zap, title: 'سرعة', desc: 'توصيل خلال 20-45 دقيقة' },
  { icon: Heart, title: 'جودة', desc: 'مكونات طازجة 100%' },
  { icon: Award, title: 'موثوقية', desc: 'أعلى معايير الأمان' },
  { icon: Users, title: 'ثقة', desc: 'أكثر من 5000 عميل' },
]

export function AboutSection() {
  return (
    <section id="about" className="py-12 md:py-16 bg-white">
      <div className="container mx-auto px-4 max-w-2xl">
        <h2 className="text-2xl font-bold text-black mb-2">من نحن</h2>
        <p className="text-gray-500 text-sm leading-relaxed mb-8">
          AVEX منصة توصيل عالمية تقدم تجربة طلب سلسة وسريعة. من الطعام إلى البقالة، نوصل كل ما تحتاجه بكفاءة وأمان.
        </p>
        <div className="grid grid-cols-2 gap-4">
          {VALUES.map((v, i) => {
            const Icon = v.icon
            return (
              <div key={i} className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-gray-50 flex items-center justify-center flex-shrink-0">
                  <Icon className="w-4 h-4 text-black" />
                </div>
                <div>
                  <p className="font-medium text-sm text-black">{v.title}</p>
                  <p className="text-xs text-gray-400">{v.desc}</p>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
