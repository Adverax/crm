import { http } from './http'
import type { Portal } from '@/types/portals'
import type { ApiResponse } from '@/types/metadata'

export const viewsApi = {
  getByAPIName(portalApiName: string): Promise<ApiResponse<Portal>> {
    return http.get<ApiResponse<Portal>>(`/api/v1/portal/${portalApiName}`)
  },
}
