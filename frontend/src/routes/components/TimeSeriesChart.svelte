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

      // Import Chart.js only when canvas is ready
      const { default: Chart } = await import('chart.js/auto');

      // Preprocess once to avoid double map
      const labels = new Array(data.length);
      const counts = new Array(data.length);
      for (let i = 0; i < data.length; i++) {
        labels[i] = data[i].bucket;
        counts[i] = data[i].count;
      }

      new Chart(canvas.getContext('2d'), {
        type: 'line',
        data: {
          labels,
          datasets: [{
            label: 'Events',
            data: counts,
            fill: true,
            tension: 0.3, // optional smoothing
            pointRadius: 0, // hides points for performance
            borderWidth: 1,
          }]
        },
        options: {
          responsive: true,
          animation: false,
          plugins: {
            legend: { display: false },
            tooltip: { mode: 'index', intersect: false }
          },
          scales: {
            x: {
              type: 'time',
              time: {
                unit: 'hour',
                tooltipFormat: 'HH:mm',
              },
              ticks: {
                autoSkip: true,
                maxTicksLimit: 12
              }
            },
            y: {
              beginAtZero: true
            }
          },
          interaction: {
            mode: 'nearest',
            axis: 'x',
            intersect: false
          },
          elements: {
            line: {
              borderColor: 'rgba(75, 192, 192, 1)'
            },
            point: {
              backgroundColor: 'rgba(75, 192, 192, 1)'
            }
          }
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

