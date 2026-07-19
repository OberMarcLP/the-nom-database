import { X } from 'lucide-react';
import { ReactNode, useEffect, useId } from 'react';
import { useEscapeKey } from '../hooks/useEscapeKey';
import { useFocusTrap } from '../hooks/useFocusTrap';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export function Modal({ isOpen, onClose, title, children }: ModalProps) {
  useEscapeKey(onClose, isOpen);
  const containerRef = useFocusTrap<HTMLDivElement>(isOpen);
  const titleId = useId();

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
      <div
        ref={containerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="modal-glass"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="admin-modal-header">
          <h2 className="admin-modal-title" id={titleId}>{title}</h2>
          <button
            onClick={onClose}
            className="admin-modal-close"
            aria-label="Close dialog"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="admin-modal-body">{children}</div>
      </div>
    </div>
  );
}
