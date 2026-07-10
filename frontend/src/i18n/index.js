import { createI18n } from 'vue-i18n';
import en from './locales/en.json';
import es from './locales/es.json';
import fr from './locales/fr.json';
import de from './locales/de.json';

/**
 * Application i18n (spec: locale follows the environment).
 *
 * The active locale is resolved once at startup from, in order:
 *   1. an explicit user override persisted in localStorage
 *      ('plexcord-locale', set via the Settings › App language picker), then
 *   2. the environment locale reported by the OS/webview
 *      (navigator.languages / navigator.language), then
 *   3. English as the ultimate fallback.
 *
 * Only the base language subtag is matched (e.g. 'fr-CA' → 'fr'), so a
 * regional environment locale still lands on the closest supported language.
 */

export const LOCALE_STORAGE_KEY = 'plexcord-locale';

export const SUPPORTED_LOCALES = ['en', 'es', 'fr', 'de'];

export const FALLBACK_LOCALE = 'en';

const messages = { en, es, fr, de };

/** Normalise a raw locale tag to a supported base language, or ''. */
function matchSupported(tag) {
    if (!tag) return '';
    const base = String(tag).toLowerCase().split('-')[0];
    return SUPPORTED_LOCALES.includes(base) ? base : '';
}

/** Resolve the startup locale from the persisted override or the environment. */
export function detectLocale() {
    try {
        const saved = matchSupported(localStorage.getItem(LOCALE_STORAGE_KEY));
        if (saved) return saved;
    } catch {
        // localStorage unavailable — fall through to environment detection.
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

/**
 * Switch the active language and persist the choice so it survives restarts
 * (overriding the environment locale on the next launch).
 * @param {string} locale - one of SUPPORTED_LOCALES
 */
export function setLocale(locale) {
    const matched = matchSupported(locale) || FALLBACK_LOCALE;
    i18n.global.locale.value = matched;
    try {
        localStorage.setItem(LOCALE_STORAGE_KEY, matched);
    } catch {
        // Persisting is best-effort; the in-memory switch still applies.
    }
    if (typeof document !== 'undefined') {
        document.documentElement.setAttribute('lang', matched);
    }
}

/** Translate outside of a component setup scope (plain JS modules). */
export function t(...args) {
    return i18n.global.t(...args);
}

if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('lang', i18n.global.locale.value);
}

export default i18n;
