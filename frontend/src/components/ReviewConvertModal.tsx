import { useEffect, useRef, useState } from 'react';
import { Camera, X } from 'lucide-react';
import { Modal } from './Modal';
import { StarRating } from './StarRating';
import { AlertDialog } from './AlertDialog';

interface PhotoWithCaption {
  file: File;
  caption: string;
}

interface PhotoDraft extends PhotoWithCaption {
  previewUrl: string;
}

interface ReviewConvertModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: {
    foodRating: number;
    serviceRating: number;
    ambianceRating: number;
    comment: string;
    description: string;
    photos?: PhotoWithCaption[];
  }) => void;
  restaurantName: string;
}

export function ReviewConvertModal({ isOpen, onClose, onSubmit, restaurantName }: ReviewConvertModalProps) {
  const [foodRating, setFoodRating] = useState(0);
  const [serviceRating, setServiceRating] = useState(0);
  const [ambianceRating, setAmbianceRating] = useState(0);
  const [comment, setComment] = useState('');
  const [description, setDescription] = useState('');
  const [photos, setPhotos] = useState<PhotoDraft[]>([]);
  const [alertMessage, setAlertMessage] = useState('');

  // Revoke all preview object URLs on unmount to avoid leaking blob memory
  const photosRef = useRef<PhotoDraft[]>([]);
  photosRef.current = photos;
  useEffect(() => {
    return () => {
      photosRef.current.forEach((photo) => URL.revokeObjectURL(photo.previewUrl));
    };
  }, []);

  const resetForm = () => {
    setFoodRating(0);
    setServiceRating(0);
    setAmbianceRating(0);
    setComment('');
    setDescription('');
    photos.forEach((photo) => URL.revokeObjectURL(photo.previewUrl));
    setPhotos([]);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newPhotos = Array.from(e.target.files).map(file => ({
        file,
        caption: '',
        previewUrl: URL.createObjectURL(file),
      }));
      setPhotos([...photos, ...newPhotos]);
    }
  };

  const removePhoto = (index: number) => {
    URL.revokeObjectURL(photos[index].previewUrl);
    setPhotos(photos.filter((_, i) => i !== index));
  };

  const updateCaption = (index: number, caption: string) => {
    const updated = [...photos];
    updated[index].caption = caption;
    setPhotos(updated);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (foodRating === 0 || serviceRating === 0 || ambianceRating === 0) {
      setAlertMessage('Please provide all ratings');
      return;
    }

    if (photos.some(photo => !photo.caption.trim())) {
      setAlertMessage('Please add a description for all photos');
      return;
    }

    onSubmit({
      foodRating,
      serviceRating,
      ambianceRating,
      comment,
      description,
      photos: photos.length > 0
        ? photos.map(({ file, caption }) => ({ file, caption }))
        : undefined,
    });

    resetForm();
  };

  const handleClose = () => {
    resetForm();
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

          <div className="admin-form-group">
            <label className="admin-label">
              Photos (Optional)
            </label>
            <div className="space-y-3">
              <label className="btn-glass inline-flex items-center gap-2 cursor-pointer">
                <Camera className="w-4 h-4" />
                Add Photos
                <input
                  type="file"
                  accept="image/*"
                  multiple
                  onChange={handleFileChange}
                  className="hidden"
                />
              </label>

              {photos.length > 0 && (
                <div className="space-y-3">
                  {photos.map((photo, index) => (
                    <div key={index} className="card p-3">
                      <div className="flex gap-3">
                        <img
                          src={photo.previewUrl}
                          alt={`Preview ${index + 1}`}
                          className="w-20 h-20 object-cover rounded-sm shrink-0"
                        />
                        <div className="flex-1 space-y-2">
                          <p className="text-sm font-medium text-(--text)">
                            {photo.file.name}
                          </p>
                          <input
                            type="text"
                            value={photo.caption}
                            onChange={(e) => updateCaption(index, e.target.value)}
                            placeholder="Add photo description (required) *"
                            className="input-glass text-sm w-full"
                            required
                          />
                        </div>
                        <button
                          type="button"
                          onClick={() => removePhoto(index)}
                          className="p-2 hover:bg-(--danger-dim) rounded-full transition-colors self-start"
                        >
                          <X className="w-4 h-4 text-(--danger)" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
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
