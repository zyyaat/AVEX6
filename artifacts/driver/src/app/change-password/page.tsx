import { useState, useEffect } from 'react'
import { useRouter } from '@/lib/navigation'
import { Lock, Loader2, Eye, EyeOff } from 'lucide-react'
import { useAuth } from '@/store/auth'
import { toast } from 'sonner'

export default function ChangePasswordPage() {
  const router = useRouter()
  const { logout } = useAuth()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [showOld, setShowOld] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!oldPassword || !newPassword) return
    setLoading(true)
    try {
      // TODO: call backend change-password endpoint
      toast.success('تم تغيير كلمة المرور')
      logout()
      router.replace('/login')
    } catch (err: any) {
      toast.error(err.message || 'فشل تغيير كلمة المرور')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-dvh bg-white flex flex-col" dir="rtl">
      <div className="px-4 h-14 flex items-center" />
      <div className="flex-1 flex flex-col items-center justify-center px-6">
        <h1 className="text-xl font-bold mb-6">تغيير كلمة المرور</h1>
        <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-4">
          <div className="relative">
            <Lock className="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type={showOld ? 'text' : 'password'}
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              placeholder="كلمة المرور الحالية"
              className="w-full h-12 pr-11 pl-11 rounded-xl border border-gray-200 bg-gray-50 outline-none"
            />
            <button type="button" onClick={() => setShowOld(!showOld)} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
              {showOld ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
            </button>
          </div>
          <div className="relative">
            <Lock className="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type={showNew ? 'text' : 'password'}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="كلمة المرور الجديدة"
              className="w-full h-12 pr-11 pl-11 rounded-xl border border-gray-200 bg-gray-50 outline-none"
            />
            <button type="button" onClick={() => setShowNew(!showNew)} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
              {showNew ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
            </button>
          </div>
          <button type="submit" disabled={loading} className="w-full h-12 rounded-xl font-semibold text-white" style={{ backgroundColor: '#FF6B35' }}>
            {loading ? <Loader2 className="w-5 h-5 animate-spin mx-auto" /> : 'تأكيد'}
          </button>
        </form>
      </div>
    </div>
  )
}
