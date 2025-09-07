import type { QueryClient } from "@tanstack/react-query";
import { Outlet, createRootRouteWithContext } from "@tanstack/react-router";
import * as React from "react";

import { Layout } from "@/components/layout";

const DevTools = React.lazy(() =>
	import("../integrations/tanstack-query/devtools").then((d) => ({
		default: d.DevTools,
	})),
);

interface MyRouterContext {
	queryClient: QueryClient;
}

function Root() {
	const [showDevtools, setShowDevtools] = React.useState(false);

	React.useEffect(() => {
		// @ts-ignore
		window.toggleDevtools = () => setShowDevtools((old) => !old);
	}, []);

	return (
		<>
			<Layout>
				<Outlet />
			</Layout>

			{showDevtools ? (
				<React.Suspense fallback={null}>
					<DevTools />
				</React.Suspense>
			) : null}
		</>
	);
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
	component: Root,
});
