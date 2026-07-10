import { describe, it, expect } from 'vitest';
import { stripInlineMarkdown, parseReleaseNotes, summarizeReleaseNotes } from '../changelogFormat';

// A representative semantic-release body for a patch release.
const RELEASE_BODY = [
    '## [1.7.1](https://github.com/Wifsimster/PlexCord/compare/v1.7.0...v1.7.1) (2026-07-10)',
    '',
    '### Bug Fixes',
    '',
    '* **brand:** replace placeholder W icon with PlexCord logo ([#64](https://github.com/Wifsimster/PlexCord/issues/64)) ([d7284c6](https://github.com/Wifsimster/PlexCord/commit/d7284c6))',
    '* stop Plex tile showing Idle while polling is live ([#61](https://github.com/Wifsimster/PlexCord/issues/61))',
    '',
    '### Features',
    '',
    '* **i18n:** add environment-based localization ([b42c137](https://github.com/Wifsimster/PlexCord/commit/b42c137))'
].join('\n');

describe('stripInlineMarkdown', () => {
    it('drops trailing PR and commit reference links', () => {
        expect(
            stripInlineMarkdown('replace icon ([#64](https://x/issues/64)) ([d7284c6](https://x/commit/d7284c6))')
        ).toBe('replace icon');
    });

    it('collapses inline links to their text and strips bold/italic/code', () => {
        expect(stripInlineMarkdown('see [the docs](https://x) for **bold** and `code`')).toBe('see the docs for bold and code');
        expect(stripInlineMarkdown('**scope:** a change')).toBe('scope: a change');
    });

    it('trims a reference fragment left by mid-link truncation', () => {
        expect(stripInlineMarkdown('replace icon ([#64](https://github.com/Wifs')).toBe('replace icon');
    });

    it('handles empty and nullish input', () => {
        expect(stripInlineMarkdown('')).toBe('');
        expect(stripInlineMarkdown(null)).toBe('');
        expect(stripInlineMarkdown(undefined)).toBe('');
    });
});

describe('parseReleaseNotes', () => {
    it('drops the version header and groups bullets under their sections', () => {
        const sections = parseReleaseNotes(RELEASE_BODY);
        expect(sections).toEqual([
            {
                title: 'Bug Fixes',
                items: ['brand: replace placeholder W icon with PlexCord logo', 'stop Plex tile showing Idle while polling is live']
            },
            {
                title: 'Features',
                items: ['i18n: add environment-based localization']
            }
        ]);
    });

    it('collects leading content into an untitled section', () => {
        const sections = parseReleaseNotes('Some intro line\n\n* a bullet');
        expect(sections).toEqual([{ title: '', items: ['Some intro line', 'a bullet'] }]);
    });

    it('omits sections that end up with no items', () => {
        const sections = parseReleaseNotes('### Empty Section\n\n### Bug Fixes\n\n* only item');
        expect(sections).toEqual([{ title: 'Bug Fixes', items: ['only item'] }]);
    });

    it('returns an empty array for empty or nullish input', () => {
        expect(parseReleaseNotes('')).toEqual([]);
        expect(parseReleaseNotes(null)).toEqual([]);
    });
});

describe('summarizeReleaseNotes', () => {
    it('joins parsed items with a bullet separator', () => {
        expect(summarizeReleaseNotes('### Features\n\n* add A\n* add B')).toBe('add A • add B');
    });

    it('truncates on a word boundary with an ellipsis', () => {
        const summary = summarizeReleaseNotes(RELEASE_BODY, 40);
        expect(summary.length).toBeLessThanOrEqual(41);
        expect(summary.endsWith('…')).toBe(true);
        expect(summary).not.toContain('##');
    });
});
