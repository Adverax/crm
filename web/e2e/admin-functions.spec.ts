import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockFunctions } from './fixtures/mock-api'

test.describe('Function list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows functions', async ({ page }) => {
    await page.goto('/admin/metadata/functions')
    await expect(page.getByText('fn.discount')).toBeVisible()
    await expect(page.getByText('fn.is_premium')).toBeVisible()
  })

  test('shows description', async ({ page }) => {
    await page.goto('/admin/metadata/functions')
    await expect(page.getByText('Рассчитывает скидку по сумме')).toBeVisible()
  })

  test('shows return type badge', async ({ page }) => {
    await page.goto('/admin/metadata/functions')
    await expect(page.getByText('number').first()).toBeVisible()
  })

  test('shows params count badge', async ({ page }) => {
    await page.goto('/admin/metadata/functions')
    await expect(page.getByText('2 пар.')).toBeVisible()
  })

  test('has create function button', async ({ page }) => {
    await page.goto('/admin/metadata/functions')
    await expect(page.locator('[data-testid="create-function-btn"]')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/functions')
    await page.locator('[data-testid="create-function-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/functions\/new/)
  })

  test('clicking function row navigates to detail', async ({ page }) => {
    await page.goto('/admin/metadata/functions')
    await page.locator('[data-testid="function-row"]').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/metadata/functions/${mockFunctions[0].id}`),
    )
  })

  test('shows empty state when no functions', async ({ page }) => {
    // Override function route with empty list
    await page.route('**/api/v1/admin/functions', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.goto('/admin/metadata/functions')
    await expect(page.getByText('Нет функций')).toBeVisible()
  })
})

test.describe('Function create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form with all fields', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await expect(page.locator('[data-testid="field-name"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-return-type"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-description"]')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await expect(page.locator('[data-testid="submit-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await page.locator('[data-testid="cancel-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/functions$/)
  })

  test('has add parameter button', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await expect(page.locator('[data-testid="add-param-btn"]')).toBeVisible()
  })

  test('adding parameter shows param row', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await page.locator('[data-testid="add-param-btn"]').click()
    await expect(page.locator('[data-testid="param-row"]')).toBeVisible()
    await expect(page.locator('[data-testid="param-name-0"]')).toBeVisible()
  })

  test('removing parameter removes param row', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await page.locator('[data-testid="add-param-btn"]').click()
    await expect(page.locator('[data-testid="param-row"]')).toBeVisible()
    await page.locator('[data-testid="remove-param-0"]').click()
    await expect(page.locator('[data-testid="param-row"]')).not.toBeVisible()
  })

  test('submit calls POST', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')

    await page.locator('[data-testid="field-name"]').fill('test_fn')

    // Fill body in CodeMirror editor
    const editor = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await editor.click()
    await page.keyboard.type('42')

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/admin/functions') &&
        req.method() === 'POST',
    )
    await page.locator('[data-testid="submit-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })

  test('shows fn.name preview', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await page.locator('[data-testid="field-name"]').fill('my_func')
    await expect(page.getByText('Вызывается как fn.my_func()')).toBeVisible()
  })

  test('shows breadcrumbs', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await expect(page.getByText('Функции').first()).toBeVisible()
    await expect(page.getByText('Создание').first()).toBeVisible()
  })
})

test.describe('Function detail page', () => {
  const fn = mockFunctions[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays function heading', async ({ page }) => {
    await page.goto(`/admin/metadata/functions/${fn.id}`)
    await expect(
      page.getByRole('heading', { name: `fn.${fn.name}` }),
    ).toBeVisible()
  })

  test('shows disabled function name', async ({ page }) => {
    await page.goto(`/admin/metadata/functions/${fn.id}`)
    const nameInput = page.locator('input[disabled]').first()
    await expect(nameInput).toHaveValue(fn.name)
  })

  test('has editable form fields', async ({ page }) => {
    await page.goto(`/admin/metadata/functions/${fn.id}`)
    await expect(page.locator('[data-testid="field-description"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-return-type"]')).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/metadata/functions/${fn.id}`)
    await expect(page.locator('[data-testid="save-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="delete-function-btn"]')).toBeVisible()
  })

  test('submit calls PUT', async ({ page }) => {
    await page.goto(`/admin/metadata/functions/${fn.id}`)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/functions/${fn.id}`) &&
        req.method() === 'PUT',
    )
    await page.locator('[data-testid="save-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })

  test('delete button shows confirmation dialog', async ({ page }) => {
    await page.goto(`/admin/metadata/functions/${fn.id}`)
    await page.locator('[data-testid="delete-function-btn"]').click()
    await expect(page.getByText('Удалить функцию?')).toBeVisible()
  })

  test('shows parameter rows', async ({ page }) => {
    await page.goto(`/admin/metadata/functions/${fn.id}`)
    await expect(page.locator('[data-testid="param-row"]')).toBeVisible()
    await expect(page.locator('[data-testid="param-name-0"]')).toHaveValue('amount')
  })
})

test.describe('Sidebar navigation', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('functions link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await expect(page.getByRole('link', { name: 'Функции' })).toBeVisible()
  })

  test('functions link navigates to functions list', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.getByRole('link', { name: 'Функции' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/functions/)
  })
})
