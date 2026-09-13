/**
 * Presence format rendering (spec §5.0.4).
 *
 * Mirrors the backend token replacement and the per-media-type defaults in
 * internal/discord/builder.go so the <DiscordSpecimen> shows exactly what the
 * relay transmits. Supported tokens: {track} {artist} {album} {year}
 * {player} {show} {season} {episode}. A token whose field is empty is
 * stripped, matching the backend's behavior; {season} and {episode} are
 * numeric there and render as 0 when unknown.
 *
 * The literal wording below ("by", "Movie", "TV Episode") is deliberately not
 * translated: it is what Discord displays, and Discord shows it in English to
 * everyone who sees the profile.
 */

const MEDIA_MUSIC = 'music';
const MEDIA_MOVIE = 'movie';
const MEDIA_TV = 'tv';

/**
 * The media type of a session, defaulting to music for the pre-video session
 * shape (and for anything unrecognized, which is what the backend registry
 * falls back to).
 *
 * @param {Object|null} track - Track/session object
 * @returns {string} 'music' | 'movie' | 'tv'
 */
export function mediaTypeOf(track) {
    const type = track?.mediaType;
    return type === MEDIA_MOVIE || type === MEDIA_TV ? type : MEDIA_MUSIC;
}

/**
 * Render a presence format string against a session object.
 *
 * @param {string} format - Format string, e.g. '{track} by {artist}'
 * @param {Object|null} track - Track/session object. Accepts both the
 *   MediaSession shape ({title, showTitle, …}) and the older MusicSession
 *   shape ({track, artist, album, playerName}); year may be a number or string.
 * @returns {string} The rendered line ('' when format or track missing)
 */
export function renderPresenceFormat(format, track) {
    if (!format || !track) {
        return '';
    }

    const year = track.year ? String(track.year) : '';
    const tokens = {
        '{track}': track.track ?? track.title ?? '',
        '{artist}': track.artist ?? '',
        '{album}': track.album ?? '',
        '{year}': year,
        '{player}': track.playerName ?? track.player ?? '',
        '{show}': track.showTitle ?? '',
        '{season}': String(track.season ?? 0),
        '{episode}': String(track.episode ?? 0)
    };

    return format.replace(/\{track\}|\{artist\}|\{album\}|\{year\}|\{player\}|\{show\}|\{season\}|\{episode\}/g, (token) => tokens[token]);
}

/**
 * Pad a season or episode number the way the backend's S%02dE%02d does.
 *
 * @param {number} value
 * @returns {string}
 */
function pad2(value) {
    return String(value).padStart(2, '0');
}

/**
 * The state line each media type shows when no custom format is configured.
 *
 * @param {Object} track - Track/session object
 * @returns {string}
 */
function defaultState(track) {
    switch (mediaTypeOf(track)) {
        case MEDIA_MOVIE:
            return track.year ? `Movie • ${track.year}` : 'Movie';
        case MEDIA_TV:
            if (track.showTitle && track.season > 0 && track.episode > 0) {
                return `${track.showTitle} • S${pad2(track.season)}E${pad2(track.episode)}`;
            }
            return track.showTitle || 'TV Episode';
        default:
            if (track.artist) {
                return track.album ? `by ${track.artist} • ${track.album}` : `by ${track.artist}`;
            }
            // No artist at all: the backend narrates the playback state instead.
            if (track.state === 'paused') return 'Paused';
            if (track.state) return 'Playing on Plex';
            return '';
    }
}

/**
 * Render both presence lines with the backend's default fallback.
 *
 * When no custom formats are set, the Go builder for the session's media type
 * supplies the lines; this helper reproduces them so consumers get faithful
 * output whether or not formats are configured.
 *
 * @param {Object|null} formats - { details, state } format strings (either may be '')
 * @param {Object|null} track - Track/session object (see renderPresenceFormat)
 * @returns {{ details: string, state: string }}
 */
export function renderPresenceLines(formats, track) {
    if (!track) {
        return { details: '', state: '' };
    }

    const detailsFormat = formats?.details ?? '';
    const stateFormat = formats?.state ?? '';

    if (detailsFormat || stateFormat) {
        return {
            details: renderPresenceFormat(detailsFormat, track),
            state: renderPresenceFormat(stateFormat, track)
        };
    }

    return {
        details: track.track ?? track.title ?? '',
        state: defaultState(track)
    };
}
