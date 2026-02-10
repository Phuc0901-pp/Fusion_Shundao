// i18n configuration for multi-language support
// To enable: npm install react-i18next i18next

import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

// Import translation files
import vi from './locales/vi.json';
import en from './locales/en.json';

i18n
    .use(initReactI18next)
    .init({
        resources: {
            vi: { translation: vi },
            en: { translation: en }
        },
        lng: 'vi', // Default language
        fallbackLng: 'vi',
        interpolation: {
            escapeValue: false // React already escapes
        }
    });

export default i18n;
