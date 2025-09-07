import { tanstackRouter } from '@tanstack/router-plugin/vite';
import viteReact from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  plugins: [tanstackRouter({ autoCodeSplitting: true }), viteReact(), tsconfigPaths()],
  server: {
    port: parseInt(process.env.PORT || "3000"),
    host: true,
    hmr: {
      protocol: 'ws',
      host: '0.0.0.0',
    }
  },
});
