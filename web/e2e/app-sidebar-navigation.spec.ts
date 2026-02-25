import { test, expect } from '@playwright/test'
import { setupAllRoutes } from './fixtures/mock-api'

test.beforeEach(async ({ page }) => {
  await setupAllRoutes(page)
})

test.describe('App sidebar navigation', () => {
  test('shows grouped navigation when config exists', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByText('Sales')).toBeVisible()
    await expect(page.getByText('Accounts')).toBeVisible()
  })

  test('clicking object navigates to list page', async ({ page }) => {
    await page.goto('/app')
    await page.getByText('Accounts').click()
    await expect(page).toHaveURL(/\/app\/Account/)
  })

  test('shows CRM brand link', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByText('CRM').first()).toBeVisible()
  })

  test('shows Settings link', async ({ page }) => {
    await page.goto('/app')
    await expect(page.getByText('Settings')).toBeVisible()
  })

  test('falls back to flat list when no navigation config', async ({ page }) => {
    await page.route('**/api/v1/navigation', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: { data: { groups: [{ key: '_default', label: '', items: [] }] } },
        })
      }
      return route.continue()
    })
    await page.goto('/app')
    // In fallback mode, it uses the describe endpoint
    await expect(page.getByText('Settings')).toBeVisible()
  })
})
