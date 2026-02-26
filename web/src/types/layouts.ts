// Since OpenAPI types may not be generated yet, define inline types

export interface Layout {
  id: string
  objectViewId: string
  formFactor: string
  mode: string
  config: LayoutConfig
  createdAt: string
  updatedAt: string
}

export interface LayoutConfig {
  root?: LayoutComponent
  sectionConfig?: Record<string, SectionConfig>
  fieldConfig?: Record<string, LayoutFieldConfig>
  listConfig?: ListConfig
}

export interface LayoutComponent {
  type: string
  key?: string
  columns?: number
  colSpan?: number
  limit?: number
  children?: LayoutComponent[]
}

export interface SectionConfig {
  columns?: number
  collapsed?: boolean
  collapsible?: boolean
  visibilityExpr?: string
}

export interface LayoutFieldConfig {
  layoutRef?: string
  colSpan?: number
  uiKind?: unknown
  requiredExpr?: string
  readonlyExpr?: string
  visibilityExpr?: string
  reference?: RefConfig
}

export interface RefConfig {
  displayFields?: string[]
  searchFields?: string[]
  target?: string
  hint?: string
  filter?: RefFilterConfig
}

export interface RefFilterConfig {
  items: RefFilterItem[]
}

export interface RefFilterItem {
  field: string
  operator: string
  value: unknown
}

export interface ListConfig {
  view?: string
  columns?: ListColumnConfig[]
  sortBy?: ListSortConfig[]
  search?: ListSearchConfig
  rowActions?: unknown[]
}

export interface ListColumnConfig {
  field: string
  label?: string
  width?: string
  align?: string
  sortable?: boolean
  sortDir?: string
  filterable?: boolean
  uiKind?: unknown
}

export interface ListSortConfig {
  field: string
  direction: string
}

export interface ListSearchConfig {
  fields: string[]
  placeholder?: string
}

export interface CreateLayoutRequest {
  objectViewId: string
  formFactor: string
  mode: string
  config: LayoutConfig
}

export interface UpdateLayoutRequest {
  config: LayoutConfig
}

export interface SharedLayout {
  id: string
  apiName: string
  type: string
  label: string
  config: unknown
  createdAt: string
  updatedAt: string
}

export interface CreateSharedLayoutRequest {
  apiName: string
  type: string
  label: string
  config: unknown
}

export interface UpdateSharedLayoutRequest {
  label: string
  config: unknown
}
