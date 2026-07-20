import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { User, Star, MapPin, List, MessageSquare, Calendar, Edit2, ArrowLeft, Lock, Globe } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { StarRating } from '../components/StarRating';
import { EditProfileModal } from '../components/EditProfileModal';
import { useUserLists, useUserProfile, useUserReviews } from '../hooks/useApi';

type TabType = 'reviews' | 'lists' | 'about';

export default function UserProfilePage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const [activeTab, setActiveTab] = useState<TabType>('reviews');
  const [showEditModal, setShowEditModal] = useState(false);

  const userId = Number(id);
  const isOwnProfile = currentUser?.id === userId;

  const {
    data: profile = null,
    isLoading: loading,
    refetch: refetchProfile,
  } = useUserProfile(userId);
  const { data: reviews = [] } = useUserReviews(userId, {
    enabled: userId > 0 && activeTab === 'reviews',
  });
  const { data: lists = [] } = useUserLists({
    // Only load lists for own profile for now
    enabled: isOwnProfile && activeTab === 'lists',
  });

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div style={{ color: 'var(--text-muted)' }}>Loading profile...</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-2" style={{ color: 'var(--text)' }}>User not found</h2>
          <button
            onClick={() => navigate('/')}
            className="transition-colors"
            style={{ color: 'var(--accent)' }}
            onMouseEnter={(e) => e.currentTarget.style.opacity = '0.8'}
            onMouseLeave={(e) => e.currentTarget.style.opacity = '1'}
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
          className="flex items-center gap-2 transition-colors mb-6"
          style={{ color: 'var(--text-muted)' }}
          onMouseEnter={(e) => e.currentTarget.style.color = 'var(--text)'}
          onMouseLeave={(e) => e.currentTarget.style.color = 'var(--text-muted)'}
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>

        {/* Profile Header */}
        <div className="card-glass p-8 mb-6 animate-slide-down">
          <div className="flex items-start gap-6">
            {/* Avatar */}
            <div className="shrink-0">
              {user.avatar_url ? (
                <img
                  src={user.avatar_url}
                  alt={user.username}
                  className="w-24 h-24 rounded-full object-cover border-4 border-white/50 dark:border-white/20"
                />
              ) : (
                <div className="w-24 h-24 rounded-full flex items-center justify-center border-4" style={{
                  background: 'var(--surface)',
                  borderColor: 'var(--border)'
                }}>
                  <User className="w-12 h-12" style={{ color: 'var(--accent)' }} />
                </div>
              )}
            </div>

            {/* User Info */}
            <div className="flex-1">
              <div className="flex items-start justify-between">
                <div>
                  <h1 className="text-3xl font-bold mb-1" style={{ color: 'var(--text)' }}>
                    {user.full_name || user.username}
                  </h1>
                  <p style={{ color: 'var(--text-muted)' }}>@{user.username}</p>
                  <div className="flex items-center gap-2 mt-2 text-sm" style={{ color: 'var(--text-muted)' }}>
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
                <div className="rounded-xl p-4 border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
                  <div className="flex items-center gap-2 mb-1" style={{ color: 'var(--text-muted)' }}>
                    <MessageSquare className="w-4 h-4" />
                    <span className="text-sm">Reviews</span>
                  </div>
                  <p className="text-2xl font-bold" style={{ color: 'var(--text)' }}>{stats.total_reviews}</p>
                </div>

                <div className="rounded-xl p-4 border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
                  <div className="flex items-center gap-2 mb-1" style={{ color: 'var(--text-muted)' }}>
                    <MapPin className="w-4 h-4" />
                    <span className="text-sm">Restaurants</span>
                  </div>
                  <p className="text-2xl font-bold" style={{ color: 'var(--text)' }}>{stats.total_restaurants}</p>
                </div>

                <div className="rounded-xl p-4 border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
                  <div className="flex items-center gap-2 mb-1" style={{ color: 'var(--text-muted)' }}>
                    <List className="w-4 h-4" />
                    <span className="text-sm">Lists</span>
                  </div>
                  <p className="text-2xl font-bold" style={{ color: 'var(--text)' }}>{stats.total_lists}</p>
                </div>

                <div className="rounded-xl p-4 border col-span-2 md:col-span-3" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
                  <div className="flex items-center gap-2 mb-2" style={{ color: 'var(--text-muted)' }}>
                    <Star className="w-4 h-4" />
                    <span className="text-sm">Average Ratings</span>
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <p className="text-xs mb-1" style={{ color: 'var(--text-muted)' }}>Food</p>
                      <div className="flex items-center gap-2">
                        <StarRating rating={stats.avg_food_rating} readonly size="sm" />
                        <span className="text-sm font-semibold" style={{ color: 'var(--text)' }}>
                          {stats.avg_food_rating.toFixed(1)}
                        </span>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs mb-1" style={{ color: 'var(--text-muted)' }}>Service</p>
                      <div className="flex items-center gap-2">
                        <StarRating rating={stats.avg_service_rating} readonly size="sm" />
                        <span className="text-sm font-semibold" style={{ color: 'var(--text)' }}>
                          {stats.avg_service_rating.toFixed(1)}
                        </span>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs mb-1" style={{ color: 'var(--text-muted)' }}>Ambiance</p>
                      <div className="flex items-center gap-2">
                        <StarRating rating={stats.avg_ambiance_rating} readonly size="sm" />
                        <span className="text-sm font-semibold" style={{ color: 'var(--text)' }}>
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
          <div className="flex border-b" style={{ borderColor: 'var(--border)' }}>
            <button
              onClick={() => setActiveTab('reviews')}
              className={`px-6 py-3 font-medium transition-colors border-b-2 ${
                activeTab === 'reviews'
                  ? 'text-(--accent)'
                  : 'border-transparent hover:text-(--text)'
              }`}
              style={activeTab === 'reviews' ? { borderColor: 'var(--accent)', color: 'var(--accent)' } : { color: 'var(--text-muted)' }}
            >
              Reviews ({stats.total_reviews})
            </button>
            <button
              onClick={() => setActiveTab('lists')}
              className={`px-6 py-3 font-medium transition-colors border-b-2 ${
                activeTab === 'lists'
                  ? 'text-(--accent)'
                  : 'border-transparent hover:text-(--text)'
              }`}
              style={activeTab === 'lists' ? { borderColor: 'var(--accent)', color: 'var(--accent)' } : { color: 'var(--text-muted)' }}
            >
              Lists ({stats.total_lists})
            </button>
            <button
              onClick={() => setActiveTab('about')}
              className={`px-6 py-3 font-medium transition-colors border-b-2 ${
                activeTab === 'about'
                  ? 'text-(--accent)'
                  : 'border-transparent hover:text-(--text)'
              }`}
              style={activeTab === 'about' ? { borderColor: 'var(--accent)', color: 'var(--accent)' } : { color: 'var(--text-muted)' }}
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
                  <MessageSquare className="w-12 h-12 mx-auto mb-3" style={{ color: 'var(--text-muted)' }} />
                  <p style={{ color: 'var(--text-muted)' }}>No reviews yet</p>
                </div>
              ) : (
                reviews.map((review) => (
                  <div key={review.id} className="card-glass p-6 hover:shadow-xl transition-shadow">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <h3
                          onClick={() => navigate(`/restaurants/${review.restaurant_id}`)}
                          className="text-lg font-semibold mb-1 cursor-pointer transition-colors"
                          style={{ color: 'var(--text)' }}
                          onMouseEnter={(e) => e.currentTarget.style.color = 'var(--accent)'}
                          onMouseLeave={(e) => e.currentTarget.style.color = 'var(--text)'}
                        >
                          {review.restaurant?.name || `Restaurant #${review.restaurant_id}`}
                        </h3>
                        <div className="flex items-center gap-4 text-sm" style={{ color: 'var(--text-muted)' }}>
                          {review.restaurant?.address && (
                            <span className="flex items-center gap-1">
                              <MapPin className="w-3 h-3" />
                              {review.restaurant.address}
                            </span>
                          )}
                          <span>{new Date(review.created_at).toLocaleDateString()}</span>
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-3 gap-4 mb-3">
                      <div>
                        <p className="text-xs mb-1" style={{ color: 'var(--text-muted)' }}>Food</p>
                        <StarRating rating={review.food_rating} readonly size="sm" />
                      </div>
                      <div>
                        <p className="text-xs mb-1" style={{ color: 'var(--text-muted)' }}>Service</p>
                        <StarRating rating={review.service_rating} readonly size="sm" />
                      </div>
                      <div>
                        <p className="text-xs mb-1" style={{ color: 'var(--text-muted)' }}>Ambiance</p>
                        <StarRating rating={review.ambiance_rating} readonly size="sm" />
                      </div>
                    </div>

                    {review.comment && (
                      <p className="mb-3" style={{ color: 'var(--text)' }}>{review.comment}</p>
                    )}

                    {review.photos && review.photos.length > 0 && (
                      <div className="flex flex-wrap gap-2 mt-3">
                        {review.photos.map((photo) => (
                          <div
                            key={photo.id}
                            className="relative group cursor-pointer"
                            onClick={(e) => {
                              e.stopPropagation();
                              window.open(photo.photo_url, '_blank');
                            }}
                          >
                            <img
                              src={photo.photo_url}
                              alt={photo.caption || 'Review photo'}
                              className="w-24 h-24 object-cover rounded-lg border-2 transition-all"
                              style={{ borderColor: 'var(--border)' }}
                              onMouseEnter={(e) => {
                                e.currentTarget.style.borderColor = 'var(--accent)';
                                e.currentTarget.style.transform = 'scale(1.05)';
                              }}
                              onMouseLeave={(e) => {
                                e.currentTarget.style.borderColor = 'var(--border)';
                                e.currentTarget.style.transform = 'scale(1)';
                              }}
                            />
                            {photo.caption && (
                              <div
                                className="absolute bottom-0 left-0 right-0 px-2 py-1 text-xs rounded-b-lg"
                                style={{
                                  background: 'rgba(0, 0, 0, 0.7)',
                                  color: 'var(--text)'
                                }}
                              >
                                {photo.caption}
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'lists' && (
            <div className="space-y-4">
              {!isOwnProfile ? (
                <div className="card-glass p-12 text-center">
                  <Lock className="w-12 h-12 mx-auto mb-3" style={{ color: 'var(--text-muted)' }} />
                  <p style={{ color: 'var(--text-muted)' }}>User lists are private</p>
                </div>
              ) : lists.length === 0 ? (
                <div className="card-glass p-12 text-center">
                  <List className="w-12 h-12 mx-auto mb-3" style={{ color: 'var(--text-muted)' }} />
                  <p className="mb-4" style={{ color: 'var(--text-muted)' }}>No lists yet</p>
                  <button
                    onClick={() => navigate('/lists')}
                    className="btn-primary inline-flex items-center gap-2"
                  >
                    <List className="w-4 h-4" />
                    Create Your First List
                  </button>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {lists.map((list) => (
                    <div
                      key={list.id}
                      onClick={() => navigate(`/lists/${list.id}`)}
                      className="card-glass p-6 hover:shadow-xl transition-all cursor-pointer"
                    >
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex-1">
                          <h3
                            className="text-lg font-semibold mb-1 list-title"
                            style={{ color: 'var(--text)' }}
                          >
                            {list.name}
                          </h3>
                          {list.description && (
                            <p className="text-sm line-clamp-2" style={{ color: 'var(--text-muted)' }}>
                              {list.description}
                            </p>
                          )}
                        </div>
                        <div className="flex items-center gap-1" style={{ color: 'var(--text-muted)' }}>
                          {list.is_public ? (
                            <div title="Public"><Globe className="w-4 h-4" /></div>
                          ) : (
                            <div title="Private"><Lock className="w-4 h-4" /></div>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-4 text-sm" style={{ color: 'var(--text-muted)' }}>
                        <span className="flex items-center gap-1">
                          <MapPin className="w-4 h-4" />
                          {list.restaurant_count || 0} restaurant{list.restaurant_count !== 1 ? 's' : ''}
                        </span>
                        <span>{new Date(list.created_at).toLocaleDateString()}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'about' && (
            <div className="card-glass p-6">
              <h3 className="text-lg font-semibold mb-4" style={{ color: 'var(--text)' }}>About</h3>
              <div className="space-y-3">
                <div className="flex items-center gap-3">
                  <User className="w-5 h-5" style={{ color: 'var(--text-muted)' }} />
                  <div>
                    <p className="text-sm" style={{ color: 'var(--text-muted)' }}>Username</p>
                    <p className="font-medium" style={{ color: 'var(--text)' }}>{user.username}</p>
                  </div>
                </div>
                {user.email && (
                  <div className="flex items-center gap-3">
                    <span className="text-xl">📧</span>
                    <div>
                      <p className="text-sm" style={{ color: 'var(--text-muted)' }}>Email</p>
                      <p className="font-medium" style={{ color: 'var(--text)' }}>{user.email}</p>
                    </div>
                  </div>
                )}
                <div className="flex items-center gap-3">
                  <Calendar className="w-5 h-5" style={{ color: 'var(--text-muted)' }} />
                  <div>
                    <p className="text-sm" style={{ color: 'var(--text-muted)' }}>Member since</p>
                    <p className="font-medium" style={{ color: 'var(--text)' }}>
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
        onSuccess={() => refetchProfile()}
      />
    </div>
  );
}
