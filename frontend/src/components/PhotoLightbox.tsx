import { X, ChevronLeft, ChevronRight } from 'lucide-react';
import { useEffect } from 'react';

interface Photo {
  id: number;
  url: string;
  caption: string;
}

interface PhotoLightboxProps {
  photos: Photo[];
  currentIndex: number;
  onClose: () => void;
  onNext: () => void;
  onPrevious: () => void;
}

export function PhotoLightbox({ photos, currentIndex, onClose, onNext, onPrevious }: PhotoLightboxProps) {
  const currentPhoto = photos[currentIndex];

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      if (e.key === 'ArrowLeft') onPrevious();
      if (e.key === 'ArrowRight') onNext();
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose, onNext, onPrevious]);

  return (
    <div className="fixed inset-0 z-50 bg-black/95 flex items-center justify-center">
      {/* Close button */}
      <button
        onClick={onClose}
        className="absolute top-4 right-4 admin-btn-icon bg-(--surface) hover:bg-(--surface-hover) z-10"
        title="Close (Esc)"
      >
        <X className="w-5 h-5" />
      </button>

      {/* Previous button */}
      {photos.length > 1 && (
        <button
          onClick={onPrevious}
          className="absolute left-4 top-1/2 -translate-y-1/2 admin-btn-icon bg-(--surface) hover:bg-(--surface-hover) z-10"
          title="Previous (←)"
        >
          <ChevronLeft className="w-6 h-6" />
        </button>
      )}

      {/* Next button */}
      {photos.length > 1 && (
        <button
          onClick={onNext}
          className="absolute right-4 top-1/2 -translate-y-1/2 admin-btn-icon bg-(--surface) hover:bg-(--surface-hover) z-10"
          title="Next (→)"
        >
          <ChevronRight className="w-6 h-6" />
        </button>
      )}

      {/* Image container */}
      <div className="max-w-7xl max-h-[90vh] w-full h-full flex flex-col items-center justify-center p-8">
        <img
          src={currentPhoto.url}
          alt={currentPhoto.caption}
          className="max-w-full max-h-full object-contain"
          onClick={(e) => e.stopPropagation()}
        />

        {/* Caption */}
        <div className="mt-4 text-center">
          <p className="text-white text-lg font-bold mb-2">{currentPhoto.caption}</p>
          {photos.length > 1 && (
            <p className="text-white/60 font-mono text-sm">
              {currentIndex + 1} / {photos.length}
            </p>
          )}
        </div>
      </div>

      {/* Background overlay - clicking closes */}
      <div
        className="absolute inset-0 -z-10"
        onClick={onClose}
      />
    </div>
  );
}
