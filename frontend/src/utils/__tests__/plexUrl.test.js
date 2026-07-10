import { describe, it, expect } from 'vitest';
import { validatePlexServerUrl, PLEX_URL_PLACEHOLDER } from '../plexUrl';

describe('validatePlexServerUrl', () => {
    it('rejects empty or whitespace-only input', () => {
        expect(validatePlexServerUrl('')).toEqual({ valid: false, error: 'Server URL is required' });
        expect(validatePlexServerUrl('   ')).toEqual({ valid: false, error: 'Server URL is required' });
        expect(validatePlexServerUrl(null)).toEqual({ valid: false, error: 'Server URL is required' });
        expect(validatePlexServerUrl(undefined).valid).toBe(false);
    });

    it('requires an http or https protocol', () => {
        expect(validatePlexServerUrl('192.168.1.10:32400')).toEqual({ valid: false, error: 'URL must start with http:// or https://' });
        expect(validatePlexServerUrl('ftp://plex.local:32400').valid).toBe(false);
        expect(validatePlexServerUrl('plex.local').valid).toBe(false);
    });

    it('rejects malformed URLs', () => {
        expect(validatePlexServerUrl('http://')).toEqual({ valid: false, error: 'Invalid URL format' });
        expect(validatePlexServerUrl('https://:32400').valid).toBe(false);
    });

    it('validates the port range when a port is present', () => {
        expect(validatePlexServerUrl('http://plex.local:0')).toEqual({ valid: false, error: 'Port must be between 1 and 65535' });
        expect(validatePlexServerUrl('http://plex.local:70000').valid).toBe(false);
        expect(validatePlexServerUrl('http://plex.local:65535').valid).toBe(true);
        expect(validatePlexServerUrl('http://plex.local:1').valid).toBe(true);
    });

    it('accepts typical Plex server URLs', () => {
        expect(validatePlexServerUrl('http://192.168.1.100:32400')).toEqual({ valid: true, error: '' });
        expect(validatePlexServerUrl('http://plex.local:32400').valid).toBe(true);
        expect(validatePlexServerUrl('https://plex.example.com:32400').valid).toBe(true);
        expect(validatePlexServerUrl('http://localhost:32400').valid).toBe(true);
        expect(validatePlexServerUrl(PLEX_URL_PLACEHOLDER).valid).toBe(true);
    });

    it('accepts URLs without an explicit port', () => {
        expect(validatePlexServerUrl('https://plex.example.com').valid).toBe(true);
    });

    it('trims surrounding whitespace before validating', () => {
        expect(validatePlexServerUrl('  http://192.168.1.100:32400  ').valid).toBe(true);
    });
});
