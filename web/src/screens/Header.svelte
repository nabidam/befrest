<script lang="ts">
  import { self } from '../lib/stores';

  export let onRename: (name: string) => void;

  let editing = false;
  let name = '';

  function beginRename(): void {
    name = $self?.name ?? '';
    editing = true;
  }

  function submitRename(): void {
    if (!name.trim()) return;
    onRename(name);
    editing = false;
  }
</script>

<header>
  <h1 id="main-title">befrest</h1>
  {#if $self}
    {#if editing}
      <form on:submit|preventDefault={submitRename}>
        <label for="device-name">Your name</label>
        <input id="device-name" bind:value={name} maxlength="32" required />
        <button type="submit">Save</button>
        <button type="button" on:click={() => (editing = false)}>Cancel</button>
      </form>
    {:else}
      <p>You: {$self.name} <button type="button" on:click={beginRename} aria-label="Rename this device">✎</button></p>
    {/if}
  {/if}
</header>
