import { useState, lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, NavLink, useNavigate } from 'react-router-dom';
import { Home, Loader2, LogIn, LogOut, UserPlus, Bookmark, Shield } from 'lucide-react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { usePermissions } from './hooks/usePermissions';
import { GlobalSearch } from './components/GlobalSearch';
import { useCategories, useFoodTypes } from './hooks/useApi';
import { RestaurantFilters } from './services/api';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ToastProvider } from './hooks/useToast';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import { AuthModal } from './components/AuthModal';
import { UserProfileModal } from './components/UserProfileModal';
import { Avatar } from './components/Avatar';

// Lazy load page components for code splitting
const HomePage = lazy(() => import('./pages/HomePage').then(m => ({ default: m.HomePage })));
const ChangePasswordPage = lazy(() => import('./pages/ChangePasswordPage'));
const ListsPage = lazy(() => import('./pages/ListsPage').then(m => ({ default: m.ListsPage })));
const UserProfilePage = lazy(() => import('./pages/UserProfilePage'));

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
    <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
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

// User menu component
interface UserMenuProps {
  onLoginClick: () => void;
  onRegisterClick: () => void;
}

function UserMenu({ onLoginClick, onRegisterClick }: UserMenuProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

  const handleProfileClick = () => {
    if (user) {
      navigate(`/users/${user.id}`);
    }
  };

  if (!user) {
    return (
      <div className="flex items-center gap-2">
        <button
          onClick={onLoginClick}
          className="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md"
        >
          <LogIn className="w-4 h-4" />
          <span className="text-sm">Login</span>
        </button>
        <button
          onClick={onRegisterClick}
          className="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 bg-blue-500/20 hover:bg-blue-500/30 backdrop-blur-md"
        >
          <UserPlus className="w-4 h-4" />
          <span className="text-sm">Register</span>
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <button
        onClick={handleProfileClick}
        className="flex items-center gap-2 px-2 py-1.5 rounded-full transition-all duration-300 hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md"
        title="View profile"
      >
        <Avatar
          src={user.avatar_url}
          alt={user.username}
          size="sm"
          fallbackText={user.full_name || user.username}
        />
        <span className="text-sm text-gray-700 dark:text-gray-300 hidden sm:inline">
          {user.username}
        </span>
      </button>
      <button
        onClick={handleLogout}
        className="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md"
        title="Logout"
      >
        <LogOut className="w-4 h-4" />
        <span className="text-sm hidden sm:inline">Logout</span>
      </button>
    </div>
  );
}

function AppContent() {
  const { user } = useAuth();
  const { isAdmin } = usePermissions();
  const [filters, setFilters] = useState<RestaurantFilters>({});
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [authModalMode, setAuthModalMode] = useState<'login' | 'register'>('login');
  const [profileModalOpen, setProfileModalOpen] = useState(false);

  // Use React Query hooks instead of manual fetching
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
              <nav className="nav-glass sticky top-0 z-40">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                  <div className="flex flex-col gap-3 py-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-8">
                        <span className="text-xl font-bold text-gradient">
                          The Nom Database
                        </span>
                        <div className="hidden lg:flex gap-2">
                          <NavLink
                            to="/"
                            className={({ isActive }) =>
                              `flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300 ${
                                isActive
                                  ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                                  : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                              }`
                            }
                          >
                            <Home className="w-4 h-4" />
                            <span>Restaurants</span>
                          </NavLink>
                          {user && (
                            <>
                              <NavLink
                                to="/lists"
                                className={({ isActive }) =>
                                  `flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300 ${
                                    isActive
                                      ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                                      : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                                  }`
                                }
                              >
                                <Bookmark className="w-4 h-4" />
                                <span>My Lists</span>
                              </NavLink>
                              {isAdmin() && (
                                <NavLink
                                  to="/admin"
                                  className={({ isActive }) =>
                                    `flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300 ${
                                      isActive
                                        ? 'bg-gradient-to-r from-red-500/20 to-orange-500/20 backdrop-blur-md border border-red-500/30 shadow-lg shadow-red-500/20'
                                        : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                                    }`
                                  }
                                >
                                  <Shield className="w-4 h-4" />
                                  <span>Admin</span>
                                </NavLink>
                              )}
                            </>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <UserMenu onLoginClick={handleLoginClick} onRegisterClick={handleRegisterClick} />
                        
                      </div>
                    </div>

                    {/* Global Search Bar */}
                    <div className="w-full">
                      <GlobalSearch
                        categories={categories}
                        foodTypes={foodTypes}
                        filters={filters}
                        onFiltersChange={setFilters}
                      />
                    </div>

                    {/* Mobile Navigation */}
                    <div className="flex lg:hidden gap-2 overflow-x-auto pb-1">
                      <NavLink
                        to="/"
                        className={({ isActive }) =>
                          `flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 whitespace-nowrap ${
                            isActive
                              ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                              : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                          }`
                        }
                      >
                        <Home className="w-4 h-4" />
                        <span className="text-sm">Restaurants</span>
                      </NavLink>
                      {user && (
                        <>
                          <NavLink
                            to="/lists"
                            className={({ isActive }) =>
                              `flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 whitespace-nowrap ${
                                isActive
                                  ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                                  : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                              }`
                            }
                          >
                            <Bookmark className="w-4 h-4" />
                            <span className="text-sm">My Lists</span>
                          </NavLink>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              </nav>

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
              <nav className="nav-glass sticky top-0 z-40">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                  <div className="flex flex-col gap-3 py-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-8">
                        <span className="text-xl font-bold text-gradient">
                          The Nom Database
                        </span>
                        <div className="hidden lg:flex gap-2">
                          <NavLink
                            to="/"
                            className={({ isActive }) =>
                              `flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300 ${
                                isActive
                                  ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                                  : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                              }`
                            }
                          >
                            <Home className="w-4 h-4" />
                            <span>Restaurants</span>
                          </NavLink>
                          {user && (
                            <>
                              <NavLink
                                to="/lists"
                                className={({ isActive }) =>
                                  `flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300 ${
                                    isActive
                                      ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                                      : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                                  }`
                                }
                              >
                                <Bookmark className="w-4 h-4" />
                                <span>My Lists</span>
                              </NavLink>
                              {isAdmin() && (
                                <NavLink
                                  to="/admin"
                                  className={({ isActive }) =>
                                    `flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300 ${
                                      isActive
                                        ? 'bg-gradient-to-r from-red-500/20 to-orange-500/20 backdrop-blur-md border border-red-500/30 shadow-lg shadow-red-500/20'
                                        : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                                    }`
                                  }
                                >
                                  <Shield className="w-4 h-4" />
                                  <span>Admin</span>
                                </NavLink>
                              )}
                            </>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <UserMenu onLoginClick={handleLoginClick} onRegisterClick={handleRegisterClick} />
                        
                      </div>
                    </div>

                    {/* Global Search Bar */}
                    <div className="w-full">
                      <GlobalSearch
                        categories={categories}
                        foodTypes={foodTypes}
                        filters={filters}
                        onFiltersChange={setFilters}
                      />
                    </div>

                    {/* Mobile Navigation */}
                    <div className="flex lg:hidden gap-2 overflow-x-auto pb-1">
                      <NavLink
                        to="/"
                        className={({ isActive }) =>
                          `flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 whitespace-nowrap ${
                            isActive
                              ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                              : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                          }`
                        }
                      >
                        <Home className="w-4 h-4" />
                        <span className="text-sm">Restaurants</span>
                      </NavLink>
                      {user && (
                        <>
                          <NavLink
                            to="/lists"
                            className={({ isActive }) =>
                              `flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 whitespace-nowrap ${
                                isActive
                                  ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                                  : 'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md'
                              }`
                            }
                          >
                            <Bookmark className="w-4 h-4" />
                            <span className="text-sm">My Lists</span>
                          </NavLink>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              </nav>

              <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <Routes>
                  <Route path="/" element={<HomePage key={user?.id || 'guest'} filters={filters} isAuthenticated={!!user} />} />
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
            <ReactQueryDevtools initialIsOpen={false} />
          </ToastProvider>
        </AuthProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;
