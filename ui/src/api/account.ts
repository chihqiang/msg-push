// 账号 API（个人中心：资料与改密码）
import { get, put } from './client'
import type { AccountProfile } from '@/types'

// 当前登录账号资料（GET /account/me）
export const getAccountProfile = () => get<AccountProfile>('/account/me')

// 修改密码（PUT /account/password { old_password, new_password }）
export const changePassword = (payload: { old_password: string; new_password: string }) =>
  put<void>('/account/password', payload)
