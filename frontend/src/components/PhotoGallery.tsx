import { useState, useEffect, useRef } from 'react';
import { MenuPhoto } from '../services/api';
import { Trash2, Edit2, Check, X, Star, User } from 'lucide-react';
import { ConfirmDialog } from './ConfirmDialog';
import { AlertDialog } from './AlertDialog';
import { LazyImage } from './LazyImage';
import { PhotoLightbox } from './PhotoLightbox';

interface ExtendedPhoto extends MenuPhoto {
  source?: 'menu' | 'review';
  originalId?: number;
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
  highlightedPhotoId?: number;
}

export function PhotoGallery({ photos, onCaptionUpdate, onDelete, highlightedPhotoId }: PhotoGalleryProps) {
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editCaption, setEditCaption] = useState('');
  const [deletingPhotoId, setDeletingPhotoId] = useState<number | null>(null);
  const [alertMessage, setAlertMessage] = useState('');
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [lightboxIndex, setLightboxIndex] = useState(0);
  const photoRefs = useRef<{ [key: number]: HTMLDivElement | null }>({});

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

  const openLightbox = (index: number) => {
    setLightboxIndex(index);
    setLightboxOpen(true);
  };

  const closeLightbox = () => {
    setLightboxOpen(false);
  };

  const nextPhoto = () => {
    setLightboxIndex((prev) => (prev + 1) % photos.length);
  };

  const previousPhoto = () => {
    setLightboxIndex((prev) => (prev - 1 + photos.length) % photos.length);
  };

  // Scroll to highlighted photo
  useEffect(() => {
    if (highlightedPhotoId) {
      // Find the photo by originalId and get its actual id
      const targetPhoto = photos.find(p => p.originalId === highlightedPhotoId || p.id === highlightedPhotoId);
      if (targetPhoto && photoRefs.current[targetPhoto.id]) {
        setTimeout(() => {
          photoRefs.current[targetPhoto.id]?.scrollIntoView({
            behavior: 'smooth',
            block: 'center',
          });
        }, 400); // Delay to ensure modal and photos are rendered
      }
    }
  }, [highlightedPhotoId, photos]);

  if (photos.length === 0) {
    return (
      <div className="text-center py-12 text-[var(--text-muted)] font-mono text-sm">
        NO PHOTOS UPLOADED YET
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {photos.map((photo) => (
        <div
          key={photo.id}
          ref={(el) => (photoRefs.current[photo.id] = el)}
          className={`relative overflow-hidden flex flex-col h-full border-2 bg-[var(--surface)] rounded-lg transition-all ${
            highlightedPhotoId === photo.originalId || highlightedPhotoId === photo.id
              ? 'border-[var(--accent)] shadow-[0_0_20px_var(--accent-dim)] scale-[1.02]'
              : 'border-[var(--border)]'
          }`}
        >
          <div className="relative group flex-1 cursor-pointer" onClick={() => openLightbox(photos.indexOf(photo))}>
            <LazyImage
              src={photo.url}
              alt={photo.caption}
              className="w-full h-full min-h-[400px] object-cover"
              onError={(e) => {
                e.currentTarget.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="100" height="100"%3E%3Crect fill="%23141414" width="100" height="100"/%3E%3Ctext fill="%23888" x="50%" y="50%" text-anchor="middle" dy=".3em"%3ENo Image%3C/text%3E%3C/svg%3E';
              }}
            />
            <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity flex gap-2 z-10">
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  handleStartEdit(photo);
                }}
                className="admin-btn-icon"
                title="Edit caption"
              >
                <Edit2 className="w-4 h-4" />
              </button>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  handleDelete(photo.id);
                }}
                className="admin-btn-icon bg-[var(--danger)] hover:bg-[var(--danger)]"
                title="Delete photo"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>

            {/* Overlay caption and info on the image */}
            {editingId !== photo.id && (
              <div className="absolute inset-0 bg-gradient-to-t from-black/90 via-black/50 to-transparent flex flex-col justify-end p-4">
                <h3 className="text-white font-bold text-base mb-2">{photo.caption}</h3>

                {photo.source === 'review' && photo.reviewInfo ? (
                  <div className="space-y-1.5">
                    <div className="flex items-center gap-1.5 text-xs">
                      <User className="w-3.5 h-3.5 text-[var(--accent)]" />
                      <span className="text-white font-mono">{photo.reviewInfo.username}</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                      <span className="text-white/90 font-mono font-semibold">FOOD:</span>
                      <div className="flex items-center gap-0.5">
                        {[...Array(5)].map((_, i) => (
                          <Star
                            key={i}
                            className={`w-3.5 h-3.5 ${
                              i < photo.reviewInfo!.ratings.food
                                ? 'fill-[var(--accent)] text-[var(--accent)]'
                                : 'text-white/20'
                            }`}
                          />
                        ))}
                      </div>
                    </div>
                    <p className="text-xs text-white/80 font-mono">
                      {new Date(photo.reviewInfo.date).toLocaleDateString()}
                    </p>
                  </div>
                ) : (
                  <p className="text-xs text-white/80 font-mono">
                    {new Date(photo.created_at).toLocaleDateString()}
                  </p>
                )}
              </div>
            )}
          </div>

          {editingId === photo.id && (
            <div className="p-4 space-y-3 bg-[var(--surface)]">
              <input
                type="text"
                value={editCaption}
                onChange={(e) => setEditCaption(e.target.value)}
                className="admin-input text-sm"
                placeholder="Dish name"
                autoFocus
              />
              <div className="flex gap-2">
                <button
                  onClick={() => handleSaveEdit(photo.id)}
                  className="flex-1 admin-btn-sm admin-btn-primary flex items-center justify-center gap-1"
                >
                  <Check className="w-4 h-4" />
                  Save
                </button>
                <button
                  onClick={handleCancelEdit}
                  className="flex-1 admin-btn-sm flex items-center justify-center gap-1"
                >
                  <X className="w-4 h-4" />
                  Cancel
                </button>
              </div>
            </div>
          )}
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
        isDangerous
      />
      <AlertDialog
        isOpen={alertMessage !== ''}
        onClose={() => setAlertMessage('')}
        message={alertMessage}
      />
      {lightboxOpen && (
        <PhotoLightbox
          photos={photos.map(p => ({ id: p.id, url: p.url, caption: p.caption }))}
          currentIndex={lightboxIndex}
          onClose={closeLightbox}
          onNext={nextPhoto}
          onPrevious={previousPhoto}
        />
      )}
    </div>
  );
}
