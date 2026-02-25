export type CredentialType = 'api_key' | 'basic' | 'oauth2_client'

export interface Credential {
  id: string
  code: string
  name: string
  description: string
  type: CredentialType
  baseUrl: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface ApiKeyAuthData {
  header: string
  value: string
}

export interface BasicAuthData {
  username: string
  password: string
}

export interface OAuth2ClientAuthData {
  clientId: string
  clientSecret: string
  tokenUrl: string
  scope?: string
}

export type AuthData = ApiKeyAuthData | BasicAuthData | OAuth2ClientAuthData

export interface CreateCredentialRequest {
  code: string
  name: string
  description?: string
  type: CredentialType
  baseUrl: string
  authData: AuthData
}

export interface UpdateCredentialRequest {
  name: string
  description?: string
  baseUrl: string
  authData?: AuthData
}

export interface UsageLogEntry {
  id: string
  credentialId: string
  procedureCode: string
  requestUrl: string
  responseStatus: number | null
  success: boolean
  errorMessage: string
  durationMs: number
  userId: string | null
  createdAt: string
}
