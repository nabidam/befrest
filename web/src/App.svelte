<script lang="ts">
  import { onMount } from 'svelte';
  import { connection, connectionError, devices, devicesLoaded, interfaceChoices, needsName, offers, self } from './lib/stores';
  import { cancelOffer, cancelTransfer, connect, pickInterface, respondToOffer, setName } from './lib/ws';
  import InterfacePickerModal from './screens/InterfacePickerModal.svelte';
  import MainScreen from './screens/MainScreen.svelte';
  import NameScreen from './screens/NameScreen.svelte';
  import OfferModal from './screens/OfferModal.svelte';
  import ToastHost from './screens/ToastHost.svelte';

  onMount(connect);

  let interfacePickerOpen = false;
  let lastChoices = $interfaceChoices;

  $: otherDevices = $devices.filter((device) => device.id !== $self?.id);
  $: if ($interfaceChoices && $interfaceChoices !== lastChoices) {
    interfacePickerOpen = true;
    lastChoices = $interfaceChoices;
  }

  function cancel(transferID: string, pending: boolean): void {
    if (pending) cancelOffer(transferID);
    else cancelTransfer(transferID);
  }
</script>

{#if $needsName}
  <NameScreen joining={$connection === 'joining'} error={$connectionError} onJoin={setName} />
{:else}
  <MainScreen devices={otherDevices} loading={!$devicesLoaded} onRename={setName} onCancel={cancel} onChangeNetwork={() => (interfacePickerOpen = true)} />
{/if}

<OfferModal offer={$offers[0] ?? null} onRespond={respondToOffer} />
<InterfacePickerModal choices={$interfaceChoices} open={interfacePickerOpen} onPick={(interfaceID) => { pickInterface(interfaceID); interfacePickerOpen = false; }} onClose={() => (interfacePickerOpen = false)} />
<ToastHost />
