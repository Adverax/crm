import { test, expect } from '@playwright/test'
import { setupAllRoutes } from './fixtures/mock-api'

test.beforeEach(async ({ page }) => {
  await setupAllRoutes(page)
})

test.describe('App dashboard', () => {
  test('shows dashboard title', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByTestId('dashboard-title')).toContainText('Dashboard')
  })

  test('renders widget grid', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByTestId('dashboard-grid')).toBeVisible()
  })

  test('renders metric widget with value', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByTestId('metric-widget')).toBeVisible()
    await expect(page.getByText('42')).toBeVisible()
  })

  test('renders list widget with records', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByTestId('list-widget')).toBeVisible()
    await expect(page.getByText('Call client')).toBeVisible()
  })

  test('renders link list widget', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByTestId('link-list-widget')).toBeVisible()
    await expect(page.getByText('New Account')).toBeVisible()
  })

  test('shows empty state when no dashboard config', async ({ page }) => {
    await page.route('**/api/v1/dashboard', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: { widgets: [] } } })
      }
      return route.continue()
    })
    await page.goto('/app')
    await expect(page.getByTestId('dashboard-empty')).toBeVisible()
  })
})
