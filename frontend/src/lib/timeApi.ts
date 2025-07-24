// frontend/src/lib/timeApi.ts
const API = import.meta.env.VITE_API_BASE;

export interface TimeSeriesPoint {
  ts: string;
  count: number;
}

export async function fetchTimeSeries(
  fetchFn: typeof fetch,
  from: string,
  to: string,
  interval: string = '1 minute'
): Promise<TimeSeriesPoint[]> {
  const params = new URLSearchParams({ from, to, interval });
  const res = await fetchFn(`${API}/metrics/time-series?${params.toString()}`);
  if (!res.ok) throw new Error(`Time series fetch failed: ${res.status}`);
  return res.json();
}

export async function fetchTypeBreakdown(
  fetchFn: typeof fetch
): Promise<Record<string, number>> {
  const res = await fetchFn(`${API}/metrics/type-breakdown`);
  if (!res.ok) throw new Error(`Type breakdown fetch failed: ${res.status}`);
  return res.json();
}

