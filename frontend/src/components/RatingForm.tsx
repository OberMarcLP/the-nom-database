import { useEffect, useRef, useState } from 'react';
import { Camera, X } from 'lucide-react';
import { StarRating } from './StarRating';
import { useToast } from '../hooks/useToast';

interface PhotoWithCaption {
  file: File;
  caption: string;
}

interface PhotoDraft extends PhotoWithCaption {
  previewUrl: string;
}

interface RatingFormProps {
  onSubmit: (data: {
    food_rating: number;
    service_rating: number;
    ambiance_rating: number;
    comment?: string;
    photos?: PhotoWithCaption[];
  }) => void;
  onCancel: () => void;
}

export function RatingForm({ onSubmit, onCancel }: RatingFormProps) {
  const [foodRating, setFoodRating] = useState(0);
  const [serviceRating, setServiceRating] = useState(0);
  const [ambianceRating, setAmbianceRating] = useState(0);
  const [comment, setComment] = useState('');
  const [photos, setPhotos] = useState<PhotoDraft[]>([]);
  const { showWarning } = useToast();

  // Revoke all preview object URLs on unmount to avoid leaking blob memory
  const photosRef = useRef<PhotoDraft[]>([]);
  photosRef.current = photos;
  useEffect(() => {
    return () => {
      photosRef.current.forEach((photo) => URL.revokeObjectURL(photo.previewUrl));
    };
  }, []);

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
      return;
    }

    // Validate that all photos have captions
    if (photos.some(photo => !photo.caption.trim())) {
      showWarning('Please add a description for all photos');
      return;
    }

    onSubmit({
      food_rating: foodRating,
      service_rating: serviceRating,
      ambiance_rating: ambianceRating,
      comment: comment || undefined,
      photos: photos.length > 0
        ? photos.map(({ file, caption }) => ({ file, caption }))
        : undefined,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label className="label mb-2">Food Rating *</label>
        <StarRating rating={foodRating} onRatingChange={setFoodRating} size="lg" />
      </div>

      <div>
        <label className="label mb-2">Service Rating *</label>
        <StarRating rating={serviceRating} onRatingChange={setServiceRating} size="lg" />
      </div>

      <div>
        <label className="label mb-2">Ambiance Rating *</label>
        <StarRating rating={ambianceRating} onRatingChange={setAmbianceRating} size="lg" />
      </div>

      <div>
        <label className="label">Comment (optional)</label>
        <textarea
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          className="input-glass min-h-[100px]"
          placeholder="Share your experience..."
          rows={3}
        />
      </div>

      <div>
        <label className="label mb-2">Photos (optional)</label>
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

      <div className="flex gap-3">
        <button
          type="submit"
          disabled={foodRating === 0 || serviceRating === 0 || ambianceRating === 0}
          className="btn-glass-primary flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Submit Rating
        </button>
        <button type="button" onClick={onCancel} className="btn-glass">
          Cancel
        </button>
      </div>
    </form>
  );
}
