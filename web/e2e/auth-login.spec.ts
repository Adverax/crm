import { test, expect } from '@playwright/test'
import { setupAuthRoutes, setupAllRoutes, setupMetadataRoutes } from './fixtures/mock-api'

test.describe('Login page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthRoutes(page)
  })

  test('renders login form', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByRole('heading', { name: 'Sign in to CRM' })).toBeVisible()
    await expect(page.locator('#username')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible()
  })

  test('has forgot password link', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByText('Forgot password?')).toBeVisible()
  })

  test('forgot password link navigates to forgot page', async ({ page }) => {
    await page.goto('/login')
    await page.getByText('Forgot password?').click()
    await expect(page).toHaveURL(/\/forgot-password/)
  })

  test('shows error on failed login', async ({ page }) => {
    await page.route('**/api/v1/auth/login', (route) => {
      return route.fulfill({
        status: 401,
        json: { error: { code: 'UNAUTHORIZED', message: 'Invalid credentials' } },
      })
    })

    await page.goto('/login')
    await page.locator('#username').fill('wrong')
    await page.locator('#password').fill('wrongpass')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page.getByText('Invalid credentials')).toBeVisible()
  })

  test('submits login and redirects to admin', async ({ page }) => {
    // Setup metadata routes so the redirect target (/admin/metadata/objects) works
    await setupMetadataRoutes(page)

    await page.goto('/login')
    await page.locator('#username').fill('admin')
    await page.locator('#password').fill('password123')

    const requestPromise = page.waitForRequest('**/api/v1/auth/login')
    await page.getByRole('button', { name: 'Sign in' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')

    await expect(page).toHaveURL(/\/admin/)
  })

  test('redirects unauthenticated users from admin to login', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await expect(page).toHaveURL(/\/login/)
  })

  test('redirects authenticated users from login to app', async ({ page }) => {
    await setupAllRoutes(page)
    await page.goto('/login')
    await expect(page).toHaveURL(/\/app/)
  })
})

test.describe('Forgot password page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthRoutes(page)
  })

  test('renders forgot password form', async ({ page }) => {
    await page.goto('/forgot-password')
    await expect(page.getByRole('heading', { name: 'Reset Password' })).toBeVisible()
    await expect(page.locator('#email')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Submit' })).toBeVisible()
  })

  test('has back to login link', async ({ page }) => {
    await page.goto('/forgot-password')
    await expect(page.getByText('Back to login')).toBeVisible()
  })

  test('submits email and shows success', async ({ page }) => {
    await page.goto('/forgot-password')
    await page.locator('#email').fill('admin@example.com')

    const requestPromise = page.waitForRequest('**/api/v1/auth/forgot-password')
    await page.getByRole('button', { name: 'Submit' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')

    await expect(page.getByText('password reset link has been sent')).toBeVisible()
  })
})

test.describe('Reset password page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthRoutes(page)
  })

  test('renders reset form with valid token', async ({ page }) => {
    await page.goto('/reset-password?token=valid-token')
    await expect(page.getByRole('heading', { name: 'New Password' })).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('#confirmPassword')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Reset Password' })).toBeVisible()
  })

  test('shows error when no token', async ({ page }) => {
    await page.goto('/reset-password')
    await expect(page.getByText('reset link is invalid')).toBeVisible()
  })

  test('shows error on password mismatch', async ({ page }) => {
    await page.goto('/reset-password?token=valid-token')
    await page.locator('#password').fill('newpassword1')
    await page.locator('#confirmPassword').fill('different')
    await page.getByRole('button', { name: 'Reset Password' }).click()

    await expect(page.getByText('Passwords do not match')).toBeVisible()
  })

  test('shows error on short password', async ({ page }) => {
    await page.goto('/reset-password?token=valid-token')
    await page.locator('#password').fill('short')
    await page.locator('#confirmPassword').fill('short')
    await page.getByRole('button', { name: 'Reset Password' }).click()

    await expect(page.getByText('at least 8 characters')).toBeVisible()
  })

  test('submits and shows success', async ({ page }) => {
    await page.goto('/reset-password?token=valid-token')
    await page.locator('#password').fill('newpassword123')
    await page.locator('#confirmPassword').fill('newpassword123')

    const requestPromise = page.waitForRequest('**/api/v1/auth/reset-password')
    await page.getByRole('button', { name: 'Reset Password' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')

    await expect(page.getByText('Password changed successfully')).toBeVisible()
  })
})
