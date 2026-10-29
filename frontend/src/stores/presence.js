import { defineStore } from 'pinia';
import { IsPresencePaused, TogglePresencePause } from '../../wailsjs/go/main/App';

export const usePresenceStore = defineStore('presence', {
    state: () => ({
        paused: false,
        loading: false,
        initialized: false
    }),

    actions: {
        async initialize() {
            if (this.initialized) return;
            this.initialized = true;
            await this.refresh();
        },

        async refresh() {
            try {
                this.paused = await IsPresencePaused();
            } catch (error) {
                console.error('Failed to read presence pause state:', error);
            }
        },

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
