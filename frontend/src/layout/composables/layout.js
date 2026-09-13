import { computed, reactive } from 'vue';

const THEME_STORAGE_KEY = 'plexcord-theme';

// main.js applies the persisted theme to <html> before mount; this state
// mirrors it so components can react. Default is dark.
const layoutConfig = reactive({
    darkTheme: localStorage.getItem(THEME_STORAGE_KEY) !== 'light'
});

export function useLayout() {
    const applyDarkMode = (dark) => {
        layoutConfig.darkTheme = dark;
        document.documentElement.classList.toggle('dark', dark);
        localStorage.setItem(THEME_STORAGE_KEY, dark ? 'dark' : 'light');
    };

    // Wraps the swap in a view transition (M-transition path) unless the
    // platform lacks the API or the user asked for reduced motion.
    const transition = (apply) => {
        const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

        if (!document.startViewTransition || prefersReducedMotion) {
            apply();

            return;
        }

        document.startViewTransition(() => apply());
    };

    const setDarkMode = (dark) => {
        if (dark === layoutConfig.darkTheme) return;
        transition(() => applyDarkMode(dark));
    };

    const toggleDarkMode = () => {
        transition(() => applyDarkMode(!layoutConfig.darkTheme));
    };

    const isDarkTheme = computed(() => layoutConfig.darkTheme);

    return {
        isDarkTheme,
        setDarkMode,
        toggleDarkMode
    };
}
