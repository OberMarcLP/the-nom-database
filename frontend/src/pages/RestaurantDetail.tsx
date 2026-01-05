import { useState } from 'react';
import { MapPin, Tag, Utensils, Edit, Trash2, Plus, Loader2, Phone, Globe, Camera, ChevronUp } from 'lucide-react';
import { Restaurant } from '../services/api';
import { useRatings, useCreateRating, useMenuPhotos, useUploadMenuPhoto, useUpdatePhotoCaption, useDeleteMenuPhoto } from '../hooks/useApi';
import { StarRating } from '../components/StarRating';
import { RestaurantMap } from '../components/RestaurantMap';
import { RatingForm } from '../components/RatingForm';
import { PhotoUpload } from '../components/PhotoUpload';
import { PhotoGallery } from '../components/PhotoGallery';
import { UserBadge } from '../components/UserBadge';

interface RestaurantDetailProps {
  restaurant: Restaurant;
  onEdit: () => void;
  onDelete: () => void;
}

export function RestaurantDetail({ restaurant, onEdit, onDelete }: RestaurantDetailProps) {
  const [showRatingForm, setShowRatingForm] = useState(false);
  const [showPhotoUpload, setShowPhotoUpload] = useState(false);

  // Use React Query hooks
  const { data: ratings = [], isLoading: loading } = useRatings(restaurant.id);
  const { data: photos = [], isLoading: loadingPhotos } = useMenuPhotos(restaurant.id);
  const createRatingMutation = useCreateRating();
  const uploadPhotoMutation = useUploadMenuPhoto();
  const updateCaptionMutation = useUpdatePhotoCaption();
  const deletePhotoMutation = useDeleteMenuPhoto();

  const handleAddRating = async (data: {
    food_rating: number;
    service_rating: number;
    ambiance_rating: number;
    comment?: string;
  }) => {
    createRatingMutation.mutate(
      { ...data, restaurant_id: restaurant.id },
      {
        onSuccess: () => {
          setShowRatingForm(false);
        },
      }
    );
  };

  const handlePhotoUpload = async (file: File, caption: string) => {
    uploadPhotoMutation.mutate(
      { restaurantId: restaurant.id, photo: file, caption },
      {
        onSuccess: () => {
          setShowPhotoUpload(false);
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

  return (
    <div className="space-y-6">
      <div className="flex gap-2">
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
        <p className="text-gray-600 dark:text-gray-400">{restaurant.description}</p>
      )}

      <div className="flex flex-wrap gap-2">
        {restaurant.category && (
          <span className="inline-flex items-center gap-1 px-3 py-1.5 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-full">
            <Tag className="w-4 h-4" />
            {restaurant.category.name}
          </span>
        )}
        {restaurant.food_types?.map((ft) => (
          <span key={ft.id} className="inline-flex items-center gap-1 px-3 py-1.5 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded-full">
            <Utensils className="w-4 h-4" />
            {ft.name}
          </span>
        ))}
      </div>

      {restaurant.address && (
        <div className="flex items-start gap-2 text-gray-600 dark:text-gray-400">
          <MapPin className="w-5 h-5 mt-0.5 flex-shrink-0" />
          <span>{restaurant.address}</span>
        </div>
      )}

      {restaurant.phone && (
        <a href={`tel:${restaurant.phone}`} className="flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-blue-500 transition-colors">
          <Phone className="w-5 h-5 flex-shrink-0" />
          <span>{restaurant.phone}</span>
        </a>
      )}

      {restaurant.website && (
        <a href={restaurant.website} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-blue-500 transition-colors">
          <Globe className="w-5 h-5 flex-shrink-0" />
          <span className="truncate">{restaurant.website}</span>
        </a>
      )}

      {restaurant.avg_rating && (
        <div className="card">
          <h3 className="font-semibold mb-3">Average Ratings</h3>
          <div className="grid grid-cols-3 gap-4 text-center">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">Food</p>
              <p className="text-2xl font-bold">{restaurant.avg_rating.food.toFixed(1)}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">Service</p>
              <p className="text-2xl font-bold">{restaurant.avg_rating.service.toFixed(1)}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">Ambiance</p>
              <p className="text-2xl font-bold">{restaurant.avg_rating.ambiance.toFixed(1)}</p>
            </div>
          </div>
          <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-center">
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">Overall</p>
            <p className="text-3xl font-bold text-blue-600 dark:text-blue-400">
              {restaurant.avg_rating.overall.toFixed(1)}
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
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
        <div>
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold flex items-center gap-2">
              <Camera className="w-5 h-5" />
              Menu Photos
            </h3>
            <button
              onClick={() => setShowPhotoUpload(!showPhotoUpload)}
              className="btn btn-primary flex items-center gap-2 text-sm"
            >
              {showPhotoUpload ? (
                <>
                  <ChevronUp className="w-4 h-4" />
                  Hide Upload
                </>
              ) : (
                <>
                  <Plus className="w-4 h-4" />
                  Upload Photo
                </>
              )}
            </button>
          </div>

          {showPhotoUpload && (
            <div className="card mb-4">
              <PhotoUpload onUpload={handlePhotoUpload} />
            </div>
          )}

          {loadingPhotos ? (
            <div className="flex justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-blue-500" />
            </div>
          ) : (
            <PhotoGallery
              photos={photos}
              onCaptionUpdate={handleCaptionUpdate}
              onDelete={handlePhotoDelete}
            />
          )}
        </div>
      )}

      {!restaurant.is_suggestion && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold">Reviews</h3>
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

          {loading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-blue-500" />
            </div>
          ) : ratings.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400 text-center py-8">
              No reviews yet. Be the first to review!
            </p>
          ) : (
            <div className="space-y-4">
              {ratings.map((rating) => (
                <div key={rating.id} className="card">
                  <div className="grid grid-cols-3 gap-4 mb-3">
                    <div>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Food</p>
                      <StarRating rating={rating.food_rating} readonly size="sm" />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Service</p>
                      <StarRating rating={rating.service_rating} readonly size="sm" />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Ambiance</p>
                      <StarRating rating={rating.ambiance_rating} readonly size="sm" />
                    </div>
                  </div>
                  {rating.comment && (
                    <p className="text-gray-600 dark:text-gray-400 text-sm">{rating.comment}</p>
                  )}
                  <div className="flex items-center justify-between mt-2">
                    <p className="text-xs text-gray-400 dark:text-gray-500">
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
