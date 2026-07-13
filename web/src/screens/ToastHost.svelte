<script lang="ts">
  import { toasts } from '../lib/stores';
</script>

<section aria-live="polite" aria-atomic="true">
  {#each $toasts as toast (toast.id)}
    <p data-tone={toast.tone}>{toast.message}</p>
  {/each}
</section>
