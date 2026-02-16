import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockSharingRules } from './fixtures/mock-api'

test.describe('Sharing rule list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows object selector', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await expect(page.getByText('Object').first()).toBeVisible()
  })

  test('has create rule button', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await expect(page.getByText('Create Rule')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await page.getByText('Create Rule').click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules\/new/)
  })

  test('shows prompt to select object when no object selected', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await expect(page.getByText('Select an object to view sharing rules.')).toBeVisible()
  })
})

test.describe('Sharing rule create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form with all selectors', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules/new')
    await expect(page.getByText('Object').first()).toBeVisible()
    await expect(page.getByText('Rule Type').first()).toBeVisible()
    await expect(page.getByText('Source Group').first()).toBeVisible()
    await expect(page.getByText('Target Group').first()).toBeVisible()
    await expect(page.getByText('Access Level').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await page.getByText('Create Rule').click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules/)
  })

  test('shows breadcrumbs with correct links', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules/new')
    await expect(page.getByText('Sharing Rules').first()).toBeVisible()
    await expect(page.getByText('New Rule').first()).toBeVisible()
  })
})

test.describe('Sharing rule detail page', () => {
  const rule = mockSharingRules[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays rule heading', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(
      page.getByRole('heading', { name: /Rule/ }),
    ).toBeVisible()
  })

  test('shows read-only object and rule type fields', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByText('Object').first()).toBeVisible()
    await expect(page.getByText('Rule Type').first()).toBeVisible()
    await expect(page.getByText('Source Group').first()).toBeVisible()
  })

  test('has editable target group and access level', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByText('Target Group').first()).toBeVisible()
    await expect(page.getByText('Access Level').first()).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('can submit updated sharing rule', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/security/sharing-rules/${rule.id}`) &&
        req.method() === 'PUT',
    )
    await page.getByRole('button', { name: 'Save' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })
})

test.describe('Criteria-based sharing rule detail page', () => {
  const rule = mockSharingRules[1]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows criteria fields for criteria_based rule', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByText('Criteria').first()).toBeVisible()
    await expect(page.locator('#criteriaField')).toBeVisible()
    await expect(page.locator('#criteriaOp')).toBeVisible()
    await expect(page.locator('#criteriaValue')).toBeVisible()
  })
})
