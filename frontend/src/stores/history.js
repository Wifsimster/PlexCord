import { defineStore } from 'pinia';
import { GetListeningHistory, GetListeningStats, ClearListeningHistory } from '../../wailsjs/go/main/App';

/**
 * History Store
 * Manages listening history state and interactions with the backend.
 */
export const useHistoryStore = defineStore('history', {
    state: () => ({
        /** @type {Array<{track: string, artist: string, album: string, duration: number, startedAt: string, thumbUrl?: string}>} */
        entries: [],

        /** @type {{totalTracks: number, uniqueArtists: number, mostPlayedArtist: string}} */
        stats: {
            totalTracks: 0,
            uniqueArtists: 0,
            mostPlayedArtist: ''
        },

        /** Whether a fetch is in progress */
        loading: false
    }),

    getters: {
        /**
         * Whether there are any history entries
         * @returns {boolean}
         */
        hasHistory: (state) => state.entries.length > 0
    },

    actions: {
        /**
         * Fetch recent listening history from the backend.
         * @param {number} limit - Maximum number of entries to retrieve
         */
        async fetchHistory(limit = 20) {
            this.loading = true;
            try {
                const result = await GetListeningHistory(limit);
                this.entries = result || [];
            } catch (error) {
                console.error('Failed to fetch listening history:', error);
                this.entries = [];
            } finally {
                this.loading = false;
            }
        },

        /**
         * Fetch aggregate listening statistics from the backend.
         */
        async fetchStats() {
            try {
                const result = await GetListeningStats();
                this.stats = result || { totalTracks: 0, uniqueArtists: 0, mostPlayedArtist: '' };
            } catch (error) {
                console.error('Failed to fetch listening stats:', error);
            }
        },

        /**
         * Clear all listening history.
         */
        async clearHistory() {
            try {
                await ClearListeningHistory();
                this.entries = [];
                this.stats = { totalTracks: 0, uniqueArtists: 0, mostPlayedArtist: '' };
            } catch (error) {
                console.error('Failed to clear listening history:', error);
            }
        }
    }
});
