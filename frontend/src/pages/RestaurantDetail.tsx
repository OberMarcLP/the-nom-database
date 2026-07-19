import { useState, useEffect, useRef } from 'react';
import { MapPin, Tag, Utensils, Edit, Trash2, Plus, Loader2, Phone, Globe, Camera, ThumbsUp, ThumbsDown } from 'lucide-react';
import { Restaurant, voteOnReview, removeVote, uploadReviewPhoto, isSafeHttpUrl } from '../services/api';
import { useRatings, useCreateRating, useMenuPhotos, useUpdatePhotoCaption, useDeleteMenuPhoto } from '../hooks/useApi';
import { StarRating } from '../components/StarRating';
import { RestaurantMap } from '../components/RestaurantMap';
import { RatingForm } from '../components/RatingForm';
import { PhotoGallery } from '../components/PhotoGallery';
import { UserBadge } from '../components/UserBadge';
import { AddToListButton } from '../components/AddToListButton';

interface RestaurantDetailProps {
  restaurant: Restaurant;
  onEdit: () => void;
  onDelete: () => void;
  highlightedRatingId?: number;
  highlightedPhotoId?: number;
}

export function RestaurantDetail({ restaurant, onEdit, onDelete, highlightedRatingId, highlightedPhotoId }: RestaurantDetailProps) {
  const [showRatingForm, setShowRatingForm] = useState(false);
  const [sortBy, setSortBy] = useState<'recent' | 'helpful' | 'rating'>('recent');
  const ratingRefs = useRef<{ [key: number]: HTMLDivElement | null }>({});
  const photosRef = useRef<HTMLDivElement>(null);

  // Use React Query hooks
  const { data: ratings = [], isLoading: loading, refetch: refetchRatings } = useRatings(restaurant.id);
  const { data: photos = [], isLoading: loadingPhotos } = useMenuPhotos(restaurant.id);
  const createRatingMutation = useCreateRating();
  const updateCaptionMutation = useUpdatePhotoCaption();
  const deletePhotoMutation = useDeleteMenuPhoto();

  // Combine menu photos and review photos for the gallery
  const allPhotos = [
    ...photos.map(photo => ({
      ...photo,
      source: 'menu' as const,
      reviewInfo: null,
      originalId: photo.id, // Store original ID for matching
    })),
    ...ratings.flatMap(rating =>
      (rating.photos || []).map(photo => ({
        id: photo.id,
        restaurant_id: restaurant.id,
        filename: photo.photo_url,
        original_filename: null,
        caption: photo.caption || 'Review photo',
        file_size: null,
        mime_type: null,
        url: photo.photo_url,
        created_at: photo.created_at,
        updated_at: photo.created_at,
        source: 'review' as const,
        originalId: photo.id, // Store original ID for matching
        reviewInfo: {
          username: rating.user?.username || 'Anonymous',
          date: rating.created_at,
          ratings: {
            food: rating.food_rating,
            service: rating.service_rating,
            ambiance: rating.ambiance_rating,
          },
        },
      }))
    ),
  ];

  const handleVote = async (ratingId: number, voteType: 'helpful' | 'not_helpful', currentVote?: string | null) => {
    try {
      if (currentVote === voteType) {
        // Remove vote if clicking the same button
        await removeVote(ratingId);
      } else {
        // Add or change vote
        await voteOnReview(ratingId, voteType);
      }
      refetchRatings();
    } catch (error) {
      console.error('Failed to vote:', error);
    }
  };

  const handleAddRating = async (data: {
    food_rating: number;
    service_rating: number;
    ambiance_rating: number;
    comment?: string;
    photos?: { file: File; caption: string }[];
  }) => {
    createRatingMutation.mutate(
      { ...data, restaurant_id: restaurant.id },
      {
        onSuccess: async (newRating) => {
          // Upload photos if any
          if (data.photos && data.photos.length > 0) {
            console.log('Uploading photos for rating ID:', newRating.id);
            try {
              for (const photo of data.photos) {
                console.log('Uploading photo:', photo.file.name, 'with caption:', photo.caption);
                const result = await uploadReviewPhoto(newRating.id, photo.file, photo.caption);
                console.log('Photo uploaded successfully:', result);
              }
              // Refetch ratings to show the new photos
              refetchRatings();
            } catch (error) {
              console.error('Failed to upload photos:', error);
              alert(`Failed to upload photos: ${error}`);
            }
          }
          setShowRatingForm(false);
        },
      }
    );
  };


  const handleCaptionUpdate = async (id: number, caption: string) => {
    updateCaptionMutation.mutate({ id, caption, restaurantId: restaurant.id });
  };

  const handlePhotoDelete = async (id: number) => {
    deletePhotoMutation.mutate({ id, restaurantId: restaurant.id });
  };

  // Sort ratings based on selected option
  const sortedRatings = [...ratings].sort((a, b) => {
    if (sortBy === 'recent') {
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    } else if (sortBy === 'helpful') {
      return (b.helpful_count - b.not_helpful_count) - (a.helpful_count - a.not_helpful_count);
    } else if (sortBy === 'rating') {
      const avgA = (a.food_rating + a.service_rating + a.ambiance_rating) / 3;
      const avgB = (b.food_rating + b.service_rating + b.ambiance_rating) / 3;
      return avgB - avgA;
    }
    return 0;
  });

  // Scroll to highlighted rating
  useEffect(() => {
    if (highlightedRatingId && ratingRefs.current[highlightedRatingId]) {
      setTimeout(() => {
        ratingRefs.current[highlightedRatingId]?.scrollIntoView({
          behavior: 'smooth',
          block: 'center',
        });
      }, 300); // Delay to ensure modal is rendered
    }
  }, [highlightedRatingId, ratings]);

  // Scroll to photos section when highlightedPhotoId is present
  useEffect(() => {
    if (highlightedPhotoId && photosRef.current) {
      setTimeout(() => {
        photosRef.current?.scrollIntoView({
          behavior: 'smooth',
          block: 'start',
        });
      }, 300); // Delay to ensure modal is rendered
    }
  }, [highlightedPhotoId, photos]);

  return (
    <div className="space-y-6">
      <div className="flex gap-2 flex-wrap">
        <AddToListButton restaurantId={restaurant.id} restaurantName={restaurant.name} />
        <button onClick={onEdit} className="btn btn-secondary flex items-center gap-2">
          <Edit className="w-4 h-4" />
          Edit
        </button>
        <button onClick={onDelete} className="btn btn-danger flex items-center gap-2">
          <Trash2 className="w-4 h-4" />
          Delete
        </button>
      </div>

      {restaurant.description && (
        <p className="text-muted">{restaurant.description}</p>
      )}

        <div className="flex flex-wrap gap-2">
          {restaurant.category && (
            <span className="badge-category">
              <Tag className="w-4 h-4" />
              {restaurant.category.name}
            </span>
          )}
          {restaurant.food_types?.map((ft) => (
            <span key={ft.id} className="badge-food-type">
              <Utensils className="w-4 h-4" />
              {ft.name}
            </span>
          ))}
        </div>

      {restaurant.address && (
        <div className="flex items-start gap-2 text-muted">
          <MapPin className="w-5 h-5 mt-0.5 shrink-0" />
          <span>{restaurant.address}</span>
        </div>
      )}

      {restaurant.phone && (
        <a href={`tel:${restaurant.phone}`} className="flex items-center gap-2 text-muted hover:text-accent transition-colors">
          <Phone className="w-5 h-5 shrink-0" />
          <span>{restaurant.phone}</span>
        </a>
      )}

      {restaurant.website && isSafeHttpUrl(restaurant.website) && (
        <a href={restaurant.website} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-muted hover:text-accent transition-colors">
          <Globe className="w-5 h-5 shrink-0" />
          <span className="truncate">{restaurant.website.replace(/^https?:\/\//, '').replace(/\/$/, '')}</span>
        </a>
      )}

      {restaurant.avg_rating && (
        <div className="card">
          <h3 className="font-semibold mb-3">Average Ratings</h3>
          <div className="grid grid-cols-3 gap-4 text-center">
            <div>
              <p className="text-sm text-muted mb-1">Food</p>
              <p className="text-2xl font-bold">{restaurant.avg_rating.food.toFixed(1)}</p>
            </div>
            <div>
              <p className="text-sm text-muted mb-1">Service</p>
              <p className="text-2xl font-bold">{restaurant.avg_rating.service.toFixed(1)}</p>
            </div>
            <div>
              <p className="text-sm text-muted mb-1">Ambiance</p>
              <p className="text-2xl font-bold">{restaurant.avg_rating.ambiance.toFixed(1)}</p>
            </div>
          </div>
          <div className="mt-4 pt-4 border-t border-default text-center">
            <p className="text-sm text-muted mb-1">Overall</p>
            <p className="text-3xl font-bold text-accent">
              {restaurant.avg_rating.overall.toFixed(1)}
            </p>
            <p className="text-sm text-muted">
              {restaurant.avg_rating.count} review{restaurant.avg_rating.count !== 1 ? 's' : ''}
            </p>
          </div>
        </div>
      )}

      {restaurant.latitude && restaurant.longitude && (
        <div>
          <h3 className="font-semibold mb-3">Location</h3>
          <RestaurantMap
            latitude={restaurant.latitude}
            longitude={restaurant.longitude}
            name={restaurant.name}
          />
        </div>
      )}

      {!restaurant.is_suggestion && (
        <div ref={photosRef}>
          <h3 className="font-semibold flex items-center gap-2 mb-4">
            <Camera className="w-5 h-5" />
            Photos ({allPhotos.length})
          </h3>

          {loadingPhotos ? (
            <div className="flex justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-(--info)" />
            </div>
          ) : allPhotos.length === 0 ? (
            <p className="text-muted text-center py-8">
              No photos yet. Add photos when you write a review!
            </p>
          ) : (
            <PhotoGallery
              photos={allPhotos}
              onCaptionUpdate={handleCaptionUpdate}
              onDelete={handlePhotoDelete}
              highlightedPhotoId={highlightedPhotoId}
            />
          )}
        </div>
      )}

      {!restaurant.is_suggestion && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold">Reviews ({ratings.length})</h3>
            <button
              onClick={() => setShowRatingForm(true)}
              className="btn btn-primary flex items-center gap-2 text-sm"
            >
              <Plus className="w-4 h-4" />
              Add Review
            </button>
          </div>

          {showRatingForm && (
            <div className="card mb-4">
              <RatingForm onSubmit={handleAddRating} onCancel={() => setShowRatingForm(false)} />
            </div>
          )}

          {ratings.length > 0 && (
            <div className="mb-4 flex items-center gap-2">
              <label className="text-sm text-muted">Sort by:</label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as 'recent' | 'helpful' | 'rating')}
                className="input-glass text-sm py-1 px-3"
              >
                <option value="recent">Most Recent</option>
                <option value="helpful">Most Helpful</option>
                <option value="rating">Highest Rating</option>
              </select>
            </div>
          )}

          {loading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-(--info)" />
            </div>
          ) : ratings.length === 0 ? (
            <p className="text-muted text-center py-8">
              No reviews yet. Be the first to review!
            </p>
          ) : (
            <div className="space-y-4">
              {sortedRatings.map((rating) => (
                <div
                  key={rating.id}
                  ref={(el) => {
                    ratingRefs.current[rating.id] = el;
                  }}
                  className={highlightedRatingId === rating.id ? 'card-glass-strong' : 'card'}
                  style={{
                    transition: 'all 0.3s ease',
                    ...(highlightedRatingId === rating.id
                      ? {
                          padding: '20px',
                          margin: '8px 0',
                        }
                      : {}),
                  }}
                >
                  <div className="grid grid-cols-3 gap-4 mb-3">
                    <div>
                      <p className="text-xs text-muted mb-1">Food</p>
                      <StarRating rating={rating.food_rating} readonly size="sm" />
                    </div>
                    <div>
                      <p className="text-xs text-muted mb-1">Service</p>
                      <StarRating rating={rating.service_rating} readonly size="sm" />
                    </div>
                    <div>
                      <p className="text-xs text-muted mb-1">Ambiance</p>
                      <StarRating rating={rating.ambiance_rating} readonly size="sm" />
                    </div>
                  </div>
                  {rating.comment && (
                    <p className="text-muted text-sm mb-3">{rating.comment}</p>
                  )}

                  {/* Voting buttons */}
                  <div className="flex items-center gap-4 mt-3 pt-3 border-t border-default">
                    <button
                      onClick={() => handleVote(rating.id, 'helpful', rating.user_vote)}
                      className={`flex items-center gap-1 text-xs transition-colors ${
                        rating.user_vote === 'helpful'
                          ? 'text-info font-medium'
                          : 'text-muted hover:text-info'
                      }`}
                    >
                      <ThumbsUp className="w-4 h-4" />
                      <span>{rating.helpful_count || 0}</span>
                    </button>
                    <button
                      onClick={() => handleVote(rating.id, 'not_helpful', rating.user_vote)}
                      className={`flex items-center gap-1 text-xs transition-colors ${
                        rating.user_vote === 'not_helpful'
                          ? 'text-danger font-medium'
                          : 'text-muted hover:text-danger'
                      }`}
                    >
                      <ThumbsDown className="w-4 h-4" />
                      <span>{rating.not_helpful_count || 0}</span>
                    </button>
                    <div className="flex-1" />
                    <p className="text-xs text-muted">
                      {new Date(rating.created_at).toLocaleDateString()}
                    </p>
                    {rating.user && (
                      <UserBadge user={rating.user} size="sm" />
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
