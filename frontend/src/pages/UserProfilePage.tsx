import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { User, Star, MapPin, List, MessageSquare, Calendar, Edit2, ArrowLeft } from 'lucide-react';
import { getUserProfile, getUserReviews, UserProfile, Rating } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { StarRating } from '../components/StarRating';
import { EditProfileModal } from '../components/EditProfileModal';

type TabType = 'reviews' | 'lists' | 'about';

export default function UserProfilePage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [reviews, setReviews] = useState<Rating[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<TabType>('reviews');
  const [showEditModal, setShowEditModal] = useState(false);

  const isOwnProfile = currentUser?.id === Number(id);

  useEffect(() => {
    loadProfile();
  }, [id]);

  useEffect(() => {
    if (activeTab === 'reviews') {
      loadReviews();
    }
  }, [activeTab, id]);

  const loadProfile = async () => {
    try {
      setLoading(true);
      const data = await getUserProfile(Number(id));
      setProfile(data);
    } catch (error) {
      console.error('Failed to load profile:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadReviews = async () => {
    try {
      const data = await getUserReviews(Number(id));
      setReviews(data);
    } catch (error) {
      console.error('Failed to load reviews:', error);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600 dark:text-gray-400">Loading profile...</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">User not found</h2>
          <button
            onClick={() => navigate('/')}
            className="text-blue-500 hover:text-blue-600 transition-colors"
          >
            Go back home
          </button>
        </div>
      </div>
    );
  }

  const { user, stats } = profile;

  return (
    <div className="min-h-screen py-8">
      <div className="max-w-6xl mx-auto px-4">
        {/* Back button */}
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>

        {/* Profile Header */}
        <div className="card-glass p-8 mb-6 animate-slide-down">
          <div className="flex items-start gap-6">
            {/* Avatar */}
            <div className="flex-shrink-0">
              {user.avatar_url ? (
                <img
                  src={user.avatar_url}
                  alt={user.username}
                  className="w-24 h-24 rounded-full object-cover border-4 border-white/50 dark:border-white/20"
                />
              ) : (
                <div className="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500/20 to-purple-500/20 backdrop-blur-md border-4 border-white/50 dark:border-white/20 flex items-center justify-center">
                  <User className="w-12 h-12 text-blue-500" />
                </div>
              )}
            </div>

            {/* User Info */}
            <div className="flex-1">
              <div className="flex items-start justify-between">
                <div>
                  <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-1">
                    {user.full_name || user.username}
                  </h1>
                  <p className="text-gray-600 dark:text-gray-400">@{user.username}</p>
                  <div className="flex items-center gap-2 mt-2 text-sm text-gray-500 dark:text-gray-400">
                    <Calendar className="w-4 h-4" />
                    Joined {new Date(user.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
                  </div>
                </div>
                {isOwnProfile && (
                  <button
                    onClick={() => setShowEditModal(true)}
                    className="btn-glass flex items-center gap-2"
                  >
                    <Edit2 className="w-4 h-4" />
                    Edit Profile
                  </button>
                )}
              </div>

              {/* Stats Grid */}
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mt-6">
                <div className="bg-white/40 dark:bg-white/5 backdrop-blur-md rounded-xl p-4 border border-white/30 dark:border-white/10">
                  <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400 mb-1">
                    <MessageSquare className="w-4 h-4" />
                    <span className="text-sm">Reviews</span>
                  </div>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total_reviews}</p>
                </div>

                <div className="bg-white/40 dark:bg-white/5 backdrop-blur-md rounded-xl p-4 border border-white/30 dark:border-white/10">
                  <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400 mb-1">
                    <MapPin className="w-4 h-4" />
                    <span className="text-sm">Restaurants</span>
                  </div>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total_restaurants}</p>
                </div>

                <div className="bg-white/40 dark:bg-white/5 backdrop-blur-md rounded-xl p-4 border border-white/30 dark:border-white/10">
                  <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400 mb-1">
                    <List className="w-4 h-4" />
                    <span className="text-sm">Lists</span>
                  </div>
                  <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total_lists}</p>
                </div>

                <div className="bg-white/40 dark:bg-white/5 backdrop-blur-md rounded-xl p-4 border border-white/30 dark:border-white/10 col-span-2 md:col-span-3">
                  <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400 mb-2">
                    <Star className="w-4 h-4" />
                    <span className="text-sm">Average Ratings</span>
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Food</p>
                      <div className="flex items-center gap-2">
                        <StarRating rating={stats.avg_food_rating} readonly size="sm" />
                        <span className="text-sm font-semibold text-gray-900 dark:text-white">
                          {stats.avg_food_rating.toFixed(1)}
                        </span>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Service</p>
                      <div className="flex items-center gap-2">
                        <StarRating rating={stats.avg_service_rating} readonly size="sm" />
                        <span className="text-sm font-semibold text-gray-900 dark:text-white">
                          {stats.avg_service_rating.toFixed(1)}
                        </span>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Ambiance</p>
                      <div className="flex items-center gap-2">
                        <StarRating rating={stats.avg_ambiance_rating} readonly size="sm" />
                        <span className="text-sm font-semibold text-gray-900 dark:text-white">
                          {stats.avg_ambiance_rating.toFixed(1)}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="card-glass mb-6 animate-slide-down" style={{ animationDelay: '0.1s' }}>
          <div className="flex border-b border-white/20 dark:border-white/10">
            <button
              onClick={() => setActiveTab('reviews')}
              className={`px-6 py-3 font-medium transition-colors border-b-2 ${
                activeTab === 'reviews'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
              }`}
            >
              Reviews ({stats.total_reviews})
            </button>
            <button
              onClick={() => setActiveTab('lists')}
              className={`px-6 py-3 font-medium transition-colors border-b-2 ${
                activeTab === 'lists'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
              }`}
            >
              Lists ({stats.total_lists})
            </button>
            <button
              onClick={() => setActiveTab('about')}
              className={`px-6 py-3 font-medium transition-colors border-b-2 ${
                activeTab === 'about'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
              }`}
            >
              About
            </button>
          </div>
        </div>

        {/* Tab Content */}
        <div className="animate-slide-down" style={{ animationDelay: '0.2s' }}>
          {activeTab === 'reviews' && (
            <div className="space-y-4">
              {reviews.length === 0 ? (
                <div className="card-glass p-12 text-center">
                  <MessageSquare className="w-12 h-12 text-gray-400 mx-auto mb-3" />
                  <p className="text-gray-600 dark:text-gray-400">No reviews yet</p>
                </div>
              ) : (
                reviews.map((review) => (
                  <div key={review.id} className="card-glass p-6 hover:shadow-xl transition-shadow">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <h3
                          onClick={() => navigate(`/restaurants/${review.restaurant_id}`)}
                          className="text-lg font-semibold text-gray-900 dark:text-white mb-1 cursor-pointer hover:text-blue-500 transition-colors"
                        >
                          Restaurant #{review.restaurant_id}
                        </h3>
                        <div className="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
                          <span>{new Date(review.created_at).toLocaleDateString()}</span>
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-3 gap-4 mb-3">
                      <div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Food</p>
                        <StarRating rating={review.food_rating} readonly size="sm" />
                      </div>
                      <div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Service</p>
                        <StarRating rating={review.service_rating} readonly size="sm" />
                      </div>
                      <div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Ambiance</p>
                        <StarRating rating={review.ambiance_rating} readonly size="sm" />
                      </div>
                    </div>

                    {review.comment && (
                      <p className="text-gray-700 dark:text-gray-300 mb-3">{review.comment}</p>
                    )}

                    {review.photos && review.photos.length > 0 && (
                      <div className="flex flex-wrap gap-2">
                        {review.photos.map((photo) => (
                          <img
                            key={photo.id}
                            src={photo.photo_url}
                            alt={photo.caption || 'Review photo'}
                            className="w-20 h-20 object-cover rounded-lg"
                          />
                        ))}
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'lists' && (
            <div className="card-glass p-12 text-center">
              <List className="w-12 h-12 text-gray-400 mx-auto mb-3" />
              <p className="text-gray-600 dark:text-gray-400">Lists feature coming soon</p>
            </div>
          )}

          {activeTab === 'about' && (
            <div className="card-glass p-6">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">About</h3>
              <div className="space-y-3 text-gray-700 dark:text-gray-300">
                <div className="flex items-center gap-3">
                  <User className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Username</p>
                    <p className="font-medium">{user.username}</p>
                  </div>
                </div>
                {user.email && (
                  <div className="flex items-center gap-3">
                    <span className="text-xl">📧</span>
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Email</p>
                      <p className="font-medium">{user.email}</p>
                    </div>
                  </div>
                )}
                <div className="flex items-center gap-3">
                  <Calendar className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Member since</p>
                    <p className="font-medium">
                      {new Date(user.created_at).toLocaleDateString('en-US', {
                        month: 'long',
                        day: 'numeric',
                        year: 'numeric',
                      })}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Edit Profile Modal */}
      <EditProfileModal
        isOpen={showEditModal}
        onClose={() => setShowEditModal(false)}
        onSuccess={loadProfile}
      />
    </div>
  );
}
