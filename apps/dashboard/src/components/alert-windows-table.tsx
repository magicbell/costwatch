import { Table } from '@chakra-ui/react';
import React from 'react';

import { formatCurrency, formatDateTime, formatNumber } from '../lib/format';

export type AlertWindowItem = {
  service: string;
  metric: string;
  start: string; // ISO date
  end: string; // ISO date
  expected_cost: number;
  real_cost: number;
};

export type AlertWindowsResponse = {
  from_date: string;
  to_date: string;
  interval: number; // seconds
  items: AlertWindowItem[];
};

export type AlertWindowsTableProps = {
  data: AlertWindowsResponse;
};

function AlertWindowsTableComponent({ data }: AlertWindowsTableProps) {
  const rows = React.useMemo(() => {
    return data.items.slice().sort((a, b) => new Date(b.start).getTime() - new Date(a.start).getTime());
  }, [data.items]);

  if (rows.length === 0) return null;

  return (
    <Table.Root size="sm">
      <Table.Header>
        <Table.Row>
          <Table.ColumnHeader>Start</Table.ColumnHeader>
          <Table.ColumnHeader>End</Table.ColumnHeader>
          <Table.ColumnHeader>Service</Table.ColumnHeader>
          <Table.ColumnHeader>Metric</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">Expected</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">Real</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">Diff</Table.ColumnHeader>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {rows.map((r, idx) => {
          const diff = r.real_cost - r.expected_cost;
          const pct = r.expected_cost > 0 ? (diff / r.expected_cost) * 100 : 0;
          return (
            <Table.Row key={idx}>
              <Table.Cell>{formatDateTime(r.start)}</Table.Cell>
              <Table.Cell>{formatDateTime(r.end)}</Table.Cell>
              <Table.Cell>{r.service}</Table.Cell>
              <Table.Cell>{r.metric}</Table.Cell>
              <Table.Cell textAlign="right">{formatCurrency(r.expected_cost)}</Table.Cell>
              <Table.Cell textAlign="right">{formatCurrency(r.real_cost)}</Table.Cell>
              <Table.Cell textAlign="right">
                {formatCurrency(diff)} {` ( ${formatNumber(pct, { maximumFractionDigits: 0 })}% )`}
              </Table.Cell>
            </Table.Row>
          );
        })}
      </Table.Body>
    </Table.Root>
  );
}

export const AlertWindowsTable = React.memo(AlertWindowsTableComponent);
