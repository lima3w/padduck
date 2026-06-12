import { test, expect } from '@playwright/test'
import { loginAsAdmin } from './helpers'

test('creates and deletes a network through the UI', async ({ page }) => {
  const name = `e2e-net-${Date.now()}`
  await loginAsAdmin(page)

  await page.goto('/networks')
  await page.getByRole('button', { name: '+ New Network' }).click()
  await page.getByLabel('Name').fill(name)
  await page.getByLabel('Description').fill('created by the e2e suite')
  await page.getByRole('button', { name: 'Save' }).click()

  const row = page.getByRole('row', { name: new RegExp(name) })
  await expect(row).toBeVisible()

  await row.getByRole('button', { name: 'Delete' }).click()
  await row.getByRole('button', { name: 'Yes' }).click()
  await expect(row).toBeHidden()
})
