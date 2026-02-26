import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockSharedLayouts } from './fixtures/mock-api'

test.beforeEach(async ({ page }) => {
  await setupAllRoutes(page)
})

// ─── List page ───────────────────────────────────────────────

test.describe('Shared Layout list page', () => {
  test('shows shared layout entries', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts')
    await expect(page.getByTestId('shared-layout-row')).toHaveCount(mockSharedLayouts.length)
    await expect(page.getByText('Compact Address Fields')).toBeVisible()
    await expect(page.getByText('Sales List Config')).toBeVisible()
  })

  test('shows api names in list', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts')
    await expect(page.getByText('compact_address')).toBeVisible()
    await expect(page.getByText('sales_list')).toBeVisible()
  })

  test('shows type badges', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts')
    await expect(page.getByText('field').first()).toBeVisible()
    await expect(page.getByText('list').first()).toBeVisible()
  })

  test('has create button', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts')
    await expect(page.getByTestId('create-shared-layout-btn')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts')
    await page.getByTestId('create-shared-layout-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/shared-layouts\/new/)
  })

  test('clicking row navigates to detail page', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts')
    await page.getByTestId('shared-layout-row').first().click()
    await expect(page).toHaveURL(/\/admin\/metadata\/shared-layouts\/sl111111/)
  })

  test('shows empty state when no shared layouts', async ({ page }) => {
    await page.route('**/api/v1/admin/shared-layouts', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.goto('/admin/metadata/shared-layouts')
    await expect(page.getByText('No shared layouts')).toBeVisible()
  })
})

// ─── Create page ─────────────────────────────────────────────

test.describe('Shared Layout create page', () => {
  test('renders form with all fields', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts/new')
    await expect(page.getByTestId('field-api-name')).toBeVisible()
    await expect(page.getByTestId('field-label')).toBeVisible()
    await expect(page.getByTestId('field-type')).toBeVisible()
    await expect(page.getByTestId('field-config')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts/new')
    await expect(page.getByTestId('submit-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts/new')
    await page.getByTestId('cancel-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/shared-layouts$/)
  })

  test('shows breadcrumbs', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts/new')
    await expect(page.getByText('Shared Layouts').first()).toBeVisible()
    await expect(page.getByText('Create').first()).toBeVisible()
  })

  test('submit calls POST', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts/new')

    await page.getByTestId('field-api-name').fill('test_layout')
    await page.getByTestId('field-label').fill('Test Layout')

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/admin/shared-layouts') &&
        req.method() === 'POST',
    )
    await page.getByTestId('submit-btn').click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

// ─── Detail page ─────────────────────────────────────────────

test.describe('Shared Layout detail page', () => {
  const sl = mockSharedLayouts[0]
  const detailUrl = `/admin/metadata/shared-layouts/${sl.id}`

  test('loads and displays heading with label', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(
      page.getByRole('heading', { name: sl.label }),
    ).toBeVisible()
  })

  test('shows api name display', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('display-api-name')).toContainText(sl.api_name)
  })

  test('shows type badge', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('badge-type')).toContainText(sl.type)
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('save-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
    await expect(page.getByTestId('delete-shared-layout-btn')).toBeVisible()
  })

  test('submit calls PUT', async ({ page }) => {
    await page.goto(detailUrl)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/shared-layouts/${sl.id}`) &&
        req.method() === 'PUT',
    )
    await page.getByTestId('save-btn').click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })

  test('delete button shows confirmation dialog', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('delete-shared-layout-btn').click()
    await expect(page.getByText('Delete shared layout?')).toBeVisible()
  })
})

// ─── Sidebar ─────────────────────────────────────────────────

test.describe('Sidebar navigation', () => {
  test('Shared Layouts link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/shared-layouts')
    await expect(page.locator('aside').getByRole('link', { name: 'Shared Layouts' })).toBeVisible()
  })

  test('Shared Layouts link navigates to list', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.locator('aside').getByText('Presentation').click()
    await page.getByRole('link', { name: 'Shared Layouts' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/shared-layouts/)
  })
})
