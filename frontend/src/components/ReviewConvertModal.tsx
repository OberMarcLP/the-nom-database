import { useState } from 'react';
import { Modal } from './Modal';
import { StarRating } from './StarRating';
import { AlertDialog } from './AlertDialog';

interface ReviewConvertModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: {
    foodRating: number;
    serviceRating: number;
    ambianceRating: number;
    comment: string;
    description: string;
  }) => void;
  restaurantName: string;
}

export function ReviewConvertModal({ isOpen, onClose, onSubmit, restaurantName }: ReviewConvertModalProps) {
  const [foodRating, setFoodRating] = useState(0);
  const [serviceRating, setServiceRating] = useState(0);
  const [ambianceRating, setAmbianceRating] = useState(0);
  const [comment, setComment] = useState('');
  const [description, setDescription] = useState('');
  const [alertMessage, setAlertMessage] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (foodRating === 0 || serviceRating === 0 || ambianceRating === 0) {
      setAlertMessage('Please provide all ratings');
      return;
    }

    onSubmit({
      foodRating,
      serviceRating,
      ambianceRating,
      comment,
      description,
    });

    // Reset form
    setFoodRating(0);
    setServiceRating(0);
    setAmbianceRating(0);
    setComment('');
    setDescription('');
  };

  const handleClose = () => {
    setFoodRating(0);
    setServiceRating(0);
    setAmbianceRating(0);
    setComment('');
    setDescription('');
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title={`Test & Review: ${restaurantName}`}>
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="admin-form-group">
          <label className="admin-label">
            Description (Optional)
          </label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Add a description for this restaurant..."
            className="admin-textarea"
            rows={3}
          />
        </div>

        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-(--text)">
            Your Review
          </h3>

          <div className="admin-form-group">
            <label className="admin-label">
              Food Rating *
            </label>
            <StarRating rating={foodRating} onRatingChange={setFoodRating} />
          </div>

          <div className="admin-form-group">
            <label className="admin-label">
              Service Rating *
            </label>
            <StarRating rating={serviceRating} onRatingChange={setServiceRating} />
          </div>

          <div className="admin-form-group">
            <label className="admin-label">
              Ambiance Rating *
            </label>
            <StarRating rating={ambianceRating} onRatingChange={setAmbianceRating} />
          </div>

          <div className="admin-form-group">
            <label className="admin-label">
              Comment (Optional)
            </label>
            <textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder="Share your experience..."
              className="admin-textarea"
              rows={4}
            />
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-(--border)">
          <button
            type="button"
            onClick={handleClose}
            className="admin-btn"
          >
            Cancel
          </button>
          <button
            type="submit"
            className="admin-btn-primary"
          >
            Submit Review & Convert
          </button>
        </div>
      </form>
      <AlertDialog
        isOpen={alertMessage !== ''}
        onClose={() => setAlertMessage('')}
        message={alertMessage}
      />
    </Modal>
  );
}
