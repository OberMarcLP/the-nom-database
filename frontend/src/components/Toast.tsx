import { useEffect, useState } from 'react';
import { X, AlertCircle, CheckCircle, Info, AlertTriangle } from 'lucide-react';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface ToastMessage {
  id: string;
  message: string;
  type: ToastType;
  duration?: number;
}

interface ToastProps {
  toast: ToastMessage;
  onClose: (id: string) => void;
}

const Toast = ({ toast, onClose }: ToastProps) => {
  const [isExiting, setIsExiting] = useState(false);

  useEffect(() => {
    const duration = toast.duration || 5000;
    const timer = setTimeout(() => {
      handleClose();
    }, duration);

    return () => clearTimeout(timer);
  }, [toast]);

  const handleClose = () => {
    setIsExiting(true);
    setTimeout(() => {
      onClose(toast.id);
    }, 300); // Match animation duration
  };

  const icons = {
    success: <CheckCircle className="w-5 h-5 text-(--success)" />,
    error: <AlertCircle className="w-5 h-5 text-(--danger)" />,
    warning: <AlertTriangle className="w-5 h-5 text-(--warning)" />,
    info: <Info className="w-5 h-5 text-(--info)" />,
  };

  const bgColors = {
    success: 'bg-(--success)/10 border-(--success)',
    error: 'bg-(--danger)/10 border-(--danger)',
    warning: 'bg-(--warning)/10 border-(--warning)',
    info: 'bg-(--info)/10 border-(--info)',
  };

  return (
    <div
      className={`flex items-start gap-3 p-4 rounded border-2 shadow-lg bg-(--surface) transition-all duration-300 ${
        bgColors[toast.type]
      } ${
        isExiting
          ? 'opacity-0 translate-x-full'
          : 'opacity-100 translate-x-0'
      }`}
    >
      <div className="shrink-0 mt-0.5">{icons[toast.type]}</div>

      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-(--text) wrap-break-word">
          {toast.message}
        </p>
      </div>

      <button
        onClick={handleClose}
        className="shrink-0 text-(--text-muted) hover:text-(--text) transition-colors"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  );
};

interface ToastContainerProps {
  toasts: ToastMessage[];
  onClose: (id: string) => void;
}

export const ToastContainer = ({ toasts, onClose }: ToastContainerProps) => {
  return (
    <div className="fixed top-20 right-4 z-50 flex flex-col gap-2 max-w-sm w-full pointer-events-none">
      {toasts.map((toast) => (
        <div key={toast.id} className="pointer-events-auto">
          <Toast toast={toast} onClose={onClose} />
        </div>
      ))}
    </div>
  );
};
