<script lang="ts">
  import type { Device } from '../lib/proto';
  import { chooseFiles } from '../lib/upload';
  import { transfers } from '../lib/stores';
  import { formatBytes } from '../lib/format';

  export let device: Device;

  $: transfer = Object.values($transfers).find(
    (item) => item.direction === 'sending' && item.transfer.receiverId === device.id,
  );

  function pickFiles(): void {
    if (!transfer) chooseFiles(device.id);
  }
</script>

<button type="button" aria-label={`Send files to ${device.name}`} on:click={pickFiles} disabled={Boolean(transfer)}>
  {#if transfer}
    <h3>{device.name}</h3>
    <p>{transfer.transfer.files[transfer.index]?.name ?? 'Sending file'}</p>
    <progress value={transfer.totalSent} max={transfer.totalSize}></progress>
    <p>{formatBytes(transfer.totalSent)} / {formatBytes(transfer.totalSize)}</p>
    <p>{transfer.totalSize === 0 ? 100 : Math.round((transfer.totalSent / transfer.totalSize) * 100)}%</p>
  {:else}
    <p aria-hidden="true">{device.kind === 'mobile' ? '📱' : '💻'}</p>
    <h3>{device.name}</h3>
  {/if}
</button>
