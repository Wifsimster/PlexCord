<script setup>
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useToast } from 'primevue/usetoast';
import { useConfirm } from 'primevue/useconfirm';
import InputText from 'primevue/inputtext';
import InputNumber from 'primevue/inputnumber';
import ToggleSwitch from 'primevue/toggleswitch';
import Dialog from 'primevue/dialog';
import DiscordSpecimen from '@/components/DiscordSpecimen.vue';
import SavedIndicator from '@/components/settings/SavedIndicator.vue';
import { useSetupStore } from '@/stores/setup';
import { usePresenceStore } from '@/stores/presence';
import { usePlayback } from '@/composables/usePlayback';
import { useVersion } from '@/composables/useVersion';
import { validatePlexServerUrl, PLEX_URL_PLACEHOLDER } from '@/utils/plexUrl';
import { setLocale, SUPPORTED_LOCALES } from '@/i18n';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import {
    GetPollingInterval,
    SetPollingInterval,
    GetAutoStart,
    SetAutoStart,
    GetMinimizeToTray,
    SetMinimizeToTray,
    GetDiscordClientID,
    GetDefaultDiscordClientID,
    SaveDiscordClientID,
    ValidateDiscordClientID,
    ConnectDiscord,
    DisconnectDiscord,
    TestDiscordPresence,
    CheckForUpdate,
    CanSelfUpdate,
    DownloadAndInstallUpdate,
    RestartApplication,
    OpenReleasesPage,
    OpenReleaseURL,
    ResetApplication,
    GetHideWhenPaused,
    SetHideWhenPaused,
    GetPresenceFormat,
    SetPresenceFormat,
    GetServers,
    AddServer,
    RemoveServer,
    SetServerActive,
    ValidatePlexConnection,
    GetPlexToken,
    DiscoverPlexServers
} from '../../../wailsjs/go/main/App';

const router = useRouter();
const { t, locale } = useI18n();
const toast = useToast();
const confirm = useConfirm();
const setupStore = useSetupStore();
const presenceStore = usePresenceStore();
const { currentTrack, hasActiveSession } = usePlayback();
const { version, commit, buildDate } = useVersion();

const AUTOSAVE_DEBOUNCE = 600;
const FORMAT_TOKENS = ['{track}', '{artist}', '{album}', '{year}', '{player}'];
const SAMPLE_TRACK = Object.freeze({
    sessionKey: 'sample-queen',
    track: 'Bohemian Rhapsody',
    artist: 'Queen',
    album: 'A Night at the Opera',
    year: 1975,
    playerName: 'Plexamp',
    duration: 354000,
    viewOffset: 83000,
    state: 'playing'
});

// ---------------- Settings state ----------------
const loaded = ref(false);
const pollingInterval = ref(2);
const autoStart = ref(false);
const minimizeToTray = ref(true);
const hideWhenPaused = ref(false);
const hideWhenPausedDelay = ref(0);
const detailsFormat = ref('');
const stateFormat = ref('');
const discordClientId = ref('');
const defaultClientId = ref('');
const servers = ref([]);
const hasPlexToken = ref(false);

// Per-server health from the lightweight validate ping (keyed by URL):
// { status: 'unknown'|'testing'|'ok'|'fail'|'auth', message }
const serverHealth = reactive({});

// Last values acknowledged by the backend (autosave dedupe)
const lastSaved = { polling: null, details: null, state: null, hideDelay: null };

// ---------------- Inline "✓ Saved" flashes (M15) ----------------
const savedFlags = reactive({});
const savedTimers = {};
function flashSaved(key) {
    savedFlags[key] = true;
    clearTimeout(savedTimers[key]);
    savedTimers[key] = setTimeout(() => {
        savedFlags[key] = false;
    }, 1600);
}

function friendlyError(error, fallback) {
    if (typeof error === 'string' && error) return error;
    return error?.message || fallback;
}

function toastFailure(summary, error, fallback) {
    toast.add({ severity: 'error', summary, detail: friendlyError(error, fallback), life: 8000 });
}

// ---------------- Rail: scroll-spy + keyboard listbox ----------------
const sections = [
    { id: 'section-connection', labelKey: 'settings.sectionConnection' },
    { id: 'section-presence', labelKey: 'settings.sectionPresence' },
    { id: 'section-app', labelKey: 'settings.sectionApp' },
    { id: 'section-advanced', labelKey: 'settings.sectionAdvanced' },
    { id: 'section-about', labelKey: 'settings.sectionAbout' }
];

// ---------------- Language (spec: display language override) ----------------
// The active locale is resolved from the environment at startup; this picker
// lets the user pin a specific language, persisted via setLocale().
const languageOptions = SUPPORTED_LOCALES.map((code) => ({ code, label: t(`settings.languageNames.${code}`) }));
const selectedLanguage = computed(() => locale.value);
function changeLanguage(code) {
    setLocale(code);
}
const activeSectionId = ref(sections[0].id);
const railFocusIndex = ref(0);
const activeIndex = computed(() =>
    Math.max(
        0,
        sections.findIndex((s) => s.id === activeSectionId.value)
    )
);
const visibleSections = new Set();
let sectionObserver = null;
let suppressSpyUntil = 0;

function scrollToSection(id) {
    activeSectionId.value = id;
    railFocusIndex.value = sections.findIndex((s) => s.id === id);
    suppressSpyUntil = Date.now() + 800;
    const el = document.getElementById(id);
    if (el) {
        const reduce = window.matchMedia?.('(prefers-reduced-motion: reduce)')?.matches;
        el.scrollIntoView({ behavior: reduce ? 'auto' : 'smooth', block: 'start' });
    }
}

function onRailKeydown(event) {
    const max = sections.length - 1;
    switch (event.key) {
        case 'ArrowDown':
            event.preventDefault();
            railFocusIndex.value = Math.min(max, railFocusIndex.value + 1);
            break;
        case 'ArrowUp':
            event.preventDefault();
            railFocusIndex.value = Math.max(0, railFocusIndex.value - 1);
            break;
        case 'Home':
            event.preventDefault();
            railFocusIndex.value = 0;
            break;
        case 'End':
            event.preventDefault();
            railFocusIndex.value = max;
            break;
        case 'Enter':
        case ' ':
            event.preventDefault();
            scrollToSection(sections[railFocusIndex.value].id);
            break;
    }
}

function onRailFocus() {
    railFocusIndex.value = activeIndex.value;
}

// ---------------- Initial load ----------------
onMounted(async () => {
    try {
        pollingInterval.value = await GetPollingInterval();
        autoStart.value = await GetAutoStart();
        minimizeToTray.value = await GetMinimizeToTray();
        discordClientId.value = await GetDiscordClientID();
        defaultClientId.value = await GetDefaultDiscordClientID();

        const pauseSettings = await GetHideWhenPaused();
        hideWhenPaused.value = !!pauseSettings?.enabled;
        hideWhenPausedDelay.value = pauseSettings?.delaySeconds ?? 0;

        const formatSettings = await GetPresenceFormat();
        detailsFormat.value = formatSettings?.detailsFormat ?? '';
        stateFormat.value = formatSettings?.stateFormat ?? '';

        servers.value = await GetServers();

        lastSaved.polling = pollingInterval.value;
        lastSaved.details = detailsFormat.value;
        lastSaved.state = stateFormat.value;
        lastSaved.hideDelay = hideWhenPausedDelay.value;

        try {
            hasPlexToken.value = !!(await GetPlexToken());
        } catch {
            hasPlexToken.value = false;
        }
        try {
            canSelfUpdate.value = !!(await CanSelfUpdate());
        } catch {
            canSelfUpdate.value = false;
        }
        refreshAllServerHealth();
    } catch (error) {
        toastFailure(t('settings.toast.loadFailed'), error, t('settings.toast.loadFailedDetail'));
    } finally {
        loaded.value = true;
    }

    sectionObserver = new IntersectionObserver(
        (entries) => {
            for (const entry of entries) {
                if (entry.isIntersecting) visibleSections.add(entry.target.id);
                else visibleSections.delete(entry.target.id);
            }
            if (Date.now() < suppressSpyUntil) return;
            const first = sections.find((s) => visibleSections.has(s.id));
            if (first) activeSectionId.value = first.id;
        },
        { rootMargin: '-72px 0px -55% 0px', threshold: 0 }
    );
    sections.forEach((s) => {
        const el = document.getElementById(s.id);
        if (el) sectionObserver.observe(el);
    });

    // In-app update lifecycle: progress, completion, and failure all arrive
    // as runtime events emitted by the Go updater.
    EventsOn('UpdateDownloadProgress', (p) => {
        updateProgress.value = Math.round(p?.percent ?? 0);
    });
    EventsOn('UpdateReady', () => {
        updateProgress.value = 100;
        installingUpdate.value = false;
        updateReady.value = true;
    });
    EventsOn('UpdateError', (message) => {
        installingUpdate.value = false;
        toastFailure(t('settings.toast.updateFailed'), message, t('settings.toast.updateFailedDetail'));
    });
});

onBeforeUnmount(() => {
    EventsOff('UpdateDownloadProgress');
    EventsOff('UpdateReady');
    EventsOff('UpdateError');
    sectionObserver?.disconnect();
    Object.values(savedTimers).forEach(clearTimeout);
    clearTimeout(pollingTimer);
    clearTimeout(formatTimer);
    clearTimeout(hideDelayTimer);
    clearTimeout(testResultTimer);
});

// ---------------- Connection: polling interval (autosave) ----------------
let pollingTimer = null;
watch(pollingInterval, () => {
    if (!loaded.value) return;
    clearTimeout(pollingTimer);
    pollingTimer = setTimeout(savePollingInterval, AUTOSAVE_DEBOUNCE);
});

function flushPollingSave() {
    clearTimeout(pollingTimer);
    nextTick(savePollingInterval);
}

async function savePollingInterval() {
    const value = pollingInterval.value;
    if (value == null || value < 1 || value > 60 || value === lastSaved.polling) return;
    try {
        await SetPollingInterval(value);
        lastSaved.polling = value;
        flashSaved('polling');
    } catch (error) {
        toastFailure(t('settings.toast.pollingFailed'), error, t('settings.toast.pollingFailedDetail'));
    }
}

// ---------------- Connection: servers ----------------
async function loadServers() {
    try {
        servers.value = await GetServers();
    } catch (error) {
        toastFailure(t('settings.toast.loadServersFailed'), error, t('settings.toast.loadServersFailedDetail'));
    }
}

function refreshAllServerHealth() {
    servers.value.forEach((server) => {
        if (!hasPlexToken.value) {
            serverHealth[server.url] = { status: 'auth', message: '' };
        } else if (!serverHealth[server.url] || serverHealth[server.url].status === 'unknown') {
            testServer(server);
        }
    });
}

async function testServer(server) {
    if (!hasPlexToken.value) {
        serverHealth[server.url] = { status: 'auth', message: '' };
        return;
    }
    serverHealth[server.url] = { status: 'testing', message: '' };
    try {
        await ValidatePlexConnection(server.url);
        serverHealth[server.url] = { status: 'ok', message: '' };
    } catch (error) {
        serverHealth[server.url] = { status: 'fail', message: friendlyError(error, t('settings.toast.serverUnreachable')) };
    }
}

function serverDotClass(server) {
    const status = serverHealth[server.url]?.status ?? 'unknown';
    switch (status) {
        case 'ok':
            return 'pc-dot--success';
        case 'fail':
            return 'pc-dot--danger';
        case 'auth':
            return 'pc-dot--warn';
        case 'testing':
            return 'pc-dot--idle pc-dot--blink';
        default:
            return 'pc-dot--idle';
    }
}

async function toggleServerActive(server) {
    try {
        await SetServerActive(server.url, !server.active);
        await loadServers();
    } catch (error) {
        toastFailure(t('settings.toast.updateServerFailed'), error, t('settings.toast.updateServerFailedDetail'));
    }
}

function confirmRemoveServer(server) {
    const isOnlyActive = server.active && servers.value.filter((s) => s.active).length === 1;
    confirm.require({
        header: t('settings.confirm.removeHeader'),
        message: isOnlyActive ? t('settings.confirm.removeOnlyActive', { name: server.name }) : t('settings.confirm.removeMessage', { name: server.name }),
        acceptProps: { label: t('settings.confirm.remove'), severity: 'danger' },
        rejectProps: { label: t('common.cancel'), severity: 'secondary', text: true },
        accept: () => removeServer(server)
    });
}

async function removeServer(server) {
    try {
        await RemoveServer(server.url);
        delete serverHealth[server.url];
        await loadServers();
    } catch (error) {
        toastFailure(t('settings.toast.removeServerFailed'), error, t('settings.toast.removeServerFailedDetail'));
    }
}

function goToPlexAuth() {
    router.push('/setup/plex');
}

// Add-server dialog
const showAddServerDialog = ref(false);
const newServerName = ref('');
const newServerURL = ref('');
const addServerError = ref('');
const addingServer = ref(false);
const newServerUrlValidation = computed(() => validatePlexServerUrl(newServerURL.value));
const newServerUrlTouched = computed(() => newServerURL.value.trim().length > 0);
const canAddServer = computed(() => newServerName.value.trim().length > 0 && newServerUrlValidation.value.valid);

// Server auto-discovery (GDM) inside the add-server dialog
const isDiscovering = ref(false);
const hasDiscovered = ref(false);
const discoveryError = ref('');
const discoveredServers = ref([]);

function openAddServerDialog() {
    newServerName.value = '';
    newServerURL.value = '';
    addServerError.value = '';
    isDiscovering.value = false;
    hasDiscovered.value = false;
    discoveryError.value = '';
    discoveredServers.value = [];
    showAddServerDialog.value = true;
}

async function discoverServers() {
    isDiscovering.value = true;
    discoveryError.value = '';
    try {
        discoveredServers.value = (await DiscoverPlexServers()) || [];
        hasDiscovered.value = true;
    } catch {
        discoveredServers.value = [];
        hasDiscovered.value = false;
        discoveryError.value = t('settings.discoveryFailed');
    } finally {
        isDiscovering.value = false;
    }
}

const discoveredServerURL = (server) => `http://${server.address}:${server.port}`;

function isServerAlreadyAdded(server) {
    return servers.value.some((s) => s.url === discoveredServerURL(server));
}

function selectDiscoveredServer(server) {
    if (isServerAlreadyAdded(server)) return;
    newServerName.value = server.name || t('settings.defaultServerName');
    newServerURL.value = discoveredServerURL(server);
}

async function addServer() {
    if (!canAddServer.value || addingServer.value) return;
    addingServer.value = true;
    addServerError.value = '';
    const url = newServerURL.value.trim();
    try {
        await AddServer(newServerName.value.trim(), url, '', '');
        showAddServerDialog.value = false;
        await loadServers();
        const added = servers.value.find((s) => s.url === url);
        if (added) testServer(added);
    } catch (error) {
        addServerError.value = friendlyError(error, t('settings.toast.addServerFailedDetail'));
    } finally {
        addingServer.value = false;
    }
}

// ---------------- Presence: format editor (autosave on blur / 600ms) ----------------
const detailsInputRef = ref(null);
const stateInputRef = ref(null);
const lastFocusedFormat = ref('details');
let formatTimer = null;

watch([detailsFormat, stateFormat], () => {
    if (!loaded.value) return;
    clearTimeout(formatTimer);
    formatTimer = setTimeout(savePresenceFormat, AUTOSAVE_DEBOUNCE);
});

function flushFormatSave() {
    clearTimeout(formatTimer);
    nextTick(savePresenceFormat);
}

async function savePresenceFormat() {
    if (detailsFormat.value === lastSaved.details && stateFormat.value === lastSaved.state) return;
    try {
        await SetPresenceFormat(detailsFormat.value, stateFormat.value);
        lastSaved.details = detailsFormat.value;
        lastSaved.state = stateFormat.value;
        flashSaved('format');
    } catch (error) {
        toastFailure(t('settings.toast.formatFailed'), error, t('settings.toast.formatFailedDetail'));
    }
}

function resetFormats() {
    detailsFormat.value = '';
    stateFormat.value = '';
    flushFormatSave();
}

function insertToken(token) {
    const useState = lastFocusedFormat.value === 'state';
    const model = useState ? stateFormat : detailsFormat;
    const el = (useState ? stateInputRef : detailsInputRef).value?.$el;
    const current = model.value ?? '';
    let start = current.length;
    let end = current.length;
    if (el && typeof el.selectionStart === 'number') {
        start = el.selectionStart;
        end = el.selectionEnd ?? el.selectionStart;
    }
    model.value = current.slice(0, start) + token + current.slice(end);
    nextTick(() => {
        if (el) {
            el.focus();
            const caret = start + token.length;
            el.setSelectionRange(caret, caret);
        }
    });
}

// Specimen: live playback when present, else the Queen sample (SAMPLE badge)
const specimenIsSample = computed(() => !hasActiveSession.value);
const specimenTrack = computed(() => (hasActiveSession.value ? currentTrack.value : SAMPLE_TRACK));
const specimenPaused = computed(() => presenceStore.paused && hasActiveSession.value);
const specimenFormats = computed(() => ({ details: detailsFormat.value, state: stateFormat.value }));

// ---------------- Presence: hide when paused ----------------
const hideWhenPausedSaving = ref(false);
let hideDelayTimer = null;

async function updateHideWhenPaused(value) {
    hideWhenPausedSaving.value = true;
    hideWhenPaused.value = value; // optimistic
    try {
        await SetHideWhenPaused(value, hideWhenPausedDelay.value ?? 0);
        flashSaved('hidePaused');
    } catch (error) {
        hideWhenPaused.value = !value; // revert
        toastFailure(t('settings.toast.hidePausedFailed'), error, t('settings.toast.hidePausedFailedDetail'));
    } finally {
        hideWhenPausedSaving.value = false;
    }
}

watch(hideWhenPausedDelay, () => {
    if (!loaded.value) return;
    clearTimeout(hideDelayTimer);
    hideDelayTimer = setTimeout(saveHideWhenPausedDelay, AUTOSAVE_DEBOUNCE);
});

function flushHideDelaySave() {
    clearTimeout(hideDelayTimer);
    nextTick(saveHideWhenPausedDelay);
}

async function saveHideWhenPausedDelay() {
    const delay = hideWhenPausedDelay.value;
    if (delay == null || delay < 0 || delay > 300 || delay === lastSaved.hideDelay) return;
    try {
        await SetHideWhenPaused(hideWhenPaused.value, delay);
        lastSaved.hideDelay = delay;
        flashSaved('hideDelay');
    } catch (error) {
        toastFailure(t('settings.toast.delayFailed'), error, t('settings.toast.delayFailedDetail'));
    }
}

// ---------------- App: toggles (instant, optimistic + revert) ----------------
const autoStartSaving = ref(false);
const minimizeToTraySaving = ref(false);

async function updateAutoStart(value) {
    autoStartSaving.value = true;
    autoStart.value = value;
    try {
        await SetAutoStart(value);
        flashSaved('autoStart');
    } catch (error) {
        autoStart.value = !value;
        toastFailure(t('settings.toast.autoStartFailed'), error, t('settings.toast.settingFailedDetail'));
    } finally {
        autoStartSaving.value = false;
    }
}

async function updateMinimizeToTray(value) {
    minimizeToTraySaving.value = true;
    minimizeToTray.value = value;
    try {
        await SetMinimizeToTray(value);
        flashSaved('minimizeToTray');
    } catch (error) {
        minimizeToTray.value = !value;
        toastFailure(t('settings.toast.trayFailed'), error, t('settings.toast.settingFailedDetail'));
    } finally {
        minimizeToTraySaving.value = false;
    }
}

// ---------------- Advanced: Discord Client ID (explicit Apply) ----------------
const applyingClientId = ref(false);
const clientIdError = ref('');
const clientIdWarning = ref('');
const isUsingDefaultClientId = computed(() => !discordClientId.value || discordClientId.value === defaultClientId.value);

watch(discordClientId, () => {
    clientIdError.value = '';
    clientIdWarning.value = '';
});

async function applyClientId() {
    if (applyingClientId.value) return;
    applyingClientId.value = true;
    clientIdError.value = '';
    clientIdWarning.value = '';
    const id = (discordClientId.value ?? '').trim();
    try {
        if (id && id !== defaultClientId.value) {
            await ValidateDiscordClientID(id);
        }
        await SaveDiscordClientID(id);
        discordClientId.value = id;
        // The caption promises it: applying reconnects Discord with the new app.
        try {
            await DisconnectDiscord();
        } catch {
            // Not connected — nothing to disconnect.
        }
        try {
            await ConnectDiscord(id);
            flashSaved('clientId');
        } catch (error) {
            clientIdWarning.value = friendlyError(error, t('settings.warnReconnect'));
        }
    } catch (error) {
        clientIdError.value = friendlyError(error, t('settings.toast.invalidClientId'));
    } finally {
        applyingClientId.value = false;
    }
}

function resetClientIdToDefault() {
    discordClientId.value = '';
    applyClientId();
}

// ---------------- Advanced: send test presence ----------------
const sendingTestPresence = ref(false);
const testPresenceResult = ref(null);
let testResultTimer = null;

async function sendTestPresence() {
    if (sendingTestPresence.value) return;
    sendingTestPresence.value = true;
    clearTimeout(testResultTimer);
    testPresenceResult.value = null;
    try {
        await TestDiscordPresence();
        testPresenceResult.value = { ok: true, message: t('settings.testSent') };
        testResultTimer = setTimeout(() => {
            testPresenceResult.value = null;
        }, 4000);
    } catch (error) {
        testPresenceResult.value = { ok: false, message: friendlyError(error, t('settings.testNotConnected')) };
        testResultTimer = setTimeout(() => {
            testPresenceResult.value = null;
        }, 8000);
    } finally {
        sendingTestPresence.value = false;
    }
}

// ---------------- About: updates + changelog ----------------
const checkingUpdate = ref(false);
const updateInfo = ref(null);
const truncatedReleaseNotes = computed(() => {
    const notes = updateInfo.value?.releaseNotes ?? '';
    return notes.length > 120 ? `${notes.slice(0, 120)}…` : notes;
});

async function checkForUpdates() {
    if (checkingUpdate.value) return;
    checkingUpdate.value = true;
    try {
        updateInfo.value = await CheckForUpdate();
        if (!updateInfo.value?.available) {
            flashSaved('upToDate');
        }
    } catch (error) {
        toastFailure(t('settings.toast.updateCheckFailed'), error, t('settings.toast.updateCheckFailedDetail'));
    } finally {
        checkingUpdate.value = false;
    }
}

function openUpdatePage() {
    if (updateInfo.value?.releaseUrl) {
        OpenReleaseURL(updateInfo.value.releaseUrl);
    } else {
        OpenReleasesPage();
    }
}

// In-app self-update (platforms where the binary can replace itself)
const canSelfUpdate = ref(false);
const installingUpdate = ref(false);
const updateProgress = ref(0);
const updateReady = ref(false);

// Progress, completion, and failure are surfaced through the
// UpdateDownloadProgress / UpdateReady / UpdateError events subscribed in
// onMounted, so the promise here only resets local state on rejection.
async function installUpdate() {
    if (installingUpdate.value) return;
    installingUpdate.value = true;
    updateProgress.value = 0;
    updateReady.value = false;
    try {
        await DownloadAndInstallUpdate();
    } catch {
        // The UpdateError event handler surfaces the message.
        installingUpdate.value = false;
    }
}

async function restartApp() {
    try {
        await RestartApplication();
    } catch (error) {
        toastFailure(t('settings.toast.restartFailed'), error, t('settings.toast.restartFailedDetail'));
    }
}

function openReleases() {
    OpenReleasesPage();
}

// ---------------- Danger zone: reset application ----------------
const resetting = ref(false);

function confirmReset() {
    confirm.require({
        header: t('settings.confirm.resetHeader'),
        message: t('settings.confirm.resetMessage'),
        acceptProps: { label: t('settings.confirm.resetAccept'), severity: 'danger' },
        rejectProps: { label: t('common.cancel'), severity: 'secondary', text: true },
        accept: executeReset
    });
}

async function executeReset() {
    resetting.value = true;
    try {
        await ResetApplication();
        setupStore.resetWizard();
        toast.add({ severity: 'success', summary: t('settings.toast.resetComplete'), detail: t('settings.toast.resetCompleteDetail'), life: 3000 });
        router.push('/setup/welcome');
    } catch (error) {
        toastFailure(t('settings.toast.resetFailed'), error, t('settings.toast.resetFailedDetail'));
    } finally {
        resetting.value = false;
    }
}
</script>

<template>
    <div class="settings-page">
        <h1 class="settings-title">{{ $t('settings.title') }}</h1>

        <div class="settings-layout">
            <!-- Rail: scroll-spy + keyboard listbox (§5.3) -->
            <nav class="settings-rail" :aria-label="$t('settings.railAria')">
                <p class="pc-eyebrow rail-eyebrow" id="settings-rail-label">{{ $t('settings.railEyebrow') }}</p>
                <div class="rail-items" role="listbox" tabindex="0" aria-labelledby="settings-rail-label" :aria-activedescendant="`rail-opt-${sections[railFocusIndex].id}`" @keydown="onRailKeydown" @focus="onRailFocus">
                    <span class="rail-bar" aria-hidden="true" :style="{ transform: `translateY(${activeIndex * 32 + 6}px)` }"></span>
                    <button
                        v-for="(section, index) in sections"
                        :id="`rail-opt-${section.id}`"
                        :key="section.id"
                        type="button"
                        role="option"
                        tabindex="-1"
                        :aria-selected="activeSectionId === section.id"
                        class="rail-item"
                        :class="{ 'rail-item--active': activeSectionId === section.id, 'rail-item--focused': railFocusIndex === index }"
                        @click="scrollToSection(section.id)"
                    >
                        {{ $t(section.labelKey) }}
                    </button>
                </div>
            </nav>

            <div class="settings-sections">
                <!-- ============ Connection ============ -->
                <section id="section-connection" class="pc-panel settings-section" aria-labelledby="hd-connection">
                    <div class="section-head">
                        <h2 id="hd-connection" class="pc-eyebrow section-eyebrow">{{ $t('settings.sectionConnection') }}</h2>
                        <button v-if="loaded" type="button" class="pc-btn pc-btn--secondary pc-btn--sm" @click="openAddServerDialog"><i class="pi pi-plus" aria-hidden="true"></i>{{ $t('settings.addServer') }}</button>
                    </div>

                    <div v-if="!loaded" class="skeleton-rows" aria-hidden="true">
                        <div class="pc-skeleton" style="width: 72%"></div>
                        <div class="pc-skeleton" style="width: 58%"></div>
                        <div class="pc-skeleton" style="width: 64%"></div>
                    </div>

                    <template v-else>
                        <div class="server-list">
                            <p v-if="servers.length === 0" class="row-caption empty-caption">{{ $t('settings.noServers') }}</p>
                            <div v-for="server in servers" :key="server.url" class="server-row">
                                <span class="pc-dot" :class="serverDotClass(server)" aria-hidden="true"></span>
                                <div class="server-main">
                                    <div class="server-line">
                                        <span class="server-name">{{ server.name }}</span>
                                        <span class="pc-chip-mono server-url">{{ server.url }}</span>
                                    </div>
                                    <p v-if="server.userName" class="row-caption">{{ $t('settings.monitoringUser', { user: server.userName }) }}</p>
                                    <p v-if="serverHealth[server.url]?.status === 'auth'" class="row-caption row-caption--warn">{{ $t('settings.needsSignIn') }} <button type="button" class="pc-link" @click="goToPlexAuth">{{ $t('settings.authenticate') }}</button></p>
                                    <p v-else-if="serverHealth[server.url]?.status === 'fail'" class="row-caption row-caption--danger" role="alert"><i class="pi pi-exclamation-circle" aria-hidden="true"></i> {{ serverHealth[server.url].message }}</p>
                                </div>
                                <div class="server-actions">
                                    <i v-if="serverHealth[server.url]?.status === 'testing'" class="pi pi-spinner pi-spin inline-spinner" aria-hidden="true"></i>
                                    <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" :disabled="serverHealth[server.url]?.status === 'testing'" @click="testServer(server)">{{ $t('settings.test') }}</button>
                                    <ToggleSwitch :modelValue="server.active" :aria-label="$t('settings.serverActiveAria', { name: server.name })" @update:modelValue="toggleServerActive(server)" />
                                    <button type="button" class="pc-btn pc-btn--ghost pc-btn--icon icon-danger" :aria-label="$t('settings.removeServerAria', { name: server.name })" @click="confirmRemoveServer(server)"><i class="pi pi-trash" aria-hidden="true"></i></button>
                                </div>
                            </div>
                        </div>

                        <div class="setting-row polling-row">
                            <div class="row-text">
                                <label class="row-label" for="polling-interval">{{ $t('settings.pollingInterval') }}</label>
                                <p class="row-caption">{{ $t('settings.pollingCaption') }}</p>
                            </div>
                            <div class="row-control">
                                <SavedIndicator :visible="!!savedFlags.polling" />
                                <InputNumber v-model="pollingInterval" inputId="polling-interval" class="num-input" :min="1" :max="60" suffix=" s" @blur="flushPollingSave" />
                            </div>
                        </div>
                    </template>
                </section>

                <!-- ============ Presence ============ -->
                <section id="section-presence" class="pc-panel settings-section" aria-labelledby="hd-presence">
                    <div class="section-head">
                        <h2 id="hd-presence" class="pc-eyebrow section-eyebrow">{{ $t('settings.sectionPresence') }}</h2>
                        <SavedIndicator :visible="!!savedFlags.format" />
                    </div>

                    <div v-if="!loaded" class="skeleton-rows" aria-hidden="true">
                        <div class="pc-skeleton" style="width: 64%"></div>
                        <div class="pc-skeleton" style="width: 80%"></div>
                        <div class="pc-skeleton" style="width: 46%"></div>
                    </div>

                    <template v-else>
                        <div class="format-fields">
                            <div class="format-field">
                                <label class="row-label" for="format-details">{{ $t('settings.detailsLine') }}</label>
                                <InputText
                                    id="format-details"
                                    ref="detailsInputRef"
                                    v-model="detailsFormat"
                                    class="format-input"
                                    placeholder="{track}"
                                    spellcheck="false"
                                    autocomplete="off"
                                    @focus="lastFocusedFormat = 'details'"
                                    @blur="flushFormatSave"
                                />
                            </div>
                            <div class="format-field">
                                <label class="row-label" for="format-state">{{ $t('settings.stateLine') }}</label>
                                <InputText
                                    id="format-state"
                                    ref="stateInputRef"
                                    v-model="stateFormat"
                                    class="format-input"
                                    placeholder="by {artist} • {album}"
                                    spellcheck="false"
                                    autocomplete="off"
                                    @focus="lastFocusedFormat = 'state'"
                                    @blur="flushFormatSave"
                                />
                            </div>
                        </div>

                        <div class="token-row" role="group" :aria-label="$t('settings.insertTokenAria')">
                            <button v-for="token in FORMAT_TOKENS" :key="token" type="button" class="pc-chip-mono token-chip" @mousedown.prevent @click="insertToken(token)">{{ token }}</button>
                            <button type="button" class="pc-link token-reset" @click="resetFormats">{{ $t('settings.resetToDefaults') }}</button>
                        </div>
                        <p class="row-caption">{{ $t('settings.tokenCaption') }}</p>

                        <div class="format-specimen">
                            <DiscordSpecimen :track="specimenTrack" :formats="specimenFormats" :sample="specimenIsSample" :paused="specimenPaused" />
                        </div>

                        <div class="setting-row divided-row">
                            <div class="row-text">
                                <span class="row-label" id="lbl-hide-paused">{{ $t('settings.hideWhenPaused') }}</span>
                                <p class="row-caption">{{ $t('settings.hideWhenPausedCaption') }}</p>
                            </div>
                            <div class="row-control">
                                <SavedIndicator :visible="!!savedFlags.hidePaused" />
                                <ToggleSwitch :modelValue="hideWhenPaused" :disabled="hideWhenPausedSaving" aria-labelledby="lbl-hide-paused" @update:modelValue="updateHideWhenPaused" />
                            </div>
                        </div>
                        <div v-if="hideWhenPaused" class="setting-row sub-row">
                            <div class="row-text">
                                <label class="row-label" for="hide-delay">{{ $t('settings.delayBeforeClearing') }}</label>
                                <p class="row-caption">{{ $t('settings.delayImmediate') }}</p>
                            </div>
                            <div class="row-control">
                                <SavedIndicator :visible="!!savedFlags.hideDelay" />
                                <InputNumber v-model="hideWhenPausedDelay" inputId="hide-delay" class="num-input" :min="0" :max="300" suffix=" s" @blur="flushHideDelaySave" />
                            </div>
                        </div>
                    </template>
                </section>

                <!-- ============ App ============ -->
                <section id="section-app" class="pc-panel settings-section" aria-labelledby="hd-app">
                    <div class="section-head">
                        <h2 id="hd-app" class="pc-eyebrow section-eyebrow">{{ $t('settings.sectionApp') }}</h2>
                    </div>

                    <div v-if="!loaded" class="skeleton-rows" aria-hidden="true">
                        <div class="pc-skeleton" style="width: 55%"></div>
                        <div class="pc-skeleton" style="width: 62%"></div>
                    </div>

                    <template v-else>
                        <div class="setting-row">
                            <div class="row-text">
                                <span class="row-label" id="lbl-autostart">{{ $t('settings.startOnLogin') }}</span>
                                <p class="row-caption">{{ $t('settings.startOnLoginCaption') }}</p>
                            </div>
                            <div class="row-control">
                                <SavedIndicator :visible="!!savedFlags.autoStart" />
                                <ToggleSwitch :modelValue="autoStart" :disabled="autoStartSaving" aria-labelledby="lbl-autostart" @update:modelValue="updateAutoStart" />
                            </div>
                        </div>
                        <div class="setting-row">
                            <div class="row-text">
                                <span class="row-label" id="lbl-tray">{{ $t('settings.minimizeToTray') }}</span>
                                <p class="row-caption">{{ $t('settings.minimizeToTrayCaption') }}</p>
                            </div>
                            <div class="row-control">
                                <SavedIndicator :visible="!!savedFlags.minimizeToTray" />
                                <ToggleSwitch :modelValue="minimizeToTray" :disabled="minimizeToTraySaving" aria-labelledby="lbl-tray" @update:modelValue="updateMinimizeToTray" />
                            </div>
                        </div>
                        <div class="setting-row">
                            <div class="row-text">
                                <label class="row-label" for="language-select">{{ $t('settings.language') }}</label>
                                <p class="row-caption">{{ $t('settings.languageCaption') }}</p>
                            </div>
                            <div class="row-control">
                                <select id="language-select" class="language-select" :value="selectedLanguage" @change="changeLanguage($event.target.value)">
                                    <option v-for="option in languageOptions" :key="option.code" :value="option.code">{{ option.label }}</option>
                                </select>
                            </div>
                        </div>
                    </template>
                </section>

                <!-- ============ Advanced ============ -->
                <section id="section-advanced" class="pc-panel settings-section" aria-labelledby="hd-advanced">
                    <div class="section-head">
                        <h2 id="hd-advanced" class="pc-eyebrow section-eyebrow">{{ $t('settings.sectionAdvanced') }}</h2>
                    </div>

                    <div v-if="!loaded" class="skeleton-rows" aria-hidden="true">
                        <div class="pc-skeleton" style="width: 68%"></div>
                        <div class="pc-skeleton" style="width: 44%"></div>
                    </div>

                    <template v-else>
                        <div class="client-id-block">
                            <div class="client-id-head">
                                <label class="row-label" for="discord-client-id">{{ $t('settings.discordClientId') }}</label>
                                <span v-if="!isUsingDefaultClientId" class="pc-badge pc-badge--accent">{{ $t('settings.customIdBadge') }}</span>
                            </div>
                            <p class="row-caption">
                                {{ isUsingDefaultClientId ? $t('settings.usingDefault') : $t('settings.usingCustom') }}
                                <template v-if="!isUsingDefaultClientId"> · <button type="button" class="pc-link" @click="resetClientIdToDefault">{{ $t('settings.resetToDefault') }}</button> </template>
                            </p>
                            <div class="client-id-controls">
                                <InputText
                                    id="discord-client-id"
                                    v-model="discordClientId"
                                    class="client-id-input"
                                    :placeholder="defaultClientId"
                                    :invalid="!!clientIdError"
                                    spellcheck="false"
                                    autocomplete="off"
                                    aria-describedby="client-id-help"
                                    @keyup.enter="applyClientId"
                                />
                                <button type="button" class="pc-btn pc-btn--primary" :class="{ 'is-loading': applyingClientId }" :disabled="applyingClientId" @click="applyClientId">
                                    <span class="btn-label">{{ $t('settings.apply') }}</span>
                                    <i v-if="applyingClientId" class="pi pi-spinner pi-spin btn-spinner" aria-hidden="true"></i>
                                </button>
                                <SavedIndicator :visible="!!savedFlags.clientId" :label="$t('common.applied')" />
                            </div>
                            <p id="client-id-help" class="row-caption">{{ $t('settings.applyReconnects') }}</p>
                            <p v-if="clientIdError" class="row-caption row-caption--danger" role="alert"><i class="pi pi-exclamation-circle" aria-hidden="true"></i> {{ clientIdError }}</p>
                            <p v-else-if="clientIdWarning" class="row-caption row-caption--warn" role="alert"><i class="pi pi-exclamation-triangle" aria-hidden="true"></i> {{ clientIdWarning }}</p>
                        </div>

                        <div class="setting-row divided-row">
                            <div class="row-text">
                                <span class="row-label">{{ $t('settings.sendTestPresence') }}</span>
                                <p class="row-caption">{{ $t('settings.sendTestCaption') }}</p>
                            </div>
                            <div class="row-control">
                                <Transition name="pc-fade">
                                    <span v-if="testPresenceResult" class="test-result pc-fade-ok" :class="testPresenceResult.ok ? 'row-caption--success' : 'row-caption--danger'" role="status">
                                        <i :class="testPresenceResult.ok ? 'pi pi-check' : 'pi pi-exclamation-circle'" aria-hidden="true"></i>
                                        {{ testPresenceResult.message }}
                                    </span>
                                </Transition>
                                <button type="button" class="pc-btn pc-btn--secondary" :class="{ 'is-loading': sendingTestPresence }" :disabled="sendingTestPresence" @click="sendTestPresence">
                                    <span class="btn-label">{{ $t('settings.sendTestPresence') }}</span>
                                    <i v-if="sendingTestPresence" class="pi pi-spinner pi-spin btn-spinner" aria-hidden="true"></i>
                                </button>
                            </div>
                        </div>
                    </template>
                </section>

                <!-- ============ About + Danger zone ============ -->
                <section id="section-about" class="pc-panel settings-section" aria-labelledby="hd-about">
                    <div class="section-head">
                        <h2 id="hd-about" class="pc-eyebrow section-eyebrow">{{ $t('settings.sectionAbout') }}</h2>
                    </div>

                    <div class="setting-row">
                        <div class="row-text">
                            <span class="row-label">{{ $t('settings.version') }}</span>
                            <p class="row-caption version-chips">
                                <span class="pc-chip-mono" :title="buildDate ? $t('footer.buildDate', { date: buildDate }) : undefined">v{{ version || '—' }}</span>
                                <span v-if="commit" class="pc-chip-mono">{{ commit }}</span>
                            </p>
                        </div>
                        <div class="row-control">
                            <SavedIndicator :visible="!!savedFlags.upToDate" :label="$t('settings.upToDate')" />
                            <button type="button" class="pc-btn pc-btn--secondary" :class="{ 'is-loading': checkingUpdate }" :disabled="checkingUpdate" @click="checkForUpdates">
                                <span class="btn-label">{{ $t('settings.checkForUpdates') }}</span>
                                <i v-if="checkingUpdate" class="pi pi-spinner pi-spin btn-spinner" aria-hidden="true"></i>
                            </button>
                        </div>
                    </div>

                    <div v-if="updateInfo?.available" class="pc-panel--raised update-row">
                        <div class="update-row-main">
                            <div class="row-text">
                                <span class="row-label">{{ $t('settings.updateAvailable', { version: updateInfo.latestVersion }) }}</span>
                                <p v-if="truncatedReleaseNotes" class="row-caption">{{ truncatedReleaseNotes }}</p>
                                <p v-if="updateReady" class="row-caption row-caption--success">{{ $t('settings.updateInstalled', { version: updateInfo.latestVersion }) }}</p>
                            </div>
                            <!-- Self-updating platforms install in place; the rest fall back to the release page -->
                            <button v-if="updateReady" type="button" class="pc-btn pc-btn--success" @click="restartApp"><i class="pi pi-refresh" aria-hidden="true"></i>{{ $t('settings.restartNow') }}</button>
                            <button v-else-if="canSelfUpdate" type="button" class="pc-btn pc-btn--primary" :class="{ 'is-loading': installingUpdate }" :disabled="installingUpdate" @click="installUpdate">
                                <span class="btn-label"><i class="pi pi-download" aria-hidden="true"></i>{{ $t('settings.downloadInstall') }}</span>
                                <i v-if="installingUpdate" class="pi pi-spinner pi-spin btn-spinner" aria-hidden="true"></i>
                            </button>
                            <button v-else type="button" class="pc-btn pc-btn--primary" @click="openUpdatePage"><i class="pi pi-download" aria-hidden="true"></i>{{ $t('settings.download') }}</button>
                        </div>
                        <div v-if="installingUpdate" class="update-progress">
                            <div class="update-progress-track" role="progressbar" :aria-label="$t('settings.updateProgressAria')" :aria-valuenow="updateProgress" aria-valuemin="0" aria-valuemax="100">
                                <div class="update-progress-fill" :style="{ width: `${updateProgress}%` }"></div>
                            </div>
                            <p class="row-caption">{{ $t('settings.downloading', { percent: updateProgress }) }}</p>
                        </div>
                    </div>

                    <div class="setting-row">
                        <div class="row-text">
                            <span class="row-label">{{ $t('settings.changelog') }}</span>
                            <p class="row-caption">{{ $t('settings.changelogCaption') }}</p>
                        </div>
                        <button type="button" class="pc-btn pc-btn--ghost" @click="openReleases">{{ $t('settings.viewChangelog') }}<i class="pi pi-external-link" aria-hidden="true"></i></button>
                    </div>

                    <div class="danger-zone">
                        <h3 class="pc-eyebrow danger-eyebrow">{{ $t('settings.dangerZone') }}</h3>
                        <div class="setting-row">
                            <div class="row-text">
                                <span class="row-label">{{ $t('settings.resetApplication') }}</span>
                                <p class="row-caption">{{ $t('settings.resetCaption') }}</p>
                            </div>
                            <button type="button" class="pc-btn pc-btn--ghost-danger" :class="{ 'is-loading': resetting }" :disabled="resetting" @click="confirmReset">
                                <span class="btn-label">{{ $t('settings.resetApplicationEllipsis') }}</span>
                                <i v-if="resetting" class="pi pi-spinner pi-spin btn-spinner" aria-hidden="true"></i>
                            </button>
                        </div>
                    </div>
                </section>
            </div>
        </div>

        <!-- Add server dialog -->
        <Dialog v-model:visible="showAddServerDialog" modal :header="$t('settings.dialogAddServer')" :style="{ width: '420px' }">
            <div class="dialog-body">
                <!-- Auto-discovery (GDM) -->
                <div class="dialog-discovery">
                    <button v-if="!isDiscovering" type="button" class="pc-btn pc-btn--secondary discovery-btn" @click="discoverServers">
                        <i class="pi pi-search" aria-hidden="true"></i>{{ hasDiscovered ? $t('common.searchAgain') : $t('settings.discoverServers') }}
                    </button>
                    <p v-else class="row-caption discovery-searching" role="status"><i class="pi pi-spinner pi-spin inline-spinner" aria-hidden="true"></i> {{ $t('settings.searching') }}</p>

                    <p v-if="discoveryError" class="row-caption row-caption--danger" role="alert"><i class="pi pi-exclamation-circle" aria-hidden="true"></i> {{ discoveryError }}</p>

                    <ul v-if="hasDiscovered && discoveredServers.length > 0" class="discovered-list">
                        <li v-for="server in discoveredServers" :key="`${server.address}:${server.port}`">
                            <button
                                type="button"
                                class="discovered-row"
                                :class="{ 'discovered-row--added': isServerAlreadyAdded(server), 'discovered-row--selected': newServerURL === discoveredServerURL(server) }"
                                :disabled="isServerAlreadyAdded(server)"
                                :aria-pressed="newServerURL === discoveredServerURL(server)"
                                @click="selectDiscoveredServer(server)"
                            >
                                <span class="server-name discovered-name">{{ server.name || $t('settings.defaultServerName') }}</span>
                                <span class="pc-chip-mono discovered-url">{{ server.address }}:{{ server.port }}</span>
                                <span v-if="isServerAlreadyAdded(server)" class="pc-badge">{{ $t('settings.added') }}</span>
                                <span v-else class="pc-badge">{{ server.isLocal ? $t('settings.local') : $t('settings.remote') }}</span>
                            </button>
                        </li>
                    </ul>

                    <p v-if="hasDiscovered && !discoveryError && discoveredServers.length === 0" class="row-caption">{{ $t('settings.noServersFound') }}</p>
                </div>

                <div class="dialog-divider" aria-hidden="true">
                    <span class="pc-eyebrow">{{ $t('settings.orEnterManually') }}</span>
                </div>

                <div class="dialog-field">
                    <label class="row-label" for="new-server-name">{{ $t('settings.serverName') }}</label>
                    <InputText id="new-server-name" v-model="newServerName" :placeholder="$t('settings.serverNamePlaceholder')" class="dialog-input" autocomplete="off" @keyup.enter="addServer" />
                </div>
                <div class="dialog-field">
                    <label class="row-label" for="new-server-url">{{ $t('settings.serverUrl') }}</label>
                    <InputText
                        id="new-server-url"
                        v-model="newServerURL"
                        :placeholder="PLEX_URL_PLACEHOLDER"
                        class="dialog-input dialog-input--mono"
                        :invalid="newServerUrlTouched && !newServerUrlValidation.valid"
                        spellcheck="false"
                        autocomplete="off"
                        aria-describedby="new-server-url-help"
                        @keyup.enter="addServer"
                    />
                    <p v-if="newServerUrlTouched && !newServerUrlValidation.valid" id="new-server-url-help" class="row-caption row-caption--danger" role="alert">
                        <i class="pi pi-exclamation-circle" aria-hidden="true"></i> {{ newServerUrlValidation.error }}
                    </p>
                    <p v-else-if="newServerUrlValidation.valid" id="new-server-url-help" class="row-caption row-caption--success"><i class="pi pi-check-circle" aria-hidden="true"></i> {{ $t('settings.validUrlFormat') }}</p>
                    <p v-else id="new-server-url-help" class="row-caption">{{ $t('settings.urlHelp') }}</p>
                </div>
                <p v-if="addServerError" class="row-caption row-caption--danger" role="alert"><i class="pi pi-exclamation-circle" aria-hidden="true"></i> {{ addServerError }}</p>
            </div>
            <template #footer>
                <p v-if="!canAddServer" class="row-caption dialog-gate-caption">{{ $t('settings.gateCaption') }}</p>
                <button type="button" class="pc-btn pc-btn--ghost" @click="showAddServerDialog = false">{{ $t('common.cancel') }}</button>
                <button type="button" class="pc-btn pc-btn--primary" :class="{ 'is-loading': addingServer }" :disabled="!canAddServer || addingServer" @click="addServer">
                    <span class="btn-label">{{ $t('settings.addServer') }}</span>
                    <i v-if="addingServer" class="pi pi-spinner pi-spin btn-spinner" aria-hidden="true"></i>
                </button>
            </template>
        </Dialog>
    </div>
</template>

<style scoped>
.settings-page {
    max-width: 1040px;
    margin: 0 auto;
}
.settings-title {
    margin: 0 0 20px;
    font-size: var(--pc-text-title);
    font-weight: 600;
    line-height: 1.3;
    letter-spacing: -0.015em;
    color: var(--pc-text);
}
.settings-layout {
    display: grid;
    grid-template-columns: 200px minmax(0, 1fr);
    gap: var(--pc-space-section);
    align-items: start;
}

/* ---------------- Rail ---------------- */
.settings-rail {
    position: sticky;
    top: 72px;
}
.rail-eyebrow {
    margin: 0 0 8px;
    padding-left: 12px;
}
.rail-items {
    position: relative;
    display: flex;
    flex-direction: column;
    border-radius: var(--pc-radius-sm);
}
.rail-bar {
    position: absolute;
    left: 0;
    top: 0;
    width: 2px;
    height: 20px;
    border-radius: 1px;
    background: var(--pc-accent);
    transition: transform var(--pc-dur-3) var(--pc-ease-inout);
}
.rail-item {
    display: block;
    width: 100%;
    height: 32px;
    padding: 0 12px;
    border: none;
    background: transparent;
    border-radius: var(--pc-radius-sm);
    text-align: left;
    font-family: var(--pc-font-ui);
    font-size: var(--pc-text-body);
    line-height: 32px;
    color: var(--pc-text-secondary);
    cursor: pointer;
    transition: color var(--pc-dur-1) var(--pc-ease-out);
}
.rail-item:hover {
    color: var(--pc-text);
}
.rail-item--active {
    color: var(--pc-accent);
    font-weight: 500;
}
.rail-items:focus-visible .rail-item--focused {
    box-shadow: var(--pc-ring-focus);
}

/* ---------------- Sections ---------------- */
.settings-section {
    margin-bottom: var(--pc-space-panel-gap);
    scroll-margin-top: 72px;
}
.settings-section:last-child {
    margin-bottom: 0;
}
.section-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-height: 28px;
    margin-bottom: 12px;
}
.section-eyebrow {
    margin: 0;
}

.skeleton-rows {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 6px 0 10px;
}
.skeleton-rows .pc-skeleton {
    height: 14px;
}

/* ---------------- Rows ---------------- */
.setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-height: 44px;
    padding: 12px 0;
}
.setting-row + .setting-row,
.divided-row {
    border-top: 1px solid var(--pc-border-subtle);
}
.sub-row {
    padding-left: 16px;
}
.row-text {
    min-width: 0;
}
.row-label {
    display: inline-block;
    font-size: var(--pc-text-body);
    font-weight: 500;
    color: var(--pc-text);
}
.row-caption {
    margin: 2px 0 0;
    font-size: var(--pc-text-caption);
    line-height: 1.45;
    color: var(--pc-text-muted);
}
.row-caption .pi {
    font-size: 11px;
}
.row-caption--danger {
    color: var(--pc-danger);
}
.row-caption--warn {
    color: var(--pc-warn);
}
.row-caption--success {
    color: var(--pc-success);
}
.row-control {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: none;
}
.pc-link {
    padding: 0;
    border: none;
    background: none;
    font-size: inherit;
    color: var(--pc-accent);
    cursor: pointer;
}
.pc-link:hover {
    text-decoration: underline;
    color: var(--pc-accent-hover);
}
.inline-spinner {
    font-size: 12px;
    color: var(--pc-text-muted);
}

/* Button loading (§5.0.2): label persists at 40%, 12px absolute spinner */
.pc-btn {
    position: relative;
    white-space: nowrap;
}
.pc-btn.is-loading .btn-label {
    opacity: 0.4;
}
.btn-spinner {
    position: absolute;
    left: 50%;
    top: 50%;
    margin: -6px 0 0 -6px;
    font-size: 12px;
}
.pc-btn .pi {
    font-size: 12px;
}

/* ---------------- Servers ---------------- */
.server-list {
    margin-bottom: 4px;
}
.empty-caption {
    padding: 12px 0;
}
.server-row {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 44px;
    padding: 12px 0;
}
.server-row + .server-row {
    border-top: 1px solid var(--pc-border-subtle);
}
.server-main {
    flex: 1;
    min-width: 0;
}
.server-line {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex-wrap: wrap;
}
.server-name {
    font-weight: 500;
    color: var(--pc-text);
}
.server-url {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.server-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: none;
}
.icon-danger {
    color: var(--pc-danger);
}
.icon-danger:hover:not(:disabled) {
    background: var(--pc-danger-dim);
    color: var(--pc-danger);
}
.polling-row {
    border-top: 1px solid var(--pc-border-subtle);
}

/* ---------------- Format editor ---------------- */
.format-fields {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 12px;
}
.format-field {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
}
.format-input {
    width: 100%;
    font-family: var(--pc-font-mono);
    font-size: var(--pc-text-mono);
}
.token-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 10px;
}
.token-chip {
    border: 1px solid var(--pc-border);
    cursor: pointer;
    transition:
        border-color var(--pc-dur-1) var(--pc-ease-out),
        color var(--pc-dur-1) var(--pc-ease-out);
}
.token-chip:hover {
    border-color: var(--pc-accent);
    color: var(--pc-accent);
}
.token-reset {
    margin-left: auto;
    font-size: var(--pc-text-caption);
}
.format-specimen {
    margin: 16px 0;
}

/* ---------------- Advanced ---------------- */
.client-id-block {
    padding: 12px 0;
}
.client-id-head {
    display: flex;
    align-items: center;
    gap: 8px;
}
.client-id-controls {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 10px;
    flex-wrap: wrap;
}
.client-id-input {
    width: 240px;
    max-width: 100%;
    font-family: var(--pc-font-mono);
    font-size: var(--pc-text-mono);
}
.test-result {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: var(--pc-text-caption);
    white-space: nowrap;
}
.test-result .pi {
    font-size: 11px;
}

/* ---------------- About + danger zone ---------------- */
.version-chips {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
}
.update-row {
    padding: 12px 16px;
    margin: 4px 0 8px;
    border: 1px solid var(--pc-border);
}
.update-row-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}
.update-row-main .pc-btn {
    flex: none;
}
.update-progress {
    margin-top: 10px;
}
.update-progress-track {
    height: 4px;
    border-radius: var(--pc-radius-full);
    background: var(--pc-raised);
    border: 1px solid var(--pc-border-subtle);
    overflow: hidden;
}
.update-progress-fill {
    height: 100%;
    border-radius: var(--pc-radius-full);
    background: var(--pc-accent);
    transition: width var(--pc-dur-2) var(--pc-ease-out);
}
.update-progress .row-caption {
    margin-top: 6px;
    font-variant-numeric: tabular-nums;
}
/* Success-severity confirm for the restart step (tokens only) */
.pc-btn--success {
    background: var(--pc-success);
    border-color: transparent;
    color: var(--pc-accent-contrast);
}
.pc-btn--success:hover:not(:disabled) {
    background: color-mix(in srgb, var(--pc-success) 88%, var(--pc-text));
}
.danger-zone {
    margin-top: 8px;
    padding-top: 12px;
    border-top: 1px solid var(--pc-border-subtle);
}
.danger-eyebrow {
    margin: 0;
    color: var(--pc-danger);
}

/* ---------------- Dialog ---------------- */
.dialog-body {
    display: flex;
    flex-direction: column;
    gap: 14px;
}
.dialog-discovery {
    display: flex;
    flex-direction: column;
    gap: 10px;
}
.discovery-btn {
    width: 100%;
}
.discovery-searching {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 32px;
    margin: 0;
}
.discovered-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.discovered-row {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    min-height: 44px;
    padding: 10px 12px;
    background: var(--pc-raised);
    border: 1px solid var(--pc-border);
    border-radius: var(--pc-radius-md);
    text-align: left;
    cursor: pointer;
    color: var(--pc-text);
    font-family: var(--pc-font-ui);
    transition:
        border-color var(--pc-dur-1) var(--pc-ease-out),
        background-color var(--pc-dur-1) var(--pc-ease-out);
}
.discovered-row:hover:not(:disabled) {
    border-color: var(--pc-border-strong);
}
.discovered-row--selected {
    border-color: var(--pc-accent);
    background: var(--pc-accent-dim);
}
.discovered-row--added {
    opacity: 0.6;
    cursor: not-allowed;
}
.discovered-name {
    font-size: var(--pc-text-body);
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.discovered-url {
    margin-left: auto;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.dialog-divider {
    display: flex;
    align-items: center;
    gap: 10px;
}
.dialog-divider::before,
.dialog-divider::after {
    content: '';
    flex: 1;
    border-top: 1px solid var(--pc-border-subtle);
}
.dialog-field {
    display: flex;
    flex-direction: column;
    gap: 6px;
}
.dialog-input {
    width: 100%;
}
.dialog-input--mono {
    font-family: var(--pc-font-mono);
    font-size: var(--pc-text-mono);
}
.dialog-gate-caption {
    margin-right: auto;
}

/* ---------------- Controls sizing ---------------- */
.num-input :deep(.p-inputnumber-input) {
    width: 96px;
}

/* Language picker: native select styled to match the form-field recipe. */
.language-select {
    min-width: 140px;
    height: 34px;
    padding: 0 10px;
    border-radius: var(--pc-radius-sm);
    border: 1px solid var(--p-inputtext-border-color, var(--pc-border));
    background: var(--p-inputtext-background, var(--pc-raised));
    color: var(--pc-text);
    font-family: var(--pc-font-ui);
    font-size: var(--pc-text-body);
    cursor: pointer;
}
.language-select:focus-visible {
    outline: none;
    box-shadow: var(--pc-ring-focus);
    border-color: var(--pc-accent);
}

/* ---------------- Responsive ---------------- */
@media (max-width: 860px) {
    .settings-layout {
        grid-template-columns: 1fr;
        gap: 16px;
    }
    .settings-rail {
        position: static;
    }
    .rail-eyebrow {
        display: none;
    }
    .rail-items {
        flex-direction: row;
        flex-wrap: wrap;
        gap: 4px;
    }
    .rail-bar {
        display: none;
    }
    .rail-item {
        width: auto;
    }
    .rail-item--active {
        background: var(--pc-accent-dim);
    }
}
</style>
