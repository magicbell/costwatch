import { Card, Grid, HStack, Spinner, Text, VStack } from '@chakra-ui/react';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import React from 'react';

import { AlertWindowsTable } from '../components/alert-windows-table';
import { AveragesTable, type PercentilesResponse } from '../components/averages-table';
import { type AlertWindowsResponse, UsageChart, type UsageResponse } from '../components/usage-chart';

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
  const [hoveredAlert, setHoveredAlert] = React.useState<{ start: number; end: number } | null>(null);
  const usage = useQuery<UsageResponse>({
    queryKey: ['usage'],
    queryFn: () => fetch('http://localhost:4000/v1/usage').then((res) => res.json()),
  });

  const alertWindows = useQuery<AlertWindowsResponse>({
    queryKey: ['alert-windows'],
    queryFn: () => fetch('http://localhost:4000/v1/alert-windows').then((res) => res.json()),
  });

  const averages = useQuery<PercentilesResponse>({
    queryKey: ['percentiles'],
    queryFn: () => fetch('http://localhost:4000/v1/usage-percentiles').then((res) => res.json()),
  });

  return (
    <VStack align="stretch" gap={6} p={4}>
      {(usage.error || alertWindows.error) && (
        <Text color="red.500">
          Failed to load
          {usage.error ? ' usage' : ''}
          {alertWindows.error ? ' alert windows' : ''}
        </Text>
      )}

      <Grid templateColumns="repeat(2, 1fr)" gap={6}>
        <Card.Root variant="subtle">
          <Card.Body gap={2}>
            <Card.Title>Usage</Card.Title>
            <Card.Description>Usage during the last 28 days</Card.Description>

            {usage.isLoading ? (
              <Loader />
            ) : (
              <UsageChart
                data={usage.data!}
                alertWindows={alertWindows.data ?? undefined}
                hoveredAlertWindow={hoveredAlert}
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
      </Grid>

      <Card.Root variant="subtle">
        <Card.Body gap={2}>
          <Card.Title>Alert windows</Card.Title>
          <Card.Description>Contiguous periods when cost exceeded configured thresholds.</Card.Description>

          {alertWindows.isLoading ? (
            <Loader />
          ) : (
            <AlertWindowsTable data={alertWindows.data!} onHover={setHoveredAlert} />
          )}
        </Card.Body>
      </Card.Root>
    </VStack>
  );
}
