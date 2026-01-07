import { useState, useEffect } from 'react';
import { Activity, MessageSquare, MapPin, Star } from 'lucide-react';
import { getActivityFeed, ActivityItem } from '../services/api';
import { Avatar } from './Avatar';

interface ActivityFeedProps {
  onRestaurantClick?: (restaurantId: number) => void;
}

export function ActivityFeed({ onRestaurantClick }: ActivityFeedProps) {
  const [activities, setActivities] = useState<ActivityItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadActivities();
  }, []);

  const loadActivities = async () => {
    setLoading(true);
    try {
      const data = await getActivityFeed(10);
      setActivities(data);
    } catch (error) {
      console.error('Failed to load activity feed:', error);
      setActivities([]);
    } finally {
      setLoading(false);
    }
  };

  const formatTimeAgo = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (seconds < 60) return 'just now';
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
    return date.toLocaleDateString();
  };

  const getAverageRating = (review: { food_rating: number; service_rating: number; ambiance_rating: number }) => {
    return ((review.food_rating + review.service_rating + review.ambiance_rating) / 3).toFixed(1);
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="card-glass p-4 animate-pulse">
            <div className="flex items-start gap-3">
              <div className="w-10 h-10 bg-gray-200 dark:bg-gray-700 rounded-full"></div>
              <div className="flex-1">
                <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2"></div>
                <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (activities.length === 0) {
    return (
      <div className="card-glass p-12 text-center">
        <Activity className="w-12 h-12 text-gray-400 mx-auto mb-3" />
        <p className="text-gray-600 dark:text-gray-400">No recent activity</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {activities.map((activity, index) => (
        <div
          key={`${activity.type}-${index}`}
          className="card-glass p-4 hover:shadow-xl transition-all duration-300 cursor-pointer"
          onClick={() => activity.restaurant && onRestaurantClick?.(activity.restaurant.id)}
        >
          <div className="flex items-start gap-3">
            {/* User Avatar */}
            {activity.user && (
              <Avatar
                src={activity.user.avatar_url}
                alt={activity.user.username}
                size="md"
                fallbackText={activity.user.full_name || activity.user.username}
              />
            )}

            {/* Activity Content */}
            <div className="flex-1 min-w-0">
              {activity.type === 'review' && activity.review ? (
                <>
                  {/* Review Activity */}
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-gray-900 dark:text-white">
                        <span className="font-semibold">
                          {activity.user?.full_name || activity.user?.username || 'Someone'}
                        </span>
                        {' '}reviewed{' '}
                        <span className="font-semibold text-blue-600 dark:text-blue-400">
                          {activity.restaurant?.name}
                        </span>
                      </p>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        {formatTimeAgo(activity.timestamp)}
                      </p>
                    </div>
                    <div className="flex items-center gap-1 bg-yellow-400/20 dark:bg-yellow-500/20 px-2 py-1 rounded-full shrink-0">
                      <Star className="w-3 h-3 fill-yellow-400 text-yellow-400" />
                      <span className="text-xs font-semibold text-yellow-700 dark:text-yellow-300">
                        {getAverageRating(activity.review)}
                      </span>
                    </div>
                  </div>

                  {activity.review.comment && (
                    <div className="flex items-start gap-2 mt-2 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                      <MessageSquare className="w-4 h-4 text-gray-400 mt-0.5 shrink-0" />
                      <p className="text-sm text-gray-700 dark:text-gray-300 line-clamp-2">
                        {activity.review.comment}
                      </p>
                    </div>
                  )}

                  {activity.restaurant?.address && (
                    <div className="flex items-center gap-1 mt-2 text-xs text-gray-500 dark:text-gray-400">
                      <MapPin className="w-3 h-3" />
                      <span className="truncate">{activity.restaurant.address}</span>
                    </div>
                  )}
                </>
              ) : (
                <>
                  {/* Restaurant Activity */}
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-gray-900 dark:text-white">
                        <span className="font-semibold">
                          {activity.user?.full_name || activity.user?.username || 'Someone'}
                        </span>
                        {' '}added a new restaurant{' '}
                        <span className="font-semibold text-green-600 dark:text-green-400">
                          {activity.restaurant?.name}
                        </span>
                      </p>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        {formatTimeAgo(activity.timestamp)}
                      </p>
                    </div>
                  </div>

                  {activity.restaurant?.address && (
                    <div className="flex items-center gap-1 mt-2 text-xs text-gray-500 dark:text-gray-400">
                      <MapPin className="w-3 h-3" />
                      <span className="truncate">{activity.restaurant.address}</span>
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
