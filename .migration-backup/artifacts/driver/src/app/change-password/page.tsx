
import { useState } from 'react'
import { useRouter } from '@/lib/navigation'
import { motion } from 'framer-motion'
import {
  ShieldCheck, Lock, ArrowLeft, Loader2, Eye, EyeOff, AlertCircle, Check,
} from 'lucide-react'
import { useAuth } from '@/store/auth'
import { driverAuthAPI } from '@/lib/api'
import { toast } from 'sonner'

export default function ChangePasswordPage() {
  const router = useRouter()
  const { token, logout, setMustChangePassword } = useAuth()
  const [oldPwd, setOldPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [confirmPwd, setConfirmPwd] = useState('')
  const [showOld, setShowOld] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const strength = computeStrength(newPwd)
  const canSubmit =
    oldPwd.length > 0 &&
    newPwd.length >= 6 &&
    confirmPwd === newPwd &&
    !submitting

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (newPwd.length < 6) {
      setError('كلمة المرور الجديدة يجب أن تكون 6 أحرف على الأقل')
      return
    }
    if (newPwd !== confirmPwd) {
      setError('كلمة المرور وتأكيدها غير متطابقين')
      return
    }
    if (newPwd === oldPwd) {
      setError('كلمة المرور الجديدة يجب أن تكون مختلفة عن الحالية')
      return
    }
    setSubmitting(true)
    try {
      await driverAuthAPI.changePassword({ oldPassword: oldPwd, newPassword: newPwd })
      setMustChangePassword(false)
      toast.success('تم تغيير كلمة المرور بنجاح')
      router.replace('/')
    } catch (err: any) {
      setError(err.message || 'فشل تغيير كلمة المرور')
    } finally {
      setSubmitting(false)
    }
  }

  const handleLogout = () => {
    logout()
    router.replace('/login')
  }

  return (
    <div className="min-h-dvh bg-white flex flex-col" dir="rtl">
      {/* Top bar */}
      <div className="px-4 h-14 flex items-center justify-between">
        <button
          onClick={handleLogout}
          className="w-9 h-9 rounded-full hover:bg-gray-100 flex items-center justify-center transition-fluent"
          aria-label="خروج"
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <span className="text-sm text-gray-500">تغيير كلمة المرور</span>
        <div className="w-9" />
      </div>

      {/* Hero */}
      <div className="flex-1 flex flex-col items-center justify-center px-6 -mt-6">
        <motion.div
          initial={{ opacity: 0, scale: 0.85, y: 8 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          transition={{ duration: 0.4, ease: 'easeOut' }}
          className="w-16 h-16 rounded-2xl bg-black flex items-center justify-center mb-4 shadow-fluent-lg"
        >
          <ShieldCheck className="w-8 h-8 text-white" strokeWidth={2} />
        </motion.div>

        <h1 className="text-xl font-bold mb-1">تغيير كلمة المرور</h1>
        <p className="text-sm text-gray-500 mb-6 text-center max-w-xs">
          لأمان حسابك، يجب تغيير كلمة المرور الافتراضية قبل البدء في العمل
        </p>

        <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-3" noValidate>
          {error && (
            <motion.div
              initial={{ opacity: 0, y: -4 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-gray-50 border border-gray-200 rounded-lg p-3 flex items-start gap-2 text-sm"
              role="alert"
            >
              <AlertCircle className="w-4 h-4 text-black flex-shrink-0 mt-0.5" />
              <span className="text-gray-700">{error}</span>
            </motion.div>
          )}

          {/* Old password */}
          <PasswordField
            label="كلمة المرور الحالية"
            value={oldPwd}
            onChange={setOldPwd}
            show={showOld}
            toggle={() => setShowOld(!showOld)}
            placeholder="••••••"
            autoComplete="current-password"
          />

          {/* New password */}
          <div>
            <PasswordField
              label="كلمة المرور الجديدة"
              value={newPwd}
              onChange={setNewPwd}
              show={showNew}
              toggle={() => setShowNew(!showNew)}
              placeholder="6 أحرف على الأقل"
              autoComplete="new-password"
            />
            {/* Strength indicator */}
            {newPwd.length > 0 && (
              <div className="mt-1.5 flex items-center gap-1.5">
                {[1, 2, 3, 4].map((i) => (
                  <div
                    key={i}
                    className={`h-1 flex-1 rounded-full transition-fluent ${
                      i <= strength
                        ? strength <= 1 ? 'bg-gray-400' : strength === 2 ? 'bg-gray-600' : 'bg-black'
                        : 'bg-gray-200'
                    }`}
                  />
                ))}
                <span className="text-[10px] text-gray-500 w-10 text-left">
                  {strength <= 1 ? 'ضعيفة' : strength === 2 ? 'متوسطة' : strength === 3 ? 'جيدة' : 'قوية'}
                </span>
              </div>
            )}
          </div>

          {/* Confirm password */}
          <div>
            <PasswordField
              label="تأكيد كلمة المرور"
              value={confirmPwd}
              onChange={setConfirmPwd}
              show={showConfirm}
              toggle={() => setShowConfirm(!showConfirm)}
              placeholder="أعد إدخال كلمة المرور"
              autoComplete="new-password"
            />
            {confirmPwd.length > 0 && confirmPwd === newPwd && (
              <div className="mt-1.5 flex items-center gap-1 text-[11px] text-black">
                <Check className="w-3 h-3" />
                <span>متطابقة</span>
              </div>
            )}
          </div>

          <button
            type="submit"
            disabled={!canSubmit}
            className="w-full h-12 rounded-lg bg-black text-white font-medium hover:bg-gray-800 active:bg-gray-900 transition-fluent disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-2"
          >
            {submitting ? <Loader2 className="w-5 h-5 animate-spin" /> : 'حفظ ومتابعة'}
          </button>
        </form>

        <button
          onClick={handleLogout}
          className="mt-4 text-xs text-gray-400 hover:text-gray-600 transition-fluent"
        >
          تسجيل الخروج
        </button>
      </div>
    </div>
  )
}

// ===== Password field component =====
function PasswordField({
  label, value, onChange, show, toggle, placeholder, autoComplete,
}: {
  label: string
  value: string
  onChange: (v: string) => void
  show: boolean
  toggle: () => void
  placeholder: string
  autoComplete: string
}) {
  return (
    <div>
      <label className="text-xs text-gray-500 mb-1 block">{label}</label>
      <div className="relative">
        <Lock className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        <input
          type={show ? 'text' : 'password'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          autoComplete={autoComplete}
          className="w-full h-12 pr-10 pl-10 rounded-lg border border-gray-200 bg-white text-right focus:outline-none focus:border-black focus:ring-1 focus:ring-black transition-fluent"
        />
        <button
          type="button"
          onClick={toggle}
          className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-700 transition-fluent"
          aria-label={show ? 'إخفاء' : 'إظهار'}
        >
          {show ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
        </button>
      </div>
    </div>
  )
}

// ===== Strength helper =====
function computeStrength(p: string): number {
  let s = 0
  if (p.length >= 6) s++
  if (p.length >= 10) s++
  if (/[A-Z]/.test(p) && /[a-z]/.test(p)) s++
  if (/\d/.test(p) && /[^A-Za-z0-9]/.test(p)) s++
  return Math.min(s, 4)
}
