import { defineConfig } from 'orval';

export default defineConfig({
  api: {
    input: '../../openapi.json',
    output: {
      target: './src/lib/api/client.ts',
      client: 'fetch',
      mode: 'single',
      baseUrl: '__API_URL__',
      namingConvention: 'kebab-case',
    },
    hooks: {
      afterAllFilesWrite: 'npx -y biome check --unsafe --write src/lib/api/client.ts',
    }
  },
});
