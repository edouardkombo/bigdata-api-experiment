<script lang="ts">
  import { onMount, tick } from 'svelte';
  import WithSkeleton from '$lib/WithSkeleton.svelte';
  import { fetchTimeSeries } from '$lib/timeApi';
  import 'chartjs-adapter-date-fns';

  let canvas: HTMLCanvasElement;
  let loading = true;
  let error = false;

  onMount(async () => {
    try {
      const now  = new Date();
      const from = new Date(now.getTime() - 24*60*60*1000).toISOString();
      const to   = now.toISOString();
      const data = await fetchTimeSeries(fetch, from, to, '5 minute');

      // Data is ready—hide skeleton and allow canvas to mount
      loading = false;
      await tick();

      // Now canvas is in the DOM
      const { default: Chart } = await import('chart.js/auto');
      new Chart(canvas.getContext('2d'), {
        type: 'line',
        data: {
          labels: data.map(d => d.bucket),
          datasets: [{ label: 'Events', data: data.map(d => d.count), fill: true }]
        },
        options: {
          scales: { x: { type: 'time', time: { unit: 'hour' } } }
        }
      });
    } catch (e) {
      console.error('TimeSeriesChart error:', e);
      loading = false;
      error = true;
    }
  });
</script>

<style>
  .chart-container { position: relative; width:100%; height:300px; }
</style>

<div class="chart-container">
  <WithSkeleton {loading} {error}>
    <div slot="error" style="color:#a00; text-align:center; padding-top:1rem;">
      Failed to load time-series.
    </div>
    <canvas bind:this={canvas} style="width:100%; height:100%"></canvas>
  </WithSkeleton>
</div>

