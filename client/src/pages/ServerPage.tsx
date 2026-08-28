import { useEffect, useState } from 'react'
import { api, type AppConfig } from '../api/config'
import { useToast } from '../store/useToast'

export default function ServerPage() {
  const [cfg, setCfg] = useState<AppConfig | null>(null)
  const [port, setPort] = useState<number>(8080)
  const [saving, setSaving] = useState(false)
  const { push } = useToast()

  useEffect(() => {
    api.getConfig().then((c) => {
      setCfg(c)
      setPort(c.server.port)
    }).catch((e) => push('error', String(e)))
  }, [push])

  function save() {
    if (!cfg) return
    if (port < 1 || port > 65535) {
      push('error', '端口必须在 1-65535 之间')
      return
    }
    setSaving(true)
    api.putConfig({ ...cfg, server: { port } })
      .then(() => push('success', '已保存。下次重启后端生效（新端口：' + port + '）'))
      .catch((e) => push('error', String(e)))
      .finally(() => setSaving(false))
  }

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="text-2xl font-semibold">服务设置</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">管理后端服务监听端口</p>

      <div className="card mt-6 p-5">
        <label className="label">监听端口</label>
        <div className="flex items-center gap-2">
          <input
            className="input w-40"
            type="number"
            min={1}
            max={65535}
            value={port}
            onChange={(e) => setPort(Number(e.target.value))}
          />
          <span className="text-sm text-slate-500 dark:text-slate-400">后端将监听 :{port}</span>
        </div>
        <p className="mt-3 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-800 dark:bg-amber-500/10 dark:text-amber-200">
          ⚠️ 端口变更需要重启后端进程后才能生效。配置文件已更新（config.yaml.bak 为旧版）。
        </p>
      </div>

      <div className="mt-5 flex gap-2">
        <button className="btn-primary" disabled={saving} onClick={save}>
          {saving ? '保存中…' : '保存'}
        </button>
      </div>
    </div>
  )
}
