import { useState, useEffect } from 'react'
import { Users, Loader2, Check, X, Plus, Phone, IdCard } from 'lucide-react'
import { adminAPI } from '@/lib/api'
import { toast } from 'sonner'

export default function AdminApplicationsPage() {
  const [apps, setApps] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', phone: '', nationalId: '', licenseNumber: '', vehicleType: 'motorcycle', vehiclePlate: '', address: '', emergencyPhone: '' })

  const load = () => {
    setLoading(true)
    adminAPI.getApplications().then((r) => setApps(r.applications || [])).finally(() => setLoading(false))
  }
  useEffect(() => { load() }, [])

  const verify = async (a: any) => {
    try {
      const r = await adminAPI.verifyApplication(a.id)
      toast.success(`تم التوثيق — كلمة المرور الأولية للمندوب: ${r.initialPassword}`)
      load()
    } catch (e: any) { toast.error(e.message) }
  }
  const reject = async (a: any) => {
    const reason = prompt('سبب الرفض:')
    if (!reason) return
    try { await adminAPI.rejectApplication(a.id, reason); toast.success('تم الرفض'); load() }
    catch (e: any) { toast.error(e.message) }
  }
  const create = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await adminAPI.createApplication(form)
      toast.success('تم إنشاء الطلب')
      setShowCreate(false)
      setForm({ name: '', phone: '', nationalId: '', licenseNumber: '', vehicleType: 'motorcycle', vehiclePlate: '', address: '', emergencyPhone: '' })
      load()
    } catch (err: any) { toast.error(err.message) }
  }

  return (
    <div dir="rtl">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-bold">طلبات الالتحاق</h1>
        <button onClick={() => setShowCreate(true)} className="px-3 h-9 rounded-lg bg-black text-white text-sm font-medium flex items-center gap-2">
          <Plus className="w-4 h-4" /> طلب جديد
        </button>
      </div>
      {loading ? <div className="py-20 text-center"><Loader2 className="w-6 h-6 animate-spin mx-auto" /></div> :
       apps.length === 0 ? <div className="py-20 text-center text-gray-400">لا توجد طلبات</div> :
       <div className="space-y-3">
         {apps.map((a) => (
           <div key={a.id} className="bg-white rounded-lg border border-gray-200 p-4">
             <div className="flex items-start justify-between mb-2">
               <div>
                 <p className="font-bold">{a.name}</p>
                 <p className="text-xs text-gray-500" dir="ltr">{a.phone}</p>
               </div>
               <span className={`text-xs px-2 py-0.5 rounded-full ${
                 a.status === 'pending' ? 'bg-gray-100 text-gray-700' :
                 a.status === 'verified' ? 'bg-black text-white' : 'bg-gray-200 text-gray-500'
               }`}>
                 {a.status === 'pending' ? 'بانتظار المراجعة' : a.status === 'verified' ? 'موثّق' : 'مرفوض'}
               </span>
             </div>
             <div className="grid grid-cols-2 md:grid-cols-3 gap-2 text-xs text-gray-600 mb-3">
               <div>الرقم القومي: {a.nationalId}</div>
               <div>رخصة: {a.licenseNumber}</div>
               <div>مركبة: {a.vehicleType} {a.vehiclePlate && `(${a.vehiclePlate})`}</div>
               {a.address && <div>العنوان: {a.address}</div>}
               {a.emergencyPhone && <div>طوارئ: {a.emergencyPhone}</div>}
               {a.rejectionReason && <div className="text-red-600">سبب الرفض: {a.rejectionReason}</div>}
             </div>
             {a.status === 'pending' && (
               <div className="flex gap-2">
                 <button onClick={() => verify(a)} className="flex-1 h-9 rounded-lg bg-black text-white text-sm font-medium flex items-center justify-center gap-1.5">
                   <Check className="w-4 h-4" /> توثيق وإنشاء حساب
                 </button>
                 <button onClick={() => reject(a)} className="px-4 h-9 rounded-lg border border-gray-200 text-sm font-medium flex items-center gap-1.5">
                   <X className="w-4 h-4" /> رفض
                 </button>
               </div>
             )}
             {a.status === 'verified' && a.driverId && (
               <div className="text-xs text-gray-500">رقم حساب المندوب: {a.driverId}</div>
             )}
           </div>
         ))}
       </div>}

      {showCreate && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onClick={(e) => e.target === e.currentTarget && setShowCreate(false)}>
          <div className="bg-white rounded-xl w-full max-w-md max-h-[90vh] overflow-y-auto p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-bold">طلب التحاق جديد</h3>
              <button onClick={() => setShowCreate(false)} className="w-8 h-8 rounded-full hover:bg-gray-100 flex items-center justify-center">
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={create} className="space-y-3">
              <input required placeholder="الاسم الكامل" value={form.name} onChange={(e) => setForm({...form, name: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 focus:outline-none focus:border-black" />
              <input required type="tel" dir="ltr" placeholder="رقم الهاتف 01xxxxxxxxx" value={form.phone} onChange={(e) => setForm({...form, phone: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 text-right focus:outline-none focus:border-black" />
              <input required placeholder="الرقم القومي" value={form.nationalId} onChange={(e) => setForm({...form, nationalId: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 focus:outline-none focus:border-black" />
              <input required placeholder="رقم الرخصة" value={form.licenseNumber} onChange={(e) => setForm({...form, licenseNumber: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 focus:outline-none focus:border-black" />
              <select value={form.vehicleType} onChange={(e) => setForm({...form, vehicleType: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 bg-white focus:outline-none focus:border-black">
                <option value="motorcycle">دراجة بخارية</option>
                <option value="bicycle">دراجة</option>
                <option value="car">سيارة</option>
              </select>
              <input placeholder="رقم اللوحة" value={form.vehiclePlate} onChange={(e) => setForm({...form, vehiclePlate: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 focus:outline-none focus:border-black" />
              <input placeholder="العنوان" value={form.address} onChange={(e) => setForm({...form, address: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 focus:outline-none focus:border-black" />
              <input type="tel" dir="ltr" placeholder="هاتف طوارئ" value={form.emergencyPhone} onChange={(e) => setForm({...form, emergencyPhone: e.target.value})}
                className="w-full h-11 px-3 rounded-lg border border-gray-200 text-right focus:outline-none focus:border-black" />
              <button type="submit" className="w-full h-11 rounded-lg bg-black text-white font-medium hover:bg-gray-800">إنشاء الطلب</button>
              <p className="text-xs text-gray-500">عند التوثيق، سيُنشأ حساب مندوب بكلمة مرور = الرقم القومي.</p>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
