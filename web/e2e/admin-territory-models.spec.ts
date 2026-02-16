import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockTerritoryModels, mockTerritories } from './fixtures/mock-api'

test.describe('Territory model list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays model api names', async ({ page }) => {
    await page.goto('/admin/territory/models')
    const main = page.locator('main')
    await expect(main.getByText('q1_2026')).toBeVisible()
    await expect(main.getByText('q4_2025')).toBeVisible()
  })

  test('shows model labels', async ({ page }) => {
    await page.goto('/admin/territory/models')
    const main = page.locator('main')
    await expect(main.getByText('Q1 2026')).toBeVisible()
    await expect(main.getByText('Q4 2025')).toBeVisible()
  })

  test('shows status badges', async ({ page }) => {
    await page.goto('/admin/territory/models')
    const main = page.locator('main')
    await expect(main.getByText('Planning').first()).toBeVisible()
    await expect(main.getByText('Active').first()).toBeVisible()
  })

  test('has create model button', async ({ page }) => {
    await page.goto('/admin/territory/models')
    await expect(page.getByText('Create model')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/territory/models')
    await page.getByText('Create model').click()
    await expect(page).toHaveURL(/\/admin\/territory\/models\/new/)
  })

  test('clicking model navigates to detail', async ({ page }) => {
    await page.goto('/admin/territory/models')
    await page.locator('main').getByText('q1_2026').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/territory/models/${mockTerritoryModels[0].id}`),
    )
  })

  test('shows empty state when no models', async ({ page }) => {
    await page.route('**/api/v1/admin/territory/models?*', (route) => {
      route.fulfill({ json: { data: [], meta: { page: 1, per_page: 20, total: 0, total_pages: 0 } } })
    })
    await page.route('**/api/v1/admin/territory/models', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [], meta: { page: 1, per_page: 20, total: 0, total_pages: 0 } } })
      }
      return route.continue()
    })
    await page.goto('/admin/territory/models')
    await expect(page.getByText('No territory models')).toBeVisible()
  })
})

test.describe('Territory model create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form', async ({ page }) => {
    await page.goto('/admin/territory/models/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/territory/models/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/territory/models')
    await page.getByText('Create model').click()
    await expect(page).toHaveURL(/\/admin\/territory\/models\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/territory\/models/)
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/territory/models/new')

    await page.locator('#apiName').fill('test_model')
    await page.locator('#label').fill('Test Model')

    const requestPromise = page.waitForRequest('**/api/v1/admin/territory/models')
    await page.getByRole('button', { name: 'Create' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Territory model detail page', () => {
  const model = mockTerritoryModels[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays model heading', async ({ page }) => {
    await page.goto(`/admin/territory/models/${model.id}`)
    await expect(
      page.getByRole('heading', { name: model.label }),
    ).toBeVisible()
  })

  test('shows status badge', async ({ page }) => {
    await page.goto(`/admin/territory/models/${model.id}`)
    await expect(page.getByText('Planning').first()).toBeVisible()
  })

  test('has activate button for planning model', async ({ page }) => {
    await page.goto(`/admin/territory/models/${model.id}`)
    await expect(page.getByRole('button', { name: /Activate/ })).toBeVisible()
  })

  test('has delete button', async ({ page }) => {
    await page.goto(`/admin/territory/models/${model.id}`)
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('has territories link', async ({ page }) => {
    await page.goto(`/admin/territory/models/${model.id}`)
    await expect(page.locator('main').getByRole('button', { name: /Territories/ })).toBeVisible()
  })

  test('shows save and cancel buttons', async ({ page }) => {
    await page.goto(`/admin/territory/models/${model.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })
})

test.describe('Territory list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays territory api names after selecting model', async ({ page }) => {
    await page.goto('/admin/territory/territories')
    // Wait for models to load and select first model
    await page.waitForResponse('**/api/v1/admin/territory/models?*')
    await expect(page.locator('main').getByText('north_america')).toBeVisible()
  })

  test('shows territory labels', async ({ page }) => {
    await page.goto('/admin/territory/territories')
    await page.waitForResponse('**/api/v1/admin/territory/models?*')
    await expect(page.locator('main').getByText('North America').first()).toBeVisible()
    await expect(page.locator('main').getByText('US East')).toBeVisible()
  })

  test('has create territory button', async ({ page }) => {
    await page.goto('/admin/territory/territories')
    await page.waitForResponse('**/api/v1/admin/territory/models?*')
    await expect(page.getByText('Create territory')).toBeVisible()
  })

  test('clicking territory navigates to detail', async ({ page }) => {
    await page.goto('/admin/territory/territories')
    await page.waitForResponse('**/api/v1/admin/territory/models?*')
    await page.locator('main').getByText('north_america').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/territory/territories/${mockTerritories[0].id}`),
    )
  })
})

test.describe('Territory create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form', async ({ page }) => {
    await page.goto('/admin/territory/territories/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/territory/territories/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })
})

test.describe('Territory detail page', () => {
  const territory = mockTerritories[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays territory heading', async ({ page }) => {
    await page.goto(`/admin/territory/territories/${territory.id}`)
    await expect(
      page.getByRole('heading', { name: territory.label }),
    ).toBeVisible()
  })

  test('shows tabs: info, users, objects', async ({ page }) => {
    await page.goto(`/admin/territory/territories/${territory.id}`)
    await expect(page.getByRole('tab', { name: /General/ })).toBeVisible()
    await expect(page.getByRole('tab', { name: /Users/ })).toBeVisible()
    await expect(page.getByRole('tab', { name: /Objects/ })).toBeVisible()
  })

  test('has delete button', async ({ page }) => {
    await page.goto(`/admin/territory/territories/${territory.id}`)
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('has save and cancel buttons on info tab', async ({ page }) => {
    await page.goto(`/admin/territory/territories/${territory.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('can switch to users tab', async ({ page }) => {
    await page.goto(`/admin/territory/territories/${territory.id}`)
    await page.getByRole('tab', { name: /Users/ }).click()
    // Should show the user ID from mock data
    await expect(page.getByText('u1111111').first()).toBeVisible()
  })

  test('can switch to objects tab', async ({ page }) => {
    await page.goto(`/admin/territory/territories/${territory.id}`)
    await page.getByRole('tab', { name: /Objects/ }).click()
    // Should show the access level from mock data
    await expect(page.getByText('read_write').first()).toBeVisible()
  })
})
