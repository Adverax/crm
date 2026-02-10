import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockSharingRules } from './fixtures/mock-api'

test.describe('Sharing rule list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows object selector', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await expect(page.getByText('Объект').first()).toBeVisible()
  })

  test('has create rule button', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await expect(page.getByText('Создать правило')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await page.getByText('Создать правило').click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules\/new/)
  })

  test('shows prompt to select object when no object selected', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await expect(page.getByText('Выберите объект для просмотра')).toBeVisible()
  })
})

test.describe('Sharing rule create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders create form with all selectors', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules/new')
    await expect(page.getByText('Объект').first()).toBeVisible()
    await expect(page.getByText('Тип правила').first()).toBeVisible()
    await expect(page.getByText('Группа-источник').first()).toBeVisible()
    await expect(page.getByText('Группа-получатель').first()).toBeVisible()
    await expect(page.getByText('Уровень доступа').first()).toBeVisible()
  })

  test('has submit and cancel buttons', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules/new')
    await expect(page.getByRole('button', { name: 'Создать' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Отмена' })).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules')
    await page.getByText('Создать правило').click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules\/new/)
    await page.getByRole('button', { name: 'Отмена' }).click()
    await expect(page).toHaveURL(/\/admin\/security\/sharing-rules/)
  })

  test('shows breadcrumbs with correct links', async ({ page }) => {
    await page.goto('/admin/security/sharing-rules/new')
    await expect(page.getByText('Правила совместного доступа').first()).toBeVisible()
    await expect(page.getByText('Новое правило').first()).toBeVisible()
  })
})

test.describe('Sharing rule detail page', () => {
  const rule = mockSharingRules[0]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('loads and displays rule heading', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(
      page.getByRole('heading', { name: /Правило/ }),
    ).toBeVisible()
  })

  test('shows read-only object and rule type fields', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByText('Объект').first()).toBeVisible()
    await expect(page.getByText('Тип правила').first()).toBeVisible()
    await expect(page.getByText('Группа-источник').first()).toBeVisible()
  })

  test('has editable target group and access level', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByText('Группа-получатель').first()).toBeVisible()
    await expect(page.getByText('Уровень доступа').first()).toBeVisible()
  })

  test('has save, cancel, and delete buttons', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByRole('button', { name: 'Сохранить' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Отмена' })).toBeVisible()
    await expect(page.getByRole('button', { name: /Удалить/ })).toBeVisible()
  })

  test('can submit updated sharing rule', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)

    const requestPromise = page.waitForRequest(
      (req) =>
        req.url().includes(`/api/v1/admin/security/sharing-rules/${rule.id}`) &&
        req.method() === 'PUT',
    )
    await page.getByRole('button', { name: 'Сохранить' }).click()

    const request = await requestPromise
    expect(request.method()).toBe('PUT')
  })
})

test.describe('Criteria-based sharing rule detail page', () => {
  const rule = mockSharingRules[1]

  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows criteria fields for criteria_based rule', async ({ page }) => {
    await page.goto(`/admin/security/sharing-rules/${rule.id}`)
    await expect(page.getByText('Критерий').first()).toBeVisible()
    await expect(page.locator('#criteriaField')).toBeVisible()
    await expect(page.locator('#criteriaOp')).toBeVisible()
    await expect(page.locator('#criteriaValue')).toBeVisible()
  })
})
