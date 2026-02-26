import { http } from './http'
import type {
  ObjectView,
  CreateObjectViewRequest,
  UpdateObjectViewRequest,
} from '@/types/object-views'
import type { ApiResponse } from '@/types/metadata'

const BASE = '/api/v1/admin'

export const objectViewsApi = {
  list(): Promise<ApiResponse<ObjectView[]>> {
    return http.get<ApiResponse<ObjectView[]>>(`${BASE}/object-views`)
  },

  get(viewId: string): Promise<ApiResponse<ObjectView>> {
    return http.get<ApiResponse<ObjectView>>(`${BASE}/object-views/${viewId}`)
  },

  create(data: CreateObjectViewRequest): Promise<ApiResponse<ObjectView>> {
    return http.post<ApiResponse<ObjectView>>(`${BASE}/object-views`, data)
  },

  update(viewId: string, data: UpdateObjectViewRequest): Promise<ApiResponse<ObjectView>> {
    return http.put<ApiResponse<ObjectView>>(`${BASE}/object-views/${viewId}`, data)
  },

  delete(viewId: string): Promise<void> {
    return http.delete(`${BASE}/object-views/${viewId}`)
  },
}
