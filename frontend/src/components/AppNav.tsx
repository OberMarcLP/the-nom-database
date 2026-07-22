import { NavLink, useNavigate } from 'react-router-dom';
import { Home, LogIn, LogOut, UserPlus, Bookmark, Shield, Sun, Moon, Monitor } from 'lucide-react';
import { GlobalSearch } from './GlobalSearch';
import { Avatar } from './Avatar';
import { useAuth } from '../contexts/AuthContext';
import { usePermissions } from '../hooks/usePermissions';
import { useTheme, type ThemePreference } from '../hooks/useTheme';
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
          className="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 hover:bg-(--surface-hover)"
        >
          <LogIn className="w-4 h-4" />
          <span className="text-sm">Login</span>
        </button>
        <button
          onClick={onRegisterClick}
          className="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 bg-(--accent-dim) border border-(--accent) text-(--accent) hover:bg-(--accent) hover:text-(--surface)"
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
        className="flex items-center gap-2 px-2 py-1.5 rounded-full transition-all duration-300 hover:bg-(--surface-hover)"
        title="View profile"
      >
        <Avatar
          src={user.avatar_url}
          alt={user.username}
          size="sm"
          fallbackText={user.full_name || user.username}
        />
        <span className="text-sm text-(--text-muted) hidden sm:inline">
          {user.username}
        </span>
      </button>
      <button
        onClick={handleLogout}
        className="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300 hover:bg-(--surface-hover)"
        title="Logout"
      >
        <LogOut className="w-4 h-4" />
        <span className="text-sm hidden sm:inline">Logout</span>
      </button>
    </div>
  );
}

function ThemeToggle() {
  const { preference, setTheme } = useTheme();
  const next: Record<ThemePreference, ThemePreference> = {
    system: 'light',
    light: 'dark',
    dark: 'system',
  };
  const Icon = preference === 'system' ? Monitor : preference === 'light' ? Sun : Moon;
  const label =
    preference === 'system' ? 'Theme: System' : preference === 'light' ? 'Theme: Light' : 'Theme: Dark';

  return (
    <button
      onClick={() => setTheme(next[preference])}
      className="flex items-center justify-center w-9 h-9 rounded-full border border-(--border) text-(--text-muted) transition-all duration-300 hover:text-(--accent) hover:border-(--accent)"
      title={`${label} (click to switch)`}
      aria-label={`${label} (click to switch)`}
    >
      <Icon className="w-4 h-4" />
    </button>
  );
}

export type AppNavVariant = 'admin' | 'main';

const navLinkBase =
  'flex items-center gap-2 px-4 py-2 rounded-full transition-all duration-300';
const navLinkActive =
  'bg-(--accent-dim) border border-(--accent) text-(--accent)';
const navLinkInactive =
  'hover:bg-(--surface-hover)';
const adminNavLinkActive =
  'bg-(--danger-dim) border border-(--danger) text-(--danger)';
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
              <NavLink to="/" className="flex items-center gap-2.5">
                <img src="/logo-mark.svg" alt="" className="w-8 h-8" />
                <span className="text-xl font-bold text-gradient">
                  The Nom Database
                </span>
              </NavLink>
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
              <ThemeToggle />
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
