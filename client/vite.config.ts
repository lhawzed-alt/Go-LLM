import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// 监听 0.0.0.0:10234，代理 /api -> 后端 :8080
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 10234,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api/, ''),
      },
    },
  },
})
