import { tanstackRouter } from '@tanstack/router-plugin/vite';
import viteReact from '@vitejs/plugin-react';
import { defineConfig, loadEnv } from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';
import replace from '@rollup/plugin-replace';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const apiUrl = (env.API_URL || 'http://localhost:3010/v1').replace(/\/$/, '');

  return {
    plugins: [
      tanstackRouter({ autoCodeSplitting: true }),
      viteReact(),
      tsconfigPaths(),
      replace({
        preventAssignment: true,
        values: {
          __API_URL__: apiUrl,
        },
      }),
    ],
    server: {
      port: parseInt(process.env.PORT || "3000"),
        host: true,
        hmr: {
        protocol: 'ws',
          host: '0.0.0.0',
      }
    },
  };
})
