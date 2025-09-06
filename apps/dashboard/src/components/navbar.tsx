import { Flex, HStack, Image, Spacer } from '@chakra-ui/react';

import { ColorModeButton } from '@/components/ui/color-mode';

export function Navbar() {
  return (
    <Flex as="header" w="full" align="center" px={6} py={4} gap={4} borderBottomWidth="1px">
      <HStack gap={3} align="center">
        <Image src="/logo.svg" alt="logo" h={8} />
      </HStack>
      <Spacer />
      <HStack gap={2} align="center">
        <ColorModeButton />
      </HStack>
    </Flex>
  );
}
