const MEDIA_MUSIC = 'music';
const MEDIA_MOVIE = 'movie';
const MEDIA_TV = 'tv';

export function mediaTypeOf(track) {
    const type = track?.mediaType;
    return type === MEDIA_MOVIE || type === MEDIA_TV ? type : MEDIA_MUSIC;
}

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

function pad2(value) {
    return String(value).padStart(2, '0');
}

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
            if (track.state === 'paused') return 'Paused';
            if (track.state) return 'Playing on Plex';
            return '';
    }
}

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
