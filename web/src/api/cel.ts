import { http } from './http'
import type { CelValidateResponse, CelContext, CelParamDef } from '@/types/functions'

const BASE = '/api/v1/admin'

export interface CelValidatePayload {
  expression: string
  context: CelContext
  objectApiName?: string
  params?: CelParamDef[]
}

export const celApi = {
  validate(data: CelValidatePayload): Promise<CelValidateResponse> {
    return http.post<CelValidateResponse>(`${BASE}/cel/validate`, data)
  },
}
