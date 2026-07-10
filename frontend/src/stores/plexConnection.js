import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetPlexConnectionStatus, GetConnectionHistory, GetPlexRetryState, RetryPlexConnection, GetErrorInfo } from '../../wailsjs/go/main/App';
import { formatRelativeTime } from '../utils/timeUtils';
import { t } from '@/i18n';

/**
 * Plex Connection Store
 * Manages Plex server connection status with real-time updates.
 * Separated from Discord to follow single responsibility principle.
 */
export const usePlexConnectionStore = defineStore('plexConnection', {
    state: () => ({
        connected: false,
        polling: false,
        inErrorState: false,
        serverUrl: '',
        userId: '',
        userName: '',
        lastConnected: null,
        retryState: null,
        loading: false,
        error: null,
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
         * Check if Plex is retrying connection
         */
        isRetrying: (state) => state.retryState?.isRetrying || false,

        /**
         * Check if there's an active error
         */
        hasError: (state) => state.inErrorState || !!state.error,

        /**
         * Get connection status label
         * Note: Uses function form to access other getters via `this`
         */
        statusLabel() {
            if (this.connected && this.polling) return t('connectionStore.statusConnected');
            if (this.loading) return t('connectionStore.statusConnecting');
            if (this.isRetrying) return t('connectionStore.statusRetrying');
            if (this.inErrorState) return t('connectionStore.statusDisconnected');
            return t('connectionStore.statusNotConnected');
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
                // Nothing to connect to yet — an active server and a selected
                // user are both required. Without this guard the dashboard
                // fires a retry that can never succeed and gets stuck showing
                // "Retrying..." indefinitely.
                if (!this.serverUrl || !this.userId) {
                    return;
                }
                if ((!this.connected || !this.polling) && !this.loading && !this.isRetrying) {
                    console.log('Auto-reconnecting to Plex after page refresh...');
                    try {
                        await this.retry();
                    } catch (error) {
                        console.log('Auto-reconnect to Plex failed:', error);
                    }
                }
            }, 500);
        },

        /**
         * Setup Wails event listeners for Plex
         */
        setupEventListeners() {
            EventsOn('PlexConnectionError', async (data) => {
                this.connected = false;
                this.inErrorState = true;
                console.log('Plex connection error:', data);

                const errorCode = data?.errorCode || data?.code || 'PLEX_UNREACHABLE';
                await this.setError(errorCode);
            });

            EventsOn('PlexConnectionRestored', () => {
                this.connected = true;
                this.inErrorState = false;
                this.lastConnected = new Date().toISOString();
                this.clearError();
            });

            EventsOn('PlexRetryState', (state) => {
                this.retryState = state;
            });
        },

        /**
         * Cleanup event listeners
         */
        cleanup() {
            EventsOff('PlexConnectionError');
            EventsOff('PlexConnectionRestored');
            EventsOff('PlexRetryState');
            this.initialized = false;
        },

        /**
         * Refresh connection status from backend
         */
        async refreshStatus() {
            try {
                const status = await GetPlexConnectionStatus();
                this.connected = status.connected;
                this.polling = status.polling;
                this.inErrorState = status.inErrorState;
                this.serverUrl = status.serverUrl;
                this.userId = status.userId;
                this.userName = status.userName;

                const history = await GetConnectionHistory();
                this.lastConnected = history.plexLastConnected;

                this.retryState = await GetPlexRetryState();
            } catch (error) {
                console.error('Failed to refresh Plex status:', error);
            }
        },

        /**
         * Manually retry connection
         */
        async retry() {
            this.loading = true;
            try {
                await RetryPlexConnection();
            } catch (error) {
                // Surface the failure (e.g. no server/user configured yet)
                // instead of leaving the Reconnect button looking like it
                // did nothing.
                this.inErrorState = true;
                this.error = {
                    code: 'PLEX_RETRY_FAILED',
                    title: t('connectionStore.reconnectFailedTitle'),
                    description: error?.message || t('connectionStore.reconnectFailedPlex'),
                    suggestion: t('connectionStore.reconnectFailedSuggestion'),
                    retryable: true
                };
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
                    title: t('connectionStore.errorTitle'),
                    description: t('connectionStore.errorPlex'),
                    suggestion: t('connectionStore.errorSuggestion'),
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
