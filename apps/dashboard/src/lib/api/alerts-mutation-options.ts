import { mutationOptions, type QueryClient } from '@tanstack/react-query';

import { toaster } from '@/components/ui/toaster.tsx';
import { formatCurrency } from '@/lib/format.ts';

import { type AlertRule, updateAlertRule } from './alerts';

export const alertRuleMutationOptions = (client: QueryClient) =>
  mutationOptions({
    mutationKey: ['update-alert-rule'],
    mutationFn: async (payload: AlertRule) => {
      await updateAlertRule(payload);
      return payload;
    },
    onSuccess: (x) => {
      void client.invalidateQueries({ queryKey: ['alert-windows'] });
      toaster.create({
        title: 'Threshold updated',
        description: `threshold for ${x.service} / ${x.metric} set to ${formatCurrency(x.threshold)} `,
      });
    },
    onError: (x) => {
      toaster.error({
        title: 'Failed to save threshold',
        description: x.message,
      });
    },
  });
