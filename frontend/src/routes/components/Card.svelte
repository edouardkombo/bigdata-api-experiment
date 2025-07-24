<script lang="ts">
  export let title: string;
  export let value: number | string;
</script>

<div class="p-4 bg-white rounded-lg shadow-md">
  <h3 class="text-sm font-medium text-gray-500">{title}</h3>
  <p class="mt-2 text-2xl font-bold text-gray-900">{value}</p>
</div>

