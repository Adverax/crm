import { test, expect } from '@playwright/test'
import { setupAllRoutes, mockFunctions } from './fixtures/mock-api'

test.describe('ExpressionBuilder — Functions tab', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('shows Functions tab in ExpressionBuilder', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    // Add a parameter so field picker is visible
    await page.locator('[data-testid="add-param-btn"]').click()
    await page.locator('[data-testid="param-name-0"]').fill('x')

    // Focus the editor to reveal toolbar
    await page.locator('[data-testid="expression-builder"] [data-testid="codemirror-editor"]').first().click()

    // Open helper popover
    await page.locator('[data-testid="helper-btn"]').click()

    await expect(page.locator('[data-testid="functions-tab"]')).toBeVisible()
  })

  test('FunctionPicker shows custom functions from mock', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await page.locator('[data-testid="add-param-btn"]').click()
    await page.locator('[data-testid="param-name-0"]').fill('x')

    // Focus the editor to reveal toolbar
    await page.locator('[data-testid="expression-builder"] [data-testid="codemirror-editor"]').first().click()

    // Open helper popover and switch to Functions tab
    await page.locator('[data-testid="helper-btn"]').click()
    await page.locator('[data-testid="functions-tab"]').click()

    // Check that custom functions are rendered
    await expect(
      page.locator('[data-testid="function-picker"]').getByText(`fn.${mockFunctions[0].name}`),
    ).toBeVisible()
    await expect(
      page.locator('[data-testid="function-picker"]').getByText(`fn.${mockFunctions[1].name}`),
    ).toBeVisible()
  })

  test('clicking function in FunctionPicker inserts text into editor', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')
    await page.locator('[data-testid="add-param-btn"]').click()
    await page.locator('[data-testid="param-name-0"]').fill('x')

    // Focus the editor to reveal toolbar
    await page.locator('[data-testid="expression-builder"] [data-testid="codemirror-editor"]').first().click()

    // Open helper popover and switch to Functions tab
    await page.locator('[data-testid="helper-btn"]').click()
    await page.locator('[data-testid="functions-tab"]').click()

    // Click on the first custom function
    const fnButton = page.locator('[data-testid="function-picker"] button').filter({
      hasText: `fn.${mockFunctions[0].name}`,
    })
    await fnButton.click()

    // Check that text was inserted into CodeMirror
    const editorContent = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await expect(editorContent).toContainText(`fn.${mockFunctions[0].name}`)
  })
})

test.describe('ExpressionBuilder — Preview', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('preview toggle shows/hides preview section', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')

    // Type something first so expression is not empty
    const editor = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await editor.click()
    await page.keyboard.type('42')

    // Preview should not be visible initially
    await expect(page.locator('[data-testid="expression-preview"]')).not.toBeVisible()

    // Click preview toggle
    await page.locator('[data-testid="preview-toggle"]').click()

    // Preview should be visible
    await expect(page.locator('[data-testid="expression-preview"]')).toBeVisible()
  })

  test('ExpressionPreview shows result for simple expression', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')

    // Type expression in CodeMirror editor
    const editor = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await editor.click()
    await page.keyboard.type('1 + 2')

    // Show preview
    await page.locator('[data-testid="preview-toggle"]').click()

    // Wait for debounced evaluation
    await expect(page.locator('[data-testid="preview-result"]')).toBeVisible({ timeout: 2000 })
    await expect(page.locator('[data-testid="preview-result"]')).toContainText('3')
  })

  test('ExpressionPreview shows error for invalid expression', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')

    // Type invalid expression
    const editor = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await editor.click()
    await page.keyboard.type('1 +')

    // Show preview
    await page.locator('[data-testid="preview-toggle"]').click()

    // Wait for debounced evaluation
    await expect(page.locator('[data-testid="preview-error"]')).toBeVisible({ timeout: 2000 })
  })

  test('function_body context shows parameter test inputs', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')

    // Add parameters
    await page.locator('[data-testid="add-param-btn"]').click()
    await page.locator('[data-testid="param-name-0"]').fill('amount')

    // Type expression
    const editor = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await editor.click()
    await page.keyboard.type('amount')

    // Show preview
    await page.locator('[data-testid="preview-toggle"]').click()

    // Parameter input should be visible
    await expect(page.locator('[data-testid="preview-param-amount"]')).toBeVisible()
  })
})

test.describe('ExpressionBuilder — Autocomplete', () => {
  test.beforeEach(async ({ page }) => {
    await setupAllRoutes(page)
  })

  test('autocomplete appears after typing fn.', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')

    // Type in CodeMirror editor
    const editor = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await editor.click()
    await page.keyboard.type('fn.')

    // Wait for autocomplete tooltip
    await expect(page.locator('.cm-tooltip-autocomplete')).toBeVisible({ timeout: 2000 })
  })

  test('autocomplete appears for parameter names in function_body context', async ({ page }) => {
    await page.goto('/admin/metadata/functions/new')

    // Add a parameter named "amount"
    await page.locator('[data-testid="add-param-btn"]').click()
    await page.locator('[data-testid="param-name-0"]').fill('amount')

    // Wait for computed properties to update
    await page.waitForTimeout(300)

    // Type "amo" in the editor — autocomplete should show "amount" parameter
    const editor = page.locator('[data-testid="codemirror-editor"] .cm-content')
    await editor.click()
    await page.keyboard.type('amo')

    // Wait for autocomplete tooltip
    await expect(page.locator('.cm-tooltip-autocomplete')).toBeVisible({ timeout: 3000 })
  })
})
