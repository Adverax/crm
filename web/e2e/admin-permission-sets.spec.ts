import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockPermissionSets } from './fixtures/mock-api'

test.describe('Permission set list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays permission set list', async ({ page }) => {
    await page.goto('/admin/security/permission-sets')
    const main = page.locator('main')
    await expect(main.getByText('read_all')).toBeVisible()
    await expect(main.getByText('deny_delete')).toBeVisible()
  })

  test('shows permission set labels', async ({ page }) => {
    await page.goto('/admin/security/permission-sets')
    const main = page.locator('main')
    await expect(main.getByText('Read All')).toBeVisible()
    await expect(main.getByText('Deny Delete')).toBeVisible()
  })

  test('has create button', async ({ page }) => {
    await page.goto('/admin/security/permission-sets')
    await expect(page.getByText('Create Permission Set')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/permission-sets')
    await page.getByText('Create Permission Set').click()
    await expect(page).toHaveURL(/\/admin\/security\/permission-sets\/new/)
  })

  test('clicking permission set navigates to detail', async ({ page }) => {
    await page.goto('/admin/security/permission-sets')
    await page.locator('main').getByText('read_all').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/security/permission-sets/${mockPermissionSets[0].id}`),
    )
  })

  test('shows type badges (grant/deny)', async ({ page }) => {
    await page.goto('/admin/security/permission-sets')
    const main = page.locator('main')
    await expect(main.getByText('grant').first()).toBeVisible()
    await expect(main.getByText('deny').first()).toBeVisible()
  })
})

test.describe('Permission set create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form', async ({ page }) => {
    await page.goto('/admin/security/permission-sets/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('has type selector (grant/deny)', async ({ page }) => {
    await page.goto('/admin/security/permission-sets/new')
    await expect(page.getByText('Type').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/permission-sets/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/permission-sets')
    await page.getByText('Create Permission Set').click()
    await expect(page).toHaveURL(/\/admin\/security\/permission-sets\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/permission-sets/)
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/security/permission-sets/new')

    await page.locator('#apiName').fill('new_ps')
    await page.locator('#label').fill('New Permission Set')
    await page.locator('#description').fill('Description')

    const requestPromise = page.waitForRequest(
      '**/api/v1/admin/security/permission-sets',
    )
    await page.getByRole('button', { name: 'Create' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Permission set detail page', () => {
  const ps = mockPermissionSets[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays permission set heading', async ({ page }) => {
    await page.goto(`/admin/security/permission-sets/${ps.id}`)
    await expect(
      page.getByRole('heading', { name: ps.label }),
    ).toBeVisible()
  })

  test('shows tabs: info, OLS, FLS', async ({ page }) => {
    await page.goto(`/admin/security/permission-sets/${ps.id}`)
    await expect(page.getByRole('tab', { name: /General/ })).toBeVisible()
    await expect(
      page.getByRole('tab', { name: /Object Permissions/ }),
    ).toBeVisible()
    await expect(
      page.getByRole('tab', { name: /Field Permissions/ }),
    ).toBeVisible()
  })

  test('info tab shows editable form', async ({ page }) => {
    await page.goto(`/admin/security/permission-sets/${ps.id}`)
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('can switch to OLS tab', async ({ page }) => {
    await page.goto(`/admin/security/permission-sets/${ps.id}`)
    await page.getByRole('tab', { name: /Object Permissions/ }).click()
    await expect(page.locator('main')).toBeVisible()
  })

  test('can switch to FLS tab', async ({ page }) => {
    await page.goto(`/admin/security/permission-sets/${ps.id}`)
    await page.getByRole('tab', { name: /Field Permissions/ }).click()
    await expect(page.locator('main')).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/security/permission-sets/${ps.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('can submit updated permission set', async ({ page }) => {
    await page.goto(`/admin/security/permission-sets/${ps.id}`)
    await page.locator('#label').fill('Updated Permission Set')

    const requestPromise = page.waitForRequest(
      `**/api/v1/admin/security/permission-sets/${ps.id}`,
    )
    await page.getByRole('button', { name: 'Save' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })
})
