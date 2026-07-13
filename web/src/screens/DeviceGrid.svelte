<script lang="ts">
  import type { Device } from '../lib/proto';
  import DeviceCard from './DeviceCard.svelte';

  export let devices: Device[] = [];
  export let disabled = false;
  export let onCancel: (transferID: string, pending: boolean) => void;
</script>

<section aria-labelledby="send-to-title">
  <h2 id="send-to-title">Send to</h2>
  <div role="list">
    {#each devices as device (device.id)}
      <div role="listitem"><DeviceCard {device} {disabled} {onCancel} /></div>
    {/each}
  </div>
</section>

<style>
  section { margin-block: var(--space-xl); }
  h2 { margin: 0 0 var(--space-md); color: var(--color-text-muted); font-size: var(--text-caption); font-weight: var(--weight-medium); letter-spacing: var(--tracking-caption); line-height: var(--leading-caption); text-transform: uppercase; }
  [role='list'] { display: grid; grid-template-columns: repeat(auto-fill, minmax(var(--size-card-col), 1fr)); gap: var(--space-md); }
</style>
