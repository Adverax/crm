import { test, expect } from '@playwright/test'
import { setupAllRoutes } from './fixtures/mock-api'

test.describe('Admin layout and sidebar', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders sidebar with CRM Admin title', async ({ page }) => {
    await page.goto('/admin')
    await expect(page.locator('aside')).toBeVisible()
    await expect(page.locator('aside').getByText('CRM Admin')).toBeVisible()
  })

  test('sidebar shows top-level navigation items', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Объекты')).toBeVisible()
    await expect(sidebar.getByText('Пользователи')).toBeVisible()
  })

  test('sidebar shows security group with toggle', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Безопасность')).toBeVisible()
  })

  test('security group expands and shows children', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Безопасность').click()
    await expect(sidebar.getByText('Роли')).toBeVisible()
    await expect(sidebar.getByText('Наборы разрешений')).toBeVisible()
    await expect(sidebar.getByText('Профили')).toBeVisible()
    await expect(sidebar.getByText('Группы')).toBeVisible()
    await expect(sidebar.getByText('Правила доступа')).toBeVisible()
  })

  test('redirects /admin to /admin/metadata/objects', async ({ page }) => {
    await page.goto('/admin')
    await expect(page).toHaveURL(/\/admin\/metadata\/objects/)
  })

  test('navigates to objects via sidebar link', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Объекты').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/objects/)
  })

  test('navigates to roles via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Безопасность').click()
    await sidebar.getByText('Роли').click()
    await expect(page).toHaveURL(/\/admin\/security\/roles/)
  })

  test('navigates to users via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Пользователи').click()
    await expect(page).toHaveURL(/\/admin\/security\/users/)
  })

  test('security section auto-expands when on security route', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Роли')).toBeVisible()
    await expect(sidebar.getByText('Наборы разрешений')).toBeVisible()
    await expect(sidebar.getByText('Профили')).toBeVisible()
    await expect(sidebar.getByText('Группы')).toBeVisible()
    await expect(sidebar.getByText('Правила доступа')).toBeVisible()
  })

  test('navigates to groups via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Безопасность').click()
    await sidebar.getByText('Группы').click()
    await expect(page).toHaveURL(/\/admin\/security\/groups/)
  })

  test('navigates to sharing rules via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Безопасность').click()
    await sidebar.getByText('Правила доступа').click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules/)
  })
})
