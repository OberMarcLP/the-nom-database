import { useEffect, useState } from 'react';

export type ThemePreference = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'nomdb-theme';
const media = window.matchMedia('(prefers-color-scheme: dark)');

function storedPreference(): ThemePreference {
  const value = localStorage.getItem(STORAGE_KEY);
  return value === 'light' || value === 'dark' ? value : 'system';
}

function resolve(preference: ThemePreference): 'light' | 'dark' {
  return preference === 'system' ? (media.matches ? 'dark' : 'light') : preference;
}

function apply(preference: ThemePreference) {
  document.documentElement.classList.toggle('dark', resolve(preference) === 'dark');
}

/**
 * White Table theme preference: follows the OS by default ("system"),
 * explicit choices are persisted in localStorage. The inline script in
 * index.html applies the same logic before first paint (no FOUC).
 */
export function useTheme() {
  const [preference, setPreference] = useState<ThemePreference>(storedPreference);

  useEffect(() => {
    apply(preference);
    if (preference === 'system') {
      const onChange = () => apply('system');
      media.addEventListener('change', onChange);
      return () => media.removeEventListener('change', onChange);
    }
  }, [preference]);

  const setTheme = (next: ThemePreference) => {
    if (next === 'system') {
      localStorage.removeItem(STORAGE_KEY);
    } else {
      localStorage.setItem(STORAGE_KEY, next);
    }
    setPreference(next);
  };

  return { preference, resolved: resolve(preference), setTheme };
}
