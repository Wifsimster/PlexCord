import { describe, it, expect } from 'vitest';
import { renderPresenceFormat, renderPresenceLines } from '../presenceFormat';

const track = {
    sessionKey: '1',
    track: 'Bohemian Rhapsody',
    artist: 'Queen',
    album: 'A Night at the Opera',
    playerName: 'Plexamp'
};

describe('renderPresenceFormat', () => {
    it('replaces {track}, {artist}, {album} and {player} tokens', () => {
        expect(renderPresenceFormat('{track} by {artist} on {album} via {player}', track)).toBe('Bohemian Rhapsody by Queen on A Night at the Opera via Plexamp');
    });

    it('substitutes {year} when the track has one', () => {
        expect(renderPresenceFormat('{album} ({year})', { ...track, year: 1975 })).toBe('A Night at the Opera (1975)');
    });

    it('accepts a string year', () => {
        expect(renderPresenceFormat('{year}', { ...track, year: '1975' })).toBe('1975');
    });

    it('strips the {year} token when the track has no year', () => {
        expect(renderPresenceFormat('{track} {year}', track)).toBe('Bohemian Rhapsody ');
    });

    it('strips the {year} token when year is 0', () => {
        expect(renderPresenceFormat('{year}', { ...track, year: 0 })).toBe('');
    });

    it('leaves unknown tokens untouched (backend parity)', () => {
        expect(renderPresenceFormat('{track} {bogus}', track)).toBe('Bohemian Rhapsody {bogus}');
    });

    it('handles repeated tokens', () => {
        expect(renderPresenceFormat('{artist} — {artist}', track)).toBe('Queen — Queen');
    });

    it('supports the MediaSession shape (title instead of track)', () => {
        expect(renderPresenceFormat('{track}', { title: 'Midnight City' })).toBe('Midnight City');
    });

    it('renders literal text without tokens as-is', () => {
        expect(renderPresenceFormat('Listening on Plex', track)).toBe('Listening on Plex');
    });

    it('returns an empty string for a missing format', () => {
        expect(renderPresenceFormat('', track)).toBe('');
        expect(renderPresenceFormat(null, track)).toBe('');
        expect(renderPresenceFormat(undefined, track)).toBe('');
    });

    it('returns an empty string for a missing track', () => {
        expect(renderPresenceFormat('{track}', null)).toBe('');
    });

    it('replaces tokens for missing fields with empty strings', () => {
        expect(renderPresenceFormat('{track} by {artist}', { track: 'Solo' })).toBe('Solo by ');
    });
});

describe('renderPresenceLines', () => {
    it('renders custom formats when provided', () => {
        expect(renderPresenceLines({ details: '{track}', state: '{artist} · {album}' }, track)).toEqual({
            details: 'Bohemian Rhapsody',
            state: 'Queen · A Night at the Opera'
        });
    });

    it('renders a partial custom format without falling back', () => {
        expect(renderPresenceLines({ details: '{track}', state: '' }, track)).toEqual({
            details: 'Bohemian Rhapsody',
            state: ''
        });
    });

    it('falls back to the backend default lines when no formats are set', () => {
        expect(renderPresenceLines(null, track)).toEqual({
            details: 'Bohemian Rhapsody',
            state: 'by Queen • A Night at the Opera'
        });
        expect(renderPresenceLines({ details: '', state: '' }, track)).toEqual({
            details: 'Bohemian Rhapsody',
            state: 'by Queen • A Night at the Opera'
        });
    });

    it('omits the album from the default state line when missing', () => {
        expect(renderPresenceLines(null, { track: 'Solo', artist: 'Someone' })).toEqual({
            details: 'Solo',
            state: 'by Someone'
        });
    });

    it('returns empty lines when there is no track', () => {
        expect(renderPresenceLines({ details: '{track}', state: '{artist}' }, null)).toEqual({ details: '', state: '' });
    });
});
