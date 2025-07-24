import adapter from '@sveltejs/adapter-auto';
import preprocess from 'svelte-preprocess';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: preprocess(),

  kit: {
    adapter: adapter(),

    // This makes all your URLs start with /
    paths: {
      base: '/bigdata'
    }
  }	
};

export default config;
