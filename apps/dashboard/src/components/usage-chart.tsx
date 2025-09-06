import { Chart, useChart } from '@chakra-ui/charts';
import React from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ReferenceArea,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

import type { AlertWindowsResponse } from '@/lib/api/alerts';
import type { UsageResponse } from '@/lib/api/usage';

import { formatCurrency, formatDate, formatDateTime } from '../lib/format';

export type UsageChartProps = {
  data: UsageResponse;
  alertWindows?: AlertWindowsResponse;
  hoveredAlertWindow?: { start: number; end: number } | null;
};

function UsageChartComponent({ data, alertWindows, hoveredAlertWindow }: UsageChartProps) {
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

  const alertAreas = React.useMemo(() => {
    if (!alertWindows || alertWindows.items.length === 0) return [] as { x1: number; x2: number }[];
    const domainMax = Math.max(...chart.data.map((x) => x[chart.key('ts')]));
    return alertWindows.items.map((w) => ({
      x1: new Date(w.start).getTime(),
      x2: w.end ? new Date(w.end).getTime() : domainMax,
    }));
  }, [alertWindows, chart]);

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

          {alertAreas.map((a, i) => {
            const isHovered =
              !!hoveredAlertWindow && a.x1 === hoveredAlertWindow.start && a.x2 === hoveredAlertWindow.end;
            return (
              <ReferenceArea
                key={`alert-${i}`}
                ifOverflow="extendDomain"
                x1={a.x1}
                x2={a.x2}
                fill={chart.color('bg.error')}
                fillOpacity={hoveredAlertWindow && !isHovered ? 0.1 : 1}
                stroke={chart.color('border.error')}
                strokeDasharray="4 4"
                strokeOpacity={hoveredAlertWindow && !isHovered ? 0.1 : 0.4}
                strokeWidth={1}
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
