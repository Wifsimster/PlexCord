import { defineStore } from 'pinia';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { CanSelfUpdate, CheckForUpdate, DownloadAndInstallUpdate, GetUpdateStatus, RestartApplication } from '../../wailsjs/go/main/App';

export const useUpdatesStore = defineStore('updates', {
    state: () => ({
        status: 'idle',
        info: null,
        progress: 0,
        canSelfUpdate: false,
        lastError: null,
        dismissedVersion: null,
        initialized: false
    }),

    getters: {
        updateAvailable: (state) => !!state.info?.available || state.status === 'available',

        updateReady: (state) => state.status === 'ready',

        installing: (state) => state.status === 'downloading',

        showUpdatePanel() {
            return this.updateAvailable || this.updateReady || this.installing;
        },

        shouldToast(state) {
            const version = state.info?.latestVersion;
            if (!version || version === state.dismissedVersion) return false;
            return this.updateReady || (this.updateAvailable && !state.canSelfUpdate);
        }
    },

    actions: {
        async initialize() {
            if (this.initialized) return;
            this.initialized = true;

            this.setupEventListeners();

            try {
                this.canSelfUpdate = !!(await CanSelfUpdate());
            } catch {
                this.canSelfUpdate = false;
            }
            await this.refreshStatus();
        },

        async refreshStatus() {
            try {
                const status = await GetUpdateStatus();
                if (!status) return;
                if (this.status === 'idle') {
                    this.status = status.state || 'idle';
                    this.progress = Math.round(status.progress ?? 0);
                }
                if (status.info && !this.info) this.info = status.info;
            } catch (error) {
                console.error('Failed to refresh update status:', error);
            }
        },

        setupEventListeners() {
            this._cancelFns = [
                EventsOn('UpdateAvailable', (info) => {
                    this.info = info;
                    if (this.status !== 'downloading' && this.status !== 'ready') {
                        this.status = 'available';
                    }
                }),
                EventsOn('UpdateDownloadProgress', (p) => {
                    this.status = 'downloading';
                    this.progress = Math.round(p?.percent ?? 0);
                }),
                EventsOn('UpdateReady', (info) => {
                    this.status = 'ready';
                    this.progress = 100;
                    if (info) this.info = info;
                }),
                EventsOn('UpdateError', (message) => {
                    this.status = this.info ? 'available' : 'idle';
                    this.progress = 0;
                    this.lastError = message;
                })
            ];
        },

        cleanup() {
            (this._cancelFns || []).forEach((cancel) => cancel());
            this._cancelFns = [];
            this.initialized = false;
        },

        async checkNow() {
            this.info = await CheckForUpdate();
            if (this.info?.available && this.status === 'idle') {
                this.status = 'available';
            }
            return this.info;
        },

        async install() {
            this.lastError = null;
            this.status = 'downloading';
            this.progress = 0;
            try {
                await DownloadAndInstallUpdate();
            } catch (error) {
                if (this.status === 'downloading') {
                    this.status = this.info ? 'available' : 'idle';
                }
                throw error;
            }
        },

        async restart() {
            await RestartApplication();
        },

        dismissToast() {
            this.dismissedVersion = this.info?.latestVersion ?? this.dismissedVersion;
        }
    }
});
