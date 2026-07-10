<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Popover from 'primevue/popover';
import BrandMark from '@/components/BrandMark.vue';
import { useLayout } from '@/layout/composables/layout';
import { usePresenceStatus } from '@/composables/usePresenceStatus';
import { usePlaybackStore } from '@/stores/playback';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePresenceStore } from '@/stores/presence';
import { WindowMinimise, WindowToggleMaximise, WindowIsMaximised, Quit } from '../../wailsjs/runtime/runtime';

const { toggleDarkMode, isDarkTheme } = useLayout();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const playbackStore = usePlaybackStore();
const plexStore = usePlexConnectionStore();
const discordStore = useDiscordConnectionStore();
const presenceStore = usePresenceStore();
const { status, headline, severity, isErrored, relayHealthy } = usePresenceStatus();

// ---- Right-side actions -------------------------------------------------
const isSettings = computed(() => route.path.startsWith('/settings'));

const toggleSettings = () => {
    router.push(isSettings.value ? '/' : '/settings');
};

const modKey = /mac/i.test(navigator.platform || navigator.userAgent) ? '⌘' : 'Ctrl';
const shortcutTooltip = (label, keys) => ({
    value: `${label} <span class="pc-chip-mono">${modKey}+${keys}</span>`,
    escape: false,
    showDelay: 300
});

// ---- Node dot states -----------------------------------------------------
const plexDotClass = computed(() => {
    if (plexStore.hasError) return 'pc-dot--danger';
    if (plexStore.isRetrying) return 'pc-dot--warn pc-dot--blink';
    if (plexStore.connected) return playbackStore.isPlaying && !presenceStore.paused ? 'pc-dot--success pc-dot--pulse' : 'pc-dot--success';
    return 'pc-dot--idle';
});

const discordDotClass = computed(() => {
    if (discordStore.hasError) return 'pc-dot--danger';
    if (discordStore.isRetrying) return 'pc-dot--warn pc-dot--blink';
    if (discordStore.connected) return playbackStore.isPlaying && !presenceStore.paused ? 'pc-dot--success pc-dot--pulse' : 'pc-dot--success';
    return 'pc-dot--idle';
});

// Connectors tint success when both adjacent ends are healthy (§5.1).
const plexOk = computed(() => plexStore.connected && !plexStore.hasError);
const discordOk = computed(() => discordStore.connected && !discordStore.hasError);
const centerOk = computed(() => status.value === 'live' || status.value === 'idle' || status.value === 'track-paused');
const leftConnectorOk = computed(() => plexOk.value && centerOk.value);
const rightConnectorOk = computed(() => centerOk.value && discordOk.value);

// ---- Headline node (the Afterglow graft) ---------------------------------
const headlineParts = computed(() => {
    const value = headline.value;
    const sep = ' — ';
    const idx = value.indexOf(sep);
    if (status.value === 'live' && idx !== -1) {
        return { state: value.slice(0, idx + sep.length), title: value.slice(idx + sep.length) };
    }
    return { state: value, title: '' };
});

const headlineGlyph = computed(() => {
    switch (status.value) {
        case 'paused':
        case 'track-paused':
            return 'pi pi-pause';
        case 'plex-error':
        case 'discord-error':
            return 'pi pi-exclamation-triangle';
        case 'idle':
            return 'pi pi-minus';
        default:
            return '';
    }
});

const headlineTitle = computed(() => {
    if (isErrored.value) return t('topbar.openDashboard');
    return presenceStore.paused ? t('topbar.resumePresence') : t('topbar.pausePresence');
});

const onHeadlineClick = () => {
    if (isErrored.value) {
        router.push('/');
        return;
    }
    presenceStore.toggle();
};

// ---- Popovers -------------------------------------------------------------
const plexPopover = ref(null);
const discordPopover = ref(null);

const togglePlexPopover = (event) => {
    discordPopover.value?.hide();
    plexPopover.value?.toggle(event);
};
const toggleDiscordPopover = (event) => {
    plexPopover.value?.hide();
    discordPopover.value?.toggle(event);
};

const plexHost = computed(() => {
    if (!plexStore.serverUrl) return '—';
    return plexStore.serverUrl.replace(/^[a-z]+:\/\//i, '').replace(/\/+$/, '');
});

const discordPresenceState = computed(() => {
    if (!discordStore.connected) return t('presenceState.inactive');
    if (presenceStore.paused) return t('presenceState.hiddenPaused');
    return playbackStore.hasActiveSession ? t('presenceState.active') : t('presenceState.inactive');
});

const reconnectPlex = async () => {
    plexPopover.value?.hide();
    try {
        await plexStore.retry();
        await plexStore.refreshStatus();
    } catch (error) {
        console.error('Plex reconnect failed:', error);
    }
};

const reconnectDiscord = async () => {
    discordPopover.value?.hide();
    try {
        await discordStore.retry();
        await discordStore.refreshStatus();
    } catch (error) {
        console.error('Discord reconnect failed:', error);
    }
};

// ---- Window controls (frameless title bar, spec §5.1) --------------------
const isMaximised = ref(false);

const syncMaximised = async () => {
    try {
        isMaximised.value = await WindowIsMaximised();
    } catch {
        // Runtime unavailable (e.g. browser preview) — leave as-is.
    }
};

const minimiseWindow = () => WindowMinimise();
const toggleMaximise = () => {
    WindowToggleMaximise();
    // The runtime has no maximise event; re-read shortly after the toggle.
    setTimeout(syncMaximised, 60);
};
// Mirrors the native close button: Quit() runs the OnBeforeClose hook, which
// hides to the background or quits depending on the "Minimize to tray" setting.
const closeWindow = () => Quit();

onMounted(() => {
    syncMaximised();
    window.addEventListener('resize', syncMaximised);
});
onBeforeUnmount(() => {
    window.removeEventListener('resize', syncMaximised);
});
</script>

<template>
    <header class="layout-topbar" style="--wails-draggable: drag">
        <!-- Left: brand lockup -->
        <router-link to="/" class="topbar-brand" :aria-label="$t('topbar.dashboardAria')" style="--wails-draggable: no-drag">
            <BrandMark />
        </router-link>

        <!-- Center: the signal path -->
        <nav class="signal-path" :aria-label="$t('topbar.signalPathAria')" style="--wails-draggable: no-drag">
            <button type="button" class="signal-node" aria-haspopup="dialog" @click="togglePlexPopover">
                <span class="pc-dot" :class="plexDotClass" aria-hidden="true"></span>
                <span class="signal-node-name">Plex</span>
            </button>

            <span class="signal-connector" :class="{ 'signal-connector--ok': leftConnectorOk }" aria-hidden="true"></span>

            <Transition name="pc-state" mode="out-in">
                <button type="button" :key="status + headlineParts.title" class="signal-node signal-node--headline" :class="`signal-headline--${severity}`" :title="headlineTitle" @click="onHeadlineClick">
                    <span v-if="status === 'live'" class="pc-eq" aria-hidden="true"><i></i><i></i><i></i></span>
                    <i v-else-if="headlineGlyph" :class="headlineGlyph" class="signal-headline-glyph" aria-hidden="true"></i>
                    <span class="signal-headline-state">{{ headlineParts.state }}</span>
                    <span v-if="headlineParts.title" class="signal-headline-title">{{ headlineParts.title }}</span>
                </button>
            </Transition>

            <span class="signal-connector" :class="{ 'signal-connector--ok': rightConnectorOk }" aria-hidden="true"></span>

            <button type="button" class="signal-node" aria-haspopup="dialog" @click="toggleDiscordPopover">
                <span class="pc-dot" :class="discordDotClass" aria-hidden="true"></span>
                <span class="signal-node-name">Discord</span>
            </button>
        </nav>

        <!-- Right: global actions + window controls -->
        <div class="topbar-right">
            <div class="topbar-actions" style="--wails-draggable: no-drag">
                <button
                    type="button"
                    class="topbar-action"
                    v-tooltip.bottom="shortcutTooltip(presenceStore.paused ? $t('topbar.resumePresence') : $t('topbar.pausePresence'), 'P')"
                    :aria-label="presenceStore.paused ? $t('topbar.resumePresence') : $t('topbar.pausePresence')"
                    :aria-pressed="presenceStore.paused"
                    @click="presenceStore.toggle()"
                >
                    <i :class="presenceStore.paused ? 'pi pi-play' : 'pi pi-pause'"></i>
                </button>
                <button type="button" class="topbar-action" v-tooltip.bottom="shortcutTooltip(isSettings ? $t('topbar.backToDashboard') : $t('topbar.settings'), ',')" :aria-label="isSettings ? $t('topbar.backToDashboard') : $t('topbar.settings')" @click="toggleSettings">
                    <i :class="isSettings ? 'pi pi-arrow-left' : 'pi pi-cog'"></i>
                </button>
                <button type="button" class="topbar-action" v-tooltip.bottom="{ value: isDarkTheme ? $t('topbar.lightTheme') : $t('topbar.darkTheme'), showDelay: 300 }" :aria-label="isDarkTheme ? $t('topbar.switchToLight') : $t('topbar.switchToDark')" @click="toggleDarkMode">
                    <i :class="isDarkTheme ? 'pi pi-sun' : 'pi pi-moon'"></i>
                </button>
            </div>

            <!-- Window caption buttons (frameless title bar) -->
            <div class="window-controls" style="--wails-draggable: no-drag">
                <button type="button" class="caption-btn" :aria-label="$t('topbar.minimise')" @click="minimiseWindow">
                    <svg width="10" height="10" viewBox="0 0 10 10" aria-hidden="true"><rect x="0" y="4.5" width="10" height="1" fill="currentColor" /></svg>
                </button>
                <button type="button" class="caption-btn" :aria-label="isMaximised ? $t('topbar.restore') : $t('topbar.maximise')" @click="toggleMaximise">
                    <svg v-if="isMaximised" width="10" height="10" viewBox="0 0 10 10" aria-hidden="true">
                        <rect x="0.5" y="2.5" width="6" height="6" fill="none" stroke="currentColor" stroke-width="1" />
                        <path d="M2.5 2.5 V0.5 H9.5 V7.5 H7.5" fill="none" stroke="currentColor" stroke-width="1" />
                    </svg>
                    <svg v-else width="10" height="10" viewBox="0 0 10 10" aria-hidden="true"><rect x="0.5" y="0.5" width="9" height="9" fill="none" stroke="currentColor" stroke-width="1" /></svg>
                </button>
                <button type="button" class="caption-btn caption-btn--close" :aria-label="$t('topbar.close')" @click="closeWindow">
                    <svg width="10" height="10" viewBox="0 0 10 10" aria-hidden="true"><path d="M0.5 0.5 L9.5 9.5 M9.5 0.5 L0.5 9.5" stroke="currentColor" stroke-width="1" /></svg>
                </button>
            </div>
        </div>

        <!-- Plex node popover -->
        <Popover ref="plexPopover">
            <div class="node-popover">
                <span class="pc-eyebrow">{{ $t('topbar.plex') }}</span>
                <div class="pop-row">
                    <span class="pop-label">{{ $t('topbar.server') }}</span>
                    <span class="pc-chip-mono">{{ plexHost }}</span>
                </div>
                <div class="pop-row">
                    <span class="pop-label">{{ $t('topbar.account') }}</span>
                    <span class="pop-value">{{ plexStore.userName || $t('common.none') }}</span>
                </div>
                <div class="pop-row">
                    <span class="pop-label">{{ $t('topbar.lastSync') }}</span>
                    <span class="pop-value">{{ plexStore.lastConnectedRelative }}</span>
                </div>
                <div class="pop-actions">
                    <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="reconnectPlex">{{ $t('common.reconnect') }}</button>
                </div>
            </div>
        </Popover>

        <!-- Discord node popover -->
        <Popover ref="discordPopover">
            <div class="node-popover">
                <span class="pc-eyebrow">{{ $t('topbar.discord') }}</span>
                <div class="pop-row">
                    <span class="pop-label">{{ $t('topbar.presence') }}</span>
                    <span class="pop-value">{{ discordPresenceState }}</span>
                </div>
                <div class="pop-row">
                    <span class="pop-label">{{ $t('topbar.lastSync') }}</span>
                    <span class="pop-value">{{ discordStore.lastConnectedRelative }}</span>
                </div>
                <div class="pop-actions">
                    <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="reconnectDiscord">{{ $t('common.reconnect') }}</button>
                </div>
            </div>
        </Popover>
    </header>
</template>

<style scoped>
/* The signal strip (spec §5.1): fixed 48px, overlay surface, hairline bottom. */
.layout-topbar {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 48px;
    z-index: 997;
    display: flex;
    align-items: center;
    /* No right gutter: window caption buttons sit flush in the corner. */
    padding: 0 0 0 var(--pc-page-gutter);
    background: var(--pc-overlay);
    border-bottom: 1px solid var(--pc-border);
}

.topbar-brand {
    display: inline-flex;
    align-items: center;
    height: 32px;
    padding: 0 6px;
    margin-left: -6px;
    border-radius: var(--pc-radius-sm);
    flex: none;
}

/* ---- Center signal path ---- */
.signal-path {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    align-items: center;
    max-width: min(60vw, 640px);
}

.signal-node {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 28px;
    padding: 0 8px;
    border: none;
    border-radius: var(--pc-radius-sm);
    background: transparent;
    color: var(--pc-text-secondary);
    font-family: var(--pc-font-ui);
    cursor: pointer;
    transition:
        background-color var(--pc-dur-1) var(--pc-ease-out),
        color var(--pc-dur-1) var(--pc-ease-out);
}
.signal-node:hover {
    background: var(--pc-raised);
    color: var(--pc-text);
}
.signal-node .pc-dot {
    transition: color var(--pc-dur-2) var(--pc-ease-out); /* M6 dot color crossfade */
}
.signal-node-name {
    font-size: 12.5px;
    line-height: 1;
}

.signal-connector {
    width: 16px;
    height: 1px;
    flex: none;
    background: var(--pc-border);
    transition: background-color var(--pc-dur-2) var(--pc-ease-out);
}
.signal-connector--ok {
    background: var(--pc-success);
}

/* ---- Headline node ---- */
.signal-node--headline {
    min-width: 0;
}
.signal-headline-glyph {
    font-size: 10px;
    flex: none;
}
.signal-headline-state {
    font-size: var(--pc-text-caption);
    font-weight: 500;
    white-space: nowrap;
}
.signal-headline--success .signal-headline-state,
.signal-headline--success .signal-headline-glyph {
    color: var(--pc-success);
}
.signal-headline--warn .signal-headline-state,
.signal-headline--warn .signal-headline-glyph {
    color: var(--pc-warn);
}
.signal-headline--danger .signal-headline-state,
.signal-headline--danger .signal-headline-glyph {
    color: var(--pc-danger);
}
.signal-headline--muted .signal-headline-state,
.signal-headline--muted .signal-headline-glyph {
    color: var(--pc-text-muted);
}
.signal-headline-title {
    font-size: var(--pc-text-body);
    font-weight: 500;
    color: var(--pc-text);
    max-width: 280px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

/* ---- Right cluster: app actions + window caption buttons ---- */
.topbar-right {
    margin-left: auto;
    display: flex;
    align-items: center;
    align-self: stretch;
    flex: none;
}

/* ---- App actions (32px ghost icon buttons) ---- */
.topbar-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: none;
    /* Divider between app actions and the window caption buttons. */
    padding: 0 var(--pc-page-gutter) 0 0;
    margin-right: 4px;
    border-right: 1px solid var(--pc-border);
    height: 24px;
}
.topbar-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border: none;
    border-radius: var(--pc-radius-sm);
    background: transparent;
    color: var(--pc-text-secondary);
    cursor: pointer;
    transition:
        background-color var(--pc-dur-1) var(--pc-ease-out),
        color var(--pc-dur-1) var(--pc-ease-out);
}
.topbar-action:hover {
    background: var(--pc-raised);
    color: var(--pc-text);
}
.topbar-action i {
    font-size: 14px;
}

/* ---- Window caption buttons (full-height, flush to the corner) ---- */
.window-controls {
    display: flex;
    align-self: stretch;
    flex: none;
}
.caption-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 46px;
    height: 100%;
    border: none;
    background: transparent;
    color: var(--pc-text-secondary);
    cursor: pointer;
    transition:
        background-color var(--pc-dur-1) var(--pc-ease-out),
        color var(--pc-dur-1) var(--pc-ease-out);
}
.caption-btn:hover {
    background: var(--pc-raised);
    color: var(--pc-text);
}
.caption-btn:active {
    background: var(--pc-surface-700);
}
.caption-btn--close:hover {
    background: var(--pc-danger);
    color: #ffffff;
}
.caption-btn--close:active {
    background: var(--pc-danger);
    filter: brightness(0.9);
}

/* ---- Node popovers (overlay recipe; content teleports with scoped attrs) ---- */
.node-popover {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 220px;
    padding: 4px;
}
.pop-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-height: 24px;
}
.pop-label {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.pop-value {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
}
.pop-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 4px;
    padding-top: 8px;
    border-top: 1px solid var(--pc-border-subtle);
}

/* Narrow windows: keep the strip usable — hide the path labels first. */
@media (max-width: 720px) {
    .signal-headline-title {
        max-width: 140px;
    }
}
@media (max-width: 560px) {
    .signal-node-name {
        display: none;
    }
}
</style>
