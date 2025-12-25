import { test, expect } from "@playwright/test";

test.describe("SRE Monitoring Platform", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    // Wait for dashboard to load (the Overview title appears once data is loaded)
    await page.waitForSelector("h1:has-text('Overview')", { timeout: 60000 });
  });

  test("dashboard loads with real stats", async ({ page }) => {
    // Should show Overview title
    await expect(page.locator("h1")).toContainText("Overview");

    // Should show real monitor count (not loading spinner)
    await expect(page.getByText(/Total Monitors/)).toBeVisible();
    await expect(page.getByText(/Healthy/)).toBeVisible();
    await expect(page.getByText(/Avg Latency/)).toBeVisible();

    // Should show scheduler status
    await expect(page.getByText(/Scheduler:/)).toBeVisible();
  });

  test("monitors page shows real monitors", async ({ page }) => {
    // Navigate to Monitors page using sidebar (exact match to avoid Quick Actions button)
    await page.getByRole("button", { name: "Monitors", exact: true }).click();

    // Wait for page to load
    await expect(page.locator("h1")).toContainText("Monitors");

    // Should show healthy count
    await expect(page.getByText(/of \d+ healthy/)).toBeVisible();

    // Should have monitor cards with real data
    const monitorCards = page.locator(".glass-card");
    await expect(monitorCards.first()).toBeVisible();

    // Each monitor should show latency and uptime
    await expect(page.getByText("latency").first()).toBeVisible();
    await expect(page.getByText("uptime").first()).toBeVisible();
  });

  test("can create and delete a monitor", async ({ page }) => {
    // Navigate to Monitors page
    await page.getByRole("button", { name: "Monitors", exact: true }).click();
    await expect(page.locator("h1")).toContainText("Monitors");

    // Click Add monitor button
    await page.getByRole("button", { name: /add monitor/i }).click();

    // Fill in the form
    await page.getByPlaceholder("Production API").fill("E2E Test Monitor");
    await page.getByPlaceholder(/api\.example\.com/).fill("example.com");

    // Select ping type
    await page.locator('[role="combobox"]').first().click();
    await page.getByRole("option", { name: "Ping" }).click();

    // Submit
    await page.getByRole("button", { name: "Add" }).click();

    // Wait for dialog to close and monitor to appear
    await expect(page.getByText("E2E Test Monitor")).toBeVisible({ timeout: 10000 });

    // Delete the monitor we just created
    const monitorRow = page.locator(".glass-card", { hasText: "E2E Test Monitor" });
    await monitorRow.getByRole("button").last().click(); // Delete button

    // Verify it's removed
    await expect(page.getByText("E2E Test Monitor")).not.toBeVisible({ timeout: 10000 });
  });

  test("quick diagnostic works", async ({ page }) => {
    // Navigate to Monitors page
    await page.getByRole("button", { name: "Monitors", exact: true }).click();

    // Click Quick test button
    await page.getByRole("button", { name: /quick test/i }).click();

    // Fill in target
    await page.getByPlaceholder(/google\.com/).fill("8.8.8.8");

    // Click Run button
    await page.getByRole("button", { name: "Run" }).click();

    // Wait for result
    await expect(page.getByText(/Success|Failed/)).toBeVisible({ timeout: 20000 });

    // Should show output
    await expect(page.locator("pre")).toBeVisible();
  });

  test("alerts page loads", async ({ page }) => {
    // Navigate to Alerts page (exact match to avoid Quick Actions button)
    await page.getByRole("button", { name: "Alerts", exact: true }).click();

    // Wait for page to load
    await expect(page.locator("h1")).toContainText("Alerts");

    // Should show filter buttons
    await expect(page.getByRole("button", { name: "All" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Active" })).toBeVisible();
  });

  test("sidebar navigation works", async ({ page }) => {
    // Test navigation to each page (use exact match to avoid Quick Actions buttons)
    const pages = [
      { button: "Overview", title: "Overview" },
      { button: "Monitors", title: "Monitors" },
      { button: "Alerts", title: "Alerts" },
      { button: "DNS", title: "DNS" },
    ];

    for (const nav of pages) {
      await page.getByRole("button", { name: nav.button, exact: true }).click();
      await expect(page.locator("h1")).toContainText(nav.title);
    }
  });

  test("scheduler can be toggled", async ({ page }) => {
    // Navigate to Monitors page
    await page.getByRole("button", { name: "Monitors", exact: true }).click();
    await expect(page.locator("h1")).toContainText("Monitors");

    // The scheduler toggle is the second button in the header area
    const headerButtons = page.locator(".flex.gap-2 > button");

    // Should have refresh button (first) and scheduler toggle (second)
    const schedulerToggle = headerButtons.nth(1);
    await expect(schedulerToggle).toBeVisible();

    // Click to toggle scheduler
    await schedulerToggle.click();

    // Wait a moment for the state to update
    await page.waitForTimeout(2000);

    // Page should still be functional
    await expect(page.locator("h1")).toContainText("Monitors");
  });

  test("monitors refresh button works", async ({ page }) => {
    // Navigate to Monitors page
    await page.getByRole("button", { name: "Monitors", exact: true }).click();

    // Find refresh button (first button with rotate icon)
    const refreshButton = page.locator(".flex.gap-2 > button").first();
    await expect(refreshButton).toBeVisible();

    // Click refresh
    await refreshButton.click();

    // Should still show monitors
    await expect(page.locator("h1")).toContainText("Monitors");
  });

  test("dashboard shows SLO compliance", async ({ page }) => {
    // Dashboard should show SLO data
    await expect(page.getByText("SLO Compliance")).toBeVisible({ timeout: 10000 });

    // Should show monitor status grid
    await expect(page.getByText("Monitor Status")).toBeVisible();

    // Should show quick actions
    await expect(page.getByText("Quick Actions")).toBeVisible();
  });

  test("reports page can generate uptime report", async ({ page }) => {
    // Navigate to Reports page
    await page.getByRole("button", { name: "Reports", exact: true }).click();
    await expect(page.locator("h1")).toContainText("Reports");

    // Click Generate button
    await page.getByRole("button", { name: /generate/i }).click();

    // Wait for report to be generated
    await expect(page.getByText("Overall Uptime")).toBeVisible({ timeout: 10000 });

    // Export buttons should appear
    await expect(page.getByRole("button", { name: /export markdown/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /export json/i })).toBeVisible();
  });

  test("reports page can switch report types", async ({ page }) => {
    // Navigate to Reports page
    await page.getByRole("button", { name: "Reports", exact: true }).click();

    // Switch to Performance tab
    await page.getByRole("tab", { name: /performance/i }).click();
    await page.getByRole("button", { name: /generate/i }).click();
    await expect(page.getByText("Performance Metrics")).toBeVisible({ timeout: 10000 });

    // Switch to Incidents tab
    await page.getByRole("tab", { name: /incidents/i }).click();
    await page.getByRole("button", { name: /generate/i }).click();
    await expect(page.getByText(/Total Incidents/)).toBeVisible({ timeout: 10000 });
  });
});
