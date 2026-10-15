import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 管理端依赖 admin-api（1005）与 aiagent-api（1006），开发阶段直接代理，不需要 nginx。
const API = {
  '/admin/v1': 'http://127.0.0.1:1005',
  '/ai/v1': 'http://127.0.0.1:1006',
}

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5174,
    host: true,
    proxy: Object.fromEntries(
      Object.entries(API).map(([prefix, target]) => [prefix, { target, changeOrigin: true }]),
    ),
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 1200,
  },
})
