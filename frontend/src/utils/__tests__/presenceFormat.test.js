import { describe, it, expect } from 'vitest';
import { mediaTypeOf, renderPresenceFormat, renderPresenceLines } from '../presenceFormat';

const track = {
    sessionKey: '1',
    track: 'Bohemian Rhapsody',
    artist: 'Queen',
    album: 'A Night at the Opera',
    playerName: 'Plexamp'
};

const movie = {
    sessionKey: '2',
    mediaType: 'movie',
    title: 'Blade Runner',
    year: 1982,
    playerName: 'Plex for Apple TV'
};

const episode = {
    sessionKey: '3',
    mediaType: 'tv',
    title: 'Good News About Hell',
    showTitle: 'Severance',
    season: 1,
    episode: 2,
    year: 2022,
    playerName: 'Plex Web'
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

describe('mediaTypeOf', () => {
    it('defaults to music for the pre-video session shape', () => {
        expect(mediaTypeOf(track)).toBe('music');
    });

    it('reads the media type when there is one', () => {
        expect(mediaTypeOf(movie)).toBe('movie');
        expect(mediaTypeOf(episode)).toBe('tv');
    });

    it('falls back to music for an unrecognized type, as the backend registry does', () => {
        expect(mediaTypeOf({ mediaType: 'photo' })).toBe('music');
        expect(mediaTypeOf(null)).toBe('music');
    });
});

describe('video tokens', () => {
    it('replaces {show}, {season} and {episode}', () => {
        expect(renderPresenceFormat('{show} S{season}E{episode}: {track}', episode)).toBe('Severance S1E2: Good News About Hell');
    });

    it('renders an unknown season or episode as 0, matching the backend %d', () => {
        expect(renderPresenceFormat('S{season}E{episode}', movie)).toBe('S0E0');
    });
});

describe('renderPresenceLines for video', () => {
    it('renders a film as its title and year', () => {
        expect(renderPresenceLines(null, movie)).toEqual({
            details: 'Blade Runner',
            state: 'Movie • 1982'
        });
    });

    it('renders a film with no year as just Movie', () => {
        expect(renderPresenceLines(null, { ...movie, year: 0 })).toEqual({
            details: 'Blade Runner',
            state: 'Movie'
        });
    });

    it('renders an episode as its title and a padded show • SxxExx line', () => {
        expect(renderPresenceLines(null, episode)).toEqual({
            details: 'Good News About Hell',
            state: 'Severance • S01E02'
        });
    });

    it('falls back to the show alone when the numbering is unknown', () => {
        expect(renderPresenceLines(null, { ...episode, season: 0, episode: 0 })).toEqual({
            details: 'Good News About Hell',
            state: 'Severance'
        });
    });

    it('falls back to a generic label with no show at all', () => {
        expect(renderPresenceLines(null, { mediaType: 'tv', title: 'Pilot' })).toEqual({
            details: 'Pilot',
            state: 'TV Episode'
        });
    });

    it('still honours a custom format for video', () => {
        expect(renderPresenceLines({ details: '{show}', state: '{track}' }, episode)).toEqual({
            details: 'Severance',
            state: 'Good News About Hell'
        });
    });
});

describe('renderPresenceLines for music without an artist', () => {
    it('narrates the playback state, as the backend builder does', () => {
        expect(renderPresenceLines(null, { title: 'Untitled', state: 'playing' }).state).toBe('Playing on Plex');
        expect(renderPresenceLines(null, { title: 'Untitled', state: 'paused' }).state).toBe('Paused');
    });
});
