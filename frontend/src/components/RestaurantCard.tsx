import { MapPin, Utensils, Tag, Lightbulb, CheckCircle, XCircle } from 'lucide-react';
import { Restaurant } from '../services/api';
import { StarRating } from './StarRating';
import { UserBadge } from './UserBadge';

interface RestaurantCardProps {
  restaurant: Restaurant;
  onClick: () => void;
  onReview?: (restaurant: Restaurant) => void;
  onReject?: (restaurant: Restaurant) => void;
}

export function RestaurantCard({ restaurant, onClick, onReview, onReject }: RestaurantCardProps) {
  return (
    <div
      className="card-glass hover:shadow-2xl hover:scale-105 transition-all duration-300 group overflow-hidden relative p-6 cursor-pointer"
    >
      <div onClick={onClick}>
        <div className="flex items-start justify-between gap-2 mb-2">
          <h3 className="text-xl font-semibold flex-1">{restaurant.name}</h3>
          {restaurant.is_suggestion && (
            <span className="badge-suggestion animate-pulse shrink-0">
              <Lightbulb className="w-3 h-3" />
              Suggestion
            </span>
          )}
        </div>

      {restaurant.description && (
        <p className="text-muted text-sm mb-3 line-clamp-2">
          {restaurant.description}
        </p>
      )}

      <div className="flex flex-wrap gap-2 mb-3">
        {restaurant.category && (
          <span className="badge-category">
            <Tag className="w-3 h-3" />
            {restaurant.category.name}
          </span>
        )}
        {restaurant.food_types?.map((ft) => (
          <span key={ft.id} className="badge-food-type">
            <Utensils className="w-3 h-3" />
            {ft.name}
          </span>
        ))}
      </div>

      {(restaurant.address || restaurant.distance !== undefined) && (
        <div className="flex items-start gap-2 text-sm text-muted mb-3">
          <MapPin className="w-4 h-4 mt-0.5 shrink-0" />
          <div className="flex-1">
            {restaurant.address && <span className="line-clamp-2">{restaurant.address}</span>}
            {restaurant.distance !== undefined && (
              <span className="text-info font-medium">
                {restaurant.address ? ' · ' : ''}{restaurant.distance.toFixed(1)} km away
              </span>
            )}
          </div>
        </div>
      )}

        {!restaurant.is_suggestion && restaurant.avg_rating && (
          <div className="flex items-center gap-2 pt-3 border-t border-default">
            <StarRating rating={Math.round(restaurant.avg_rating.overall)} readonly size="sm" />
            <span className="text-sm text-muted">
              {restaurant.avg_rating.overall.toFixed(1)} ({restaurant.avg_rating.count} reviews)
            </span>
          </div>
        )}

        {restaurant.created_by_user && (
          <div className="flex items-center gap-2 pt-2 mt-2 border-t border-default">
            <span className="text-xs text-muted">Created by</span>
            <UserBadge user={restaurant.created_by_user} size="sm" />
          </div>
        )}
      </div>

      {restaurant.is_suggestion && onReview && onReject && (
        <div className="flex gap-2 mt-3 pt-3 border-t border-default">
          <button
            onClick={(e) => {
              e.stopPropagation();
              onReview(restaurant);
            }}
            className="btn-glass-success flex-1 flex items-center justify-center gap-2"
          >
            <CheckCircle className="w-4 h-4" />
            Review
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onReject(restaurant);
            }}
            className="btn-glass-danger flex-1 flex items-center justify-center gap-2"
          >
            <XCircle className="w-4 h-4" />
            Reject
          </button>
        </div>
      )}
    </div>
  );
}
