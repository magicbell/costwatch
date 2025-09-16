import { Flex, HStack, IconButton, Image, Link, Spacer } from '@chakra-ui/react';

import { VscGithubInverted } from 'react-icons/vsc';
import { ColorModeButton } from '@/components/ui/color-mode.tsx';

function GithubButton() {
  return (
    <IconButton
      variant="ghost"
      aria-label="Toggle color mode"
      size="sm"
      css={{
        _icon: {
          width: "5",
          height: "5",
        },
      }}
      asChild
    >
      <Link href="https://github.com/magicbell/costwatch" target="_blank">
        <VscGithubInverted />
      </Link>
    </IconButton>
  )
}

export function Navbar() {
	return (
		<Flex
			as="header"
			w="full"
			align="center"
			px={6}
			py={4}
			gap={4}
			borderBottomWidth="1px"
		>
			<HStack gap={3} align="center">
				<Image src="/logo.svg" alt="logo" h={8} />
			</HStack>
			<Spacer />
			<HStack gap={2} align="center">
        <GithubButton />
				<ColorModeButton />
			</HStack>
		</Flex>
	);
}
