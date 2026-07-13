<script lang="ts">
  import QRCode from 'qrcode';
  import { invite } from '../lib/stores';

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

<section aria-labelledby="empty-invite-title">
  <h2 id="empty-invite-title">No other devices yet</h2>
  {#if $invite}
    <div aria-label="QR code for joining this hub">{@html qrCode}</div>
    <p>Scan with a phone camera, or copy a link to open on another device.</p>
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
</section>

<style>
  section { display: grid; justify-items: center; gap: var(--space-md); margin-block: var(--space-xl); text-align: center; }
  h2, p { margin: 0; }
  h2 { font-size: var(--text-display); font-weight: var(--weight-semibold); letter-spacing: var(--tracking-display); line-height: var(--leading-display); }
  p { color: var(--color-text-muted); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); }
  div[aria-label] { display: grid; place-items: center; inline-size: var(--size-qr); block-size: var(--size-qr); border-radius: var(--radius-md); background: var(--color-accent-hover); color: var(--color-on-accent); padding: var(--space-sm); }
  div[aria-label] :global(svg) { inline-size: 100%; block-size: 100%; }
  button { min-block-size: var(--size-touch); max-inline-size: 100%; border: thin solid var(--color-border); border-radius: var(--radius-sm); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-family: var(--font-mono); font-size: var(--text-body-lg); line-height: var(--leading-body-lg); overflow-wrap: anywhere; padding-inline: var(--space-md); transition: background var(--dur-fast) var(--ease-standard), transform var(--dur-fast) var(--ease-standard); }
  button:hover { background: var(--color-surface-hover); }
  button:active { transform: translateY(var(--press-shift)); }
</style>
