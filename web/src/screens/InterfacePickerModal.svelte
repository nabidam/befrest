<script lang="ts">
  import type { InterfaceChoicesMessage } from '../lib/proto';

  export let choices: InterfaceChoicesMessage | null = null;
  export let open = false;
  export let onPick: (interfaceID: string) => void;
  export let onClose: () => void;

  let dialog: HTMLDialogElement;
  let selected = '';

  $: if (choices && !choices.choices.some((choice) => choice.id === selected)) selected = choices.preselected;
  $: if (open && choices && dialog && !dialog.open) dialog.showModal();
  $: if ((!open || !choices) && dialog?.open) dialog.close();
</script>

{#if open && choices}
  <dialog bind:this={dialog} aria-labelledby="interface-picker-title" on:cancel={onClose}>
    <header>
      <h2 id="interface-picker-title">Which network are your devices on?</h2>
      <button type="button" aria-label="Close network picker" on:click={onClose}>Close</button>
    </header>
    <fieldset aria-label="Available network interfaces">
      {#each choices.choices as choice}
        <label>
          <input type="radio" bind:group={selected} value={choice.id} />
          <span>{choice.kind}</span>
          <span>{choice.address}</span>
        </label>
      {/each}
    </fieldset>
    <button class="primary" type="button" disabled={!selected} on:click={() => onPick(selected)}>Use this</button>
  </dialog>
{/if}

<style>
  dialog { inline-size: min(100% - var(--space-md), var(--size-content-max)); border: thin solid var(--color-border); border-radius: var(--radius-lg); background: var(--color-surface-raised); box-shadow: var(--shadow-layer); color: var(--color-text); padding: var(--space-lg); }
  dialog::backdrop { background: var(--color-scrim); }
  header { display: flex; align-items: start; justify-content: space-between; gap: var(--space-md); }
  h2 { margin: 0; font-size: var(--text-headline); font-weight: var(--weight-semibold); line-height: var(--leading-headline); }
  fieldset { display: grid; gap: var(--space-sm); margin-block: var(--space-lg); border: 0; padding: 0; }
  label { display: grid; grid-template-columns: auto 1fr auto; align-items: center; gap: var(--space-sm); min-block-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface); cursor: pointer; padding-inline: var(--space-md); }
  label:has(input:checked) { border-color: var(--color-accent); background: var(--color-surface-hover); }
  label span:last-child { color: var(--color-text-muted); font-family: var(--font-mono); font-size: var(--text-label-mono); line-height: var(--leading-label-mono); }
  button { min-block-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-weight: var(--weight-semibold); padding-inline: var(--space-md); }
  button:hover { background: var(--color-surface-hover); }
  .primary { inline-size: 100%; border-color: var(--color-accent); background: var(--color-accent); color: var(--color-on-accent); }
  .primary:hover { background: var(--color-accent-hover); }
  button:disabled { cursor: not-allowed; opacity: var(--opacity-disabled); }
</style>
