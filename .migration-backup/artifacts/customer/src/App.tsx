import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from '@/components/ui/toaster';
import { Toaster as SonnerToaster } from '@/components/ui/sonner';
import { Router as WouterRouter } from 'wouter';
import CustomerPage from '@/app/page';

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <WouterRouter base={import.meta.env.BASE_URL.replace(/\/$/, '')}>
        <CustomerPage />
      </WouterRouter>
      <SonnerToaster position="top-center" />
      <Toaster />
    </QueryClientProvider>
  );
}
