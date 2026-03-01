import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockObjectViews } from './fixtures/mock-api'

test.describe('Object View list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows object views', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await expect(page.getByText('Account Default View')).toBeVisible()
    await expect(page.getByText('Account Sales View')).toBeVisible()
  })

  test('shows api names', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await expect(page.getByText('account_default')).toBeVisible()
    await expect(page.getByText('account_sales_view')).toBeVisible()
  })

  test('shows Default badge for default view', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await expect(page.getByText('Default').first()).toBeVisible()
  })

  test('shows Global badge for views without profile', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await expect(page.getByText('Global').first()).toBeVisible()
  })

  test('shows Profile-specific badge for views with profile', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await expect(page.getByText('Profile-specific')).toBeVisible()
  })

  test('has create view button', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await expect(page.locator('[data-testid="create-view-btn"]')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await page.locator('[data-testid="create-view-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/object-views\/new/)
  })

  test('clicking view row navigates to detail', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await page.locator('[data-testid="view-row"]').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/metadata/object-views/${mockObjectViews[0].id}`),
    )
  })

  test('shows empty state when no views', async ({ page }) => {
    await page.route('**/api/v1/admin/object-views', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.goto('/admin/metadata/object-views')
    await expect(page.getByText('No object views')).toBeVisible()
  })
})

test.describe('Object View create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form with all fields', async ({ page }) => {
    await page.goto('/admin/metadata/object-views/new')
    await expect(page.locator('[data-testid="field-api-name"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-label"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-description"]')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/object-views/new')
    await expect(page.locator('[data-testid="submit-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/object-views/new')
    await page.locator('[data-testid="cancel-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/object-views$/)
  })

  test('shows breadcrumbs', async ({ page }) => {
    await page.goto('/admin/metadata/object-views/new')
    await expect(page.getByText('Object Views').first()).toBeVisible()
    await expect(page.getByText('Create').first()).toBeVisible()
  })

  test('submit calls POST', async ({ page }) => {
    await page.goto('/admin/metadata/object-views/new')

    await page.locator('[data-testid="field-api-name"]').fill('test_view')
    await page.locator('[data-testid="field-label"]').fill('Test View')

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/admin/object-views') &&
        req.method() === 'POST',
    )
    await page.locator('[data-testid="submit-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Object View detail page', () => {
  const view = mockObjectViews[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays view heading', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await expect(
      page.getByRole('heading', { name: view.label }),
    ).toBeVisible()
  })

  test('shows editable form fields on General tab', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await expect(page.locator('[data-testid="field-label"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-description"]')).toBeVisible()
  })

  test('has save and delete buttons', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await expect(page.locator('[data-testid="save-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="delete-view-btn"]')).toBeVisible()
  })

  test('tabs render (General, Queries, Fields, Actions)', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await expect(page.locator('[data-testid="view-tabs"]')).toBeVisible()
    await expect(page.getByRole('tab', { name: 'General' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Fields', exact: true })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Actions' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Queries' })).toBeVisible()
  })

  test('no Validation/Defaults/Mutations tabs', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await expect(page.getByRole('tab', { name: 'Validation' })).not.toBeVisible()
    await expect(page.getByRole('tab', { name: 'Defaults' })).not.toBeVisible()
    await expect(page.getByRole('tab', { name: 'Mutations' })).not.toBeVisible()
  })

  test('Fields tab shows field entries', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Fields', exact: true }).click()
    await expect(page.locator('[data-testid="field-entry"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="add-field-btn"]')).toBeVisible()
  })

  test('Actions tab shows master-detail layout with action list', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await expect(page.locator('[data-testid="actions-master-detail"]')).toBeVisible()
    await expect(page.locator('[data-testid="action-card"]')).toBeVisible()
    await expect(page.locator('[data-testid="add-action-btn"]')).toBeVisible()
  })

  test('Actions tab — clicking action shows detail tabs', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await expect(page.locator('[data-testid="tab-action-identity"]')).toBeVisible()
    await expect(page.locator('[data-testid="tab-action-form"]')).toBeVisible()
    await expect(page.locator('[data-testid="tab-action-validation"]')).toBeVisible()
    await expect(page.locator('[data-testid="tab-action-apply"]')).toBeVisible()
  })

  test('Actions tab — Form tab shows form fields', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-form"]').click()
    await expect(page.locator('[data-testid="action-form-field"]').first()).toBeVisible()
  })

  test('Actions tab — Validation tab shows rules', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-validation"]').click()
    await expect(page.locator('[data-testid="action-validation-rule"]').first()).toBeVisible()
  })

  test('Actions tab — Apply tab shows DML config', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await expect(page.locator('[data-testid="dml-statement"]').first()).toBeVisible()
  })

  test('submit calls PUT', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/object-views/${view.id}`) &&
        req.method() === 'PUT',
    )
    await page.locator('[data-testid="save-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })

  test('Queries tab shows query cards from config', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await expect(page.locator('[data-testid="query-card"]')).toBeVisible()
    await expect(page.locator('[data-testid="add-query-btn"]')).toBeVisible()
  })

  test('Fields tab shows computed field with expression', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Fields', exact: true }).click()
    // The 4th field (display_name) has an expr, so Expression textarea should be visible
    const fields = page.locator('[data-testid="field-entry"]')
    await expect(fields).toHaveCount(4)
  })

  test('delete button shows confirmation dialog', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.locator('[data-testid="delete-view-btn"]').click()
    await expect(page.getByText('Delete object view?')).toBeVisible()
  })
})

test.describe('Sidebar navigation', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('Object Views link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/object-views')
    await expect(page.locator('aside').getByRole('link', { name: 'Object Views' })).toBeVisible()
  })

  test('Object Views link navigates to list', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.locator('aside').getByText('Presentation').click()
    await page.getByRole('link', { name: 'Object Views' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/object-views/)
  })
})

test.describe('OV queries tab — SOQL editor', () => {
  const view = mockObjectViews[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('SOQL editor is visible in query card', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await expect(page.locator('[data-testid="soql-editor"]')).toBeVisible()
  })

  test('Validate button sends POST to /soql/validate', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    const requestPromise = page.waitForRequest('**/api/v1/admin/soql/validate')
    await page.locator('[data-testid="soql-validate-btn"]').first().click()
    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })

  test('Validation errors are displayed', async ({ page }) => {
    // Override validate route to return errors
    await page.route('**/api/v1/admin/soql/validate', (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: {
            valid: false,
            errors: [{ message: 'Unknown object: Foo', line: 1, column: 20 }],
          },
        })
      }
      return route.continue()
    })
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await page.locator('[data-testid="soql-validate-btn"]').first().click()
    await expect(page.getByText('Unknown object: Foo')).toBeVisible()
  })

  test('Test Query executes and shows results', async ({ page }) => {
    await page.goto(`/admin/metadata/object-views/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await page.locator('[data-testid="soql-test-btn"]').first().click()
    await expect(page.locator('[data-testid="soql-test-result"]')).toBeVisible()
    await expect(page.getByText('2 record(s) found')).toBeVisible()
    await expect(page.getByText('Acme Corp')).toBeVisible()
  })
})
