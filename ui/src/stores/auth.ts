// 认证状态：登录态以 cookie 为准，页面加载时恢复；后端无 /auth/me 接口，
// 用户名为登录时持久化到 cookie（仅展示用）。
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { post } from '@/api/client'
import type { LoginResponse } from '@/types'
import { appCookie } from '@/plugins/cookie'

const ACCESS_TOKEN_COOKIE = 'msgpush.access_token'
const REFRESH_TOKEN_COOKIE = 'msgpush.refresh_token'
const EXPIRES_AT_COOKIE = 'msgpush.expires_at'
const USERNAME_COOKIE = 'msgpush.username'

const authCookies = appCookie

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(authCookies.get(ACCESS_TOKEN_COOKIE))
  const refreshToken = ref(authCookies.get(REFRESH_TOKEN_COOKIE))
  const expiresAt = ref(authCookies.getNumber(EXPIRES_AT_COOKIE))
  const username = ref(authCookies.get(USERNAME_COOKIE))

  const isLoggedIn = computed(() => !!accessToken.value && Date.now() < expiresAt.value * 1000)

  function writeAuthCookies(opts: {
    accessToken: string
    refreshToken?: string
    expiresAt: number
    username?: string
  }) {
    const maxAge = Math.max(Math.floor((opts.expiresAt - Date.now() / 1000)), 60)
    authCookies.set(ACCESS_TOKEN_COOKIE, opts.accessToken, { maxAge })
    if (opts.refreshToken) {
      authCookies.set(REFRESH_TOKEN_COOKIE, opts.refreshToken, { maxAge: 7 * 24 * 3600 })
    }
    authCookies.set(EXPIRES_AT_COOKIE, String(opts.expiresAt), { maxAge })
    if (opts.username) {
      authCookies.set(USERNAME_COOKIE, opts.username, { maxAge: 7 * 24 * 3600 })
    }
  }

  // 登录（后端：POST /account/auth/login { username, password }）
  async function login(usernameArg: string, password: string) {
    const resp = await post<LoginResponse>('/account/auth/login', { username: usernameArg, password })
    accessToken.value = resp.access_token
    refreshToken.value = resp.refresh_token ?? ''
    expiresAt.value = resp.expires_at
    username.value = usernameArg
    writeAuthCookies({
      accessToken: resp.access_token,
      refreshToken: resp.refresh_token,
      expiresAt: resp.expires_at,
      username: usernameArg,
    })
  }

  // 刷新令牌（后端：POST /account/auth/refresh { refresh_token }）
  async function tryRefresh(): Promise<boolean> {
    if (!refreshToken.value) return false
    try {
      const resp = await post<LoginResponse>('/account/auth/refresh', { refresh_token: refreshToken.value })
      accessToken.value = resp.access_token
      refreshToken.value = resp.refresh_token ?? refreshToken.value
      expiresAt.value = resp.expires_at
      writeAuthCookies({
        accessToken: resp.access_token,
        refreshToken: resp.refresh_token ?? refreshToken.value,
        expiresAt: resp.expires_at,
        username: username.value, // 刷新 username cookie 过期时间，避免 7 天后展示名丢失
      })
      return true
    } catch {
      return false
    }
  }

  // 退出
  function logout() {
    accessToken.value = ''
    refreshToken.value = ''
    expiresAt.value = 0
    authCookies.remove(ACCESS_TOKEN_COOKIE)
    authCookies.remove(REFRESH_TOKEN_COOKIE)
    authCookies.remove(EXPIRES_AT_COOKIE)
    authCookies.remove(USERNAME_COOKIE)
  }

  return {
    accessToken,
    refreshToken,
    expiresAt,
    username,
    isLoggedIn,
    login,
    tryRefresh,
    logout,
  }
})
