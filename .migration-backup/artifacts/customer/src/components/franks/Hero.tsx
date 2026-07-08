
import { motion } from 'framer-motion'
import { ChevronLeft, Truck, Clock, ShieldCheck, Star } from 'lucide-react'
import { Button } from '@/components/ui/button'

const FEATURES = [
  { icon: Truck, title: 'توصيل سريع', desc: 'خلال 20-45 دقيقة' },
  { icon: Clock, title: '7 أيام', desc: 'من 10ص حتى 12م' },
  { icon: ShieldCheck, title: 'دفع آمن', desc: 'حماية كاملة' },
  { icon: Star, title: '4.8', desc: 'تقييم العملاء' },
]

export function Hero() {
  return (
    <section className="bg-white border-b border-gray-100">
      <div className="container mx-auto px-4 py-12 md:py-20">
        <div className="max-w-2xl">
          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-4xl md:text-5xl font-bold tracking-tight text-black leading-tight"
          >
            كل ما تحتاجه
            <br />
            <span className="text-gray-400">يصلك بسرعة</span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="mt-4 text-base text-gray-500 max-w-md"
          >
            اطلب طعامك المفضل من أفضل المطاعم. توصيل سريع، دفع آمن، تجربة سهلة.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="mt-6"
          >
            <Button size="lg" className="bg-black hover:bg-gray-800 text-white rounded-lg h-11 px-6 text-sm font-medium" asChild>
              <a href="#menu">
                اطلب الآن
                <ChevronLeft className="w-4 h-4 mr-1" />
              </a>
            </Button>
          </motion.div>

          {/* Features */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="grid grid-cols-2 md:grid-cols-4 gap-3 mt-10"
          >
            {FEATURES.map((f, i) => {
              const Icon = f.icon
              return (
                <div key={i} className="flex items-center gap-2.5">
                  <div className="w-9 h-9 rounded-lg bg-gray-50 flex items-center justify-center flex-shrink-0">
                    <Icon className="w-4 h-4 text-black" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-black">{f.title}</p>
                    <p className="text-xs text-gray-400">{f.desc}</p>
                  </div>
                </div>
              )
            })}
          </motion.div>
        </div>
      </div>
    </section>
  )
}
