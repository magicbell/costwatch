import { queryOptions } from '@tanstack/react-query';

import { fetchUsage, fetchUsagePercentiles } from './usage';

export const usageQueryOptions = queryOptions({
  queryKey: ['usage'] as const,
  queryFn: () => fetchUsage(),
});

export const percentilesQueryOptions = queryOptions({
  queryKey: ['percentiles'] as const,
  queryFn: () => fetchUsagePercentiles(),
});
