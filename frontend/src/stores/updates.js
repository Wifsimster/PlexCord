import { defineStore } from 'pinia';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { CanSelfUpdate, CheckForUpdate, DownloadAndInstallUpdate, GetUpdateStatus, RestartApplication } from '../../wailsjs/go/main/App';

/**
 * Updates Store
 * Owns all update state and the update runtime event subscriptions so the
 * state survives page navigation: the backend checks for updates
 * automatically (at startup and periodically) and may download one in the
 * background while the user is on any page — or none at all.
 *
 * Event names must match internal/events/events.go verbatim.
 */
export const useUpdatesStore = defineStore('updates', {
    state: () => ({
        // Mirrors the backend updater state machine: idle | available | downloading | ready
        status: 'idle',
        info: null, // version.UpdateInfo from the last check / download
        progress: 0, // download progress 0-100
        canSelfUpdate: false,
        lastError: null,
        // Toast dedup: version the user dismissed the notification for
        dismissedVersion: null,
        initialized: false
    }),

    getters: {
        /**
         * An update exists but has not been applied yet
         */
        updateAvailable: (state) => !!state.info?.available || state.status === 'available',

        /**
         * Update applied on disk — restart pending
         */
        updateReady: (state) => state.status === 'ready',

        /**
         * A download/install is currently running
         */
        installing: (state) => state.status === 'downloading',

        /**
         * The About panel row shows for any known update, including one that
         * finished installing (whose info.available is false by then).
         */
        showUpdatePanel() {
            return this.updateAvailable || this.updateReady || this.installing;
        },

        /**
         * Whether the global "update" toast should be shown: an update is
         * ready to apply (or merely available on platforms that cannot
         * self-update), and the user has not dismissed it for this version.
         */
        shouldToast(state) {
            const version = state.info?.latestVersion;
            if (!version || version === state.dismissedVersion) return false;
            return this.updateReady || (this.updateAvailable && !state.canSelfUpdate);
        }
    },

    actions: {
        /**
         * Initialize: hydrate from the backend (an automatic check/download
         * may have completed before this page loaded) and register listeners.
         */
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

        /**
         * Hydrate state from the backend updater snapshot.
         * Listeners are registered before this runs, so any transition after
         * that point already advanced our state via events. Only apply the
         * snapshot while still idle — otherwise a snapshot dispatched before a
         * just-delivered event could roll the state backwards (e.g. clobber a
         * fresh 'ready' with a stale 'downloading').
         */
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

        /**
         * Setup Wails event listeners for the update lifecycle.
         * IMPORTANT: keep the cancel functions returned by EventsOn and call
         * those in cleanup() — EventsOff(name) would remove ALL listeners for
         * the event name, including subscriptions owned by other components.
         */
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

        /**
         * Cleanup event listeners (via their cancel functions — see above)
         */
        cleanup() {
            (this._cancelFns || []).forEach((cancel) => cancel());
            this._cancelFns = [];
            this.initialized = false;
        },

        /**
         * Manual "Check for updates" (Settings button)
         */
        async checkNow() {
            this.info = await CheckForUpdate();
            if (this.info?.available && this.status === 'idle') {
                this.status = 'available';
            }
            return this.info;
        },

        /**
         * Manual "Download & install" (Settings button). Progress and
         * completion arrive through the runtime events; rejection carries
         * the error for the caller to surface.
         */
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

        /**
         * Restart the app to apply an installed update
         */
        async restart() {
            await RestartApplication();
        },

        /**
         * Dismiss the update toast for the current version so periodic
         * re-checks don't nag about it again.
         */
        dismissToast() {
            this.dismissedVersion = this.info?.latestVersion ?? this.dismissedVersion;
        }
    }
});
