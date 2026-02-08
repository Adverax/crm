export type ObjectType = 'standard' | 'custom'

export interface ObjectDefinition {
  id: string
  apiName: string
  label: string
  pluralLabel: string
  description: string
  objectType: ObjectType
  isPlatformManaged: boolean
  isVisibleInSetup: boolean
  isCustomFieldsAllowed: boolean
  isDeleteableObject: boolean
  isCreateable: boolean
  isUpdateable: boolean
  isDeleteable: boolean
  isQueryable: boolean
  isSearchable: boolean
  hasActivities: boolean
  hasNotes: boolean
  hasHistoryTracking: boolean
  hasSharingRules: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateObjectRequest {
  apiName: string
  label: string
  pluralLabel: string
  description: string
  objectType: ObjectType
  isVisibleInSetup: boolean
  isCustomFieldsAllowed: boolean
  isDeleteableObject: boolean
  isCreateable: boolean
  isUpdateable: boolean
  isDeleteable: boolean
  isQueryable: boolean
  isSearchable: boolean
  hasActivities: boolean
  hasNotes: boolean
  hasHistoryTracking: boolean
  hasSharingRules: boolean
}

export interface UpdateObjectRequest {
  label: string
  pluralLabel: string
  description?: string
  isVisibleInSetup?: boolean
  isCustomFieldsAllowed?: boolean
  isDeleteableObject?: boolean
  isCreateable?: boolean
  isUpdateable?: boolean
  isDeleteable?: boolean
  isQueryable?: boolean
  isSearchable?: boolean
  hasActivities?: boolean
  hasNotes?: boolean
  hasHistoryTracking?: boolean
  hasSharingRules?: boolean
}

export type FieldType = 'text' | 'number' | 'boolean' | 'datetime' | 'picklist' | 'reference'
export type FieldSubtype =
  | 'plain' | 'area' | 'rich' | 'email' | 'phone' | 'url'
  | 'integer' | 'decimal' | 'currency' | 'percent' | 'auto_number'
  | 'date' | 'datetime' | 'time'
  | 'single' | 'multi'
  | 'association' | 'composition' | 'polymorphic'

export type OnDeleteAction = 'set_null' | 'cascade' | 'restrict'

export interface FieldConfig {
  maxLength?: number
  precision?: number
  scale?: number
  format?: string
  startValue?: number
  relationshipName?: string
  onDelete?: OnDeleteAction
  isReparentable?: boolean
  defaultValue?: string
}

export interface FieldDefinition {
  id: string
  objectId: string
  apiName: string
  label: string
  description: string
  helpText: string
  fieldType: FieldType
  fieldSubtype?: FieldSubtype
  referencedObjectId?: string
  isRequired: boolean
  isUnique: boolean
  config: FieldConfig
  isSystemField: boolean
  isCustom: boolean
  isPlatformManaged: boolean
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface CreateFieldRequest {
  apiName: string
  label: string
  fieldType: FieldType
  fieldSubtype?: FieldSubtype
  description?: string
  helpText?: string
  referencedObjectId?: string
  isRequired?: boolean
  isUnique?: boolean
  isCustom?: boolean
  config?: FieldConfig
  sortOrder?: number
}

export interface UpdateFieldRequest {
  label: string
  description?: string
  helpText?: string
  isRequired?: boolean
  isUnique?: boolean
  config?: FieldConfig
  sortOrder?: number
}

export interface PaginationMeta {
  page: number
  perPage: number
  total: number
  totalPages: number
}

export interface ApiResponse<T> {
  data: T
}

export interface ApiListResponse<T> {
  data: T[]
  pagination: PaginationMeta
}

export interface ApiError {
  code: string
  message: string
}

export interface ObjectFilter {
  page?: number
  perPage?: number
  objectType?: ObjectType
}
