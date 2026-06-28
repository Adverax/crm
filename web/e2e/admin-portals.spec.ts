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

  test('tabs render (General, Args, Gates)', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(page.locator('[data-testid="view-tabs"]')).toBeVisible()
    await expect(page.getByRole('tab', { name: 'General' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Args' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Gates' })).toBeVisible()
  })

  test('no legacy tabs (Queries, Fields, Actions)', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await expect(page.getByRole('tab', { name: 'Queries' })).not.toBeVisible()
    await expect(page.getByRole('tab', { name: 'Fields', exact: true })).not.toBeVisible()
    await expect(page.getByRole('tab', { name: 'Actions' })).not.toBeVisible()
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

  test('delete button shows confirmation dialog', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.locator('[data-testid="delete-view-btn"]').click()
    await expect(page.getByText('Delete object view?')).toBeVisible()
  })
})

test.describe('Portal Gates tab', () => {
  const view = mockPortals[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('Gates tab shows master-detail layout', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await expect(page.locator('[data-testid="gates-master-detail"]')).toBeVisible()
  })

  test('gate cards show gate names from config', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    const cards = page.locator('[data-testid="gate-card"]')
    await expect(cards).toHaveCount(3)
    await expect(cards.first()).toContainText('list')
    await expect(cards.nth(1)).toContainText('detail')
    await expect(cards.nth(2)).toContainText('action')
  })

  test('entry gate is marked with badge', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    // 'list' is entry gate
    await expect(page.locator('[data-testid="gate-card"]').first()).toContainText('entry')
  })

  test('clicking gate shows detail with sub-tabs', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await expect(page.locator('[data-testid="gate-sub-tabs"]')).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Instructions' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Layout' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Outcomes' })).toBeVisible()
  })

  test('gate detail shows label input', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await expect(page.locator('[data-testid="gate-label-input"]')).toBeVisible()
  })

  test('add gate button works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    const cardsBefore = page.locator('[data-testid="gate-card"]')
    await expect(cardsBefore).toHaveCount(3)
    await page.locator('[data-testid="add-gate-btn"]').click()
    await expect(page.locator('[data-testid="gate-card"]')).toHaveCount(4)
  })

  test('delete gate button works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').nth(2).click()
    await page.locator('[data-testid="delete-gate-btn"]').click()
    await expect(page.locator('[data-testid="gate-card"]')).toHaveCount(2)
  })

  test('Instructions sub-tab shows steps', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await expect(page.locator('[data-testid="body-steps-editor"]')).toBeVisible()
    await expect(page.locator('[data-testid="step-card"]')).toHaveCount(1)
    await expect(page.locator('[data-testid="step-card"]').first()).toContainText('accounts')
    await expect(page.locator('[data-testid="step-card"]').first()).toContainText('soql')
  })

  test('Instructions — clicking step shows detail', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-name-input"]')).toBeVisible()
    await expect(page.getByText('Instruction').first()).toBeVisible()
    await expect(page.locator('[data-testid="step-instruction-editor"]')).toBeVisible()
  })

  test('Instructions — add step works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.locator('[data-testid="add-step-btn"]').click()
    await expect(page.locator('[data-testid="step-card"]')).toHaveCount(2)
  })

  test('Layout sub-tab shows field editor', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.getByRole('tab', { name: 'Layout' }).click()
    await expect(page.locator('[data-testid="layout-fields-editor"]')).toBeVisible()
    const fields = page.locator('[data-testid="field-card"]')
    await expect(fields).toHaveCount(4)
    await expect(fields.first()).toContainText('Name')
  })

  test('Layout — computed badge on fields with expr', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.getByRole('tab', { name: 'Layout' }).click()
    // display_name has expr → "computed" badge
    await expect(page.locator('[data-testid="field-card"]').nth(3)).toContainText('computed')
  })

  test('Layout — clicking field shows detail', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.getByRole('tab', { name: 'Layout' }).click()
    await page.locator('[data-testid="field-card"]').first().click()
    await expect(page.getByText('Field Details')).toBeVisible()
    await expect(page.locator('[data-testid="field-name-input"]')).toBeVisible()
  })

  test('Outcomes sub-tab shows outcomes', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.getByRole('tab', { name: 'Outcomes' }).click()
    await expect(page.locator('[data-testid="outcomes-editor"]')).toBeVisible()
    const outcomes = page.locator('[data-testid="outcome-card"]')
    await expect(outcomes).toHaveCount(2)
    await expect(outcomes.first()).toContainText('detail')
  })

  test('Outcomes — clicking outcome shows detail with target gate selector', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.getByRole('tab', { name: 'Outcomes' }).click()
    await page.locator('[data-testid="outcome-card"]').first().click()
    await expect(page.locator('[data-testid="outcome-name-input"]')).toBeVisible()
    await expect(page.getByText('Target Gate')).toBeVisible()
  })

  test('Outcomes — add outcome works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.getByRole('tab', { name: 'Outcomes' }).click()
    await page.locator('[data-testid="add-outcome-btn"]').click()
    await expect(page.locator('[data-testid="outcome-card"]')).toHaveCount(3)
  })

  test('set as entry button changes entry gate', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    // Click the second gate (detail) which is not entry
    await page.locator('[data-testid="gate-card"]').nth(1).click()
    await page.locator('[data-testid="set-entry-btn"]').click()
    // Now detail gate should have entry badge
    await expect(page.locator('[data-testid="gate-card"]').nth(1)).toContainText('entry')
  })

  test('Instructions — type badge shows soql for SELECT step', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-type-badge"]')).toHaveText('soql')
  })

  test('Instructions — type badge shows dml for DML step', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    // action gate has a DML step
    await page.locator('[data-testid="gate-card"]').nth(2).click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-type-badge"]')).toHaveText('dml')
  })

  test('Instructions — new step shows auto badge when empty', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.locator('[data-testid="add-step-btn"]').click()
    // New step is auto-selected; click it to be sure
    await page.locator('[data-testid="step-card"]').last().click()
    await expect(page.locator('[data-testid="step-type-badge"]')).toHaveText('auto')
  })

  test('Instructions — page size visible only for soql', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    // soql step
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-page-size-input"]')).toBeVisible()
    // dml step
    await page.locator('[data-testid="gate-card"]').nth(2).click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-page-size-input"]')).not.toBeVisible()
  })

  test('Instructions — restrict filters visible only for soql', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    // soql step
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-restrict-filters-input"]')).toBeVisible()
    // dml step
    await page.locator('[data-testid="gate-card"]').nth(2).click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-restrict-filters-input"]')).not.toBeVisible()
  })

  test('Instructions — no type dropdown present', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Gates' }).click()
    await page.locator('[data-testid="gate-card"]').first().click()
    await page.locator('[data-testid="step-card"]').first().click()
    await expect(page.locator('[data-testid="step-type-select"]')).not.toBeVisible()
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

  test('arg rules section is visible under args', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await expect(page.getByText('Arg Validation Rules')).toBeVisible()
    await expect(page.locator('[data-testid="arg-rules-editor"]')).toBeVisible()
  })

  test('arg rule cards show rule names', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    const ruleCards = page.locator('[data-testid="arg-rule-card"]')
    await expect(ruleCards).toHaveCount(1)
    await expect(ruleCards.first()).toContainText('limit_range')
  })

  test('add arg rule button works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await page.locator('[data-testid="add-arg-rule-btn"]').click()
    await expect(page.locator('[data-testid="arg-rule-card"]')).toHaveCount(2)
  })

  test('delete arg rule button works', async ({ page }) => {
    await page.goto(`/admin/metadata/portals/${view.id}`)
    await page.getByRole('tab', { name: 'Args' }).click()
    await page.locator('[data-testid="arg-rule-card"]').first().click()
    await page.locator('[data-testid="delete-arg-rule-btn"]').click()
    await expect(page.locator('[data-testid="arg-rule-card"]')).toHaveCount(0)
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
