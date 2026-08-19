// Cookie 管理：统一封装 js-cookie，支持子域共享（VITE_AUTH_DOMAIN）
import Cookies from 'js-cookie'

const domain = import.meta.env.VITE_AUTH_DOMAIN || undefined

export const appCookie = {
  get(key: string): string {
    return Cookies.get(key) || ''
  },
  getNumber(key: string): number {
    const v = Cookies.get(key)
    if (!v) return 0
    const n = Number(v)
    return Number.isFinite(n) ? n : 0
  },
  set(key: string, value: string, opts?: { maxAge?: number }) {
    Cookies.set(key, value, {
      domain,
      expires: opts?.maxAge ? opts.maxAge / 86400 : undefined,
      path: '/',
      sameSite: 'lax',
    })
  },
  remove(key: string) {
    Cookies.remove(key, { domain, path: '/' })
  },
}
