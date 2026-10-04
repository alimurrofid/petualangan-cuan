import { createI18n } from 'vue-i18n';
import id from '@/locales/id.json';
import en from '@/locales/en.json';

export const getSavedLocale = (): 'id' | 'en' => {
  try {
    const raw = localStorage.getItem('cuan_settings');
    if (raw) {
      const parsed = JSON.parse(raw);
      if (parsed.language === 'en' || parsed.language === 'id') {
        return parsed.language;
      }
    }
  } catch {
    // ignore
  }
  return 'id';
};

const i18n = createI18n({
  legacy: false,
  locale: getSavedLocale(),
  fallbackLocale: 'id',
  messages: {
    id,
    en,
  },
});

export default i18n;
