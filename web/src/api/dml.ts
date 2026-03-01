import { http } from './http'

const BASE = '/api/v1/admin'

export interface DmlValidatePayload {
  statement: string
}

export interface DmlValidateError {
  message: string
  line?: number
  column?: number
  code?: string
}

export interface DmlValidateResponse {
  valid: boolean
  operation?: string
  object?: string
  fields?: string[]
  sql?: string
  errors?: DmlValidateError[]
}

export interface DmlTestResponse {
  operation: string
  object: string
  rowsAffected: number
  rolledBack: boolean
  error?: string
}

export const dmlApi = {
  validate(data: DmlValidatePayload): Promise<DmlValidateResponse> {
    return http.post<DmlValidateResponse>(`${BASE}/dml/validate`, data)
  },

  test(data: DmlValidatePayload): Promise<DmlTestResponse> {
    return http.post<DmlTestResponse>(`${BASE}/dml/test`, data)
  },
}
