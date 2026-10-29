import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetPlexConnectionStatus, GetConnectionHistory, GetPlexRetryState, RetryPlexConnection, GetErrorInfo } from '../../wailsjs/go/main/App';
import { formatRelativeTime } from '../utils/timeUtils';
import { t } from '@/i18n';

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
        lastConnectedRelative: (state) => {
            return formatRelativeTime(state.lastConnected);
        },

        isRetrying: (state) => state.retryState?.isRetrying || false,

        hasError: (state) => state.inErrorState || !!state.error,

        statusLabel() {
            if (this.connected && this.polling) return t('connectionStore.statusConnected');
            if (this.loading) return t('connectionStore.statusConnecting');
            if (this.isRetrying) return t('connectionStore.statusRetrying');
            if (this.inErrorState) return t('connectionStore.statusDisconnected');
            return t('connectionStore.statusNotConnected');
        }
    },

    actions: {
        async initialize() {
            if (this.initialized) return;

            this.setupEventListeners();
            await this.refreshStatus();
            this.autoReconnect();

            this.initialized = true;
        },

        async autoReconnect() {
            setTimeout(async () => {
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
                this.polling = true;
                this.inErrorState = false;
                this.lastConnected = new Date().toISOString();
                this.clearError();
            });

            EventsOn('PlexRetryState', (state) => {
                this.retryState = state;
            });
        },

        markLivePoll() {
            if (this.connected && this.polling && !this.inErrorState) {
                return;
            }
            this.connected = true;
            this.polling = true;
            this.inErrorState = false;
            this.clearError();
        },

        cleanup() {
            EventsOff('PlexConnectionError');
            EventsOff('PlexConnectionRestored');
            EventsOff('PlexRetryState');
            this.initialized = false;
        },

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

        async retry() {
            this.loading = true;
            try {
                await RetryPlexConnection();
            } catch (error) {
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

        clearError() {
            this.error = null;
        }
    }
});
