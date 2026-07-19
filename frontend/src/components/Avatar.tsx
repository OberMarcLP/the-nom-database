import { User } from 'lucide-react';

interface AvatarProps {
  src?: string | null;
  alt?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  fallbackText?: string;
  className?: string;
}

const sizeClasses = {
  sm: 'w-8 h-8 text-xs',
  md: 'w-10 h-10 text-sm',
  lg: 'w-12 h-12 text-base',
  xl: 'w-16 h-16 text-lg',
};

/**
 * Avatar component with Gravatar support and fallback to initials or icon
 *
 * Features:
 * - Displays user avatar from Gravatar URL
 * - Falls back to user initials if provided
 * - Falls back to user icon if no initials
 * - Responsive sizing
 * - Dark mode support
 *
 * @param src - Avatar image URL (typically from Gravatar)
 * @param alt - Alt text for the image
 * @param size - Size variant (sm, md, lg, xl)
 * @param fallbackText - Text to display as initials (e.g., "John Doe" -> "JD")
 * @param className - Additional CSS classes
 */
export function Avatar({ src, alt, size = 'md', fallbackText, className = '' }: AvatarProps) {
  const sizeClass = sizeClasses[size];

  // Generate initials from fallback text
  const getInitials = (text: string): string => {
    const words = text.trim().split(/\s+/);
    if (words.length === 1) {
      return words[0].charAt(0).toUpperCase();
    }
    return (words[0].charAt(0) + words[words.length - 1].charAt(0)).toUpperCase();
  };

  const initials = fallbackText ? getInitials(fallbackText) : null;

  return (
    <div
      className={`${sizeClass} rounded-full overflow-hidden flex items-center justify-center bg-linear-to-br from-(--accent) to-(--info) text-white font-semibold ${className}`}
    >
      {src ? (
        <img
          src={src}
          alt={alt || 'User avatar'}
          className="w-full h-full object-cover"
          onError={(e) => {
            // Hide image on error and show fallback
            e.currentTarget.style.display = 'none';
          }}
        />
      ) : initials ? (
        <span>{initials}</span>
      ) : (
        <User className="w-1/2 h-1/2" />
      )}
    </div>
  );
}
