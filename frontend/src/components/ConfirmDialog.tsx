import { useEffect } from 'react';
import { AlertTriangle, Globe } from 'lucide-react';

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

  const handleConfirm = () => {
    onConfirm();
    onClose();
  };

  return (
    <div className="admin-confirm-overlay" onClick={onClose}>
      <div className="admin-confirm-dialog" onClick={(e) => e.stopPropagation()}>
        <div className="admin-confirm-header">
          <div className="admin-confirm-icon">
            {isDangerous ? <AlertTriangle size={24} /> : <Globe size={24} />}
          </div>
          <h2 className="admin-confirm-title">{title}</h2>
        </div>
        <div className="admin-confirm-body">
          {message}
        </div>
        <div className="admin-confirm-footer">
          <button
            className="admin-btn"
            onClick={onClose}
          >
            {cancelText}
          </button>
          <button
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
