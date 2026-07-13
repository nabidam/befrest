<script lang="ts">
  import QRCode from 'qrcode';
  import { invite } from '../lib/stores';

  export let open = false;
  export let onClose: () => void;

  let qrCode = '';
  let copiedURL: string | null = null;

  $: void createQRCode($invite?.urls.ip);

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
  <div role="dialog" aria-modal="true" aria-labelledby="invite-sheet-title">
    <header>
      <h2 id="invite-sheet-title">Add a device</h2>
      <button type="button" aria-label="Close invite panel" on:click={onClose}>Close</button>
    </header>
    {#if $invite}
      <div aria-label="QR code for joining this hub">{@html qrCode}</div>
      <p>Scan with a phone camera, or type either of these:</p>
      <p>
        <button type="button" on:click={() => copyURL($invite.urls.mdns)}>{$invite.urls.mdns}</button>
      </p>
      <p>
        <button type="button" on:click={() => copyURL($invite.urls.ip)}>{$invite.urls.ip}</button>
      </p>
      {#if copiedURL}
        <p aria-live="polite">Copied {copiedURL}</p>
      {/if}
    {:else}
      <p aria-live="polite">Preparing an invite…</p>
    {/if}
  </div>
{/if}
