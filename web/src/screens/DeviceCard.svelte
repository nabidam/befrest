<script lang="ts">
  import type { Device } from '../lib/proto';
  import { chooseFiles, offerSelectedFiles } from '../lib/upload';
  import { self, transfers } from '../lib/stores';
  import { formatBytes, formatFilePosition } from '../lib/format';

  export let device: Device;
  export let disabled = false;
  export let onCancel: (transferID: string, pending: boolean) => void;

  let dropping = false;

  $: transfer = Object.values($transfers).find(
    (item) => item.direction === 'sending' && item.transfer.receiverId === device.id,
  );

  function pickFiles(): void {
    if (!transfer) chooseFiles(device.id);
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      pickFiles();
    }
  }

  function cancel(): void {
    if (!transfer) return;
    onCancel(transfer.transfer.id, transfer.transfer.state === 'offered');
  }

  function onDragover(event: DragEvent): void {
    if (disabled || transfer || !event.dataTransfer?.types.includes('Files')) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
    dropping = true;
  }

  function onDragleave(event: DragEvent): void {
    if (!event.currentTarget.contains(event.relatedTarget as Node | null)) dropping = false;
  }

  function onDrop(event: DragEvent): void {
    event.preventDefault();
    dropping = false;
    if (disabled || transfer) return;
    offerSelectedFiles(device.id, Array.from(event.dataTransfer?.files ?? []));
  }
</script>

<div
  role="button"
  tabindex={disabled || Boolean(transfer) ? undefined : 0}
  aria-label={`Send files to ${device.name}`}
  aria-disabled={disabled || Boolean(transfer)}
  data-drop-active={dropping}
  on:click={pickFiles}
  on:keydown={onKeydown}
  on:dragover={onDragover}
  on:dragleave={onDragleave}
  on:drop={onDrop}
>
  {#if transfer}
    <h3>{device.name}</h3>
    <p>{transfer.transfer.files[transfer.index]?.name ?? 'Sending file'}</p>
    <p>{formatFilePosition(transfer.index, transfer.transfer.files.length)}</p>
    <progress value={transfer.totalSent} max={transfer.totalSize}></progress>
    <p>{formatBytes(transfer.totalSent)} / {formatBytes(transfer.totalSize)}</p>
    <p>{transfer.totalSize === 0 ? 100 : Math.round((transfer.totalSent / transfer.totalSize) * 100)}%</p>
    {#if $self?.kind === 'mobile'}
      <p>Keep this screen on until sending finishes.</p>
    {/if}
    <button type="button" on:click|stopPropagation={cancel}>Cancel</button>
  {:else if dropping}
    <p>Drop to send to {device.name}</p>
  {:else}
    <p aria-hidden="true">{device.kind === 'mobile' ? '📱' : '💻'}</p>
    <h3>{device.name}</h3>
  {/if}
</div>
