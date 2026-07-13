<script lang="ts">
  export let onOpen: () => void;
</script>

<footer>
  <button type="button" on:click={onOpen}>Add device</button>
</footer>
