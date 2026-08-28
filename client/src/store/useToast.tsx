import { createContext, useCallback, useContext, useState } from 'react'

export type ToastKind = 'success' | 'error' | 'info'
export interface ToastItem {
  id: number
  kind: ToastKind
  message: string
}

interface Ctx {
  toasts: ToastItem[]
  push: (kind: ToastKind, message: string) => void
  dismiss: (id: number) => void
}

const ToastCtx = createContext<Ctx | null>(null)

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([])
  const dismiss = useCallback((id: number) => {
    setToasts((arr) => arr.filter((t) => t.id !== id))
  }, [])
  const push = useCallback(
    (kind: ToastKind, message: string) => {
      const id = Date.now() + Math.random()
      setToasts((arr) => [...arr, { id, kind, message }])
      setTimeout(() => dismiss(id), 3500)
    },
    [dismiss],
  )
  return (
    <ToastCtx.Provider value={{ toasts, push, dismiss }}>{children}</ToastCtx.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastCtx)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}
