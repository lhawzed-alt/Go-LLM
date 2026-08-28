import { NavLink, Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import Toast from './Toast'
import { useTheme } from '../store/useTheme'

export default function Layout() {
  const { mode, setMode, resolved } = useTheme()
  return (
    <div className="flex h-full">
      <Sidebar />
      <main className="flex-1 overflow-y-auto">
        <header className="sticky top-0 z-20 flex items-center justify-between border-b border-slate-200/60 bg-white/70 px-6 py-3 backdrop-blur dark:border-slate-700/60 dark:bg-slate-900/70">
          <div className="text-sm text-slate-500 dark:text-slate-400">
            <NavLink to="/dashboard" className="hover:text-indigo-600">控制台</NavLink>
            <span className="mx-1.5">/</span>
            <NavLink to="/channels" className="hover:text-indigo-600">渠道</NavLink>
            <span className="mx-1.5">/</span>
            <NavLink to="/keys" className="hover:text-indigo-600">Key</NavLink>
            <span className="mx-1.5">/</span>
            <NavLink to="/server" className="hover:text-indigo-600">服务</NavLink>
          </div>
          <div className="flex items-center gap-1 rounded-lg bg-slate-100 p-0.5 text-xs dark:bg-slate-800">
            {(['light', 'system', 'dark'] as const).map((m) => (
              <button
                key={m}
                onClick={() => setMode(m)}
                className={
                  'rounded-md px-2.5 py-1 transition-colors ' +
                  (mode === m
                    ? 'bg-white text-slate-900 shadow dark:bg-slate-700 dark:text-white'
                    : 'text-slate-500 hover:text-slate-900 dark:hover:text-white')
                }
              >
                {m === 'light' ? '浅色' : m === 'dark' ? '深色' : '跟随系统'}
                {m === 'system' && (
                  <span className="ml-1 text-[10px] text-slate-400">({resolved === 'dark' ? '暗' : '亮'})</span>
                )}
              </button>
            ))}
          </div>
        </header>
        <div className="px-6 py-6">
          <Outlet />
        </div>
      </main>
      <Toast />
    </div>
  )
}
