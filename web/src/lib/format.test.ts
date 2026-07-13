import { describe, expect, it } from 'vitest';
import { formatBytes } from './format';

describe('formatBytes', () => {
  it.each([
    [0, '0 B'],
    [1_300_000_000, '1.3 GB'],
    [2_000_000_000, '2.0 GB'],
  ])('formats %d bytes as %s', (bytes, expected) => {
    expect(formatBytes(bytes)).toBe(expected);
  });
});
