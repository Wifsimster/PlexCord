/**
 * Plex server URL format validation (spec §5.3/§5.4).
 *
 * Extracted from the setup wizard's Plex step so the Settings "Add server"
 * dialog and the wizard's manual-entry field share one validator. Format
 * check only — reachability is verified separately via
 * ValidatePlexConnection.
 */

/** Canonical placeholder for Plex server URL inputs (spec §5.4). */
export const PLEX_URL_PLACEHOLDER = 'http://192.168.1.10:32400';

/**
 * Validate a Plex server URL's format.
 *
 * @param {string} url - The URL to validate (leading/trailing whitespace ignored)
 * @returns {{ valid: boolean, error: string }} error is '' when valid
 */
export function validatePlexServerUrl(url) {
    if (!url || url.trim().length === 0) {
        return { valid: false, error: 'Server URL is required' };
    }

    const trimmedUrl = url.trim();

    // Check for protocol
    if (!trimmedUrl.startsWith('http://') && !trimmedUrl.startsWith('https://')) {
        return { valid: false, error: 'URL must start with http:// or https://' };
    }

    // Parse URL to validate structure
    try {
        const urlObj = new URL(trimmedUrl);

        // Validate port if present
        if (urlObj.port) {
            const portNum = parseInt(urlObj.port, 10);
            if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
                return { valid: false, error: 'Port must be between 1 and 65535' };
            }
        }

        // Validate hostname (IP or domain)
        if (!urlObj.hostname || urlObj.hostname.length === 0) {
            return { valid: false, error: 'Invalid hostname or IP address' };
        }

        return { valid: true, error: '' };
    } catch {
        return { valid: false, error: 'Invalid URL format' };
    }
}
