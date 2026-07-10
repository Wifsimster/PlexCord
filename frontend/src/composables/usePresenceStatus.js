import { computed } from 'vue';
import { usePlaybackStore } from '@/stores/playback';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePresenceStore } from '@/stores/presence';

/**
 * The headline state machine (spec §5.0.6 / §5.1).
 *
 * Pure computed layer over the playback store, both connection stores and
 * the presence-pause store. Owns NO lifecycle: consumers must ensure the
 * stores are initialized (AppLayout does this app-wide).
 *
 * Status values:
 *   'plex-error'     — Plex failure blocks the relay   → '▲ Plex unreachable'
 *   'discord-error'  — Discord failure blocks the relay → '▲ Discord disconnected'
 *   'paused'         — presence paused by the user      → '⏸ Paused'
 *   'track-paused'   — playback paused in the player    → '⏸ Paused'
 *   'live'           — playing & relaying               → '▶ Live — {title}'
 *   'idle'           — nothing playing                  → '– Idle'
 */
export function usePresenceStatus() {
    const playback = usePlaybackStore();
    const plex = usePlexConnectionStore();
    const discord = useDiscordConnectionStore();
    const presence = usePresenceStore();

    const status = computed(() => {
        if (plex.hasError) return 'plex-error';
        if (discord.hasError) return 'discord-error';
        if (presence.paused) return 'paused';
        if (playback.isPlaying) return 'live';
        if (playback.isPaused) return 'track-paused';
        return 'idle';
    });

    const trackTitle = computed(() => playback.currentTrack?.track ?? playback.currentTrack?.title ?? '');

    /** Headline text WITHOUT the leading glyph (renderers add ▶/⏸/–/▲). */
    const headline = computed(() => {
        switch (status.value) {
            case 'plex-error':
                return 'Plex unreachable';
            case 'discord-error':
                return 'Discord disconnected';
            case 'paused':
            case 'track-paused':
                return 'Paused';
            case 'live':
                return trackTitle.value ? `Live — ${trackTitle.value}` : 'Live';
            default:
                return 'Idle';
        }
    });

    /** Semantic color bucket for the headline: success | warn | danger | muted. */
    const severity = computed(() => {
        switch (status.value) {
            case 'live':
                return 'success';
            case 'paused':
            case 'track-paused':
                return 'warn';
            case 'plex-error':
            case 'discord-error':
                return 'danger';
            default:
                return 'muted';
        }
    });

    const isErrored = computed(() => status.value === 'plex-error' || status.value === 'discord-error');

    /** Both endpoints healthy — the relay can publish. */
    const relayHealthy = computed(() => plex.connected && discord.connected && !plex.hasError && !discord.hasError);

    const presencePaused = computed(() => presence.paused);

    return {
        status,
        headline,
        severity,
        trackTitle,
        isErrored,
        relayHealthy,
        presencePaused,
        /** Toggle the app-wide presence pause. @returns {Promise<boolean>} new state */
        togglePause: () => presence.toggle()
    };
}
