import { http } from './http'
import type {
  Function,
  CreateFunctionRequest,
  UpdateFunctionRequest,
} from '@/types/functions'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin'

export const functionsApi = {
  list(): Promise<ApiResponse<Function[]>> {
    return http.get<ApiResponse<Function[]>>(`${BASE}/functions`)
  },

  get(functionId: string): Promise<ApiResponse<Function>> {
    return http.get<ApiResponse<Function>>(`${BASE}/functions/${functionId}`)
  },

  create(data: CreateFunctionRequest): Promise<ApiResponse<Function>> {
    return http.post<ApiResponse<Function>>(`${BASE}/functions`, data)
  },

  update(functionId: string, data: UpdateFunctionRequest): Promise<ApiResponse<Function>> {
    return http.put<ApiResponse<Function>>(`${BASE}/functions/${functionId}`, data)
  },

  delete(functionId: string): Promise<void> {
    return http.delete(`${BASE}/functions/${functionId}`)
  },
}
