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

  test('sidebar shows navigation groups', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Schema')).toBeVisible()
    await expect(sidebar.getByText('Presentation')).toBeVisible()
    await expect(sidebar.getByText('Automation')).toBeVisible()
    await expect(sidebar.getByText('Security')).toBeVisible()
    await expect(sidebar.getByText('Users')).toBeVisible()
  })

  test('sidebar shows security group with toggle', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Security')).toBeVisible()
  })

  test('security group expands and shows children', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Security').click()
    await expect(sidebar.getByText('Roles')).toBeVisible()
    await expect(sidebar.getByText('Permission Sets')).toBeVisible()
    await expect(sidebar.getByText('Profiles')).toBeVisible()
    await expect(sidebar.getByText('Groups')).toBeVisible()
    await expect(sidebar.getByText('Sharing Rules')).toBeVisible()
  })

  test('redirects /admin to /admin/metadata/objects', async ({ page }) => {
    await page.goto('/admin')
    await expect(page).toHaveURL(/\/admin\/metadata\/objects/)
  })

  test('navigates to objects via sidebar link', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Schema').click()
    await sidebar.getByText('Objects').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/objects/)
  })

  test('navigates to roles via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Security').click()
    await sidebar.getByText('Roles').click()
    await expect(page).toHaveURL(/\/admin\/security\/roles/)
  })

  test('navigates to users via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Users').click()
    await expect(page).toHaveURL(/\/admin\/security\/users/)
  })

  test('security section auto-expands when on security route', async ({ page }) => {
    await page.goto('/admin/security/roles')
    const sidebar = page.locator('aside')
    await expect(sidebar.getByText('Roles')).toBeVisible()
    await expect(sidebar.getByText('Permission Sets')).toBeVisible()
    await expect(sidebar.getByText('Profiles')).toBeVisible()
    await expect(sidebar.getByText('Groups')).toBeVisible()
    await expect(sidebar.getByText('Sharing Rules')).toBeVisible()
  })

  test('navigates to groups via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Security').click()
    await sidebar.getByText('Groups').click()
    await expect(page).toHaveURL(/\/admin\/security\/groups/)
  })

  test('navigates to sharing rules via sidebar', async ({ page }) => {
    await page.goto('/admin')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Security').click()
    await sidebar.getByText('Sharing Rules').click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules/)
  })
})
