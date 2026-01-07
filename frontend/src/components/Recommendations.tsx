import { useState, useEffect } from 'react';
import { TrendingUp, Clock, Sparkles, ChevronRight } from 'lucide-react';
import { getPopularRestaurants, getRecentRestaurants, getPersonalizedRecommendations, Restaurant } from '../services/api';
import { RestaurantCard } from './RestaurantCard';
import { useAuth } from '../contexts/AuthContext';

interface RecommendationsProps {
  onRestaurantClick?: (restaurant: Restaurant) => void;
}

type TabType = 'popular' | 'recent' | 'personalized';

export function Recommendations({ onRestaurantClick }: RecommendationsProps) {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<TabType>('popular');
  const [restaurants, setRestaurants] = useState<Restaurant[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadRecommendations();
  }, [activeTab, user]);

  const loadRecommendations = async () => {
    setLoading(true);
    try {
      let data: Restaurant[];
      switch (activeTab) {
        case 'popular':
          data = await getPopularRestaurants(6, 3);
          break;
        case 'recent':
          data = await getRecentRestaurants(6);
          break;
        case 'personalized':
          if (!user) {
            data = [];
          } else {
            data = await getPersonalizedRecommendations(6);
          }
          break;
        default:
          data = [];
      }
      setRestaurants(data);
    } catch (error) {
      console.error('Failed to load recommendations:', error);
      setRestaurants([]);
    } finally {
      setLoading(false);
    }
  };

  const tabs = [
    { id: 'popular' as TabType, label: 'Popular', icon: TrendingUp },
    { id: 'recent' as TabType, label: 'Recently Added', icon: Clock },
    ...(user ? [{ id: 'personalized' as TabType, label: 'For You', icon: Sparkles }] : []),
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          Discover Restaurants
        </h2>
      </div>

      {/* Tabs */}
      <div className="flex flex-wrap gap-2">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all duration-300 ${
                activeTab === tab.id
                  ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                  : 'btn-glass'
              }`}
            >
              <Icon className="w-4 h-4" />
              <span>{tab.label}</span>
            </button>
          );
        })}
      </div>

      {/* Content */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="card-glass p-6 animate-pulse">
              <div className="h-32 bg-gray-200 dark:bg-gray-700 rounded-lg mb-4"></div>
              <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2"></div>
              <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
            </div>
          ))}
        </div>
      ) : restaurants.length === 0 ? (
        <div className="card-glass p-12 text-center">
          {activeTab === 'personalized' ? (
            <>
              <Sparkles className="w-12 h-12 text-gray-400 mx-auto mb-3" />
              <p className="text-gray-600 dark:text-gray-400">
                Rate more restaurants to get personalized recommendations!
              </p>
            </>
          ) : (
            <>
              <TrendingUp className="w-12 h-12 text-gray-400 mx-auto mb-3" />
              <p className="text-gray-600 dark:text-gray-400">
                No restaurants found
              </p>
            </>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {restaurants.map((restaurant) => (
            <RestaurantCard
              key={restaurant.id}
              restaurant={restaurant}
              onClick={() => onRestaurantClick?.(restaurant)}
            />
          ))}
        </div>
      )}

      {/* View All Link */}
      {restaurants.length > 0 && (
        <div className="flex justify-center">
          <button
            onClick={() => {
              // Scroll to top or navigate to main list
              window.scrollTo({ top: 0, behavior: 'smooth' });
            }}
            className="flex items-center gap-2 px-4 py-2 text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
          >
            <span>View all restaurants</span>
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      )}
    </div>
  );
}
