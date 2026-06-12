import { defineConfig, devices } from '@playwright/test';

const prodURL = process.env.PROD_URL;

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['json', { outputFile: 'playwright-report.json' }], ['list']] : 'list',
  use: {
    baseURL: prodURL ?? 'http://127.0.0.1:4173',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  // Local dev server only — production runs target a live URL.
  webServer: prodURL
    ? undefined
    : {
        command: 'npx --yes http-server . -p 4173 -s',
        url: 'http://127.0.0.1:4173',
        reuseExistingServer: !process.env.CI,
        timeout: 60_000,
      },
});
