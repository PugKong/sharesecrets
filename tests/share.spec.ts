import { test } from "@playwright/test";
import { ShareSecretPage, OpenSecretPage } from "./page";

const openSecretViolation = "Message not found or invalid passphrase";

const secret = { message: "my message", passphrase: "my passphrase" };

test("it stores secrets", async ({ page }) => {
  const shareSecretPage = new ShareSecretPage(page);
  await shareSecretPage.visit();
  await shareSecretPage.share(secret);
  await shareSecretPage.isShared();
  const secretUrl = await shareSecretPage.getUrl();

  const openSecretPage = new OpenSecretPage(page, secretUrl);
  await openSecretPage.visit();
  await openSecretPage.open(secret.passphrase);
  await openSecretPage.isOpened();
  await openSecretPage.hasMessage(secret.message);
});

test("it checks passphrase", async ({ page }) => {
  const shareSecretPage = new ShareSecretPage(page);
  await shareSecretPage.visit();
  await shareSecretPage.share(secret);
  await shareSecretPage.isShared();
  const secretUrl = await shareSecretPage.getUrl();

  const openSecretPage = new OpenSecretPage(page, secretUrl);
  await openSecretPage.visit();
  await openSecretPage.open(secret.passphrase + "!");
  await openSecretPage.hasViolation(openSecretViolation);
});

test("it allows to change secret lifetime", async ({ page }) => {
  const shareSecretPage = new ShareSecretPage(page);
  await shareSecretPage.visit();
  await shareSecretPage.share({ ...secret, amount: "1", unit: "seconds" });
  await shareSecretPage.isShared();
  const secretUrl = await shareSecretPage.getUrl();

  await new Promise((fn) => setTimeout(fn, 2000));

  const openSecretPage = new OpenSecretPage(page, secretUrl);
  await openSecretPage.visit();
  await openSecretPage.open(secret.passphrase);
  await openSecretPage.hasViolation(openSecretViolation);
});

test("it validates passphrase max length", async ({ page }) => {
  const maxLen = 32;

  const shareSecretPage = new ShareSecretPage(page);
  await shareSecretPage.visit();
  await shareSecretPage.share({ ...secret, passphrase: "x".repeat(maxLen + 1) });
  await shareSecretPage.hasViolation("The passphrase must be less than or equal to 32 bytes");

  await shareSecretPage.share({ ...secret, passphrase: "x".repeat(maxLen) });
  await shareSecretPage.isShared();
});

test("it validates message max length", async ({ page }) => {
  const maxLen = 4 * 1024;

  const shareSecretPage = new ShareSecretPage(page);
  await shareSecretPage.visit();
  await shareSecretPage.share({ ...secret, message: "x".repeat(maxLen + 1) });
  await shareSecretPage.hasViolation("The message must be less than or equal to 4 kilobytes");

  await shareSecretPage.share({ ...secret, message: "x".repeat(maxLen) });
  await shareSecretPage.isShared();
});

test("it validates lifetime", async ({ page }) => {
  const shareSecretPage = new ShareSecretPage(page);
  await shareSecretPage.visit();
  await shareSecretPage.share({ ...secret, amount: "-1" });
  await shareSecretPage.hasViolation("The expire field must be positive");

  await shareSecretPage.share({ ...secret, amount: "1441", unit: "minutes" });
  await shareSecretPage.hasViolation("Expire must be less than 1 day");

  await shareSecretPage.share({ ...secret, amount: "1440", unit: "minutes" });
  await shareSecretPage.isShared();
});
