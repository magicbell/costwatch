import { TanstackDevtools } from "@tanstack/react-devtools";
import { ReactQueryDevtoolsPanel } from "@tanstack/react-query-devtools";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";

export function DevTools() {
	return (
		<TanstackDevtools
			config={{
				position: "bottom-left",
			}}
			plugins={[
				{
					name: "Tanstack Router",
					render: <TanStackRouterDevtoolsPanel />,
				},
				{
					name: "Tanstack Query",
					render: <ReactQueryDevtoolsPanel />,
				},
			]}
		/>
	);
}
