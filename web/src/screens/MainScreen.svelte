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
  export let onCancel: (transferID: string, pending: boolean) => void;
  export let onChangeNetwork: () => void;

  let inviteOpen = false;
</script>

<main aria-labelledby="main-title">
  <Header {onRename} />
  <ConnectionBanner />
  <ReceiveBanner onCancel={(transferID) => onCancel(transferID, false)} />
  {#if loading}
    <section aria-label="Loading devices" aria-busy="true">
      <p class="loading-label">Loading devices…</p>
      <div class="skeleton" aria-hidden="true">Loading device</div>
      <div class="skeleton" aria-hidden="true">Loading device</div>
    </section>
  {:else if devices.length === 0}
    <EmptyInvite />
  {:else}
    <DeviceGrid {devices} disabled={$connection === 'connecting'} {onCancel} />
  {/if}
  <Footer onOpen={() => (inviteOpen = true)} />
</main>

<InviteSheet open={inviteOpen} onClose={() => (inviteOpen = false)} {onChangeNetwork} />

<style>
  main { display: grid; grid-template-rows: auto auto auto 1fr auto; min-block-size: 100dvh; max-inline-size: var(--size-content-max); margin-inline: auto; padding: var(--space-md); }
  section[aria-label='Loading devices'] { display: grid; grid-template-columns: repeat(auto-fill, minmax(var(--size-card-col), 1fr)); gap: var(--space-md); align-content: start; margin-block: var(--space-xl); }
  .loading-label { grid-column: 1 / -1; margin: 0; color: var(--color-text-muted); font-size: var(--text-caption); font-weight: var(--weight-medium); letter-spacing: var(--tracking-caption); text-transform: uppercase; }
  .skeleton { min-block-size: var(--size-card-min); overflow: hidden; border: thin solid var(--color-border); border-radius: var(--radius-md); background: linear-gradient(90deg, var(--color-surface-sunken), var(--color-surface), var(--color-surface-sunken)); background-size: 200% 100%; color: transparent; animation: shimmer var(--dur-slow) var(--ease-standard) infinite; }
  @keyframes shimmer { to { background-position: -200% 0; } }
  @media (min-width: 40rem) { main { padding: var(--space-margin); } }
</style>
