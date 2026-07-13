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

<style>
  header { display: flex; align-items: center; justify-content: space-between; gap: var(--space-md); border-block-end: thin solid var(--color-border); padding-block: var(--space-md); }
  h1 { margin: 0; color: var(--color-accent); font-size: var(--text-headline); font-weight: var(--weight-semibold); line-height: var(--leading-headline); }
  p { display: flex; align-items: center; gap: var(--space-xs); margin: 0; color: var(--color-text-muted); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); }
  form { display: flex; flex-wrap: wrap; align-items: center; justify-content: flex-end; gap: var(--space-sm); }
  label { color: var(--color-text-muted); font-size: var(--text-body-sm); }
  input { min-block-size: var(--size-touch); min-inline-size: 0; border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface-sunken); color: var(--color-text); padding-inline: var(--space-sm); }
  input:focus { border-color: var(--color-accent); }
  button { min-block-size: var(--size-touch); min-inline-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: transparent; color: var(--color-text); cursor: pointer; transition: background var(--dur-fast) var(--ease-standard), transform var(--dur-fast) var(--ease-standard); }
  button:hover { background: var(--color-surface-hover); }
  button:active { border-color: var(--color-accent); transform: translateY(var(--press-shift)); }
</style>
