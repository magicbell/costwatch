import { Chart, useChart } from '@chakra-ui/charts';
import React from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

import { formatCurrency, formatDate, formatDateTime, formatTime } from '../lib/format';

export type UsageItem = {
  service: string;
  metric: string;
  cost: number;
  timestamp: string; // ISO date
};

export type UsageResponse = {
  from_date: string;
  to_date: string;
  interval: number; // seconds
  items: UsageItem[];
};

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

export type UsageChartProps = {
  data: UsageResponse;
  anomalies?: AnomaliesResponse;
  hoveredAnomalyTs?: number | null;
};

function UsageChartComponent({ data, anomalies, hoveredAnomalyTs }: UsageChartProps) {
  const chartData = React.useMemo(
    () =>
      data.items.map((x) => {
        return {
          [`${x.service} / ${x.metric}`]: x.cost,
          ts: new Date(x.timestamp).getTime(),
        };
      }),
    [data.items],
  );

  const series = React.useMemo(
    () =>
      Array.from(new Set(data.items.map((x) => `${x.service} / ${x.metric}`))).map((x) => ({
        name: x,
        color: 'purple.solid',
      })),
    [data.items],
  );

  const chart = useChart({ data: chartData, series });

  const anomalyLines = React.useMemo(() => {
    if (!anomalies || anomalies.items.length === 0) return [] as { x: number }[];
    return anomalies.items.map((it) => ({ x: new Date(it.timestamp).getTime() }));
  }, [anomalies]);

  const xTickFormatter = React.useCallback(
    (value: number, idx: number) => {
      const d = new Date(value);
      const h = d.getUTCHours();
      const m = d.getUTCMinutes();
      if (idx === 0 || idx === chartData.length - 1) return '';
      if (h === 0 && m === 0) {
        return formatDate(d);
      }
      return '';
    },
    [chartData.length],
  );

  const yTickCount = 5;
  const yTickFormatter = React.useCallback((v: number, idx: number) => {
    if (idx === 0 || idx === yTickCount - 1) return '';
    return formatCurrency(v);
  }, []);

  const tooltipLabelFormatter = React.useCallback((value: number) => {
    if (Number.isNaN(Number(value))) return '';
    return formatDateTime(Number(value));
  }, []);

  const tooltipValueFormatter = React.useCallback((value: number) => formatCurrency(value), []);

  const domain = React.useMemo(() => {
    const min = Math.min(...chart.data.map((x) => x[chart.key('ts')]));
    const max = Math.max(...chart.data.map((x) => x[chart.key('ts')]));
    return [min, max];
  }, [chart.data, chart.key('ts')]);

  return (
    <ResponsiveContainer width="100%" height={300}>
      <Chart.Root maxH="sm" chart={chart}>
        <BarChart data={chart.data} barCategoryGap={1}>
          <CartesianGrid stroke={chart.color('border.muted')} vertical={false} />
          <XAxis
            axisLine={false}
            tickLine={false}
            type="number"
            dataKey={chart.key('ts')}
            domain={domain}
            scale="time"
            interval={0}
            tickFormatter={xTickFormatter}
          />
          <YAxis
            axisLine={false}
            tickLine={false}
            tickMargin={12}
            tickCount={yTickCount}
            tickFormatter={yTickFormatter}
          />
          <Tooltip
            cursor={{ fill: chart.color('border.muted') }}
            animationDuration={100}
            content={<Chart.Tooltip />}
            labelFormatter={tooltipLabelFormatter}
            formatter={(value) => tooltipValueFormatter(value as number)}
            labelStyle={{ fill: chart.color('fg.muted'), fontSize: 10, fontWeight: 'bold' }}
            itemStyle={{}}
          />
          <Legend content={<Chart.Legend />} />

          {anomalyLines.map((a, i) => {
            const isHovered = hoveredAnomalyTs != null && a.x === hoveredAnomalyTs;
            return (
              <ReferenceLine
                key={`anomaly-${i}`}
                x={a.x}
                ifOverflow="extendDomain"
                stroke={isHovered ? chart.color('fg.error') : chart.color('border.error')}
                strokeDasharray={isHovered ? undefined : '4 2'}
                strokeWidth={isHovered ? 3 : 1}
                label={{
                  value: formatTime(a.x),
                  position: 'top',
                  dy: 12,
                  fill: chart.color('fg.error'),
                  style: { fontSize: 10, fontWeight: 'bold' },
                }}
              />
            );
          })}

          {chart.series.map((item) => (
            <Bar
              isAnimationActive={false}
              key={item.name}
              dataKey={chart.key(item.name)}
              fill={chart.color(item.color)}
              stackId={item.stackId}
            />
          ))}
        </BarChart>
      </Chart.Root>
    </ResponsiveContainer>
  );
}

export const UsageChart = React.memo(UsageChartComponent);
