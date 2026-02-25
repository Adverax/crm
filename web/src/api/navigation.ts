import { http } from './http'
import type {
  ProfileNavigation,
  CreateProfileNavigationRequest,
  UpdateProfileNavigationRequest,
  ResolvedNavigation,
} from '@/types/navigation'
import type { ApiResponse } from '@/types/metadata'

const ADMIN_BASE = '/api/v1/admin'

export const navigationApi = {
  list(): Promise<ApiResponse<ProfileNavigation[]>> {
    return http.get<ApiResponse<ProfileNavigation[]>>(`${ADMIN_BASE}/profile-navigation`)
  },

  get(id: string): Promise<ApiResponse<ProfileNavigation>> {
    return http.get<ApiResponse<ProfileNavigation>>(`${ADMIN_BASE}/profile-navigation/${id}`)
  },

  create(data: CreateProfileNavigationRequest): Promise<ApiResponse<ProfileNavigation>> {
    return http.post<ApiResponse<ProfileNavigation>>(`${ADMIN_BASE}/profile-navigation`, data)
  },

  update(id: string, data: UpdateProfileNavigationRequest): Promise<ApiResponse<ProfileNavigation>> {
    return http.put<ApiResponse<ProfileNavigation>>(`${ADMIN_BASE}/profile-navigation/${id}`, data)
  },

  delete(id: string): Promise<void> {
    return http.delete(`${ADMIN_BASE}/profile-navigation/${id}`)
  },

  resolve(): Promise<ApiResponse<ResolvedNavigation>> {
    return http.get<ApiResponse<ResolvedNavigation>>('/api/v1/navigation')
  },
}
