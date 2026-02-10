import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockRoles } from './fixtures/mock-api'

test.describe('Role list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays role api names', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const main = page.locator('main')
    await expect(main.getByText('ceo')).toBeVisible()
    await expect(main.getByText('sales_manager')).toBeVisible()
  })

  test('shows role labels in table', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const main = page.locator('main')
    // "Генеральный директор" appears twice: as label for ceo row and as parent for sales_manager row
    await expect(main.getByText('Менеджер по продажам')).toBeVisible()
  })

  test('has create role button', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await expect(page.getByText('Создать роль')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await page.getByText('Создать роль').click()
    await expect(page).toHaveURL(/\/admin\/security\/roles\/new/)
  })

  test('clicking role navigates to detail', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await page.locator('main').getByText('ceo').first().click()
    await expect(page).toHaveURL(
      new RegExp(`/admin/security/roles/${mockRoles[0].id}`),
    )
  })
})

test.describe('Role create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form', async ({ page }) => {
    await page.goto('/admin/security/roles/new')
    await expect(page.locator('#apiName')).toBeVisible()
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('has parent role selector', async ({ page }) => {
    await page.goto('/admin/security/roles/new')
    await expect(page.getByText('Родительская роль').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/roles/new')
    await expect(page.getByRole('button', { name: 'Создать' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Отмена' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/roles')
    await page.getByText('Создать роль').click()
    await expect(page).toHaveURL(/\/admin\/security\/roles\/new/)
    await page.getByRole('button', { name: 'Отмена' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/roles/)
  })

  test('can fill and submit the form', async ({ page }) => {
    await page.goto('/admin/security/roles/new')

    await page.locator('#apiName').fill('new_role')
    await page.locator('#label').fill('Новая роль')
    await page.locator('#description').fill('Описание новой роли')

    const requestPromise = page.waitForRequest('**/api/v1/admin/security/roles')
    await page.getByRole('button', { name: 'Создать' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('POST')
  })
})

test.describe('Role detail page', () => {
  const role = mockRoles[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays role heading', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await expect(
      page.getByRole('heading', { name: role.label }),
    ).toBeVisible()
  })

  test('shows form with editable fields', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await expect(page.locator('#label')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await expect(page.getByRole('button', { name: 'Сохранить' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Отмена' })).toBeVisible()
    await expect(page.getByRole('button', { name: /Удалить/ })).toBeVisible()
  })

  test('can submit updated role', async ({ page }) => {
    await page.goto(`/admin/security/roles/${role.id}`)
    await page.locator('#label').fill('Обновлённая роль')

    const requestPromise = page.waitForRequest(
      `**/api/v1/admin/security/roles/${role.id}`,
    )
    await page.getByRole('button', { name: 'Сохранить' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })
})
