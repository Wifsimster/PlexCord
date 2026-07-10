/**
 * Presence format rendering (spec §5.0.4).
 *
 * Mirrors the backend token replacement in internal/discord/builder.go
 * (applyFormatTokens) so the <DiscordSpecimen> shows exactly what the
 * relay transmits. Supported tokens: {track} {artist} {album} {year}
 * {player}. {year} is only substituted when the track data carries a
 * year; otherwise the token is stripped (replaced with an empty string,
 * matching the backend's behavior for empty fields).
 */

/**
 * Render a presence format string against a track object.
 *
 * @param {string} format - Format string, e.g. '{track} by {artist}'
 * @param {Object|null} track - Track/session object. Accepts both the
 *   MusicSession shape ({track, artist, album, playerName}) and the
 *   MediaSession shape ({title, ...}); year may be a number or string.
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
        '{player}': track.playerName ?? track.player ?? ''
    };

    return format.replace(/\{track\}|\{artist\}|\{album\}|\{year\}|\{player\}/g, (token) => tokens[token]);
}

/**
 * Render both presence lines with the backend's default fallback.
 *
 * When no custom formats are set, the Go builder (musicBuilder.Build)
 * falls back to details = track title and state = 'by {artist} • {album}'
 * (or 'by {artist}' without an album). This helper reproduces that so
 * consumers get faithful lines whether or not formats are configured.
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

    // Backend default rendering (no custom formats configured).
    const title = track.track ?? track.title ?? '';
    let state = '';
    if (track.artist) {
        state = track.album ? `by ${track.artist} • ${track.album}` : `by ${track.artist}`;
    }
    return { details: title, state };
}
