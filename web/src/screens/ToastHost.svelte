<script lang="ts">
  import { toasts } from '../lib/stores';
</script>

<section aria-live="polite" aria-atomic="true" aria-label="Notifications">
  {#each $toasts as toast (toast.id)}
    <p data-tone={toast.tone}>{toast.message}</p>
  {/each}
</section>

<style>
  section { position: fixed; inset-inline: var(--space-md); inset-block-end: var(--space-md); display: grid; gap: var(--space-sm); margin-inline: auto; max-inline-size: var(--size-content-max); pointer-events: none; }
  p { border: thin solid var(--color-border); border-inline-start: var(--space-xs) solid var(--color-accent); border-radius: var(--radius-md); background: var(--color-surface-raised); box-shadow: var(--shadow-layer); color: var(--color-text); font-size: var(--text-body-sm); line-height: var(--leading-body-sm); margin: 0; padding: var(--space-md); }
  p[data-tone='failure'] { border-inline-start-color: var(--color-danger); }
</style>
