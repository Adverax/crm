import { http } from './http'
import type {
  LoginRequest,
  TokenPair,
  UserInfo,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  RefreshRequest,
} from '@/types/auth'

interface DataResponse<T> {
  data: T
}

export const authApi = {
  login(req: LoginRequest): Promise<DataResponse<TokenPair>> {
    return http.post('/api/v1/auth/login', req)
  },

  refresh(req: RefreshRequest): Promise<DataResponse<TokenPair>> {
    return http.post('/api/v1/auth/refresh', req)
  },

  logout(refreshToken: string): Promise<void> {
    return http.post('/api/v1/auth/logout', { refreshToken })
  },

  me(): Promise<DataResponse<UserInfo>> {
    return http.get('/api/v1/auth/me')
  },

  forgotPassword(req: ForgotPasswordRequest): Promise<void> {
    return http.post('/api/v1/auth/forgot-password', req)
  },

  resetPassword(req: ResetPasswordRequest): Promise<void> {
    return http.post('/api/v1/auth/reset-password', req)
  },
}
