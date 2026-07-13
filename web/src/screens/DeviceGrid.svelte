<script lang="ts">
  import type { Device } from '../lib/proto';
  import DeviceCard from './DeviceCard.svelte';

  export let devices: Device[] = [];
</script>

<section aria-labelledby="send-to-title">
  <h2 id="send-to-title">Send to</h2>
  <div role="list">
    {#each devices as device (device.id)}
      <div role="listitem"><DeviceCard {device} /></div>
    {/each}
  </div>
</section>
