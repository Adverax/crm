import { http } from './http'
import type { ObjectView } from '@/types/object-views'
import type { ApiResponse } from '@/types/metadata'

export const viewsApi = {
  getByAPIName(ovApiName: string): Promise<ApiResponse<ObjectView>> {
    return http.get<ApiResponse<ObjectView>>(`/api/v1/view/${ovApiName}`)
  },
}
