import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from '@/components/ui/toaster';
import { Toaster as SonnerToaster } from '@/components/ui/sonner';
import { Route, Switch, Router as WouterRouter } from 'wouter';
import NotFound from '@/pages/not-found';

import LoginPage from '@/app/login/page';
import AgentLayoutWrapper from '@/app/(agent)/layout';
import AgentDefaultPage from '@/app/(agent)/page';
import InboxPage from '@/app/(agent)/inbox/page';
import SearchPage from '@/app/(agent)/search/page';
import TicketPage from '@/app/(agent)/tickets/[id]/page';
import OrderDetailPage from '@/app/(agent)/orders/[id]/page';
import DriverDetailPage from '@/app/(agent)/drivers/[id]/page';

const queryClient = new QueryClient();

function AgentRoutes() {
  return (
    <AgentLayoutWrapper>
      <Switch>
        <Route path="/" component={AgentDefaultPage} />
        <Route path="/inbox" component={InboxPage} />
        <Route path="/search" component={SearchPage} />
        <Route path="/tickets/:id" component={TicketPage} />
        <Route path="/orders/:id" component={OrderDetailPage} />
        <Route path="/drivers/:id" component={DriverDetailPage} />
        <Route component={NotFound} />
      </Switch>
    </AgentLayoutWrapper>
  );
}

function Router() {
  return (
    <Switch>
      <Route path="/login" component={LoginPage} />
      <Route path="/:rest*" component={AgentRoutes} />
    </Switch>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <WouterRouter base={import.meta.env.BASE_URL.replace(/\/$/, '')}>
        <Router />
      </WouterRouter>
      <SonnerToaster />
      <Toaster />
    </QueryClientProvider>
  );
}
