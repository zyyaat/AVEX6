import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from '@/components/ui/toaster';
import { Toaster as SonnerToaster } from '@/components/ui/sonner';
import { Route, Switch, Router as WouterRouter } from 'wouter';
import NotFound from '@/pages/not-found';

import LoginPage from '@/app/login/page';
import AdminLayoutWrapper from '@/app/(admin)/layout';
import DashboardPage from '@/app/(admin)/page';
import ApplicationsPage from '@/app/(admin)/applications/page';
import DriversPage from '@/app/(admin)/drivers/page';
import OrdersPage from '@/app/(admin)/orders/page';
import RestaurantsPage from '@/app/(admin)/restaurants/page';
import SettingsPage from '@/app/(admin)/settings/page';
import AdminSupportPage from '@/app/(admin)/support/page';
import TierPricesPage from '@/app/(admin)/tier-prices/page';
import TiersPage from '@/app/(admin)/tiers/page';
import ZonesPage from '@/app/(admin)/zones/page';

const queryClient = new QueryClient();

function AdminRoutes() {
  return (
    <AdminLayoutWrapper>
      <Switch>
        <Route path="/" component={DashboardPage} />
        <Route path="/applications" component={ApplicationsPage} />
        <Route path="/drivers" component={DriversPage} />
        <Route path="/orders" component={OrdersPage} />
        <Route path="/restaurants" component={RestaurantsPage} />
        <Route path="/settings" component={SettingsPage} />
        <Route path="/support" component={AdminSupportPage} />
        <Route path="/tier-prices" component={TierPricesPage} />
        <Route path="/tiers" component={TiersPage} />
        <Route path="/zones" component={ZonesPage} />
        <Route component={NotFound} />
      </Switch>
    </AdminLayoutWrapper>
  );
}

function Router() {
  return (
    <Switch>
      <Route path="/login" component={LoginPage} />
      <Route path="/" component={AdminRoutes} />
      <Route path="/:rest*" component={AdminRoutes} />
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
