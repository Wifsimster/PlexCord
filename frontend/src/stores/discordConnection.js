import { defineStore } from 'pinia';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { IsDiscordConnected, GetConnectionHistory, GetDiscordRetryState, RetryDiscordConnection, ConnectDiscord, GetErrorInfo } from '../../wailsjs/go/main/App';
import { formatRelativeTime } from '../utils/timeUtils';
import { t } from '@/i18n';

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
        lastConnectedRelative: (state) => {
            return formatRelativeTime(state.lastConnected);
        },

        isRetrying: (state) => state.retryState?.isRetrying || false,

        hasError: (state) => !!state.error,

        statusLabel() {
            if (this.connected) return t('connectionStore.statusConnected');
            if (this.loading) return t('connectionStore.statusConnecting');
            if (this.isRetrying) return t('connectionStore.statusRetrying');
            if (this.error) return t('connectionStore.statusDisconnected');
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

        setupEventListeners() {
            EventsOn('DiscordConnected', () => {
                this.connected = true;
                this.error = null;
                this.lastConnected = new Date().toISOString();
            });

            EventsOn('DiscordDisconnected', async (data) => {
                this.connected = false;

                const errorCode = data?.error?.code || data?.code;
                if (!errorCode) {
                    this.error = null;
                    return;
                }
                await this.setError(errorCode);
                if (this.connected) {
                    this.error = null;
                }
            });

            EventsOn('DiscordRetryState', (state) => {
                this.retryState = state;
            });
        },

        cleanup() {
            EventsOff('DiscordConnected');
            EventsOff('DiscordDisconnected');
            EventsOff('DiscordRetryState');
            this.initialized = false;
        },

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

        async retry() {
            this.loading = true;
            try {
                await RetryDiscordConnection();
            } finally {
                this.loading = false;
            }
        },

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

        async setError(errorCode) {
            try {
                this.error = await GetErrorInfo(errorCode);
            } catch (err) {
                this.error = {
                    code: errorCode,
                    title: t('connectionStore.errorTitle'),
                    description: t('connectionStore.errorDiscord'),
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
