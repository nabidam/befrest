import { describe, expect, it } from 'vitest';
import { formatBytes, formatOfferCancellation, formatTransferFailure } from './format';

describe('formatBytes', () => {
  it.each([
    [0, '0 B'],
    [1_300_000_000, '1.3 GB'],
    [2_000_000_000, '2.0 GB'],
  ])('formats %d bytes as %s', (bytes, expected) => {
    expect(formatBytes(bytes)).toBe(expected);
  });
});

describe('transfer failure copy', () => {
  it.each([
    ['sender-disconnected', 'Transfer failed — Pixel 8 disconnected'],
    ['receiver-disconnected', 'Transfer failed — Pixel 8 disconnected'],
    ['cancelled-by-sender', 'Pixel 8 cancelled the transfer'],
    ['cancelled-by-receiver', 'Pixel 8 cancelled the transfer'],
    ['stream-error', 'Transfer failed'],
  ] as const)('maps %s to the approved copy', (reason, expected) => {
    expect(formatTransferFailure(reason, 'Pixel 8')).toBe(expected);
  });

  it.each([
    ['sender-cancelled', 'Pixel 8 cancelled'],
    ['sender-disconnected', 'Pixel 8 cancelled'],
  ] as const)('maps an offer cancellation %s to the approved copy', (reason, expected) => {
    expect(formatOfferCancellation(reason, 'Pixel 8')).toBe(expected);
  });
});
