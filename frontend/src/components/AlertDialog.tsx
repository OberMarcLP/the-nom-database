import { useEffect } from 'react';

interface AlertDialogProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  message: string;
  buttonText?: string;
}

export function AlertDialog({
  isOpen,
  onClose,
  title = 'Alert',
  message,
  buttonText = 'OK',
}: AlertDialogProps) {
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-9999 flex items-center justify-center p-4">
      <div
        className="modal-overlay z-9998"
        onClick={onClose}
      />
      <div className="modal-glass w-full max-w-md relative z-9999">
        <div className="p-6">
          {title && (
            <h3 className="text-lg font-semibold mb-4 text-(--text) flex items-center gap-2">
              <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="12" r="3" fill="currentColor"/>
              </svg>
              {title}
            </h3>
          )}
          <p className="text-(--text-muted) mb-6 wrap-break-word">
            {message}
          </p>
          <div className="flex justify-end">
            <button
              onClick={onClose}
              className="btn-glass-primary min-w-[80px]"
              autoFocus
            >
              {buttonText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
