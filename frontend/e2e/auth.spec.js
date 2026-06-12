import { test, expect } from '@playwright/test'
import { loginAsAdmin, ADMIN_USER } from './helpers'

test('rejects wrong credentials with a generic error', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Username').fill(ADMIN_USER)
  await page.getByLabel('Password').fill('definitely-wrong')
  await page.getByRole('button', { name: 'Sign In' }).click()
  await expect(page.getByText('invalid username or password')).toBeVisible()
  await expect(page).toHaveURL(/\/login/)
})

test('logs in and lands on the dashboard', async ({ page }) => {
  await loginAsAdmin(page)
  await expect(page.getByRole('heading', { name: 'Recent Activity' })).toBeVisible()
})

test('redirects unauthenticated visitors to login', async ({ page }) => {
  await page.goto('/networks')
  await expect(page).toHaveURL(/\/login/)
})

test('session survives a page reload', async ({ page }) => {
  await loginAsAdmin(page)
  await page.reload()
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
  await expect(page).not.toHaveURL(/\/login/)
})
