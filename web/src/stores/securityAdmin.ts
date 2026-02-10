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
  Group,
  CreateGroupRequest,
  GroupFilter,
  GroupMember,
  AddGroupMemberRequest,
  SharingRule,
  CreateSharingRuleRequest,
  UpdateSharingRuleRequest,
  SharingRuleFilter,
} from '@/types/security'
import type { PaginationMeta } from '@/types/metadata'

export const useSecurityAdminStore = defineStore('securityAdmin', () => {
  // Roles
  const roles = ref<UserRole[]>([])
  const currentRole = ref<UserRole | null>(null)
  const rolesPagination = ref<PaginationMeta | null>(null)
  const rolesLoading = ref(false)
  const rolesError = ref<string | null>(null)

  // Permission Sets
  const permissionSets = ref<PermissionSet[]>([])
  const currentPermissionSet = ref<PermissionSet | null>(null)
  const permissionSetsPagination = ref<PaginationMeta | null>(null)
  const permissionSetsLoading = ref(false)
  const permissionSetsError = ref<string | null>(null)

  // Profiles
  const profiles = ref<Profile[]>([])
  const currentProfile = ref<Profile | null>(null)
  const profilesPagination = ref<PaginationMeta | null>(null)
  const profilesLoading = ref(false)
  const profilesError = ref<string | null>(null)

  // Users
  const users = ref<User[]>([])
  const currentUser = ref<User | null>(null)
  const usersPagination = ref<PaginationMeta | null>(null)
  const usersLoading = ref(false)
  const usersError = ref<string | null>(null)

  // User PS assignments
  const userPermissionSets = ref<PermissionSetAssignment[]>([])

  // --- Roles ---

  async function fetchRoles(filter?: RoleFilter) {
    rolesLoading.value = true
    rolesError.value = null
    try {
      const response = await securityApi.listRoles(filter)
      roles.value = response.data ?? []
      rolesPagination.value = response.pagination
    } catch (err) {
      rolesError.value = err instanceof Error ? err.message : 'Ошибка загрузки ролей'
      throw err
    } finally {
      rolesLoading.value = false
    }
  }

  async function fetchRole(roleId: string) {
    rolesLoading.value = true
    rolesError.value = null
    try {
      const response = await securityApi.getRole(roleId)
      currentRole.value = response.data
      return response.data
    } catch (err) {
      rolesError.value = err instanceof Error ? err.message : 'Ошибка загрузки роли'
      throw err
    } finally {
      rolesLoading.value = false
    }
  }

  async function createRole(data: CreateRoleRequest) {
    rolesLoading.value = true
    rolesError.value = null
    try {
      const response = await securityApi.createRole(data)
      return response.data
    } catch (err) {
      rolesError.value = err instanceof Error ? err.message : 'Ошибка создания роли'
      throw err
    } finally {
      rolesLoading.value = false
    }
  }

  async function updateRole(roleId: string, data: UpdateRoleRequest) {
    rolesLoading.value = true
    rolesError.value = null
    try {
      const response = await securityApi.updateRole(roleId, data)
      currentRole.value = response.data
      return response.data
    } catch (err) {
      rolesError.value = err instanceof Error ? err.message : 'Ошибка обновления роли'
      throw err
    } finally {
      rolesLoading.value = false
    }
  }

  async function deleteRole(roleId: string) {
    rolesLoading.value = true
    rolesError.value = null
    try {
      await securityApi.deleteRole(roleId)
    } catch (err) {
      rolesError.value = err instanceof Error ? err.message : 'Ошибка удаления роли'
      throw err
    } finally {
      rolesLoading.value = false
    }
  }

  // --- Permission Sets ---

  async function fetchPermissionSets(filter?: PermissionSetFilter) {
    permissionSetsLoading.value = true
    permissionSetsError.value = null
    try {
      const response = await securityApi.listPermissionSets(filter)
      permissionSets.value = response.data ?? []
      permissionSetsPagination.value = response.pagination
    } catch (err) {
      permissionSetsError.value = err instanceof Error ? err.message : 'Ошибка загрузки наборов разрешений'
      throw err
    } finally {
      permissionSetsLoading.value = false
    }
  }

  async function fetchPermissionSet(psId: string) {
    permissionSetsLoading.value = true
    permissionSetsError.value = null
    try {
      const response = await securityApi.getPermissionSet(psId)
      currentPermissionSet.value = response.data
      return response.data
    } catch (err) {
      permissionSetsError.value = err instanceof Error ? err.message : 'Ошибка загрузки набора разрешений'
      throw err
    } finally {
      permissionSetsLoading.value = false
    }
  }

  async function createPermissionSet(data: CreatePermissionSetRequest) {
    permissionSetsLoading.value = true
    permissionSetsError.value = null
    try {
      const response = await securityApi.createPermissionSet(data)
      return response.data
    } catch (err) {
      permissionSetsError.value = err instanceof Error ? err.message : 'Ошибка создания набора разрешений'
      throw err
    } finally {
      permissionSetsLoading.value = false
    }
  }

  async function updatePermissionSet(psId: string, data: UpdatePermissionSetRequest) {
    permissionSetsLoading.value = true
    permissionSetsError.value = null
    try {
      const response = await securityApi.updatePermissionSet(psId, data)
      currentPermissionSet.value = response.data
      return response.data
    } catch (err) {
      permissionSetsError.value = err instanceof Error ? err.message : 'Ошибка обновления набора разрешений'
      throw err
    } finally {
      permissionSetsLoading.value = false
    }
  }

  async function deletePermissionSet(psId: string) {
    permissionSetsLoading.value = true
    permissionSetsError.value = null
    try {
      await securityApi.deletePermissionSet(psId)
    } catch (err) {
      permissionSetsError.value = err instanceof Error ? err.message : 'Ошибка удаления набора разрешений'
      throw err
    } finally {
      permissionSetsLoading.value = false
    }
  }

  // --- Profiles ---

  async function fetchProfiles(filter?: ProfileFilter) {
    profilesLoading.value = true
    profilesError.value = null
    try {
      const response = await securityApi.listProfiles(filter)
      profiles.value = response.data ?? []
      profilesPagination.value = response.pagination
    } catch (err) {
      profilesError.value = err instanceof Error ? err.message : 'Ошибка загрузки профилей'
      throw err
    } finally {
      profilesLoading.value = false
    }
  }

  async function fetchProfile(profileId: string) {
    profilesLoading.value = true
    profilesError.value = null
    try {
      const response = await securityApi.getProfile(profileId)
      currentProfile.value = response.data
      return response.data
    } catch (err) {
      profilesError.value = err instanceof Error ? err.message : 'Ошибка загрузки профиля'
      throw err
    } finally {
      profilesLoading.value = false
    }
  }

  async function createProfile(data: CreateProfileRequest) {
    profilesLoading.value = true
    profilesError.value = null
    try {
      const response = await securityApi.createProfile(data)
      return response.data
    } catch (err) {
      profilesError.value = err instanceof Error ? err.message : 'Ошибка создания профиля'
      throw err
    } finally {
      profilesLoading.value = false
    }
  }

  async function updateProfile(profileId: string, data: UpdateProfileRequest) {
    profilesLoading.value = true
    profilesError.value = null
    try {
      const response = await securityApi.updateProfile(profileId, data)
      currentProfile.value = response.data
      return response.data
    } catch (err) {
      profilesError.value = err instanceof Error ? err.message : 'Ошибка обновления профиля'
      throw err
    } finally {
      profilesLoading.value = false
    }
  }

  async function deleteProfile(profileId: string) {
    profilesLoading.value = true
    profilesError.value = null
    try {
      await securityApi.deleteProfile(profileId)
    } catch (err) {
      profilesError.value = err instanceof Error ? err.message : 'Ошибка удаления профиля'
      throw err
    } finally {
      profilesLoading.value = false
    }
  }

  // --- Users ---

  async function fetchUsers(filter?: UserFilter) {
    usersLoading.value = true
    usersError.value = null
    try {
      const response = await securityApi.listUsers(filter)
      users.value = response.data ?? []
      usersPagination.value = response.pagination
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка загрузки пользователей'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  async function fetchUser(userId: string) {
    usersLoading.value = true
    usersError.value = null
    try {
      const response = await securityApi.getUser(userId)
      currentUser.value = response.data
      return response.data
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка загрузки пользователя'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  async function createUser(data: CreateUserRequest) {
    usersLoading.value = true
    usersError.value = null
    try {
      const response = await securityApi.createUser(data)
      return response.data
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка создания пользователя'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  async function updateUser(userId: string, data: UpdateUserRequest) {
    usersLoading.value = true
    usersError.value = null
    try {
      const response = await securityApi.updateUser(userId, data)
      currentUser.value = response.data
      return response.data
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка обновления пользователя'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  async function deleteUser(userId: string) {
    usersLoading.value = true
    usersError.value = null
    try {
      await securityApi.deleteUser(userId)
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка удаления пользователя'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  // --- User Permission Set Assignments ---

  async function fetchUserPermissionSets(userId: string) {
    usersLoading.value = true
    usersError.value = null
    try {
      const response = await securityApi.listUserPermissionSets(userId)
      userPermissionSets.value = response.data ?? []
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка загрузки назначенных наборов'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  async function assignPermissionSet(userId: string, permissionSetId: string) {
    usersLoading.value = true
    usersError.value = null
    try {
      await securityApi.assignPermissionSet(userId, permissionSetId)
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка назначения набора разрешений'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  async function revokePermissionSet(userId: string, assignmentId: string) {
    usersLoading.value = true
    usersError.value = null
    try {
      await securityApi.revokePermissionSet(userId, assignmentId)
    } catch (err) {
      usersError.value = err instanceof Error ? err.message : 'Ошибка отзыва набора разрешений'
      throw err
    } finally {
      usersLoading.value = false
    }
  }

  // --- Groups ---

  const groups = ref<Group[]>([])
  const currentGroup = ref<Group | null>(null)
  const groupsPagination = ref<PaginationMeta | null>(null)
  const groupsLoading = ref(false)
  const groupsError = ref<string | null>(null)
  const groupMembers = ref<GroupMember[]>([])

  async function fetchGroups(filter?: GroupFilter) {
    groupsLoading.value = true
    groupsError.value = null
    try {
      const response = await securityApi.listGroups(filter)
      groups.value = response.data ?? []
      groupsPagination.value = response.pagination
    } catch (err) {
      groupsError.value = err instanceof Error ? err.message : 'Ошибка загрузки групп'
      throw err
    } finally {
      groupsLoading.value = false
    }
  }

  async function fetchGroup(groupId: string) {
    groupsLoading.value = true
    groupsError.value = null
    try {
      const response = await securityApi.getGroup(groupId)
      currentGroup.value = response.data
      return response.data
    } catch (err) {
      groupsError.value = err instanceof Error ? err.message : 'Ошибка загрузки группы'
      throw err
    } finally {
      groupsLoading.value = false
    }
  }

  async function createGroup(data: CreateGroupRequest) {
    groupsLoading.value = true
    groupsError.value = null
    try {
      const response = await securityApi.createGroup(data)
      return response.data
    } catch (err) {
      groupsError.value = err instanceof Error ? err.message : 'Ошибка создания группы'
      throw err
    } finally {
      groupsLoading.value = false
    }
  }

  async function deleteGroup(groupId: string) {
    groupsLoading.value = true
    groupsError.value = null
    try {
      await securityApi.deleteGroup(groupId)
    } catch (err) {
      groupsError.value = err instanceof Error ? err.message : 'Ошибка удаления группы'
      throw err
    } finally {
      groupsLoading.value = false
    }
  }

  async function fetchGroupMembers(groupId: string) {
    groupsLoading.value = true
    groupsError.value = null
    try {
      const response = await securityApi.listGroupMembers(groupId)
      groupMembers.value = response.data ?? []
    } catch (err) {
      groupsError.value = err instanceof Error ? err.message : 'Ошибка загрузки членов группы'
      throw err
    } finally {
      groupsLoading.value = false
    }
  }

  async function addGroupMember(groupId: string, data: AddGroupMemberRequest) {
    groupsLoading.value = true
    groupsError.value = null
    try {
      const response = await securityApi.addGroupMember(groupId, data)
      return response.data
    } catch (err) {
      groupsError.value = err instanceof Error ? err.message : 'Ошибка добавления члена группы'
      throw err
    } finally {
      groupsLoading.value = false
    }
  }

  async function removeGroupMember(groupId: string, memberId: string) {
    groupsLoading.value = true
    groupsError.value = null
    try {
      await securityApi.removeGroupMember(groupId, memberId)
    } catch (err) {
      groupsError.value = err instanceof Error ? err.message : 'Ошибка удаления члена группы'
      throw err
    } finally {
      groupsLoading.value = false
    }
  }

  // --- Sharing Rules ---

  const sharingRules = ref<SharingRule[]>([])
  const currentSharingRule = ref<SharingRule | null>(null)
  const sharingRulesLoading = ref(false)
  const sharingRulesError = ref<string | null>(null)

  async function fetchSharingRules(filter: SharingRuleFilter) {
    sharingRulesLoading.value = true
    sharingRulesError.value = null
    try {
      const response = await securityApi.listSharingRules(filter)
      sharingRules.value = response.data ?? []
    } catch (err) {
      sharingRulesError.value = err instanceof Error ? err.message : 'Ошибка загрузки правил'
      throw err
    } finally {
      sharingRulesLoading.value = false
    }
  }

  async function fetchSharingRule(ruleId: string) {
    sharingRulesLoading.value = true
    sharingRulesError.value = null
    try {
      const response = await securityApi.getSharingRule(ruleId)
      currentSharingRule.value = response.data
      return response.data
    } catch (err) {
      sharingRulesError.value = err instanceof Error ? err.message : 'Ошибка загрузки правила'
      throw err
    } finally {
      sharingRulesLoading.value = false
    }
  }

  async function createSharingRule(data: CreateSharingRuleRequest) {
    sharingRulesLoading.value = true
    sharingRulesError.value = null
    try {
      const response = await securityApi.createSharingRule(data)
      return response.data
    } catch (err) {
      sharingRulesError.value = err instanceof Error ? err.message : 'Ошибка создания правила'
      throw err
    } finally {
      sharingRulesLoading.value = false
    }
  }

  async function updateSharingRule(ruleId: string, data: UpdateSharingRuleRequest) {
    sharingRulesLoading.value = true
    sharingRulesError.value = null
    try {
      const response = await securityApi.updateSharingRule(ruleId, data)
      currentSharingRule.value = response.data
      return response.data
    } catch (err) {
      sharingRulesError.value = err instanceof Error ? err.message : 'Ошибка обновления правила'
      throw err
    } finally {
      sharingRulesLoading.value = false
    }
  }

  async function deleteSharingRule(ruleId: string) {
    sharingRulesLoading.value = true
    sharingRulesError.value = null
    try {
      await securityApi.deleteSharingRule(ruleId)
    } catch (err) {
      sharingRulesError.value = err instanceof Error ? err.message : 'Ошибка удаления правила'
      throw err
    } finally {
      sharingRulesLoading.value = false
    }
  }

  return {
    // Roles
    roles,
    currentRole,
    rolesPagination,
    rolesLoading,
    rolesError,
    fetchRoles,
    fetchRole,
    createRole,
    updateRole,
    deleteRole,

    // Permission Sets
    permissionSets,
    currentPermissionSet,
    permissionSetsPagination,
    permissionSetsLoading,
    permissionSetsError,
    fetchPermissionSets,
    fetchPermissionSet,
    createPermissionSet,
    updatePermissionSet,
    deletePermissionSet,

    // Profiles
    profiles,
    currentProfile,
    profilesPagination,
    profilesLoading,
    profilesError,
    fetchProfiles,
    fetchProfile,
    createProfile,
    updateProfile,
    deleteProfile,

    // Users
    users,
    currentUser,
    usersPagination,
    usersLoading,
    usersError,
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

    // Groups
    groups,
    currentGroup,
    groupsPagination,
    groupsLoading,
    groupsError,
    groupMembers,
    fetchGroups,
    fetchGroup,
    createGroup,
    deleteGroup,
    fetchGroupMembers,
    addGroupMember,
    removeGroupMember,

    // Sharing Rules
    sharingRules,
    currentSharingRule,
    sharingRulesLoading,
    sharingRulesError,
    fetchSharingRules,
    fetchSharingRule,
    createSharingRule,
    updateSharingRule,
    deleteSharingRule,
  }
})
