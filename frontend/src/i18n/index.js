import { createI18n } from 'vue-i18n';
import en from './locales/en.json';
import es from './locales/es.json';
import fr from './locales/fr.json';
import de from './locales/de.json';

export const LOCALE_STORAGE_KEY = 'plexcord-locale';

export const SUPPORTED_LOCALES = ['en', 'es', 'fr', 'de'];

export const FALLBACK_LOCALE = 'en';

const messages = { en, es, fr, de };

function matchSupported(tag) {
    if (!tag) return '';
    const base = String(tag).toLowerCase().split('-')[0];
    return SUPPORTED_LOCALES.includes(base) ? base : '';
}

export function detectLocale() {
    try {
        const saved = matchSupported(localStorage.getItem(LOCALE_STORAGE_KEY));
        if (saved) return saved;
    } catch {
    }

    const candidates = [];
    if (typeof navigator !== 'undefined') {
        if (Array.isArray(navigator.languages)) candidates.push(...navigator.languages);
        if (navigator.language) candidates.push(navigator.language);
    }
    for (const candidate of candidates) {
        const matched = matchSupported(candidate);
        if (matched) return matched;
    }

    return FALLBACK_LOCALE;
}

const i18n = createI18n({
    legacy: false,
    globalInjection: true,
    locale: detectLocale(),
    fallbackLocale: FALLBACK_LOCALE,
    messages
});

export function setLocale(locale) {
    const matched = matchSupported(locale) || FALLBACK_LOCALE;
    i18n.global.locale.value = matched;
    try {
        localStorage.setItem(LOCALE_STORAGE_KEY, matched);
    } catch {
    }
    if (typeof document !== 'undefined') {
        document.documentElement.setAttribute('lang', matched);
    }
}

export function t(...args) {
    return i18n.global.t(...args);
}

if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('lang', i18n.global.locale.value);
}

export default i18n;
