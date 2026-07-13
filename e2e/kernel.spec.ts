import { expect, test, type BrowserContext, type Page } from '@playwright/test';

async function join(context: BrowserContext, name: string): Promise<Page> {
  const page = await context.newPage();
  await page.goto('/');
  await page.getByLabel('Device name').fill(name);
  await page.getByRole('button', { name: 'Join' }).click();
  await expect(page.getByText(`You: ${name}`)).toBeVisible();
  return page;
}

async function selectFile(page: Page, recipient: string, name: string, contents: string): Promise<void> {
  const chooser = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: `Send files to ${recipient}` }).click();
  await (await chooser).setFiles({ name, mimeType: 'text/plain', buffer: Buffer.from(contents) });
}

test('kernel journey: presence, transfer, decline, and rejoin', async ({ browser }) => {
  const senderContext = await browser.newContext({ acceptDownloads: true });
  const receiverContext = await browser.newContext({ acceptDownloads: true });
  const sender = await join(senderContext, 'Kernel Sender');
  const receiver = await join(receiverContext, 'Kernel Receiver');

  await expect(sender.getByRole('button', { name: 'Send files to Kernel Receiver', exact: true })).toBeVisible();
  await expect(receiver.getByRole('button', { name: 'Send files to Kernel Sender', exact: true })).toBeVisible();

  await selectFile(sender, 'Kernel Receiver', 'hello.txt', 'befrest end-to-end payload');
  await expect(receiver.getByRole('dialog')).toContainText('Kernel Sender wants to send you');
  await expect(receiver.getByRole('dialog')).toContainText('hello.txt');
  const download = receiver.waitForEvent('download');
  await receiver.getByRole('button', { name: 'Accept' }).click();
  const stream = await (await download).createReadStream();
  const chunks: Buffer[] = [];
  for await (const chunk of stream!) chunks.push(Buffer.from(chunk));
  expect(Buffer.concat(chunks).toString()).toBe('befrest end-to-end payload');
  await expect(sender.getByText('Sent ✓')).toBeVisible();
  await expect(receiver.getByText('Saved to Downloads ✓')).toBeVisible();

  await selectFile(sender, 'Kernel Receiver', 'declined.txt', 'nothing should download');
  await expect(receiver.getByRole('dialog')).toContainText('declined.txt');
  await receiver.getByRole('button', { name: 'Decline' }).click();
  await expect(sender.getByText('declined.txt was declined')).toBeVisible();

  await receiver.close();
  await expect(sender.getByRole('button', { name: 'Send files to Kernel Receiver', exact: true })).toHaveCount(0);
  const rejoined = await receiverContext.newPage();
  await rejoined.goto('/');
  await expect(rejoined.getByLabel('Device name')).toHaveCount(0);
  await expect(rejoined.getByText('You: Kernel Receiver')).toBeVisible();
  await expect(sender.getByRole('button', { name: 'Send files to Kernel Receiver', exact: true })).toBeVisible();

  await senderContext.close();
  await receiverContext.close();
});
