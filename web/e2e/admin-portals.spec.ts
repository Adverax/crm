import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockPortals } from './fixtures/mock-api'

test.describe('Portal list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows object views', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await expect(page.getByText('Account Default View')).toBeVisible()
    await expect(page.getByText('Account Sales View')).toBeVisible()
  })

  test('shows api names', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await expect(page.getByText('account_default')).toBeVisible()
    await expect(page.getByText('account_sales_view')).toBeVisible()
  })

  test('shows Default badge for default view', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await expect(page.getByText('Default').first()).toBeVisible()
  })

  test('shows Global badge for views without profile', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await expect(page.getByText('Global').first()).toBeVisible()
  })

  test('shows Profile-specific badge for views with profile', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await expect(page.getByText('Profile-specific')).toBeVisible()
  })

  test('has create view button', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await expect(page.locator('[data-testid="create-view-btn"]')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await page.locator('[data-testid="create-view-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/portals\/new/)
  })

  test('clicking view row navigates to detail', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await page.locator('[data-testid="view-row"]').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/metadata/portals/${mockPortals[0].id}`),
    )
  })

  test('shows empty state when no views', async ({ page }) => {
    await page.route('**/api/v1/admin/portals', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.goto('/admin/metadata/portals')
    await expect(page.getByText('No object views')).toBeVisible()
  })
})

test.describe('Portal create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form with all fields', async ({ page }) => {
    await page.goto('/admin/metadata/portals/new')
    await expect(page.locator('[data-testid="field-api-name"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-label"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-description"]')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/portals/new')
    await expect(page.locator('[data-testid="submit-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="cancel-btn"]')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/portals/new')
    await page.locator('[data-testid="cancel-btn"]').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/portals$/)
  })

  test('shows breadcrumbs', async ({ page }) => {
    await page.goto('/admin/metadata/portals/new')
    await expect(page.getByText('Portals').first()).toBeVisible()
    await expect(page.getByText('Create').first()).toBeVisible()
  })

  test('submit calls POST', async ({ page }) => {
    await page.goto('/admin/metadata/portals/new')

    await page.locator('[data-testid="field-api-name"]').fill('test_view')
    await page.locator('[data-testid="field-label"]').fill('Test View')

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/admin/portals') &&
        req.method() === 'POST',
    )
    await page.locator('[data-testid="submit-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Portal detail page', () => {
  const view = mockPortals[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays view heading', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(
      page.getByRole('heading', { name: view.label }),
    ).toBeVisible()
  })

  test('shows editable form fields on General tab', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(page.locator('[data-testid="field-label"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-description"]')).toBeVisible()
  })

  test('has save and delete buttons', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(page.locator('[data-testid="save-btn"]')).toBeVisible()
    await expect(page.locator('[data-testid="delete-view-btn"]')).toBeVisible()
  })

  test('tabs render (General, Args, Queries, Fields, Actions)', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(page.locator('[data-testid="view-tabs"]')).toBeVisible()
    await expect(page.getByRole('tab', { name: 'General' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Args' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Fields', exact: true })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Actions' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Queries' })).toBeVisible()
  })

  test('no Validation/Defaults/Mutations tabs', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(page.getByRole('tab', { name: 'Validation' })).not.toBeVisible()
    await expect(page.getByRole('tab', { name: 'Defaults' })).not.toBeVisible()
    await expect(page.getByRole('tab', { name: 'Mutations' })).not.toBeVisible()
  })

  test('Fields tab shows master-detail layout with field list', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Fields', exact: true }).click()
    await expect(page.locator('[data-testid="fields-master-detail"]')).toBeVisible()
    await expect(page.locator('[data-testid="field-card"]').first()).toBeVisible()
    await expect(page.locator('[data-testid="add-field-btn"]')).toBeVisible()
  })

  test('Actions tab shows master-detail layout with action list', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await expect(page.locator('[data-testid="actions-master-detail"]')).toBeVisible()
    await expect(page.locator('[data-testid="action-card"]')).toBeVisible()
    await expect(page.locator('[data-testid="add-action-btn"]')).toBeVisible()
  })

  test('Actions tab — clicking action shows detail tabs', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await expect(page.locator('[data-testid="tab-action-identity"]')).toBeVisible()
    await expect(page.locator('[data-testid="tab-action-apply"]')).toBeVisible()
  })

  test('Actions tab — Apply tab shows DML config', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await expect(page.locator('[data-testid="dml-statement"]').first()).toBeVisible()
  })

  test('submit calls PUT', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/portals/${view.id}`) &&
        req.method() === 'PUT',
    )
    await page.locator('[data-testid="save-btn"]').click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })

  test('Queries tab shows query cards from config', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await expect(page.locator('[data-testid="query-card"]')).toBeVisible()
    await expect(page.locator('[data-testid="add-query-btn"]')).toBeVisible()
  })

  test('Fields tab shows computed badge for fields with expr', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Fields', exact: true }).click()
    const fields = page.locator('[data-testid="field-card"]')
    await expect(fields).toHaveCount(4)
    // display_name has expr → "computed" badge
    await expect(fields.nth(3)).toContainText('computed')
  })

  test('Fields tab — clicking field shows detail with ExpressionBuilder', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Fields', exact: true }).click()
    await page.locator('[data-testid="field-card"]').nth(3).click()
    await expect(page.getByText('Field Details')).toBeVisible()
    await expect(page.getByText('Expression (CEL)')).toBeVisible()
    await expect(page.getByText('When (CEL)')).toBeVisible()
  })

  test('Fields tab — shows available query names in description', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Fields', exact: true }).click()
    await page.locator('[data-testid="field-card"]').first().click()
    await expect(page.getByText('recent_activities')).toBeVisible()
  })

  test('delete button shows confirmation dialog', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.locator('[data-testid="delete-view-btn"]').click()
    await expect(page.getByText('Delete object view?')).toBeVisible()
  })
})

test.describe('Portal Args tab', () => {
  const view = mockPortals[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('Args tab is visible in tab list', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(page.getByRole('tab', { name: 'Args' })).toBeVisible()
  })

  test('master-detail layout renders', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await expect(page.locator('[data-testid="args-master-detail"]')).toBeVisible()
  })

  test('arg cards show name and type badge', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    const cards = page.locator('[data-testid="arg-card"]')
    await expect(cards).toHaveCount(2)
    await expect(cards.first()).toContainText('account_id')
    await expect(cards.first()).toContainText('string')
  })

  test('required/optional badges display correctly', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    const cards = page.locator('[data-testid="arg-card"]')
    // account_id has no default → required
    await expect(cards.first()).toContainText('required')
    // limit has default → optional
    await expect(cards.nth(1)).toContainText('optional')
  })

  test('clicking arg shows detail form', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await page.locator('[data-testid="arg-card"]').first().click()
    await expect(page.locator('[data-testid="arg-name-input"]')).toBeVisible()
    await expect(page.locator('[data-testid="arg-default-input"]')).toBeVisible()
  })

  test('add arg button works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    const cardsBefore = page.locator('[data-testid="arg-card"]')
    await expect(cardsBefore).toHaveCount(2)
    await page.locator('[data-testid="add-arg-btn"]').click()
    await expect(page.locator('[data-testid="arg-card"]')).toHaveCount(3)
  })

  test('delete arg button works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await page.locator('[data-testid="arg-card"]').first().click()
    await page.locator('[data-testid="delete-arg-btn"]').click()
    await expect(page.locator('[data-testid="arg-card"]')).toHaveCount(1)
  })

  test('shows name validation error for empty name', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await page.locator('[data-testid="add-arg-btn"]').click()
    // New arg has empty name — click it to see detail
    await page.locator('[data-testid="arg-card"]').last().click()
    await expect(page.locator('[data-testid="arg-name-error"]')).toContainText('Name is required')
  })

  test('shows name validation error for invalid format', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await page.locator('[data-testid="arg-card"]').first().click()
    await page.locator('[data-testid="arg-name-input"]').fill('BadName')
    await expect(page.locator('[data-testid="arg-name-error"]')).toContainText('Must match')
  })

  test('shows default validation error for int type with non-numeric value', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    // Click the 'limit' arg (int type with default)
    await page.locator('[data-testid="arg-card"]').nth(1).click()
    await page.locator('[data-testid="arg-default-input"]').fill('abc')
    await expect(page.locator('[data-testid="arg-default-error"]')).toContainText('not a valid int')
  })

  test('error indicator on arg card with invalid name', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await page.locator('[data-testid="add-arg-btn"]').click()
    // New arg with empty name should have error border
    const lastCard = page.locator('[data-testid="arg-card"]').last()
    await expect(lastCard).toHaveClass(/border-l-destructive/)
  })

  test('validation and error_message fields are visible in arg detail', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    // Click the 'limit' arg which has validation set
    await page.locator('[data-testid="arg-card"]').nth(1).click()
    await expect(page.locator('[data-testid="arg-validation-input"]')).toBeVisible()
    await expect(page.locator('[data-testid="arg-error-message-input"]')).toBeVisible()
  })

  test('error_message shows error when validation is set but error_message is empty', async ({
    page,
  }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    // Click account_id arg (no validation)
    await page.locator('[data-testid="arg-card"]').first().click()
    // Type a validation expression
    const editor = page.locator('[data-testid="arg-validation-input"]')
    await editor.locator('[data-testid="codemirror-editor"]').click()
    await page.keyboard.type('args.account_id != ""')
    // Error message should show validation error
    await expect(page.locator('[data-testid="arg-validation-error"]')).toContainText(
      'Error message is required',
    )
  })
})

test.describe('Sidebar navigation', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('Portals link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/portals')
    await expect(page.locator('aside').getByRole('link', { name: 'Portals' })).toBeVisible()
  })

  test('Portals link navigates to list', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.locator('aside').getByText('Presentation').click()
    await page.getByRole('link', { name: 'Portals' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/portals/)
  })
})

test.describe('Portal queries tab — SOQL editor', () => {
  const view = mockPortals[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('query list shows items with name and type badge', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await expect(page.locator('[data-testid="query-card"]')).toHaveCount(1)
    await expect(page.locator('[data-testid="query-card"]').first()).toContainText('recent_activities')
    await expect(page.locator('[data-testid="query-card"]').first()).toContainText('scalar')
  })

  test('selecting query shows detail panel with SOQL editor', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await page.locator('[data-testid="query-card"]').first().click()
    await expect(page.locator('[data-testid="soql-editor"]')).toBeVisible()
  })

  test('Validate button sends POST to /soql/validate', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await page.locator('[data-testid="query-card"]').first().click()
    await page.locator('[data-testid="soql-editor"] [data-testid="codemirror-editor"]').first().click()
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
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await page.locator('[data-testid="query-card"]').first().click()
    await page.locator('[data-testid="soql-editor"] [data-testid="codemirror-editor"]').first().click()
    await page.locator('[data-testid="soql-validate-btn"]').first().click()
    await expect(page.getByText('Unknown object: Foo')).toBeVisible()
  })

  test('Test Query executes and shows results', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Queries' }).click()
    await page.locator('[data-testid="query-card"]').first().click()
    await page.locator('[data-testid="soql-editor"] [data-testid="codemirror-editor"]').first().click()
    await page.locator('[data-testid="soql-test-btn"]').first().click()
    await expect(page.locator('[data-testid="soql-test-result"]')).toBeVisible()
    await expect(page.getByText('2 record(s) found')).toBeVisible()
    await expect(page.getByText('Acme Corp')).toBeVisible()
  })
})

test.describe('Portal actions tab — DML editor', () => {
  const view = mockPortals[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('Apply tab shows DML editor instead of plain textarea', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await expect(page.locator('[data-testid="dml-editor"]').first()).toBeVisible()
  })

  test('DML editor shows validate button on focus', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await page.locator('[data-testid="dml-editor"] [data-testid="codemirror-editor"]').first().click()
    await expect(page.locator('[data-testid="dml-validate-btn"]').first()).toBeVisible()
  })

  test('DML validate sends POST to /dml/validate', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await page.locator('[data-testid="dml-editor"] [data-testid="codemirror-editor"]').first().click()
    const requestPromise = page.waitForRequest('**/api/v1/admin/dml/validate')
    await page.locator('[data-testid="dml-validate-btn"]').first().click()
    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })

  test('DML test execute sends POST to /dml/test', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await page.locator('[data-testid="dml-editor"] [data-testid="codemirror-editor"]').first().click()
    const requestPromise = page.waitForRequest('**/api/v1/admin/dml/test')
    await page.locator('[data-testid="dml-test-btn"]').first().click()
    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })

  test('DML test result shows rolled back badge', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await page.locator('[data-testid="dml-editor"] [data-testid="codemirror-editor"]').first().click()
    await page.locator('[data-testid="dml-test-btn"]').first().click()
    await expect(page.locator('[data-testid="dml-test-result"]')).toBeVisible()
    await expect(page.getByText('Rolled back')).toBeVisible()
    await expect(page.getByText('1 row(s) affected')).toBeVisible()
  })

  test('DML validation errors are displayed', async ({ page }) => {
    await page.route('**/api/v1/admin/dml/validate', (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({
          json: {
            valid: false,
            errors: [{ message: 'Unknown object: Foo', line: 1, column: 13, code: 'UnknownObject' }],
          },
        })
      }
      return route.continue()
    })
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Actions' }).click()
    await page.locator('[data-testid="action-card"]').first().click()
    await page.locator('[data-testid="tab-action-apply"]').click()
    await page.locator('[data-testid="dml-editor"] [data-testid="codemirror-editor"]').first().click()
    await page.locator('[data-testid="dml-validate-btn"]').first().click()
    await expect(page.getByText('Unknown object: Foo')).toBeVisible()
  })
})
