import { Star } from 'lucide-react';

interface StarRatingProps {
  rating: number;
  onRatingChange?: (rating: number) => void;
  readonly?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export function StarRating({
  rating,
  onRatingChange,
  readonly = false,
  size = 'md',
}: StarRatingProps) {
  const sizeClasses = {
    sm: 'w-4 h-4',
    md: 'w-5 h-5',
    lg: 'w-6 h-6',
  };

  return (
    <div className="flex gap-1">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          onClick={() => !readonly && onRatingChange?.(star)}
          disabled={readonly}
          className={`${readonly ? 'cursor-default' : 'cursor-pointer hover:scale-125'} transition-all duration-200`}
        >
          <Star
            className={`${sizeClasses[size]} transition-all duration-200 ${
              star <= rating
                ? 'fill-[var(--warning)] text-[var(--warning)]'
                : 'fill-none text-[var(--border)]'
            }`}
            style={star <= rating ? { filter: 'drop-shadow(0 0 8px var(--warning))' } : undefined}
          />
        </button>
      ))}
    </div>
  );
}
