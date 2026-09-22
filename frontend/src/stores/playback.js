import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetCurrentSession } from '../../wailsjs/go/main/App';
import { usePlexConnectionStore } from './plexConnection';

/**
 * Playback Store
 * Manages the current playback state and metadata from Plex — a track, a film
 * or a TV episode. Subscribes to Wails events for real-time updates.
 */
export const usePlaybackStore = defineStore('playback', {
    state: () => ({
        // Current track information
        currentTrack: null,

        // Playback state
        isPlaying: false,
        isPaused: false,
        isStopped: true,

        // Event listeners initialized
        initialized: false
    }),

    getters: {
        /**
         * Check if something is currently active (playing or paused)
         * @returns {boolean}
         */
        hasActiveSession: (state) => {
            return state.currentTrack !== null;
        },

        /**
         * Get formatted current position as mm:ss
         * @returns {string}
         */
        formattedPosition: (state) => {
            if (!state.currentTrack || !state.currentTrack.viewOffset) {
                return '0:00';
            }
            return formatDuration(state.currentTrack.viewOffset);
        },

        /**
         * Get formatted total duration as mm:ss
         * @returns {string}
         */
        formattedDuration: (state) => {
            if (!state.currentTrack || !state.currentTrack.duration) {
                return '0:00';
            }
            return formatDuration(state.currentTrack.duration);
        },

        /**
         * Get playback progress as percentage (0-100)
         * @returns {number}
         */
        progressPercent: (state) => {
            if (!state.currentTrack || !state.currentTrack.duration) {
                return 0;
            }
            const percent = (state.currentTrack.viewOffset / state.currentTrack.duration) * 100;
            return Math.min(100, Math.max(0, percent));
        },

        /**
         * Get display-friendly playback state
         * @returns {string}
         */
        playbackState: (state) => {
            if (state.isPlaying) return 'playing';
            if (state.isPaused) return 'paused';
            return 'stopped';
        }
    },

    actions: {
        /**
         * Initialize event listeners for Wails events
         * Should be called once when the app starts
         */
        async initializeEventListeners() {
            if (this.initialized) {
                return;
            }

            // Listen for PlaybackUpdated events. A playback event is proof the
            // Plex poller reached the server, so tell the connection store the
            // link is live — this keeps its tile from getting stuck on "Idle"
            // when the startup status snapshot raced ahead of the poller.
            EventsOn('PlaybackUpdated', (session) => {
                usePlexConnectionStore().markLivePoll();
                this.setTrack(session);
            });

            // Listen for PlaybackStopped events
            EventsOn('PlaybackStopped', () => {
                usePlexConnectionStore().markLivePoll();
                this.clearTrack();
            });

            this.initialized = true;

            // Restore current session after page refresh
            // This ensures the preview shows current playback immediately
            try {
                const currentSession = await GetCurrentSession();
                if (currentSession) {
                    console.log('Restoring current playback session after page refresh');
                    usePlexConnectionStore().markLivePoll();
                    this.setTrack(currentSession);
                } else {
                    // Nothing is playing now: drop any track kept from before
                    // the listeners were removed, whose stop event was missed.
                    this.clearTrack();
                }
            } catch (error) {
                console.error('Failed to restore current session:', error);
            }
        },

        /**
         * Clean up event listeners
         * Should be called when the app is unmounted
         */
        cleanupEventListeners() {
            if (!this.initialized) {
                return;
            }

            EventsOff('PlaybackUpdated');
            EventsOff('PlaybackStopped');
            this.initialized = false;
        },

        /**
         * Set the current item from a MediaSession event.
         *
         * `track` is kept alongside `title` because every consumer reads it:
         * for a film or an episode it holds the title, which is what those
         * consumers are showing anyway.
         *
         * @param {Object} session - MediaSession object from backend
         */
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

            // Update playback state flags
            this.isPlaying = session.state === 'playing';
            this.isPaused = session.state === 'paused';
            this.isStopped = session.state === 'stopped';
        },

        /**
         * Clear the current item (playback stopped)
         */
        clearTrack() {
            this.currentTrack = null;
            this.isPlaying = false;
            this.isPaused = false;
            this.isStopped = true;
        }
    }
});

/**
 * Format milliseconds to mm:ss display format
 * @param {number} ms - Duration in milliseconds
 * @returns {string} Formatted duration string
 */
function formatDuration(ms) {
    if (!ms || ms <= 0) {
        return '0:00';
    }
    const totalSeconds = Math.floor(ms / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}
