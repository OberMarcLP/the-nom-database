import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useRestaurant, useUpdateRestaurant } from '../hooks/useApi';
import { RestaurantDetail } from './RestaurantDetail';
import { useState, useEffect } from 'react';
import { Modal } from '../components/Modal';
import { RestaurantForm } from '../components/RestaurantForm';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { deleteRestaurant, CreateRestaurantData } from '../services/api';

export function RestaurantPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const restaurantId = parseInt(id || '0', 10);
  const highlightedRatingId = searchParams.get('rating') ? parseInt(searchParams.get('rating')!, 10) : undefined;
  const highlightedPhotoId = searchParams.get('photo') ? parseInt(searchParams.get('photo')!, 10) : undefined;
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const { data: restaurant, isLoading, error, refetch } = useRestaurant(restaurantId);
  const updateRestaurantMutation = useUpdateRestaurant();

  // Automatically close if restaurant not found after loading
  useEffect(() => {
    if (!isLoading && (error || !restaurant)) {
      navigate(-1); // Go back to previous page
    }
  }, [isLoading, error, restaurant, navigate]);

  // Handle ESC key to close modals
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (showDeleteConfirm) {
          setShowDeleteConfirm(false);
        } else if (showEditModal) {
          setShowEditModal(false);
        }
      }
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [showEditModal, showDeleteConfirm]);

  const handleClose = () => {
    navigate(-1); // Go back to previous page (admin or home)
  };

  const handleEdit = () => {
    setShowEditModal(true);
  };

  const handleUpdate = async (data: CreateRestaurantData) => {
    if (!restaurant) return;
    updateRestaurantMutation.mutate(
      { id: restaurant.id, data },
      {
        onSuccess: () => {
          setShowEditModal(false);
          refetch();
        },
      }
    );
  };

  const handleDelete = () => {
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!restaurant) return;

    try {
      await deleteRestaurant(restaurant.id);
      setShowDeleteConfirm(false);
      navigate('/');
    } catch (error) {
      console.error('Failed to delete restaurant:', error);
      alert('Failed to delete restaurant. Please try again.');
      setShowDeleteConfirm(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <Loader2 className="w-8 h-8 animate-spin text-(--accent)" />
      </div>
    );
  }

  // Display the restaurant in a modal, just like on the home page
  return (
    <>
      <Modal
        isOpen={!!restaurant}
        onClose={handleClose}
        title={restaurant?.name || ''}
      >
        {restaurant && (
          <RestaurantDetail
            restaurant={restaurant}
            onEdit={handleEdit}
            onDelete={handleDelete}
            highlightedRatingId={highlightedRatingId}
            highlightedPhotoId={highlightedPhotoId}
          />
        )}
      </Modal>

      <Modal
        isOpen={showEditModal}
        onClose={() => setShowEditModal(false)}
        title="Edit Restaurant"
      >
        {restaurant && (
          <RestaurantForm
            restaurant={restaurant}
            onSubmit={handleUpdate}
            onCancel={() => setShowEditModal(false)}
          />
        )}
      </Modal>

      <ConfirmDialog
        isOpen={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={confirmDelete}
        title="Delete Restaurant"
        message={`Delete "${restaurant?.name}"? This will also delete all associated ratings and photos. This action cannot be undone.`}
        confirmText="Delete"
        cancelText="Cancel"
        isDangerous={true}
      />
    </>
  );
}
