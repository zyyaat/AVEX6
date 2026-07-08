
import { Home, Package, TrendingUp, User, LifeBuoy } from 'lucide-react'
import { useRouter, usePathname } from '@/lib/navigation'

export function BottomTabBar() {
  const router = useRouter()
  const pathname = usePathname()

  const tabs = [
    { id: 'home', icon: Home, label: 'الرئيسية', path: '/' },
    { id: 'history', icon: Package, label: 'سجلّي', path: '/history' },
    { id: 'earnings', icon: TrendingUp, label: 'أرباحي', path: '/earnings' },
    { id: 'support', icon: LifeBuoy, label: 'الدعم', path: '/support' },
    { id: 'profile', icon: User, label: 'حسابي', path: '/profile' },
  ]

  return (
    <nav
      className="sm:hidden fixed bottom-0 left-0 right-0 z-30 bg-white border-t border-gray-200 flex items-center justify-around h-14 px-1"
      style={{ paddingBottom: 'env(safe-area-inset-bottom, 0px)' }}
    >
      {tabs.map((tab) => {
        const Icon = tab.icon
        const active = pathname === tab.path
        return (
          <button
            key={tab.id}
            onClick={() => router.push(tab.path)}
            className={`flex flex-col items-center justify-center gap-0.5 flex-1 h-full hover:bg-gray-50 transition-colors active:bg-gray-100 ${
              active ? 'text-black' : 'text-gray-400'
            }`}
          >
            <Icon className="w-5 h-5" strokeWidth={active ? 2.2 : 1.5} />
            <span className={`text-[10px] ${active ? 'font-bold' : 'font-medium'}`}>{tab.label}</span>
          </button>
        )
      })}
    </nav>
  )
}
