import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockObjects } from './fixtures/mock-api'

test.describe('Object list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays object list with data', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    const main = page.locator('main')
    await expect(main.getByText('account', { exact: true })).toBeVisible()
    await expect(main.getByText('custom_obj', { exact: true })).toBeVisible()
  })

  test('shows object labels', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    const main = page.locator('main')
    await expect(main.getByText('Account', { exact: true })).toBeVisible()
    await expect(main.getByText('Custom Object', { exact: true })).toBeVisible()
  })

  test('has create object button', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await expect(page.getByText('Create object')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.getByText('Create object').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/objects\/new/)
  })

  test('clicking object row navigates to detail', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.locator('main').getByText('account').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/metadata/objects/${mockObjects[0].id}`),
    )
  })
})

test.describe('Object create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form with all fields', async ({ page }) => {
    await page.goto('/admin/metadata/objects/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#pluralLabel')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/objects/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/objects')
    await page.getByText('Create object').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/objects\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/objects/)
  })

  test('has visibility selector', async ({ page }) => {
    await page.goto('/admin/metadata/objects/new')
    await expect(page.getByText('Visibility (OWD)').first()).toBeVisible()
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/metadata/objects/new')

    await page.locator('#apiName').fill('test_object')
    await page.locator('#label').fill('Test Object')
    await page.locator('#pluralLabel').fill('Test Objects')
    await page.locator('#description').fill('Test object description')

    const requestPromise = page.waitForRequest('**/api/v1/admin/metadata/objects')
    await page.getByRole('button', { name: 'Create' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Object detail page', () => {
  const obj = mockObjects[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays object details', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await expect(page.getByRole('heading', { name: obj.label })).toBeVisible()
  })

  test('shows info tab with form fields', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#pluralLabel')).toBeVisible()
  })

  test('shows visibility selector in detail form', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await expect(page.getByText('Visibility (OWD)').first()).toBeVisible()
  })

  test('shows fields tab with field count', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    const fieldsTab = page.getByRole('tab', { name: /Fields/ })
    await expect(fieldsTab).toBeVisible()
  })

  test('can switch to fields tab and see field data', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await page.getByRole('tab', { name: /Fields/ }).click()
    // Check for field api_name which is unique to the fields table
    await expect(page.locator('main').getByText('name').first()).toBeVisible()
  })

  test('has save and cancel buttons', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('shows object type badge', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await expect(page.getByText('Standard')).toBeVisible()
  })
})

test.describe('Custom object detail page', () => {
  const obj = mockObjects[1]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows Custom badge for custom objects', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await expect(page.getByText('Custom', { exact: true })).toBeVisible()
  })

  test('loads and displays custom object heading', async ({ page }) => {
    await page.goto(`/admin/metadata/objects/${obj.id}`)
    await expect(
      page.getByRole('heading', { name: obj.label }),
    ).toBeVisible()
  })
})
