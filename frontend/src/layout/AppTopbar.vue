<script setup>
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Popover from 'primevue/popover';
import BrandMark from '@/components/BrandMark.vue';
import { useLayout } from '@/layout/composables/layout';
import { usePresenceStatus } from '@/composables/usePresenceStatus';
import { usePlaybackStore } from '@/stores/playback';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePresenceStore } from '@/stores/presence';

const { toggleDarkMode, isDarkTheme } = useLayout();
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
    if (isErrored.value) return 'Open the dashboard';
    return presenceStore.paused ? 'Resume presence' : 'Pause presence';
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
    if (!discordStore.connected) return 'Inactive';
    if (presenceStore.paused) return 'Hidden (paused)';
    return playbackStore.hasActiveSession ? 'Active' : 'Inactive';
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
</script>

<template>
    <header class="layout-topbar">
        <!-- Left: brand lockup -->
        <router-link to="/" class="topbar-brand" aria-label="PlexCord dashboard">
            <BrandMark />
        </router-link>

        <!-- Center: the signal path -->
        <nav class="signal-path" aria-label="Signal path">
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

        <!-- Right: global actions -->
        <div class="topbar-actions">
            <button
                type="button"
                class="topbar-action"
                v-tooltip.bottom="shortcutTooltip(presenceStore.paused ? 'Resume presence' : 'Pause presence', 'P')"
                :aria-label="presenceStore.paused ? 'Resume presence' : 'Pause presence'"
                :aria-pressed="presenceStore.paused"
                @click="presenceStore.toggle()"
            >
                <i :class="presenceStore.paused ? 'pi pi-play' : 'pi pi-pause'"></i>
            </button>
            <button type="button" class="topbar-action" v-tooltip.bottom="shortcutTooltip(isSettings ? 'Back to dashboard' : 'Settings', ',')" :aria-label="isSettings ? 'Back to dashboard' : 'Settings'" @click="toggleSettings">
                <i :class="isSettings ? 'pi pi-arrow-left' : 'pi pi-cog'"></i>
            </button>
            <button type="button" class="topbar-action" v-tooltip.bottom="{ value: isDarkTheme ? 'Light theme' : 'Dark theme', showDelay: 300 }" :aria-label="isDarkTheme ? 'Switch to light theme' : 'Switch to dark theme'" @click="toggleDarkMode">
                <i :class="isDarkTheme ? 'pi pi-sun' : 'pi pi-moon'"></i>
            </button>
        </div>

        <!-- Plex node popover -->
        <Popover ref="plexPopover">
            <div class="node-popover">
                <span class="pc-eyebrow">Plex</span>
                <div class="pop-row">
                    <span class="pop-label">Server</span>
                    <span class="pc-chip-mono">{{ plexHost }}</span>
                </div>
                <div class="pop-row">
                    <span class="pop-label">Account</span>
                    <span class="pop-value">{{ plexStore.userName || '—' }}</span>
                </div>
                <div class="pop-row">
                    <span class="pop-label">Last sync</span>
                    <span class="pop-value">{{ plexStore.lastConnectedRelative }}</span>
                </div>
                <div class="pop-actions">
                    <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="reconnectPlex">Reconnect</button>
                </div>
            </div>
        </Popover>

        <!-- Discord node popover -->
        <Popover ref="discordPopover">
            <div class="node-popover">
                <span class="pc-eyebrow">Discord</span>
                <div class="pop-row">
                    <span class="pop-label">Presence</span>
                    <span class="pop-value">{{ discordPresenceState }}</span>
                </div>
                <div class="pop-row">
                    <span class="pop-label">Last sync</span>
                    <span class="pop-value">{{ discordStore.lastConnectedRelative }}</span>
                </div>
                <div class="pop-actions">
                    <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="reconnectDiscord">Reconnect</button>
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
    padding: 0 var(--pc-page-gutter);
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

/* ---- Right actions (32px ghost icon buttons) ---- */
.topbar-actions {
    margin-left: auto;
    display: flex;
    gap: 8px;
    flex: none;
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
