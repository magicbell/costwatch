import { Input, Table } from '@chakra-ui/react';
import { useMutation, useQuery } from '@tanstack/react-query';
import React from 'react';

import { formatCurrency, formatNumber } from '../lib/format';

export type PercentileItem = {
  service: string;
  metric: string;
  p50: number;
  p90: number;
  p95: number;
  pmax: number;
};

export type PercentilesResponse = {
  from_date: string;
  to_date: string;
  interval: number; // seconds
  items: PercentileItem[];
};

export type AlertRule = {
  service: string;
  metric: string;
  threshold: number;
};

export type AveragesTableProps = {
  data: PercentilesResponse;
};

function AveragesTableComponent({ data }: AveragesTableProps) {
  const rows = React.useMemo(() => {
    return data.items.slice().sort((a, b) => {
      if (a.service === b.service) return a.metric.localeCompare(b.metric);
      return a.service.localeCompare(b.service);
    });
  }, [data]);

  const query = useQuery<{ items: AlertRule[] }>({
    queryKey: ['alert-rules'],
    queryFn: async () => {
      const res = await fetch('http://localhost:4000/v1/alert-rules');
      if (!res.ok) throw new Error('Failed to fetch alert rules');
      return res.json();
    },
    initialData: { items: [] },
    refetchInterval: 10000,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  });

  const mutation = useMutation({
    mutationKey: ['update-threshold'],
    mutationFn: async (payload: { service: string; metric: string; threshold: number }) => {
      const res = await fetch('http://localhost:4000/v1/alert-rules', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error('Failed to update threshold');
      return true;
    },
  });

  if (rows.length === 0) return null;

  return (
    <Table.Root size="sm">
      <Table.Header>
        <Table.Row>
          <Table.ColumnHeader>Service</Table.ColumnHeader>
          <Table.ColumnHeader>Metric</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">P50</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">P90</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">P95</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right" pr={8}>
            Max
          </Table.ColumnHeader>
          <Table.ColumnHeader textAlign="left" w={32}>
            Alert threshold
          </Table.ColumnHeader>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {rows.map((r, idx) => {
          const value = (query.data?.items || []).find(
            (x) => x.service === r.service && x.metric === r.metric,
          )?.threshold;

          const defaultValue =
            value != null ? formatNumber(value, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : undefined;

          return (
            <Table.Row key={idx}>
              <Table.Cell>{r.service}</Table.Cell>
              <Table.Cell>{r.metric}</Table.Cell>
              <Table.Cell textAlign="right">{formatCurrency(r.p50)}</Table.Cell>
              <Table.Cell textAlign="right">{formatCurrency(r.p90)}</Table.Cell>
              <Table.Cell textAlign="right">{formatCurrency(r.p95)}</Table.Cell>
              <Table.Cell textAlign="right" pr={8}>
                {formatCurrency(r.pmax)}
              </Table.Cell>
              <Table.Cell textAlign="right" py={0} w={32}>
                <Input
                  size="sm"
                  variant="flushed"
                  border={0}
                  placeholder={formatNumber(r.p95, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  defaultValue={defaultValue}
                  onBlur={(e) => {
                    const value = Number(e.currentTarget.value.replaceAll(',', '.'));
                    mutation.mutate({ service: r.service, metric: r.metric, threshold: value });
                  }}
                />
              </Table.Cell>
            </Table.Row>
          );
        })}
      </Table.Body>
    </Table.Root>
  );
}

export const AveragesTable = React.memo(AveragesTableComponent);
