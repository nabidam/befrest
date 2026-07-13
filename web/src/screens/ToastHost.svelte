<script lang="ts">
  import { toasts } from '../lib/stores';
</script>

<section aria-live="polite" aria-atomic="true">
  {#each $toasts as toast (toast.id)}
    <p data-tone={toast.tone} role={toast.tone === 'failure' ? 'alert' : undefined}>{toast.message}</p>
  {/each}
</section>
