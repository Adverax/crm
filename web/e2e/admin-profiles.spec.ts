import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockProfiles } from './fixtures/mock-api'

test.describe('Profile list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays profile list', async ({ page }) => {
    await page.goto('/admin/security/profiles')
    const main = page.locator('main')
    await expect(main.getByText('system_admin')).toBeVisible()
    await expect(main.getByText('standard_user')).toBeVisible()
  })

  test('shows profile labels', async ({ page }) => {
    await page.goto('/admin/security/profiles')
    const main = page.locator('main')
    await expect(main.getByText('System Administrator')).toBeVisible()
    await expect(main.getByText('Standard User')).toBeVisible()
  })

  test('has create button', async ({ page }) => {
    await page.goto('/admin/security/profiles')
    await expect(page.getByText('Create Profile')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/profiles')
    await page.getByText('Create Profile').click()
    await expect(page).toHaveURL(/\/admin\/security\/profiles\/new/)
  })

  test('clicking profile navigates to detail', async ({ page }) => {
    await page.goto('/admin/security/profiles')
    await page.locator('main').getByText('system_admin').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/security/profiles/${mockProfiles[0].id}`),
    )
  })
})

test.describe('Profile create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form', async ({ page }) => {
    await page.goto('/admin/security/profiles/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/profiles/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/profiles')
    await page.getByText('Create Profile').click()
    await expect(page).toHaveURL(/\/admin\/security\/profiles\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/profiles/)
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/security/profiles/new')

    await page.locator('#apiName').fill('new_profile')
    await page.locator('#label').fill('New Profile')
    await page.locator('#description').fill('Description')

    const requestPromise = page.waitForRequest(
      '**/api/v1/admin/security/profiles',
    )
    await page.getByRole('button', { name: 'Create' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Profile detail page', () => {
  const profile = mockProfiles[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays profile heading', async ({ page }) => {
    await page.goto(`/admin/security/profiles/${profile.id}`)
    await expect(
      page.getByRole('heading', { name: profile.label }),
    ).toBeVisible()
  })

  test('shows editable form fields', async ({ page }) => {
    await page.goto(`/admin/security/profiles/${profile.id}`)
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('shows link to base permission set', async ({ page }) => {
    await page.goto(`/admin/security/profiles/${profile.id}`)
    await expect(
      page.getByText('Open Base Permission Set'),
    ).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/security/profiles/${profile.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('can submit updated profile', async ({ page }) => {
    await page.goto(`/admin/security/profiles/${profile.id}`)
    await page.locator('#label').fill('Updated Profile')

    const requestPromise = page.waitForRequest(
      `**/api/v1/admin/security/profiles/${profile.id}`,
    )
    await page.getByRole('button', { name: 'Save' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })
})
