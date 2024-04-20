import { expect, Page, Locator } from "@playwright/test";

export class ShareSecretPage {
  readonly page: Page;
  readonly url: string = "/";

  readonly headingLocator: Locator;

  readonly messageTextareaLocator: Locator;
  readonly passphraseInputLocator: Locator;
  readonly expireAmountLocator: Locator;
  readonly expireUnitLocator: Locator;
  readonly shareButtonLocator: Locator;
  readonly violationsLocator: Locator;

  readonly secretUrlLocator: Locator;
  readonly copyButtonLocator: Locator;

  constructor(page: Page) {
    this.page = page;

    this.headingLocator = page.getByRole("heading");

    this.messageTextareaLocator = page.getByLabel("Message", { exact: true });
    this.passphraseInputLocator = page.getByLabel("Passphrase", { exact: true });
    this.expireAmountLocator = page.locator('input[name="expire_amount"]');
    this.expireUnitLocator = page.getByRole("combobox");
    this.shareButtonLocator = page.getByRole("button", { name: "Share", exact: true });
    this.violationsLocator = page.getByRole("alert");

    this.secretUrlLocator = page.getByLabel("Secret URL", { exact: true });
    this.copyButtonLocator = page.getByRole("button", { name: "Copy", exact: true });
  }

  async visit() {
    await this.page.goto(this.url);

    await expect(this.headingLocator).toHaveText("Share secret");
    await expect(this.messageTextareaLocator).toBeVisible();
    await expect(this.passphraseInputLocator).toBeVisible();
    await expect(this.expireAmountLocator).toBeVisible();
    await expect(this.expireUnitLocator).toBeVisible();
    await expect(this.shareButtonLocator).toBeVisible();
    await expect(this.violationsLocator).toBeHidden();

    await expect(this.secretUrlLocator).toBeHidden();
    await expect(this.copyButtonLocator).toBeHidden();
  }

  async share(secret: { message: string; passphrase: string; amount?: string; unit?: string }) {
    await this.messageTextareaLocator.fill(secret.message);
    await this.passphraseInputLocator.fill(secret.passphrase);

    if (secret.amount !== undefined) {
      await this.expireAmountLocator.fill(secret.amount);
    }

    if (secret.unit !== undefined) {
      await this.expireUnitLocator.selectOption(secret.unit);
    }

    await this.shareButtonLocator.click();
  }

  async isShared() {
    await expect(this.headingLocator).toHaveText("Secret shared");

    await expect(this.messageTextareaLocator).toBeHidden();
    await expect(this.passphraseInputLocator).toBeHidden();
    await expect(this.expireAmountLocator).toBeHidden();
    await expect(this.expireUnitLocator).toBeHidden();
    await expect(this.shareButtonLocator).toBeHidden();
    await expect(this.violationsLocator).toBeHidden();

    await expect(this.secretUrlLocator).toBeVisible();
    await expect(this.copyButtonLocator).toBeVisible();

    await this.copyButtonLocator.click();
    const secretUrl: string = await this.page.evaluate("navigator.clipboard.readText()");
    await expect(this.secretUrlLocator).toHaveValue(secretUrl);
  }

  async hasViolation(violation: string) {
    await expect(this.headingLocator).toHaveText("Share secret");

    await expect(this.messageTextareaLocator).toBeVisible();
    await expect(this.passphraseInputLocator).toBeVisible();
    await expect(this.expireAmountLocator).toBeVisible();
    await expect(this.expireUnitLocator).toBeVisible();
    await expect(this.shareButtonLocator).toBeVisible();
    await expect(this.violationsLocator).toBeVisible();

    await expect(this.secretUrlLocator).toBeHidden();
    await expect(this.copyButtonLocator).toBeHidden();

    await expect(this.violationsLocator).toHaveText(violation);
  }

  async getUrl(): Promise<string> {
    return await this.secretUrlLocator.inputValue();
  }
}

export class OpenSecretPage {
  readonly page: Page;
  readonly url: string;

  readonly headingLocator: Locator;

  readonly passphraseInputLocator: Locator;
  readonly openButtonLocator: Locator;
  readonly violationsLocator: Locator;

  readonly messageTextareaLocator: Locator;
  readonly copyButtonLocator: Locator;

  constructor(page: Page, url: string) {
    this.page = page;
    this.url = url;

    this.headingLocator = page.getByRole("heading");

    this.passphraseInputLocator = page.getByLabel("Passphrase", { exact: true });
    this.openButtonLocator = page.getByRole("button", { name: "Open", exact: true });
    this.violationsLocator = page.getByRole("alert");

    this.messageTextareaLocator = page.getByLabel("Message", { exact: true });
    this.copyButtonLocator = page.getByRole("button", { name: "Copy", exact: true });
  }

  async visit() {
    await this.page.goto(this.url);

    await expect(this.headingLocator).toHaveText("Open secret");

    await expect(this.passphraseInputLocator).toBeVisible();
    await expect(this.openButtonLocator).toBeVisible();
    await expect(this.violationsLocator).toBeHidden();

    await expect(this.messageTextareaLocator).toBeHidden();
    await expect(this.copyButtonLocator).toBeHidden();
  }

  async open(passphrase: string) {
    await this.passphraseInputLocator.fill(passphrase);
    await this.openButtonLocator.click();
  }

  async isOpened() {
    await expect(this.headingLocator).toHaveText("Secret opened");

    await expect(this.passphraseInputLocator).toBeHidden();
    await expect(this.openButtonLocator).toBeHidden();
    await expect(this.violationsLocator).toBeHidden();

    await expect(this.messageTextareaLocator).toBeVisible();
    await expect(this.copyButtonLocator).toBeVisible();

    await this.copyButtonLocator.click();
    const secretValue: string = await this.page.evaluate("navigator.clipboard.readText()");
    await expect(this.messageTextareaLocator).toHaveValue(secretValue);
  }

  async hasMessage(message: string) {
    await expect(this.messageTextareaLocator).toHaveText(message);
  }

  async hasViolation(violation: string) {
    await expect(this.headingLocator).toHaveText("Open secret");

    await expect(this.passphraseInputLocator).toBeVisible();
    await expect(this.openButtonLocator).toBeVisible();
    await expect(this.violationsLocator).toBeVisible();

    await expect(this.violationsLocator).toHaveText(violation);
  }
}
