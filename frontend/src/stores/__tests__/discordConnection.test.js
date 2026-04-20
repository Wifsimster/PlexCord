import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';

// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
    EventsOn: vi.fn(),
    EventsOff: vi.fn()
}));

// Mock backend calls
vi.mock('../../../wailsjs/go/main/App', () => ({
    IsDiscordConnected: vi.fn().mockResolvedValue(false),
    GetConnectionHistory: vi.fn().mockResolvedValue({
        discordLastConnected: null
    }),
    GetDiscordRetryState: vi.fn().mockResolvedValue(null),
    RetryDiscordConnection: vi.fn().mockResolvedValue(undefined),
    ConnectDiscord: vi.fn().mockResolvedValue(undefined),
    GetErrorInfo: vi.fn().mockResolvedValue({
        code: 'DISCORD_NOT_RUNNING',
        title: 'Discord Not Running',
        description: 'Discord is not running',
        suggestion: 'Please start Discord',
        retryable: true
    })
}));

import { useDiscordConnectionStore } from '../discordConnection';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import { IsDiscordConnected, GetConnectionHistory, GetDiscordRetryState, RetryDiscordConnection, ConnectDiscord, GetErrorInfo } from '../../../wailsjs/go/main/App';

describe('discordConnection store', () => {
    let store;

    beforeEach(() => {
        setActivePinia(createPinia());
        store = useDiscordConnectionStore();
        vi.clearAllMocks();
    });

    describe('initial state', () => {
        it('has correct default values', () => {
            expect(store.connected).toBe(false);
            expect(store.lastConnected).toBeNull();
            expect(store.retryState).toBeNull();
            expect(store.error).toBeNull();
            expect(store.loading).toBe(false);
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

            it('returns true when retryState.isRetrying is true', () => {
                store.retryState = { isRetrying: true };
                expect(store.isRetrying).toBe(true);
            });
        });

        describe('hasError', () => {
            it('returns false when no error', () => {
                expect(store.hasError).toBe(false);
            });

            it('returns true when error is set', () => {
                store.error = { code: 'TEST' };
                expect(store.hasError).toBe(true);
            });
        });

        describe('statusLabel', () => {
            it('returns "Connected" when connected', () => {
                store.connected = true;
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

            it('returns "Disconnected" when there is an error', () => {
                store.error = { code: 'DISCORD_NOT_RUNNING' };
                expect(store.statusLabel).toBe('Disconnected');
            });

            it('returns "Not Connected" by default', () => {
                expect(store.statusLabel).toBe('Not Connected');
            });

            it('prioritizes "Connected" over other states', () => {
                store.connected = true;
                store.loading = true;
                store.error = { code: 'TEST' };
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
            it('registers event listeners for Discord events', () => {
                store.setupEventListeners();

                expect(EventsOn).toHaveBeenCalledTimes(3);
                expect(EventsOn).toHaveBeenCalledWith('DiscordConnected', expect.any(Function));
                expect(EventsOn).toHaveBeenCalledWith('DiscordDisconnected', expect.any(Function));
                expect(EventsOn).toHaveBeenCalledWith('DiscordRetryState', expect.any(Function));
            });
        });

        describe('cleanup', () => {
            it('removes event listeners and resets initialized', () => {
                store.initialized = true;
                store.cleanup();

                expect(EventsOff).toHaveBeenCalledWith('DiscordConnected');
                expect(EventsOff).toHaveBeenCalledWith('DiscordDisconnected');
                expect(EventsOff).toHaveBeenCalledWith('DiscordRetryState');
                expect(store.initialized).toBe(false);
            });
        });

        describe('refreshStatus', () => {
            it('fetches and sets connection status from backend', async () => {
                IsDiscordConnected.mockResolvedValue(true);
                GetConnectionHistory.mockResolvedValue({
                    discordLastConnected: '2026-01-01T00:00:00Z'
                });
                GetDiscordRetryState.mockResolvedValue({ isRetrying: false, attemptNumber: 0 });

                await store.refreshStatus();

                expect(store.connected).toBe(true);
                expect(store.lastConnected).toBe('2026-01-01T00:00:00Z');
                expect(store.retryState).toEqual({ isRetrying: false, attemptNumber: 0 });
            });

            it('handles errors gracefully', async () => {
                IsDiscordConnected.mockRejectedValue(new Error('fetch failed'));
                const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

                await store.refreshStatus();

                expect(store.connected).toBe(false);
                consoleSpy.mockRestore();
            });
        });

        describe('retry', () => {
            it('calls RetryDiscordConnection and manages loading state', async () => {
                RetryDiscordConnection.mockResolvedValue(undefined);

                const promise = store.retry();
                expect(store.loading).toBe(true);

                await promise;
                expect(store.loading).toBe(false);
                expect(RetryDiscordConnection).toHaveBeenCalled();
            });

            it('clears loading even if retry fails', async () => {
                RetryDiscordConnection.mockRejectedValue(new Error('retry failed'));

                await expect(store.retry()).rejects.toThrow('retry failed');
                expect(store.loading).toBe(false);
            });
        });

        describe('connect', () => {
            it('calls ConnectDiscord with client ID and manages loading', async () => {
                ConnectDiscord.mockResolvedValue(undefined);

                const promise = store.connect('my-client-id');
                expect(store.loading).toBe(true);

                await promise;
                expect(store.loading).toBe(false);
                expect(ConnectDiscord).toHaveBeenCalledWith('my-client-id');
            });

            it('defaults to empty string client ID', async () => {
                ConnectDiscord.mockResolvedValue(undefined);

                await store.connect();
                expect(ConnectDiscord).toHaveBeenCalledWith('');
            });

            it('sets error and re-throws on failure', async () => {
                const error = new Error('connect failed');
                ConnectDiscord.mockRejectedValue(error);

                await expect(store.connect('id')).rejects.toThrow('connect failed');
                expect(store.loading).toBe(false);
                expect(store.error).toBe(error);
            });
        });

        describe('setError', () => {
            it('fetches error info from backend', async () => {
                await store.setError('DISCORD_NOT_RUNNING');

                expect(GetErrorInfo).toHaveBeenCalledWith('DISCORD_NOT_RUNNING');
                expect(store.error).toEqual({
                    code: 'DISCORD_NOT_RUNNING',
                    title: 'Discord Not Running',
                    description: 'Discord is not running',
                    suggestion: 'Please start Discord',
                    retryable: true
                });
            });

            it('uses fallback error when GetErrorInfo fails', async () => {
                GetErrorInfo.mockRejectedValue(new Error('backend down'));

                await store.setError('UNKNOWN_CODE');

                expect(store.error).toEqual({
                    code: 'UNKNOWN_CODE',
                    title: 'Connection Error',
                    description: 'Failed to connect to Discord',
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
                const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

                await store.initialize();
                expect(store.initialized).toBe(true);
                expect(EventsOn).toHaveBeenCalledTimes(3);

                // Second call should be a no-op
                await store.initialize();
                expect(EventsOn).toHaveBeenCalledTimes(3);

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

            it('DiscordConnected sets connected state and clears error', () => {
                store.error = { code: 'old error' };
                eventHandlers['DiscordConnected']();

                expect(store.connected).toBe(true);
                expect(store.error).toBeNull();
                expect(store.lastConnected).toBeTruthy();
            });

            it('DiscordDisconnected sets disconnected state with error', async () => {
                store.connected = true;
                await eventHandlers['DiscordDisconnected']({ code: 'DISCORD_NOT_RUNNING' });

                expect(store.connected).toBe(false);
                expect(GetErrorInfo).toHaveBeenCalledWith('DISCORD_NOT_RUNNING');
            });

            it('DiscordDisconnected falls back to DISCORD_NOT_RUNNING code', async () => {
                await eventHandlers['DiscordDisconnected']({});

                expect(GetErrorInfo).toHaveBeenCalledWith('DISCORD_NOT_RUNNING');
            });

            it('DiscordRetryState updates retry state', () => {
                const state = { isRetrying: true, attemptNumber: 2 };
                eventHandlers['DiscordRetryState'](state);

                expect(store.retryState).toEqual(state);
            });
        });
    });
});
