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

<style>
  [role='button'] { position: relative; display: grid; align-content: center; gap: var(--space-xs); min-block-size: var(--size-card-min); overflow: hidden; border: thin solid var(--color-border); border-radius: var(--radius-md); background: var(--color-surface); color: var(--color-text); cursor: pointer; padding: var(--space-md); transition: background var(--dur-fast) var(--ease-standard), border-color var(--dur-fast) var(--ease-standard), opacity var(--dur-fast) var(--ease-standard), transform var(--dur-fast) var(--ease-standard); }
  [role='button']::after { position: absolute; inset: 0; border: var(--focus-ring) solid var(--color-accent); border-radius: var(--radius-full); content: ''; opacity: 0; pointer-events: none; animation: join-pulse var(--dur-pulse) var(--ease-decel) 1; }
  [role='button']:hover:not([aria-disabled='true']) { border-color: var(--color-accent); transform: translateY(var(--lift-hover)); }
  [role='button']:active:not([aria-disabled='true']) { transform: scale(var(--press-scale)); }
  [role='button'][aria-disabled='true'] { cursor: not-allowed; opacity: var(--opacity-disabled); }
  [role='button'][aria-disabled='true']:hover { transform: none; }
  [role='button'][data-drop-active='true'] { border-color: var(--color-accent); border-style: dashed; background: var(--color-surface-hover); }
  h3, p { margin: 0; }
  h3 { font-size: var(--text-headline); font-weight: var(--weight-semibold); line-height: var(--leading-headline); }
  p { color: var(--color-text-muted); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); }
  p[aria-hidden='true'] { color: var(--color-icon); font-size: var(--text-display); line-height: var(--leading-display); }
  progress { inline-size: 100%; block-size: var(--size-bar); accent-color: var(--color-accent); }
  progress::-webkit-progress-bar { border-radius: var(--radius-full); background: var(--color-surface-sunken); }
  progress::-webkit-progress-value { border-radius: var(--radius-full); background: var(--color-accent); transition: inline-size var(--dur-base) var(--ease-standard); }
  progress::-moz-progress-bar { border-radius: var(--radius-full); background: var(--color-accent); }
  button { justify-self: start; min-block-size: var(--size-touch); border: thin solid transparent; background: transparent; color: var(--color-danger); cursor: pointer; font-weight: var(--weight-semibold); padding-inline: var(--space-xs); }
  button:hover { text-decoration: underline; }
  @keyframes join-pulse { 0% { inset: var(--space-md); opacity: 0; } 20% { opacity: 1; } 100% { inset: calc(var(--space-md) * -1); opacity: 0; } }
  @media (prefers-reduced-motion: reduce) { [role='button']::after { animation: none; } }
</style>
