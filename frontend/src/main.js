import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import i18n from './i18n';

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

import '@fontsource-variable/bricolage-grotesque/standard.css';
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
        // "Ink" — interaction is monochrome (spec §1.2). The ramp is the
        // graphite surface ramp; which end is "primary" flips per scheme.
        ink: {
            50: '#FBFAF8',
            100: '#F2EEE8',
            200: '#E2DED8',
            300: '#C6C1B9',
            400: '#948E86',
            500: '#716B64',
            600: '#524D47',
            700: '#34302D',
            800: '#211F1D',
            900: '#1A1817',
            950: '#0E0D0C'
        }
    },
    semantic: {
        primary: {
            50: '{ink.50}',
            100: '{ink.100}',
            200: '{ink.200}',
            300: '{ink.300}',
            400: '{ink.400}',
            500: '{ink.500}',
            600: '{ink.600}',
            700: '{ink.700}',
            800: '{ink.800}',
            900: '{ink.900}',
            950: '{ink.950}'
        },
        transitionDuration: '0.15s',
        focusRing: { width: '2px', style: 'solid', color: '{text.muted.color}', offset: '2px' },
        colorScheme: {
            light: {
                surface: {
                    0: '#FFFFFF',
                    50: '#F5F3EF',
                    100: '#EEEBE6',
                    200: '#E2DED8',
                    300: '#C6C1B9',
                    400: '#948E86',
                    500: '#716B64',
                    600: '#524D47',
                    700: '#34302D',
                    800: '#211F1D',
                    900: '#121110',
                    950: '#0E0D0C'
                },
                primary: {
                    color: '{ink.900}',
                    contrastColor: '#FFFFFF',
                    hoverColor: '{ink.700}',
                    activeColor: '#000000'
                },
                highlight: {
                    background: 'rgba(26,24,23,.06)',
                    focusBackground: 'rgba(26,24,23,.10)',
                    color: '#1A1817',
                    focusColor: '#1A1817'
                },
                formField: {
                    background: '#F3F0EB',
                    disabledBackground: '#F3F0EB',
                    borderColor: '#E6E2DC',
                    hoverBorderColor: '#CBC6BE',
                    focusBorderColor: '{ink.900}',
                    color: '#1A1817',
                    placeholderColor: '#958F88',
                    floatLabelColor: '#645E58'
                },
                text: {
                    color: '#1A1817',
                    hoverColor: '#0E0D0C',
                    mutedColor: '#645E58',
                    hoverMutedColor: '#4F4A45'
                },
                content: {
                    background: '#FFFFFF',
                    hoverBackground: '#F3F0EB',
                    borderColor: '#E6E2DC',
                    color: '#1A1817'
                },
                overlay: {
                    modal: { background: '#FFFFFF', borderColor: '#E6E2DC', color: '#1A1817' },
                    popover: { background: '#FFFFFF', borderColor: '#E6E2DC', color: '#1A1817' }
                }
            },
            dark: {
                /* NB: --p-surface-900 is deliberately #171514 (our panel step, --pc-surface-850).
                   The true near-black rail value #121110 is available only via --pc-overlay. */
                surface: {
                    0: '#FFFFFF',
                    50: '#F5F3EF',
                    100: '#EEEBE6',
                    200: '#E2DED8',
                    300: '#C6C1B9',
                    400: '#948E86',
                    500: '#716B64',
                    600: '#524D47',
                    700: '#34302D',
                    800: '#211F1D',
                    900: '#171514',
                    950: '#0E0D0C'
                },
                primary: {
                    color: '{ink.100}',
                    contrastColor: '#141210',
                    hoverColor: '#FFFFFF',
                    activeColor: '{ink.200}'
                },
                highlight: {
                    background: 'rgba(242,238,232,.10)',
                    focusBackground: 'rgba(242,238,232,.16)',
                    color: 'rgba(255,255,255,.94)',
                    focusColor: 'rgba(255,255,255,.94)'
                },
                formField: {
                    background: '{surface.800}',
                    disabledBackground: '{surface.800}',
                    borderColor: '#292624',
                    hoverBorderColor: '{surface.600}',
                    focusBorderColor: '{ink.300}',
                    color: '#F2EEE8',
                    placeholderColor: '{surface.500}',
                    floatLabelColor: '{surface.400}'
                },
                text: {
                    color: '#F2EEE8',
                    hoverColor: '#FFFFFF',
                    mutedColor: '#948E86',
                    hoverMutedColor: '#B3ADA5'
                },
                content: {
                    background: '{surface.900}',
                    hoverBackground: '{surface.800}',
                    borderColor: '#292624',
                    color: '#F2EEE8'
                },
                overlay: {
                    modal: { background: '#121110', borderColor: '#292624', color: '#F2EEE8' },
                    popover: { background: '#121110', borderColor: '#292624', color: '#F2EEE8' }
                }
            }
        }
    },
    components: {
        button: { root: { borderRadius: '999px', paddingX: '1rem', paddingY: '0.4375rem' } },
        card: { root: { borderRadius: '14px', background: '{content.background}' } },
        dialog: { root: { borderRadius: '14px' } },
        inputtext: { root: { borderRadius: '8px', paddingY: '0.4375rem' } },
        inputnumber: { root: { borderRadius: '8px' } },
        select: { root: { borderRadius: '8px' } },
        toggleswitch: {
            root: { width: '2.25rem', height: '1.25rem' },
            handle: { size: '0.875rem' }
        },
        toast: { root: { borderRadius: '10px' } }
    }
});

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(i18n);
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
