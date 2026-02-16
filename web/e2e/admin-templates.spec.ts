import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockTemplates } from './fixtures/mock-api'

test.describe('Template list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays template labels', async ({ page }) => {
    await page.goto('/admin/templates')
    const main = page.locator('main')
    await expect(main.getByText('Sales CRM')).toBeVisible()
    await expect(main.getByText('Recruiting')).toBeVisible()
  })

  test('displays template descriptions', async ({ page }) => {
    await page.goto('/admin/templates')
    const main = page.locator('main')
    await expect(main.getByText(mockTemplates[0].description)).toBeVisible()
    await expect(main.getByText(mockTemplates[1].description)).toBeVisible()
  })

  test('displays object and field counts', async ({ page }) => {
    await page.goto('/admin/templates')
    const main = page.locator('main')
    await expect(main.getByText('4 objects, 36 fields')).toBeVisible()
    await expect(main.getByText('4 objects, 28 fields')).toBeVisible()
  })

  test('has apply buttons for each template', async ({ page }) => {
    await page.goto('/admin/templates')
    const buttons = page.getByRole('button', { name: 'Apply' })
    await expect(buttons).toHaveCount(2)
  })

  test('clicking apply sends POST request', async ({ page }) => {
    await page.goto('/admin/templates')

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/admin/templates/sales_crm/apply') &&
        req.method() === 'POST',
    )

    const buttons = page.getByRole('button', { name: 'Apply' })
    await buttons.first().click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })

  test('shows page header', async ({ page }) => {
    await page.goto('/admin/templates')
    await expect(
      page.getByRole('heading', { name: 'App Templates' }),
    ).toBeVisible()
  })

  test('shows description text', async ({ page }) => {
    await page.goto('/admin/templates')
    await expect(
      page.getByText('Choose a template to create standard objects and fields'),
    ).toBeVisible()
  })
})

test.describe('Sidebar navigation — Templates', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('sidebar has Templates link', async ({ page }) => {
    await page.goto('/admin/templates')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Templates')).toBeVisible()
  })

  test('clicking Templates navigates to templates page', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Templates').click()
    await expect(page).toHaveURL(/\/admin\/templates/)
  })
})
