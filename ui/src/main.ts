import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { VueQueryPlugin } from '@tanstack/vue-query'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import './assets/main.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(VueQueryPlugin)

// 恢复登录态：已持久化 token 且未过期时无需额外请求（无 /auth/me 接口）
const auth = useAuthStore()
if (auth.isLoggedIn && !auth.accessToken) {
  auth.logout()
}

app.mount('#app')
