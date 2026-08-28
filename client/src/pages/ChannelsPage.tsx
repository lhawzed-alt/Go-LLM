import { useEffect, useMemo, useState } from 'react'
import { api, type AppConfig, type Channel } from '../api/config'
import { useToast } from '../store/useToast'
import Drawer from '../components/Drawer'
import ConfirmDialog from '../components/ConfirmDialog'

type Mode = { kind: 'create' } | { kind: 'edit'; name: string }

export default function ChannelsPage() {
  const [cfg, setCfg] = useState<AppConfig | null>(null)
  const [loading, setLoading] = useState(false)
  const [mode, setMode] = useState<Mode | null>(null)
  const [pendingDelete, setPendingDelete] = useState<string | null>(null)
  const { push } = useToast()

  useEffect(() => {
    setLoading(true)
    api.getConfig()
      .then(setCfg)
      .catch((e) => push('error', String(e)))
      .finally(() => setLoading(false))
  }, [push])

  const current = useMemo<Channel | null>(() => {
    if (!mode || !cfg) return null
    if (mode.kind === 'create') {
      return { name: '', base_url: 'https://api.openai.com', api_key: '', models: [], weight: 1 }
    }
    return cfg.channels.find((c) => c.name === mode.name) ?? null
  }, [mode, cfg])

  function save(channel: Channel) {
    if (!cfg) return
    if (!channel.name.trim()) {
      push('error', '请填写渠道名')
      return
    }
    const updated: AppConfig = (() => {
      if (mode?.kind === 'create') {
        if (cfg.channels.some((c) => c.name === channel.name)) {
          push('error', `渠道名 "${channel.name}" 已存在`)
          return cfg
        }
        return { ...cfg, channels: [...cfg.channels, channel] }
      }
      if (mode?.kind === 'edit') {
        return {
          ...cfg,
          channels: cfg.channels.map((c) => (c.name === mode.name ? channel : c)),
        }
      }
      return cfg
    })()
    if (updated === cfg) return
    setLoading(true)
    api.putConfig(updated)
      .then((res) => {
        setCfg(res.config)
        setMode(null)
        push('success', mode?.kind === 'create' ? '已新增渠道' : '已保存')
      })
      .catch((e) => push('error', String(e)))
      .finally(() => setLoading(false))
  }

  function remove(name: string) {
    if (!cfg) return
    const next: AppConfig = { ...cfg, channels: cfg.channels.filter((c) => c.name !== name) }
    setLoading(true)
    api.putConfig(next)
      .then((res) => {
        setCfg(res.config)
        setPendingDelete(null)
        push('success', `已删除 ${name}`)
      })
      .catch((e) => push('error', String(e)))
      .finally(() => setLoading(false))
  }

  return (
    <div className="mx-auto max-w-6xl">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">上游渠道</h1>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
            增删改查 OpenAI 兼容上游。保存后立即热更新路由表。
          </p>
        </div>
        <button className="btn-primary" onClick={() => setMode({ kind: 'create' })}>+ 新增渠道</button>
      </div>

      <div className="card overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50/80 text-left text-xs uppercase tracking-wider text-slate-500 dark:bg-slate-800/40 dark:text-slate-400">
            <tr>
              <th className="px-4 py-2.5">名称</th>
              <th className="px-4 py-2.5">Base URL</th>
              <th className="px-4 py-2.5">API Key</th>
              <th className="px-4 py-2.5">模型</th>
              <th className="px-4 py-2.5">权重</th>
              <th className="px-4 py-2.5 text-right">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-700/60">
            {cfg?.channels.map((c) => (
              <tr key={c.name} className="hover:bg-slate-50/60 dark:hover:bg-slate-800/40">
                <td className="px-4 py-2.5 font-medium">{c.name}</td>
                <td className="px-4 py-2.5 font-mono text-xs text-slate-500 dark:text-slate-400">{c.base_url}</td>
                <td className="px-4 py-2.5 font-mono text-xs text-slate-500 dark:text-slate-400">{c.api_key || '—'}</td>
                <td className="px-4 py-2.5">
                  <div className="flex max-w-xs flex-wrap gap-1">
                    {c.models.slice(0, 3).map((m) => (
                      <span key={m} className="tag">{m}</span>
                    ))}
                    {c.models.length > 3 && (
                      <span className="text-xs text-slate-400">+{c.models.length - 3}</span>
                    )}
                  </div>
                </td>
                <td className="px-4 py-2.5">{c.weight}</td>
                <td className="px-4 py-2.5 text-right">
                  <button className="btn-ghost px-2 py-1" onClick={() => setMode({ kind: 'edit', name: c.name })}>编辑</button>
                  <button
                    className="btn-ghost px-2 py-1 text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-500/10"
                    onClick={() => setPendingDelete(c.name)}
                  >
                    删除
                  </button>
                </td>
              </tr>
            ))}
            {cfg && cfg.channels.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-slate-400">
                  暂无渠道，点击右上角新增
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {loading && <div className="mt-2 text-xs text-slate-400">同步中…</div>}

      <Drawer
        open={!!mode}
        title={mode?.kind === 'create' ? '新增渠道' : `编辑：${(mode as { name: string } | null)?.name ?? ''}`}
        onClose={() => setMode(null)}
      >
        {current && (
          <ChannelForm
            channel={current}
            readonlyName={mode?.kind === 'edit'}
            onSubmit={save}
            onCancel={() => setMode(null)}
          />
        )}
      </Drawer>

      <ConfirmDialog
        open={!!pendingDelete}
        title="删除渠道？"
        message={`将删除渠道 "${pendingDelete}"。如果该渠道是某些模型的唯一来源，相关请求将失败。`}
        confirmText="删除"
        danger
        onConfirm={() => pendingDelete && remove(pendingDelete)}
        onCancel={() => setPendingDelete(null)}
      />
    </div>
  )
}

function ChannelForm({
  channel,
  readonlyName,
  onSubmit,
  onCancel,
}: {
  channel: Channel
  readonlyName?: boolean
  onSubmit: (c: Channel) => void
  onCancel: () => void
}) {
  const [form, setForm] = useState<Channel>(channel)
  const [modelInput, setModelInput] = useState('')

  useEffect(() => setForm(channel), [channel])

  function addModel() {
    const v = modelInput.trim()
    if (!v) return
    if (form.models.includes(v)) return
    setForm({ ...form, models: [...form.models, v] })
    setModelInput('')
  }
  function removeModel(m: string) {
    setForm({ ...form, models: form.models.filter((x) => x !== m) })
  }

  const apiKeyIsMasked = form.api_key.includes('****')

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        onSubmit(form)
      }}
      className="space-y-4"
    >
      <div>
        <label className="label">渠道名（唯一标识）</label>
        <input
          className="input"
          value={form.name}
          disabled={readonlyName}
          onChange={(e) => setForm({ ...form, name: e.target.value })}
          placeholder="例如 openai / deepseek"
          required
        />
      </div>
      <div>
        <label className="label">Base URL</label>
        <input
          className="input font-mono"
          value={form.base_url}
          onChange={(e) => setForm({ ...form, base_url: e.target.value })}
          placeholder="https://api.openai.com"
          required
        />
        <p className="mt-1 text-xs text-slate-400">如使用 v1 路径，可写为 https://api.openai.com/v1</p>
      </div>
      <div>
        <label className="label">API Key</label>
        <input
          className="input font-mono"
          value={form.api_key}
          onChange={(e) => setForm({ ...form, api_key: e.target.value })}
          placeholder="sk-..."
        />
        {apiKeyIsMasked && (
          <p className="mt-1 text-xs text-amber-600 dark:text-amber-400">
            显示为脱敏值。如需更换请直接修改，否则保存时保留原值。
          </p>
        )}
      </div>
      <div>
        <label className="label">支持模型</label>
        <div className="flex gap-2">
          <input
            className="input"
            value={modelInput}
            onChange={(e) => setModelInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addModel()
              }
            }}
            placeholder="如 gpt-4o，按回车添加"
          />
          <button type="button" className="btn-ghost" onClick={addModel}>添加</button>
        </div>
        <div className="mt-2 flex flex-wrap gap-1.5">
          {form.models.map((m) => (
            <span key={m} className="tag">
              {m}
              <button type="button" onClick={() => removeModel(m)} className="text-indigo-500 hover:text-rose-500">×</button>
            </span>
          ))}
          {form.models.length === 0 && <span className="text-xs text-slate-400">尚未添加模型</span>}
        </div>
      </div>
      <div>
        <label className="label">权重（多渠道支持同一模型时按权重随机）</label>
        <input
          className="input w-32"
          type="number"
          min={1}
          value={form.weight}
          onChange={(e) => setForm({ ...form, weight: Number(e.target.value) || 1 })}
        />
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <button type="button" className="btn-ghost" onClick={onCancel}>取消</button>
        <button type="submit" className="btn-primary">保存</button>
      </div>
    </form>
  )
}
