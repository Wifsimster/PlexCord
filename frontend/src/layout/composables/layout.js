import { computed, reactive } from 'vue';

const layoutConfig = reactive({
    preset: 'Aura',
    primary: 'purple',
    surface: null,
    darkTheme: true,
    menuMode: 'static'
});

const layoutState = reactive({
    activeMenuItem: null
});

export function useLayout() {
    const setActiveMenuItem = (item) => {
        layoutState.activeMenuItem = item.value || item;
    };

    const toggleDarkMode = () => {
        if (!document.startViewTransition) {
            executeDarkModeToggle();

            return;
        }

        document.startViewTransition(() => executeDarkModeToggle(event));
    };

    const executeDarkModeToggle = () => {
        layoutConfig.darkTheme = !layoutConfig.darkTheme;
        document.documentElement.classList.toggle('app-dark');
    };

    const toggleMenu = () => {
        // Menu toggle functionality removed
    };

    const isDarkTheme = computed(() => layoutConfig.darkTheme);

    const getPrimary = computed(() => layoutConfig.primary);

    const getSurface = computed(() => layoutConfig.surface);

    return {
        layoutConfig,
        layoutState,
        toggleMenu,
        isDarkTheme,
        getPrimary,
        getSurface,
        setActiveMenuItem,
        toggleDarkMode
    };
}
