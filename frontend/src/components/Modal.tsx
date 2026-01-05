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
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="modal-overlay"
        onClick={onClose}
      />
      <div className="modal-glass w-full max-w-2xl max-h-[90vh] overflow-y-auto m-4 shadow-2xl shadow-black/20">
        <div className="sticky top-0 bg-gradient-to-r from-white/95 to-white/90 dark:from-gray-800/95 dark:to-gray-800/90 backdrop-blur-xl flex items-center justify-between px-6 py-4 border-b border-white/20 dark:border-white/10">
          <h2 className="text-xl font-semibold">{title}</h2>
          <button
            onClick={onClose}
            className="btn-glass p-2 rounded-full hover:bg-red-500/20 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
}
