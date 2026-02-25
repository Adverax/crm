import { test, expect } from '@playwright/test'
import {
  setupAllRoutes,
  mockProcedures,
  mockProcedureVersions,
} from './fixtures/mock-api'

test.beforeEach(async ({ page }) => {
  await setupAllRoutes(page)
})

// ─── List page ───────────────────────────────────────────────

test.describe('Procedure list page', () => {
  test('shows procedures from API', async ({ page }) => {
    await page.goto('/admin/metadata/procedures')
    await expect(page.getByTestId('procedure-row')).toHaveCount(mockProcedures.length)
    await expect(page.getByText('create_account')).toBeVisible()
    await expect(page.getByText('send_welcome')).toBeVisible()
  })

  test('shows create button', async ({ page }) => {
    await page.goto('/admin/metadata/procedures')
    await expect(page.getByTestId('create-procedure-btn')).toBeVisible()
  })

  test('create button navigates to create page', async ({ page }) => {
    await page.goto('/admin/metadata/procedures')
    await page.getByTestId('create-procedure-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/procedures\/new/)
  })

  test('clicking row navigates to detail page', async ({ page }) => {
    await page.goto('/admin/metadata/procedures')
    await page.getByTestId('procedure-row').first().click()
    await expect(page).toHaveURL(/\/admin\/metadata\/procedures\/pr111111/)
  })

  test('shows status badges', async ({ page }) => {
    await page.goto('/admin/metadata/procedures')
    await expect(page.getByText('Published + Draft')).toBeVisible()
    await expect(page.getByText('Draft').first()).toBeVisible()
  })

  test('shows empty state when no procedures', async ({ page }) => {
    await page.route('**/api/v1/admin/procedures', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: [] } })
      }
      return route.continue()
    })
    await page.goto('/admin/metadata/procedures')
    await expect(page.getByText('No procedures')).toBeVisible()
  })
})

// ─── Create page ─────────────────────────────────────────────

test.describe('Procedure create page', () => {
  test('renders form with all fields', async ({ page }) => {
    await page.goto('/admin/metadata/procedures/new')
    await expect(page.getByTestId('field-code')).toBeVisible()
    await expect(page.getByTestId('field-name')).toBeVisible()
    await expect(page.getByTestId('field-description')).toBeVisible()
  })

  test('has create and cancel buttons', async ({ page }) => {
    await page.goto('/admin/metadata/procedures/new')
    await expect(page.getByTestId('submit-btn')).toBeVisible()
    await expect(page.getByTestId('cancel-btn')).toBeVisible()
  })

  test('cancel navigates back to list', async ({ page }) => {
    await page.goto('/admin/metadata/procedures/new')
    await page.getByTestId('cancel-btn').click()
    await expect(page).toHaveURL(/\/admin\/metadata\/procedures$/)
  })

  test('submitting form sends POST request', async ({ page }) => {
    let postCalled = false
    await page.route('**/api/v1/admin/procedures', (route) => {
      if (route.request().method() === 'POST') {
        postCalled = true
        return route.fulfill({
          json: {
            data: {
              procedure: { ...mockProcedures[0], id: 'new-proc-id' },
              draft_version: mockProcedureVersions[0],
              published_version: null,
            },
          },
        })
      }
      return route.continue()
    })

    await page.goto('/admin/metadata/procedures/new')
    await page.getByTestId('field-code').fill('test_proc')
    await page.getByTestId('field-name').fill('Test Procedure')
    await page.getByTestId('submit-btn').click()

    await page.waitForTimeout(500)
    expect(postCalled).toBe(true)
  })
})

// ─── Detail page ─────────────────────────────────────────────

test.describe('Procedure detail page', () => {
  const detailUrl = `/admin/metadata/procedures/${mockProcedures[0].id}`

  test('shows procedure heading', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByRole('heading', { name: 'create_account' })).toBeVisible()
  })

  test('shows tabs', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('tab-definition')).toBeVisible()
    await expect(page.getByTestId('tab-versions')).toBeVisible()
    await expect(page.getByTestId('tab-settings')).toBeVisible()
    await expect(page.getByTestId('tab-dry-run')).toBeVisible()
  })

  test('shows save draft and publish buttons', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('save-draft-btn')).toBeVisible()
    await expect(page.getByTestId('publish-btn')).toBeVisible()
  })

  test('shows rollback button for published procedure', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('rollback-btn')).toBeVisible()
  })

  test('shows delete button', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('delete-procedure-btn')).toBeVisible()
  })

  test('versions tab shows version history', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('tab-versions').click()
    await expect(page.getByTestId('version-row')).toHaveCount(2)
    await expect(page.getByText('v2')).toBeVisible()
    await expect(page.getByText('v1')).toBeVisible()
  })

  test('settings tab shows form', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('tab-settings').click()
    await expect(page.getByTestId('field-name')).toBeVisible()
    await expect(page.getByTestId('field-description')).toBeVisible()
    await expect(page.getByTestId('save-settings-btn')).toBeVisible()
  })

  test('dry-run tab shows input and run button', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('tab-dry-run').click()
    await expect(page.getByTestId('dry-run-input')).toBeVisible()
    await expect(page.getByTestId('dry-run-btn')).toBeVisible()
  })

  test('dry-run shows result after execution', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('tab-dry-run').click()
    await page.getByTestId('dry-run-btn').click()
    await expect(page.getByText('Success')).toBeVisible()
  })
})

// ─── Constructor ─────────────────────────────────────────────

test.describe('Procedure constructor', () => {
  const detailUrl = `/admin/metadata/procedures/${mockProcedures[0].id}`

  test('shows command cards from definition', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('command-card')).toHaveCount(2)
  })

  test('shows add command button', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('add-command-btn')).toBeVisible()
  })

  test('remove command button removes a command', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('command-card')).toHaveCount(2)
    await page.getByTestId('remove-command-btn').first().click()
    await expect(page.getByTestId('command-card')).toHaveCount(1)
  })
})

// ─── Sidebar ─────────────────────────────────────────────────

test.describe('Sidebar navigation', () => {
  test('procedures link appears in sidebar', async ({ page }) => {
    await page.goto('/admin/metadata/procedures')
    await expect(page.getByRole('complementary').getByRole('link', { name: 'Procedures' })).toBeVisible()
  })

  test('procedures link navigates correctly', async ({ page }) => {
    await page.goto('/admin')
    await page.getByRole('complementary').getByRole('link', { name: 'Procedures' }).click()
    await expect(page).toHaveURL(/\/admin\/metadata\/procedures/)
  })
})

// ─── Try/Catch & Retry ──────────────────────────────────────

test.describe('Try/Catch command', () => {
  const detailUrl = `/admin/metadata/procedures/${mockProcedures[0].id}`

  test('flow.try appears in command picker', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('add-command-btn').click()
    await expect(page.getByTestId('cmd-flow.try')).toBeVisible()
  })

  test('adding flow.try shows try/catch blocks', async ({ page }) => {
    await page.goto(detailUrl)
    await page.getByTestId('add-command-btn').click()
    await page.getByTestId('cmd-flow.try').click()
    await expect(page.getByTestId('flow-try-info')).toBeVisible()
    await expect(page.getByTestId('try-block-label')).toBeVisible()
    await expect(page.getByTestId('catch-block-label')).toBeVisible()
  })
})

// ─── KeyValueEditor ─────────────────────────────────────────

test.describe('KeyValueEditor', () => {
  const detailUrl = `/admin/metadata/procedures/${mockProcedures[0].id}`

  test('record.create shows data key-value editor', async ({ page }) => {
    await page.goto(detailUrl)
    // Second command card is record.create with data
    const card = page.getByTestId('command-card').nth(1)
    await expect(card.getByTestId('kv-editor')).toBeVisible()
  })

  test('record.create data editor shows existing entries', async ({ page }) => {
    await page.goto(detailUrl)
    const card = page.getByTestId('command-card').nth(1)
    const kvEditor = card.getByTestId('kv-editor')
    // mock has data: { Name: '$.input.name' }
    await expect(kvEditor.locator('input').first()).toHaveValue('Name')
    await expect(kvEditor.locator('input').nth(1)).toHaveValue('$.input.name')
  })

  test('add button appends a new entry', async ({ page }) => {
    // Use compute.transform with a known value map
    const procWithTransform = {
      procedure: mockProcedures[0],
      draft_version: {
        ...mockProcedureVersions[0],
        definition: {
          commands: [
            { type: 'compute.transform', as: 'step', value: { total: '$.input.a + $.input.b' } },
          ],
          result: {},
        },
      },
      published_version: mockProcedureVersions[1],
    }
    await page.route(`**/api/v1/admin/procedures/${mockProcedures[0].id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: procWithTransform } })
      }
      return route.continue()
    })
    await page.goto(detailUrl)
    const kvEditor = page.getByTestId('command-card').first().getByTestId('kv-editor')
    await expect(kvEditor.getByRole('button', { name: 'Remove' })).toHaveCount(1)
    await kvEditor.getByTestId('kv-add-btn').click()
    await expect(kvEditor.getByRole('button', { name: 'Remove' })).toHaveCount(2)
  })

  test('compute.transform shows value key-value editor', async ({ page }) => {
    const procWithTransform = {
      procedure: mockProcedures[0],
      draft_version: {
        ...mockProcedureVersions[0],
        definition: {
          commands: [
            {
              type: 'compute.transform',
              as: 'totals',
              value: { greeting: '$.input.first + " " + $.input.last', tax: '$.input.amount * 0.2' },
            },
          ],
          result: {},
        },
      },
      published_version: mockProcedureVersions[1],
    }
    await page.route(`**/api/v1/admin/procedures/${mockProcedures[0].id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: procWithTransform } })
      }
      return route.continue()
    })
    await page.goto(detailUrl)
    const card = page.getByTestId('command-card').first()
    const kvEditor = card.getByTestId('kv-editor')
    await expect(kvEditor).toBeVisible()
    // Should have 2 entries with remove buttons
    await expect(kvEditor.getByRole('button', { name: 'Remove' })).toHaveCount(2)
    await expect(kvEditor.locator('input').first()).toHaveValue('greeting')
  })

  test('flow.call shows input key-value editor', async ({ page }) => {
    const procWithCall = {
      procedure: mockProcedures[0],
      draft_version: {
        ...mockProcedureVersions[0],
        definition: {
          commands: [
            {
              type: 'flow.call',
              as: 'sub',
              procedure: 'other_proc',
              input: { name: '$.input.name' },
            },
          ],
          result: {},
        },
      },
      published_version: mockProcedureVersions[1],
    }
    await page.route(`**/api/v1/admin/procedures/${mockProcedures[0].id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: procWithCall } })
      }
      return route.continue()
    })
    await page.goto(detailUrl)
    const card = page.getByTestId('command-card').first()
    await expect(card.getByTestId('kv-editor')).toBeVisible()
    await expect(card.getByTestId('kv-editor').locator('input').first()).toHaveValue('name')
  })

  test('integration.http shows headers key-value editor', async ({ page }) => {
    const procWithHttp = {
      procedure: mockProcedures[0],
      draft_version: {
        ...mockProcedureVersions[0],
        definition: {
          commands: [
            {
              type: 'integration.http',
              as: 'call',
              credential: 'my_api',
              method: 'POST',
              path: '/endpoint',
              headers: { 'Content-Type': 'application/json' },
            },
          ],
          result: {},
        },
      },
      published_version: mockProcedureVersions[1],
    }
    await page.route(`**/api/v1/admin/procedures/${mockProcedures[0].id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: procWithHttp } })
      }
      return route.continue()
    })
    await page.goto(detailUrl)
    const card = page.getByTestId('command-card').first()
    await expect(card.getByTestId('kv-editor')).toBeVisible()
    await expect(card.getByTestId('kv-editor').locator('input').first()).toHaveValue('Content-Type')
  })

  test('empty key-value editor shows placeholder text', async ({ page }) => {
    const procWithEmptyTransform = {
      procedure: mockProcedures[0],
      draft_version: {
        ...mockProcedureVersions[0],
        definition: {
          commands: [
            { type: 'compute.transform', as: 'empty_step' },
          ],
          result: {},
        },
      },
      published_version: mockProcedureVersions[1],
    }
    await page.route(`**/api/v1/admin/procedures/${mockProcedures[0].id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: procWithEmptyTransform } })
      }
      return route.continue()
    })
    await page.goto(detailUrl)
    const card = page.getByTestId('command-card').first()
    await expect(card.getByText('No entries')).toBeVisible()
  })
})

test.describe('Retry config', () => {
  const detailUrl = `/admin/metadata/procedures/${mockProcedures[0].id}`

  test('retry section is visible on command card', async ({ page }) => {
    await page.goto(detailUrl)
    await expect(page.getByTestId('retry-section').first()).toBeVisible()
  })

  test('retry config fields render when command has retry', async ({ page }) => {
    // Override route to return a procedure with retry config in a command
    const procWithRetry = {
      procedure: mockProcedures[0],
      draft_version: {
        ...mockProcedureVersions[0],
        definition: {
          commands: [
            {
              type: 'integration.http',
              as: 'api_call',
              credential: 'stripe',
              method: 'POST',
              path: '/charge',
              retry: { max_attempts: 3, delay_ms: 1000, backoff_mult: 2 },
            },
          ],
          result: {},
        },
      },
      published_version: mockProcedureVersions[1],
    }
    await page.route(`**/api/v1/admin/procedures/${mockProcedures[0].id}`, (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: { data: procWithRetry } })
      }
      return route.continue()
    })

    await page.goto(detailUrl)
    await expect(page.getByTestId('retry-config')).toBeVisible()
    await expect(page.getByTestId('retry-max-attempts')).toBeVisible()
    await expect(page.getByTestId('retry-delay-ms')).toBeVisible()
    await expect(page.getByTestId('retry-backoff-mult')).toBeVisible()
  })
})
