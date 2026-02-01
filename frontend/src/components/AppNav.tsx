import { NavLink, useNavigate } from 'react-router-dom';
import { Home, LogIn, LogOut, UserPlus, Bookmark, Shield } from 'lucide-react';
import { GlobalSearch } from './GlobalSearch';
import { Avatar } from './Avatar';
import { useAuth } from '../contexts/AuthContext';
import { usePermissions } from '../hooks/usePermissions';
import type { Category, FoodType } from '../services/api';
import type { RestaurantFilters } from '../services/api';

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

export type AppNavVariant = 'admin' | 'main';

const navLinkBase =
  'flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300';
const navLinkActive =
  'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20';
const navLinkInactive =
  'hover:bg-white/20 dark:hover:bg-white/10 hover:backdrop-blur-md hover:shadow-md';
const adminNavLinkActive =
  'bg-gradient-to-r from-red-500/20 to-orange-500/20 backdrop-blur-md border border-red-500/30 shadow-lg shadow-red-500/20';
const mobileNavBase =
  'flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 whitespace-nowrap';

interface AppNavProps {
  variant: AppNavVariant;
  categories: Category[];
  foodTypes: FoodType[];
  filters: RestaurantFilters;
  onFiltersChange: (f: RestaurantFilters) => void;
  onLoginClick: () => void;
  onRegisterClick: () => void;
}

export function AppNav({
  variant: _variant,
  categories,
  foodTypes,
  filters,
  onFiltersChange,
  onLoginClick,
  onRegisterClick,
}: AppNavProps) {
  const { user } = useAuth();
  const { isAdmin } = usePermissions();

  return (
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
                    `${navLinkBase} ${isActive ? navLinkActive : navLinkInactive}`
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
                        `${navLinkBase} ${isActive ? navLinkActive : navLinkInactive}`
                      }
                    >
                      <Bookmark className="w-4 h-4" />
                      <span>My Lists</span>
                    </NavLink>
                    {isAdmin() && (
                      <NavLink
                        to="/admin"
                        className={({ isActive }) =>
                          `${navLinkBase} ${isActive ? adminNavLinkActive : navLinkInactive}`
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
              <UserMenu onLoginClick={onLoginClick} onRegisterClick={onRegisterClick} />
            </div>
          </div>

          <div className="w-full">
            <GlobalSearch
              categories={categories}
              foodTypes={foodTypes}
              filters={filters}
              onFiltersChange={onFiltersChange}
            />
          </div>

          <div className="flex lg:hidden gap-2 overflow-x-auto pb-1">
            <NavLink
              to="/"
              className={({ isActive }) =>
                `${mobileNavBase} ${isActive ? navLinkActive : navLinkInactive}`
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
                    `${mobileNavBase} ${isActive ? navLinkActive : navLinkInactive}`
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
  );
}
