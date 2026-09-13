import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ref, nextTick } from 'vue';
import { mount, flushPromises } from '@vue/test-utils';

// The theme moved out of the topbar into Settings → App. Settings pulls in the
// whole Wails bridge, PrimeVue and the stores, so everything outside the theme
// select is mocked away — these tests only exercise that control's wiring.
const setDarkMode = vi.fn();
const isDarkTheme = ref(true);

vi.mock('@/layout/composables/layout', () => ({
    useLayout: () => ({ isDarkTheme, setDarkMode, toggleDarkMode: vi.fn() })
}));

vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }));
// Partial: the real createI18n still has to build the app-wide instance that
// @/i18n (pulled in through the utils) creates at import time.
vi.mock('vue-i18n', async (importOriginal) => ({
    ...(await importOriginal()),
    useI18n: () => ({ t: (key) => key, locale: ref('en') })
}));
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: vi.fn() }) }));
vi.mock('primevue/useconfirm', () => ({ useConfirm: () => ({ require: vi.fn() }) }));
vi.mock('primevue/inputtext', () => ({ default: { name: 'InputText', template: '<input />' } }));
vi.mock('primevue/inputnumber', () => ({ default: { name: 'InputNumber', template: '<input />' } }));
vi.mock('primevue/toggleswitch', () => ({ default: { name: 'ToggleSwitch', template: '<button />' } }));
vi.mock('primevue/dialog', () => ({ default: { name: 'Dialog', template: '<div />' } }));
vi.mock('@/components/DiscordSpecimen.vue', () => ({ default: { name: 'DiscordSpecimen', template: '<div />' } }));
vi.mock('@/components/settings/SavedIndicator.vue', () => ({ default: { name: 'SavedIndicator', template: '<span />' } }));
vi.mock('@/composables/usePlayback', () => ({
    usePlayback: () => ({ currentTrack: ref(null), hasActiveSession: ref(false) })
}));
vi.mock('@/composables/useVersion', () => ({
    useVersion: () => ({ version: ref('4.3.0'), commit: ref('abc1234'), buildDate: ref('') })
}));
vi.mock('@/stores/setup', () => ({ useSetupStore: () => ({ resetWizard: vi.fn() }) }));
vi.mock('@/stores/presence', () => ({ usePresenceStore: () => ({ paused: false }) }));
vi.mock('@/stores/updates', () => ({
    useUpdatesStore: () => ({ initialize: vi.fn(), updateReady: false, installing: false, showUpdatePanel: false, progress: 0, info: null, canSelfUpdate: true })
}));
// storeToRefs expects a real Pinia store; the mocks above are plain objects.
vi.mock('pinia', async (importOriginal) => ({
    ...(await importOriginal()),
    storeToRefs: (store) => Object.fromEntries(Object.entries(store).map(([key, value]) => [key, ref(value)]))
}));
vi.mock('../../../../wailsjs/runtime/runtime', () => ({ BrowserOpenURL: vi.fn() }));
vi.mock('../../../../wailsjs/go/main/App', () => ({
    GetPollingInterval: vi.fn(async () => 2),
    SetPollingInterval: vi.fn(),
    GetAutoStart: vi.fn(async () => false),
    SetAutoStart: vi.fn(),
    GetMinimizeToTray: vi.fn(async () => true),
    SetMinimizeToTray: vi.fn(),
    GetStartMinimized: vi.fn(async () => false),
    SetStartMinimized: vi.fn(),
    GetStartMinimizedOnLogin: vi.fn(async () => true),
    SetStartMinimizedOnLogin: vi.fn(),
    GetAutoUpdateCheck: vi.fn(async () => true),
    SetAutoUpdateCheck: vi.fn(),
    GetDiscordClientID: vi.fn(async () => '123'),
    GetDefaultDiscordClientID: vi.fn(async () => '123'),
    SaveDiscordClientID: vi.fn(),
    ValidateDiscordClientID: vi.fn(),
    ConnectDiscord: vi.fn(),
    DisconnectDiscord: vi.fn(),
    TestDiscordPresence: vi.fn(),
    OpenReleasesPage: vi.fn(),
    OpenReleaseURL: vi.fn(),
    ResetApplication: vi.fn(),
    GetHideWhenPaused: vi.fn(async () => ({ enabled: false, delaySeconds: 0 })),
    SetHideWhenPaused: vi.fn(),
    GetPresenceFormat: vi.fn(async () => ({ detailsFormat: '{track}', stateFormat: '{artist}' })),
    SetPresenceFormat: vi.fn(),
    GetPresenceOptions: vi.fn(async () => ({ activityStyle: 'media', statusDisplay: 'state', artworkLookup: true })),
    SetPresenceOptions: vi.fn(),
    GetServers: vi.fn(async () => []),
    AddServer: vi.fn(),
    RemoveServer: vi.fn(),
    SetServerActive: vi.fn(),
    ValidatePlexConnection: vi.fn(),
    GetPlexToken: vi.fn(async () => ''),
    DiscoverPlexServers: vi.fn()
}));

const Settings = (await import('@/views/pages/Settings.vue')).default;

// The scroll-spy observer is set up on mount; jsdom has no IntersectionObserver.
class NoopObserver {
    observe() {}
    disconnect() {}
}

const mountSettings = async () => {
    const wrapper = mount(Settings, {
        global: {
            mocks: { $t: (key) => key },
            stubs: { Transition: false, RouterLink: true }
        }
    });
    await flushPromises();
    return wrapper;
};

beforeEach(() => {
    vi.clearAllMocks();
    isDarkTheme.value = true;
    vi.stubGlobal('IntersectionObserver', NoopObserver);
});

describe('Settings → App theme select', () => {
    it('offers dark and light, with the active theme selected', async () => {
        const wrapper = await mountSettings();
        const select = wrapper.get('#theme-select');

        expect(select.findAll('option').map((o) => o.attributes('value'))).toEqual(['dark', 'light']);
        expect(select.element.value).toBe('dark');

        wrapper.unmount();
    });

    it('switches to the light theme when light is picked', async () => {
        const wrapper = await mountSettings();

        await wrapper.get('#theme-select').setValue('light');

        expect(setDarkMode).toHaveBeenCalledTimes(1);
        expect(setDarkMode).toHaveBeenCalledWith(false);

        wrapper.unmount();
    });

    it('switches back to the dark theme when dark is picked', async () => {
        isDarkTheme.value = false;
        const wrapper = await mountSettings();
        expect(wrapper.get('#theme-select').element.value).toBe('light');

        await wrapper.get('#theme-select').setValue('dark');

        expect(setDarkMode).toHaveBeenCalledWith(true);

        wrapper.unmount();
    });

    it('follows the theme when it changes elsewhere (Alt+D)', async () => {
        const wrapper = await mountSettings();

        isDarkTheme.value = false;
        await nextTick();

        expect(wrapper.get('#theme-select').element.value).toBe('light');

        wrapper.unmount();
    });
});
