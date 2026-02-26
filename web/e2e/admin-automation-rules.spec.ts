import { test, expect } from '@playwright/test'
import {
  setupAllRoutes,
  mockAutomationRules,
  mockObjects,
} from './fixtures/mock-api'

test.beforeEach(async ({ page }) => {
  await setupAllRoutes(page)
})

// ─── List page ───────────────────────────────────────────────

test.describe('Automation rule list page', () => {
  test('shows rules from API', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await expect(page.getByTestId('rule-row')).toHaveCount(mockAutomationRules.length)
    await expect(page.getByText('Notify on insert')).toBeVisible()
    await expect(page.getByText('Auto-approve update')).toBeVisible()
  })

  test('shows object selector', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await expect(page.getByTestId('object-select')).toBeVisible()
  })

  test('shows create button', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await expect(page.getByTestId('create-rule-btn')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await page.getByTestId('create-rule-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/automation-rules\/new\//)
  })

  test('clicking row navigates to detail page', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await page.getByTestId('rule-row').first().click()
    await expect(page).toHaveURL(/\/admin\/metadata\/automation-rules\/ar111111/)
  })

  test('shows empty state when no rules', async ({ page }) => {
    await page.route(`**/api/v1/admin/metadata/objects/${mockObjects[0]!.id}/automation-rules`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.fallback()
    })
    await page.goto('/admin/metadata/automation-rules')
    await expect(page.getByText('No automation rules')).toBeVisible()
  })

  test('shows event type and procedure code for each rule', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await expect(page.getByText('after insert')).toBeVisible()
    await expect(page.getByText('notify_manager')).toBeVisible()
  })

  test('shows execution mode badge', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await expect(page.getByText('per_record').first()).toBeVisible()
  })
})

// ─── Create page ─────────────────────────────────────────────

test.describe('Automation rule create page', () => {
  const createUrl = `/admin/metadata/automation-rules/new/${mockObjects[0]!.id}`

  test('renders form with fields', async ({ page }) => {
    await page.goto(createUrl)
    await expect(page.getByTestId('field-name')).toBeVisible()
    await expect(page.getByTestId('field-event-type')).toBeVisible()
    await expect(page.getByTestId('field-procedure-code')).toBeVisible()
  })

  test('has create and cancel buttons', async ({ page }) => {
    await page.goto(createUrl)
    await expect(page.getByTestId('submit-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto(createUrl)
    await page.getByTestId('cancel-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/automation-rules$/)
  })

  test('shows condition and execution mode fields', async ({ page }) => {
    await page.goto(createUrl)
    await expect(page.getByTestId('field-condition')).toBeVisible()
    await expect(page.getByTestId('field-execution-mode')).toBeVisible()
    await expect(page.getByTestId('field-sort-order')).toBeVisible()
    await expect(page.getByTestId('field-is-active')).toBeVisible()
  })

  test('submitting form sends POST request', async ({ page }) => {
    let postCalled = false
    await page.route(`**/api/v1/admin/metadata/objects/${mockObjects[0]!.id}/automation-rules`, (route) => {
      if (route.request().method() === 'POST') {
        postCalled = true
        return route.fulfill({
          json: { data: { ...mockAutomationRules[0], id: 'new-rule-id' } },
        })
      }
      return route.fallback()
    })

    await page.goto(createUrl)
    await page.getByTestId('field-name').fill('Test rule')
    await page.getByTestId('field-procedure-code').fill('test_proc')
    await page.getByTestId('submit-btn').click()

    await page.waitForTimeout(500)
    expect(postCalled).toBe(true)
  })
})

// ─── Detail page ─────────────────────────────────────────────

test.describe('Automation rule detail page', () => {
  const detailUrl = `/admin/metadata/automation-rules/${mockAutomationRules[0]!.id}`

  test('shows rule heading', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByRole('heading', { name: 'Notify on insert' })).toBeVisible()
  })

  test('shows form with editable fields', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('field-name')).toBeVisible()
    await expect(page.getByTestId('field-description')).toBeVisible()
    await expect(page.getByTestId('field-event-type')).toBeVisible()
    await expect(page.getByTestId('field-procedure-code')).toBeVisible()
    await expect(page.getByTestId('field-condition')).toBeVisible()
    await expect(page.getByTestId('field-execution-mode')).toBeVisible()
  })

  test('shows save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('save-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
    await expect(page.getByTestId('delete-rule-btn')).toBeVisible()
  })

  test('shows event type and procedure code badges', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByText('after insert', { exact: true })).toBeVisible()
    await expect(page.getByText('notify_manager', { exact: true })).toBeVisible()
  })

  test('submitting form sends PUT request', async ({ page }) => {
    let putCalled = false
    await page.route(`**/api/v1/admin/metadata/automation-rules/${mockAutomationRules[0]!.id}`, (route) => {
      if (route.request().method() === 'PUT') {
        putCalled = true
        return route.fulfill({
          json: { data: mockAutomationRules[0] },
        })
      }
      return route.fallback()
    })

    await page.goto(detailUrl)
    await page.getByTestId('field-name').fill('Updated rule name')
    await page.getByTestId('save-btn').click()

    await page.waitForTimeout(500)
    expect(putCalled).toBe(true)
  })
})

// ─── Sidebar ─────────────────────────────────────────────────

test.describe('Sidebar navigation', () => {
  test('automation rules link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/automation-rules')
    await expect(
      page.getByRole('complementary').getByRole('link', { name: 'Automation Rules' }),
    ).toBeVisible()
  })

  test('automation rules link navigates correctly', async ({ page }) => {
    await page.goto('/admin')
    await page.locator('aside').getByText('Automation').click()
    await page.getByRole('complementary').getByRole('link', { name: 'Automation Rules' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/automation-rules/)
  })
})
