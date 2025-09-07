import { Table } from "@chakra-ui/react";
import React from "react";

import { formatCurrency, formatDateTime, formatNumber } from "../lib/format";
import type { AlertWindowsQueryResponse } from '../lib/api/client';

export type AlertWindowsTableProps = {
	data: AlertWindowsQueryResponse;
	onHover?: (range: { start: number; end: number } | null) => void;
};

function AlertWindowsTableComponent({ data, onHover }: AlertWindowsTableProps) {
	const rows = React.useMemo(() => {
		return data.items
			.slice()
			.sort(
				(a, b) => new Date(b.start).getTime() - new Date(a.start).getTime(),
			);
	}, [data.items]);

	const formatDuration = React.useCallback(
		(startISO: string, endISO: string) => {
			const start = new Date(startISO).getTime();
			const end = new Date(endISO).getTime();
			const ms = Math.max(0, end - start);
			const totalMinutes = Math.round(ms / 60000);
			const days = Math.floor(totalMinutes / (24 * 60));
			const hours = Math.floor((totalMinutes % (24 * 60)) / 60);
			const minutes = totalMinutes % 60;
			if (days > 0 && hours === 0 && minutes === 0) return `${days}d`;
			if (days > 0 && hours === 0) return `${days}d ${minutes}m`;
			if (days > 0 && minutes === 0) return `${days}d ${hours}h`;
			if (days > 0) return `${days}d ${hours}h ${minutes}m`;
			if (hours > 0 && minutes === 0) return `${hours}h`;
			if (hours > 0) return `${hours}h ${minutes}m`;
			return `${minutes}m`;
		},
		[],
	);

	if (rows.length === 0) return null;

	return (
		<Table.Root size="sm" interactive>
			<Table.Header>
				<Table.Row>
					<Table.ColumnHeader>Service</Table.ColumnHeader>
					<Table.ColumnHeader>Metric</Table.ColumnHeader>
					<Table.ColumnHeader>Start</Table.ColumnHeader>
					<Table.ColumnHeader>End</Table.ColumnHeader>
					<Table.ColumnHeader>Duration</Table.ColumnHeader>
					<Table.ColumnHeader textAlign="end">Expected</Table.ColumnHeader>
					<Table.ColumnHeader textAlign="end">Real</Table.ColumnHeader>
					<Table.ColumnHeader textAlign="end">Diff</Table.ColumnHeader>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{rows.map((r) => {
					const diff = r.real_cost - r.expected_cost;
					const pct = r.expected_cost > 0 ? (diff / r.expected_cost) * 100 : 0;
					const startNum = new Date(r.start).getTime();
					const fallbackEndISO = data.to_date;
					const effectiveEndISO = r.end ?? fallbackEndISO;
					const endNum = new Date(effectiveEndISO).getTime();
					return (
						<Table.Row
							key={`${r.service}-${r.metric}-${r.start}-${r.end}`}
							bg={!r.end ? "bg.error" : undefined}
							onMouseEnter={() => onHover?.({ start: startNum, end: endNum })}
							onMouseLeave={() => onHover?.(null)}
						>
							<Table.Cell>{r.service}</Table.Cell>
							<Table.Cell>{r.metric}</Table.Cell>
							<Table.Cell>{formatDateTime(r.start)}</Table.Cell>
							<Table.Cell>
								{!r.end ? "ongoing" : formatDateTime(r.end)}
							</Table.Cell>
							<Table.Cell>
								{formatDuration(r.start, effectiveEndISO)}
							</Table.Cell>
							<Table.Cell textAlign="end">
								{formatCurrency(r.expected_cost)}
							</Table.Cell>
							<Table.Cell textAlign="end">
								{formatCurrency(r.real_cost)}
							</Table.Cell>
							<Table.Cell textAlign="end">
								{formatCurrency(diff)}{" "}
								{` ( ${formatNumber(pct, { maximumFractionDigits: 0 })}% )`}
							</Table.Cell>
						</Table.Row>
					);
				})}
			</Table.Body>
		</Table.Root>
	);
}

export const AlertWindowsTable = React.memo(AlertWindowsTableComponent);
