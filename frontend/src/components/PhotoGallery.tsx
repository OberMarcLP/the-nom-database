import { useState } from 'react';
import { MenuPhoto } from '../services/api';
import { Trash2, Edit2, Check, X, Star, User } from 'lucide-react';
import { ConfirmDialog } from './ConfirmDialog';
import { AlertDialog } from './AlertDialog';
import { LazyImage } from './LazyImage';

interface ExtendedPhoto extends MenuPhoto {
  source?: 'menu' | 'review';
  reviewInfo?: {
    username: string;
    date: string;
    ratings: {
      food: number;
      service: number;
      ambiance: number;
    };
  } | null;
}

interface PhotoGalleryProps {
  photos: ExtendedPhoto[];
  onCaptionUpdate: (id: number, caption: string) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
}

export function PhotoGallery({ photos, onCaptionUpdate, onDelete }: PhotoGalleryProps) {
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editCaption, setEditCaption] = useState('');
  const [deletingPhotoId, setDeletingPhotoId] = useState<number | null>(null);
  const [alertMessage, setAlertMessage] = useState('');

  const handleStartEdit = (photo: MenuPhoto) => {
    setEditingId(photo.id);
    setEditCaption(photo.caption);
  };

  const handleCancelEdit = () => {
    setEditingId(null);
    setEditCaption('');
  };

  const handleSaveEdit = async (id: number) => {
    if (!editCaption.trim()) {
      setAlertMessage('Caption cannot be empty');
      return;
    }

    try {
      await onCaptionUpdate(id, editCaption);
      setEditingId(null);
      setEditCaption('');
    } catch (error) {
      setAlertMessage('Failed to update caption');
    }
  };

  const handleDelete = async (id: number) => {
    setDeletingPhotoId(id);
  };

  const confirmDelete = async () => {
    if (!deletingPhotoId) return;
    try {
      await onDelete(deletingPhotoId);
      setDeletingPhotoId(null);
    } catch (error) {
      setAlertMessage('Failed to delete photo');
      setDeletingPhotoId(null);
    }
  };

  if (photos.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        No photos uploaded yet
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {photos.map((photo) => (
        <div
          key={photo.id}
          className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden hover:shadow-lg transition-shadow bg-white dark:bg-gray-800"
        >
          <div className="relative group">
            <LazyImage
              src={photo.url}
              alt={photo.caption}
              className="w-full h-48 object-cover"
              onError={(e) => {
                e.currentTarget.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="100" height="100"%3E%3Crect fill="%23ddd" width="100" height="100"/%3E%3Ctext fill="%23999" x="50%" y="50%" text-anchor="middle" dy=".3em"%3ENo Image%3C/text%3E%3C/svg%3E';
              }}
            />
            <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-2">
              <button
                onClick={() => handleStartEdit(photo)}
                className="p-2 bg-blue-500 hover:bg-blue-600 text-white rounded-full shadow-lg"
                title="Edit caption"
              >
                <Edit2 className="w-4 h-4" />
              </button>
              <button
                onClick={() => handleDelete(photo.id)}
                className="p-2 bg-red-500 hover:bg-red-600 text-white rounded-full shadow-lg"
                title="Delete photo"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div className="p-3">
            {editingId === photo.id ? (
              <div className="space-y-2">
                <input
                  type="text"
                  value={editCaption}
                  onChange={(e) => setEditCaption(e.target.value)}
                  className="input text-sm py-1"
                  placeholder="Dish name"
                  autoFocus
                />
                <div className="flex gap-2">
                  <button
                    onClick={() => handleSaveEdit(photo.id)}
                    className="flex-1 btn btn-sm bg-green-500 hover:bg-green-600 text-white flex items-center justify-center gap-1"
                  >
                    <Check className="w-4 h-4" />
                    Save
                  </button>
                  <button
                    onClick={handleCancelEdit}
                    className="flex-1 btn btn-sm bg-gray-500 hover:bg-gray-600 text-white flex items-center justify-center gap-1"
                  >
                    <X className="w-4 h-4" />
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <>
                <h3 className="font-semibold text-sm mb-1">{photo.caption}</h3>

                {photo.source === 'review' && photo.reviewInfo ? (
                  <div className="space-y-1 mt-2 pt-2 border-t border-gray-200 dark:border-gray-700">
                    <div className="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-400">
                      <User className="w-3 h-3" />
                      <span className="font-medium">{photo.reviewInfo.username}</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                      <span className="text-gray-500 dark:text-gray-400">Food:</span>
                      <div className="flex items-center gap-0.5">
                        {[...Array(5)].map((_, i) => (
                          <Star
                            key={i}
                            className={`w-3 h-3 ${
                              i < photo.reviewInfo!.ratings.food
                                ? 'fill-yellow-400 text-yellow-400'
                                : 'text-gray-300 dark:text-gray-600'
                            }`}
                          />
                        ))}
                      </div>
                    </div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {new Date(photo.reviewInfo.date).toLocaleDateString()}
                    </p>
                  </div>
                ) : (
                  <>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {new Date(photo.created_at).toLocaleDateString()}
                    </p>
                    {photo.file_size && (
                      <p className="text-xs text-gray-400 dark:text-gray-500">
                        {(photo.file_size / 1024).toFixed(1)} KB
                      </p>
                    )}
                  </>
                )}
              </>
            )}
          </div>
        </div>
      ))}
      <ConfirmDialog
        isOpen={deletingPhotoId !== null}
        onClose={() => setDeletingPhotoId(null)}
        onConfirm={confirmDelete}
        title="Delete Photo"
        message="Are you sure you want to delete this photo?"
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
