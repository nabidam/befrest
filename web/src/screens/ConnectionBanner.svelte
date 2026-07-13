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
  <aside role="status" aria-live="polite" aria-atomic="true">
    Connection lost — reconnecting<span class="ellipsis" aria-hidden="true">…</span>{#if elapsed >= 30} Still trying. Is the hub running?{/if}
  </aside>
{/if}

<style>
  aside { margin-inline: calc(var(--space-md) * -1); background: var(--color-warn-bg); color: var(--color-warn-text); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); padding: var(--space-sm) var(--space-md); }
  .ellipsis { display: inline-block; animation: pulse var(--dur-pulse) var(--ease-standard) infinite; }
  @keyframes pulse { 50% { opacity: var(--opacity-disabled); } }
</style>
