import adapter from '@sveltejs/adapter-auto';
import preprocess from 'svelte-preprocess';
import * as dotenv from 'dotenv';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: preprocess(),

  kit: {
    adapter: adapter(),

    // This makes all your URLs start with /
    paths: {
      base: process.env.VITE_BASE_PATH || ''
    }
  }	
};

export default config;
