import { createContext, useContext, useEffect, useMemo, useState } from 'react'

type Mode = 'light' | 'dark' | 'system'
type Resolved = 'light' | 'dark'

interface Ctx {
  mode: Mode
  resolved: Resolved
  setMode: (m: Mode) => void
}

const ThemeCtx = createContext<Ctx | null>(null)

function getSystem(): Resolved {
  if (typeof window === 'undefined') return 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyTheme(resolved: Resolved) {
  const root = document.documentElement
  if (resolved === 'dark') root.classList.add('dark')
  else root.classList.remove('dark')
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [mode, setModeState] = useState<Mode>(() => {
    if (typeof window === 'undefined') return 'system'
    return (localStorage.getItem('theme-mode') as Mode) || 'system'
  })
  const [system, setSystem] = useState<Resolved>(getSystem())

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = () => setSystem(mq.matches ? 'dark' : 'light')
    mq.addEventListener('change', onChange)
    return () => mq.removeEventListener('change', onChange)
  }, [])

  const resolved: Resolved = mode === 'system' ? system : mode

  useEffect(() => {
    applyTheme(resolved)
  }, [resolved])

  const setMode = (m: Mode) => {
    setModeState(m)
    localStorage.setItem('theme-mode', m)
  }

  const value = useMemo(() => ({ mode, resolved, setMode }), [mode, resolved])
  return <ThemeCtx.Provider value={value}>{children}</ThemeCtx.Provider>
}

export function useTheme() {
  const ctx = useContext(ThemeCtx)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}
