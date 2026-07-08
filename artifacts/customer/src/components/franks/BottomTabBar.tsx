
import { Home, Search, Package, User } from 'lucide-react'
import { useRouter } from '@/lib/navigation'
import { useAuth } from '@/store/auth'

export function BottomTabBar() {
  const router = useRouter()
  const { isAuthenticated } = useAuth()

  const tabs = [
    { id: 'home', icon: Home, label: 'الرئيسية', action: () => router.push('/') },
    { id: 'search', icon: Search, label: 'بحث', action: () => router.push('/?search=1') },
    { id: 'orders', icon: Package, label: 'طلباتي', action: () => router.push('/?myorders=1') },
    { id: 'account', icon: User, label: 'حسابي', action: () => router.push('/?account=1') },
  ]

  return (
    <nav className="sm:hidden fixed bottom-0 left-0 right-0 z-30 bg-white border-t border-gray-200 flex items-center justify-around h-14 px-2"
      style={{ paddingBottom: 'env(safe-area-inset-bottom, 0px)' }}
    >
      {tabs.map((tab) => {
        const Icon = tab.icon
        return (
          <button
            key={tab.id}
            onClick={tab.action}
            className="flex flex-col items-center justify-center gap-0.5 flex-1 h-full hover:bg-gray-50 transition-colors active:bg-gray-100"
          >
            <Icon className="w-5 h-5 text-gray-500" strokeWidth={1.5} />
            <span className="text-[10px] text-gray-500 font-medium">{tab.label}</span>
          </button>
        )
      })}
    </nav>
  )
}
