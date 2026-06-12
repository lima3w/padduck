import { defineConfig } from '@playwright/test'

// The stack must already be running (make e2e boots it via
// docker-compose.e2e.yml). E2E_BASE_URL points at the frontend.
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://127.0.0.1:3000',
    trace: 'retain-on-failure',
  },
})
