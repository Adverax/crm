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
    await expect(main.getByText('Все пользователи')).toBeVisible()
    await expect(main.getByText('Отдел продаж')).toBeVisible()
  })

  test('shows group type labels in Russian', async ({ page }) => {
    await page.goto('/admin/security/groups')
    const main = page.locator('main')
    await expect(main.getByText('Публичная').first()).toBeVisible()
  })

  test('shows Роль type for role group', async ({ page }) => {
    await page.goto('/admin/security/groups')
    const main = page.locator('main')
    await expect(main.getByText('Роль', { exact: true }).first()).toBeVisible()
  })

  test('has create group button', async ({ page }) => {
    await page.goto('/admin/security/groups')
    await expect(page.getByText('Создать группу')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/groups')
    await page.getByText('Создать группу').click()
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
    await expect(page.getByText('Тип группы').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/groups/new')
    await expect(page.getByRole('button', { name: 'Создать' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Отмена' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/groups')
    await page.getByText('Создать группу').click()
    await expect(page).toHaveURL(/\/admin\/security\/groups\/new/)
    await page.getByRole('button', { name: 'Отмена' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/groups/)
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/security/groups/new')

    await page.locator('#apiName').fill('test_group')
    await page.locator('#label').fill('Тестовая группа')

    const requestPromise = page.waitForRequest('**/api/v1/admin/security/groups')
    await page.getByRole('button', { name: 'Создать' }).click()

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
    await expect(page.getByRole('tab', { name: /Основное/ })).toBeVisible()
    await expect(page.getByRole('tab', { name: /Участники/ })).toBeVisible()
  })

  test('info tab shows group details', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await expect(page.getByText('API Name').first()).toBeVisible()
    await expect(page.getByText('Тип группы').first()).toBeVisible()
  })

  test('has delete button', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await expect(page.getByRole('button', { name: /Удалить/ })).toBeVisible()
  })

  test('can switch to members tab', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await page.getByRole('tab', { name: /Участники/ }).click()
    await expect(page.getByText('Добавить участника')).toBeVisible()
  })

  test('members tab shows member list', async ({ page }) => {
    await page.goto(`/admin/security/groups/${group.id}`)
    await page.getByRole('tab', { name: /Участники/ }).click()
    // Members table should be visible with user info
    await expect(page.getByText('Иван Иванов').first()).toBeVisible()
  })
})
