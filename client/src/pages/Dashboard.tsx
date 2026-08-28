import { useEffect, useState } from 'react'
import { api, type AppConfig } from '../api/config'

export default function Dashboard() {
  const [cfg, setCfg] = useState<AppConfig | null>(null)
  const [hitRate, setHitRate] = useState<number | null>(null)
  const [err, setErr] = useState<string>('')

  useEffect(() => {
    refresh()
  }, [])

  function refresh() {
    setErr('')
    Promise.all([api.getConfig().catch((e) => { setErr(String(e)); return null }), api.stats().catch(() => null)])
      .then(([c, s]) => {
        if (c) setCfg(c)
        if (s) setHitRate(s.cache_hit_rate_percent)
      })
  }

  return (
    <div className="mx-auto max-w-6xl">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">概览</h1>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
            实时查看中转站运行状态和已加载的配置
          </p>
        </div>
        <button className="btn-ghost" onClick={refresh}>刷新</button>
      </div>

      {err && (
        <div className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-2 text-sm text-rose-700 dark:border-rose-500/40 dark:bg-rose-900/40 dark:text-rose-200">
          {err}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <Card label="上游渠道" value={cfg?.channels.length ?? '-'} />
        <Card label="可用模型" value={cfg ? cfg.channels.reduce((s, c) => s + c.models.length, 0) : '-'} />
        <Card label="客户端 Key" value={cfg?.auth.api_keys.length ?? '-'} />
        <Card label="缓存命中率" value={hitRate != null ? `${hitRate.toFixed(1)}%` : '-'} />
        <Card label="监听端口" value={cfg?.server.port ?? '-'} />
        <Card label="API 地址" value="http://localhost:8080" mono />
      </div>

      <div className="mt-8">
        <h2 className="mb-3 text-sm font-semibold text-slate-500 dark:text-slate-400">渠道速览</h2>
        <div className="card overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-50/80 text-left text-xs uppercase tracking-wider text-slate-500 dark:bg-slate-800/40 dark:text-slate-400">
              <tr>
                <th className="px-4 py-2.5">名称</th>
                <th className="px-4 py-2.5">Base URL</th>
                <th className="px-4 py-2.5">模型数</th>
                <th className="px-4 py-2.5">权重</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-700/60">
              {(cfg?.channels ?? []).map((c) => (
                <tr key={c.name} className="hover:bg-slate-50/60 dark:hover:bg-slate-800/40">
                  <td className="px-4 py-2.5 font-medium">{c.name}</td>
                  <td className="px-4 py-2.5 font-mono text-xs text-slate-500 dark:text-slate-400">{c.base_url}</td>
                  <td className="px-4 py-2.5">{c.models.length}</td>
                  <td className="px-4 py-2.5">{c.weight}</td>
                </tr>
              ))}
              {cfg && cfg.channels.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-6 text-center text-slate-400">暂无渠道</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

function Card({ label, value, mono }: { label: string; value: string | number; mono?: boolean }) {
  return (
    <div className="card p-5">
      <div className="text-xs text-slate-500 dark:text-slate-400">{label}</div>
      <div className={'mt-1.5 text-2xl font-semibold ' + (mono ? 'font-mono text-base' : '')}>{value}</div>
    </div>
  )
}
