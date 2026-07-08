
import { useState, useEffect, useRef } from 'react'
import { useRouter, useParams } from '@/lib/navigation'
import { ArrowLeft, Send, Loader2 } from 'lucide-react'
import { useAuth } from '@/store/auth'
import { driverAPI } from '@/lib/api'

export default function SupportTicketPage() {
  const router = useRouter()
  const params = useParams<{ id: string }>()
  const { isAuthenticated } = useAuth()
  const [ticket, setTicket] = useState<any>(null)
  const [messages, setMessages] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [body, setBody] = useState('')
  const [sending, setSending] = useState(false)
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!isAuthenticated) { router.replace('/login'); return }
  }, [isAuthenticated, router])

  const load = () => {
    driverAPI.getTicket(params.id).then((d) => {
      setTicket(d.ticket)
      setMessages(d.messages || [])
    }).finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    const id = setInterval(load, 5000)
    return () => clearInterval(id)
  }, [params.id])

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!body.trim()) return
    setSending(true)
    try {
      await driverAPI.sendMessage(params.id, body)
      setBody('')
      load()
    } catch (err: any) {
      // ignore
    } finally {
      setSending(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-dvh bg-gray-50 flex items-center justify-center" dir="rtl">
        <Loader2 className="w-6 h-6 animate-spin" />
      </div>
    )
  }

  return (
    <div className="min-h-dvh bg-gray-50 flex flex-col" dir="rtl">
      <header className="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 h-14 flex items-center gap-3">
        <button onClick={() => router.back()} className="w-9 h-9 rounded-full hover:bg-gray-100 flex items-center justify-center">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div>
          <h1 className="font-bold text-sm">تذكرة دعم</h1>
          <p className="text-[10px] text-gray-400">
            {ticket?.status === 'open' ? 'مفتوحة' : ticket?.status === 'resolved' ? 'تم الحل' : 'مغلقة'}
          </p>
        </div>
      </header>

      <div className="flex-1 container mx-auto px-4 py-4 max-w-md overflow-y-auto pb-4">
        {/* Ticket info */}
        {ticket && (
          <div className="bg-white rounded-lg border border-gray-200 p-3 mb-4">
            <p className="text-xs text-gray-500 mb-1">{ticket.type === 'cancellation_request' ? 'طلب إلغاء' : ticket.type}</p>
            <p className="text-sm">{ticket.reason}</p>
          </div>
        )}

        {/* Messages */}
        <div className="space-y-2 mb-4">
          {messages.map((m) => (
            <div
              key={m.id}
              className={`flex ${m.sender === 'driver' ? 'justify-start' : 'justify-end'}`}
            >
              <div
                className={`max-w-[75%] rounded-2xl px-3 py-2 text-sm ${
                  m.sender === 'driver'
                    ? 'bg-black text-white rounded-tl-md'
                    : 'bg-white border border-gray-200 rounded-tr-md'
                }`}
              >
                <p>{m.body}</p>
                <p className={`text-[9px] mt-1 ${m.sender === 'driver' ? 'text-gray-300' : 'text-gray-400'}`}>
                  {new Date(m.createdAt).toLocaleTimeString('ar-EG', { hour: '2-digit', minute: '2-digit' })}
                </p>
              </div>
            </div>
          ))}
          <div ref={endRef} />
        </div>
      </div>

      {/* Input */}
      {ticket?.status !== 'resolved' && (
        <form
          onSubmit={handleSend}
          className="bg-white border-t border-gray-200 p-3 flex items-center gap-2"
          style={{ paddingBottom: 'calc(0.75rem + env(safe-area-inset-bottom, 0px))' }}
        >
          <input
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="اكتب رسالة..."
            className="flex-1 h-11 px-4 rounded-full border border-gray-200 focus:outline-none focus:border-black"
          />
          <button
            type="submit"
            disabled={sending || !body.trim()}
            className="w-11 h-11 rounded-full bg-black text-white flex items-center justify-center disabled:opacity-50"
          >
            {sending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
          </button>
        </form>
      )}
    </div>
  )
}
