import { expect } from '@playwright/test'

export const ADMIN_USER = 'admin'
export const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD || 'e2e-admin-password'

// Logs in through the real form and waits for the dashboard.
// On a fresh install the admin is redirected to the first-run setup wizard first — this helper
// detects that and clicks "Skip setup for now" so subsequent tests land on the dashboard as expected.
export async function loginAsAdmin(page) {
  await page.goto('/login')
  await page.getByLabel('Username').fill(ADMIN_USER)
  await page.getByLabel('Password').fill(ADMIN_PASSWORD)
  await page.getByRole('button', { name: 'Sign In' }).click()
  try {
    await page.waitForURL('**/setup', { timeout: 3000 })
    await page.getByRole('button', { name: 'Skip setup for now' }).click()
  } catch {
    // Setup wizard already completed — already heading to the dashboard.
  }
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
}
