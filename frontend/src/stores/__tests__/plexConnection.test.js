import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';

// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
    EventsOn: vi.fn(),
    EventsOff: vi.fn()
}));

// Mock backend calls
vi.mock('../../../wailsjs/go/main/App', () => ({
    GetPlexConnectionStatus: vi.fn().mockResolvedValue({
        connected: false,
        polling: false,
        inErrorState: false,
        serverUrl: '',
        userId: '',
        userName: ''
    }),
    GetConnectionHistory: vi.fn().mockResolvedValue({
        plexLastConnected: null
    }),
    GetPlexRetryState: vi.fn().mockResolvedValue(null),
    RetryPlexConnection: vi.fn().mockResolvedValue(undefined),
    GetErrorInfo: vi.fn().mockResolvedValue({
        code: 'PLEX_UNREACHABLE',
        title: 'Plex Unreachable',
        description: 'Cannot reach Plex server',
        suggestion: 'Check your connection',
        retryable: true
    })
}));

import { usePlexConnectionStore } from '../plexConnection';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import { GetPlexConnectionStatus, GetConnectionHistory, GetPlexRetryState, RetryPlexConnection, GetErrorInfo } from '../../../wailsjs/go/main/App';

describe('plexConnection store', () => {
    let store;

    beforeEach(() => {
        setActivePinia(createPinia());
        store = usePlexConnectionStore();
        vi.clearAllMocks();
    });

    describe('initial state', () => {
        it('has correct default values', () => {
            expect(store.connected).toBe(false);
            expect(store.polling).toBe(false);
            expect(store.inErrorState).toBe(false);
            expect(store.serverUrl).toBe('');
            expect(store.userId).toBe('');
            expect(store.userName).toBe('');
            expect(store.lastConnected).toBeNull();
            expect(store.retryState).toBeNull();
            expect(store.loading).toBe(false);
            expect(store.error).toBeNull();
            expect(store.initialized).toBe(false);
        });
    });

    describe('getters', () => {
        describe('lastConnectedRelative', () => {
            it('returns "Never" when lastConnected is null', () => {
                expect(store.lastConnectedRelative).toBe('Never');
            });

            it('returns "Just now" for recent timestamp', () => {
                store.lastConnected = new Date().toISOString();
                expect(store.lastConnectedRelative).toBe('Just now');
            });
        });

        describe('isRetrying', () => {
            it('returns false when retryState is null', () => {
                expect(store.isRetrying).toBe(false);
            });

            it('returns false when retryState.isRetrying is false', () => {
                store.retryState = { isRetrying: false };
                expect(store.isRetrying).toBe(false);
            });

            it('returns true when retryState.isRetrying is true', () => {
                store.retryState = { isRetrying: true };
                expect(store.isRetrying).toBe(true);
            });
        });

        describe('hasError', () => {
            it('returns false when no error and not in error state', () => {
                expect(store.hasError).toBe(false);
            });

            it('returns true when inErrorState is true', () => {
                store.inErrorState = true;
                expect(store.hasError).toBe(true);
            });

            it('returns true when error is set', () => {
                store.error = { code: 'TEST' };
                expect(store.hasError).toBe(true);
            });
        });

        describe('statusLabel', () => {
            it('returns "Connected" when connected and polling', () => {
                store.connected = true;
                store.polling = true;
                expect(store.statusLabel).toBe('Connected');
            });

            it('returns "Connecting..." when loading', () => {
                store.loading = true;
                expect(store.statusLabel).toBe('Connecting...');
            });

            it('returns "Retrying..." when retry is active', () => {
                store.retryState = { isRetrying: true };
                expect(store.statusLabel).toBe('Retrying...');
            });

            it('returns "Disconnected" when in error state', () => {
                store.inErrorState = true;
                expect(store.statusLabel).toBe('Disconnected');
            });

            it('returns "Not Connected" by default', () => {
                expect(store.statusLabel).toBe('Not Connected');
            });

            it('prioritizes "Connected" over other states', () => {
                store.connected = true;
                store.polling = true;
                store.loading = true;
                store.inErrorState = true;
                expect(store.statusLabel).toBe('Connected');
            });

            it('prioritizes "Connecting..." over "Retrying..."', () => {
                store.loading = true;
                store.retryState = { isRetrying: true };
                expect(store.statusLabel).toBe('Connecting...');
            });
        });
    });

    describe('actions', () => {
        describe('setupEventListeners', () => {
            it('registers event listeners for Plex events', () => {
                store.setupEventListeners();

                expect(EventsOn).toHaveBeenCalledTimes(3);
                expect(EventsOn).toHaveBeenCalledWith('PlexConnectionError', expect.any(Function));
                expect(EventsOn).toHaveBeenCalledWith('PlexConnectionRestored', expect.any(Function));
                expect(EventsOn).toHaveBeenCalledWith('PlexRetryState', expect.any(Function));
            });
        });

        describe('cleanup', () => {
            it('removes event listeners and resets initialized', () => {
                store.initialized = true;
                store.cleanup();

                expect(EventsOff).toHaveBeenCalledWith('PlexConnectionError');
                expect(EventsOff).toHaveBeenCalledWith('PlexConnectionRestored');
                expect(EventsOff).toHaveBeenCalledWith('PlexRetryState');
                expect(store.initialized).toBe(false);
            });
        });

        describe('refreshStatus', () => {
            it('fetches and sets connection status from backend', async () => {
                GetPlexConnectionStatus.mockResolvedValue({
                    connected: true,
                    polling: true,
                    inErrorState: false,
                    serverUrl: 'http://plex:32400',
                    userId: '123',
                    userName: 'TestUser'
                });
                GetConnectionHistory.mockResolvedValue({
                    plexLastConnected: '2026-01-01T00:00:00Z'
                });
                GetPlexRetryState.mockResolvedValue({ isRetrying: false, attemptNumber: 0 });

                await store.refreshStatus();

                expect(store.connected).toBe(true);
                expect(store.polling).toBe(true);
                expect(store.inErrorState).toBe(false);
                expect(store.serverUrl).toBe('http://plex:32400');
                expect(store.userId).toBe('123');
                expect(store.userName).toBe('TestUser');
                expect(store.lastConnected).toBe('2026-01-01T00:00:00Z');
                expect(store.retryState).toEqual({ isRetrying: false, attemptNumber: 0 });
            });

            it('handles errors gracefully', async () => {
                GetPlexConnectionStatus.mockRejectedValue(new Error('fetch failed'));
                const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

                await store.refreshStatus();

                // Should not throw, store retains previous state
                expect(store.connected).toBe(false);
                consoleSpy.mockRestore();
            });
        });

        describe('retry', () => {
            it('calls RetryPlexConnection and manages loading state', async () => {
                RetryPlexConnection.mockResolvedValue(undefined);

                const promise = store.retry();
                expect(store.loading).toBe(true);

                await promise;
                expect(store.loading).toBe(false);
                expect(RetryPlexConnection).toHaveBeenCalled();
            });

            it('clears loading even if retry fails', async () => {
                RetryPlexConnection.mockRejectedValue(new Error('retry failed'));

                await expect(store.retry()).rejects.toThrow('retry failed');
                expect(store.loading).toBe(false);
            });
        });

        describe('setError', () => {
            it('fetches error info from backend', async () => {
                await store.setError('PLEX_UNREACHABLE');

                expect(GetErrorInfo).toHaveBeenCalledWith('PLEX_UNREACHABLE');
                expect(store.error).toEqual({
                    code: 'PLEX_UNREACHABLE',
                    title: 'Plex Unreachable',
                    description: 'Cannot reach Plex server',
                    suggestion: 'Check your connection',
                    retryable: true
                });
            });

            it('uses fallback error when GetErrorInfo fails', async () => {
                GetErrorInfo.mockRejectedValue(new Error('backend down'));

                await store.setError('UNKNOWN_CODE');

                expect(store.error).toEqual({
                    code: 'UNKNOWN_CODE',
                    title: 'Connection Error',
                    description: 'Failed to connect to Plex',
                    suggestion: 'Please check your connection and try again.',
                    retryable: true
                });
            });
        });

        describe('clearError', () => {
            it('clears error state', () => {
                store.error = { code: 'TEST' };
                store.clearError();
                expect(store.error).toBeNull();
            });
        });

        describe('initialize', () => {
            it('only initializes once', async () => {
                // Suppress console.log from autoReconnect
                const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

                await store.initialize();
                expect(store.initialized).toBe(true);
                expect(EventsOn).toHaveBeenCalledTimes(3);

                // Second call should be a no-op
                await store.initialize();
                expect(EventsOn).toHaveBeenCalledTimes(3); // unchanged

                consoleSpy.mockRestore();
            });
        });

        describe('event handlers', () => {
            let eventHandlers;

            beforeEach(() => {
                eventHandlers = {};
                EventsOn.mockImplementation((event, handler) => {
                    eventHandlers[event] = handler;
                });
                store.setupEventListeners();
            });

            it('PlexConnectionRestored sets connected state', () => {
                eventHandlers['PlexConnectionRestored']();

                expect(store.connected).toBe(true);
                expect(store.inErrorState).toBe(false);
                expect(store.lastConnected).toBeTruthy();
                expect(store.error).toBeNull();
            });

            it('PlexConnectionError sets error state', async () => {
                await eventHandlers['PlexConnectionError']({ errorCode: 'PLEX_UNREACHABLE' });

                expect(store.connected).toBe(false);
                expect(store.inErrorState).toBe(true);
                expect(GetErrorInfo).toHaveBeenCalledWith('PLEX_UNREACHABLE');
            });

            it('PlexConnectionError falls back to PLEX_UNREACHABLE code', async () => {
                await eventHandlers['PlexConnectionError']({});

                expect(GetErrorInfo).toHaveBeenCalledWith('PLEX_UNREACHABLE');
            });

            it('PlexRetryState updates retry state', () => {
                const state = { isRetrying: true, attemptNumber: 3 };
                eventHandlers['PlexRetryState'](state);

                expect(store.retryState).toEqual(state);
            });
        });
    });
});
