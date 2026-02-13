import { type Page } from '@playwright/test'

// ─── Mock data ──────────────────────────────────────────────

export const mockObjects = [
  {
    id: '11111111-1111-1111-1111-111111111111',
    api_name: 'account',
    label: 'Аккаунт',
    plural_label: 'Аккаунты',
    object_type: 'standard',
    description: 'Компании и организации',
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
    label: 'Кастомный объект',
    plural_label: 'Кастомные объекты',
    object_type: 'custom',
    description: 'Пользовательский объект',
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
    label: 'Название',
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
    label: 'Генеральный директор',
    parent_id: null,
    description: 'Главная роль',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'r2222222-2222-2222-2222-222222222222',
    api_name: 'sales_manager',
    label: 'Менеджер по продажам',
    parent_id: 'r1111111-1111-1111-1111-111111111111',
    description: 'Управляет отделом продаж',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockPermissionSets = [
  {
    id: 'ps111111-1111-1111-1111-111111111111',
    api_name: 'read_all',
    label: 'Чтение всего',
    ps_type: 'grant',
    description: 'Даёт права на чтение',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'ps222222-2222-2222-2222-222222222222',
    api_name: 'deny_delete',
    label: 'Запрет удаления',
    ps_type: 'deny',
    description: 'Запрещает удаление',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
]

export const mockProfiles = [
  {
    id: 'pf111111-1111-1111-1111-111111111111',
    api_name: 'system_admin',
    label: 'Системный администратор',
    description: 'Полный доступ',
    base_permission_set_id: 'ps111111-1111-1111-1111-111111111111',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'pf222222-2222-2222-2222-222222222222',
    api_name: 'standard_user',
    label: 'Стандартный пользователь',
    description: 'Базовый доступ',
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
    first_name: 'Иван',
    last_name: 'Иванов',
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
    first_name: 'Пётр',
    last_name: 'Петров',
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
    label: 'Все пользователи',
    group_type: 'public',
    related_role_id: null,
    related_user_id: null,
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'g2222222-2222-2222-2222-222222222222',
    api_name: 'sales_team',
    label: 'Отдел продаж',
    group_type: 'public',
    related_role_id: null,
    related_user_id: null,
    created_at: '2026-02-01T12:00:00Z',
    updated_at: '2026-02-01T12:00:00Z',
  },
  {
    id: 'g3333333-3333-3333-3333-333333333333',
    api_name: 'ceo_role',
    label: 'Роль CEO',
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
  first_name: 'Иван',
  last_name: 'Иванов',
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

// ─── Territory mock data ─────────────────────────────────────

export const mockTerritoryModels = [
  {
    id: 'tm111111-1111-1111-1111-111111111111',
    api_name: 'q1_2026',
    label: 'Q1 2026',
    description: 'Территориальная модель первого квартала',
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
    description: 'Территориальная модель четвёртого квартала',
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
    label: 'Северная Америка',
    description: 'Регион Северной Америки',
    created_at: '2026-01-16T10:00:00Z',
    updated_at: '2026-01-16T10:00:00Z',
  },
  {
    id: 'tt222222-2222-2222-2222-222222222222',
    model_id: 'tm111111-1111-1111-1111-111111111111',
    parent_id: 'tt111111-1111-1111-1111-111111111111',
    api_name: 'us_east',
    label: 'Восток США',
    description: 'Восточное побережье',
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

export async function setupAllRoutes(page: Page) {
  await seedAuthToken(page)
  await setupAuthRoutes(page)
  await setupMetadataRoutes(page)
  await setupSecurityRoutes(page)
  await setupGroupRoutes(page)
  await setupSharingRuleRoutes(page)
  await setupTerritoryRoutes(page)
}
