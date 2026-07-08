
import { useEffect, useState } from 'react'
import { ShoppingBag, Search, User, LogOut, Package, ChevronDown, MapPin } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useCart } from '@/store/cart'
import { useAuth } from '@/store/auth'
import { useRouter } from '@/lib/navigation'

export function Header() {
  const { getTotalItems, setOpen } = useCart()
  const { user, isAuthenticated, logout, initialize } = useAuth()
  const itemCount = getTotalItems()
  const [scrolled, setScrolled] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const router = useRouter()

  useEffect(() => {
    initialize()
    const handleScroll = () => setScrolled(window.scrollY > 8)
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [initialize])

  return (
    <header className={`sticky top-0 z-40 w-full bg-white transition-all ${
      scrolled ? 'border-b border-gray-200 shadow-fluent' : 'border-b border-gray-100'
    }`}>
      <div className="container mx-auto px-4 h-14 flex items-center justify-between gap-4">
        {/* Logo */}
        <a href="/" className="flex items-center gap-2">
          <span className="text-xl font-bold tracking-tight text-black">AVEX</span>
        </a>

        {/* Location - desktop */}
        <button className="hidden md:flex items-center gap-1.5 text-sm text-gray-500 hover:text-black transition-colors">
          <MapPin className="w-3.5 h-3.5" />
          <span>توصيل إلى</span>
          <span className="font-medium text-black">منزلك</span>
          <ChevronDown className="w-3.5 h-3.5" />
        </button>

        {/* Search - desktop */}
        <div className="hidden lg:flex flex-1 max-w-sm relative">
          <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="ابحث عن مطاعم، أطباق..."
            className="w-full h-9 bg-gray-50 border border-gray-200 rounded-lg pr-9 pl-4 text-sm placeholder:text-gray-400 focus:bg-white focus:border-gray-400 focus:outline-none transition-all"
            onClick={() => router.push('/#menu')}
            readOnly
          />
        </div>

        {/* Right */}
        <div className="flex items-center gap-2">
          {isAuthenticated && user ? (
            <div className="relative">
              <button
                onClick={() => setUserMenuOpen(!userMenuOpen)}
                className="flex items-center gap-2 hover:bg-gray-50 rounded-lg px-2 py-1.5 transition-colors"
              >
                <div className="w-7 h-7 rounded-full bg-black flex items-center justify-center">
                  <span className="text-xs font-medium text-white">{user.name.charAt(0)}</span>
                </div>
                <span className="hidden sm:inline text-sm font-medium text-gray-700 max-w-[80px] truncate">{user.name}</span>
                <ChevronDown className="w-3.5 h-3.5 text-gray-400" />
              </button>
              {userMenuOpen && (
                <>
                  <div className="fixed inset-0 z-40" onClick={() => setUserMenuOpen(false)} />
                  <div className="absolute left-0 mt-2 w-56 bg-white rounded-lg shadow-fluent-lg border border-gray-200 overflow-hidden z-50">
                    <div className="p-3 border-b border-gray-100">
                      <p className="font-medium text-sm text-black">{user.name}</p>
                      <p className="text-xs text-gray-500" dir="ltr">{user.phone}</p>
                      {user.loyaltyPoints > 0 && <p className="text-[10px] text-gray-500 mt-1">{user.loyaltyPoints} نقطة</p>}
                    </div>
                    <button onClick={() => { setUserMenuOpen(false); router.push('/?myorders=1') }} className="w-full flex items-center gap-2 px-3 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors text-right">
                      <Package className="w-4 h-4" /> طلباتي
                    </button>
                    <button onClick={() => { setUserMenuOpen(false); router.push('/?account=1') }} className="w-full flex items-center gap-2 px-3 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors text-right">
                      <User className="w-4 h-4" /> حسابي
                    </button>
                    <button onClick={() => { logout(); setUserMenuOpen(false); router.push('/') }} className="w-full flex items-center gap-2 px-3 py-2.5 text-sm text-red-600 hover:bg-gray-50 transition-colors text-right border-t border-gray-100">
                      <LogOut className="w-4 h-4" /> خروج
                    </button>
                  </div>
                </>
              )}
            </div>
          ) : (
            <Button onClick={() => router.push('/?auth=login')} variant="ghost" size="sm" className="rounded-lg text-gray-700 hover:bg-gray-50 text-sm font-medium">
              <User className="w-4 h-4 ml-1.5" />
              <span className="hidden sm:inline">دخول</span>
            </Button>
          )}

          {/* Cart */}
          <Button onClick={() => setOpen(true)} size="sm" className="relative bg-black hover:bg-gray-800 text-white rounded-lg h-9 px-3 transition-fluent">
            <ShoppingBag className="w-4 h-4 ml-1.5" />
            <span className="hidden sm:inline text-sm font-medium">السلة</span>
            {itemCount > 0 && (
              <span className="absolute -top-1 -right-1 flex h-4.5 min-w-4.5 items-center justify-center rounded-full bg-gray-200 text-black text-[10px] font-bold border border-white">
                {itemCount}
              </span>
            )}
          </Button>
        </div>
      </div>
    </header>
  )
}
