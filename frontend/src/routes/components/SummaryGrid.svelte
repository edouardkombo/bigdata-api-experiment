<script lang="ts">
  import WithSkeleton from '$lib/WithSkeleton.svelte';
  export let total_events: number|null = null;
  export let unique_users: number|null = null;
  export let events_last_hour: number|null = null;
</script>

<style>
  .grid { display: grid; grid-template-columns: repeat(3,1fr); gap:1rem; margin-bottom:1.5rem; }
  .card { padding:1rem; background:white; border-radius:0.5rem; box-shadow:0 1px 3px rgba(0,0,0,0.1); text-align:center; }
  .title { font-size:0.9rem; color:#555; }
  .value { font-size:1.8rem; font-weight:600; margin-top:0.5rem; }
</style>

<div class="grid">
  {#each [
    { label: 'Total Events',       value: total_events },
    { label: 'Unique Users',       value: unique_users },
    { label: 'Events Last Hour',   value: events_last_hour }
  ] as card}
    <div class="card">
      <div class="title">{card.label}</div>
      <WithSkeleton loading={card.value === null} error={false}>
        <div class="value">{card.value?.toLocaleString()}</div>
      </WithSkeleton>
    </div>
  {/each}
</div>

