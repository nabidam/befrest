import type { Transfer } from './proto';
import { offerFiles } from './ws';

const filesByRecipient = new Map<string, File[]>();
const filesByTransfer = new Map<string, File[]>();

function fileURL(transferID: string, index: number): string {
  return `/api/transfers/${encodeURIComponent(transferID)}/files/${index}`;
}

export function chooseFiles(recipientID: string): void {
  const picker = document.createElement('input');
  picker.type = 'file';
  picker.multiple = true;
  picker.addEventListener('change', () => {
    const files = Array.from(picker.files ?? []);
    if (files.length === 0) return;
    offerSelectedFiles(recipientID, files);
  });
  picker.click();
}

export function offerSelectedFiles(recipientID: string, files: File[]): boolean {
  if (files.length === 0) return false;
  filesByRecipient.set(recipientID, files);
  if (offerFiles(recipientID, files)) return true;
  filesByRecipient.delete(recipientID);
  return false;
}

export function registerOfferFiles(transfer: Transfer): void {
  const files = filesByRecipient.get(transfer.receiverId);
  if (!files) return;
  filesByRecipient.delete(transfer.receiverId);
  filesByTransfer.set(transfer.id, files);
}

export async function beginUpload(transferID: string): Promise<void> {
  const files = filesByTransfer.get(transferID);
  if (!files) return;

  try {
    for (const [index, file] of files.entries()) {
      const response = await fetch(fileURL(transferID, index), { method: 'POST', body: file });
      if (!response.ok) return;
    }
  } catch {
    // The WebSocket transfer verdict is the user-visible source of truth.
  } finally {
    filesByTransfer.delete(transferID);
  }
}
