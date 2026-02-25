import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockProfileDashboards } from './fixtures/mock-api'

test.beforeEach(async ({ page }) => {
  await setupAllRoutes(page)
})

// ─── List page ───────────────────────────────────────────────

test.describe('Dashboard list page', () => {
  test('shows dashboard configs from API', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards')
    await expect(page.getByTestId('dash-row')).toHaveCount(mockProfileDashboards.length)
    await expect(page.getByText('2 widget(s)')).toBeVisible()
  })

  test('shows create button', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards')
    await expect(page.getByTestId('create-dash-btn')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards')
    await page.getByTestId('create-dash-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/dashboards\/new/)
  })

  test('clicking row navigates to detail page', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards')
    await page.getByTestId('dash-row').first().click()
    await expect(page).toHaveURL(/\/admin\/metadata\/dashboards\/dsh11111/)
  })

  test('shows empty state when no dashboard configs', async ({ page }) => {
    await page.route('**/api/v1/admin/profile-dashboards', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.goto('/admin/metadata/dashboards')
    await expect(page.getByText('No dashboard configs')).toBeVisible()
  })
})

// ─── Create page ─────────────────────────────────────────────

test.describe('Dashboard create page', () => {
  test('renders form with fields', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards/new')
    await expect(page.getByTestId('field-profile-id')).toBeVisible()
    await expect(page.getByTestId('field-config')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards/new')
    await expect(page.getByTestId('submit-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards/new')
    await page.getByTestId('cancel-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/dashboards$/)
  })

  test('submitting form calls POST', async ({ page }) => {
    await page.goto('/admin/metadata/dashboards/new')
    await page.getByTestId('field-profile-id').fill('prf11111-1111-1111-1111-111111111111')
    await page.getByTestId('submit-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/dashboards\/dsh/)
  })
})

// ─── Detail page ─────────────────────────────────────────────

test.describe('Dashboard detail page', () => {
  const detailUrl = `/admin/metadata/dashboards/${mockProfileDashboards[0]!.id}`

  test('loads and shows profile ID', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('profile-id')).toContainText('prf11111')
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('save-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
    await expect(page.getByTestId('delete-btn')).toBeVisible()
  })

  test('shows config in JSON editor', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('field-config')).toBeVisible()
  })
})
