import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockRecords, mockDescribeList } from './fixtures/mock-api'

test.describe('Record list page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays record names in table', async ({ page }) => {
    await page.goto('/app/Account')
    const main = page.locator('main')
    await expect(main.getByText('Acme Corp')).toBeVisible()
    await expect(main.getByText('Globex Inc')).toBeVisible()
  })

  test('displays table column headers from metadata', async ({ page }) => {
    await page.goto('/app/Account')
    const main = page.locator('main')
    await expect(main.getByText('Название')).toBeVisible()
  })

  test('has create button', async ({ page }) => {
    await page.goto('/app/Account')
    await expect(page.getByRole('button', { name: 'Создать' })).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/app/Account')
    await page.getByRole('button', { name: 'Создать' }).click()
    await expect(page).toHaveURL(/\/app\/Account\/new/)
  })

  test('clicking row navigates to detail', async ({ page }) => {
    await page.goto('/app/Account')
    await page.locator('main').getByText('Acme Corp').click()
    await expect(page).toHaveURL(
      new RegExp(`/app/Account/${mockRecords[0].Id}`),
    )
  })

  test('shows empty state when no records', async ({ page }) => {
    await page.route('**/api/v1/records/Account?*', (route) => {
      return route.fulfill({
        json: { data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 0 } },
      })
    })
    await page.route('**/api/v1/records/Account', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: { data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 0 } },
        })
      }
      return route.continue()
    })
    await page.goto('/app/Account')
    await expect(page.getByText('Нет записей')).toBeVisible()
  })
})

test.describe('Record create page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('renders form with fields from metadata', async ({ page }) => {
    await page.goto('/app/Account/new')
    await expect(page.getByLabel('Название')).toBeVisible()
    await expect(page.getByLabel('Телефон')).toBeVisible()
  })

  test('has create and cancel buttons', async ({ page }) => {
    await page.goto('/app/Account/new')
    await expect(page.getByRole('button', { name: 'Создать' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Отмена' })).toBeVisible()
  })

  test('cancel button navigates back', async ({ page }) => {
    await page.goto('/app/Account')
    await page.getByRole('button', { name: 'Создать' }).click()
    await expect(page).toHaveURL(/\/app\/Account\/new/)
    await page.getByRole('button', { name: 'Отмена' }).click()
    await expect(page).toHaveURL(/\/app\/Account$/)
  })

  test('submitting form sends POST request', async ({ page }) => {
    let postCalled = false
    await page.route('**/api/v1/records/Account', (route) => {
      if (route.request().method() === 'POST') {
        postCalled = true
        return route.fulfill({
          status: 201,
          json: { data: { id: 'rec33333-3333-3333-3333-333333333333' } },
        })
      }
      return route.continue()
    })
    await page.goto('/app/Account/new')
    await page.getByLabel('Название').fill('Test Company')
    await page.getByRole('button', { name: 'Создать' }).click()
    await page.waitForTimeout(500)
    expect(postCalled).toBe(true)
  })
})

test.describe('Record detail page', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('displays record data in form', async ({ page }) => {
    await page.goto(`/app/Account/${mockRecords[0].Id}`)
    await expect(page.getByRole('heading', { name: 'Acme Corp' })).toBeVisible()
  })

  test('has save and cancel buttons', async ({ page }) => {
    await page.goto(`/app/Account/${mockRecords[0].Id}`)
    await expect(page.getByRole('button', { name: 'Сохранить' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Отмена' })).toBeVisible()
  })

  test('has delete button', async ({ page }) => {
    await page.goto(`/app/Account/${mockRecords[0].Id}`)
    await expect(page.getByRole('button', { name: 'Удалить' })).toBeVisible()
  })

  test('save button sends PUT request', async ({ page }) => {
    let putCalled = false
    await page.route(`**/api/v1/records/Account/${mockRecords[0].Id}`, (route) => {
      if (route.request().method() === 'PUT') {
        putCalled = true
        return route.fulfill({ json: { data: { success: true } } })
      }
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: mockRecords[0] } })
      }
      return route.continue()
    })
    await page.goto(`/app/Account/${mockRecords[0].Id}`)
    await page.getByRole('button', { name: 'Сохранить' }).click()
    await page.waitForTimeout(500)
    expect(putCalled).toBe(true)
  })

  test('delete button triggers DELETE request', async ({ page }) => {
    let deleteCalled = false
    await page.route(`**/api/v1/records/Account/${mockRecords[0].Id}`, (route) => {
      if (route.request().method() === 'DELETE') {
        deleteCalled = true
        return route.fulfill({ status: 204 })
      }
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: mockRecords[0] } })
      }
      return route.continue()
    })
    await page.goto(`/app/Account/${mockRecords[0].Id}`)
    await page.getByRole('button', { name: 'Удалить' }).click()
    // Confirm delete in dialog
    const dialog = page.locator('[role="dialog"]')
    await dialog.getByRole('button', { name: 'Удалить' }).click()
    await page.waitForTimeout(500)
    expect(deleteCalled).toBe(true)
  })
})

test.describe('App sidebar', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows objects from describe in navigation', async ({ page }) => {
    await page.goto('/app/Account')
    const sidebar = page.locator('aside')
    for (const obj of mockDescribeList) {
      await expect(sidebar.getByText(obj.plural_label)).toBeVisible()
    }
  })

  test('settings link navigates to admin', async ({ page }) => {
    await page.goto('/app/Account')
    const sidebar = page.locator('aside')
    await sidebar.getByText('Настройки').click()
    await expect(page).toHaveURL(/\/admin/)
  })
})
