import { http } from './http'
import type {
  Layout,
  CreateLayoutRequest,
  UpdateLayoutRequest,
} from '@/types/layouts'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin'

export const layoutsApi = {
  list(portalId?: string): Promise<ApiResponse<Layout[]>> {
    const params: Record<string, string | undefined> = {}
    if (portalId) {
      params.portal_id = portalId
    }
    return http.get<ApiResponse<Layout[]>>(`${BASE}/layouts`, params)
  },

  get(layoutId: string): Promise<ApiResponse<Layout>> {
    return http.get<ApiResponse<Layout>>(`${BASE}/layouts/${layoutId}`)
  },

  create(data: CreateLayoutRequest): Promise<ApiResponse<Layout>> {
    return http.post<ApiResponse<Layout>>(`${BASE}/layouts`, data)
  },

  update(layoutId: string, data: UpdateLayoutRequest): Promise<ApiResponse<Layout>> {
    return http.put<ApiResponse<Layout>>(`${BASE}/layouts/${layoutId}`, data)
  },

  delete(layoutId: string): Promise<void> {
    return http.delete(`${BASE}/layouts/${layoutId}`)
  },
}
