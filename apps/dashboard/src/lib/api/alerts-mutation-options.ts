import { mutationOptions, type QueryClient } from '@tanstack/react-query';

import { type AlertRule, updateAlertRule } from './alerts';

export const alertRuleMutationOptions = (client: QueryClient) =>
  mutationOptions({
    mutationKey: ['update-alert-rule'],
    mutationFn: (payload: AlertRule) => updateAlertRule(payload),
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: ['alert-windows'] });
    },
  });
