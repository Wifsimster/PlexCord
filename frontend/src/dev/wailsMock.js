/**
 * Dev-only mock of the Wails bridge (window.go / window.runtime).
 *
 * Lets the frontend run in a plain browser (`npm run dev`) without the Go
 * backend, so pages can be developed and visually tested. Never bundled in
 * production: only imported behind `import.meta.env.DEV` from main.js.
 *
 * URL flags:
 *   ?mock=empty  — nothing playing, everything disconnected (empty states)
 *   ?mock=error  — Plex unreachable + auto-retrying (error/countdown states)
 *   ?mock=users  — multiple Plex users (wizard Select User grid)
 */

const ALBUM_ART =
    'data:image/svg+xml;utf8,' +
    encodeURIComponent(`<svg xmlns="http://www.w3.org/2000/svg" width="300" height="300">
        <defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0" stop-color="#e5a00d"/><stop offset="1" stop-color="#5865F2"/>
        </linearGradient></defs>
        <rect width="300" height="300" fill="url(#g)"/>
        <circle cx="150" cy="150" r="70" fill="rgba(0,0,0,0.35)"/>
        <circle cx="150" cy="150" r="18" fill="rgba(255,255,255,0.85)"/>
    </svg>`);

function nowIso(offsetMs = 0) {
    return new Date(Date.now() + offsetMs).toISOString();
}

export function installWailsMock() {
    const params = new URLSearchParams(window.location.search);
    const scenario = params.get('mock') || '';
    const empty = scenario === 'empty';
    const plexError = scenario === 'error';
    const multiUsers = scenario === 'users';

    const listeners = new Map();

    const session =
        empty || plexError
            ? null
            : {
                  sessionKey: 'mock-1',
                  track: 'Midnight City',
                  artist: 'M83',
                  album: 'Hurry Up, We’re Dreaming',
                  thumb: '/library/thumb/mock',
                  thumbUrl: ALBUM_ART,
                  duration: 243000,
                  viewOffset: 97000,
                  state: 'playing',
                  playerName: 'Plexamp'
              };

    const state = {
        presencePaused: false,
        autoStart: true,
        minimizeToTray: true,
        hideWhenPaused: false,
        hideWhenPausedDelay: 0,
        pollingInterval: 5,
        presenceFormat: { detailsFormat: '{track}', stateFormat: 'by {artist}' },
        discordClientId: '',
        servers: empty ? [] : [{ name: 'Home Server', url: 'http://192.168.1.10:32400', userId: '1', userName: 'demo-user', active: true }]
    };

    // Matches the Go retry.RetryState JSON shape (nextRetryIn is NANOSECONDS).
    const idleRetry = { isRetrying: false, attemptNumber: 0, nextRetryIn: 0, maxIntervalReached: false };

    // ?mock=error — Plex unreachable, auto-retry countdown re-arming forever.
    const retrySim = plexError ? { attemptNumber: 3, deadline: Date.now() + 12000 } : null;
    const plexRetryState = () =>
        retrySim
            ? {
                  isRetrying: true,
                  attemptNumber: retrySim.attemptNumber,
                  nextRetryIn: Math.max(0, retrySim.deadline - Date.now()) * 1e6,
                  nextRetryAt: new Date(retrySim.deadline).toISOString(),
                  lastError: 'connection refused',
                  lastErrorCode: 'PLEX_UNREACHABLE',
                  maxIntervalReached: false
              }
            : { ...idleRetry };

    const appMocks = {
        GetVersion: () => ({ version: '4.3.0-dev', commit: 'a1b2c3d4e5f6', buildDate: nowIso() }),
        CheckSetupComplete: () => true,
        CheckForUpdate: () => ({ available: false, latestVersion: '4.3.0' }),
        // Browsers can't swap the running binary; false exercises the
        // open-release-page fallback in Settings → About.
        CanSelfUpdate: () => false,
        DownloadAndInstallUpdate: () => {},
        RestartApplication: () => {},
        GetCurrentSession: () => session,
        IsPresencePaused: () => state.presencePaused,
        TogglePresencePause: () => {
            state.presencePaused = !state.presencePaused;
            return state.presencePaused;
        },
        GetPlexConnectionStatus: () => ({
            connected: !empty && !plexError,
            polling: !empty && !plexError,
            inErrorState: plexError,
            serverUrl: empty ? '' : 'http://192.168.1.10:32400',
            userId: empty ? '' : '1',
            userName: empty ? '' : 'demo-user'
        }),
        IsDiscordConnected: () => !empty,
        GetConnectionHistory: () => ({
            plexLastConnected: empty ? null : nowIso(-42000),
            discordLastConnected: empty ? null : nowIso(-42000)
        }),
        GetPlexRetryState: () => plexRetryState(),
        GetDiscordRetryState: () => ({ ...idleRetry }),
        RetryPlexConnection: () => {},
        RetryDiscordConnection: () => {},
        ConnectDiscord: () => {},
        DisconnectDiscord: () => {},
        GetErrorInfo: (code) =>
            code === 'PLEX_UNREACHABLE'
                ? {
                      code,
                      title: 'Plex server unreachable',
                      description: 'PlexCord could not reach your Plex server.',
                      suggestion: 'Check that the server is running and reachable from this computer.',
                      retryable: true
                  }
                : {
                      code,
                      title: 'Connection Error',
                      description: 'Mock error description',
                      suggestion: 'Mock suggestion',
                      retryable: true
                  },
        GetAutoStart: () => state.autoStart,
        SetAutoStart: (v) => {
            state.autoStart = v;
        },
        GetMinimizeToTray: () => state.minimizeToTray,
        SetMinimizeToTray: (v) => {
            state.minimizeToTray = v;
        },
        // Matches the Go App.GetHideWhenPaused map shape
        GetHideWhenPaused: () => ({ enabled: state.hideWhenPaused, delaySeconds: state.hideWhenPausedDelay }),
        SetHideWhenPaused: (enabled, delaySeconds) => {
            state.hideWhenPaused = enabled;
            state.hideWhenPausedDelay = Math.max(0, delaySeconds ?? 0);
        },
        GetPollingInterval: () => state.pollingInterval,
        SetPollingInterval: (v) => {
            state.pollingInterval = v;
        },
        GetPresenceFormat: () => ({ ...state.presenceFormat }),
        SetPresenceFormat: (details, stateFormat) => {
            state.presenceFormat = { detailsFormat: details, stateFormat };
        },
        GetDefaultDiscordClientID: () => '1234567890',
        // Like the backend: returns the custom ID when set, else the default
        GetDiscordClientID: () => state.discordClientId || '1234567890',
        SaveDiscordClientID: (id) => {
            state.discordClientId = id;
        },
        // Backend validates a Discord snowflake (17–20 digits); '' = default
        ValidateDiscordClientID: (id) => {
            if (id && !/^\d{17,20}$/.test(id)) {
                throw new Error('Discord Client ID must be a 17–20 digit number');
            }
        },
        GetPlexToken: () => (empty ? '' : 'mock-token'),
        GetServers: () => state.servers.map((s) => ({ ...s })),
        AddServer: (name, url, userId, userName) => {
            if (state.servers.some((s) => s.url === url)) {
                throw new Error('a server with this URL already exists');
            }
            state.servers.push({ name, url, userId: userId || '', userName: userName || '', active: true });
        },
        RemoveServer: (url) => {
            state.servers = state.servers.filter((s) => s.url !== url);
        },
        SetServerActive: (url, active) => {
            const server = state.servers.find((s) => s.url === url);
            if (server) server.active = active;
        },
        GetPlexUsers: () => {
            if (empty) return [];
            if (multiUsers) {
                return [
                    { id: '1', name: 'demo-user', thumb: '' },
                    { id: '2', name: 'kids', thumb: '' },
                    { id: '3', name: 'guest', thumb: '' }
                ];
            }
            return [{ id: '1', name: 'demo-user', thumb: '' }];
        },
        // Matches plex.Server (GDM discovery result shape). Superset of what
        // both consumers read: SetupPlex keys rows by `id`; Settings builds
        // URLs from `name`/`address`/`port`/`isLocal`. The first entry's URL
        // matches the pre-seeded server so the "Added" state is exercisable.
        DiscoverPlexServers: () =>
            empty
                ? []
                : [
                      { id: 'srv-1', name: 'Home Server', address: '192.168.1.10', port: '32400', version: '1.40.0.7998', isLocal: true },
                      { id: 'srv-2', name: 'Remote NAS', address: 'plex.example.com', port: '32400', version: '1.40.0.7998', isLocal: false }
                  ],
        // PIN auth: authorizes on the second status poll (~4s), like a user
        // completing the code at plex.tv/link.
        StartPlexPINAuth: () => {
            state.pinPolls = 0;
            return { pinCode: 'ABCD', pinID: 314159, authURL: 'https://plex.tv/link?pin=ABCD', expiresIn: 900, expiresAt: nowIso(900000) };
        },
        CheckPlexPINAuth: () => {
            state.pinPolls = (state.pinPolls || 0) + 1;
            return state.pinPolls >= 2 ? { authorized: true, authToken: 'mock-plex-token' } : { authorized: false, expired: false };
        },
        SavePlexToken: () => {},
        SaveServerURL: () => {},
        SavePlexUserSelection: () => {},
        StartSessionPolling: () => {},
        CompleteSetup: () => {},
        SkipSetup: () => {},
        // Matches plex.ValidationResult; fails in the ?mock=error scenario
        ValidatePlexConnection: () => {
            if (plexError) {
                throw new Error('connection refused');
            }
            return { success: true, serverName: 'Home Server', serverVersion: '1.40.0.7998', libraryCount: 3 };
        },
        GetResourceStats: () => ({ cpuPercent: 0.4, memoryMB: 38 }),
        IsPollingActive: () => !empty,
        TestDiscordPresence: () => {
            if (empty) {
                throw new Error('not connected to Discord');
            }
        },
        ResetApplication: () => {},
        OpenReleasesPage: () => {},
        OpenReleaseURL: () => {}
    };

    window.go = {
        main: {
            App: new Proxy(appMocks, {
                get(target, prop) {
                    if (prop in target) {
                        return async (...args) => target[prop](...args);
                    }
                    return async () => {
                        console.warn(`[wailsMock] unmocked App.${String(prop)} called`);
                        return null;
                    };
                }
            })
        }
    };

    const noop = () => {};
    window.runtime = new Proxy(
        {
            EventsOn: (name, cb) => {
                if (!listeners.has(name)) listeners.set(name, []);
                listeners.get(name).push(cb);
                return () =>
                    listeners.set(
                        name,
                        (listeners.get(name) || []).filter((f) => f !== cb)
                    );
            },
            // The wailsjs runtime.js wrapper routes EventsOn/EventsOnce
            // through EventsOnMultiple — without this, wrapper-based
            // subscriptions are silent no-ops in the browser.
            EventsOnMultiple: (name, cb) => window.runtime.EventsOn(name, cb),
            EventsOnce: (name, cb) => window.runtime.EventsOn(name, cb),
            EventsOff: (name) => listeners.delete(name),
            EventsEmit: (name, ...data) => (listeners.get(name) || []).forEach((cb) => cb(...data)),
            BrowserOpenURL: (url) => window.open(url, '_blank')
        },
        {
            get(target, prop) {
                return prop in target ? target[prop] : noop;
            }
        }
    );

    // Simulate playback progress so time-based UI (progress bars) animates.
    if (session) {
        setInterval(() => {
            session.viewOffset = (session.viewOffset + 1000) % session.duration;
            window.runtime.EventsEmit('PlaybackUpdated', { ...session });
        }, 1000);
    }

    // Simulate the Plex failure + auto-retry cycle (?mock=error). Mirrors the
    // backend: PlexConnectionError once, then a PlexRetryState per scheduled
    // attempt (the frontend ticks the countdown locally from nextRetryAt).
    if (retrySim) {
        setTimeout(() => {
            window.runtime.EventsEmit('PlexConnectionError', { errorCode: 'PLEX_UNREACHABLE' });
            window.runtime.EventsEmit('PlexRetryState', plexRetryState());
        }, 400);
        setInterval(() => {
            if (Date.now() >= retrySim.deadline) {
                retrySim.attemptNumber += 1;
                retrySim.deadline = Date.now() + 12000;
                window.runtime.EventsEmit('PlexConnectionError', { errorCode: 'PLEX_UNREACHABLE' });
                window.runtime.EventsEmit('PlexRetryState', plexRetryState());
            }
        }, 1000);
    }

    console.info(`[wailsMock] installed (${scenario || 'playing'} scenario)`);
}
