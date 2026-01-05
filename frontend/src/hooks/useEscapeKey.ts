import { useEffect } from 'react';

/**
 * Custom hook to handle ESC key press
 * Useful for closing modals, dialogs, and other dismissible UI elements
 *
 * @param callback - Function to call when ESC key is pressed
 * @param enabled - Whether the hook is enabled (default: true)
 *
 * @example
 * function Modal({ isOpen, onClose }) {
 *   useEscapeKey(onClose, isOpen);
 *   return <div>Modal content</div>;
 * }
 */
export function useEscapeKey(callback: () => void, enabled: boolean = true) {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        callback();
      }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [callback, enabled]);
}
