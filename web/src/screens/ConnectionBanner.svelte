<script lang="ts">
  import { onDestroy } from 'svelte';
  import { connection, self } from '../lib/stores';

  let elapsed = 0;
  let timer: number | undefined;

  $: reconnecting = $connection === 'connecting' && $self !== null;
  $: if (reconnecting && timer === undefined) {
    elapsed = 0;
    timer = window.setInterval(() => (elapsed += 1), 1_000);
  } else if (!reconnecting && timer !== undefined) {
    window.clearInterval(timer);
    timer = undefined;
    elapsed = 0;
  }

  onDestroy(() => {
    if (timer !== undefined) window.clearInterval(timer);
  });
</script>

{#if reconnecting}
  <aside role="status" aria-live="polite">
    Connection lost — reconnecting…{#if elapsed >= 30} Still trying. Is the hub running?{/if}
  </aside>
{/if}
