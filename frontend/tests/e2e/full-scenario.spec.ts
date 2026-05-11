import { expect, test } from "@playwright/test";

test.describe("full scenario training flow", () => {
  test("trainee completes the hearsay objection flow and opens debrief", async ({
    page,
  }) => {
    await page.goto("/scenarios");

    await expect(
      page.getByRole("heading", { name: /objection training scenarios/i })
    ).toBeVisible();

	const hearsayCard = page
	.locator(".card")
	.filter({ hasText: /basic hearsay on direct examination/i });

	await expect(hearsayCard).toBeVisible();

	await hearsayCard.getByRole("link", { name: /view scenario/i }).click();

	await expect(page).toHaveURL(/\/scenarios\/scenario-hearsay-001/);

	await expect(
	page.getByRole("heading", {
		name: /basic hearsay on direct examination/i,
	})
	).toBeVisible();

    await page
      .getByRole("button", { name: /start training session/i })
      .click();

    await expect(
      page.getByRole("heading", { name: /training session/i })
    ).toBeVisible();

    for (const expectedText of [
      /where were you on the evening of march 12/i,
      /front porch/i,
      /did you speak with your neighbor/i,
      /defendant admitted he caused the accident/i,
    ]) {
      await page.getByRole("button", { name: /next line/i }).click();
      await expect(page.getByText(expectedText).first()).toBeVisible();
    }

    await page
      .getByLabel(/type your objection/i)
      .fill("Objection, hearsay.");

    await page
      .getByRole("button", { name: /submit objection/i })
      .click();

   	await expect(page.getByText(/objection, hearsay/i).first()).toBeVisible();
	await expect(page.getByText(/sustained/i).first()).toBeVisible();
	await expect(page.getByText(/coach feedback/i).first()).toBeVisible();
	await expect(page.getByText(/correct/i).first()).toBeVisible();

	await page.getByRole("link", { name: /view debrief/i }).click();
    await expect(
      page.getByRole("heading", { name: /session debrief/i })
    ).toBeVisible();

    await expect(page.getByText(/summary/i).first()).toBeVisible();
    await expect(page.getByText(/action review/i).first()).toBeVisible();
    await expect(page.getByText(/full transcript/i).first()).toBeVisible();
    await expect(page.getByText(/objection, hearsay/i).first()).toBeVisible();
  });
});