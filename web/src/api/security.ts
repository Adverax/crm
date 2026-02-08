import { http } from './http'
import type {
  UserRole,
  CreateRoleRequest,
  UpdateRoleRequest,
  RoleFilter,
  PermissionSet,
  CreatePermissionSetRequest,
  UpdatePermissionSetRequest,
  PermissionSetFilter,
  Profile,
  CreateProfileRequest,
  UpdateProfileRequest,
  ProfileFilter,
  User,
  CreateUserRequest,
  UpdateUserRequest,
  UserFilter,
  PermissionSetAssignment,
  ObjectPermission,
  SetObjectPermissionRequest,
  FieldPermission,
  SetFieldPermissionRequest,
} from '@/types/security'
import type { ApiResponse, ApiListResponse } from '@/types/metadata'

const BASE = '/api/v1/admin/security'

export const securityApi = {
  // Roles
  listRoles(filter?: RoleFilter): Promise<ApiListResponse<UserRole>> {
    const params: Record<string, string | number | undefined> = {}
    if (filter?.page) params['page'] = filter.page
    if (filter?.perPage) params['per_page'] = filter.perPage
    return http.get<ApiListResponse<UserRole>>(`${BASE}/roles`, params)
  },

  getRole(roleId: string): Promise<ApiResponse<UserRole>> {
    return http.get<ApiResponse<UserRole>>(`${BASE}/roles/${roleId}`)
  },

  createRole(data: CreateRoleRequest): Promise<ApiResponse<UserRole>> {
    return http.post<ApiResponse<UserRole>>(`${BASE}/roles`, data)
  },

  updateRole(roleId: string, data: UpdateRoleRequest): Promise<ApiResponse<UserRole>> {
    return http.put<ApiResponse<UserRole>>(`${BASE}/roles/${roleId}`, data)
  },

  deleteRole(roleId: string): Promise<void> {
    return http.delete(`${BASE}/roles/${roleId}`)
  },

  // Permission Sets
  listPermissionSets(filter?: PermissionSetFilter): Promise<ApiListResponse<PermissionSet>> {
    const params: Record<string, string | number | undefined> = {}
    if (filter?.page) params['page'] = filter.page
    if (filter?.perPage) params['per_page'] = filter.perPage
    if (filter?.psType) params['ps_type'] = filter.psType
    return http.get<ApiListResponse<PermissionSet>>(`${BASE}/permission-sets`, params)
  },

  getPermissionSet(psId: string): Promise<ApiResponse<PermissionSet>> {
    return http.get<ApiResponse<PermissionSet>>(`${BASE}/permission-sets/${psId}`)
  },

  createPermissionSet(data: CreatePermissionSetRequest): Promise<ApiResponse<PermissionSet>> {
    return http.post<ApiResponse<PermissionSet>>(`${BASE}/permission-sets`, data)
  },

  updatePermissionSet(psId: string, data: UpdatePermissionSetRequest): Promise<ApiResponse<PermissionSet>> {
    return http.put<ApiResponse<PermissionSet>>(`${BASE}/permission-sets/${psId}`, data)
  },

  deletePermissionSet(psId: string): Promise<void> {
    return http.delete(`${BASE}/permission-sets/${psId}`)
  },

  // Profiles
  listProfiles(filter?: ProfileFilter): Promise<ApiListResponse<Profile>> {
    const params: Record<string, string | number | undefined> = {}
    if (filter?.page) params['page'] = filter.page
    if (filter?.perPage) params['per_page'] = filter.perPage
    return http.get<ApiListResponse<Profile>>(`${BASE}/profiles`, params)
  },

  getProfile(profileId: string): Promise<ApiResponse<Profile>> {
    return http.get<ApiResponse<Profile>>(`${BASE}/profiles/${profileId}`)
  },

  createProfile(data: CreateProfileRequest): Promise<ApiResponse<Profile>> {
    return http.post<ApiResponse<Profile>>(`${BASE}/profiles`, data)
  },

  updateProfile(profileId: string, data: UpdateProfileRequest): Promise<ApiResponse<Profile>> {
    return http.put<ApiResponse<Profile>>(`${BASE}/profiles/${profileId}`, data)
  },

  deleteProfile(profileId: string): Promise<void> {
    return http.delete(`${BASE}/profiles/${profileId}`)
  },

  // Users
  listUsers(filter?: UserFilter): Promise<ApiListResponse<User>> {
    const params: Record<string, string | number | undefined> = {}
    if (filter?.page) params['page'] = filter.page
    if (filter?.perPage) params['per_page'] = filter.perPage
    if (filter?.isActive !== undefined) params['is_active'] = filter.isActive ? 1 : 0
    return http.get<ApiListResponse<User>>(`${BASE}/users`, params)
  },

  getUser(userId: string): Promise<ApiResponse<User>> {
    return http.get<ApiResponse<User>>(`${BASE}/users/${userId}`)
  },

  createUser(data: CreateUserRequest): Promise<ApiResponse<User>> {
    return http.post<ApiResponse<User>>(`${BASE}/users`, data)
  },

  updateUser(userId: string, data: UpdateUserRequest): Promise<ApiResponse<User>> {
    return http.put<ApiResponse<User>>(`${BASE}/users/${userId}`, data)
  },

  deleteUser(userId: string): Promise<void> {
    return http.delete(`${BASE}/users/${userId}`)
  },

  // User Permission Set Assignments
  listUserPermissionSets(userId: string): Promise<ApiListResponse<PermissionSetAssignment>> {
    return http.get<ApiListResponse<PermissionSetAssignment>>(`${BASE}/users/${userId}/permission-sets`)
  },

  assignPermissionSet(userId: string, permissionSetId: string): Promise<ApiResponse<PermissionSetAssignment>> {
    return http.post<ApiResponse<PermissionSetAssignment>>(
      `${BASE}/users/${userId}/permission-sets`,
      { permissionSetId },
    )
  },

  revokePermissionSet(userId: string, assignmentId: string): Promise<void> {
    return http.delete(`${BASE}/users/${userId}/permission-sets/${assignmentId}`)
  },

  // Object Permissions (OLS)
  listObjectPermissions(psId: string): Promise<ApiListResponse<ObjectPermission>> {
    return http.get<ApiListResponse<ObjectPermission>>(`${BASE}/permission-sets/${psId}/object-permissions`)
  },

  setObjectPermission(psId: string, data: SetObjectPermissionRequest): Promise<ApiResponse<ObjectPermission>> {
    return http.put<ApiResponse<ObjectPermission>>(
      `${BASE}/permission-sets/${psId}/object-permissions`,
      data,
    )
  },

  deleteObjectPermission(psId: string, objectPermissionId: string): Promise<void> {
    return http.delete(`${BASE}/permission-sets/${psId}/object-permissions/${objectPermissionId}`)
  },

  // Field Permissions (FLS)
  listFieldPermissions(psId: string, objectId: string): Promise<ApiListResponse<FieldPermission>> {
    return http.get<ApiListResponse<FieldPermission>>(
      `${BASE}/permission-sets/${psId}/field-permissions`,
      { object_id: objectId },
    )
  },

  setFieldPermission(psId: string, data: SetFieldPermissionRequest): Promise<ApiResponse<FieldPermission>> {
    return http.put<ApiResponse<FieldPermission>>(
      `${BASE}/permission-sets/${psId}/field-permissions`,
      data,
    )
  },

  deleteFieldPermission(psId: string, fieldPermissionId: string): Promise<void> {
    return http.delete(`${BASE}/permission-sets/${psId}/field-permissions/${fieldPermissionId}`)
  },
}
