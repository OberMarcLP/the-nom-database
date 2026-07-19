import { Avatar } from './Avatar';
import type { UserSummary } from '../services/api';

interface UserBadgeProps {
  user: UserSummary;
  size?: 'sm' | 'md';
  showAvatar?: boolean;
  className?: string;
}

/**
 * UserBadge component displays user information with optional avatar
 *
 * Features:
 * - Shows user avatar (optional)
 * - Displays username with @ prefix
 * - Two sizes: sm (compact) and md (default)
 * - Dark mode support
 */
export function UserBadge({ user, size = 'md', showAvatar = true, className = '' }: UserBadgeProps) {
  const avatarSize = size === 'sm' ? 'sm' : 'md';
  const textSize = size === 'sm' ? 'text-xs' : 'text-sm';

  return (
    <div className={`inline-flex items-center gap-2 ${className}`}>
      {showAvatar && (
        <Avatar
          src={user.avatar_url}
          alt={user.username}
          size={avatarSize}
          fallbackText={user.full_name || user.username}
        />
      )}
      <span className={`${textSize} text-(--text)`}>
        @{user.username}
      </span>
    </div>
  );
}
