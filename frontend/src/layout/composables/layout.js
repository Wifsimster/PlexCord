import { computed, reactive } from 'vue';

const THEME_STORAGE_KEY = 'plexcord-theme';

// main.js applies the persisted theme to <html> before mount; this state
// mirrors it so components can react. Default is dark.
const layoutConfig = reactive({
    darkTheme: localStorage.getItem(THEME_STORAGE_KEY) !== 'light'
});

export function useLayout() {
    const executeDarkModeToggle = () => {
        layoutConfig.darkTheme = !layoutConfig.darkTheme;
        document.documentElement.classList.toggle('dark', layoutConfig.darkTheme);
        localStorage.setItem(THEME_STORAGE_KEY, layoutConfig.darkTheme ? 'dark' : 'light');
    };

    const toggleDarkMode = () => {
        const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

        if (!document.startViewTransition || prefersReducedMotion) {
            executeDarkModeToggle();

            return;
        }

        document.startViewTransition(() => executeDarkModeToggle());
    };

    const isDarkTheme = computed(() => layoutConfig.darkTheme);

    return {
        isDarkTheme,
        toggleDarkMode
    };
}
