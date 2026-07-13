<script lang="ts">
  import type { Device } from '../lib/proto';
  import DeviceGrid from './DeviceGrid.svelte';
  import EmptyInvite from './EmptyInvite.svelte';
  import Footer from './Footer.svelte';
  import Header from './Header.svelte';
  import InviteSheet from './InviteSheet.svelte';
  import ReceiveBanner from './ReceiveBanner.svelte';

  export let devices: Device[] = [];

  let inviteOpen = false;
</script>

<main aria-labelledby="main-title">
  <Header />
  <ReceiveBanner />
  {#if devices.length === 0}
    <EmptyInvite />
  {:else}
    <DeviceGrid {devices} />
  {/if}
  <Footer onOpen={() => (inviteOpen = true)} />
</main>

<InviteSheet open={inviteOpen} onClose={() => (inviteOpen = false)} />
