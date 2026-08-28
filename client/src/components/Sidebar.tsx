import { NavLink } from 'react-router-dom'

const items = [
  { to: '/dashboard', label: '概览', icon: '📊' },
  { to: '/server', label: '服务设置', icon: '⚙️' },
  { to: '/channels', label: '上游渠道', icon: '🔌' },
  { to: '/keys', label: '客户端 Key', icon: '🔑' },
]

export default function Sidebar() {
  return (
    <aside className="hidden w-60 shrink-0 flex-col border-r border-slate-200/60 bg-white/60 px-4 py-6 backdrop-blur dark:border-slate-800/60 dark:bg-slate-900/50 md:flex">
      <div className="mb-8 flex items-center gap-2 px-2">
        <div className="grid h-8 w-8 place-items-center rounded-lg bg-gradient-to-br from-indigo-500 to-emerald-500 text-white">
          <span className="font-bold">G</span>
        </div>
        <div>
          <div className="text-sm font-semibold">Go LLM</div>
          <div className="text-[11px] text-slate-500 dark:text-slate-400">控制台</div>
        </div>
      </div>
      <nav className="flex flex-col gap-1">
        {items.map((it) => (
          <NavLink
            key={it.to}
            to={it.to}
            className={({ isActive }) =>
              'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ' +
              (isActive
                ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300'
                : 'text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800/60')
            }
          >
            <span>{it.icon}</span>
            <span>{it.label}</span>
          </NavLink>
        ))}
      </nav>
      <div className="mt-auto px-2 pt-6 text-[11px] leading-relaxed text-slate-400 dark:text-slate-500">
        监听 <span className="font-mono">0.0.0.0:10234</span>
        <br />
        修改将持久化到 <span className="font-mono">config.yaml</span>
      </div>
    </aside>
  )
}
