export interface ObjectNavItem {
  apiName: string
  label: string
  pluralLabel: string
  isCreateable: boolean
  isQueryable: boolean
}

export interface FieldConfigDescribe {
  maxLength?: number
  precision?: number
  scale?: number
  defaultValue?: string | null
  values?: PicklistValueDescribe[]
}

export interface PicklistValueDescribe {
  id: string
  value: string
  label: string
  sortOrder: number
  isDefault: boolean
  isActive: boolean
}

export interface FieldDescribe {
  apiName: string
  label: string
  fieldType: string
  fieldSubtype: string | null
  isRequired: boolean
  isReadOnly: boolean
  isSystemField: boolean
  sortOrder: number
  config: FieldConfigDescribe
}

export interface ObjectDescribe {
  apiName: string
  label: string
  pluralLabel: string
  isCreateable: boolean
  isUpdateable: boolean
  isDeleteable: boolean
  fields: FieldDescribe[]
}

export type RecordData = Record<string, unknown>

export interface RecordPagination {
  page: number
  perPage: number
  total: number
  totalPages: number
}
