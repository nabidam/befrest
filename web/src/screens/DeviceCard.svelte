<script lang="ts">
  import type { Device } from '../lib/proto';

  export let device: Device;
</script>

<button type="button" aria-label={`Send files to ${device.name}`}>
  <p aria-hidden="true">{device.kind === 'mobile' ? '📱' : '💻'}</p>
  <h3>{device.name}</h3>
</button>
