import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Enum types (inline unions from OpenAPI) ---

export type ObjectType = components['schemas']['CreateObjectRequest']['object_type']
export type Visibility = components['schemas']['CreateObjectRequest']['visibility']
export type FieldType = components['schemas']['CreateFieldRequest']['field_type']
export type FieldSubtype = NonNullable<components['schemas']['CreateFieldRequest']['field_subtype']>
export type OnDeleteAction = NonNullable<components['schemas']['FieldConfig']['on_delete']>

// --- Derived from OpenAPI spec (single source of truth) ---

export type ObjectDefinition = CamelCaseKeys<components['schemas']['ObjectDefinition']>
export type CreateObjectRequest = CamelCaseKeys<components['schemas']['CreateObjectRequest']>
export type UpdateObjectRequest = CamelCaseKeys<components['schemas']['UpdateObjectRequest']>

export type FieldConfig = CamelCaseKeys<components['schemas']['FieldConfig']>
export type FieldDefinition = CamelCaseKeys<components['schemas']['FieldDefinitionSchema']>
export type CreateFieldRequest = CamelCaseKeys<components['schemas']['CreateFieldRequest']>
export type UpdateFieldRequest = CamelCaseKeys<components['schemas']['UpdateFieldRequest']>

export type PaginationMeta = CamelCaseKeys<components['schemas']['PaginationMeta']>

// --- Wrapper types (not part of OpenAPI schemas) ---

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

// --- Business-specific filter (not from API) ---

export interface ObjectFilter {
  page?: number
  perPage?: number
  objectType?: ObjectType
}
