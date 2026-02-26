// Profile Navigation types (ADR-0032)

export interface NavItem {
  type: 'object' | 'link' | 'divider' | 'page'
  objectApiName?: string
  ovApiName?: string
  label?: string
  url?: string
  icon?: string
}

export interface NavGroup {
  key: string
  label: string
  icon?: string
  items: NavItem[]
}

export interface NavConfig {
  groups: NavGroup[]
}

export interface ProfileNavigation {
  id: string
  profileId: string
  config: NavConfig
  createdAt: string
  updatedAt: string
}

export interface CreateProfileNavigationRequest {
  profile_id: string
  config: NavConfig
}

export interface UpdateProfileNavigationRequest {
  config: NavConfig
}

// Resolved navigation (from GET /api/v1/navigation)
export interface ResolvedNavItem {
  type: 'object' | 'link' | 'divider' | 'page'
  objectApiName?: string
  ovApiName?: string
  label?: string
  pluralLabel?: string
  url?: string
  icon?: string
}

export interface ResolvedNavGroup {
  key: string
  label: string
  icon?: string
  items: ResolvedNavItem[]
}

export interface ResolvedNavigation {
  groups: ResolvedNavGroup[]
}
