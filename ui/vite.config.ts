import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  // 开发环境代理目标：由 VITE_API_PROXY_TARGET 提供，缺省 8080（见后端 config.yaml）
  const proxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080'

  return {
    plugins: [vue(), tailwindcss()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 5173,
      proxy: {
        // 开发环境代理：前端 /api 请求转发到后端 msg-push
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
