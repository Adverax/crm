import { http } from './http'
import type {
  ProfileDashboard,
  CreateProfileDashboardRequest,
  UpdateProfileDashboardRequest,
  ResolvedDashboard,
} from '@/types/dashboard'
import type { ApiResponse } from '@/types/metadata'

const ADMIN_BASE = '/api/v1/admin'

export const dashboardApi = {
  list(): Promise<ApiResponse<ProfileDashboard[]>> {
    return http.get<ApiResponse<ProfileDashboard[]>>(`${ADMIN_BASE}/profile-dashboards`)
  },

  get(id: string): Promise<ApiResponse<ProfileDashboard>> {
    return http.get<ApiResponse<ProfileDashboard>>(`${ADMIN_BASE}/profile-dashboards/${id}`)
  },

  create(data: CreateProfileDashboardRequest): Promise<ApiResponse<ProfileDashboard>> {
    return http.post<ApiResponse<ProfileDashboard>>(`${ADMIN_BASE}/profile-dashboards`, data)
  },

  update(id: string, data: UpdateProfileDashboardRequest): Promise<ApiResponse<ProfileDashboard>> {
    return http.put<ApiResponse<ProfileDashboard>>(`${ADMIN_BASE}/profile-dashboards/${id}`, data)
  },

  delete(id: string): Promise<void> {
    return http.delete(`${ADMIN_BASE}/profile-dashboards/${id}`)
  },

  resolve(): Promise<ApiResponse<ResolvedDashboard>> {
    return http.get<ApiResponse<ResolvedDashboard>>('/api/v1/dashboard')
  },
}
