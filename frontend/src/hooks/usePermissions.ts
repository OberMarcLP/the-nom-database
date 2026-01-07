import { useContext } from 'react';
import { AuthContext } from '../contexts/AuthContext';
import { Role } from '../services/api';

export const usePermissions = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('usePermissions must be used within an AuthProvider');
  }
  const { user } = context;

  const hasPermission = (permission: string): boolean => {
    if (!user) return false;
    return user.permissions?.includes(permission) ?? false;
  };

  const hasAnyPermission = (permissions: string[]): boolean => {
    if (!user) return false;
    return permissions.some(perm => user.permissions?.includes(perm));
  };

  const hasAllPermissions = (permissions: string[]): boolean => {
    if (!user) return false;
    return permissions.every(perm => user.permissions?.includes(perm));
  };

  const hasRole = (roleName: string): boolean => {
    if (!user) return false;
    return user.roles?.some((role: Role) => role.name === roleName) ?? false;
  };

  const isAdmin = (): boolean => {
    return hasRole('admin');
  };

  const isModerator = (): boolean => {
    return hasRole('moderator') || hasRole('admin');
  };

  return {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    hasRole,
    isAdmin,
    isModerator,
    user,
  };
};
