import { useEffect } from 'react'

interface Props {
  open: boolean
  title: string
  onClose: () => void
  children: React.ReactNode
  footer?: React.ReactNode
}

export default function Drawer({ open, title, onClose, children, footer }: Props) {
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open) return null
  return (
    <div className="fixed inset-0 z-30 flex">
      <div className="flex-1 bg-slate-900/40 backdrop-blur-sm" onClick={onClose} />
      <div className="flex h-full w-full max-w-lg flex-col border-l border-slate-200/60 bg-white shadow-2xl dark:border-slate-700/60 dark:bg-slate-900">
        <div className="flex items-center justify-between border-b border-slate-200/60 px-5 py-3 dark:border-slate-700/60">
          <h3 className="text-sm font-semibold">{title}</h3>
          <button onClick={onClose} className="btn-ghost px-2 py-1">✕</button>
        </div>
        <div className="flex-1 overflow-y-auto px-5 py-4">{children}</div>
        {footer && (
          <div className="border-t border-slate-200/60 px-5 py-3 dark:border-slate-700/60">
            {footer}
          </div>
        )}
      </div>
    </div>
  )
}
