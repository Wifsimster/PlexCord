import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetCurrentSession } from '../../wailsjs/go/main/App';
import { usePlexConnectionStore } from './plexConnection';

export const usePlaybackStore = defineStore('playback', {
    state: () => ({
        currentTrack: null,

        isPlaying: false,
        isPaused: false,
        isStopped: true,

        initialized: false
    }),

    getters: {
        hasActiveSession: (state) => {
            return state.currentTrack !== null;
        },

        formattedPosition: (state) => {
            if (!state.currentTrack || !state.currentTrack.viewOffset) {
                return '0:00';
            }
            return formatDuration(state.currentTrack.viewOffset);
        },

        formattedDuration: (state) => {
            if (!state.currentTrack || !state.currentTrack.duration) {
                return '0:00';
            }
            return formatDuration(state.currentTrack.duration);
        },

        progressPercent: (state) => {
            if (!state.currentTrack || !state.currentTrack.duration) {
                return 0;
            }
            const percent = (state.currentTrack.viewOffset / state.currentTrack.duration) * 100;
            return Math.min(100, Math.max(0, percent));
        },

        playbackState: (state) => {
            if (state.isPlaying) return 'playing';
            if (state.isPaused) return 'paused';
            return 'stopped';
        }
    },

    actions: {
        async initializeEventListeners() {
            if (this.initialized) {
                return;
            }

            EventsOn('PlaybackUpdated', (session) => {
                usePlexConnectionStore().markLivePoll();
                this.setTrack(session);
            });

            EventsOn('PlaybackStopped', () => {
                usePlexConnectionStore().markLivePoll();
                this.clearTrack();
            });

            this.initialized = true;

            try {
                const currentSession = await GetCurrentSession();
                if (currentSession) {
                    console.log('Restoring current playback session after page refresh');
                    usePlexConnectionStore().markLivePoll();
                    this.setTrack(currentSession);
                } else {
                    this.clearTrack();
                }
            } catch (error) {
                console.error('Failed to restore current session:', error);
            }
        },

        cleanupEventListeners() {
            if (!this.initialized) {
                return;
            }

            EventsOff('PlaybackUpdated');
            EventsOff('PlaybackStopped');
            this.initialized = false;
        },

        setTrack(session) {
            if (!session) {
                this.clearTrack();
                return;
            }

            const title = session.title ?? session.track ?? '';

            this.currentTrack = {
                sessionKey: session.sessionKey,
                mediaType: session.mediaType || 'music',
                title,
                track: title,
                artist: session.artist,
                album: session.album,
                year: session.year,
                showTitle: session.showTitle,
                season: session.season,
                episode: session.episode,
                thumb: session.thumb,
                thumbUrl: session.thumbUrl,
                duration: session.duration,
                viewOffset: session.viewOffset,
                state: session.state,
                playerName: session.playerName
            };

            this.isPlaying = session.state === 'playing';
            this.isPaused = session.state === 'paused';
            this.isStopped = session.state === 'stopped';
        },

        clearTrack() {
            this.currentTrack = null;
            this.isPlaying = false;
            this.isPaused = false;
            this.isStopped = true;
        }
    }
});

function formatDuration(ms) {
    if (!ms || ms <= 0) {
        return '0:00';
    }
    const totalSeconds = Math.floor(ms / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}
