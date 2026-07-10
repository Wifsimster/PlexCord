/**
 * Time Utility Functions
 * Shared utilities for time formatting across the application.
 */
import { t } from '@/i18n';

/**
 * Format a timestamp as relative time (e.g., "5 minutes ago")
 * @param {string|Date|null} timestamp - ISO timestamp or Date object
 * @returns {string} Formatted relative time string
 */
export function formatRelativeTime(timestamp) {
    if (!timestamp) return t('time.never');

    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return t('time.justNow');
    if (diffMin < 60) return t('time.minutesAgo', { n: diffMin }, diffMin);
    if (diffHour < 24) return t('time.hoursAgo', { n: diffHour }, diffHour);
    if (diffDay < 7) return t('time.daysAgo', { n: diffDay }, diffDay);

    return date.toLocaleDateString();
}

/**
 * Format duration in milliseconds to human-readable string
 * @param {number} ms - Duration in milliseconds
 * @returns {string} Formatted duration (e.g., "2h 30m")
 */
export function formatDuration(ms) {
    if (!ms || ms < 0) return '0s';

    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
        const remainingMinutes = minutes % 60;
        return remainingMinutes > 0 ? `${hours}h ${remainingMinutes}m` : `${hours}h`;
    }

    if (minutes > 0) {
        const remainingSeconds = seconds % 60;
        return remainingSeconds > 0 ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`;
    }

    return `${seconds}s`;
}
