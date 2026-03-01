import { http } from './http'
import type {
  Portal,
  CreatePortalRequest,
  UpdatePortalRequest,
} from '@/types/portals'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin'

export const portalsApi = {
  list(): Promise<ApiResponse<Portal[]>> {
    return http.get<ApiResponse<Portal[]>>(`${BASE}/portals`)
  },

  get(portalId: string): Promise<ApiResponse<Portal>> {
    return http.get<ApiResponse<Portal>>(`${BASE}/portals/${portalId}`)
  },

  create(data: CreatePortalRequest): Promise<ApiResponse<Portal>> {
    return http.post<ApiResponse<Portal>>(`${BASE}/portals`, data)
  },

  update(portalId: string, data: UpdatePortalRequest): Promise<ApiResponse<Portal>> {
    return http.put<ApiResponse<Portal>>(`${BASE}/portals/${portalId}`, data)
  },

  delete(portalId: string): Promise<void> {
    return http.delete(`${BASE}/portals/${portalId}`)
  },
}
