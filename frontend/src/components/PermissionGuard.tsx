import { ReactNode } from 'react';
import { usePermissions } from '../hooks/usePermissions';

interface PermissionGuardProps {
  children: ReactNode;
  permission?: string;
  permissions?: string[];
  requireAll?: boolean;
  role?: string;
  fallback?: ReactNode;
}

export const PermissionGuard = ({
  children,
  permission,
  permissions,
  requireAll = false,
  role,
  fallback = null,
}: PermissionGuardProps) => {
  const { hasPermission, hasAnyPermission, hasAllPermissions, hasRole } = usePermissions();

  let hasAccess = false;

  if (role) {
    hasAccess = hasRole(role);
  } else if (permission) {
    hasAccess = hasPermission(permission);
  } else if (permissions) {
    hasAccess = requireAll
      ? hasAllPermissions(permissions)
      : hasAnyPermission(permissions);
  }

  return hasAccess ? <>{children}</> : <>{fallback}</>;
};
