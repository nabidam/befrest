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
  <p class="brand" aria-hidden="true">befrest</p>
  <h1 id="name-title">You'll appear to others as:</h1>
  <form on:submit|preventDefault={submit}>
    <label class="field-label" for="device-name">Device name</label>
    <input id="device-name" bind:this={nameInput} bind:value={name} maxlength="33" autocomplete="nickname" />
    {#if !trimmedName}
      <p class="field-message">Give this device a name</p>
    {:else if nameTooLong}
      <p class="field-message">{name.length}/32 characters</p>
    {/if}
    <button type="submit" disabled={!canJoin} aria-busy={joining}>
      {joining ? 'Joining…' : 'Join'}
    </button>
  </form>
  {#if error}
    <p class="error-message" role="alert">{error}</p>
    <button type="button" on:click={() => window.location.reload()}>Retry</button>
  {/if}
</main>

<style>
  main { display: grid; align-content: center; gap: var(--space-lg); min-block-size: 100dvh; max-inline-size: var(--size-content-max); margin-inline: auto; padding: var(--space-md); }
  .brand { margin: 0; color: var(--color-accent); font-size: var(--text-headline); font-weight: var(--weight-semibold); letter-spacing: var(--tracking-display); }
  h1 { margin: 0; font-size: var(--text-display); font-weight: var(--weight-semibold); letter-spacing: var(--tracking-display); line-height: var(--leading-display); }
  form { display: grid; gap: var(--space-sm); }
  .field-label { color: var(--color-text-muted); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); }
  input { min-block-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface-sunken); color: var(--color-text); font-size: var(--text-display); font-weight: var(--weight-semibold); letter-spacing: var(--tracking-display); line-height: var(--leading-display); padding-inline: var(--space-md); }
  input:focus { border-color: var(--color-accent); }
  .field-message, .error-message { margin: 0; color: var(--color-danger); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); }
  button { min-block-size: var(--size-touch); border: thin solid var(--color-accent); border-radius: var(--radius-sm); background: var(--color-accent); color: var(--color-on-accent); cursor: pointer; font-weight: var(--weight-semibold); transition: background var(--dur-fast) var(--ease-standard), transform var(--dur-fast) var(--ease-standard); }
  button:hover:not(:disabled) { background: var(--color-accent-hover); }
  button:active:not(:disabled) { transform: translateY(var(--press-shift)); }
  button:disabled { cursor: not-allowed; opacity: var(--opacity-disabled); }
  .error-message + button { border-color: var(--color-border); background: transparent; color: var(--color-text); }
  .error-message + button:hover { background: var(--color-surface-hover); }
  @media (min-width: 40rem) { main { padding: var(--space-margin); } }
</style>
