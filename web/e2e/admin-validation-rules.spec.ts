import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockValidationRules, mockObjects } from './fixtures/mock-api'

const objectId = mockObjects[0].id

test.describe('Validation rule list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows validation rules', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules`)
    await expect(page.getByText('Имя обязательно')).toBeVisible()
    await expect(page.getByText('name_required')).toBeVisible()
  })

  test('shows severity badge', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules`)
    await expect(page.getByText('error').first()).toBeVisible()
  })

  test('shows inactive badge for inactive rules', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules`)
    await expect(page.getByText('Неактивно')).toBeVisible()
  })

  test('has create rule button', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules`)
    await expect(page.locator('[data-testid="create-rule-btn"]')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules`)
    await page.locator('[data-testid="create-rule-btn"]').click()
    await expect(page).toHaveURL(new RegExp(`/admin/metadata/objects/${objectId}/rules/new`))
  })

  test('clicking rule row navigates to detail', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules`)
    await page.locator('[data-testid="rule-row"]').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/metadata/objects/${objectId}/rules/${mockValidationRules[0].id}`),
    )
  })

  test('shows empty state when no rules', async ({ page }) => {
    const emptyObjectId = mockObjects[1].id
    await page.goto(`/admin/metadata/objects/${emptyObjectId}/rules`)
    await expect(page.getByText('Нет правил валидации')).toBeVisible()
  })
})

test.describe('Validation rule create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form with all fields', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/new`)
    await expect(page.locator('[data-testid="field-api-name"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-label"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-expression"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-error-message"]')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/new`)
    await expect(page.locator('[data-testid="submit-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/new`)
    await page.locator('[data-testid="cancel-btn"]').click()
    await expect(page).toHaveURL(new RegExp(`/admin/metadata/objects/${objectId}/rules`))
  })

  test('submit calls POST', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/new`)

    await page.locator('[data-testid="field-api-name"]').fill('test_rule')
    await page.locator('[data-testid="field-label"]').fill('Test Rule')
    await page.locator('[data-testid="field-expression"]').fill('size(record.Name) > 0')
    await page.locator('[data-testid="field-error-message"]').fill('Name is required')

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/metadata/objects/${objectId}/rules`) &&
        req.method() === 'POST',
    )
    await page.locator('[data-testid="submit-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })

  test('shows breadcrumbs', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/new`)
    await expect(page.getByText('Правила').first()).toBeVisible()
    await expect(page.getByText('Создание').first()).toBeVisible()
  })
})

test.describe('Validation rule detail page', () => {
  const rule = mockValidationRules[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays rule heading', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/${rule.id}`)
    await expect(
      page.getByRole('heading', { name: rule.label }),
    ).toBeVisible()
  })

  test('shows disabled API name', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/${rule.id}`)
    await expect(page.getByText('API Name')).toBeVisible()
  })

  test('has editable form fields', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/${rule.id}`)
    await expect(page.locator('[data-testid="field-label"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-expression"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-error-message"]')).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/${rule.id}`)
    await expect(page.locator('[data-testid="save-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="delete-rule-btn"]')).toBeVisible()
  })

  test('submit calls PUT', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/${rule.id}`)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/metadata/objects/${objectId}/rules/${rule.id}`) &&
        req.method() === 'PUT',
    )
    await page.locator('[data-testid="save-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })

  test('delete button shows confirmation dialog', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${objectId}/rules/${rule.id}`)
    await page.locator('[data-testid="delete-rule-btn"]').click()
    await expect(page.getByText('Удалить правило?')).toBeVisible()
  })
})
