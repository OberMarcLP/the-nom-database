import { createContext, useContext, useState, useEffect, useMemo, ReactNode } from 'react';
import {
  User,
  LoginRequest,
  RegisterRequest,
  LoginResponse,
  login as apiLogin,
  register as apiRegister,
  logout as apiLogout,
  getCurrentUser,
  restoreSession,
  setAccessToken,
  clearAccessToken,
} from '../services/api';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (credentials: LoginRequest) => Promise<LoginResponse>;
  register: (data: RegisterRequest) => Promise<LoginResponse>;
  logout: () => Promise<void>;
  updateUser: (user: User) => void;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // Restore the session on app start: the httpOnly refresh cookie is the
  // only persistent credential, so silently mint a fresh in-memory access
  // token first, then load the user. While this runs, `loading` is true.
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        const restored = await restoreSession();
        if (!restored || cancelled) return;
        const currentUser = await getCurrentUser();
        if (!cancelled) {
          setUser(currentUser);
        }
      } catch {
        // No valid session - stay logged out
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  const login = async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await apiLogin(credentials);
    // Refresh token lives in an httpOnly cookie set by the server;
    // the access token is held in memory only (never persisted).
    setAccessToken(response.access_token);
    localStorage.removeItem('refresh_token');
    setUser({ ...response.user });
    return response;
  };

  const register = async (data: RegisterRequest): Promise<LoginResponse> => {
    const response = await apiRegister(data);
    // Refresh token lives in an httpOnly cookie set by the server;
    // the access token is held in memory only (never persisted).
    setAccessToken(response.access_token);
    localStorage.removeItem('refresh_token');
    setUser({ ...response.user });
    return response;
  };

  const logout = async () => {
    try {
      await apiLogout();
    } catch (error) {
      // Continue with logout even if API call fails
      console.error('Logout error:', error);
    } finally {
      clearAccessToken();
      localStorage.removeItem('refresh_token');
      setUser(null);
    }
  };

  const updateUser = (updatedUser: User) => {
    setUser(updatedUser);
  };

  const value = useMemo(
    () => ({ user, loading, login, register, logout, updateUser }),
    [user, loading]
  );

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
