import { http } from './http'
import type { Portal, GateResponse } from '@/types/portals'
import type { ApiResponse } from '@/types/metadata'

export const viewsApi = {
  getByAPIName(portalApiName: string): Promise<ApiResponse<Portal>> {
    return http.get<ApiResponse<Portal>>(`/api/v1/portal/${portalApiName}`)
  },

  executeGate(
    portalApiName: string,
    gateName: string,
    params?: Record<string, string>,
  ): Promise<ApiResponse<GateResponse>> {
    const query = params ? '?' + new URLSearchParams(params).toString() : ''
    return http.get<ApiResponse<GateResponse>>(
      `/api/v1/portal/${portalApiName}/gate/${gateName}${query}`,
    )
  },

  executeGatePost(
    portalApiName: string,
    gateName: string,
    args?: Record<string, unknown>,
    data?: Record<string, unknown>,
  ): Promise<ApiResponse<GateResponse>> {
    return http.post<ApiResponse<GateResponse>>(
      `/api/v1/portal/${portalApiName}/gate/${gateName}`,
      { args, data },
    )
  },
}
