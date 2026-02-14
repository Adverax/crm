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
    await expect(main.getByText('4 объектов, 36 полей')).toBeVisible()
    await expect(main.getByText('4 объектов, 28 полей')).toBeVisible()
  })

  test('has apply buttons for each template', async ({ page }) => {
    await page.goto('/admin/templates')
    const buttons = page.getByRole('button', { name: 'Применить' })
    await expect(buttons).toHaveCount(2)
  })

  test('clicking apply sends POST request', async ({ page }) => {
    await page.goto('/admin/templates')

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/admin/templates/sales_crm/apply') &&
        req.method() === 'POST',
    )

    const buttons = page.getByRole('button', { name: 'Применить' })
    await buttons.first().click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })

  test('shows page header', async ({ page }) => {
    await page.goto('/admin/templates')
    await expect(
      page.getByRole('heading', { name: 'Шаблоны приложений' }),
    ).toBeVisible()
  })

  test('shows description text', async ({ page }) => {
    await page.goto('/admin/templates')
    await expect(
      page.getByText('Выберите шаблон для создания стандартных объектов и полей'),
    ).toBeVisible()
  })
})

test.describe('Sidebar navigation — Templates', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('sidebar has Шаблоны link', async ({ page }) => {
    await page.goto('/admin/templates')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Шаблоны')).toBeVisible()
  })

  test('clicking Шаблоны navigates to templates page', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Шаблоны').click()
    await expect(page).toHaveURL(/\/admin\/templates/)
  })
})
