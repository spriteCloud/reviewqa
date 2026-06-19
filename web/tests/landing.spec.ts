import { test, expect } from '@playwright/test';

test.describe('quail landing page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('.');
  });

  test('renders the wordmark with the spriteCloud colour split', async ({ page }) => {
    const wordmark = page.getByTestId('wordmark').first();
    await expect(wordmark).toBeVisible();
    await expect(wordmark.locator('.s')).toHaveText('sprite');
    await expect(wordmark.locator('.c')).toHaveText('Cloud');

    const blue = await wordmark.locator('.s').evaluate(
      (el) => getComputedStyle(el).color,
    );
    expect(blue).toBe('rgb(0, 144, 208)');
  });

  test('hero title and lead are visible', async ({ page }) => {
    await expect(page.getByTestId('hero-title')).toContainText('Tests');
    await expect(page.getByTestId('hero-title')).toContainText('write themselves');
    await expect(page.getByText('AST‑first scaffolding')).toBeVisible();
  });

  test('pixel motif has 15 blocks and the 9th is the copper peak', async ({ page }) => {
    const blocks = page.getByTestId('pixel-motif').locator('span');
    await expect(blocks).toHaveCount(15);

    const peakColor = await blocks.nth(8).evaluate(
      (el) => getComputedStyle(el).backgroundColor,
    );
    expect(peakColor).toBe('rgb(192, 128, 90)');
  });

  test('feature grid shows three cards', async ({ page }) => {
    await expect(page.getByTestId('feature-grid').locator('.card')).toHaveCount(3);
  });

  test('install CTA points at the marketplace listing', async ({ page }) => {
    const cta = page.getByTestId('cta-install');
    await expect(cta).toBeVisible();
    await expect(cta).toHaveAttribute(
      'href',
      'https://github.com/marketplace/actions/quail',
    );
  });

  test('usage snippet references spriteCloud/quail', async ({ page }) => {
    await expect(page.getByTestId('snippet')).toContainText('spriteCloud/quail@v1');
  });
});
