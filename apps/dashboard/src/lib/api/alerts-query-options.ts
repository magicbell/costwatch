import { queryOptions } from '@tanstack/react-query';

import { fetchAlertRules, fetchAlertWindows } from './alerts';

export const alertWindowsQueryOptions = queryOptions({
  queryKey: ['alert-windows'] as const,
  queryFn: () => fetchAlertWindows(),
  refetchInterval: 30_000,
});

export const alertRulesQueryOptions = queryOptions({
  queryKey: ['alert-rules'] as const,
  queryFn: () => fetchAlertRules(),
  initialData: { items: [] },
  refetchInterval: 30_000,
});
