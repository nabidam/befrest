function fileURL(transferID: string, index: number): string {
  return `/api/transfers/${encodeURIComponent(transferID)}/files/${index}`;
}

export function downloadFile(transferID: string, index: number): void {
  const link = document.createElement('a');
  link.href = fileURL(transferID, index);
  link.download = '';
  link.hidden = true;
  document.body.append(link);
  link.click();
  link.remove();
}
