export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const unit = Math.min(Math.floor(Math.log(bytes) / Math.log(1_000)), units.length - 1);
  const value = bytes / 1_000 ** unit;
  return `${value.toFixed(1)} ${units[unit]}`;
}

export function formatAdditionalFiles(count: number, totalSize: number): string {
  return `and ${count} more — ${formatBytes(totalSize)} total`;
}

export function formatFilePosition(index: number, count: number): string {
  return `file ${index + 1} of ${count}`;
}

export type TransferFailureReason =
  | 'sender-disconnected'
  | 'receiver-disconnected'
  | 'cancelled-by-sender'
  | 'cancelled-by-receiver'
  | 'stream-error';

export type OfferCancellationReason = 'sender-cancelled' | 'sender-disconnected';

const transferFailureCopy: Record<TransferFailureReason, (counterpart: string) => string> = {
  'sender-disconnected': (counterpart) => `Transfer failed — ${counterpart} disconnected`,
  'receiver-disconnected': (counterpart) => `Transfer failed — ${counterpart} disconnected`,
  'cancelled-by-sender': (counterpart) => `${counterpart} cancelled the transfer`,
  'cancelled-by-receiver': (counterpart) => `${counterpart} cancelled the transfer`,
  'stream-error': () => 'Transfer failed',
};

const offerCancellationCopy: Record<OfferCancellationReason, (sender: string) => string> = {
  'sender-cancelled': (sender) => `${sender} cancelled`,
  'sender-disconnected': (sender) => `${sender} cancelled`,
};

export function formatTransferFailure(reason: TransferFailureReason, counterpart = 'The other device'): string {
  return transferFailureCopy[reason](counterpart);
}

export function formatOfferCancellation(reason: OfferCancellationReason, sender = 'The sender'): string {
  return offerCancellationCopy[reason](sender);
}
