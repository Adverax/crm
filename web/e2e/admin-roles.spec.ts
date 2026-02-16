import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockRoles } from './fixtures/mock-api'

test.describe('Role list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays role api names', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const main = page.locator('main')
    await expect(main.getByText('ceo', { exact: true })).toBeVisible()
    await expect(main.getByText('sales_manager', { exact: true })).toBeVisible()
  })

  test('shows role labels in table', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const main = page.locator('main')
    // "CEO" appears twice: as label for ceo row and as parent for sales_manager row
    await expect(main.getByText('Sales Manager')).toBeVisible()
  })

  test('has create role button', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await expect(page.getByText('Create Role')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await page.getByText('Create Role').click()
    await expect(page).toHaveURL(/\/admin\/security\/roles\/new/)
  })

  test('clicking role navigates to detail', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await page.locator('main').getByText('ceo').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/security/roles/${mockRoles[0].id}`),
    )
  })
})

test.describe('Role create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form', async ({ page }) => {
    await page.goto('/admin/security/roles/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('has parent role selector', async ({ page }) => {
    await page.goto('/admin/security/roles/new')
    await expect(page.getByText('Parent Role').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/roles/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await page.getByText('Create Role').click()
    await expect(page).toHaveURL(/\/admin\/security\/roles\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/roles/)
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/security/roles/new')

    await page.locator('#apiName').fill('new_role')
    await page.locator('#label').fill('New Role')
    await page.locator('#description').fill('New role description')

    const requestPromise = page.waitForRequest('**/api/v1/admin/security/roles')
    await page.getByRole('button', { name: 'Create' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Role detail page', () => {
  const role = mockRoles[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays role heading', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await expect(
      page.getByRole('heading', { name: role.label }),
    ).toBeVisible()
  })

  test('shows form with editable fields', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('can submit updated role', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await page.locator('#label').fill('Updated Role')

    const requestPromise = page.waitForRequest(
      `**/api/v1/admin/security/roles/${role.id}`,
    )
    await page.getByRole('button', { name: 'Save' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })
})
