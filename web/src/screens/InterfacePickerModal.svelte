<script lang="ts">
  import type { InterfaceChoicesMessage } from '../lib/proto';

  export let choices: InterfaceChoicesMessage | null = null;
  export let open = false;
  export let onPick: (interfaceID: string) => void;
  export let onClose: () => void;

  let selected = '';

  $: if (choices && !choices.choices.some((choice) => choice.id === selected)) selected = choices.preselected;
</script>

{#if open && choices}
  <div role="dialog" aria-modal="true" aria-labelledby="interface-picker-title">
    <header>
      <h2 id="interface-picker-title">Which network are your devices on?</h2>
      <button type="button" aria-label="Close network picker" on:click={onClose}>Close</button>
    </header>
    {#each choices.choices as choice}
      <label>
        <input type="radio" bind:group={selected} value={choice.id} />
        {choice.kind} {choice.address}
      </label>
    {/each}
    <button type="button" disabled={!selected} on:click={() => onPick(selected)}>Use this</button>
  </div>
{/if}
