import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

export function getContext() {
  return {
    queryClient: new QueryClient(),
  };
}

export function Provider({ children, queryClient }: { children: React.ReactNode; queryClient: QueryClient }) {
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
