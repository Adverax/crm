import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

// --- Derived from OpenAPI spec (single source of truth) ---

export type ObjectNavItem = CamelCaseKeys<components['schemas']['ObjectNavItem']>
export type ObjectDescribe = CamelCaseKeys<components['schemas']['ObjectDescribe']>
export type FieldDescribe = CamelCaseKeys<components['schemas']['FieldDescribe']>

// --- Inline sub-types from FieldDescribe.config ---

export type FieldConfigDescribe = NonNullable<FieldDescribe['config']>
export type PicklistValueDescribe = NonNullable<NonNullable<FieldConfigDescribe['values']>[number]>

// --- Business types (not from API schema) ---

export type RecordData = Record<string, unknown>

export interface RecordPagination {
  page: number
  perPage: number
  total: number
  totalPages: number
}
