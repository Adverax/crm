// Profile Dashboard types (ADR-0032)

export interface DashLink {
  label: string
  url: string
  icon?: string
}

export interface DashboardWidget {
  key: string
  type: 'list' | 'metric' | 'link_list'
  label: string
  size: 'full' | 'half' | 'third'
  query?: string
  columns?: string[]
  objectApiName?: string
  format?: 'number' | 'currency' | 'percent'
  links?: DashLink[]
}

export interface DashboardConfig {
  widgets: DashboardWidget[]
}

export interface ProfileDashboard {
  id: string
  profileId: string
  config: DashboardConfig
  createdAt: string
  updatedAt: string
}

export interface CreateProfileDashboardRequest {
  profile_id: string
  config: DashboardConfig
}

export interface UpdateProfileDashboardRequest {
  config: DashboardConfig
}

// Resolved dashboard (from GET /api/v1/dashboard)
export interface ListWidgetData {
  records: Record<string, unknown>[]
  totalCount: number
}

export interface MetricWidgetData {
  value: number | string
}

export interface ResolvedWidget {
  key: string
  type: 'list' | 'metric' | 'link_list'
  label: string
  size: 'full' | 'half' | 'third'
  objectApiName?: string
  columns?: string[]
  format?: string
  links?: DashLink[]
  data: ListWidgetData | MetricWidgetData | null
}

export interface ResolvedDashboard {
  widgets: ResolvedWidget[]
}
