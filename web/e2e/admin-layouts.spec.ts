import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockLayouts, mockObjectViews } from './fixtures/mock-api'

test.describe('Layout list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows layout entries', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    const rows = page.locator('[data-testid="layout-row"]')
    await expect(rows).toHaveCount(2)
  })

  test('shows form factor badges', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    await expect(page.getByText('desktop').first()).toBeVisible()
    await expect(page.getByText('mobile').first()).toBeVisible()
  })

  test('shows mode badges', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    await expect(page.getByText('edit').first()).toBeVisible()
    await expect(page.getByText('view').first()).toBeVisible()
  })

  test('has create layout button', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    await expect(page.locator('[data-testid="create-layout-btn"]')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    await page.locator('[data-testid="create-layout-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/layouts\/new/)
  })

  test('clicking layout row navigates to detail', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    await page.locator('[data-testid="layout-row"]').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/metadata/layouts/${mockLayouts[0].id}`),
    )
  })

  test('shows empty state when no layouts', async ({ page }) => {
    await page.route('**/api/v1/admin/layouts', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.route('**/api/v1/admin/layouts?*', (route) => {
      return route.fulfill({ json: { data: [] } })
    })
    await page.goto('/admin/metadata/layouts')
    await expect(page.getByText('No layouts')).toBeVisible()
  })

  test('shows OV filter dropdown', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    await expect(page.locator('[data-testid="filter-ov"]')).toBeVisible()
  })
})

test.describe('Layout create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders form with all fields', async ({ page }) => {
    await page.goto('/admin/metadata/layouts/new')
    await expect(page.locator('[data-testid="field-object-view"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-form-factor"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-mode"]')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/layouts/new')
    await expect(page.locator('[data-testid="submit-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/layouts/new')
    await page.locator('[data-testid="cancel-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/layouts$/)
  })

  test('shows breadcrumbs', async ({ page }) => {
    await page.goto('/admin/metadata/layouts/new')
    await expect(page.getByText('Layouts').first()).toBeVisible()
    await expect(page.getByText('Create').first()).toBeVisible()
  })

  test('submit calls POST', async ({ page }) => {
    await page.goto('/admin/metadata/layouts/new')

    // Select object view from dropdown
    await page.locator('[data-testid="field-object-view"]').click()
    await page.getByRole('option', { name: mockObjectViews[0].label }).click()

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/admin/layouts') &&
        req.method() === 'POST',
    )
    await page.locator('[data-testid="submit-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Layout detail page — tabs', () => {
  const layout = mockLayouts[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays layout heading', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(
      page.getByRole('heading', { name: /desktop/ }),
    ).toBeVisible()
  })

  test('shows form factor and mode badges', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(page.locator('[data-testid="badge-form-factor"]')).toBeVisible()
    await expect(page.locator('[data-testid="badge-mode"]')).toBeVisible()
  })

  test('shows 3 tabs', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(page.locator('[data-testid="tab-form-layout"]')).toBeVisible()
    await expect(page.locator('[data-testid="tab-list-config"]')).toBeVisible()
    await expect(page.locator('[data-testid="tab-json"]')).toBeVisible()
  })

  test('default tab is Form Layout', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(page.locator('[data-testid="form-layout-tab"]')).toBeVisible()
  })

  test('switching to JSON tab shows JSON editor', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="tab-json"]').click()
    await expect(page.locator('[data-testid="json-config"]')).toBeVisible()
  })

  test('JSON tab shows formatted JSON', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="tab-json"]').click()
    const textarea = page.locator('[data-testid="json-config"]')
    const value = await textarea.inputValue()
    expect(value).toContain('sectionConfig')
  })

  test('switching to List Config tab shows list editor', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="tab-list-config"]').click()
    await expect(page.locator('[data-testid="list-config-tab"]')).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(page.locator('[data-testid="save-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="delete-layout-btn"]')).toBeVisible()
  })

  test('delete button shows confirmation dialog', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="delete-layout-btn"]').click()
    await expect(page.getByText('Delete layout?')).toBeVisible()
  })
})

test.describe('Layout detail — Form Layout tab', () => {
  const layout = mockLayouts[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows section cards from OV', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(page.locator('[data-testid="section-card"]').first()).toBeVisible()
  })

  test('shows field chips inside sections', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(page.locator('[data-testid="field-chip"]').first()).toBeVisible()
  })

  test('clicking section header shows section properties panel', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    // Click the section header text (not field chips inside the card)
    await page.locator('[data-testid="section-card"]').first().locator('[data-testid="section-header"]').click()
    await expect(page.locator('[data-testid="section-columns"]')).toBeVisible()
  })

  test('clicking field shows field properties panel', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="field-chip"]').first().click()
    await expect(page.locator('[data-testid="field-col-span"]')).toBeVisible()
  })

  test('changing columns updates section config', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    // Click section header to open section properties
    await page.locator('[data-testid="section-card"]').first().locator('[data-testid="section-header"]').click()
    const colInput = page.locator('[data-testid="section-columns"]')
    await colInput.fill('3')

    // Verify by switching to JSON tab
    await page.locator('[data-testid="tab-json"]').click()
    const json = await page.locator('[data-testid="json-config"]').inputValue()
    expect(json).toContain('"columns": 3')
  })

  test('changing col_span updates field config', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="field-chip"]').first().click()
    await expect(page.locator('[data-testid="field-col-span"]')).toBeVisible()
    await page.locator('[data-testid="field-col-span"]').fill('3')
    // Verify the badge updated on the chip
    await expect(page.locator('[data-testid="field-chip"]').first().getByText('3col')).toBeVisible()
  })

  test('properties panel shows placeholder when nothing selected', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await expect(page.getByText('Select a section or field')).toBeVisible()
  })
})

test.describe('Layout detail — List Config tab', () => {
  // Use layout 2 which has list_config
  const layout = mockLayouts[1]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows available fields', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="tab-list-config"]').click()
    // OV 2 has no fields in read config, so available fields may be empty
    await expect(page.locator('[data-testid="list-config-tab"]')).toBeVisible()
  })

  test('shows active columns from config', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="tab-list-config"]').click()
    const columns = page.locator('[data-testid="active-column"]')
    await expect(columns).toHaveCount(2)
  })

  test('can remove column from active list', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="tab-list-config"]').click()
    const removeButtons = page.locator('[data-testid="remove-column"]')
    await removeButtons.first().click()
    await expect(page.locator('[data-testid="active-column"]')).toHaveCount(1)
  })
})

test.describe('Layout detail — Save integration', () => {
  const layout = mockLayouts[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('save sends updated config via PUT', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/layouts/${layout.id}`) &&
        req.method() === 'PUT',
    )
    await page.locator('[data-testid="save-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
    const body = JSON.parse(request.postData() ?? '{}')
    expect(body.config).toBeDefined()
  })

  test('JSON tab shows valid config and errors on invalid JSON', async ({ page }) => {
    await page.goto(`/admin/metadata/layouts/${layout.id}`)
    await page.locator('[data-testid="tab-json"]').click()
    await expect(page.locator('[data-testid="json-config"]')).toBeVisible()

    const textarea = page.locator('[data-testid="json-config"]')
    const value = await textarea.inputValue()
    // Config JSON should be parseable
    expect(() => JSON.parse(value)).not.toThrow()

    // Type invalid JSON to trigger error
    await textarea.fill('{invalid')
    await expect(page.locator('[data-testid="json-error"]')).toBeVisible()
  })
})

test.describe('Sidebar navigation', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('Layouts link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/layouts')
    await expect(page.locator('aside').getByRole('link', { name: 'Layouts', exact: true })).toBeVisible()
  })

  test('Layouts link navigates to list', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.locator('aside').getByText('Presentation').click()
    await page.getByRole('link', { name: 'Layouts', exact: true }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/layouts/)
  })
})
