<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import WithSkeleton from '$lib/WithSkeleton.svelte';
  import { getEvents } from '$lib/api';

  export let initialEvents: any[] = [];
  export let initialCursor: string|null = null;

  let events = [...initialEvents];
  let cursor = initialCursor;
  let loading = false;
  let finished = false;
  const dispatch = createEventDispatcher();

  async function loadMore() {
    if (loading || finished) return;
    loading = true;
    try {
      const more = await getEvents(fetch, cursor, 50);
      if (more.length) {
        events = [...events, ...more];
        cursor = more.at(-1)?.ts ?? cursor;
      } else {
        finished = true;
        dispatch('end');
      }
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }
</script>

<style>
  .list { display:flex; flex-direction:column; gap:0.5rem; }
  .row  { padding:0.75rem; background:white; border-radius:0.25rem; box-shadow:0 1px 2px rgba(0,0,0,0.05); }
  .load { text-align:center; margin-top:1rem; }
  button[disabled] { opacity:0.6; }
</style>

<div class="list">
  {#each events as ev, i (`${ev.id}-${i}`)}
    <div class="row">
      <div class="text-xs text-gray-500">{ev.ts}</div>
      <div class="mt-1 text-sm">
        <strong>{ev.event_type}</strong> —
        <a href={ev.url} target="_blank" rel="noopener">{ev.url}</a>
      </div>
    </div>
  {:else}
    <WithSkeleton loading={events.length===0} error={false}>
      <div slot="error">No events found.</div>
      <!-- skeleton rows -->
      {#each Array(5) as _}
        <div class="row" style="background:#f0f0f0; color:transparent;">Loading…</div>
      {/each}
    </WithSkeleton>
  {/each}
</div>

<div class="load">
  {#if !finished}
    <button on:click={loadMore} disabled={loading}>
      {#if loading}Loading…{:else}Load more{/if}
    </button>
  {/if}
</div>

