// 与后端约定：/api 前缀由 Vite 代理到后端 :8080

export interface Channel {
  name: string
  base_url: string
  api_key: string
  models: string[]
  weight: number
}

export interface AppConfig {
  server: { port: number }
  channels: Channel[]
  auth: { api_keys: string[] }
}

export interface Stats {
  cache_hit_rate_percent: number
}

async function http<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    let detail = ''
    try {
      const j = await res.json()
      detail = j?.error || JSON.stringify(j)
    } catch {
      detail = await res.text()
    }
    throw new Error(detail || `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

export const api = {
  getConfig: () => http<AppConfig>('/admin/config'),
  putConfig: (cfg: AppConfig) =>
    http<{ message: string; config: AppConfig }>('/admin/config', {
      method: 'PUT',
      body: JSON.stringify(cfg),
    }),
  listKeysMasked: () => http<{ keys: string[] }>('/admin/keys'),
  health: () => http<{ status: string }>('/healthz').catch(() => ({ status: 'unknown' })) as Promise<{ status: string }>,
  stats: () => http<Stats>('/stats'),
}
