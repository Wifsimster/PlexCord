<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useSetupStore } from '@/stores/setup';
import Button from 'primevue/button';
import InputNumber from 'primevue/inputnumber';
import ToggleSwitch from 'primevue/toggleswitch';
import InputText from 'primevue/inputtext';
import Dialog from 'primevue/dialog';
import { useToast } from 'primevue/usetoast';
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
    GetVersion,
    CheckForUpdate,
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
    SetServerActive
} from '../../../wailsjs/go/main/App';

const router = useRouter();
const toast = useToast();
const setupStore = useSetupStore();

// Settings state
const pollingInterval = ref(2);
const autoStart = ref(false);
const minimizeToTray = ref(true);
const discordClientId = ref('');
const defaultClientId = ref('');

// Hide when paused
const hideWhenPaused = ref(false);
const hideWhenPausedDelay = ref(0);

// Custom presence format
const presenceDetailsFormat = ref('');
const presenceStateFormat = ref('');

// Version info
const version = ref(null);
const updateInfo = ref(null);
const checkingUpdate = ref(false);

// Loading states
const loading = ref({
    polling: false,
    autoStart: false,
    minimizeToTray: false,
    clientId: false,
    hideWhenPaused: false,
    presenceFormat: false,
    reset: false
});

// Servers
const servers = ref([]);
const showAddServerDialog = ref(false);
const newServerName = ref('');
const newServerURL = ref('');

// Reset confirmation dialog
const showResetDialog = ref(false);

// Load settings on mount
onMounted(async () => {
    try {
        pollingInterval.value = await GetPollingInterval();
        autoStart.value = await GetAutoStart();
        minimizeToTray.value = await GetMinimizeToTray();
        discordClientId.value = await GetDiscordClientID();
        defaultClientId.value = await GetDefaultDiscordClientID();
        version.value = await GetVersion();

        // Load hide-when-paused settings
        const pauseSettings = await GetHideWhenPaused();
        hideWhenPaused.value = pauseSettings.enabled;
        hideWhenPausedDelay.value = pauseSettings.delaySeconds;

        // Load presence format settings
        const formatSettings = await GetPresenceFormat();
        presenceDetailsFormat.value = formatSettings.detailsFormat;
        presenceStateFormat.value = formatSettings.stateFormat;

        // Load servers
        servers.value = await GetServers();
    } catch (error) {
        console.error('Failed to load settings:', error);
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to load settings',
            life: 3000
        });
    }
});

// Computed
const isUsingDefaultClientId = computed(() => {
    return !discordClientId.value || discordClientId.value === defaultClientId.value;
});

// Settings handlers
const updatePollingInterval = async () => {
    loading.value.polling = true;
    try {
        await SetPollingInterval(pollingInterval.value);
        toast.add({
            severity: 'success',
            summary: 'Saved',
            detail: 'Polling interval updated',
            life: 2000
        });
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to update polling interval',
            life: 3000
        });
    } finally {
        loading.value.polling = false;
    }
};

const updateAutoStart = async (value) => {
    loading.value.autoStart = true;
    try {
        await SetAutoStart(value);
        autoStart.value = value;
    } catch (error) {
        // Revert on error
        autoStart.value = !value;
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to update auto-start setting',
            life: 3000
        });
    } finally {
        loading.value.autoStart = false;
    }
};

const updateMinimizeToTray = async (value) => {
    loading.value.minimizeToTray = true;
    try {
        await SetMinimizeToTray(value);
        minimizeToTray.value = value;
    } catch (error) {
        // Revert on error
        minimizeToTray.value = !value;
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to update minimize to tray setting',
            life: 3000
        });
    } finally {
        loading.value.minimizeToTray = false;
    }
};

const saveDiscordClientId = async () => {
    loading.value.clientId = true;
    try {
        // Validate first
        if (discordClientId.value && discordClientId.value !== defaultClientId.value) {
            await ValidateDiscordClientID(discordClientId.value);
        }
        await SaveDiscordClientID(discordClientId.value);
        toast.add({
            severity: 'success',
            summary: 'Saved',
            detail: 'Discord Client ID updated',
            life: 2000
        });
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: error.message || 'Invalid Discord Client ID',
            life: 3000
        });
    } finally {
        loading.value.clientId = false;
    }
};

const resetToDefaultClientId = () => {
    discordClientId.value = '';
    saveDiscordClientId();
};

// Hide when paused handler
const updateHideWhenPaused = async (value) => {
    loading.value.hideWhenPaused = true;
    try {
        await SetHideWhenPaused(value, hideWhenPausedDelay.value);
        hideWhenPaused.value = value;
    } catch (error) {
        hideWhenPaused.value = !value;
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to update hide when paused setting',
            life: 3000
        });
    } finally {
        loading.value.hideWhenPaused = false;
    }
};

const updateHideWhenPausedDelay = async () => {
    loading.value.hideWhenPaused = true;
    try {
        await SetHideWhenPaused(hideWhenPaused.value, hideWhenPausedDelay.value);
        toast.add({
            severity: 'success',
            summary: 'Saved',
            detail: 'Hide when paused delay updated',
            life: 2000
        });
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to update delay',
            life: 3000
        });
    } finally {
        loading.value.hideWhenPaused = false;
    }
};

// Presence format handlers
const formatPreviewDetails = computed(() => {
    const fmt = presenceDetailsFormat.value || '{track}';
    return fmt.replace('{track}', 'Bohemian Rhapsody').replace('{artist}', 'Queen').replace('{album}', 'A Night at the Opera').replace('{year}', '1975').replace('{player}', 'Plexamp');
});

const formatPreviewState = computed(() => {
    const fmt = presenceStateFormat.value || 'by {artist} \u2022 {album}';
    return fmt.replace('{track}', 'Bohemian Rhapsody').replace('{artist}', 'Queen').replace('{album}', 'A Night at the Opera').replace('{year}', '1975').replace('{player}', 'Plexamp');
});

const savePresenceFormat = async () => {
    loading.value.presenceFormat = true;
    try {
        await SetPresenceFormat(presenceDetailsFormat.value, presenceStateFormat.value);
        toast.add({
            severity: 'success',
            summary: 'Saved',
            detail: 'Presence format updated',
            life: 2000
        });
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to save presence format',
            life: 3000
        });
    } finally {
        loading.value.presenceFormat = false;
    }
};

const resetPresenceFormat = () => {
    presenceDetailsFormat.value = '';
    presenceStateFormat.value = '';
    savePresenceFormat();
};

// Server management handlers
const loadServers = async () => {
    try {
        servers.value = await GetServers();
    } catch (error) {
        console.error('Failed to load servers:', error);
    }
};

const toggleServerActive = async (server) => {
    try {
        await SetServerActive(server.url, !server.active);
        await loadServers();
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to update server status',
            life: 3000
        });
    }
};

const removeServer = async (server) => {
    try {
        await RemoveServer(server.url);
        await loadServers();
        toast.add({
            severity: 'success',
            summary: 'Removed',
            detail: `Server "${server.name}" removed`,
            life: 2000
        });
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to remove server',
            life: 3000
        });
    }
};

const openAddServerDialog = () => {
    newServerName.value = '';
    newServerURL.value = '';
    showAddServerDialog.value = true;
};

const addServer = async () => {
    if (!newServerName.value || !newServerURL.value) {
        toast.add({
            severity: 'warn',
            summary: 'Validation',
            detail: 'Server name and URL are required',
            life: 3000
        });
        return;
    }
    try {
        await AddServer(newServerName.value, newServerURL.value, '', '');
        showAddServerDialog.value = false;
        await loadServers();
        toast.add({
            severity: 'success',
            summary: 'Added',
            detail: `Server "${newServerName.value}" added`,
            life: 2000
        });
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: error.message || 'Failed to add server',
            life: 3000
        });
    }
};

// Update check
const checkForUpdates = async () => {
    checkingUpdate.value = true;
    try {
        updateInfo.value = await CheckForUpdate();
        if (updateInfo.value.available) {
            toast.add({
                severity: 'info',
                summary: 'Update Available',
                detail: `Version ${updateInfo.value.latestVersion} is available`,
                life: 5000
            });
        } else {
            toast.add({
                severity: 'success',
                summary: 'Up to Date',
                detail: 'You are running the latest version',
                life: 3000
            });
        }
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to check for updates',
            life: 3000
        });
    } finally {
        checkingUpdate.value = false;
    }
};

const openUpdatePage = () => {
    if (updateInfo.value?.releaseUrl) {
        OpenReleaseURL(updateInfo.value.releaseUrl);
    } else {
        OpenReleasesPage();
    }
};

// Reset application
const confirmReset = () => {
    showResetDialog.value = true;
};

const executeReset = async () => {
    loading.value.reset = true;
    try {
        await ResetApplication();
        showResetDialog.value = false;
        // Reset the wizard state to start from the beginning
        setupStore.resetWizard();
        toast.add({
            severity: 'success',
            summary: 'Reset Complete',
            detail: 'Application has been reset. Redirecting to setup...',
            life: 3000
        });
        // Redirect to setup after short delay
        setTimeout(() => {
            router.push('/setup/welcome');
        }, 1500);
    } catch (error) {
        toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to reset application',
            life: 3000
        });
    } finally {
        loading.value.reset = false;
    }
};

// Navigation
const goToDashboard = () => {
    router.push('/');
};
</script>

<template>
    <div class="max-w-7xl mx-auto">
        <!-- Header -->
        <div class="flex items-center gap-4 mb-6">
            <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="goToDashboard" />
            <div>
                <h1 class="text-2xl font-bold text-surface-900 dark:text-surface-0">Settings</h1>
                <p class="text-muted-color">Configure PlexCord behavior</p>
            </div>
        </div>

        <!-- Connection Settings -->
        <div class="card mb-4">
            <h2 class="text-xl font-semibold mb-4">Connection</h2>

            <!-- Polling Interval -->
            <div class="py-3 border-b border-surface-200 dark:border-surface-700">
                <div class="mb-2">
                    <div class="font-medium">Polling Interval</div>
                    <div class="text-sm text-muted-color">How often to check for playback changes (1-60 seconds)</div>
                </div>
                <div class="flex gap-2">
                    <InputNumber v-model="pollingInterval" :min="1" :max="60" suffix=" sec" :disabled="loading.polling" class="flex-grow" @keyup.enter="updatePollingInterval" />
                    <Button label="Save" :loading="loading.polling" @click="updatePollingInterval" />
                </div>
            </div>

            <!-- Discord Client ID -->
            <div class="py-3">
                <div class="flex items-center justify-between mb-2">
                    <div>
                        <div class="font-medium">Discord Client ID</div>
                        <div class="text-sm text-muted-color">
                            {{ isUsingDefaultClientId ? 'Using default PlexCord application' : 'Using custom application' }}
                        </div>
                    </div>
                    <Button v-if="!isUsingDefaultClientId" label="Reset to Default" severity="secondary" size="small" @click="resetToDefaultClientId" />
                </div>
                <div class="flex gap-2">
                    <InputText v-model="discordClientId" :placeholder="defaultClientId" class="flex-grow" :disabled="loading.clientId" @keyup.enter="saveDiscordClientId" />
                    <Button label="Save" :loading="loading.clientId" @click="saveDiscordClientId" />
                </div>
            </div>
        </div>

        <!-- Servers -->
        <div class="card mb-4">
            <div class="flex items-center justify-between mb-4">
                <h2 class="text-xl font-semibold">Servers</h2>
                <Button label="Add Server" icon="pi pi-plus" size="small" @click="openAddServerDialog" />
            </div>

            <div v-if="servers.length === 0" class="py-3 text-muted-color text-sm">
                No servers configured. Add a server to get started.
            </div>

            <div v-for="(server, index) in servers" :key="server.url"
                 :class="['flex items-center justify-between py-3', { 'border-b border-surface-200 dark:border-surface-700': index < servers.length - 1 }]">
                <div class="flex-1 min-w-0 mr-4">
                    <div class="font-medium truncate">{{ server.name }}</div>
                    <div class="text-sm text-muted-color truncate">{{ server.url }}</div>
                    <div v-if="server.userName" class="text-xs text-muted-color">User: {{ server.userName }}</div>
                </div>
                <div class="flex items-center gap-3 shrink-0">
                    <ToggleSwitch :modelValue="server.active" @update:modelValue="() => toggleServerActive(server)" aria-label="Toggle server active" />
                    <Button icon="pi pi-trash" severity="danger" text rounded size="small" @click="removeServer(server)" />
                </div>
            </div>
        </div>

        <!-- Add Server Dialog -->
        <Dialog v-model:visible="showAddServerDialog" modal header="Add Server" :style="{ width: '400px' }">
            <div class="flex flex-col gap-4">
                <div>
                    <label class="block font-medium mb-1">Server Name</label>
                    <InputText v-model="newServerName" placeholder="My Plex Server" class="w-full" @keyup.enter="addServer" />
                </div>
                <div>
                    <label class="block font-medium mb-1">Server URL</label>
                    <InputText v-model="newServerURL" placeholder="http://192.168.1.100:32400" class="w-full" @keyup.enter="addServer" />
                </div>
            </div>

            <template #footer>
                <Button label="Cancel" severity="secondary" @click="showAddServerDialog = false" />
                <Button label="Add Server" @click="addServer" />
            </template>
        </Dialog>

        <!-- Behavior Settings -->
        <div class="card mb-4">
            <h2 class="text-xl font-semibold mb-4">Behavior</h2>

            <!-- Auto-start -->
            <div class="flex items-center justify-between py-3 border-b border-surface-200 dark:border-surface-700">
                <div>
                    <div class="font-medium">Start on Login</div>
                    <div class="text-sm text-muted-color">Automatically launch PlexCord when you log in</div>
                </div>
                <ToggleSwitch :modelValue="autoStart" @update:modelValue="updateAutoStart" :disabled="loading.autoStart" aria-label="Start on Login" />
            </div>

            <!-- Minimize to Tray -->
            <div class="flex items-center justify-between py-3 border-b border-surface-200 dark:border-surface-700">
                <div>
                    <div class="font-medium">Minimize to Tray</div>
                    <div class="text-sm text-muted-color">Keep running in system tray when window is closed</div>
                </div>
                <ToggleSwitch :modelValue="minimizeToTray" @update:modelValue="updateMinimizeToTray" :disabled="loading.minimizeToTray" aria-label="Minimize to Tray" />
            </div>

            <!-- Hide When Paused -->
            <div class="py-3">
                <div class="flex items-center justify-between mb-2">
                    <div>
                        <div class="font-medium">Hide When Paused</div>
                        <div class="text-sm text-muted-color">Clear Discord presence when playback is paused</div>
                    </div>
                    <ToggleSwitch :modelValue="hideWhenPaused" @update:modelValue="updateHideWhenPaused" :disabled="loading.hideWhenPaused" aria-label="Hide When Paused" />
                </div>
                <div v-if="hideWhenPaused" class="flex gap-2 mt-2">
                    <InputNumber v-model="hideWhenPausedDelay" :min="0" :max="300" suffix=" sec" :disabled="loading.hideWhenPaused" class="flex-grow" placeholder="0 = immediate" @keyup.enter="updateHideWhenPausedDelay" />
                    <Button label="Save" :loading="loading.hideWhenPaused" @click="updateHideWhenPausedDelay" />
                </div>
                <div v-if="hideWhenPaused" class="text-xs text-muted-color mt-1">Delay before clearing (0 = immediate)</div>
            </div>
        </div>

        <!-- Appearance Settings -->
        <div class="card mb-4">
            <h2 class="text-xl font-semibold mb-4">Appearance</h2>

            <!-- Presence Details Format -->
            <div class="py-3 border-b border-surface-200 dark:border-surface-700">
                <div class="mb-2">
                    <div class="font-medium">Presence Details Format</div>
                    <div class="text-sm text-muted-color">First line of Discord presence. Tokens: {track}, {artist}, {album}, {year}, {player}</div>
                </div>
                <div class="flex gap-2">
                    <InputText v-model="presenceDetailsFormat" placeholder="{track}" class="flex-grow" :disabled="loading.presenceFormat" @keyup.enter="savePresenceFormat" />
                </div>
            </div>

            <!-- Presence State Format -->
            <div class="py-3 border-b border-surface-200 dark:border-surface-700">
                <div class="mb-2">
                    <div class="font-medium">Presence State Format</div>
                    <div class="text-sm text-muted-color">Second line of Discord presence. Tokens: {track}, {artist}, {album}, {year}, {player}</div>
                </div>
                <div class="flex gap-2">
                    <InputText v-model="presenceStateFormat" placeholder="by {artist} &bull; {album}" class="flex-grow" :disabled="loading.presenceFormat" @keyup.enter="savePresenceFormat" />
                </div>
            </div>

            <!-- Preview & Save -->
            <div class="py-3">
                <div class="mb-3 p-3 rounded-lg bg-surface-100 dark:bg-surface-800">
                    <div class="text-xs text-muted-color mb-1">Preview</div>
                    <div class="font-medium text-surface-900 dark:text-surface-0">{{ formatPreviewDetails }}</div>
                    <div class="text-sm text-muted-color">{{ formatPreviewState }}</div>
                </div>
                <div class="flex gap-2 justify-end">
                    <Button label="Reset to Defaults" severity="secondary" size="small" @click="resetPresenceFormat" />
                    <Button label="Save" :loading="loading.presenceFormat" @click="savePresenceFormat" />
                </div>
            </div>
        </div>

        <!-- About Section -->
        <div class="card mb-4">
            <h2 class="text-xl font-semibold mb-4">About</h2>

            <!-- Version -->
            <div class="flex items-center justify-between py-3 border-b border-surface-200 dark:border-surface-700">
                <div>
                    <div class="font-medium">Version</div>
                    <div class="text-sm text-muted-color">
                        {{ version?.version || 'Loading...' }}
                        <span v-if="version?.commit && version.commit !== 'unknown'" class="ml-1"> ({{ version.commit.substring(0, 7) }}) </span>
                    </div>
                </div>
                <div class="flex gap-2">
                    <Button label="Check for Updates" icon="pi pi-refresh" severity="secondary" size="small" :loading="checkingUpdate" @click="checkForUpdates" />
                </div>
            </div>

            <!-- Update available banner -->
            <div v-if="updateInfo?.available" class="mt-3 p-3 rounded-lg bg-blue-100 dark:bg-blue-900/20 flex items-center justify-between">
                <div>
                    <div class="font-medium text-blue-700 dark:text-blue-400">Update Available: {{ updateInfo.latestVersion }}</div>
                    <div class="text-sm text-blue-600 dark:text-blue-500">{{ updateInfo.releaseNotes?.substring(0, 100) }}{{ updateInfo.releaseNotes?.length > 100 ? '...' : '' }}</div>
                </div>
                <Button label="Download" icon="pi pi-download" size="small" @click="openUpdatePage" />
            </div>

            <!-- Changelog link -->
            <div class="flex items-center justify-between py-3">
                <div>
                    <div class="font-medium">Changelog</div>
                    <div class="text-sm text-muted-color">View release notes and version history</div>
                </div>
                <Button label="View Changelog" icon="pi pi-external-link" severity="secondary" size="small" @click="OpenReleasesPage" />
            </div>
        </div>

        <!-- Danger Zone -->
        <div class="card border-red-200 dark:border-red-900">
            <h2 class="text-xl font-semibold mb-4 text-red-600 dark:text-red-400">Danger Zone</h2>

            <div class="flex items-center justify-between py-3">
                <div>
                    <div class="font-medium">Reset Application</div>
                    <div class="text-sm text-muted-color">Clear all settings and return to setup wizard</div>
                </div>
                <Button label="Reset" icon="pi pi-trash" severity="danger" size="small" outlined @click="confirmReset" />
            </div>
        </div>

        <!-- Reset Confirmation Dialog -->
        <Dialog v-model:visible="showResetDialog" modal header="Reset Application?" :style="{ width: '400px' }">
            <p class="mb-4">This will remove all your settings, including:</p>
            <ul class="list-disc list-inside mb-4 text-muted-color">
                <li>Plex token and server configuration</li>
                <li>Discord settings</li>
                <li>All preferences</li>
            </ul>
            <p class="font-medium text-red-600 dark:text-red-400">This action cannot be undone.</p>

            <template #footer>
                <Button label="Cancel" severity="secondary" @click="showResetDialog = false" />
                <Button label="Reset Application" severity="danger" :loading="loading.reset" @click="executeReset" />
            </template>
        </Dialog>
    </div>
</template>
