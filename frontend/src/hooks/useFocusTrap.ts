import { useEffect, useRef } from 'react';

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(', ');

/**
 * Custom hook that traps keyboard focus inside a container while active.
 * Moves focus into the container on activation (preferring an element
 * marked with `data-autofocus`), cycles Tab/Shift+Tab within it, and
 * restores focus to the previously focused element on deactivation.
 *
 * @param active - Whether the trap is active (e.g. modal is open)
 * @returns Ref to attach to the container element
 *
 * @example
 * function Modal({ isOpen }) {
 *   const containerRef = useFocusTrap<HTMLDivElement>(isOpen);
 *   return <div ref={containerRef} role="dialog">...</div>;
 * }
 */
export function useFocusTrap<T extends HTMLElement>(active: boolean) {
  const containerRef = useRef<T | null>(null);

  useEffect(() => {
    if (!active) return;

    const container = containerRef.current;
    if (!container) return;

    const previouslyFocused = document.activeElement as HTMLElement | null;

    const getFocusable = () =>
      Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
        (el) => el.offsetParent !== null || el === document.activeElement
      );

    // Move focus into the dialog: explicit data-autofocus wins, then the
    // first focusable element, then the container itself as a fallback.
    const preferred = container.querySelector<HTMLElement>('[data-autofocus]');
    const initial = preferred || getFocusable()[0];
    if (initial) {
      initial.focus();
    } else {
      container.tabIndex = -1;
      container.focus();
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return;

      const focusable = getFocusable();
      if (focusable.length === 0) {
        event.preventDefault();
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const current = document.activeElement;
      const outside = !container.contains(current);

      if (event.shiftKey) {
        if (current === first || outside) {
          event.preventDefault();
          last.focus();
        }
      } else if (current === last || outside) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener('keydown', handleKeyDown, true);

    return () => {
      document.removeEventListener('keydown', handleKeyDown, true);
      previouslyFocused?.focus();
    };
  }, [active]);

  return containerRef;
}
