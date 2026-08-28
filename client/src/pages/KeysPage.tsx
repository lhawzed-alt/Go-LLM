import { useEffect, useState } from 'react'
import { api, type AppConfig } from '../api/config'
import { useToast } from '../store/useToast'
import ConfirmDialog from '../components/ConfirmDialog'

export default function KeysPage() {
  const [cfg, setCfg] = useState<AppConfig | null>(null)
  const [newKey, setNewKey] = useState('')
  const [pendingDelete, setPendingDelete] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const { push } = useToast()

  useEffect(() => {
    refresh()
  }, [])

  function refresh() {
    api.getConfig()
      .then(setCfg)
      .catch((e) => push('error', String(e)))
  }

  function add() {
    if (!cfg) return
    const k = newKey.trim()
    if (!k) {
      push('error', '请填写 Key')
      return
    }
    if (cfg.auth.api_keys.includes(k)) {
      push('error', '该 Key 已存在')
      return
    }
    const next: AppConfig = { ...cfg, auth: { api_keys: [...cfg.auth.api_keys, k] } }
    setBusy(true)
    api.putConfig(next)
      .then((res) => {
        setCfg(res.config)
        setNewKey('')
        push('success', '已添加 Key')
      })
      .catch((e) => push('error', String(e)))
      .finally(() => setBusy(false))
  }

  function remove(k: string) {
    if (!cfg) return
    const next: AppConfig = { ...cfg, auth: { api_keys: cfg.auth.api_keys.filter((x) => x !== k) } }
    setBusy(true)
    api.putConfig(next)
      .then((res) => {
        setCfg(res.config)
        setPendingDelete(null)
        push('success', '已删除')
      })
      .catch((e) => push('error', String(e)))
      .finally(() => setBusy(false))
  }

  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="text-2xl font-semibold">客户端 API Key</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        客户端访问本网关时使用的 Bearer Key。修改立即热生效，无需重启。
      </p>

      <div className="card mt-6 p-5">
        <label className="label">添加新 Key</label>
        <div className="flex gap-2">
          <input
            className="input font-mono"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value)}
            placeholder="sk-..."
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                add()
              }
            }}
          />
          <button className="btn-primary" disabled={busy} onClick={add}>
            {busy ? '保存中…' : '添加'}
          </button>
        </div>
        <p className="mt-2 text-xs text-slate-400">
          Key 将以脱敏形式展示，保存时持久化到 config.yaml。生产环境建议为每位使用者分配独立 Key，便于按 Key 吊销。
        </p>
      </div>

      <div className="card mt-6 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50/80 text-left text-xs uppercase tracking-wider text-slate-500 dark:bg-slate-800/40 dark:text-slate-400">
            <tr>
              <th className="px-4 py-2.5">Key（脱敏）</th>
              <th className="px-4 py-2.5 text-right">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-700/60">
            {cfg?.auth.api_keys.map((k) => (
              <tr key={k} className="hover:bg-slate-50/60 dark:hover:bg-slate-800/40">
                <td className="px-4 py-2.5 font-mono text-xs">{k}</td>
                <td className="px-4 py-2.5 text-right">
                  <button
                    className="btn-ghost px-2 py-1 text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-500/10"
                    onClick={() => setPendingDelete(k)}
                  >
                    删除
                  </button>
                </td>
              </tr>
            ))}
            {cfg && cfg.auth.api_keys.length === 0 && (
              <tr>
                <td colSpan={2} className="px-4 py-8 text-center text-slate-400">
                  列表为空时后端不校验客户端 Key（开放模式）
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <ConfirmDialog
        open={!!pendingDelete}
        title="删除 Key？"
        message={`将吊销 "${pendingDelete}"。持有该 Key 的客户端请求将立即返回 401。`}
        confirmText="吊销"
        danger
        onConfirm={() => pendingDelete && remove(pendingDelete)}
        onCancel={() => setPendingDelete(null)}
      />
    </div>
  )
}
