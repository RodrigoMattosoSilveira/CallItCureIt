import { expect, test } from "@playwright/test";

test.describe("full scenario training flow", () => {
  test("trainee completes the hearsay objection flow and opens debrief", async ({
    page,
  }) => {
    await page.goto("/scenarios");

    await expect(
      page.getByRole("heading", { name: /objection training scenarios/i })
    ).toBeVisible();

    await expect(
      page.getByText(/basic hearsay on direct examination/i)
    ).toBeVisible();

    await page
      .getByRole("link", { name: /view scenario/i })
      .first()
      .click();

    await expect(
      page.getByRole("heading", {
        name: /basic hearsay on direct examination/i,
      })
    ).toBeVisible();

    await expect(page.getByText(/transcript preview/i)).toBeVisible();

    await page
      .getByRole("button", { name: /start training session/i })
      .click();

    await expect(
      page.getByRole("heading", { name: /training session/i })
    ).toBeVisible();

    await expect(page).toHaveURL(/\/sessions\/.+\/play/);

    await page.getByRole("button", { name: /next line/i }).click();

    await expect(
      page.getByText(/where were you on the evening of march 12/i)
    ).toBeVisible();

    await page.getByRole("button", { name: /next line/i }).click();

    await expect(page.getByText(/front porch/i)).toBeVisible();

    await page.getByRole("button", { name: /next line/i }).click();

    await expect(
      page.getByText(/did you speak with your neighbor/i)
    ).toBeVisible();

    await page.getByRole("button", { name: /next line/i }).click();

    await expect(
      page.getByText(/defendant admitted he caused the accident/i)
    ).toBeVisible();

    await page
      .getByLabel(/type your objection/i)
      .fill("Objection, hearsay.");

    await page
      .getByRole("button", { name: /submit objection/i })
      .click();

    await expect(page.getByText(/objection, hearsay/i)).toBeVisible();

    await expect(page.getByText(/sustained/i)).toBeVisible();

    await expect(page.getByText(/coach feedback/i)).toBeVisible();

    await expect(page.getByText(/correct/i)).toBeVisible();

    await expect(page.getByText(/current score/i)).toBeVisible();

    await expect(page.getByText(/overall/i)).toBeVisible();

    await page.getByRole("link", { name: /view debrief/i }).click();

    await expect(
      page.getByRole("heading", { name: /session debrief/i })
    ).toBeVisible();

    await expect(page).toHaveURL(/\/sessions\/.+\/debrief/);

    await expect(page.getByText(/summary/i)).toBeVisible();

    await expect(page.getByText(/action review/i)).toBeVisible();

    await expect(page.getByText(/full transcript/i)).toBeVisible();

    await expect(page.getByText(/objection, hearsay/i)).toBeVisible();

    await expect(page.getByText(/sustained/i)).toBeVisible();

    await expect(page.getByText(/overall/i)).toBeVisible();

    await expect(page.getByText(/legal accuracy/i)).toBeVisible();
  });
});