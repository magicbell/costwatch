import { queryOptions } from '@tanstack/react-query';
import * as client from './client';

export const usageQueryOptions = queryOptions({
  queryKey: ['usage'] as const,
  queryFn: () => client.usage().then(x => x.data),
  refetchInterval: 30_000,
});

export const percentilesQueryOptions = queryOptions({
  queryKey: ['percentiles'] as const,
  queryFn: () => client.usagePercentiles().then(x => x.data),
  refetchInterval: 30_000,
});
