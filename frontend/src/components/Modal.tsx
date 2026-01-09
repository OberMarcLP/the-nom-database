import { X } from 'lucide-react';
import { ReactNode, useEffect } from 'react';
import { useEscapeKey } from '../hooks/useEscapeKey';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export function Modal({ isOpen, onClose, title, children }: ModalProps) {
  useEscapeKey(onClose, isOpen);

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
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-glass" onClick={(e) => e.stopPropagation()}>
        <div className="admin-modal-header">
          <h2 className="admin-modal-title">{title}</h2>
          <button
            onClick={onClose}
            className="admin-modal-close"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="admin-modal-body">{children}</div>
      </div>
    </div>
  );
}
