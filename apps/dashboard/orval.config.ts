import { defineConfig } from 'orval';

export default defineConfig({
  api: {
    input: '../../openapi.json',
    output: {
      target: './src/lib/api/client.ts',
      client: 'fetch',
      mode: 'single',
      baseUrl: process.env.API_URL,
      namingConvention: 'kebab-case',
    },
  },
});
