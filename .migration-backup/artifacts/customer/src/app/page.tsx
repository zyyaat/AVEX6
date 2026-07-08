
import { useState, useEffect, Suspense } from 'react'
import { useSearchParams, useRouter } from '@/lib/navigation'
import { Header } from '@/components/franks/Header'
import { Hero } from '@/components/franks/Hero'
import { HomeRestaurants } from '@/components/franks/HomeRestaurants'
import { RestaurantPage } from '@/components/franks/RestaurantPage'
import { OffersSection } from '@/components/franks/OffersSection'
import { AboutSection } from '@/components/franks/AboutSection'
import { Footer } from '@/components/franks/Footer'
import { CartDrawer } from '@/components/franks/CartDrawer'
import { CheckoutDialog } from '@/components/franks/CheckoutDialog'
import { OrderSuccessDialog } from '@/components/franks/OrderSuccessDialog'
import { FloatingCartButton } from '@/components/franks/FloatingCartButton'
import { BottomTabBar } from '@/components/franks/BottomTabBar'
import { OrderTracking } from '@/components/franks/OrderTracking'
import { MyOrders } from '@/components/franks/MyOrders'
import { AccountPage } from '@/components/franks/AccountPage'
import { AuthDialog } from '@/components/franks/AuthDialog'
import { SearchPage } from '@/components/franks/SearchPage'

function HomeContent() {
  const [checkoutOpen, setCheckoutOpen] = useState(false)
  const [successOpen, setSuccessOpen] = useState(false)
  const [orderNumber, setOrderNumber] = useState('')
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login')

  const searchParams = useSearchParams()
  const router = useRouter()

  const trackOrderNumber = searchParams.get('track')
  const myOrdersMode = searchParams.get('myorders') === '1'
  const accountMode = searchParams.get('account') === '1'
  const searchMode = searchParams.get('search') === '1'
  const restaurantId = searchParams.get('restaurant')
  const authParam = searchParams.get('auth')

  useEffect(() => {
    if (authParam === 'login' || authParam === 'register') {
      const timer = setTimeout(() => {
        setAuthMode(authParam as 'login' | 'register')
        setAuthOpen(true)
        const url = new URL(window.location.href)
        url.searchParams.delete('auth')
        window.history.replaceState({}, '', url.toString())
      }, 0)
      return () => clearTimeout(timer)
    }
    return undefined
  }, [authParam])

  const goHome = () => router.push('/')
  const handleLoginRequired = () => { setAuthMode('login'); setAuthOpen(true) }
  const handleCheckout = () => setCheckoutOpen(true)
  const handleOrderSuccess = (num: string) => { setOrderNumber(num); setSuccessOpen(true) }

  if (trackOrderNumber !== null) return <><OrderTracking initialOrderNumber={trackOrderNumber || undefined} onBack={goHome} /><BottomTabBar /></>
  if (myOrdersMode) return <><MyOrders onBack={goHome} onLoginRequired={handleLoginRequired} /><AuthDialog open={authOpen} onOpenChange={setAuthOpen} initialMode={authMode} /><BottomTabBar /></>
  if (accountMode) return <><AccountPage onBack={goHome} onLoginRequired={handleLoginRequired} /><AuthDialog open={authOpen} onOpenChange={setAuthOpen} initialMode={authMode} /><BottomTabBar /></>
  if (searchMode) return <><SearchPage onBack={goHome} /><BottomTabBar /></>
  if (restaurantId) return <><RestaurantPage restaurantId={restaurantId} onBack={goHome} /><BottomTabBar /></>

  return (
    <div className="min-h-dvh flex flex-col bg-white pb-14 sm:pb-0">
      <Header />
      <main className="flex-1">
        <Hero />
        <HomeRestaurants />
        <OffersSection />
        <AboutSection />
      </main>
      <Footer />
      <CartDrawer onCheckout={handleCheckout} />
      <FloatingCartButton />
      <CheckoutDialog open={checkoutOpen} onOpenChange={setCheckoutOpen} onSuccess={handleOrderSuccess} />
      <OrderSuccessDialog key={orderNumber} open={successOpen} onOpenChange={setSuccessOpen} orderNumber={orderNumber} />
      <AuthDialog open={authOpen} onOpenChange={setAuthOpen} initialMode={authMode} />
      <BottomTabBar />
    </div>
  )
}

export default function Home() {
  return (
    <Suspense fallback={<div className="min-h-dvh flex items-center justify-center"><div className="animate-pulse text-gray-400">جاري التحميل...</div></div>}>
      <HomeContent />
    </Suspense>
  )
}
