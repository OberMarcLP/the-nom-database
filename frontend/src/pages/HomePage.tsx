import { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Plus, Loader2, CheckCircle, XCircle, Tag, Utensils, MapPin, Phone, Globe } from 'lucide-react';
import { Restaurant, CreateRestaurantData, CreateSuggestionData, RestaurantFilters } from '../services/api';
import { useRestaurants, useRestaurant, useUpdateRestaurant, useDeleteRestaurant, useCreateSuggestion, useConvertSuggestion, useDeleteSuggestion } from '../hooks/useApi';
import { RestaurantCard } from '../components/RestaurantCard';
import { RestaurantForm } from '../components/RestaurantForm';
import { SuggestionForm } from '../components/SuggestionForm';
import { ReviewConvertModal } from '../components/ReviewConvertModal';
import { Modal } from '../components/Modal';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { AlertDialog } from '../components/AlertDialog';
import { RestaurantDetail } from './RestaurantDetail';
import { UserBadge } from '../components/UserBadge';
import { RestaurantMap } from '../components/RestaurantMap';
import { Recommendations } from '../components/Recommendations';
import { ActivityFeed } from '../components/ActivityFeed';

interface HomePageProps {
  filters: RestaurantFilters;
  isAuthenticated?: boolean;
}

export function HomePage({ filters, isAuthenticated = false }: HomePageProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [showAddSuggestionModal, setShowAddSuggestionModal] = useState(false);
  const [selectedRestaurantId, setSelectedRestaurantId] = useState<number | null>(null);
  const [viewingSuggestion, setViewingSuggestion] = useState<Restaurant | null>(null);
  const [editingRestaurant, setEditingRestaurant] = useState<Restaurant | null>(null);
  const [reviewingSuggestion, setReviewingSuggestion] = useState<Restaurant | null>(null);
  const [rejectingRestaurant, setRejectingRestaurant] = useState<Restaurant | null>(null);
  const [deletingRestaurant, setDeletingRestaurant] = useState<Restaurant | null>(null);
  const [alertMessage, setAlertMessage] = useState<string>('');

  // Use React Query hooks
  const { data: restaurants = [], isLoading: loading } = useRestaurants(filters);
  const { data: selectedRestaurant } = useRestaurant(selectedRestaurantId || 0, {
    enabled: selectedRestaurantId !== null && selectedRestaurantId > 0,
  });
  const createSuggestionMutation = useCreateSuggestion();
  const updateRestaurantMutation = useUpdateRestaurant();
  const deleteRestaurantMutation = useDeleteRestaurant();
  const convertSuggestionMutation = useConvertSuggestion();
  const deleteSuggestionMutation = useDeleteSuggestion();

  // Handle restaurant selection from global search
  useEffect(() => {
    const state = location.state as { restaurantId?: number } | null;
    if (state?.restaurantId) {
      setSelectedRestaurantId(state.restaurantId);
      // Clear the location state
      navigate(location.pathname, { replace: true });
    }
  }, [location, navigate]);

  const handleCreateSuggestion = async (data: CreateSuggestionData) => {
    createSuggestionMutation.mutate(data, {
      onSuccess: () => {
        setShowAddSuggestionModal(false);
      },
      onError: (error: any) => {
        if (error.message && (error.message.includes('already exists') || error.message.includes('duplicate'))) {
          setAlertMessage('This restaurant already exists in the database. Please search for it instead.');
        } else {
          setAlertMessage('Failed to create suggestion. Please try again.');
        }
      },
    });
  };

  const handleUpdate = async (data: CreateRestaurantData) => {
    if (!editingRestaurant) return;
    updateRestaurantMutation.mutate(
      { id: editingRestaurant.id, data },
      {
        onSuccess: () => {
          setEditingRestaurant(null);
        },
      }
    );
  };

  const handleDelete = async () => {
    setDeletingRestaurant(selectedRestaurant || null);
  };

  const confirmDelete = async () => {
    if (!deletingRestaurant) return;
    deleteRestaurantMutation.mutate(deletingRestaurant.id, {
      onSuccess: () => {
        setSelectedRestaurantId(null);
        setDeletingRestaurant(null);
      },
    });
  };

  const handleReviewSuggestion = (restaurant: Restaurant) => {
    setReviewingSuggestion(restaurant);
  };

  const handleRejectSuggestion = async (restaurant: Restaurant) => {
    setRejectingRestaurant(restaurant);
  };

  const confirmReject = async () => {
    if (!rejectingRestaurant?.suggestion_id) return;
    deleteSuggestionMutation.mutate(rejectingRestaurant.suggestion_id, {
      onSuccess: () => {
        setRejectingRestaurant(null);
      },
    });
  };

  const handleConvertSuggestion = async (data: {
    foodRating: number;
    serviceRating: number;
    ambianceRating: number;
    comment: string;
    description: string;
  }) => {
    if (!reviewingSuggestion?.suggestion_id) return;
    convertSuggestionMutation.mutate(
      {
        id: reviewingSuggestion.suggestion_id,
        data: {
          description: data.description || undefined,
          category_id: reviewingSuggestion.category?.id || undefined,
          food_rating: data.foodRating,
          service_rating: data.serviceRating,
          ambiance_rating: data.ambianceRating,
          comment: data.comment || undefined,
        },
      },
      {
        onSuccess: () => {
          setReviewingSuggestion(null);
        },
        onError: (error: any) => {
          if (error.message && (error.message.includes('already exists') || error.message.includes('duplicate'))) {
            setAlertMessage('This restaurant already exists in the database. The suggestion will be rejected.');
            if (reviewingSuggestion?.suggestion_id) {
              deleteSuggestionMutation.mutate(reviewingSuggestion.suggestion_id, {
                onSuccess: () => {
                  setReviewingSuggestion(null);
                },
              });
            }
          } else {
            setAlertMessage('Failed to convert suggestion. Please try again.');
          }
        },
      }
    );
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    );
  }

  return (
    <div className="space-y-12">
      {/* Recommendations Section - Only show when no filters are active */}
      {Object.keys(filters).length === 0 && (
        <Recommendations
          onRestaurantClick={(restaurant) => {
            if (restaurant.is_suggestion) {
              setViewingSuggestion(restaurant);
            } else {
              setSelectedRestaurantId(restaurant.id);
            }
          }}
        />
      )}

      {/* Main Content Grid */}
      <div className={Object.keys(filters).length === 0 ? "grid grid-cols-1 lg:grid-cols-3 gap-8" : ""}>
        {/* Main Restaurant List */}
        <div className={Object.keys(filters).length === 0 ? "lg:col-span-2" : ""}>
          <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold">
            {Object.keys(filters).length > 0 ? 'Search Results' : 'All Restaurants'}
          </h1>
          {isAuthenticated && (
            <button onClick={() => setShowAddSuggestionModal(true)} className="btn btn-primary flex items-center gap-2">
              <Plus className="w-5 h-5" />
              Add Suggestion
            </button>
          )}
        </div>

        {loading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
          </div>
        ) : restaurants.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-500 dark:text-gray-400 mb-4">
            {Object.keys(filters).length > 0 ? 'No restaurants match your filters.' : 'No restaurants yet.'}
          </p>
          {Object.keys(filters).length === 0 && isAuthenticated && (
            <button onClick={() => setShowAddSuggestionModal(true)} className="btn btn-primary">
              Add your first suggestion
            </button>
          )}
        </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {restaurants.map((restaurant) => (
              <RestaurantCard
                key={restaurant.id}
                restaurant={restaurant}
                onClick={() => {
                  if (restaurant.is_suggestion) {
                    setViewingSuggestion(restaurant);
                  } else {
                    setSelectedRestaurantId(restaurant.id);
                  }
                }}
                onReview={handleReviewSuggestion}
                onReject={handleRejectSuggestion}
              />
            ))}
          </div>
        )}
        </div>

        {/* Activity Feed Sidebar - Only show when no filters are active */}
        {Object.keys(filters).length === 0 && (
          <div className="lg:col-span-1">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Recent Activity</h2>
            <ActivityFeed
              onRestaurantClick={(restaurantId) => setSelectedRestaurantId(restaurantId)}
            />
          </div>
        )}
      </div>

      <Modal isOpen={showAddSuggestionModal} onClose={() => setShowAddSuggestionModal(false)} title="Add Suggestion">
        <SuggestionForm onSubmit={handleCreateSuggestion} onCancel={() => setShowAddSuggestionModal(false)} />
      </Modal>

      <Modal
        isOpen={editingRestaurant !== null}
        onClose={() => setEditingRestaurant(null)}
        title="Edit Restaurant"
      >
        {editingRestaurant && (
          <RestaurantForm
            restaurant={editingRestaurant}
            onSubmit={handleUpdate}
            onCancel={() => setEditingRestaurant(null)}
          />
        )}
      </Modal>

      <Modal
        isOpen={selectedRestaurant !== undefined && selectedRestaurant !== null}
        onClose={() => setSelectedRestaurantId(null)}
        title={selectedRestaurant?.name || ''}
      >
        {selectedRestaurant && (
          <RestaurantDetail
            restaurant={selectedRestaurant}
            onEdit={() => {
              setEditingRestaurant(selectedRestaurant);
              setSelectedRestaurantId(null);
            }}
            onDelete={handleDelete}
          />
        )}
      </Modal>

      <Modal
        isOpen={viewingSuggestion !== null}
        onClose={() => setViewingSuggestion(null)}
        title={viewingSuggestion?.name || ''}
      >
        {viewingSuggestion && (
          <div className="space-y-6">
            <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
              <p className="text-sm text-yellow-800 dark:text-yellow-200">
                This is a pending suggestion. Review and convert it to add it to the database.
              </p>
            </div>

            {isAuthenticated && (
              <div className="flex gap-2">
                <button
                  onClick={() => {
                    setViewingSuggestion(null);
                    handleReviewSuggestion(viewingSuggestion);
                  }}
                  className="btn btn-primary flex items-center gap-2"
                >
                  <CheckCircle className="w-4 h-4" />
                  Review & Convert
                </button>
                <button
                  onClick={() => {
                    setViewingSuggestion(null);
                    handleRejectSuggestion(viewingSuggestion);
                  }}
                  className="btn btn-danger flex items-center gap-2"
                >
                  <XCircle className="w-4 h-4" />
                  Reject
                </button>
              </div>
            )}

            {viewingSuggestion.notes && (
              <p className="text-gray-600 dark:text-gray-400">{viewingSuggestion.notes}</p>
            )}

            <div className="flex flex-wrap gap-2">
              {viewingSuggestion.category && (
                <span className="inline-flex items-center gap-1 px-3 py-1.5 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-full">
                  <Tag className="w-4 h-4" />
                  {viewingSuggestion.category.name}
                </span>
              )}
              {viewingSuggestion.food_types?.map((ft) => (
                <span key={ft.id} className="inline-flex items-center gap-1 px-3 py-1.5 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded-full">
                  <Utensils className="w-4 h-4" />
                  {ft.name}
                </span>
              ))}
            </div>

            {viewingSuggestion.address && (
              <div className="flex items-start gap-2 text-gray-600 dark:text-gray-400">
                <MapPin className="w-5 h-5 mt-0.5 flex-shrink-0" />
                <span>{viewingSuggestion.address}</span>
              </div>
            )}

            {viewingSuggestion.phone && (
              <a href={`tel:${viewingSuggestion.phone}`} className="flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-blue-500 transition-colors">
                <Phone className="w-5 h-5 flex-shrink-0" />
                <span>{viewingSuggestion.phone}</span>
              </a>
            )}

            {viewingSuggestion.website && (
              <a href={viewingSuggestion.website} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-blue-500 transition-colors">
                <Globe className="w-5 h-5 flex-shrink-0" />
                <span className="truncate">{viewingSuggestion.website}</span>
              </a>
            )}

            {viewingSuggestion.latitude && viewingSuggestion.longitude && (
              <div>
                <h3 className="font-semibold mb-3">Location</h3>
                <RestaurantMap
                  latitude={viewingSuggestion.latitude}
                  longitude={viewingSuggestion.longitude}
                  name={viewingSuggestion.name}
                />
              </div>
            )}

            {viewingSuggestion.user && (
              <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
                <h3 className="font-semibold mb-2">Suggested by</h3>
                <UserBadge user={viewingSuggestion.user} />
              </div>
            )}
          </div>
        )}
      </Modal>

      {reviewingSuggestion && (
        <ReviewConvertModal
          isOpen={true}
          onClose={() => setReviewingSuggestion(null)}
          onSubmit={handleConvertSuggestion}
          restaurantName={reviewingSuggestion.name}
        />
      )}

      <ConfirmDialog
        isOpen={rejectingRestaurant !== null}
        onClose={() => setRejectingRestaurant(null)}
        onConfirm={confirmReject}
        title="Reject Suggestion"
        message={`Are you sure you want to reject "${rejectingRestaurant?.name}"?`}
        confirmText="Reject"
        cancelText="Cancel"
        confirmClassName="bg-red-600 hover:bg-red-700 text-white"
      />

      <ConfirmDialog
        isOpen={deletingRestaurant !== null}
        onClose={() => setDeletingRestaurant(null)}
        onConfirm={confirmDelete}
        title="Delete Restaurant"
        message={`Are you sure you want to delete "${deletingRestaurant?.name}"? This action cannot be undone.`}
        confirmText="Delete"
        cancelText="Cancel"
        confirmClassName="bg-red-600 hover:bg-red-700 text-white"
      />

      <AlertDialog
        isOpen={alertMessage !== ''}
        onClose={() => setAlertMessage('')}
        message={alertMessage}
      />
    </div>
  );
}
