import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';

// AppLayout pulls in the whole shell (Wails bindings, PrimeVue toast, child
// views). Everything outside the keyboard handler is mocked away so these
// tests only exercise the shortcut wiring (spec §5.1 / §6.7).
const toggleDarkMode = vi.fn();
const presenceToggle = vi.fn();
const push = vi.fn();
const plexRetry = vi.fn();
const discordRetry = vi.fn();
const toastAdd = vi.fn();

let plexHasError = false;
let discordHasError = false;

vi.mock('@/layout/composables/layout', () => ({
    useLayout: () => ({ toggleDarkMode, isDarkTheme: { value: true } })
}));
vi.mock('@/composables/usePlayback', () => ({ usePlayback: () => ({}) }));
vi.mock('vue-router', () => ({
    useRouter: () => ({ push, currentRoute: { value: { path: '/' } } })
}));
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key) => key }) }));
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: toastAdd, removeGroup: vi.fn() }) }));
vi.mock('primevue/toast', () => ({ default: { name: 'Toast', template: '<div />' } }));
vi.mock('../AppFooter.vue', () => ({ default: { name: 'AppFooter', template: '<div />' } }));
vi.mock('../AppTopbar.vue', () => ({ default: { name: 'AppTopbar', template: '<div />' } }));
vi.mock('../../../wailsjs/go/main/App', () => ({ OpenReleaseURL: vi.fn(), OpenReleasesPage: vi.fn() }));
vi.mock('@/stores/plexConnection', () => ({
    usePlexConnectionStore: () => ({
        initialize: vi.fn(),
        cleanup: vi.fn(),
        retry: plexRetry,
        get hasError() {
            return plexHasError;
        }
    })
}));
vi.mock('@/stores/discordConnection', () => ({
    useDiscordConnectionStore: () => ({
        initialize: vi.fn(),
        cleanup: vi.fn(),
        retry: discordRetry,
        get hasError() {
            return discordHasError;
        }
    })
}));
vi.mock('@/stores/presence', () => ({
    usePresenceStore: () => ({ initialize: vi.fn(), toggle: presenceToggle })
}));
vi.mock('@/stores/updates', () => ({
    useUpdatesStore: () => ({ initialize: vi.fn(), cleanup: vi.fn(), shouldToast: false, updateReady: false, info: null, dismissToast: vi.fn(), restart: vi.fn() })
}));

const AppLayout = (await import('@/layout/AppLayout.vue')).default;

const mountLayout = () =>
    mount(AppLayout, {
        attachTo: document.body,
        global: { stubs: { RouterView: true, Transition: false } }
    });

const press = (init) => {
    const event = new KeyboardEvent('keydown', { bubbles: true, cancelable: true, ...init });
    (init.target ?? window).dispatchEvent(event);
    return event;
};

beforeEach(() => {
    vi.clearAllMocks();
    plexHasError = false;
    discordHasError = false;
    document.body.innerHTML = '';
});

describe('AppLayout keyboard shortcuts', () => {
    it('toggles the theme on Alt+D', () => {
        const wrapper = mountLayout();

        const event = press({ key: 'd', code: 'KeyD', altKey: true });

        expect(toggleDarkMode).toHaveBeenCalledTimes(1);
        expect(event.defaultPrevented).toBe(true);
        wrapper.unmount();
    });

    it('toggles the theme when Alt remaps the character (falls back to event.code)', () => {
        const wrapper = mountLayout();

        press({ key: '∂', code: 'KeyD', altKey: true });

        expect(toggleDarkMode).toHaveBeenCalledTimes(1);
        wrapper.unmount();
    });

    it('ignores Alt+D combined with other modifiers', () => {
        const wrapper = mountLayout();

        press({ key: 'd', code: 'KeyD', altKey: true, ctrlKey: true });
        press({ key: 'd', code: 'KeyD', altKey: true, shiftKey: true });
        press({ key: 'd', code: 'KeyD', altKey: true, metaKey: true });

        expect(toggleDarkMode).not.toHaveBeenCalled();
        wrapper.unmount();
    });

    it('ignores plain D and other Alt keys', () => {
        const wrapper = mountLayout();

        press({ key: 'd', code: 'KeyD' });
        press({ key: 'k', code: 'KeyK', altKey: true });

        expect(toggleDarkMode).not.toHaveBeenCalled();
        wrapper.unmount();
    });

    it('does not fire while an input has focus', () => {
        const wrapper = mountLayout();
        const input = document.createElement('input');
        document.body.appendChild(input);

        press({ key: 'd', code: 'KeyD', altKey: true, target: input });

        expect(toggleDarkMode).not.toHaveBeenCalled();
        wrapper.unmount();
    });

    it('still handles the Ctrl/⌘ shortcuts', () => {
        const wrapper = mountLayout();

        press({ key: 'p', ctrlKey: true });
        expect(presenceToggle).toHaveBeenCalledTimes(1);

        press({ key: ',', ctrlKey: true });
        expect(push).toHaveBeenCalledWith('/settings');

        plexHasError = true;
        press({ key: 'r', ctrlKey: true });
        expect(plexRetry).toHaveBeenCalledTimes(1);
        expect(toggleDarkMode).not.toHaveBeenCalled();

        wrapper.unmount();
    });

    it('removes the listener on unmount', () => {
        const wrapper = mountLayout();
        wrapper.unmount();

        press({ key: 'd', code: 'KeyD', altKey: true });

        expect(toggleDarkMode).not.toHaveBeenCalled();
    });
});
