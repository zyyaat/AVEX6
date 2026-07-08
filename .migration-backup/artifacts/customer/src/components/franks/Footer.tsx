
import { Mail, MapPin, Clock, Facebook, Instagram, Twitter } from 'lucide-react'

export function Footer() {
  return (
    <footer id="contact" className="bg-black text-gray-400">
      <div className="container mx-auto px-4 py-10">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="col-span-2 md:col-span-1">
            <h3 className="text-xl font-bold text-white mb-2">AVEX</h3>
            <p className="text-sm leading-relaxed mb-4">منصة توصيل عالمية. تجربة سهلة، توصيل سريع.</p>
            <ul className="space-y-2 text-sm">
              <li className="flex items-center gap-2"><Mail className="w-3.5 h-3.5" /> hello@avex.com</li>
              <li className="flex items-center gap-2"><MapPin className="w-3.5 h-3.5" /> القاهرة، مصر</li>
              <li className="flex items-center gap-2"><Clock className="w-3.5 h-3.5" /> يومياً 10ص - 12م</li>
            </ul>
          </div>

          {/* Links */}
          {[
            { title: 'الشركة', links: ['من نحن', 'الوظائف', 'تواصل'] },
            { title: 'المساعدة', links: ['تتبع طلبك', 'الأسئلة الشائعة', 'الدعم'] },
            { title: 'قانوني', links: ['الخصوصية', 'الشروط', 'الكوكيز'] },
          ].map(col => (
            <div key={col.title}>
              <h4 className="text-xs font-semibold uppercase tracking-wider text-gray-500 mb-3">{col.title}</h4>
              <ul className="space-y-2">
                {col.links.map(l => <li key={l}><a href="#" className="text-sm hover:text-white transition-colors">{l}</a></li>)}
              </ul>
            </div>
          ))}

          {/* Social */}
          <div className="col-span-2 md:col-span-1">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-gray-500 mb-3">تابعنا</h4>
            <div className="flex gap-2">
              {[Facebook, Instagram, Twitter].map((Icon, i) => (
                <a key={i} href="#" className="w-9 h-9 rounded-lg bg-gray-900 hover:bg-gray-800 flex items-center justify-center transition-colors">
                  <Icon className="w-4 h-4" />
                </a>
              ))}
            </div>
          </div>
        </div>
      </div>
      <div className="border-t border-gray-900">
        <div className="container mx-auto px-4 py-4 flex flex-col sm:flex-row items-center justify-between gap-2">
          <span className="text-xs text-gray-600">© 2026 AVEX</span>
          <div className="flex gap-4">
            <a href="#" className="text-xs text-gray-600 hover:text-white transition-colors">الخصوصية</a>
            <a href="#" className="text-xs text-gray-600 hover:text-white transition-colors">الشروط</a>
            <a href="?admin=1" className="text-xs text-gray-600 hover:text-white transition-colors">لوحة التحكم</a>
          </div>
        </div>
      </div>
    </footer>
  )
}
