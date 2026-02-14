import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Derived from OpenAPI spec (single source of truth) ---

export type LoginRequest = CamelCaseKeys<components['schemas']['LoginRequest']>
export type TokenPair = CamelCaseKeys<components['schemas']['TokenPair']>
export type UserInfo = CamelCaseKeys<components['schemas']['UserInfo']>
export type ForgotPasswordRequest = CamelCaseKeys<components['schemas']['ForgotPasswordRequest']>
export type ResetPasswordRequest = CamelCaseKeys<components['schemas']['ResetPasswordRequest']>
export type RefreshRequest = CamelCaseKeys<components['schemas']['RefreshRequest']>
