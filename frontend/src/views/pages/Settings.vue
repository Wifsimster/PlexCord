<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useSetupStore } from '@/stores/setup';
import Button from 'primevue/button';
import InputNumber from 'primevue/inputnumber';
import InputSwitch from 'primevue/inputswitch';
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
    reset: false,
});

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
            <Button
                icon="pi pi-arrow-left"
                severity="secondary"
                text
                rounded
                @click="goToDashboard"
            />
            <div>
                <h1 class="text-2xl font-bold text-surface-900 dark:text-surface-0">Settings</h1>
                <p class="text-muted-color">Configure PlexCord behavior</p>
            </div>
        </div>

        <!-- Connection Settings -->
        <div class="card mb-4">
            <h2 class="text-xl font-semibold mb-4">Connection</h2>

            <!-- Polling Interval -->
            <div class="flex items-center justify-between py-3 border-b border-surface-200 dark:border-surface-700">
                <div>
                    <div class="font-medium">Polling Interval</div>
                    <div class="text-sm text-muted-color">How often to check for playback changes (1-60 seconds)</div>
                </div>
                <div class="flex items-center gap-2">
                    <InputNumber
                        v-model="pollingInterval"
                        :min="1"
                        :max="60"
                        suffix=" sec"
                        :disabled="loading.polling"
                        class="w-24"
                    />
                    <Button
                        label="Save"
                        size="small"
                        :loading="loading.polling"
                        @click="updatePollingInterval"
                    />
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
                    <Button
                        v-if="!isUsingDefaultClientId"
                        label="Reset to Default"
                        severity="secondary"
                        size="small"
                        @click="resetToDefaultClientId"
                    />
                </div>
                <div class="flex gap-2">
                    <InputText
                        v-model="discordClientId"
                        :placeholder="defaultClientId"
                        class="flex-grow"
                        :disabled="loading.clientId"
                    />
                    <Button
                        label="Save"
                        :loading="loading.clientId"
                        @click="saveDiscordClientId"
                    />
                </div>
            </div>
        </div>

        <!-- Behavior Settings -->
        <div class="card mb-4">
            <h2 class="text-xl font-semibold mb-4">Behavior</h2>

            <!-- Auto-start -->
            <div class="flex items-center justify-between py-3 border-b border-surface-200 dark:border-surface-700">
                <div>
                    <div class="font-medium">Start on Login</div>
                    <div class="text-sm text-muted-color">Automatically launch PlexCord when you log in</div>
                </div>
                <InputSwitch
                    :modelValue="autoStart"
                    @update:modelValue="updateAutoStart"
                    :disabled="loading.autoStart"
                />
            </div>

            <!-- Minimize to Tray -->
            <div class="flex items-center justify-between py-3">
                <div>
                    <div class="font-medium">Minimize to Tray</div>
                    <div class="text-sm text-muted-color">Keep running in system tray when window is closed</div>
                </div>
                <InputSwitch
                    :modelValue="minimizeToTray"
                    @update:modelValue="updateMinimizeToTray"
                    :disabled="loading.minimizeToTray"
                />
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
                        <span v-if="version?.commit && version.commit !== 'unknown'" class="ml-1">
                            ({{ version.commit.substring(0, 7) }})
                        </span>
                    </div>
                </div>
                <div class="flex gap-2">
                    <Button
                        label="Check for Updates"
                        icon="pi pi-refresh"
                        severity="secondary"
                        size="small"
                        :loading="checkingUpdate"
                        @click="checkForUpdates"
                    />
                </div>
            </div>

            <!-- Update available banner -->
            <div
                v-if="updateInfo?.available"
                class="mt-3 p-3 rounded-lg bg-blue-100 dark:bg-blue-900/20 flex items-center justify-between"
            >
                <div>
                    <div class="font-medium text-blue-700 dark:text-blue-400">
                        Update Available: {{ updateInfo.latestVersion }}
                    </div>
                    <div class="text-sm text-blue-600 dark:text-blue-500">
                        {{ updateInfo.releaseNotes?.substring(0, 100) }}{{ updateInfo.releaseNotes?.length > 100 ? '...' : '' }}
                    </div>
                </div>
                <Button
                    label="Download"
                    icon="pi pi-download"
                    size="small"
                    @click="openUpdatePage"
                />
            </div>

            <!-- Changelog link -->
            <div class="flex items-center justify-between py-3">
                <div>
                    <div class="font-medium">Changelog</div>
                    <div class="text-sm text-muted-color">View release notes and version history</div>
                </div>
                <Button
                    label="View Changelog"
                    icon="pi pi-external-link"
                    severity="secondary"
                    size="small"
                    @click="OpenReleasesPage"
                />
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
                <Button
                    label="Reset"
                    icon="pi pi-trash"
                    severity="danger"
                    size="small"
                    outlined
                    @click="confirmReset"
                />
            </div>
        </div>

        <!-- Reset Confirmation Dialog -->
        <Dialog
            v-model:visible="showResetDialog"
            modal
            header="Reset Application?"
            :style="{ width: '400px' }"
        >
            <p class="mb-4">
                This will remove all your settings, including:
            </p>
            <ul class="list-disc list-inside mb-4 text-muted-color">
                <li>Plex token and server configuration</li>
                <li>Discord settings</li>
                <li>All preferences</li>
            </ul>
            <p class="font-medium text-red-600 dark:text-red-400">
                This action cannot be undone.
            </p>

            <template #footer>
                <Button
                    label="Cancel"
                    severity="secondary"
                    @click="showResetDialog = false"
                />
                <Button
                    label="Reset Application"
                    severity="danger"
                    :loading="loading.reset"
                    @click="executeReset"
                />
            </template>
        </Dialog>
    </div>
</template>
