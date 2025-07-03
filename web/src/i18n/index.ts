import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

// Import translations
import translationEN from './locales/en/translation.json';
import translationRU from './locales/ru/translation.json';

declare const process: {
  env: {
    NODE_ENV: string;
  };
};

const resources = {
  en: {
    translation: translationEN,
  },
  ru: {
    translation: translationRU,
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    debug: process.env.NODE_ENV === 'development',
    interpolation: {
      escapeValue: false,
    },
    detection: {
      // Define the order of detection methods
      order: ['localStorage', 'navigator', 'htmlTag'],
      // Keys for localStorage detection
      lookupLocalStorage: 'i18nextLng',
      // Cache user language
      caches: ['localStorage'],
      // Only use localStorage and browser detection
      excludeCacheFor: ['cimode'],
    },
  });

export default i18n; 