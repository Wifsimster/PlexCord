import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';

// Running in a plain browser (no Wails backend): install the dev mock so
// pages render with realistic data. Excluded from production builds.
if (import.meta.env.DEV && !window.go) {
    const { installWailsMock } = await import('./dev/wailsMock');
    installWailsMock();
}

import { definePreset } from '@primeuix/themes';
import Aura from '@primeuix/themes/aura';
import PrimeVue from 'primevue/config';
import ConfirmationService from 'primevue/confirmationservice';
import ToastService from 'primevue/toastservice';
import Tooltip from 'primevue/tooltip';

import '@/assets/tokens.css';
import '@/assets/tailwind.css';
import '@/assets/styles.scss';
import '@/assets/transitions.css';
import '@/assets/components.css';

// Theme is applied pre-mount so there is no flash of the wrong scheme.
// Persisted in localStorage['plexcord-theme'] ('dark' | 'light'), default dark.
const savedTheme = localStorage.getItem('plexcord-theme') === 'light' ? 'light' : 'dark';
document.documentElement.classList.toggle('dark', savedTheme === 'dark');

const PlexCordPreset = definePreset(Aura, {
    primitive: {
        signal: {
            50: '#EEF5FF',
            100: '#D9E9FF',
            200: '#B3D3FF',
            300: '#85B8FF',
            400: '#5C9EFF',
            500: '#3B82F6',
            600: '#2A6AE0',
            700: '#2456B8',
            800: '#1F468F',
            900: '#1C3A72',
            950: '#142343'
        }
    },
    semantic: {
        primary: {
            50: '{signal.50}',
            100: '{signal.100}',
            200: '{signal.200}',
            300: '{signal.300}',
            400: '{signal.400}',
            500: '{signal.500}',
            600: '{signal.600}',
            700: '{signal.700}',
            800: '{signal.800}',
            900: '{signal.900}',
            950: '{signal.950}'
        },
        transitionDuration: '0.15s',
        focusRing: { width: '2px', style: 'solid', color: '{primary.color}', offset: '2px' },
        colorScheme: {
            light: {
                surface: {
                    0: '#FFFFFF',
                    50: '#F7F7F8',
                    100: '#EEEEF0',
                    200: '#DCDCE1',
                    300: '#B9BAC3',
                    400: '#8B8C98',
                    500: '#6E6F7B',
                    600: '#4B4C57',
                    700: '#2E2F38',
                    800: '#1E1F26',
                    900: '#121317',
                    950: '#0B0C0F'
                },
                primary: {
                    color: '{signal.600}',
                    contrastColor: '#FFFFFF',
                    hoverColor: '{signal.700}',
                    activeColor: '{signal.800}'
                },
                highlight: {
                    background: 'rgba(42,106,224,.10)',
                    focusBackground: 'rgba(42,106,224,.16)',
                    color: '#17181D',
                    focusColor: '#17181D'
                },
                formField: {
                    background: '#F1F1F3',
                    disabledBackground: '#F1F1F3',
                    borderColor: '#E3E3E8',
                    hoverBorderColor: '#C9CAD1',
                    focusBorderColor: '{signal.600}',
                    color: '#17181D',
                    placeholderColor: '#8B8C98',
                    floatLabelColor: '#5C5D68'
                },
                text: {
                    color: '#17181D',
                    hoverColor: '#0B0C0F',
                    mutedColor: '#5C5D68',
                    hoverMutedColor: '#4B4C57'
                },
                content: {
                    background: '#FFFFFF',
                    hoverBackground: '#F1F1F3',
                    borderColor: '#E3E3E8',
                    color: '#17181D'
                },
                overlay: {
                    modal: { background: '#FFFFFF', borderColor: '#E3E3E8', color: '#17181D' },
                    popover: { background: '#FFFFFF', borderColor: '#E3E3E8', color: '#17181D' }
                }
            },
            dark: {
                /* NB: --p-surface-900 is deliberately #17181D (our panel step, --pc-surface-850).
                   The true near-black rail value #121317 is available only via --pc-overlay. */
                surface: {
                    0: '#FFFFFF',
                    50: '#F7F7F8',
                    100: '#EEEEF0',
                    200: '#DCDCE1',
                    300: '#B9BAC3',
                    400: '#8B8C98',
                    500: '#6E6F7B',
                    600: '#4B4C57',
                    700: '#2E2F38',
                    800: '#1E1F26',
                    900: '#17181D',
                    950: '#0B0C0F'
                },
                primary: {
                    color: '{signal.400}',
                    contrastColor: '#0B0C0F',
                    hoverColor: '{signal.300}',
                    activeColor: '{signal.500}'
                },
                highlight: {
                    background: 'rgba(92,158,255,.16)',
                    focusBackground: 'rgba(92,158,255,.24)',
                    color: 'rgba(255,255,255,.92)',
                    focusColor: 'rgba(255,255,255,.92)'
                },
                formField: {
                    background: '{surface.800}',
                    disabledBackground: '{surface.800}',
                    borderColor: '#26272F',
                    hoverBorderColor: '{surface.600}',
                    focusBorderColor: '{signal.400}',
                    color: '#EDEDF0',
                    placeholderColor: '{surface.500}',
                    floatLabelColor: '{surface.400}'
                },
                text: {
                    color: '#EDEDF0',
                    hoverColor: '#FFFFFF',
                    mutedColor: '#8B8C98',
                    hoverMutedColor: '#A5A6B1'
                },
                content: {
                    background: '{surface.900}',
                    hoverBackground: '{surface.800}',
                    borderColor: '#26272F',
                    color: '#EDEDF0'
                },
                overlay: {
                    modal: { background: '#121317', borderColor: '#26272F', color: '#EDEDF0' },
                    popover: { background: '#121317', borderColor: '#26272F', color: '#EDEDF0' }
                }
            }
        }
    },
    components: {
        button: { root: { borderRadius: '6px', paddingX: '0.75rem', paddingY: '0.4375rem' } },
        card: { root: { borderRadius: '10px', background: '{content.background}' } },
        dialog: { root: { borderRadius: '10px' } },
        inputtext: { root: { borderRadius: '6px', paddingY: '0.4375rem' } },
        inputnumber: { root: { borderRadius: '6px' } },
        toggleswitch: {
            root: { width: '2.25rem', height: '1.25rem' },
            handle: { size: '0.875rem' }
        },
        toast: { root: { borderRadius: '8px' } }
    }
});

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(PrimeVue, {
    theme: {
        preset: PlexCordPreset,
        options: {
            darkModeSelector: '.dark'
        }
    }
});
app.use(ToastService);
app.use(ConfirmationService);
// v-tooltip is a directive — the auto-import resolver only covers components.
app.directive('tooltip', Tooltip);

app.mount('#app');
