import type { PageLoad } from './$types';
import { getOverview, getEvents } from '$lib/api';

export const load: PageLoad = async ({ fetch }) => {
  const overview = await getOverview(fetch);
  const events   = await getEvents(fetch, null, 100);
  return { overview, events, cursor: events.at(-1)?.ts };
};

