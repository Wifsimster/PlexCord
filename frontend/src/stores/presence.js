import { defineStore } from 'pinia';
import { IsPresencePaused, TogglePresencePause } from '../../wailsjs/go/main/App';

/**
 * Presence Store
 * App-wide presence pause state (spec §5.0.6 / F4). Wraps the
 * IsPresencePaused / TogglePresencePause bindings behind a single pinia
 * store so the topbar pause button, the Ctrl+P shortcut and the
 * Dashboard chip all read and mutate the same state.
 */
export const usePresenceStore = defineStore('presence', {
    state: () => ({
        paused: false,
        loading: false,
        initialized: false
    }),

    actions: {
        /**
         * Fetch the initial paused state from the backend (idempotent).
         */
        async initialize() {
            if (this.initialized) return;
            this.initialized = true;
            await this.refresh();
        },

        /**
         * Re-read the paused state from the backend.
         */
        async refresh() {
            try {
                this.paused = await IsPresencePaused();
            } catch (error) {
                console.error('Failed to read presence pause state:', error);
            }
        },

        /**
         * Toggle the presence pause state.
         * @returns {Promise<boolean>} the new paused state
         */
        async toggle() {
            if (this.loading) return this.paused;
            this.loading = true;
            try {
                this.paused = await TogglePresencePause();
            } catch (error) {
                console.error('Failed to toggle presence pause:', error);
            } finally {
                this.loading = false;
            }
            return this.paused;
        }
    }
});
