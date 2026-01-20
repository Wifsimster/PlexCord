import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetPlexConnectionStatus, IsDiscordConnected, GetConnectionHistory, GetPlexRetryState, GetDiscordRetryState, RetryPlexConnection, RetryDiscordConnection, ConnectDiscord, GetErrorInfo } from '../../wailsjs/go/main/App';

/**
 * Connection Store
 * Manages Plex and Discord connection status with real-time updates.
 */
export const useConnectionStore = defineStore('connection', {
    state: () => ({
        // Plex connection status
        plex: {
            connected: false,
            polling: false,
            inErrorState: false,
            serverUrl: '',
            userId: '',
            userName: '',
            lastConnected: null,
            retryState: null
        },

        // Discord connection status
        discord: {
            connected: false,
            lastConnected: null,
            retryState: null,
            error: null
        },

        // Event listeners initialized
        initialized: false,

        // Loading states
        loading: {
            plex: false,
            discord: false
        },

        // Active errors for ErrorBanner display
        // Array of { source: 'plex'|'discord', errorInfo: ErrorInfo, dismissed: boolean }
        errors: []
    }),

    getters: {
        /**
         * Check if both services are connected
         */
        allConnected: (state) => state.plex.connected && state.discord.connected,

        /**
         * Get overall connection health status
         */
        connectionHealth: (state) => {
            if (state.plex.connected && state.discord.connected) return 'healthy';
            if (state.plex.inErrorState || state.discord.error) return 'error';
            if (!state.plex.connected || !state.discord.connected) return 'partial';
            return 'unknown';
        },

        /**
         * Format last connected time as relative string
         */
        plexLastConnectedRelative: (state) => {
            return formatRelativeTime(state.plex.lastConnected);
        },

        /**
         * Format last connected time as relative string
         */
        discordLastConnectedRelative: (state) => {
            return formatRelativeTime(state.discord.lastConnected);
        },

        /**
         * Check if Plex is retrying
         */
        isPlexRetrying: (state) => state.plex.retryState?.isRetrying || false,

        /**
         * Check if Discord is retrying
         */
        isDiscordRetrying: (state) => state.discord.retryState?.isRetrying || false,

        /**
         * Get active (non-dismissed) errors for display
         */
        activeErrors: (state) => state.errors.filter((e) => !e.dismissed),

        /**
         * Get Plex error if exists
         */
        plexError: (state) => state.errors.find((e) => e.source === 'plex' && !e.dismissed),

        /**
         * Get Discord error if exists
         */
        discordError: (state) => state.errors.find((e) => e.source === 'discord' && !e.dismissed),

        /**
         * Check if there are any active errors
         */
        hasActiveErrors: (state) => state.errors.some((e) => !e.dismissed)
    },

    actions: {
        /**
         * Initialize event listeners and fetch initial status
         */
        async initialize() {
            if (this.initialized) return;

            // Setup event listeners
            this.setupEventListeners();

            // Fetch initial status
            await this.refreshStatus();

            // Auto-reconnect if connections are down after page refresh
            // This ensures connections are restored after frontend reload
            this.autoReconnect();

            this.initialized = true;
        },

        /**
         * Auto-reconnect to services if they're disconnected
         * Called after page refresh to restore connections
         */
        async autoReconnect() {
            // Small delay to let initial status settle
            setTimeout(async () => {
                // Auto-reconnect Discord if not connected
                if (!this.discord.connected && !this.loading.discord && !this.isDiscordRetrying) {
                    console.log('Auto-reconnecting to Discord after page refresh...');
                    try {
                        await this.connectDiscord('');
                    } catch (error) {
                        console.log('Auto-reconnect to Discord failed:', error);
                    }
                }

                // Auto-reconnect Plex if not connected or not polling
                if ((!this.plex.connected || !this.plex.polling) && !this.loading.plex && !this.isPlexRetrying) {
                    console.log('Auto-reconnecting to Plex after page refresh...');
                    try {
                        await this.retryPlex();
                    } catch (error) {
                        console.log('Auto-reconnect to Plex failed:', error);
                    }
                }
            }, 500);
        },

        /**
         * Setup Wails event listeners
         */
        setupEventListeners() {
            // Plex connection events
            EventsOn('PlexConnectionError', (_data) => {
                this.plex.connected = false;
                this.plex.inErrorState = true;
                console.log('Plex connection error:', _data);
                // Add error banner with fallback error code
                const errorCode = _data?.errorCode || _data?.code || 'PLEX_UNREACHABLE';
                this.addError('plex', errorCode);
            });

            EventsOn('PlexConnectionRestored', () => {
                this.plex.connected = true;
                this.plex.inErrorState = false;
                this.plex.lastConnected = new Date().toISOString();
                // Remove error banner
                this.removeError('plex');
            });

            // Discord connection events
            EventsOn('DiscordConnected', () => {
                this.discord.connected = true;
                this.discord.error = null;
                this.discord.lastConnected = new Date().toISOString();
                // Remove error banner
                this.removeError('discord');
            });

            EventsOn('DiscordDisconnected', (data) => {
                this.discord.connected = false;
                if (data?.error) {
                    this.discord.error = data.error;
                }
                // Add error banner
                if (data?.code) {
                    this.addError('discord', data.code);
                } else if (data?.error) {
                    // Fallback: use DISCORD_NOT_RUNNING as default code
                    this.addError('discord', 'DISCORD_NOT_RUNNING');
                }
            });

            // Retry state events
            EventsOn('PlexRetryState', (state) => {
                this.plex.retryState = state;
            });

            EventsOn('DiscordRetryState', (state) => {
                this.discord.retryState = state;
            });
        },

        /**
         * Cleanup event listeners
         */
        cleanup() {
            EventsOff('PlexConnectionError');
            EventsOff('PlexConnectionRestored');
            EventsOff('DiscordConnected');
            EventsOff('DiscordDisconnected');
            EventsOff('PlexRetryState');
            EventsOff('DiscordRetryState');
            this.initialized = false;
        },

        /**
         * Refresh all connection status from backend
         */
        async refreshStatus() {
            try {
                // Get Plex status
                const plexStatus = await GetPlexConnectionStatus();
                this.plex.connected = plexStatus.connected;
                this.plex.polling = plexStatus.polling;
                this.plex.inErrorState = plexStatus.inErrorState;
                this.plex.serverUrl = plexStatus.serverUrl;
                this.plex.userId = plexStatus.userId;
                this.plex.userName = plexStatus.userName;

                // Get Discord status
                this.discord.connected = await IsDiscordConnected();

                // Get connection history
                const history = await GetConnectionHistory();
                this.plex.lastConnected = history.plexLastConnected;
                this.discord.lastConnected = history.discordLastConnected;

                // Get retry states
                this.plex.retryState = await GetPlexRetryState();
                this.discord.retryState = await GetDiscordRetryState();
            } catch (error) {
                console.error('Failed to refresh connection status:', error);
            }
        },

        /**
         * Manually retry Plex connection
         */
        async retryPlex() {
            this.loading.plex = true;
            try {
                await RetryPlexConnection();
            } finally {
                this.loading.plex = false;
            }
        },

        /**
         * Manually retry Discord connection
         */
        async retryDiscord() {
            this.loading.discord = true;
            try {
                await RetryDiscordConnection();
            } finally {
                this.loading.discord = false;
            }
        },

        /**
         * Connect to Discord
         */
        async connectDiscord(clientId = '') {
            this.loading.discord = true;
            try {
                await ConnectDiscord(clientId);
            } catch (error) {
                this.discord.error = error;
                throw error;
            } finally {
                this.loading.discord = false;
            }
        },

        /**
         * Add an error to the errors array
         * @param {string} source - 'plex' or 'discord'
         * @param {string} errorCode - Error code from backend
         */
        async addError(source, errorCode) {
            // Remove any existing error from same source first
            this.removeError(source);

            try {
                // Fetch detailed error info from backend
                const errorInfo = await GetErrorInfo(errorCode);
                this.errors.push({
                    source,
                    errorInfo,
                    dismissed: false,
                    timestamp: new Date().toISOString()
                });
            } catch (err) {
                // Fallback if GetErrorInfo fails
                this.errors.push({
                    source,
                    errorInfo: {
                        code: errorCode,
                        title: 'Connection Error',
                        description: `Failed to connect to ${source === 'plex' ? 'Plex' : 'Discord'}`,
                        suggestion: 'Please check your connection and try again.',
                        retryable: true
                    },
                    dismissed: false,
                    timestamp: new Date().toISOString()
                });
            }
        },

        /**
         * Remove error for a specific source (when connection restored)
         * @param {string} source - 'plex' or 'discord'
         */
        removeError(source) {
            this.errors = this.errors.filter((e) => e.source !== source);
        },

        /**
         * Dismiss an error (hide it without removing)
         * @param {string} source - 'plex' or 'discord'
         */
        dismissError(source) {
            const error = this.errors.find((e) => e.source === source);
            if (error) {
                error.dismissed = true;
            }
        },

        /**
         * Clear all errors
         */
        clearAllErrors() {
            this.errors = [];
        }
    }
});

/**
 * Format a timestamp as relative time (e.g., "5 minutes ago")
 */
function formatRelativeTime(timestamp) {
    if (!timestamp) return 'Never';

    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return 'Just now';
    if (diffMin < 60) return `${diffMin} minute${diffMin !== 1 ? 's' : ''} ago`;
    if (diffHour < 24) return `${diffHour} hour${diffHour !== 1 ? 's' : ''} ago`;
    if (diffDay < 7) return `${diffDay} day${diffDay !== 1 ? 's' : ''} ago`;

    return date.toLocaleDateString();
}
