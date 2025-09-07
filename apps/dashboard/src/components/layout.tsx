import { Box } from "@chakra-ui/react";
import type { PropsWithChildren } from "react";

import { Navbar } from "@/components/navbar";

export function Layout({ children }: PropsWithChildren) {
	return (
		<Box w="full" minH="100dvh" display="flex" flexDir="column">
			<Navbar />
			<Box as="main" w="full" flex="1">
				{children}
			</Box>
		</Box>
	);
}
