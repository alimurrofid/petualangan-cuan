import { defineStore } from 'pinia';
import { ref } from 'vue';
import i18n from '@/i18n';

export type SupportedLanguage = 'id' | 'en';

export interface AppSettings {
  language: SupportedLanguage;
  showDecimal: boolean;
}

const STORAGE_KEY = 'cuan_settings';

export const useSettingsStore = defineStore('settings', () => {
  // Load initial settings from localStorage or defaults
  const savedSettings = (() => {
    try {
      const item = localStorage.getItem(STORAGE_KEY);
      return item ? JSON.parse(item) : null;
    } catch {
      return null;
    }
  })();

  const language = ref<SupportedLanguage>(savedSettings?.language === 'en' ? 'en' : 'id');
  const isDecimalActive = savedSettings?.showDecimal === true || savedSettings?.showDecimal === 'true' || savedSettings?.showDecimal === 'Show';
  const showDecimal = ref<boolean>(isDecimalActive);

  const saveSettings = (newSettings: { language: SupportedLanguage; showDecimal: boolean }) => {
    language.value = newSettings.language;
    showDecimal.value = Boolean(newSettings.showDecimal);

    try {
      // Sync i18n locale
      (i18n.global.locale as any).value = newSettings.language;
    } catch (err) {
      console.warn('Could not sync i18n locale:', err);
    }

    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          language: language.value,
          showDecimal: showDecimal.value,
        })
      );
    } catch (e) {
      console.error('Failed to save settings to localStorage:', e);
    }
  };

  const toggleShowDecimal = () => {
    showDecimal.value = !showDecimal.value;
    saveSettings({ language: language.value, showDecimal: showDecimal.value });
  };

  const setLanguage = (lang: SupportedLanguage) => {
    language.value = lang;
    saveSettings({ language: lang, showDecimal: showDecimal.value });
  };

  return {
    language,
    showDecimal,
    saveSettings,
    toggleShowDecimal,
    setLanguage,
  };
});
