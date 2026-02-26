import { http } from './http'

const BASE = '/api/v1/admin'

export interface SoqlValidatePayload {
  query: string
}

export interface SoqlValidateError {
  message: string
  line?: number
  column?: number
  code?: string
}

export interface SoqlValidateResponse {
  valid: boolean
  object?: string
  fields?: string[]
  errors?: SoqlValidateError[]
}

export const soqlApi = {
  validate(data: SoqlValidatePayload): Promise<SoqlValidateResponse> {
    return http.post<SoqlValidateResponse>(`${BASE}/soql/validate`, data)
  },
}
