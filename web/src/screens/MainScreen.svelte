<script lang="ts">
  import type { Device } from '../lib/proto';
  import DeviceGrid from './DeviceGrid.svelte';

  export let devices: Device[] = [];
</script>

<main aria-labelledby="main-title">
  <h1 id="main-title">befrest</h1>
  <DeviceGrid {devices} />
</main>
