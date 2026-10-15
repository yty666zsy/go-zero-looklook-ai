import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 后端是 go-zero 的 4 个 api 服务，这里按「带版本号的路径前缀」直接代理，开发阶段不需要 nginx 网关。
// 注意要用 /xxx/v1 而不是 /xxx —— 否则前端路由 /orders 会被 /order 前缀误代理到后端。
const API = {
  '/travel/v1': 'http://127.0.0.1:1003',
  '/order/v1': 'http://127.0.0.1:1001',
  '/payment/v1': 'http://127.0.0.1:1002',
  '/usercenter/v1': 'http://127.0.0.1:1004',
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
    port: 5173,
    host: true,
    proxy: Object.fromEntries(
      Object.entries(API).map(([prefix, target]) => [
        prefix,
        { target, changeOrigin: true },
      ]),
    ),
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 1200,
  },
})
