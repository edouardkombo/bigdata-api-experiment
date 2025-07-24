<script lang="ts">
  import { onMount, tick } from 'svelte';
  import WithSkeleton from '$lib/WithSkeleton.svelte';
  import { fetchTypeBreakdown } from '$lib/timeApi';

  let canvas: HTMLCanvasElement;
  let loading = true;
  let error = false;

  onMount(async () => {
    try {
      const data = await fetchTypeBreakdown(fetch);

      loading = false;
      await tick();

      const { default: Chart } = await import('chart.js/auto');
      new Chart(canvas.getContext('2d'), {
        type: 'pie',
        data: {
          labels: Object.keys(data),
          datasets: [{ label: 'Count', data: Object.values(data) }]
        }
      });
    } catch (e) {
      console.error('TypeBreakdownChart error:', e);
      loading = false;
      error = true;
    }
  });
</script>

<style>
  .chart-container { position: relative; width:100%; height:250px; }
</style>

<div class="chart-container">
  <WithSkeleton {loading} {error}>
    <div slot="error" style="color:#a00; text-align:center; padding-top:1rem;">
      Failed to load breakdown.
    </div>
    <canvas bind:this={canvas} style="width:100%; height:100%"></canvas>
  </WithSkeleton>
</div>

