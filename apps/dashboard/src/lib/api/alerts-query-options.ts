import { queryOptions } from "@tanstack/react-query";
import * as client from "./client";

export const alertWindowsQueryOptions = queryOptions({
	queryKey: ["alert-windows"] as const,
	queryFn: () => client.alertWindows().then((x) => x.data),
	refetchInterval: 30_000,
});

export const alertRulesQueryOptions = queryOptions({
	queryKey: ["alert-rules"] as const,
	queryFn: () => client.alertRules().then((x) => x.data),
	initialData: { items: [] },
	refetchInterval: 30_000,
});
