import { Table } from '@chakra-ui/react';
import React from 'react';

import { formatCurrency, formatDateTime, formatNumber } from '../lib/format';

export type AnomalyItem = {
  service: string;
  metric: string;
  timestamp: string; // ISO date (start time)
  sum: number;
  diff: number;
  z_score: number;
  cost: number;
};

export type AnomaliesResponse = {
  from_date: string;
  to_date: string;
  interval: number; // seconds
  items: AnomalyItem[];
};

export type AnomaliesTableProps = {
  data: AnomaliesResponse;
  onHoverTimestamp?: (ts: number) => void;
  onLeave?: () => void;
};

function AnomaliesTableComponent({ data, onHoverTimestamp, onLeave }: AnomaliesTableProps) {
  const rows = React.useMemo(() => {
    return data.items.slice().sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
  }, [data]);

  if (rows.length === 0) return null;

  return (
    <Table.Root size="sm">
      <Table.Header>
        <Table.Row>
          <Table.ColumnHeader>Time</Table.ColumnHeader>
          <Table.ColumnHeader>Service</Table.ColumnHeader>
          <Table.ColumnHeader>Metric</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">Cost</Table.ColumnHeader>
          <Table.ColumnHeader textAlign="right">Z-score</Table.ColumnHeader>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {rows.map((r, idx) => (
          <Table.Row
            key={idx}
            onMouseEnter={() => onHoverTimestamp?.(new Date(r.timestamp).getTime())}
            onMouseLeave={() => onLeave?.()}
            cursor="pointer"
            _hover={{ bg: 'gray.50', _dark: { bg: 'whiteAlpha.100' } }}
          >
            <Table.Cell>{formatDateTime(r.timestamp)}</Table.Cell>
            <Table.Cell>{r.service}</Table.Cell>
            <Table.Cell>{r.metric}</Table.Cell>
            <Table.Cell textAlign="right">{formatCurrency(r.cost)}</Table.Cell>
            <Table.Cell textAlign="right">
              {typeof r.z_score === 'number'
                ? formatNumber(r.z_score, { maximumFractionDigits: 2, minimumFractionDigits: 2 })
                : '—'}
            </Table.Cell>
          </Table.Row>
        ))}
      </Table.Body>
    </Table.Root>
  );
}

export const AnomaliesTable = React.memo(AnomaliesTableComponent);
