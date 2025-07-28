<script lang="ts">
  import { onMount, tick } from 'svelte';
  import WithSkeleton from '$lib/WithSkeleton.svelte';
  import { fetchTypeBreakdown } from '$lib/timeApi';

  let canvas: HTMLCanvasElement;
  let loading = true;
  let error = false;

  onMount(async () => {
    try {
      const raw = await fetchTypeBreakdown(fetch);
      loading = false;
      await tick();

      const entries = Object.entries(raw)
        .map(([type, count]) => [type, +count])
        .sort((a, b) => b[1] - a[1]); // sort descending

      const maxSlices = 6;
      const top = entries.slice(0, maxSlices);
      const rest = entries.slice(maxSlices);

      if (rest.length > 0) {
        const otherSum = rest.reduce((sum, [, c]) => sum + c, 0);
        top.push(['Other', otherSum]);
      }

      const labels = top.map(([t]) => t);
      const values = top.map(([, c]) => c);

      const { default: Chart } = await import('chart.js/auto');
      new Chart(canvas.getContext('2d'), {
        type: 'pie',
        data: {
          labels,
          datasets: [{
            label: 'Event Type',
            data: values,
            borderWidth: 1
          }]
        },
        options: {
          responsive: true,
          animation: false,
          plugins: {
            legend: { position: 'top' },
            tooltip: { callbacks: {
              label: (ctx) => `${ctx.label}: ${ctx.formattedValue}`
            }}
          }
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

