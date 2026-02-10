export type PsType = 'grant' | 'deny'

export interface UserRole {
  id: string
  apiName: string
  label: string
  description: string
  parentId: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateRoleRequest {
  apiName: string
  label: string
  description?: string
  parentId?: string | null
}

export interface UpdateRoleRequest {
  label: string
  description?: string
  parentId?: string | null
}

export interface PermissionSet {
  id: string
  apiName: string
  label: string
  description: string
  psType: PsType
  createdAt: string
  updatedAt: string
}

export interface CreatePermissionSetRequest {
  apiName: string
  label: string
  description?: string
  psType: PsType
}

export interface UpdatePermissionSetRequest {
  label: string
  description?: string
}

export interface Profile {
  id: string
  apiName: string
  label: string
  description: string
  basePermissionSetId: string
  createdAt: string
  updatedAt: string
}

export interface CreateProfileRequest {
  apiName: string
  label: string
  description?: string
}

export interface UpdateProfileRequest {
  label: string
  description?: string
}

export interface User {
  id: string
  username: string
  email: string
  firstName: string
  lastName: string
  profileId: string
  roleId: string | null
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateUserRequest {
  username: string
  email: string
  firstName?: string
  lastName?: string
  profileId: string
  roleId?: string | null
}

export interface UpdateUserRequest {
  email: string
  firstName?: string
  lastName?: string
  profileId: string
  roleId?: string | null
  isActive: boolean
}

export interface PermissionSetAssignment {
  id: string
  permissionSetId: string
  userId: string
  createdAt: string
}

export interface ObjectPermission {
  id: string
  permissionSetId: string
  objectId: string
  permissions: number
  createdAt: string
  updatedAt: string
}

export interface SetObjectPermissionRequest {
  objectId: string
  permissions: number
}

export interface FieldPermission {
  id: string
  permissionSetId: string
  fieldId: string
  permissions: number
  createdAt: string
  updatedAt: string
}

export interface SetFieldPermissionRequest {
  fieldId: string
  permissions: number
}

export interface RoleFilter {
  page?: number
  perPage?: number
}

export interface PermissionSetFilter {
  page?: number
  perPage?: number
  psType?: PsType
}

export interface ProfileFilter {
  page?: number
  perPage?: number
}

export interface UserFilter {
  page?: number
  perPage?: number
  isActive?: boolean
}

// --- Groups ---

export type GroupType = 'personal' | 'role' | 'role_and_subordinates' | 'public'

export interface Group {
  id: string
  apiName: string
  label: string
  groupType: GroupType
  relatedRoleId: string | null
  relatedUserId: string | null
  createdAt: string
  updatedAt: string
}

export interface GroupMember {
  id: string
  groupId: string
  memberUserId: string | null
  memberGroupId: string | null
  createdAt: string
}

export interface CreateGroupRequest {
  apiName: string
  label: string
  groupType: GroupType
  relatedRoleId?: string | null
  relatedUserId?: string | null
}

export interface AddGroupMemberRequest {
  memberUserId?: string | null
  memberGroupId?: string | null
}

export interface GroupFilter {
  page?: number
  perPage?: number
  groupType?: GroupType
}

// --- Sharing Rules ---

export type RuleType = 'owner_based' | 'criteria_based'
export type AccessLevel = 'read' | 'read_write'

export interface SharingRule {
  id: string
  objectId: string
  ruleType: RuleType
  sourceGroupId: string
  targetGroupId: string
  accessLevel: AccessLevel
  criteriaField: string | null
  criteriaOp: string | null
  criteriaValue: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateSharingRuleRequest {
  objectId: string
  ruleType: RuleType
  sourceGroupId: string
  targetGroupId: string
  accessLevel: AccessLevel
  criteriaField?: string | null
  criteriaOp?: string | null
  criteriaValue?: string | null
}

export interface UpdateSharingRuleRequest {
  targetGroupId: string
  accessLevel: AccessLevel
  criteriaField?: string | null
  criteriaOp?: string | null
  criteriaValue?: string | null
}

export interface SharingRuleFilter {
  objectId: string
}

export const OLS_READ = 1
export const OLS_CREATE = 2
export const OLS_UPDATE = 4
export const OLS_DELETE = 8
export const OLS_ALL = 15

export const FLS_READ = 1
export const FLS_WRITE = 2
export const FLS_ALL = 3

export interface BitmaskFlag {
  bit: number
  label: string
  key: string
}

export const OLS_FLAGS: BitmaskFlag[] = [
  { bit: OLS_READ, label: 'Чтение', key: 'read' },
  { bit: OLS_CREATE, label: 'Создание', key: 'create' },
  { bit: OLS_UPDATE, label: 'Обновление', key: 'update' },
  { bit: OLS_DELETE, label: 'Удаление', key: 'delete' },
]

export const FLS_FLAGS: BitmaskFlag[] = [
  { bit: FLS_READ, label: 'Чтение', key: 'read' },
  { bit: FLS_WRITE, label: 'Запись', key: 'write' },
]
