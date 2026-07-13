<script lang="ts">
  import { tick } from 'svelte';
  import type { IncomingOffer } from '../lib/stores';
  import { formatAdditionalFiles, formatBytes } from '../lib/format';

  export let offer: IncomingOffer | null = null;
  export let onRespond: (transferID: string, accepted: boolean) => void;

  let dialog: HTMLDialogElement;
  let fileInfo: HTMLElement;
  let previousFocus: HTMLElement | null = null;

  $: if (offer && dialog && !dialog.open) openDialog();
  $: if (!offer && dialog?.open) dialog.close();

  async function openDialog(): Promise<void> {
    previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    dialog.showModal();
    await tick();
    fileInfo?.focus();
  }

  function respond(accepted: boolean): void {
    if (offer) onRespond(offer.transfer.id, accepted);
  }

  function onCancel(event: Event): void {
    event.preventDefault();
    respond(false);
  }

  function restoreFocus(): void {
    previousFocus?.focus();
    previousFocus = null;
  }
</script>

<dialog bind:this={dialog} aria-labelledby="incoming-title" on:cancel={onCancel} on:close={restoreFocus}>
  {#if offer}
    <section bind:this={fileInfo} tabindex="-1" aria-labelledby="incoming-title">
      <h2 id="incoming-title">Incoming files</h2>
      <p><strong>{offer.from.name}</strong> wants to send you</p>
      <ul>
        {#each offer.transfer.files.slice(0, 4) as file}
          <li><span>{file.name}</span> <span class="file-size">{formatBytes(file.size)}</span></li>
        {/each}
        {#if offer.transfer.files.length > 4}
          <li class="file-size">
            {formatAdditionalFiles(
              offer.transfer.files.length - 4,
              offer.transfer.files.reduce((total, file) => total + file.size, 0),
            )}
          </li>
        {/if}
      </ul>
    </section>
    <div class="actions">
      <button type="button" on:click={() => respond(false)}>Decline</button>
      <button class="primary" type="button" on:click={() => respond(true)}>Accept</button>
    </div>
  {/if}
</dialog>

<style>
  dialog { inline-size: min(100% - var(--space-md), var(--size-content-max)); border: thin solid var(--color-border); border-radius: var(--radius-lg); background: var(--color-surface-raised); box-shadow: var(--shadow-layer); color: var(--color-text); padding: var(--space-lg); }
  dialog::backdrop { background: var(--color-scrim); }
  h2, p, ul { margin: 0; }
  h2 { font-size: var(--text-headline); font-weight: var(--weight-semibold); line-height: var(--leading-headline); }
  p { margin-block-start: var(--space-md); font-size: var(--text-body-lg); line-height: var(--leading-body-lg); }
  ul { display: grid; gap: var(--space-sm); margin-block: var(--space-lg); padding-inline-start: var(--space-lg); }
  li { display: flex; justify-content: space-between; gap: var(--space-md); font-size: var(--text-body-lg); line-height: var(--leading-body-lg); }
  .file-size { color: var(--color-text-muted); font-family: var(--font-mono); font-size: var(--text-label-mono); line-height: var(--leading-label-mono); }
  .actions { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-md); }
  button { min-block-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-weight: var(--weight-semibold); transition: background var(--dur-fast) var(--ease-standard), transform var(--dur-fast) var(--ease-standard); }
  button:hover { background: var(--color-surface-hover); }
  button:active { transform: translateY(var(--press-shift)); }
  .primary { border-color: var(--color-accent); background: var(--color-accent); color: var(--color-on-accent); }
  .primary:hover { background: var(--color-accent-hover); }
</style>
