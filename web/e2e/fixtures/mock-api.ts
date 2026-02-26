import { type Page } from '@playwright/test'

// ─── Mock data ──────────────────────────────────────────────

export const mockObjects = [
  {
    id: '11111111-1111-1111-1111-111111111111',
    api_name: 'account',
    label: 'Account',
    plural_label: 'Accounts',
    object_type: 'standard',
    description: 'Companies and organizations',
    is_searchable: true,
    is_creatable: true,
    is_updatable: true,
    is_deletable: true,
    is_queryable: true,
    has_sharing: false,
    has_history: false,
    has_activities: false,
    has_notes: false,
    has_attachments: false,
    has_triggers: false,
    created_at: '2026-01-15T10:00:00Z',
    updated_at: '2026-01-15T10:00:00Z',
  },
  {
    id: '22222222-2222-2222-2222-222222222222',
    api_name: 'custom_obj',
    label: 'Custom Object',
    plural_label: 'Custom Objects',
    object_type: 'custom',
    description: 'Custom object',
    is_searchable: true,
    is_creatable: true,
    is_updatable: true,
    is_deletable: true,
    is_queryable: true,
    has_sharing: false,
    has_history: false,
    has_activities: false,
    has_notes: false,
    has_attachments: false,
    has_triggers: false,
    created_at: '2026-02-01T12:00:00Z',
    updated_at: '2026-02-01T12:00:00Z',
  },
]

export const mockFields = [
  {
    id: 'f1111111-1111-1111-1111-111111111111',
    object_id: '11111111-1111-1111-1111-111111111111',
    api_name: 'name',
    label: 'Name',
    field_type: 'string',
    field_subtype: 'text',
    is_required: true,
    is_unique: false,
    is_indexed: false,
    max_length: 255,
    created_at: '2026-01-15T10:00:00Z',
    updated_at: '2026-01-15T10:00:00Z',
  },
]

export const mockRoles = [
  {
    id: 'r1111111-1111-1111-1111-111111111111',
    api_name: 'ceo',
    label: 'CEO',
    parent_id: null,
    description: 'Top-level role',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'r2222222-2222-2222-2222-222222222222',
    api_name: 'sales_manager',
    label: 'Sales Manager',
    parent_id: 'r1111111-1111-1111-1111-111111111111',
    description: 'Manages sales department',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockPermissionSets = [
  {
    id: 'ps111111-1111-1111-1111-111111111111',
    api_name: 'read_all',
    label: 'Read All',
    ps_type: 'grant',
    description: 'Grants read access',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'ps222222-2222-2222-2222-222222222222',
    api_name: 'deny_delete',
    label: 'Deny Delete',
    ps_type: 'deny',
    description: 'Denies delete access',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockProfiles = [
  {
    id: 'pf111111-1111-1111-1111-111111111111',
    api_name: 'system_admin',
    label: 'System Administrator',
    description: 'Full access',
    base_permission_set_id: 'ps111111-1111-1111-1111-111111111111',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'pf222222-2222-2222-2222-222222222222',
    api_name: 'standard_user',
    label: 'Standard User',
    description: 'Basic access',
    base_permission_set_id: 'ps222222-2222-2222-2222-222222222222',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockUsers = [
  {
    id: 'u1111111-1111-1111-1111-111111111111',
    username: 'admin',
    email: 'admin@example.com',
    first_name: 'John',
    last_name: 'Smith',
    profile_id: 'pf111111-1111-1111-1111-111111111111',
    role_id: 'r1111111-1111-1111-1111-111111111111',
    is_active: true,
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'u2222222-2222-2222-2222-222222222222',
    username: 'user1',
    email: 'user1@example.com',
    first_name: 'Peter',
    last_name: 'Johnson',
    profile_id: 'pf222222-2222-2222-2222-222222222222',
    role_id: 'r2222222-2222-2222-2222-222222222222',
    is_active: false,
    created_at: '2026-02-01T12:00:00Z',
    updated_at: '2026-02-01T12:00:00Z',
  },
]

export const mockObjectPermissions = [
  {
    id: 'op111111-1111-1111-1111-111111111111',
    permission_set_id: 'ps111111-1111-1111-1111-111111111111',
    object_id: '11111111-1111-1111-1111-111111111111',
    permissions: 15, // CRUD bitmask: 1+2+4+8
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockFieldPermissions = [
  {
    id: 'fp111111-1111-1111-1111-111111111111',
    permission_set_id: 'ps111111-1111-1111-1111-111111111111',
    field_id: 'f1111111-1111-1111-1111-111111111111',
    permissions: 3, // RW bitmask: 1+2
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockUserPermissionSets = [
  {
    id: 'upa11111-1111-1111-1111-111111111111',
    user_id: 'u1111111-1111-1111-1111-111111111111',
    permission_set_id: 'ps111111-1111-1111-1111-111111111111',
    created_at: '2026-01-10T10:00:00Z',
  },
]

export const mockGroups = [
  {
    id: 'g1111111-1111-1111-1111-111111111111',
    api_name: 'all_users',
    label: 'All Users',
    group_type: 'public',
    related_role_id: null,
    related_user_id: null,
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'g2222222-2222-2222-2222-222222222222',
    api_name: 'sales_team',
    label: 'Sales Team',
    group_type: 'public',
    related_role_id: null,
    related_user_id: null,
    created_at: '2026-02-01T12:00:00Z',
    updated_at: '2026-02-01T12:00:00Z',
  },
  {
    id: 'g3333333-3333-3333-3333-333333333333',
    api_name: 'ceo_role',
    label: 'CEO Role',
    group_type: 'role',
    related_role_id: 'r1111111-1111-1111-1111-111111111111',
    related_user_id: null,
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockGroupMembers = [
  {
    id: 'gm111111-1111-1111-1111-111111111111',
    group_id: 'g1111111-1111-1111-1111-111111111111',
    member_user_id: 'u1111111-1111-1111-1111-111111111111',
    member_group_id: null,
    created_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'gm222222-2222-2222-2222-222222222222',
    group_id: 'g1111111-1111-1111-1111-111111111111',
    member_user_id: 'u2222222-2222-2222-2222-222222222222',
    member_group_id: null,
    created_at: '2026-02-01T12:00:00Z',
  },
]

export const mockSharingRules = [
  {
    id: 'sr111111-1111-1111-1111-111111111111',
    object_id: '11111111-1111-1111-1111-111111111111',
    rule_type: 'owner_based',
    source_group_id: 'g2222222-2222-2222-2222-222222222222',
    target_group_id: 'g1111111-1111-1111-1111-111111111111',
    access_level: 'read',
    criteria_field: null,
    criteria_op: null,
    criteria_value: null,
    created_at: '2026-02-01T12:00:00Z',
    updated_at: '2026-02-01T12:00:00Z',
  },
  {
    id: 'sr222222-2222-2222-2222-222222222222',
    object_id: '11111111-1111-1111-1111-111111111111',
    rule_type: 'criteria_based',
    source_group_id: 'g1111111-1111-1111-1111-111111111111',
    target_group_id: 'g2222222-2222-2222-2222-222222222222',
    access_level: 'read_write',
    criteria_field: 'status',
    criteria_op: '=',
    criteria_value: 'active',
    created_at: '2026-02-05T12:00:00Z',
    updated_at: '2026-02-05T12:00:00Z',
  },
]

// ─── Helpers ────────────────────────────────────────────────

function listResponse<T>(items: T[], page = 1, perPage = 20) {
  return {
    data: items,
    meta: {
      page,
      per_page: perPage,
      total: items.length,
      total_pages: Math.ceil(items.length / perPage),
    },
  }
}

function singleResponse<T>(item: T) {
  return { data: item }
}

// ─── Auth mock data ────────────────────────────────────────

export const mockAuthMe = {
  id: 'u1111111-1111-1111-1111-111111111111',
  username: 'admin',
  email: 'admin@example.com',
  first_name: 'John',
  last_name: 'Smith',
}

export const mockTokenPair = {
  access_token: 'mock-access-token-jwt',
  refresh_token: 'mock-refresh-token-hex',
}

// ─── Auth helpers ──────────────────────────────────────────

export async function seedAuthToken(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('crm_access_token', 'mock-access-token-jwt')
    localStorage.setItem('crm_refresh_token', 'mock-refresh-token-hex')
  })
}

// ─── Route setup ────────────────────────────────────────────

export async function setupAuthRoutes(page: Page) {
  await page.route('**/api/v1/auth/login', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: { data: mockTokenPair } })
    }
    return route.continue()
  })

  await page.route('**/api/v1/auth/refresh', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: { data: mockTokenPair } })
    }
    return route.continue()
  })

  await page.route('**/api/v1/auth/logout', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ status: 204 })
    }
    return route.continue()
  })

  await page.route('**/api/v1/auth/me', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: { data: mockAuthMe } })
    }
    return route.continue()
  })

  await page.route('**/api/v1/auth/forgot-password', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: {} })
    }
    return route.continue()
  })

  await page.route('**/api/v1/auth/reset-password', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: {} })
    }
    return route.continue()
  })
}

export async function setupMetadataRoutes(page: Page) {
  // List objects
  await page.route('**/api/v1/admin/metadata/objects?*', (route) => {
    route.fulfill({ json: listResponse(mockObjects) })
  })
  await page.route('**/api/v1/admin/metadata/objects', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockObjects) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: singleResponse({ ...mockObjects[1], id: 'new-object-id' }) })
    }
    return route.continue()
  })

  // Single object
  for (const obj of mockObjects) {
    await page.route(`**/api/v1/admin/metadata/objects/${obj.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(obj) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(obj) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })

    // Fields for object
    await page.route(`**/api/v1/admin/metadata/objects/${obj.id}/fields?*`, (route) => {
      route.fulfill({
        json: listResponse(mockFields.filter((f) => f.object_id === obj.id)),
      })
    })
    await page.route(`**/api/v1/admin/metadata/objects/${obj.id}/fields`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: listResponse(mockFields.filter((f) => f.object_id === obj.id)),
        })
      }
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse(mockFields[0]) })
      }
      return route.continue()
    })
  }
}

export async function setupSecurityRoutes(page: Page) {
  // ── Roles ──
  await page.route('**/api/v1/admin/security/roles?*', (route) => {
    route.fulfill({ json: listResponse(mockRoles) })
  })
  await page.route('**/api/v1/admin/security/roles', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockRoles) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: singleResponse({ ...mockRoles[0], id: 'new-role-id' }) })
    }
    return route.continue()
  })
  for (const role of mockRoles) {
    await page.route(`**/api/v1/admin/security/roles/${role.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(role) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(role) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }

  // ── Permission Sets ──
  await page.route('**/api/v1/admin/security/permission-sets?*', (route) => {
    route.fulfill({ json: listResponse(mockPermissionSets) })
  })
  await page.route('**/api/v1/admin/security/permission-sets', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockPermissionSets) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockPermissionSets[0], id: 'new-ps-id' }),
      })
    }
    return route.continue()
  })
  for (const ps of mockPermissionSets) {
    await page.route(`**/api/v1/admin/security/permission-sets/${ps.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(ps) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(ps) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })

    // Object permissions for PS
    await page.route(
      `**/api/v1/admin/security/permission-sets/${ps.id}/object-permissions?*`,
      (route) => {
        route.fulfill({
          json: listResponse(
            mockObjectPermissions.filter((op) => op.permission_set_id === ps.id),
          ),
        })
      },
    )
    await page.route(
      `**/api/v1/admin/security/permission-sets/${ps.id}/object-permissions`,
      (route) => {
        if (route.request().method() === 'GET') {
          return route.fulfill({
            json: listResponse(
              mockObjectPermissions.filter((op) => op.permission_set_id === ps.id),
            ),
          })
        }
        if (route.request().method() === 'PUT') {
          return route.fulfill({ json: singleResponse(mockObjectPermissions[0]) })
        }
        return route.continue()
      },
    )

    // Field permissions for PS
    await page.route(
      `**/api/v1/admin/security/permission-sets/${ps.id}/field-permissions?*`,
      (route) => {
        route.fulfill({
          json: listResponse(
            mockFieldPermissions.filter((fp) => fp.permission_set_id === ps.id),
          ),
        })
      },
    )
    await page.route(
      `**/api/v1/admin/security/permission-sets/${ps.id}/field-permissions`,
      (route) => {
        if (route.request().method() === 'GET') {
          return route.fulfill({
            json: listResponse(
              mockFieldPermissions.filter((fp) => fp.permission_set_id === ps.id),
            ),
          })
        }
        return route.continue()
      },
    )
  }

  // ── Profiles ──
  await page.route('**/api/v1/admin/security/profiles?*', (route) => {
    route.fulfill({ json: listResponse(mockProfiles) })
  })
  await page.route('**/api/v1/admin/security/profiles', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockProfiles) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockProfiles[0], id: 'new-profile-id' }),
      })
    }
    return route.continue()
  })
  for (const profile of mockProfiles) {
    await page.route(`**/api/v1/admin/security/profiles/${profile.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(profile) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(profile) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }

  // ── Users ──
  await page.route('**/api/v1/admin/security/users?*', (route) => {
    route.fulfill({ json: listResponse(mockUsers) })
  })
  await page.route('**/api/v1/admin/security/users', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockUsers) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockUsers[0], id: 'new-user-id' }),
      })
    }
    return route.continue()
  })
  for (const user of mockUsers) {
    await page.route(`**/api/v1/admin/security/users/${user.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(user) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(user) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })

    // User permission set assignments
    await page.route(`**/api/v1/admin/security/users/${user.id}/permission-sets?*`, (route) => {
      route.fulfill({
        json: listResponse(
          mockUserPermissionSets.filter((ups) => ups.user_id === user.id),
        ),
      })
    })
    await page.route(`**/api/v1/admin/security/users/${user.id}/permission-sets`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: listResponse(
            mockUserPermissionSets.filter((ups) => ups.user_id === user.id),
          ),
        })
      }
      if (route.request().method() === 'POST') {
        return route.fulfill({ status: 201 })
      }
      return route.continue()
    })
  }
}

export async function setupGroupRoutes(page: Page) {
  // ── Groups ──
  await page.route('**/api/v1/admin/security/groups?*', (route) => {
    route.fulfill({ json: listResponse(mockGroups) })
  })
  await page.route('**/api/v1/admin/security/groups', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockGroups) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockGroups[0], id: 'new-group-id' }),
      })
    }
    return route.continue()
  })
  for (const group of mockGroups) {
    await page.route(`**/api/v1/admin/security/groups/${group.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(group) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })

    // Group members
    await page.route(`**/api/v1/admin/security/groups/${group.id}/members`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: singleResponse(
            mockGroupMembers.filter((gm) => gm.group_id === group.id),
          ),
        })
      }
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: singleResponse(mockGroupMembers[0]),
        })
      }
      return route.continue()
    })
  }

  // Delete member route (pattern: /groups/:id/members/:memberId)
  await page.route(/\/api\/v1\/admin\/security\/groups\/[^/]+\/members\/[^/]+$/, (route) => {
    if (route.request().method() === 'DELETE') {
      return route.fulfill({ status: 204 })
    }
    return route.continue()
  })
}

export async function setupSharingRuleRoutes(page: Page) {
  // ── Sharing Rules ──
  await page.route('**/api/v1/admin/security/sharing-rules?*', (route) => {
    route.fulfill({ json: singleResponse(mockSharingRules) })
  })
  await page.route('**/api/v1/admin/security/sharing-rules', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: singleResponse(mockSharingRules) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockSharingRules[0], id: 'new-rule-id' }),
      })
    }
    return route.continue()
  })
  for (const rule of mockSharingRules) {
    await page.route(`**/api/v1/admin/security/sharing-rules/${rule.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(rule) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(rule) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

// ─── Templates mock data ─────────────────────────────────────

export const mockTemplates = [
  {
    id: 'sales_crm',
    label: 'Sales CRM',
    description: 'CRM for sales teams: accounts, contacts, opportunities, and tasks',
    status: 'available',
    objects: 4,
    fields: 36,
  },
  {
    id: 'recruiting',
    label: 'Recruiting',
    description: 'Applicant tracking system: positions, candidates, applications, and interviews',
    status: 'available',
    objects: 4,
    fields: 28,
  },
]

export async function setupTemplateRoutes(page: Page) {
  await page.route('**/api/v1/admin/templates', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: { data: mockTemplates } })
    }
    return route.continue()
  })

  for (const tmpl of mockTemplates) {
    await page.route(`**/api/v1/admin/templates/${tmpl.id}/apply`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: { data: { template_id: tmpl.id, message: 'template applied successfully' } },
        })
      }
      return route.continue()
    })
  }
}

// ─── Territory mock data ─────────────────────────────────────

export const mockTerritoryModels = [
  {
    id: 'tm111111-1111-1111-1111-111111111111',
    api_name: 'q1_2026',
    label: 'Q1 2026',
    description: 'Territory model for Q1',
    status: 'planning',
    activated_at: null,
    archived_at: null,
    created_at: '2026-01-15T10:00:00Z',
    updated_at: '2026-01-15T10:00:00Z',
  },
  {
    id: 'tm222222-2222-2222-2222-222222222222',
    api_name: 'q4_2025',
    label: 'Q4 2025',
    description: 'Territory model for Q4',
    status: 'active',
    activated_at: '2025-10-01T10:00:00Z',
    archived_at: null,
    created_at: '2025-09-15T10:00:00Z',
    updated_at: '2025-10-01T10:00:00Z',
  },
]

export const mockTerritories = [
  {
    id: 'tt111111-1111-1111-1111-111111111111',
    model_id: 'tm111111-1111-1111-1111-111111111111',
    parent_id: null,
    api_name: 'north_america',
    label: 'North America',
    description: 'North America region',
    created_at: '2026-01-16T10:00:00Z',
    updated_at: '2026-01-16T10:00:00Z',
  },
  {
    id: 'tt222222-2222-2222-2222-222222222222',
    model_id: 'tm111111-1111-1111-1111-111111111111',
    parent_id: 'tt111111-1111-1111-1111-111111111111',
    api_name: 'us_east',
    label: 'US East',
    description: 'East Coast',
    created_at: '2026-01-17T10:00:00Z',
    updated_at: '2026-01-17T10:00:00Z',
  },
]

export const mockTerritoryUsers = [
  {
    id: 'tu111111-1111-1111-1111-111111111111',
    user_id: 'u1111111-1111-1111-1111-111111111111',
    territory_id: 'tt111111-1111-1111-1111-111111111111',
    created_at: '2026-01-16T10:00:00Z',
  },
]

export const mockObjectDefaults = [
  {
    id: 'od111111-1111-1111-1111-111111111111',
    territory_id: 'tt111111-1111-1111-1111-111111111111',
    object_id: '11111111-1111-1111-1111-111111111111',
    access_level: 'read_write',
    created_at: '2026-01-16T10:00:00Z',
    updated_at: '2026-01-16T10:00:00Z',
  },
]

export async function setupTerritoryRoutes(page: Page) {
  // ── Models ──
  await page.route('**/api/v1/admin/territory/models?*', (route) => {
    route.fulfill({ json: listResponse(mockTerritoryModels) })
  })
  await page.route('**/api/v1/admin/territory/models', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockTerritoryModels) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockTerritoryModels[0], id: 'new-model-id' }),
      })
    }
    return route.continue()
  })
  for (const model of mockTerritoryModels) {
    await page.route(`**/api/v1/admin/territory/models/${model.id}/activate`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/territory/models/${model.id}/archive`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/territory/models/${model.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(model) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(model) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }

  // ── Territories ──
  await page.route('**/api/v1/admin/territory/territories?*', (route) => {
    route.fulfill({ json: listResponse(mockTerritories) })
  })
  await page.route('**/api/v1/admin/territory/territories', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: listResponse(mockTerritories) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockTerritories[0], id: 'new-territory-id' }),
      })
    }
    return route.continue()
  })
  for (const territory of mockTerritories) {
    await page.route(`**/api/v1/admin/territory/territories/${territory.id}/users`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: singleResponse(
            mockTerritoryUsers.filter((tu) => tu.territory_id === territory.id),
          ),
        })
      }
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse(mockTerritoryUsers[0]) })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/territory/territories/${territory.id}/object-defaults`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: singleResponse(
            mockObjectDefaults.filter((od) => od.territory_id === territory.id),
          ),
        })
      }
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse(mockObjectDefaults[0]) })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/territory/territories/${territory.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(territory) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(territory) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }

  // Delete routes for nested resources
  await page.route(/\/api\/v1\/admin\/territory\/territories\/[^/]+\/users\/[^/]+$/, (route) => {
    if (route.request().method() === 'DELETE') {
      return route.fulfill({ status: 204 })
    }
    return route.continue()
  })
  await page.route(/\/api\/v1\/admin\/territory\/territories\/[^/]+\/object-defaults\/[^/]+$/, (route) => {
    if (route.request().method() === 'DELETE') {
      return route.fulfill({ status: 204 })
    }
    return route.continue()
  })
}

// ─── Validation Rules mock data ─────────────────────────────────

export const mockValidationRules = [
  {
    id: 'vr111111-1111-1111-1111-111111111111',
    object_id: '11111111-1111-1111-1111-111111111111',
    api_name: 'name_required',
    label: 'Name Required',
    description: 'Checks that name is not empty',
    expression: 'size(record.Name) > 0',
    error_message: 'Name field is required',
    error_code: 'validation_failed',
    severity: 'error',
    when_expression: null,
    applies_to: 'create,update',
    sort_order: 0,
    is_active: true,
    created_at: '2026-02-10T10:00:00Z',
    updated_at: '2026-02-10T10:00:00Z',
  },
  {
    id: 'vr222222-2222-2222-2222-222222222222',
    object_id: '11111111-1111-1111-1111-111111111111',
    api_name: 'phone_format',
    label: 'Phone Format',
    description: 'Warning about phone format',
    expression: 'record.Phone.startsWith("+")',
    error_message: 'International phone format is recommended',
    error_code: 'phone_format',
    severity: 'warning',
    when_expression: 'has(record.Phone)',
    applies_to: 'create,update',
    sort_order: 1,
    is_active: false,
    created_at: '2026-02-11T10:00:00Z',
    updated_at: '2026-02-11T10:00:00Z',
  },
]

export async function setupValidationRuleRoutes(page: Page) {
  for (const obj of mockObjects) {
    // List rules for object
    await page.route(`**/api/v1/admin/metadata/objects/${obj.id}/rules?*`, (route) => {
      route.fulfill({
        json: singleResponse(
          mockValidationRules.filter((r) => r.object_id === obj.id),
        ),
      })
    })
    await page.route(`**/api/v1/admin/metadata/objects/${obj.id}/rules`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: singleResponse(
            mockValidationRules.filter((r) => r.object_id === obj.id),
          ),
        })
      }
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: singleResponse({ ...mockValidationRules[0], id: 'new-rule-id' }),
        })
      }
      return route.continue()
    })
  }

  // Single rule routes
  for (const rule of mockValidationRules) {
    await page.route(`**/api/v1/admin/metadata/objects/${rule.object_id}/rules/${rule.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(rule) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(rule) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

// ─── Functions mock data ─────────────────────────────────────

export const mockFunctions = [
  {
    id: 'fn111111-1111-1111-1111-111111111111',
    name: 'discount',
    description: 'Calculates discount by amount',
    params: [
      { name: 'amount', type: 'number', description: 'Amount' },
    ],
    return_type: 'number',
    body: 'amount > 1000 ? amount * 0.1 : 0.0',
    created_at: '2026-02-15T10:00:00Z',
    updated_at: '2026-02-15T10:00:00Z',
  },
  {
    id: 'fn222222-2222-2222-2222-222222222222',
    name: 'is_premium',
    description: 'Checks premium status',
    params: [
      { name: 'total', type: 'number', description: 'Total amount' },
      { name: 'count', type: 'number', description: 'Order count' },
    ],
    return_type: 'boolean',
    body: 'total > 10000 && count > 5',
    created_at: '2026-02-15T11:00:00Z',
    updated_at: '2026-02-15T11:00:00Z',
  },
]

export async function setupFunctionRoutes(page: Page) {
  await page.route('**/api/v1/admin/functions?*', (route) => {
    route.fulfill({ json: singleResponse(mockFunctions) })
  })
  await page.route('**/api/v1/admin/functions', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: singleResponse(mockFunctions) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockFunctions[0], id: 'new-function-id' }),
      })
    }
    return route.continue()
  })
  for (const fn of mockFunctions) {
    await page.route(`**/api/v1/admin/functions/${fn.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(fn) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(fn) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }

  // CEL validate endpoint
  await page.route('**/api/v1/admin/cel/validate', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: { valid: true, return_type: 'bool' },
      })
    }
    return route.continue()
  })
}

// ─── Object Views mock data ─────────────────────────────────

export const mockObjectViews = [
  {
    id: 'ov111111-1111-1111-1111-111111111111',
    profile_id: null,
    api_name: 'account_default',
    label: 'Account Default View',
    description: 'Default view for accounts',
    config: {
      read: {
        fields: ['Name', 'Industry', 'Phone'],
        actions: [
          {
            key: 'send_email',
            label: 'Send Email',
            type: 'primary',
            icon: 'mail',
            visibility_expr: 'true',
          },
        ],
        queries: [
          {
            name: 'recent_activities',
            soql: 'SELECT Id, Subject FROM Activity WHERE WhatId = :recordId',
            when: '',
          },
        ],
        computed: [
          {
            name: 'display_name',
            type: 'string',
            expr: 'record.Name + " (" + record.Industry + ")"',
            when: '',
          },
        ],
      },
      write: {
        validation: [
          {
            expr: 'size(record.Name) > 0',
            message: 'Name is required',
            code: 'name_required',
            severity: 'error',
            when: '',
          },
        ],
        defaults: [],
        computed: [],
        mutations: [],
      },
    },
    created_at: '2026-02-16T10:00:00Z',
    updated_at: '2026-02-16T10:00:00Z',
  },
  {
    id: 'ov222222-2222-2222-2222-222222222222',
    profile_id: 'pf222222-2222-2222-2222-222222222222',
    api_name: 'account_sales_view',
    label: 'Account Sales View',
    description: 'Sales-specific view for accounts',
    config: {
      read: {
        fields: [],
        actions: [],
        queries: [],
        computed: [],
      },
      write: null,
    },
    created_at: '2026-02-16T11:00:00Z',
    updated_at: '2026-02-16T11:00:00Z',
  },
]

export async function setupObjectViewRoutes(page: Page) {
  await page.route('**/api/v1/admin/object-views', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: singleResponse(mockObjectViews) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        status: 201,
        json: singleResponse({ ...mockObjectViews[0], id: 'new-view-id' }),
      })
    }
    return route.continue()
  })
  for (const view of mockObjectViews) {
    await page.route(`**/api/v1/admin/object-views/${view.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(view) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(view) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }

  // View resolution endpoint
  for (const view of mockObjectViews) {
    await page.route(`**/api/v1/view/${view.api_name}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(view) })
      }
      return route.continue()
    })
  }
}

// ─── Procedures mock data ─────────────────────────────────

export const mockProcedures = [
  {
    id: 'pr111111-1111-1111-1111-111111111111',
    code: 'create_account',
    name: 'Create Account',
    description: 'Creates a new account with validation',
    draft_version_id: 'pv111111-1111-1111-1111-111111111111',
    published_version_id: 'pv222222-2222-2222-2222-222222222222',
    created_at: '2026-02-20T10:00:00Z',
    updated_at: '2026-02-20T10:00:00Z',
  },
  {
    id: 'pr222222-2222-2222-2222-222222222222',
    code: 'send_welcome',
    name: 'Send Welcome Email',
    description: 'Sends welcome email to new contacts',
    draft_version_id: 'pv333333-3333-3333-3333-333333333333',
    published_version_id: null,
    created_at: '2026-02-21T10:00:00Z',
    updated_at: '2026-02-21T10:00:00Z',
  },
]

export const mockProcedureVersions = [
  {
    id: 'pv111111-1111-1111-1111-111111111111',
    procedure_id: 'pr111111-1111-1111-1111-111111111111',
    version: 2,
    definition: {
      commands: [
        { type: 'compute.validate', condition: '$.input.name != ""', code: 'name_required', message: 'Name is required' },
        { type: 'record.create', object: 'Account', data: { Name: '$.input.name' }, as: 'account' },
      ],
      result: { id: '$.account.id' },
    },
    status: 'draft',
    change_summary: 'Added validation step',
    created_by: null,
    created_at: '2026-02-20T12:00:00Z',
    published_at: null,
  },
  {
    id: 'pv222222-2222-2222-2222-222222222222',
    procedure_id: 'pr111111-1111-1111-1111-111111111111',
    version: 1,
    definition: {
      commands: [
        { type: 'record.create', object: 'Account', data: { Name: '$.input.name' }, as: 'account' },
      ],
      result: { id: '$.account.id' },
    },
    status: 'published',
    change_summary: 'Initial version',
    created_by: null,
    created_at: '2026-02-20T10:00:00Z',
    published_at: '2026-02-20T11:00:00Z',
  },
  {
    id: 'pv333333-3333-3333-3333-333333333333',
    procedure_id: 'pr222222-2222-2222-2222-222222222222',
    version: 1,
    definition: {
      commands: [
        { type: 'notification.email', as: 'email' },
      ],
      result: {},
    },
    status: 'draft',
    change_summary: 'Draft v1',
    created_by: null,
    created_at: '2026-02-21T10:00:00Z',
    published_at: null,
  },
]

function getProcedureWithVersions(procId: string) {
  const proc = mockProcedures.find((p) => p.id === procId)
  if (!proc) return null
  const draft = mockProcedureVersions.find((v) => v.id === proc.draft_version_id) ?? null
  const published = mockProcedureVersions.find((v) => v.id === proc.published_version_id) ?? null
  return { procedure: proc, draft_version: draft, published_version: published }
}

export async function setupProcedureRoutes(page: Page) {
  await page.route('**/api/v1/admin/procedures?*', (route) => {
    route.fulfill({ json: singleResponse(mockProcedures) })
  })
  await page.route('**/api/v1/admin/procedures', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: singleResponse(mockProcedures) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse(getProcedureWithVersions(mockProcedures[0].id)),
      })
    }
    return route.continue()
  })
  for (const proc of mockProcedures) {
    const pvs = getProcedureWithVersions(proc.id)
    // Versions sub-route MUST be registered before the catch-all ID route
    await page.route(`**/api/v1/admin/procedures/${proc.id}/versions`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: singleResponse(
            mockProcedureVersions.filter((v) => v.procedure_id === proc.id),
          ),
        })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/procedures/${proc.id}/draft`, (route) => {
      if (route.request().method() === 'PUT') {
        const draft = mockProcedureVersions.find((v) => v.id === proc.draft_version_id)
        return route.fulfill({ json: singleResponse(draft ?? mockProcedureVersions[0]) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/procedures/${proc.id}/publish`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse(mockProcedureVersions[0]) })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/procedures/${proc.id}/rollback`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse(mockProcedureVersions[1]) })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/procedures/${proc.id}/execute`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: singleResponse({ success: true, result: { id: 'new-id' }, warnings: [] }),
        })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/procedures/${proc.id}/dry-run`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: singleResponse({
            success: true,
            result: { dry_run: true },
            warnings: [],
            trace: [
              { step: 'account', type: 'record.create', status: 'ok', duration_ms: 5 },
            ],
          }),
        })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/procedures/${proc.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(pvs) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(proc) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

// ─── Credentials mock data ─────────────────────────────────

export const mockCredentials = [
  {
    id: 'cr111111-1111-1111-1111-111111111111',
    code: 'stripe_api',
    name: 'Stripe API',
    description: 'Stripe payment API',
    type: 'api_key',
    base_url: 'https://api.stripe.com',
    is_active: true,
    created_at: '2026-02-20T10:00:00Z',
    updated_at: '2026-02-20T10:00:00Z',
  },
  {
    id: 'cr222222-2222-2222-2222-222222222222',
    code: 'slack_oauth',
    name: 'Slack OAuth',
    description: 'Slack integration',
    type: 'oauth2_client',
    base_url: 'https://slack.com/api',
    is_active: false,
    created_at: '2026-02-21T10:00:00Z',
    updated_at: '2026-02-21T10:00:00Z',
  },
]

export const mockCredentialUsageLog = [
  {
    id: 'cu111111-1111-1111-1111-111111111111',
    credential_id: 'cr111111-1111-1111-1111-111111111111',
    procedure_code: 'create_account',
    request_url: 'https://api.stripe.com/v1/charges',
    response_status: 200,
    success: true,
    error_message: '',
    duration_ms: 120,
    user_id: null,
    created_at: '2026-02-20T12:00:00Z',
  },
]

export async function setupCredentialRoutes(page: Page) {
  await page.route('**/api/v1/admin/credentials?*', (route) => {
    route.fulfill({ json: singleResponse(mockCredentials) })
  })
  await page.route('**/api/v1/admin/credentials', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: singleResponse(mockCredentials) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockCredentials[0], id: 'new-credential-id' }),
      })
    }
    return route.continue()
  })
  for (const cred of mockCredentials) {
    await page.route(`**/api/v1/admin/credentials/${cred.id}/test`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse({ success: true }) })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/credentials/${cred.id}/usage`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: singleResponse(
            mockCredentialUsageLog.filter((u) => u.credential_id === cred.id),
          ),
        })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/credentials/${cred.id}/deactivate`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse({ success: true }) })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/credentials/${cred.id}/activate`, (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ json: singleResponse({ success: true }) })
      }
      return route.continue()
    })
    await page.route(`**/api/v1/admin/credentials/${cred.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(cred) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(cred) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

// ─── Automation Rules mock data ──────────────────────────────

export const mockAutomationRules = [
  {
    id: 'ar111111-1111-1111-1111-111111111111',
    object_id: '11111111-1111-1111-1111-111111111111',
    name: 'Notify on insert',
    description: 'Send notification after record insert',
    event_type: 'after_insert',
    condition: null,
    procedure_code: 'notify_manager',
    execution_mode: 'per_record',
    sort_order: 0,
    is_active: true,
    created_at: '2026-02-20T10:00:00Z',
    updated_at: '2026-02-20T10:00:00Z',
  },
  {
    id: 'ar222222-2222-2222-2222-222222222222',
    object_id: '11111111-1111-1111-1111-111111111111',
    name: 'Auto-approve update',
    description: 'Set status to approved on update',
    event_type: 'after_update',
    condition: "new.Status == 'Pending'",
    procedure_code: 'auto_approve',
    execution_mode: 'per_record',
    sort_order: 1,
    is_active: false,
    created_at: '2026-02-20T11:00:00Z',
    updated_at: '2026-02-20T11:00:00Z',
  },
]

export async function setupAutomationRuleRoutes(page: Page) {
  // List rules for an object
  for (const obj of mockObjects) {
    await page.route(`**/api/v1/admin/metadata/objects/${obj.id}/automation-rules`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: singleResponse(
            mockAutomationRules.filter((r) => r.object_id === obj.id),
          ),
        })
      }
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: singleResponse({ ...mockAutomationRules[0], id: 'new-rule-id' }),
        })
      }
      return route.continue()
    })
  }

  // Single rule CRUD
  for (const rule of mockAutomationRules) {
    await page.route(`**/api/v1/admin/metadata/automation-rules/${rule.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(rule) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(rule) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

// ─── Describe / Records mock data ─────────────────────────────

export const mockDescribeList = [
  {
    api_name: 'Account',
    label: 'Account',
    plural_label: 'Accounts',
    is_createable: true,
    is_queryable: true,
  },
  {
    api_name: 'Contact',
    label: 'Contact',
    plural_label: 'Contacts',
    is_createable: true,
    is_queryable: true,
  },
]

export const mockAccountDescribe = {
  api_name: 'Account',
  label: 'Account',
  plural_label: 'Accounts',
  is_createable: true,
  is_updateable: true,
  is_deleteable: true,
  fields: [
    {
      api_name: 'Id',
      label: 'ID',
      field_type: 'text',
      field_subtype: null,
      is_required: false,
      is_read_only: true,
      is_system_field: true,
      sort_order: -6,
      config: {},
    },
    {
      api_name: 'Name',
      label: 'Name',
      field_type: 'text',
      field_subtype: 'plain',
      is_required: true,
      is_read_only: false,
      is_system_field: false,
      sort_order: 1,
      config: { max_length: 255, default_value: null },
    },
    {
      api_name: 'Industry',
      label: 'Industry',
      field_type: 'picklist',
      field_subtype: 'single',
      is_required: false,
      is_read_only: false,
      is_system_field: false,
      sort_order: 2,
      config: {
        values: [
          { id: 'v1', value: 'Technology', label: 'Technology', sort_order: 1, is_default: false, is_active: true },
          { id: 'v2', value: 'Finance', label: 'Finance', sort_order: 2, is_default: false, is_active: true },
        ],
      },
    },
    {
      api_name: 'Phone',
      label: 'Phone',
      field_type: 'text',
      field_subtype: 'phone',
      is_required: false,
      is_read_only: false,
      is_system_field: false,
      sort_order: 3,
      config: {},
    },
  ],
}

export const mockRecords = [
  {
    Id: 'rec11111-1111-1111-1111-111111111111',
    Name: 'Acme Corp',
    Industry: 'Technology',
    Phone: '+380441234567',
  },
  {
    Id: 'rec22222-2222-2222-2222-222222222222',
    Name: 'Globex Inc',
    Industry: 'Finance',
    Phone: '+380509876543',
  },
]

export async function setupDescribeRoutes(page: Page) {
  await page.route('**/api/v1/describe', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: { data: mockDescribeList } })
    }
    return route.continue()
  })

  await page.route('**/api/v1/describe/Account', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: { data: mockAccountDescribe } })
    }
    return route.continue()
  })
}

export async function setupRecordRoutes(page: Page) {
  // List records
  await page.route('**/api/v1/records/Account?*', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({
        json: {
          data: mockRecords,
          pagination: { page: 1, per_page: 20, total: 2, total_pages: 1 },
        },
      })
    }
    return route.continue()
  })
  await page.route('**/api/v1/records/Account', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({
        json: {
          data: mockRecords,
          pagination: { page: 1, per_page: 20, total: 2, total_pages: 1 },
        },
      })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        status: 201,
        json: { data: { id: 'rec33333-3333-3333-3333-333333333333' } },
      })
    }
    return route.continue()
  })

  // Single record
  for (const rec of mockRecords) {
    await page.route(`**/api/v1/records/Account/${rec.Id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: rec } })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: { data: { success: true } } })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

// ─── Navigation mock data ───────────────────────────────────

export const mockProfileNavigations = [
  {
    id: 'nav11111-1111-1111-1111-111111111111',
    profile_id: 'prf11111-1111-1111-1111-111111111111',
    config: {
      groups: [
        {
          key: 'sales',
          label: 'Sales',
          icon: 'briefcase',
          items: [
            { type: 'object', object_api_name: 'Account' },
            { type: 'object', object_api_name: 'Contact' },
          ],
        },
      ],
    },
    created_at: '2026-02-25T10:00:00Z',
    updated_at: '2026-02-25T10:00:00Z',
  },
]

export const mockResolvedNavigation = {
  groups: [
    {
      key: 'sales',
      label: 'Sales',
      icon: 'briefcase',
      items: [
        { type: 'object', object_api_name: 'Account', label: 'Account', plural_label: 'Accounts' },
        { type: 'object', object_api_name: 'Contact', label: 'Contact', plural_label: 'Contacts' },
      ],
    },
  ],
}

export async function setupNavigationRoutes(page: Page) {
  await page.route('**/api/v1/admin/profile-navigation', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: { data: mockProfileNavigations } })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        status: 201,
        json: { data: { ...mockProfileNavigations[0], id: 'nav-new-1111-1111-1111-111111111111' } },
      })
    }
    return route.continue()
  })

  for (const nav of mockProfileNavigations) {
    await page.route(`**/api/v1/admin/profile-navigation/${nav.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: nav } })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: { data: nav } })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }

  await page.route('**/api/v1/navigation', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: { data: mockResolvedNavigation } })
    }
    return route.continue()
  })
}

// ─── Layouts mock data ──────────────────────────────────────

export const mockLayouts = [
  {
    id: 'ly111111-1111-1111-1111-111111111111',
    object_view_id: 'ov111111-1111-1111-1111-111111111111',
    form_factor: 'desktop',
    mode: 'edit',
    config: {
      root: {
        type: 'grid',
        columns: 2,
        children: [
          { type: 'highlights', key: 'highlights' },
          { type: 'field_section', key: 'general' },
        ],
      },
      section_config: {
        general: { columns: 2, collapsible: true },
      },
      field_config: {
        Name: { col_span: 2 },
        Industry: { col_span: 1 },
      },
      list_config: null,
    },
    created_at: '2026-02-26T10:00:00Z',
    updated_at: '2026-02-26T10:00:00Z',
  },
  {
    id: 'ly222222-2222-2222-2222-222222222222',
    object_view_id: 'ov222222-2222-2222-2222-222222222222',
    form_factor: 'mobile',
    mode: 'view',
    config: {
      root: null,
      section_config: {},
      field_config: {},
      list_config: {
        columns: [
          { field: 'Name', label: 'Name', width: '200px' },
          { field: 'Industry', label: 'Industry' },
        ],
      },
    },
    created_at: '2026-02-26T11:00:00Z',
    updated_at: '2026-02-26T11:00:00Z',
  },
]

// ─── Shared Layouts mock data ───────────────────────────────

export const mockSharedLayouts = [
  {
    id: 'sl111111-1111-1111-1111-111111111111',
    api_name: 'compact_address',
    type: 'field',
    label: 'Compact Address Fields',
    config: { col_span: 2, ui_kind: 'address' },
    created_at: '2026-02-26T10:00:00Z',
    updated_at: '2026-02-26T10:00:00Z',
  },
  {
    id: 'sl222222-2222-2222-2222-222222222222',
    api_name: 'sales_list',
    type: 'list',
    label: 'Sales List Config',
    config: { view: 'table', columns: [{ field: 'Name' }, { field: 'Amount' }] },
    created_at: '2026-02-26T11:00:00Z',
    updated_at: '2026-02-26T11:00:00Z',
  },
]

export async function setupLayoutRoutes(page: Page) {
  // List with optional query param
  await page.route('**/api/v1/admin/layouts?*', (route) => {
    const url = route.request().url()
    const ovId = new URL(url).searchParams.get('object_view_id')
    const filtered = ovId
      ? mockLayouts.filter((l) => l.object_view_id === ovId)
      : mockLayouts
    route.fulfill({ json: singleResponse(filtered) })
  })
  await page.route('**/api/v1/admin/layouts', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: singleResponse(mockLayouts) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockLayouts[0], id: 'new-layout-id' }),
      })
    }
    return route.continue()
  })
  for (const layout of mockLayouts) {
    await page.route(`**/api/v1/admin/layouts/${layout.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(layout) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(layout) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

export async function setupSharedLayoutRoutes(page: Page) {
  await page.route('**/api/v1/admin/shared-layouts?*', (route) => {
    route.fulfill({ json: singleResponse(mockSharedLayouts) })
  })
  await page.route('**/api/v1/admin/shared-layouts', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: singleResponse(mockSharedLayouts) })
    }
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: singleResponse({ ...mockSharedLayouts[0], id: 'new-shared-layout-id' }),
      })
    }
    return route.continue()
  })
  for (const sl of mockSharedLayouts) {
    await page.route(`**/api/v1/admin/shared-layouts/${sl.id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: singleResponse(sl) })
      }
      if (route.request().method() === 'PUT') {
        return route.fulfill({ json: singleResponse(sl) })
      }
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.continue()
    })
  }
}

export async function setupAllRoutes(page: Page) {
  await seedAuthToken(page)
  await setupAuthRoutes(page)
  await setupMetadataRoutes(page)
  await setupSecurityRoutes(page)
  await setupGroupRoutes(page)
  await setupSharingRuleRoutes(page)
  await setupTemplateRoutes(page)
  await setupTerritoryRoutes(page)
  await setupValidationRuleRoutes(page)
  await setupFunctionRoutes(page)
  await setupObjectViewRoutes(page)
  await setupProcedureRoutes(page)
  await setupCredentialRoutes(page)
  await setupAutomationRuleRoutes(page)
  await setupNavigationRoutes(page)
  await setupLayoutRoutes(page)
  await setupSharedLayoutRoutes(page)
  await setupDescribeRoutes(page)
  await setupRecordRoutes(page)
}
