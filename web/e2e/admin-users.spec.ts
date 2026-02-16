import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockUsers } from './fixtures/mock-api'

test.describe('User list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays user usernames', async ({ page }) => {
    await page.goto('/admin/security/users')
    const main = page.locator('main')
    // Use exact match to avoid matching "CRM Admin" in sidebar
    await expect(main.getByText('admin', { exact: true })).toBeVisible()
    await expect(main.getByText('user1', { exact: true })).toBeVisible()
  })

  test('shows user emails', async ({ page }) => {
    await page.goto('/admin/security/users')
    const main = page.locator('main')
    await expect(main.getByText('admin@example.com')).toBeVisible()
    await expect(main.getByText('user1@example.com')).toBeVisible()
  })

  test('shows user names', async ({ page }) => {
    await page.goto('/admin/security/users')
    const main = page.locator('main')
    await expect(main.getByText('John Smith')).toBeVisible()
    await expect(main.getByText('Peter Johnson')).toBeVisible()
  })

  test('has create user button', async ({ page }) => {
    await page.goto('/admin/security/users')
    await expect(page.getByText('Create User')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/users')
    await page.getByText('Create User').click()
    await expect(page).toHaveURL(/\/admin\/security\/users\/new/)
  })

  test('clicking user row navigates to detail', async ({ page }) => {
    await page.goto('/admin/security/users')
    // Click on the email which is unique
    await page.locator('main').getByText('admin@example.com').click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/security/users/${mockUsers[0].id}`),
    )
  })

  test('shows active/inactive status badges', async ({ page }) => {
    await page.goto('/admin/security/users')
    const main = page.locator('main')
    await expect(main.getByText('Active', { exact: true })).toBeVisible()
    await expect(main.getByText('Inactive')).toBeVisible()
  })
})

test.describe('User create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders credentials card', async ({ page }) => {
    await page.goto('/admin/security/users/new')
    await expect(page.locator('#username')).toBeVisible()
    await expect(page.locator('#email')).toBeVisible()
  })

  test('renders personal data card', async ({ page }) => {
    await page.goto('/admin/security/users/new')
    await expect(page.locator('#firstName')).toBeVisible()
    await expect(page.locator('#lastName')).toBeVisible()
  })

  test('renders security card with selectors', async ({ page }) => {
    await page.goto('/admin/security/users/new')
    await expect(page.getByText('Profile').first()).toBeVisible()
    await expect(page.getByText('Role').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/users/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/users')
    await page.getByText('Create User').click()
    await expect(page).toHaveURL(/\/admin\/security\/users\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/users/)
  })

  test('shows validation error when profile not selected', async ({ page }) => {
    await page.goto('/admin/security/users/new')

    await page.locator('#username').fill('test_user')
    await page.locator('#email').fill('test@example.com')
    await page.locator('#firstName').fill('Test')
    await page.locator('#lastName').fill('User')

    await page.getByRole('button', { name: 'Create' }).click()
    await expect(page.getByText('Profile is required')).toBeVisible()
  })
})

test.describe('User detail page', () => {
  const user = mockUsers[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays user heading', async ({ page }) => {
    await page.goto(`/admin/security/users/${user.id}`)
    await expect(
      page.getByRole('heading', { name: user.username }),
    ).toBeVisible()
  })

  test('shows tabs: info and permission sets', async ({ page }) => {
    await page.goto(`/admin/security/users/${user.id}`)
    await expect(page.getByRole('tab', { name: /General/ })).toBeVisible()
    await expect(
      page.getByRole('tab', { name: /Permission Sets/ }),
    ).toBeVisible()
  })

  test('info tab shows editable fields', async ({ page }) => {
    await page.goto(`/admin/security/users/${user.id}`)
    await expect(page.locator('#email')).toBeVisible()
    await expect(page.locator('#firstName')).toBeVisible()
    await expect(page.locator('#lastName')).toBeVisible()
  })

  test('can switch to permission sets tab', async ({ page }) => {
    await page.goto(`/admin/security/users/${user.id}`)
    await page.getByRole('tab', { name: /Permission Sets/ }).click()
    await expect(page.locator('main')).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/security/users/${user.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('can submit updated user', async ({ page }) => {
    await page.goto(`/admin/security/users/${user.id}`)
    await page.locator('#firstName').fill('Updated Name')

    const requestPromise = page.waitForRequest(
      `**/api/v1/admin/security/users/${user.id}`,
    )
    await page.getByRole('button', { name: 'Save' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })
})
