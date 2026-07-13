<script lang="ts">
  import { transfers } from '../lib/stores';
  import { formatBytes } from '../lib/format';

  $: transfer = Object.values($transfers).find((item) => item.direction === 'receiving');
</script>

{#if transfer}
  <aside aria-live="polite">
    <p>Receiving {transfer.transfer.files[transfer.index]?.name ?? 'file'}</p>
    <progress value={transfer.totalSent} max={transfer.totalSize}></progress>
    <p>{formatBytes(transfer.totalSent)} / {formatBytes(transfer.totalSize)}</p>
  </aside>
{/if}
