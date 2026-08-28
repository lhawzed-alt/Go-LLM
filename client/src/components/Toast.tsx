import { useToast } from '../store/useToast'

export default function Toast() {
  const { toasts, dismiss } = useToast()
  return (
    <div className="pointer-events-none fixed right-4 top-4 z-50 flex w-80 flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          onClick={() => dismiss(t.id)}
          className={
            'pointer-events-auto cursor-pointer rounded-lg border px-4 py-3 text-sm shadow-lg backdrop-blur transition-all ' +
            (t.kind === 'success'
              ? 'border-emerald-200 bg-emerald-50/90 text-emerald-800 dark:border-emerald-500/40 dark:bg-emerald-900/60 dark:text-emerald-100'
              : t.kind === 'error'
              ? 'border-rose-200 bg-rose-50/90 text-rose-800 dark:border-rose-500/40 dark:bg-rose-900/60 dark:text-rose-100'
              : 'border-slate-200 bg-white/95 text-slate-700 dark:border-slate-700 dark:bg-slate-800/90 dark:text-slate-200')
          }
        >
          {t.message}
        </div>
      ))}
    </div>
  )
}
