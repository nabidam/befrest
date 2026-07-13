<script lang="ts">
  import { onMount } from 'svelte';
  import { connection, connectionError, devices, needsName, self } from './lib/stores';
  import { connect, setName } from './lib/ws';
  import MainScreen from './screens/MainScreen.svelte';
  import NameScreen from './screens/NameScreen.svelte';

  onMount(connect);

  $: otherDevices = $devices.filter((device) => device.id !== $self?.id);
</script>

{#if $needsName}
  <NameScreen joining={$connection === 'joining'} error={$connectionError} onJoin={setName} />
{:else}
  <MainScreen devices={otherDevices} />
{/if}
