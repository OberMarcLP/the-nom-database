import { useEffect, useId } from 'react';
import { AlertTriangle, Globe } from 'lucide-react';
import { useEscapeKey } from '../hooks/useEscapeKey';
import { useFocusTrap } from '../hooks/useFocusTrap';

interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  isDangerous?: boolean;
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = 'OK',
  cancelText = 'Cancel',
  isDangerous = false,
}: ConfirmDialogProps) {
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

  const handleConfirm = () => {
    onConfirm();
    onClose();
  };

  return (
    <div className="admin-confirm-overlay" onClick={onClose}>
      <div
        ref={containerRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={messageId}
        className="admin-confirm-dialog"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="admin-confirm-header">
          <div className="admin-confirm-icon">
            {isDangerous ? <AlertTriangle size={24} /> : <Globe size={24} />}
          </div>
          <h2 className="admin-confirm-title" id={titleId}>{title}</h2>
        </div>
        <div className="admin-confirm-body" id={messageId}>
          {message}
        </div>
        <div className="admin-confirm-footer">
          <button
            type="button"
            className="admin-btn"
            onClick={onClose}
            data-autofocus
          >
            {cancelText}
          </button>
          <button
            type="button"
            className={`admin-btn ${isDangerous ? 'admin-btn-danger' : 'admin-btn-primary'}`}
            onClick={handleConfirm}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}
