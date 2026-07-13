<script lang="ts">
  export let onOpen: () => void;
</script>

<footer>
  <button type="button" on:click={onOpen}>Add device</button>
</footer>

<style>
  footer { border-block-start: thin solid var(--color-border); margin-block-start: auto; padding-block: var(--space-md); }
  button { min-block-size: var(--size-touch); border: thin solid var(--color-border); border-radius: var(--radius-sm); background: transparent; color: var(--color-text); cursor: pointer; font-weight: var(--weight-semibold); padding-inline: var(--space-md); transition: background var(--dur-fast) var(--ease-standard), border-color var(--dur-fast) var(--ease-standard), transform var(--dur-fast) var(--ease-standard); }
  button:hover { background: var(--color-surface-hover); }
  button:active { border-color: var(--color-accent); transform: translateY(var(--press-shift)); }
</style>
