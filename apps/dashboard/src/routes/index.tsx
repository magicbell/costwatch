import { Card, HStack, Spinner, Text, VStack } from '@chakra-ui/react';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

import { AlertWindowsTable } from '../components/alert-windows-table';
import { AnomaliesTable } from '../components/anomalies-table';
import { AveragesTable, type PercentilesResponse } from '../components/averages-table';
import {
  type AlertWindowsResponse,
  type AnomaliesResponse,
  UsageChart,
  type UsageResponse,
} from '../components/usage-chart';

export const Route = createFileRoute('/')({
  component: App,
});

function Loader() {
  return (
    <HStack>
      <Spinner />
      <Text>Loading data…</Text>
    </HStack>
  );
}

function App() {
  const usage = useQuery<UsageResponse>({
    queryKey: ['usage'],
    queryFn: () => fetch('http://localhost:4000/v1/usage').then((res) => res.json()),
  });

  const anomalies = useQuery<AnomaliesResponse>({
    queryKey: ['anomalies'],
    queryFn: () => fetch('http://localhost:4000/v1/anomalies').then((res) => res.json()),
  });

  const alertWindows = useQuery<AlertWindowsResponse>({
    queryKey: ['alert-windows'],
    queryFn: () => fetch('http://localhost:4000/v1/alert-windows').then((res) => res.json()),
  });

  const averages = useQuery<PercentilesResponse>({
    queryKey: ['percentiles'],
    queryFn: () => fetch('http://localhost:4000/v1/usage-percentiles').then((res) => res.json()),
  });

  // Shared hover state: anomaly timestamp in ms; null when none
  const [hoveredAnomalyTs, setHoveredAnomalyTs] = React.useState<number | null>(null);
  const handleHover = React.useCallback((ts: number) => setHoveredAnomalyTs(ts), []);
  const handleLeave = React.useCallback(() => setHoveredAnomalyTs(null), []);

  return (
    <VStack align="stretch" gap={6} p={4}>
      {(usage.error || anomalies.error || alertWindows.error) && (
        <Text color="red.500">
          Failed to load
          {usage.error ? ' usage' : ''}
          {anomalies.error ? ' anomalies' : ''}
          {alertWindows.error ? ' alert windows' : ''}
        </Text>
      )}

      <Card.Root variant="subtle">
        <Card.Body gap={2}>
          <Card.Title>Usage</Card.Title>
          <Card.Description>Usage during the last 28 days</Card.Description>

          {usage.isLoading ? (
            <Loader />
          ) : (
            <UsageChart
              data={usage.data!}
              anomalies={anomalies.data ?? undefined}
              alertWindows={alertWindows.data ?? undefined}
              hoveredAnomalyTs={hoveredAnomalyTs}
            />
          )}
        </Card.Body>
      </Card.Root>

      <Card.Root variant="subtle">
        <Card.Body gap={2}>
          <Card.Title>Hourly costs</Card.Title>
          <Card.Description>Hourly cost percentiles in the recent days.</Card.Description>

          {averages.isLoading ? <Loader /> : <AveragesTable data={averages.data!} />}
        </Card.Body>
      </Card.Root>

      <Card.Root variant="subtle">
        <Card.Body gap={2}>
          <Card.Title>Alert windows</Card.Title>
          <Card.Description>Contiguous periods when cost exceeded configured thresholds.</Card.Description>

          {alertWindows.isLoading ? <Loader /> : <AlertWindowsTable data={alertWindows.data!} />}
        </Card.Body>
      </Card.Root>

      <Card.Root variant="subtle">
        <Card.Body gap={2}>
          <Card.Title>Anomalies</Card.Title>
          <Card.Description>Anomalies detected in the last 28 days.</Card.Description>

          {anomalies.isLoading ? (
            <Loader />
          ) : (
            <AnomaliesTable data={anomalies.data!} onHoverTimestamp={handleHover} onLeave={handleLeave} />
          )}
        </Card.Body>
      </Card.Root>
    </VStack>
  );
}
