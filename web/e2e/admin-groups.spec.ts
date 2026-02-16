import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockGroups } from './fixtures/mock-api'

test.describe('Group list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays group api names', async ({ page }) => {
    await page.goto('/admin/security/groups')
    const main = page.locator('main')
    await expect(main.getByText('all_users')).toBeVisible()
    await expect(main.getByText('sales_team')).toBeVisible()
  })

  test('shows group labels', async ({ page }) => {
    await page.goto('/admin/security/groups')
    const main = page.locator('main')
    await expect(main.getByText('All Users')).toBeVisible()
    await expect(main.getByText('Sales Team')).toBeVisible()
  })

  test('shows group type labels', async ({ page }) => {
    await page.goto('/admin/security/groups')
    const main = page.locator('main')
    await expect(main.getByText('Public').first()).toBeVisible()
  })

  test('shows Role type for role group', async ({ page }) => {
    await page.goto('/admin/security/groups')
    const main = page.locator('main')
    await expect(main.getByText('Role', { exact: true }).first()).toBeVisible()
  })

  test('has create group button', async ({ page }) => {
    await page.goto('/admin/security/groups')
    await expect(page.getByText('Create Group')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/groups')
    await page.getByText('Create Group').click()
    await expect(page).toHaveURL(/\/admin\/security\/groups\/new/)
  })

  test('clicking group navigates to detail', async ({ page }) => {
    await page.goto('/admin/security/groups')
    await page.locator('main').getByText('all_users').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/security/groups/${mockGroups[0].id}`),
    )
  })
})

test.describe('Group create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form', async ({ page }) => {
    await page.goto('/admin/security/groups/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
  })

  test('has group type selector', async ({ page }) => {
    await page.goto('/admin/security/groups/new')
    await expect(page.getByText('Group Type').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/groups/new')
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/groups')
    await page.getByText('Create Group').click()
    await expect(page).toHaveURL(/\/admin\/security\/groups\/new/)
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/groups/)
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/security/groups/new')

    await page.locator('#apiName').fill('test_group')
    await page.locator('#label').fill('Test Group')

    const requestPromise = page.waitForRequest('**/api/v1/admin/security/groups')
    await page.getByRole('button', { name: 'Create' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Group detail page', () => {
  const group = mockGroups[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays group heading', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await expect(
      page.getByRole('heading', { name: group.label }),
    ).toBeVisible()
  })

  test('shows tabs: info and members', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await expect(page.getByRole('tab', { name: /General/ })).toBeVisible()
    await expect(page.getByRole('tab', { name: /Members/ })).toBeVisible()
  })

  test('info tab shows group details', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await expect(page.getByText('API Name').first()).toBeVisible()
    await expect(page.getByText('Group Type').first()).toBeVisible()
  })

  test('has delete button', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await expect(page.getByRole('button', { name: /Delete/ })).toBeVisible()
  })

  test('can switch to members tab', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await page.getByRole('tab', { name: /Members/ }).click()
    await expect(page.getByText('Add Member')).toBeVisible()
  })

  test('members tab shows member list', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await page.getByRole('tab', { name: /Members/ }).click()
    // Members table should be visible with user info
    await expect(page.getByText('John Smith').first()).toBeVisible()
  })
})
