import { useState, useEffect } from 'react'
import { useRouter, usePathname } from '@/lib/navigation'
import { Link } from '@/lib/navigation'
import { Store, LayoutDashboard, Package, UtensilsCrossed, Clock, Power, LogOut, Menu, X } from 'lucide-react'
import { useAuth } from '@/store/auth'

const navItems = [
  { href: '/', label: 'لوحة المعلومات', icon: LayoutDashboard },
  { href: '/orders', label: 'الطلبات', icon: Package },
  { href: '/menu', label: 'المنيو', icon: UtensilsCrossed },
  { href: '/hours', label: 'ساعات العمل', icon: Clock },
]

export function MerchantLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const { isAuthenticated, mustChangePassword, merchant, logout, initialize } = useAuth()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [bootChecked, setBootChecked] = useState(false)

  useEffect(() => { initialize().then(() => setBootChecked(true)) }, [initialize])
  useEffect(() => {
    if (!bootChecked) return
    if (!isAuthenticated && pathname !== '/login') {
      router.replace('/login')
      return
    }
    if (isAuthenticated && mustChangePassword && pathname !== '/change-password') {
      router.replace('/change-password')
      return
    }
  }, [bootChecked, isAuthenticated, mustChangePassword, pathname, router])

  if (pathname === '/login' || pathname === '/change-password') return <>{children}</>
  if (!bootChecked) return <div className="min-h-dvh flex items-center justify-center"><div className="animate-pulse text-gray-400">جاري التحميل...</div></div>
  if (!isAuthenticated) return null

  const restaurant = merchant?.restaurant
  const isActive = restaurant?.isActive

  return (
    <div className="min-h-dvh bg-gray-50 flex" dir="rtl">
      <aside className="hidden md:flex md:flex-col md:w-60 md:fixed md:inset-y-0 md:right-0 bg-white border-l border-gray-200">
        <div className="h-14 px-4 flex items-center border-b border-gray-200">
          <span className="font-bold text-lg">🏪 {restaurant?.nameAr || 'AVEX Merchant'}</span>
        </div>
        <nav className="flex-1 py-2">
          {navItems.map((item) => {
            const Icon = item.icon
            const active = pathname === item.href
            return (
              <Link key={item.href} href={item.href}
                className={`flex items-center gap-3 px-4 py-2.5 text-sm transition-fluent ${
                  active ? 'bg-black text-white font-bold' : 'text-gray-600 hover:bg-gray-100'
                }`}>
                <Icon className="w-4 h-4" /> {item.label}
              </Link>
            )
          })}
        </nav>
        <div className="border-t border-gray-200 p-3">
          <div className="text-xs text-gray-500 mb-2 px-2">{merchant?.name}</div>
          <button onClick={() => { logout(); router.replace('/login') }}
            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">
            <LogOut className="w-4 h-4" /> خروج
          </button>
        </div>
      </aside>

      {sidebarOpen && (
        <div className="md:hidden fixed inset-0 z-40 bg-black/50" onClick={() => setSidebarOpen(false)}>
          <aside className="absolute right-0 top-0 bottom-0 w-72 bg-white flex flex-col" onClick={(e) => e.stopPropagation()}>
            <div className="h-14 px-4 flex items-center justify-between border-b border-gray-200">
              <span className="font-bold">🏪 {restaurant?.nameAr}</span>
              <button onClick={() => setSidebarOpen(false)} className="w-8 h-8 rounded-full hover:bg-gray-100 flex items-center justify-center">
                <X className="w-5 h-5" />
              </button>
            </div>
            <nav className="flex-1 py-2">
              {navItems.map((item) => {
                const Icon = item.icon
                const active = pathname === item.href
                return (
                  <Link key={item.href} href={item.href} onClick={() => setSidebarOpen(false)}
                    className={`flex items-center gap-3 px-4 py-3 text-sm transition-fluent ${
                      active ? 'bg-black text-white font-bold' : 'text-gray-600 hover:bg-gray-100'
                    }`}>
                    <Icon className="w-4 h-4" /> {item.label}
                  </Link>
                )
              })}
            </nav>
          </aside>
        </div>
      )}

      <div className="flex-1 md:mr-60">
        <header className="sticky top-0 z-30 bg-white border-b border-gray-200 h-14 flex items-center justify-between px-4">
          <div className="flex items-center gap-2">
            <button onClick={() => setSidebarOpen(true)} className="md:hidden w-9 h-9 rounded-full hover:bg-gray-100 flex items-center justify-center">
              <Menu className="w-5 h-5" />
            </button>
            <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-bold ${
              isActive ? 'bg-black text-white' : 'bg-gray-200 text-gray-600'
            }`}>
              <div className={`w-2 h-2 rounded-full ${isActive ? 'bg-white animate-pulse' : 'bg-gray-500'}`} />
              {isActive ? 'مفتوح' : 'مغلق'}
            </div>
          </div>
          <div className="md:hidden font-bold">🏪</div>
          <div className="hidden md:block text-sm text-gray-500">{merchant?.name}</div>
        </header>
        <main className="p-4 md:p-6 max-w-5xl mx-auto">{children}</main>
      </div>
    </div>
  )
}
