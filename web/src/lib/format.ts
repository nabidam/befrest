export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const unit = Math.min(Math.floor(Math.log(bytes) / Math.log(1_000)), units.length - 1);
  const value = bytes / 1_000 ** unit;
  return `${value.toFixed(1)} ${units[unit]}`;
}
