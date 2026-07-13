<script lang="ts">
  import QRCode from 'qrcode';
  import { invite, self } from '../lib/stores';

  export let open = false;
  export let onClose: () => void;
  export let onChangeNetwork: () => void;

  let dialog: HTMLDialogElement;
  let qrCode = '';
  let copiedURL: string | null = null;

  $: void createQRCode($invite?.urls.ip);
  $: if (open && dialog && !dialog.open) dialog.showModal();
  $: if (!open && dialog?.open) dialog.close();

  async function createQRCode(url: string | undefined): Promise<void> {
    qrCode = url ? await QRCode.toString(url, { type: 'svg' }) : '';
  }

  async function copyURL(url: string): Promise<void> {
    try {
      if (!navigator.clipboard) throw new Error('clipboard unavailable');
      await navigator.clipboard.writeText(url);
    } catch {
      const input = document.createElement('textarea');
      input.value = url;
      document.body.append(input);
      input.select();
      document.execCommand('copy');
      input.remove();
    }
    copiedURL = url;
  }
</script>

{#if open}
  <dialog bind:this={dialog} aria-labelledby="invite-sheet-title" on:cancel={onClose}>
    <header>
      <h2 id="invite-sheet-title">Add a device</h2>
      <button type="button" aria-label="Close invite panel" on:click={onClose}>Close</button>
    </header>
    {#if $invite}
      <div class="qr-tile" aria-label="QR code for joining this hub">{@html qrCode}</div>
      <p class="instructions">Scan with a phone camera, or type either of these:</p>
      <p class="copy-row">
        <button type="button" on:click={() => copyURL($invite.urls.mdns)}>{$invite.urls.mdns}</button>
      </p>
      <p class="copy-row">
        <button type="button" on:click={() => copyURL($invite.urls.ip)}>{$invite.urls.ip}</button>
      </p>
      {#if $self?.isHost}
        <p><button class="text-button" type="button" on:click={onChangeNetwork}>Wrong network? Change network</button></p>
      {/if}
      {#if $invite.reachabilityHint}
        <p class="warning" role="status">{$invite.reachabilityHint}</p>
      {/if}
      {#if copiedURL}
        <p class="copied" aria-live="polite">Copied {copiedURL}</p>
      {/if}
    {:else}
      <div class="qr-skeleton" aria-hidden="true"></div>
      <p class="instructions" aria-live="polite">Preparing an invite…</p>
    {/if}
  </dialog>
{/if}

<style>
  dialog { inline-size: min(100% - var(--space-md), var(--size-content-max)); max-block-size: 100%; border: thin solid var(--color-border); border-radius: var(--radius-lg); background: var(--color-surface-raised); box-shadow: var(--shadow-layer); color: var(--color-text); padding: var(--space-lg); }
  dialog::backdrop { background: var(--color-scrim); }
  header { display: flex; align-items: center; justify-content: space-between; gap: var(--space-md); margin-block-end: var(--space-lg); }
  h2, p { margin: 0; }
  h2 { font-size: var(--text-headline); font-weight: var(--weight-semibold); line-height: var(--leading-headline); }
  .qr-tile, .qr-skeleton { display: grid; place-items: center; inline-size: var(--size-qr); block-size: var(--size-qr); max-inline-size: 100%; margin-inline: auto; border-radius: var(--radius-md); background: var(--color-accent-hover); color: var(--color-on-accent); padding: var(--space-sm); }
  .qr-tile :global(svg) { inline-size: 100%; block-size: 100%; }
  .qr-skeleton { background: var(--color-surface-sunken); }
  .instructions { margin-block: var(--space-lg) var(--space-md); color: var(--color-text-muted); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); }
  .copy-row { margin-block: var(--space-sm); }
  button { min-block-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-weight: var(--weight-semibold); padding-inline: var(--space-md); transition: background var(--dur-fast) var(--ease-standard), transform var(--dur-fast) var(--ease-standard); }
  .copy-row button { inline-size: 100%; font-family: var(--font-mono); font-size: var(--text-body-lg); font-weight: var(--weight-regular); overflow-wrap: anywhere; }
  button:hover { background: var(--color-surface-hover); }
  button:active { transform: translateY(var(--press-shift)); }
  .text-button { min-block-size: auto; border: 0; background: transparent; color: var(--color-accent); padding-inline: 0; }
  .text-button:hover { background: transparent; color: var(--color-accent-hover); }
  .warning { margin-block-start: var(--space-md); border-radius: var(--radius-sm); background: var(--color-warn-bg); color: var(--color-warn-text); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); padding: var(--space-md); }
  .copied { margin-block-start: var(--space-sm); color: var(--color-text-muted); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); }
  @media (min-width: 40rem) { dialog { margin: auto; } }
</style>
