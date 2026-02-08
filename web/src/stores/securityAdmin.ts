import { ref } from 'vue'
import { defineStore } from 'pinia'
import { securityApi } from '@/api/security'
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
} from '@/types/security'
import type { PaginationMeta } from '@/types/metadata'

export const useSecurityAdminStore = defineStore('securityAdmin', () => {
  // Roles
  const roles = ref<UserRole[]>([])
  const currentRole = ref<UserRole | null>(null)
  const rolesPagination = ref<PaginationMeta | null>(null)

  // Permission Sets
  const permissionSets = ref<PermissionSet[]>([])
  const currentPermissionSet = ref<PermissionSet | null>(null)
  const permissionSetsPagination = ref<PaginationMeta | null>(null)

  // Profiles
  const profiles = ref<Profile[]>([])
  const currentProfile = ref<Profile | null>(null)
  const profilesPagination = ref<PaginationMeta | null>(null)

  // Users
  const users = ref<User[]>([])
  const currentUser = ref<User | null>(null)
  const usersPagination = ref<PaginationMeta | null>(null)

  // User PS assignments
  const userPermissionSets = ref<PermissionSetAssignment[]>([])

  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // --- Roles ---

  async function fetchRoles(filter?: RoleFilter) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.listRoles(filter)
      roles.value = response.data
      rolesPagination.value = response.pagination
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки ролей'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchRole(roleId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.getRole(roleId)
      currentRole.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки роли'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createRole(data: CreateRoleRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.createRole(data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка создания роли'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateRole(roleId: string, data: UpdateRoleRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.updateRole(roleId, data)
      currentRole.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка обновления роли'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function deleteRole(roleId: string) {
    isLoading.value = true
    error.value = null
    try {
      await securityApi.deleteRole(roleId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка удаления роли'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // --- Permission Sets ---

  async function fetchPermissionSets(filter?: PermissionSetFilter) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.listPermissionSets(filter)
      permissionSets.value = response.data
      permissionSetsPagination.value = response.pagination
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки наборов разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchPermissionSet(psId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.getPermissionSet(psId)
      currentPermissionSet.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки набора разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createPermissionSet(data: CreatePermissionSetRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.createPermissionSet(data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка создания набора разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updatePermissionSet(psId: string, data: UpdatePermissionSetRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.updatePermissionSet(psId, data)
      currentPermissionSet.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка обновления набора разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function deletePermissionSet(psId: string) {
    isLoading.value = true
    error.value = null
    try {
      await securityApi.deletePermissionSet(psId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка удаления набора разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // --- Profiles ---

  async function fetchProfiles(filter?: ProfileFilter) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.listProfiles(filter)
      profiles.value = response.data
      profilesPagination.value = response.pagination
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки профилей'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchProfile(profileId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.getProfile(profileId)
      currentProfile.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки профиля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createProfile(data: CreateProfileRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.createProfile(data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка создания профиля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateProfile(profileId: string, data: UpdateProfileRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.updateProfile(profileId, data)
      currentProfile.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка обновления профиля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function deleteProfile(profileId: string) {
    isLoading.value = true
    error.value = null
    try {
      await securityApi.deleteProfile(profileId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка удаления профиля'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // --- Users ---

  async function fetchUsers(filter?: UserFilter) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.listUsers(filter)
      users.value = response.data
      usersPagination.value = response.pagination
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки пользователей'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchUser(userId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.getUser(userId)
      currentUser.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки пользователя'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createUser(data: CreateUserRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.createUser(data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка создания пользователя'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateUser(userId: string, data: UpdateUserRequest) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.updateUser(userId, data)
      currentUser.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка обновления пользователя'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function deleteUser(userId: string) {
    isLoading.value = true
    error.value = null
    try {
      await securityApi.deleteUser(userId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка удаления пользователя'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // --- User Permission Set Assignments ---

  async function fetchUserPermissionSets(userId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.listUserPermissionSets(userId)
      userPermissionSets.value = response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка загрузки назначенных наборов'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function assignPermissionSet(userId: string, permissionSetId: string) {
    isLoading.value = true
    error.value = null
    try {
      const response = await securityApi.assignPermissionSet(userId, permissionSetId)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка назначения набора разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function revokePermissionSet(userId: string, assignmentId: string) {
    isLoading.value = true
    error.value = null
    try {
      await securityApi.revokePermissionSet(userId, assignmentId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Ошибка отзыва набора разрешений'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    // Roles
    roles,
    currentRole,
    rolesPagination,
    fetchRoles,
    fetchRole,
    createRole,
    updateRole,
    deleteRole,

    // Permission Sets
    permissionSets,
    currentPermissionSet,
    permissionSetsPagination,
    fetchPermissionSets,
    fetchPermissionSet,
    createPermissionSet,
    updatePermissionSet,
    deletePermissionSet,

    // Profiles
    profiles,
    currentProfile,
    profilesPagination,
    fetchProfiles,
    fetchProfile,
    createProfile,
    updateProfile,
    deleteProfile,

    // Users
    users,
    currentUser,
    usersPagination,
    fetchUsers,
    fetchUser,
    createUser,
    updateUser,
    deleteUser,

    // User PS assignments
    userPermissionSets,
    fetchUserPermissionSets,
    assignPermissionSet,
    revokePermissionSet,

    // Shared
    isLoading,
    error,
  }
})
