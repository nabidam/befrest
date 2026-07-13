<script lang="ts">
  import type { Device } from '../lib/proto';
  import DeviceGrid from './DeviceGrid.svelte';
  import ReceiveBanner from './ReceiveBanner.svelte';

  export let devices: Device[] = [];
</script>

<main aria-labelledby="main-title">
  <h1 id="main-title">befrest</h1>
  <ReceiveBanner />
  <DeviceGrid {devices} />
</main>
