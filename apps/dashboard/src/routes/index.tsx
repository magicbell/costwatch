import { Box, Card, Grid, Spinner, Text, VStack } from "@chakra-ui/react";
import {
	useIsFetching,
	useIsMutating,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import React from "react";
import { useSpinDelay } from "spin-delay";

import { AlertWindowsTable } from "../components/alert-windows-table";
import { AveragesTable } from "../components/averages-table";
import { UsageChart } from "../components/usage-chart";
import { alertWindowsQueryOptions } from "../lib/api/alerts-query-options";
import {
	percentilesQueryOptions,
	usageQueryOptions,
} from "../lib/api/usage-query-options";

export const Route = createFileRoute("/")({
	component: App,
	pendingComponent: Loader,
	errorComponent: () => (
		<Text color="red.500" p={4}>
			Failed to load data. Please try again.
		</Text>
	),
});

function Loader() {
	return (
		<VStack
			position="fixed"
			inset={0}
			zIndex={1000}
			w="full"
			h="100dvh"
			align="center"
			justify="center"
			textAlign="center"
			gap={4}
			px={6}
			opacity={0.5}
		>
			<Spinner size="lg" />
			<Text fontSize="lg">Loading your dashboard…</Text>
		</VStack>
	);
}

function App() {
	const [hoveredAlert, setHoveredAlert] = React.useState<{
		start: number;
		end: number;
	} | null>(null);
	const { data: usage } = useSuspenseQuery(usageQueryOptions);
	const { data: alertWindows } = useSuspenseQuery(alertWindowsQueryOptions);
	const { data: averages } = useSuspenseQuery(percentilesQueryOptions);

	// Tiny top-right spinners for background activity
	const isUsageFetching = useSpinDelay(
		useIsFetching({ queryKey: ["usage"] }) > 0,
	);
	const isAlertWindowsFetching = useSpinDelay(
		useIsFetching({ queryKey: ["alert-windows"] }) > 0,
	);
	const isPercentilesFetching = useSpinDelay(
		useIsFetching({ queryKey: ["percentiles"] }) > 0,
	);
	const isThresholdMutating = useSpinDelay(
		useIsMutating({ mutationKey: ["update-alert-rule"] }) > 0,
		{ delay: 0 },
	);

	return (
		<VStack align="stretch" gap={6} p={4}>
			<Grid templateColumns={{ base: "1fr", lg: "repeat(2, 1fr)" }} gap={6}>
				<Card.Root variant="subtle" position="relative">
					{/* Usage card spinner (usage or alert-windows refresh) */}
					{(isUsageFetching || isAlertWindowsFetching) && (
						<Box position="absolute" top={2} right={2} opacity={0.6}>
							<Spinner size="xs" />
						</Box>
					)}
					<Card.Body gap={2}>
						<Card.Title>Usage</Card.Title>
						<Card.Description>Usage during the last 7 days</Card.Description>

						<UsageChart
							data={usage}
							alertWindows={alertWindows}
							hoveredAlertWindow={hoveredAlert}
						/>
					</Card.Body>
				</Card.Root>

				<Card.Root variant="subtle" position="relative">
					{/* Hourly costs spinner (percentiles refresh or threshold mutation) */}
					{(isPercentilesFetching || isThresholdMutating) && (
						<Box position="absolute" top={2} right={2} opacity={0.6}>
							<Spinner size="xs" />
						</Box>
					)}
					<Card.Body gap={2}>
						<Card.Title>Hourly costs</Card.Title>
						<Card.Description>
							Hourly cost percentiles in the recent days.
						</Card.Description>

						<AveragesTable data={averages} />
					</Card.Body>
				</Card.Root>
			</Grid>

			<Card.Root variant="subtle" position="relative">
				{/* Alert windows spinner (alert-windows refresh) */}
				{isAlertWindowsFetching && (
					<Box position="absolute" top={2} right={2} opacity={0.6}>
						<Spinner size="xs" />
					</Box>
				)}
				<Card.Body gap={2}>
					<Card.Title>Alert windows</Card.Title>
					<Card.Description>
						Contiguous periods when cost exceeded configured thresholds.
					</Card.Description>

					<AlertWindowsTable data={alertWindows} onHover={setHoveredAlert} />
				</Card.Body>
			</Card.Root>
		</VStack>
	);
}
