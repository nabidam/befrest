<script lang="ts">
  import { self } from '../lib/stores';
</script>

<header>
  <h1 id="main-title">befrest</h1>
  {#if $self}
    <p>You: {$self.name}</p>
  {/if}
</header>
