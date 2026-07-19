import { useEffect, useId } from 'react';
import { useEscapeKey } from '../hooks/useEscapeKey';
import { useFocusTrap } from '../hooks/useFocusTrap';

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
  useEscapeKey(onClose, isOpen);
  const containerRef = useFocusTrap<HTMLDivElement>(isOpen);
  const titleId = useId();
  const messageId = useId();

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-9999 flex items-center justify-center p-4">
      <div
        className="modal-overlay z-9998"
        onClick={onClose}
      />
      <div
        ref={containerRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={title ? titleId : undefined}
        aria-describedby={messageId}
        className="modal-glass w-full max-w-md relative z-9999"
      >
        <div className="p-6">
          {title && (
            <h3 id={titleId} className="text-lg font-semibold mb-4 text-(--text) flex items-center gap-2">
              <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="12" r="3" fill="currentColor"/>
              </svg>
              {title}
            </h3>
          )}
          <p id={messageId} className="text-(--text-muted) mb-6 wrap-break-word">
            {message}
          </p>
          <div className="flex justify-end">
            <button
              type="button"
              onClick={onClose}
              className="btn-glass-primary min-w-[80px]"
              data-autofocus
            >
              {buttonText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
