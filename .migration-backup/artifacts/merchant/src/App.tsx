import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from '@/components/ui/toaster';
import { Toaster as SonnerToaster } from '@/components/ui/sonner';
import { Route, Switch, Router as WouterRouter } from 'wouter';
import NotFound from '@/pages/not-found';

import LoginPage from '@/app/login/page';
import ChangePasswordPage from '@/app/change-password/page';
import MerchantLayoutWrapper from '@/app/(merchant)/layout';
import DashboardPage from '@/app/(merchant)/page';
import OrdersPage from '@/app/(merchant)/orders/page';
import MenuPage from '@/app/(merchant)/menu/page';
import HoursPage from '@/app/(merchant)/hours/page';

const queryClient = new QueryClient();

function MerchantRoutes() {
  return (
    <MerchantLayoutWrapper>
      <Switch>
        <Route path="/" component={DashboardPage} />
        <Route path="/orders" component={OrdersPage} />
        <Route path="/menu" component={MenuPage} />
        <Route path="/hours" component={HoursPage} />
        <Route component={NotFound} />
      </Switch>
    </MerchantLayoutWrapper>
  );
}

function Router() {
  return (
    <Switch>
      <Route path="/login" component={LoginPage} />
      <Route path="/change-password" component={ChangePasswordPage} />
      <Route path="/:rest*" component={MerchantRoutes} />
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
