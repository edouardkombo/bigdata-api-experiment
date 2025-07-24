const API = import.meta.env.VITE_API_BASE;

export async function getOverview(fetch: typeof window.fetch) {
  const res = await fetch(`${API}/metrics/overview`);
  if (!res.ok) throw new Error(`Overview fetch failed: ${res.status}`);
  return res.json();
}

export async function getEvents(fetch: typeof window.fetch, cursor: string|null, limit = 50) {
  const params = new URLSearchParams();
  if (cursor) params.set('cursor', cursor);
  params.set('limit', String(limit));

  const res = await fetch(`${API}/metrics/events?` + params);
  if (!res.ok) throw new Error(`Events fetch failed: ${res.status}`);
  return res.json();
}

