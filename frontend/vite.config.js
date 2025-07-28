import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    base: env.VITE_BASE_PATH || '/',
    plugins: [sveltekit()],
    server: {
      host: '0.0.0.0',
      port: env.VITE_ALLOWED_HOST_PORT,
      allowedHosts: [
        env.VITE_ALLOWED_HOST,
        'localhost',
        '127.0.0.1'
      ],
      hmr: {
          protocol: 'ws',
          host: env.VITE_ALLOWED_HOST,     // your public IP or domain name
          port: env.VITE_ALLOWED_HOST_PORT
      },
      cors: true,
      fs: { strict: false }
    }
  };
});

