<script lang="ts">
  import { onMount } from 'svelte';
  import { suggestedName } from '../lib/stores';

  export let joining = false;
  export let error: string | null = null;
  export let onJoin: (name: string) => void;

  let name = '';
  let lastSuggested = '';
  let nameInput: HTMLInputElement;

  $: if ($suggestedName !== lastSuggested) {
    lastSuggested = $suggestedName;
    name = $suggestedName;
  }
  $: trimmedName = name.trim();
  $: nameTooLong = name.length > 32;
  $: canJoin = Boolean(trimmedName) && !nameTooLong && !joining;

  function submit(): void {
    if (canJoin) onJoin(trimmedName);
  }

  onMount(() => {
    nameInput.focus();
    nameInput.select();
  });
</script>

<main aria-labelledby="name-title">
  <p aria-hidden="true">befrest</p>
  <h1 id="name-title">You'll appear to others as:</h1>
  <form on:submit|preventDefault={submit}>
    <label for="device-name">Device name</label>
    <input id="device-name" bind:this={nameInput} bind:value={name} maxlength="33" autocomplete="nickname" />
    {#if !trimmedName}
      <p>Give this device a name</p>
    {:else if nameTooLong}
      <p>{name.length}/32 characters</p>
    {/if}
    <button type="submit" disabled={!canJoin} aria-busy={joining}>
      {joining ? 'Joining…' : 'Join'}
    </button>
  </form>
  {#if error}
    <p role="alert">{error}</p>
    <button type="button" on:click={() => window.location.reload()}>Retry</button>
  {/if}
</main>
