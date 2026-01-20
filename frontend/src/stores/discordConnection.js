import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { IsDiscordConnected, GetConnectionHistory, GetDiscordRetryState, RetryDiscordConnection, ConnectDiscord, GetErrorInfo } from '../../wailsjs/go/main/App';
import { formatRelativeTime } from '../utils/timeUtils';

/**
 * Discord Connection Store
 * Manages Discord Rich Presence connection status with real-time updates.
 * Separated from Plex to follow single responsibility principle.
 */
export const useDiscordConnectionStore = defineStore('discordConnection', {
    state: () => ({
        connected: false,
        lastConnected: null,
        retryState: null,
        error: null,
        loading: false,
        initialized: false
    }),

    getters: {
        /**
         * Format last connected time as relative string
         */
        lastConnectedRelative: (state) => {
            return formatRelativeTime(state.lastConnected);
        },

        /**
         * Check if Discord is retrying connection
         */
        isRetrying: (state) => state.retryState?.isRetrying || false,

        /**
         * Check if there's an active error
         */
        hasError: (state) => !!state.error,

        /**
         * Get connection status label
         */
        statusLabel: (state) => {
            if (state.connected) return 'Connected';
            if (state.loading) return 'Connecting...';
            if (state.isRetrying) return 'Retrying...';
            if (state.error) return 'Disconnected';
            return 'Not Connected';
        }
    },

    actions: {
        /**
         * Initialize event listeners and fetch initial status
         */
        async initialize() {
            if (this.initialized) return;

            this.setupEventListeners();
            await this.refreshStatus();
            this.autoReconnect();

            this.initialized = true;
        },

        /**
         * Auto-reconnect if disconnected after page refresh
         */
        async autoReconnect() {
            setTimeout(async () => {
                if (!this.connected && !this.loading && !this.isRetrying) {
                    console.log('Auto-reconnecting to Discord after page refresh...');
                    try {
                        await this.connect('');
                    } catch (error) {
                        console.log('Auto-reconnect to Discord failed:', error);
                    }
                }
            }, 500);
        },

        /**
         * Setup Wails event listeners for Discord
         */
        setupEventListeners() {
            EventsOn('DiscordConnected', () => {
                this.connected = true;
                this.error = null;
                this.lastConnected = new Date().toISOString();
            });

            EventsOn('DiscordDisconnected', async (data) => {
                this.connected = false;

                const errorCode = data?.code || 'DISCORD_NOT_RUNNING';
                await this.setError(errorCode);
            });

            EventsOn('DiscordRetryState', (state) => {
                this.retryState = state;
            });
        },

        /**
         * Cleanup event listeners
         */
        cleanup() {
            EventsOff('DiscordConnected');
            EventsOff('DiscordDisconnected');
            EventsOff('DiscordRetryState');
            this.initialized = false;
        },

        /**
         * Refresh connection status from backend
         */
        async refreshStatus() {
            try {
                this.connected = await IsDiscordConnected();

                const history = await GetConnectionHistory();
                this.lastConnected = history.discordLastConnected;

                this.retryState = await GetDiscordRetryState();
            } catch (error) {
                console.error('Failed to refresh Discord status:', error);
            }
        },

        /**
         * Manually retry connection
         */
        async retry() {
            this.loading = true;
            try {
                await RetryDiscordConnection();
            } finally {
                this.loading = false;
            }
        },

        /**
         * Connect to Discord with optional client ID
         */
        async connect(clientId = '') {
            this.loading = true;
            try {
                await ConnectDiscord(clientId);
            } catch (error) {
                this.error = error;
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Set error with detailed info from backend
         */
        async setError(errorCode) {
            try {
                this.error = await GetErrorInfo(errorCode);
            } catch (err) {
                this.error = {
                    code: errorCode,
                    title: 'Connection Error',
                    description: 'Failed to connect to Discord',
                    suggestion: 'Please check your connection and try again.',
                    retryable: true
                };
            }
        },

        /**
         * Clear error state
         */
        clearError() {
            this.error = null;
        }
    }
});
