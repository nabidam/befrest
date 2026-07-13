<script lang="ts">
  import { onMount } from 'svelte';
  import { connection, connectionError, devices, needsName, offers, self } from './lib/stores';
  import { connect, respondToOffer, setName } from './lib/ws';
  import MainScreen from './screens/MainScreen.svelte';
  import NameScreen from './screens/NameScreen.svelte';
  import OfferModal from './screens/OfferModal.svelte';
  import ToastHost from './screens/ToastHost.svelte';

  onMount(connect);

  $: otherDevices = $devices.filter((device) => device.id !== $self?.id);
</script>

{#if $needsName}
  <NameScreen joining={$connection === 'joining'} error={$connectionError} onJoin={setName} />
{:else}
  <MainScreen devices={otherDevices} />
{/if}

<OfferModal offer={$offers[0] ?? null} onRespond={respondToOffer} />
<ToastHost />
