<script lang="ts">
  export let loading: boolean;
  export let error: boolean;
</script>

<style>
  .skeleton {
    background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 37%, #f0f0f0 63%);
    background-size: 400% 100%;
    animation: pulse 1.4s ease infinite;
    border-radius: 4px;
  }
  @keyframes pulse { to { background-position: -200% 0; } }
</style>

{#if loading}
  <div class="skeleton" style="width:100%; height:100%"></div>
{:else if error}
  <slot name="error" />
{:else}
  <slot />
{/if}

