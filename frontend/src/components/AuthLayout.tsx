import { ReactNode } from 'react';

interface AuthLayoutProps {
  children: ReactNode;
}

export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="min-h-screen bg-(--surface-hover)">
      {/* Auth page content */}
      {children}
    </div>
  );
}
