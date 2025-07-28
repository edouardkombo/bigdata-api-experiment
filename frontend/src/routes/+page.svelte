<script lang="ts">
  import SummaryGrid         from './components/SummaryGrid.svelte';
  import TimeSeriesChart    from './components/TimeSeriesChart.svelte';
  import TypeBreakdownChart from './components/TypeBreakdownChart.svelte';
  import EventList          from './components/EventList.svelte';
  export let data: {
    overview: { total: number; unique_users: number; last_hour: number };
    events: any[];
    cursor: string|null;
  };
</script>

<div class="container py-4">
  <SummaryGrid
    total={data.overview.total}
    unique_users={data.overview.unique_users}
    last_hour={data.overview.last_hour}
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
