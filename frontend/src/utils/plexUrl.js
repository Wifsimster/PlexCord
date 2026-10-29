import { t } from '@/i18n';

export const PLEX_URL_PLACEHOLDER = 'http://192.168.1.10:32400';

export function validatePlexServerUrl(url) {
    if (!url || url.trim().length === 0) {
        return { valid: false, error: t('validation.urlRequired') };
    }

    const trimmedUrl = url.trim();

    if (!trimmedUrl.startsWith('http://') && !trimmedUrl.startsWith('https://')) {
        return { valid: false, error: t('validation.urlProtocol') };
    }

    try {
        const urlObj = new URL(trimmedUrl);

        if (urlObj.port) {
            const portNum = parseInt(urlObj.port, 10);
            if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
                return { valid: false, error: t('validation.urlPort') };
            }
        }

        if (!urlObj.hostname || urlObj.hostname.length === 0) {
            return { valid: false, error: t('validation.urlHostname') };
        }

        return { valid: true, error: '' };
    } catch {
        return { valid: false, error: t('validation.urlFormat') };
    }
}
