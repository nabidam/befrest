<script lang="ts">
  import type { IncomingOffer } from '../lib/stores';
  import { formatBytes } from '../lib/format';

  export let offer: IncomingOffer | null = null;
  export let onRespond: (transferID: string, accepted: boolean) => void;

  let dialog: HTMLDialogElement;

  $: if (offer && dialog && !dialog.open) dialog.showModal();
  $: if (!offer && dialog?.open) dialog.close();

  function respond(accepted: boolean): void {
    if (offer) onRespond(offer.transfer.id, accepted);
  }

  function onCancel(event: Event): void {
    event.preventDefault();
    respond(false);
  }
</script>

<dialog bind:this={dialog} aria-labelledby="incoming-title" on:cancel={onCancel}>
  {#if offer}
    <h2 id="incoming-title">Incoming files</h2>
    <p><strong>{offer.from.name}</strong> wants to send you</p>
    <ul>
      {#each offer.transfer.files.slice(0, 4) as file}
        <li>{file.name} — {formatBytes(file.size)}</li>
      {/each}
      {#if offer.transfer.files.length > 4}
        <li>and {offer.transfer.files.length - 4} more</li>
      {/if}
    </ul>
    <button type="button" on:click={() => respond(false)}>Decline</button>
    <button type="button" on:click={() => respond(true)} autofocus>Accept</button>
  {/if}
</dialog>
