<script lang="ts">
  import { transfers } from '../lib/stores';
  import { formatBytes, formatFilePosition } from '../lib/format';

  export let onCancel: (transferID: string) => void;

  $: transfer = Object.values($transfers).find((item) => item.direction === 'receiving');
</script>

{#if transfer}
  <aside aria-live="polite">
    <p>Receiving {transfer.transfer.files[transfer.index]?.name ?? 'file'}</p>
    <p>{formatFilePosition(transfer.index, transfer.transfer.files.length)}</p>
    <progress value={transfer.totalSent} max={transfer.totalSize}></progress>
    <p>{formatBytes(transfer.totalSent)} / {formatBytes(transfer.totalSize)}</p>
    <button type="button" on:click={() => onCancel(transfer.transfer.id)}>Cancel</button>
  </aside>
{/if}
