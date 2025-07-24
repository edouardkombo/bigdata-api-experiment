<script lang="ts">
  import SummaryGrid         from './components/SummaryGrid.svelte';
  import TimeSeriesChart    from './components/TimeSeriesChart.svelte';
  import TypeBreakdownChart from './components/TypeBreakdownChart.svelte';
  import EventList          from './components/EventList.svelte';
  export let data: {
    overview: { total_events: number; unique_users: number; events_last_hour: number };
    events: any[];
    cursor: string|null;
  };
</script>

<div class="container py-4">
  <SummaryGrid
    total_events={data.overview.total_events}
    unique_users={data.overview.unique_users}
    events_last_hour={data.overview.events_last_hour}
  />

  <div class="row mb-4">
    <div class="col-md-8">
      <TimeSeriesChart />
    </div>
    <div class="col-md-4">
      <TypeBreakdownChart />
    </div>
  </div>

  <EventList initialEvents={data.events} initialCursor={data.cursor} />
</div>
