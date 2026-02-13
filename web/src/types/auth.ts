export interface LoginRequest {
  username: string
  password: string
}

export interface TokenPair {
  accessToken: string
  refreshToken: string
}

export interface UserInfo {
  id: string
  username: string
  email: string
  firstName: string
  lastName: string
  profileId: string
  roleId: string | null
  isActive: boolean
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  password: string
}

export interface RefreshRequest {
  refreshToken: string
}
