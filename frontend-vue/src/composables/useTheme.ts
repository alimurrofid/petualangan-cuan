import { useDark, useToggle } from '@vueuse/core';

export const isDark = useDark({
  storageKey: 'theme',
  valueDark: 'dark',
  valueLight: '',
});

export const toggleTheme = useToggle(isDark);
