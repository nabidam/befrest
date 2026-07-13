<script lang="ts">
  import type { Device } from '../lib/proto';
  import DeviceGrid from './DeviceGrid.svelte';
  import EmptyInvite from './EmptyInvite.svelte';
  import Footer from './Footer.svelte';
  import Header from './Header.svelte';
  import InviteSheet from './InviteSheet.svelte';
  import ReceiveBanner from './ReceiveBanner.svelte';
  import ConnectionBanner from './ConnectionBanner.svelte';
  import { connection } from '../lib/stores';

  export let devices: Device[] = [];
  export let loading = false;
  export let onRename: (name: string) => void;

  let inviteOpen = false;
</script>

<main aria-labelledby="main-title">
  <Header {onRename} />
  <ConnectionBanner />
  <ReceiveBanner />
  {#if loading}
    <section aria-label="Loading devices" aria-busy="true">
      <p>Loading devices…</p>
      <div aria-hidden="true">Loading device</div>
      <div aria-hidden="true">Loading device</div>
    </section>
  {:else if devices.length === 0}
    <EmptyInvite />
  {:else}
    <DeviceGrid {devices} disabled={$connection === 'connecting'} />
  {/if}
  <Footer onOpen={() => (inviteOpen = true)} />
</main>

<InviteSheet open={inviteOpen} onClose={() => (inviteOpen = false)} />
