import { http } from './http'
import type {
  Credential,
  CreateCredentialRequest,
  UpdateCredentialRequest,
  UsageLogEntry,
} from '@/types/credentials'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin/credentials'

export const credentialsApi = {
  list(): Promise<ApiResponse<Credential[]>> {
    return http.get<ApiResponse<Credential[]>>(BASE)
  },

  get(id: string): Promise<ApiResponse<Credential>> {
    return http.get<ApiResponse<Credential>>(`${BASE}/${id}`)
  },

  create(data: CreateCredentialRequest): Promise<ApiResponse<Credential>> {
    return http.post<ApiResponse<Credential>>(BASE, data)
  },

  update(id: string, data: UpdateCredentialRequest): Promise<ApiResponse<Credential>> {
    return http.put<ApiResponse<Credential>>(`${BASE}/${id}`, data)
  },

  delete(id: string): Promise<void> {
    return http.delete(`${BASE}/${id}`)
  },

  testConnection(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return http.post<ApiResponse<{ success: boolean }>>(`${BASE}/${id}/test`, {})
  },

  getUsageLog(id: string, limit?: number): Promise<ApiResponse<UsageLogEntry[]>> {
    const params: Record<string, string | number | undefined> = {}
    if (limit) params['limit'] = limit
    return http.get<ApiResponse<UsageLogEntry[]>>(`${BASE}/${id}/usage`, params)
  },

  deactivate(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return http.post<ApiResponse<{ success: boolean }>>(`${BASE}/${id}/deactivate`, {})
  },

  activate(id: string): Promise<ApiResponse<{ success: boolean }>> {
    return http.post<ApiResponse<{ success: boolean }>>(`${BASE}/${id}/activate`, {})
  },
}
