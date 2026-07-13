<script lang="ts">
  import { transfers } from '../lib/stores';
  import { formatBytes, formatFilePosition } from '../lib/format';

  export let onCancel: (transferID: string) => void;

  $: transfer = Object.values($transfers).find((item) => item.direction === 'receiving');
</script>

{#if transfer}
  <aside aria-live="polite" aria-atomic="true">
    <p class="title">Receiving {transfer.transfer.files[transfer.index]?.name ?? 'file'}</p>
    <p class="position">{formatFilePosition(transfer.index, transfer.transfer.files.length)}</p>
    <progress value={transfer.totalSent} max={transfer.totalSize}></progress>
    <p class="amount">{formatBytes(transfer.totalSent)} / {formatBytes(transfer.totalSize)}</p>
    <button type="button" on:click={() => onCancel(transfer.transfer.id)}>Cancel</button>
  </aside>
{/if}

<style>
  aside { display: grid; grid-template-columns: 1fr auto; align-items: center; gap: var(--space-sm) var(--space-md); border-block-end: thin solid var(--color-border); background: var(--color-surface-raised); margin-inline: calc(var(--space-md) * -1); padding: var(--space-sm) var(--space-md); }
  p { margin: 0; }
  .title { font-size: var(--text-body-sm); font-weight: var(--weight-semibold); line-height: var(--leading-body-sm); }
  .position, .amount { color: var(--color-text-muted); font-family: var(--font-mono); font-size: var(--text-label-mono); line-height: var(--leading-label-mono); }
  .position { grid-column: 1; }
  progress { grid-column: 1; inline-size: 100%; block-size: var(--size-bar); accent-color: var(--color-accent); }
  button { grid-column: 2; grid-row: 1 / span 3; min-block-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-weight: var(--weight-semibold); padding-inline: var(--space-md); }
  button:hover { background: var(--color-surface-hover); }
</style>
