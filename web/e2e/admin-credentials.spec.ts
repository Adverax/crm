import { test, expect } from '@playwright/test'
import {
  setupAllRoutes,
  mockCredentials,
  mockCredentialUsageLog,
} from './fixtures/mock-api'

test.beforeEach(async ({ page }) => {
  await setupAllRoutes(page)
})

// ─── List page ───────────────────────────────────────────────

test.describe('Credential list page', () => {
  test('shows credentials from API', async ({ page }) => {
    await page.goto('/admin/metadata/credentials')
    await expect(page.getByTestId('credential-row')).toHaveCount(mockCredentials.length)
    await expect(page.getByText('stripe_api')).toBeVisible()
    await expect(page.getByText('slack_oauth')).toBeVisible()
  })

  test('shows create button', async ({ page }) => {
    await page.goto('/admin/metadata/credentials')
    await expect(page.getByTestId('create-credential-btn')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/credentials')
    await page.getByTestId('create-credential-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/credentials\/new/)
  })

  test('clicking row navigates to detail page', async ({ page }) => {
    await page.goto('/admin/metadata/credentials')
    await page.getByTestId('credential-row').first().click()
    await expect(page).toHaveURL(/\/admin\/metadata\/credentials\/cr111111/)
  })

  test('shows empty state when no credentials', async ({ page }) => {
    await page.route('**/api/v1/admin/credentials', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.goto('/admin/metadata/credentials')
    await expect(page.getByText('No credentials')).toBeVisible()
  })
})

// ─── Create page ─────────────────────────────────────────────

test.describe('Credential create page', () => {
  test('renders form with fields', async ({ page }) => {
    await page.goto('/admin/metadata/credentials/new')
    await expect(page.getByTestId('field-code')).toBeVisible()
    await expect(page.getByTestId('field-name')).toBeVisible()
    await expect(page.getByTestId('field-base-url')).toBeVisible()
    await expect(page.getByTestId('field-type')).toBeVisible()
  })

  test('shows type-dependent fields for api_key', async ({ page }) => {
    await page.goto('/admin/metadata/credentials/new')
    // Default type is api_key
    await expect(page.getByTestId('field-header')).toBeVisible()
    await expect(page.getByTestId('field-value')).toBeVisible()
  })

  test('has create and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/credentials/new')
    await expect(page.getByTestId('submit-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/credentials/new')
    await page.getByTestId('cancel-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/credentials$/)
  })

  test('submitting form sends POST request', async ({ page }) => {
    let postCalled = false
    await page.route('**/api/v1/admin/credentials', (route) => {
      if (route.request().method() === 'POST') {
        postCalled = true
        return route.fulfill({
          json: { data: { ...mockCredentials[0], id: 'new-cred-id' } },
        })
      }
      return route.continue()
    })

    await page.goto('/admin/metadata/credentials/new')
    await page.getByTestId('field-code').fill('test_api')
    await page.getByTestId('field-name').fill('Test API')
    await page.getByTestId('field-base-url').fill('https://api.example.com')
    await page.getByTestId('field-header').fill('X-API-Key')
    await page.getByTestId('field-value').fill('secret123')
    await page.getByTestId('submit-btn').click()

    await page.waitForTimeout(500)
    expect(postCalled).toBe(true)
  })
})

// ─── Detail page ─────────────────────────────────────────────

test.describe('Credential detail page', () => {
  const detailUrl = `/admin/metadata/credentials/${mockCredentials[0].id}`

  test('shows credential heading', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByRole('heading', { name: 'stripe_api' })).toBeVisible()
  })

  test('shows test connection button', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('test-connection-btn')).toBeVisible()
  })

  test('shows toggle active button', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('toggle-active-btn')).toBeVisible()
  })

  test('shows delete button', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('delete-credential-btn')).toBeVisible()
  })

  test('shows settings tab with form', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('tab-settings')).toBeVisible()
    await expect(page.getByTestId('field-name')).toBeVisible()
    await expect(page.getByTestId('field-base-url')).toBeVisible()
    await expect(page.getByTestId('save-btn')).toBeVisible()
  })

  test('usage tab shows log entries', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('tab-usage').click()
    await expect(page.getByTestId('usage-row')).toHaveCount(mockCredentialUsageLog.length)
  })
})

// ─── Sidebar ─────────────────────────────────────────────────

test.describe('Sidebar navigation', () => {
  test('credentials link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/credentials')
    await expect(page.getByRole('complementary').getByRole('link', { name: 'Credentials' })).toBeVisible()
  })

  test('credentials link navigates correctly', async ({ page }) => {
    await page.goto('/admin')
    await page.locator('aside').getByText('Automation').click()
    await page.getByRole('complementary').getByRole('link', { name: 'Credentials' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/credentials/)
  })
})
