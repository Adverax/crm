import { http } from './http'
import type {
  SharedLayout,
  CreateSharedLayoutRequest,
  UpdateSharedLayoutRequest,
} from '@/types/layouts'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin'

export const sharedLayoutsApi = {
  list(): Promise<ApiResponse<SharedLayout[]>> {
    return http.get<ApiResponse<SharedLayout[]>>(`${BASE}/shared-layouts`)
  },

  get(id: string): Promise<ApiResponse<SharedLayout>> {
    return http.get<ApiResponse<SharedLayout>>(`${BASE}/shared-layouts/${id}`)
  },

  create(data: CreateSharedLayoutRequest): Promise<ApiResponse<SharedLayout>> {
    return http.post<ApiResponse<SharedLayout>>(`${BASE}/shared-layouts`, data)
  },

  update(id: string, data: UpdateSharedLayoutRequest): Promise<ApiResponse<SharedLayout>> {
    return http.put<ApiResponse<SharedLayout>>(`${BASE}/shared-layouts/${id}`, data)
  },

  delete(id: string): Promise<void> {
    return http.delete(`${BASE}/shared-layouts/${id}`)
  },
}
