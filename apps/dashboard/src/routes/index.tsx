import { Card, HStack, Spinner, Text, VStack } from '@chakra-ui/react';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

import { AnomaliesTable } from '../components/anomalies-table';
import { type AnomaliesResponse, UsageChart, type UsageResponse } from '../components/usage-chart';

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

  // Shared hover state: anomaly timestamp in ms; null when none
  const [hoveredAnomalyTs, setHoveredAnomalyTs] = React.useState<number | null>(null);
  const handleHover = React.useCallback((ts: number) => setHoveredAnomalyTs(ts), []);
  const handleLeave = React.useCallback(() => setHoveredAnomalyTs(null), []);

  return (
    <VStack align="stretch" gap={6} p={4}>
      {(usage.error || anomalies.error) && (
        <Text color="red.500">Failed to load {usage.error ? 'usage' : 'anomalies'}</Text>
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
              hoveredAnomalyTs={hoveredAnomalyTs}
            />
          )}
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
