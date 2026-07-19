import { useState, lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { useCategories, useFoodTypes } from './hooks/useApi';
import { RestaurantFilters } from './services/api';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ToastProvider } from './hooks/useToast';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import { AuthModal } from './components/AuthModal';
import { UserProfileModal } from './components/UserProfileModal';
import { AppNav } from './components/AppNav';

// Lazy load page components for code splitting
const HomePage = lazy(() => import('./pages/HomePage').then(m => ({ default: m.HomePage })));
const RestaurantPage = lazy(() => import('./pages/RestaurantPage').then(m => ({ default: m.RestaurantPage })));
const ChangePasswordPage = lazy(() => import('./pages/ChangePasswordPage'));
const ListsPage = lazy(() => import('./pages/ListsPage').then(m => ({ default: m.ListsPage })));
const UserProfilePage = lazy(() => import('./pages/UserProfilePage'));
const AuthCallbackPage = lazy(() => import('./pages/AuthCallbackPage').then(m => ({ default: m.AuthCallbackPage })));

// Admin pages
const AdminDashboard = lazy(() => import('./pages/AdminDashboard').then(m => ({ default: m.AdminDashboard })));
const AdminOverview = lazy(() => import('./pages/AdminOverview').then(m => ({ default: m.AdminOverview })));
const AdminUsers = lazy(() => import('./pages/AdminUsers').then(m => ({ default: m.AdminUsers })));
const AdminRoles = lazy(() => import('./pages/AdminRoles').then(m => ({ default: m.AdminRoles })));
const AdminRestaurants = lazy(() => import('./pages/AdminRestaurants').then(m => ({ default: m.AdminRestaurants })));
const AdminCategoriesFoodTypes = lazy(() => import('./pages/AdminCategoriesFoodTypes').then(m => ({ default: m.AdminCategoriesFoodTypes })));
const AdminContent = lazy(() => import('./pages/AdminContent').then(m => ({ default: m.AdminContent })));
const AdminAnalytics = lazy(() => import('./pages/AdminAnalytics').then(m => ({ default: m.AdminAnalytics })));
const AdminAudit = lazy(() => import('./pages/AdminAudit').then(m => ({ default: m.AdminAudit })));
const AdminSettings = lazy(() => import('./pages/AdminSettings').then(m => ({ default: m.AdminSettings })));

// Loading fallback component
const PageLoader = () => (
  <div className="flex items-center justify-center h-64">
    <Loader2 className="w-8 h-8 animate-spin text-(--info)" />
  </div>
);

// Create QueryClient instance with optimized defaults and error handling
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false, // Don't refetch on window focus
      retry: 1, // Retry failed requests once
      staleTime: 5 * 60 * 1000, // 5 minutes default
    },
    mutations: {
      retry: 0, // Don't retry mutations by default
    },
  },
});

function AppContent() {
  const { user } = useAuth();
  const [filters, setFilters] = useState<RestaurantFilters>({});
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [authModalMode, setAuthModalMode] = useState<'login' | 'register'>('login');
  const [profileModalOpen, setProfileModalOpen] = useState(false);

  const { data: categories = [] } = useCategories();
  const { data: foodTypes = [] } = useFoodTypes();

  const handleLoginClick = () => {
    setAuthModalMode('login');
    setAuthModalOpen(true);
  };

  const handleRegisterClick = () => {
    setAuthModalMode('register');
    setAuthModalOpen(true);
  };

  return (
    <BrowserRouter>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          {/* Admin routes - full width */}
          <Route path="/admin/*" element={
            <div className="min-h-screen">
              <AppNav
                variant="admin"
                categories={categories}
                foodTypes={foodTypes}
                filters={filters}
                onFiltersChange={setFilters}
                onLoginClick={handleLoginClick}
                onRegisterClick={handleRegisterClick}
              />

              <ProtectedRoute>
                <AdminDashboard />
              </ProtectedRoute>
            </div>
          }>
            <Route index element={<AdminOverview />} />
            <Route path="users" element={<AdminUsers />} />
            <Route path="roles" element={<AdminRoles />} />
            <Route path="restaurants" element={<AdminRestaurants />} />
            <Route path="categories" element={<AdminCategoriesFoodTypes />} />
            <Route path="content" element={<AdminContent />} />
            <Route path="analytics" element={<AdminAnalytics />} />
            <Route path="audit" element={<AdminAudit />} />
            <Route path="settings" element={<AdminSettings />} />
          </Route>

          {/* Main app routes - with navigation and max-width */}
          <Route path="/*" element={
            <div className="min-h-screen">
              <AppNav
                variant="main"
                categories={categories}
                foodTypes={foodTypes}
                filters={filters}
                onFiltersChange={setFilters}
                onLoginClick={handleLoginClick}
                onRegisterClick={handleRegisterClick}
              />

              <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <Routes>
                  <Route path="/" element={<HomePage key={user?.id ?? 'guest'} filters={filters} isAuthenticated={!!user} />} />
                  <Route path="/restaurants/:id" element={<RestaurantPage />} />
                  <Route path="/auth/callback" element={<AuthCallbackPage />} />
                  <Route
                    path="/lists"
                    element={
                      <ProtectedRoute>
                        <ListsPage />
                      </ProtectedRoute>
                    }
                  />
                  <Route path="/users/:id" element={<UserProfilePage />} />
                  <Route
                    path="/change-password"
                    element={
                      <ProtectedRoute>
                        <ChangePasswordPage />
                      </ProtectedRoute>
                    }
                  />
                </Routes>
              </main>
            </div>
          } />
        </Routes>
      </Suspense>

      {/* Modals */}
      <AuthModal
        isOpen={authModalOpen}
        onClose={() => setAuthModalOpen(false)}
        initialMode={authModalMode}
      />
      <UserProfileModal
        isOpen={profileModalOpen}
        onClose={() => setProfileModalOpen(false)}
      />
    </BrowserRouter>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
            <ToastProvider>
            <AppContent />
            {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
          </ToastProvider>
        </AuthProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;
