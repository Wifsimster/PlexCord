import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { usePlaybackStore } from '@/stores/playback';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePresenceStore } from '@/stores/presence';

export function usePresenceStatus() {
    const { t } = useI18n();
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

    const headline = computed(() => {
        switch (status.value) {
            case 'plex-error':
                return t('presenceStatus.plexUnreachable');
            case 'discord-error':
                return t('presenceStatus.discordDisconnected');
            case 'paused':
            case 'track-paused':
                return t('presenceStatus.paused');
            case 'live':
                return trackTitle.value ? t('presenceStatus.liveWithTitle', { title: trackTitle.value }) : t('presenceStatus.live');
            default:
                return t('presenceStatus.idle');
        }
    });

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

    const tally = computed(() => {
        switch (status.value) {
            case 'live':
                return { kind: 'on', label: t('tally.onAir') };
            case 'paused':
                return { kind: 'off', label: t('tally.offAir') };
            case 'track-paused':
                return { kind: 'hold', label: t('tally.paused') };
            case 'plex-error':
                return { kind: 'fault', label: t('tally.plexDown') };
            case 'discord-error':
                return { kind: 'fault', label: t('tally.discordDown') };
            default:
                return { kind: 'idle', label: t('tally.standby') };
        }
    });

    const isErrored = computed(() => status.value === 'plex-error' || status.value === 'discord-error');

    const relayHealthy = computed(() => plex.connected && discord.connected && !plex.hasError && !discord.hasError);

    const presencePaused = computed(() => presence.paused);

    return {
        status,
        headline,
        severity,
        tally,
        trackTitle,
        isErrored,
        relayHealthy,
        presencePaused,
        togglePause: () => presence.toggle()
    };
}
