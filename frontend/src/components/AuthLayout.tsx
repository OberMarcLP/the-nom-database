import { ReactNode } from 'react';
import { useTheme } from '../hooks/useTheme';
import { ThemeToggle } from './ThemeToggle';

interface AuthLayoutProps {
  children: ReactNode;
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { isDark, toggleTheme } = useTheme();

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Minimal header with just theme toggle */}
      <div className="absolute top-4 right-4 z-10">
        <ThemeToggle isDark={isDark} onToggle={toggleTheme} />
      </div>

      {/* Auth page content */}
      {children}
    </div>
  );
}
