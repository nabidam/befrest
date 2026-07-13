import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';

async function join(context: BrowserContext, name: string): Promise<Page> {
  const page = await context.newPage();
  await page.goto('/');
  await page.getByLabel('Device name').fill(name);
  await page.getByRole('button', { name: 'Join' }).click();
  await expect(page.getByText(`You: ${name}`)).toBeVisible();
  return page;
}

async function offerFile(page: Page, recipient: string, name: string, size = 1_024): Promise<void> {
  const chooser = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: `Send files to ${recipient}` }).click();
  await (await chooser).setFiles({ name, mimeType: 'application/octet-stream', buffer: Buffer.alloc(size, 7) });
}

async function pairedPages(browser: Browser, senderName: string, receiverName: string) {
  const senderContext = await browser.newContext({ acceptDownloads: true });
  const receiverContext = await browser.newContext({ acceptDownloads: true });
  const sender = await join(senderContext, senderName);
  const receiver = await join(receiverContext, receiverName);
  return { senderContext, receiverContext, sender, receiver };
}

test('offer cancellation closes the receiver prompt', async ({ browser }) => {
  const senderName = 'Offer Sender';
  const receiverName = 'Offer Receiver';
  const { senderContext, receiverContext, sender, receiver } = await pairedPages(browser, senderName, receiverName);

  await offerFile(sender, receiverName, 'withdraw.bin');
  await expect(receiver.getByRole('dialog')).toContainText('withdraw.bin');
  await sender.getByRole('button', { name: 'Cancel' }).click();
  await expect(receiver.getByRole('dialog')).toHaveCount(0);
  await expect(receiver.getByText(`${senderName} cancelled`)).toBeVisible();

  await senderContext.close();
  await receiverContext.close();
});

test('mid-transfer cancellation renders cancelled copy', async ({ browser }) => {
  const senderName = 'Cancel Sender';
  const receiverName = 'Cancel Receiver';
  const { senderContext, receiverContext, sender, receiver } = await pairedPages(browser, senderName, receiverName);
  await offerFile(sender, receiverName, 'cancel.bin', 2 * 1024 * 1024);
  await receiver.getByRole('button', { name: 'Accept' }).click();
  await expect(receiver.getByRole('complementary').getByRole('button', { name: 'Cancel' })).toBeVisible();
  await receiver.getByRole('complementary').getByRole('button', { name: 'Cancel' }).click();
  await expect(sender.getByText(`${receiverName} cancelled the transfer`)).toBeVisible();

  await senderContext.close();
  await receiverContext.close();
});

test('closing a receiver context yields a disconnect verdict', async ({ browser }) => {
  const senderName = 'Disconnect Sender';
  const receiverName = 'Disconnect Receiver';
  const { senderContext, receiverContext, sender, receiver } = await pairedPages(browser, senderName, receiverName);
  await offerFile(sender, receiverName, 'disconnect.bin', 2 * 1024 * 1024);
  await receiver.getByRole('button', { name: 'Accept' }).click();
  await expect(receiver.getByRole('complementary')).toBeVisible();
  await receiverContext.close();
  await expect(sender.getByText(`Transfer failed — ${receiverName} disconnected`)).toBeVisible();

  await senderContext.close();
});
