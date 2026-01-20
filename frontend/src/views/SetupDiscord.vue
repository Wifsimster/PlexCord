<script setup>
import { ref, computed, watch } from 'vue';
import { useSetupStore } from '@/stores/setup';
import { 
    ConnectDiscord, 
    DisconnectDiscord,
    IsDiscordConnected,
    GetDefaultDiscordClientID,
    GetDiscordClientID,
    SaveDiscordClientID,
    ValidateDiscordClientID,
    TestDiscordPresence
} from '../../wailsjs/go/main/App';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import Message from 'primevue/message';
import ProgressSpinner from 'primevue/progressspinner';
import Tooltip from 'primevue/tooltip';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

const setupStore = useSetupStore();

// Connection state
const connectionState = ref('initial'); // 'initial', 'connecting', 'connected', 'error'
const connectionError = ref('');
const isConnecting = ref(false);

// Test presence state
const isTesting = ref(false);
const testSuccess = ref(false);
const testError = ref('');

// Client ID state
const showCustomClientId = ref(false);
const customClientId = ref('');
const defaultClientId = ref('');
const clientIdError = ref('');
const isClientIdValid = ref(true);

// Instructions state
const showInstructions = ref(false);

// Load default client ID on mount
const loadDefaultClientId = async () => {
    try {
        defaultClientId.value = await GetDefaultDiscordClientID();
        const currentId = await GetDiscordClientID();
        if (currentId !== defaultClientId.value) {
            customClientId.value = currentId;
            showCustomClientId.value = true;
        }
    } catch (error) {
        console.error('Failed to load Discord client ID:', error);
    }
};
loadDefaultClientId();

// Computed: Get the client ID to use for connection
const activeClientId = computed(() => {
    if (showCustomClientId.value && customClientId.value.trim()) {
        return customClientId.value.trim();
    }
    return defaultClientId.value;
});

// Validate custom client ID format
const validateClientId = async () => {
    if (!showCustomClientId.value || !customClientId.value.trim()) {
        isClientIdValid.value = true;
        clientIdError.value = '';
        return true;
    }

    try {
        await ValidateDiscordClientID(customClientId.value.trim());
        isClientIdValid.value = true;
        clientIdError.value = '';
        return true;
    } catch (error) {
        isClientIdValid.value = false;
        clientIdError.value = error?.message || 'Invalid Client ID format';
        return false;
    }
};

// Handle custom client ID change
const handleClientIdChange = async () => {
    await validateClientId();
};

// Save custom client ID
const saveCustomClientId = async () => {
    if (!showCustomClientId.value) {
        // Reset to default
        try {
            await SaveDiscordClientID('');
            return true;
        } catch (error) {
            console.error('Failed to save Discord client ID:', error);
            return false;
        }
    }

    const isValid = await validateClientId();
    if (!isValid) {
        return false;
    }

    try {
        await SaveDiscordClientID(customClientId.value.trim());
        return true;
    } catch (error) {
        console.error('Failed to save Discord client ID:', error);
        clientIdError.value = error?.message || 'Failed to save Client ID';
        return false;
    }
};

// Connect to Discord
const connectToDiscord = async () => {
    // Save client ID first if using custom
    const saved = await saveCustomClientId();
    if (!saved && showCustomClientId.value) {
        return;
    }

    isConnecting.value = true;
    connectionState.value = 'connecting';
    connectionError.value = '';

    try {
        await ConnectDiscord(activeClientId.value);
        
        // Verify connection
        const isConnected = await IsDiscordConnected();
        if (isConnected) {
            connectionState.value = 'connected';
            connectionError.value = '';
        } else {
            throw new Error('Connection verification failed');
        }
    } catch (error) {
        console.error('Discord connection failed:', error);
        connectionState.value = 'error';
        
        // Parse error message for user-friendly display
        let errorMessage = 'Failed to connect to Discord';
        if (error && typeof error === 'string') {
            errorMessage = error;
        } else if (error && error.message) {
            errorMessage = error.message;
        }

        // Provide specific guidance based on error
        if (errorMessage.includes('not running')) {
            connectionError.value = 'Discord is not running. Please start Discord and try again.';
        } else if (errorMessage.includes('invalid') || errorMessage.includes('Client ID')) {
            connectionError.value = 'Invalid Discord Client ID. Please check your configuration.';
        } else {
            connectionError.value = errorMessage;
        }
    } finally {
        isConnecting.value = false;
    }
};

// Retry connection
const retryConnection = () => {
    connectionError.value = '';
    connectToDiscord();
};

// Disconnect from Discord
const disconnect = async () => {
    try {
        await DisconnectDiscord();
        connectionState.value = 'initial';
        connectionError.value = '';
        testSuccess.value = false;
        testError.value = '';
    } catch (error) {
        console.error('Failed to disconnect:', error);
    }
};

// Test Discord presence
const testPresence = async () => {
    isTesting.value = true;
    testError.value = '';
    testSuccess.value = false;

    try {
        await TestDiscordPresence();
        testSuccess.value = true;
        // Auto-hide success message after 3 seconds
        setTimeout(() => {
            testSuccess.value = false;
        }, 3000);
    } catch (error) {
        console.error('Test presence failed:', error);
        testError.value = error?.message || 'Failed to send test presence';
    } finally {
        isTesting.value = false;
    }
};

// Toggle custom client ID section
const toggleCustomClientId = () => {
    showCustomClientId.value = !showCustomClientId.value;
    if (!showCustomClientId.value) {
        customClientId.value = '';
        clientIdError.value = '';
        isClientIdValid.value = true;
    }
};

// Open Discord Developer Portal
const openDeveloperPortal = () => {
    BrowserOpenURL('https://discord.com/developers/applications');
};

// Watch for custom client ID toggle to reset validation
watch(showCustomClientId, () => {
    if (connectionState.value === 'error' || connectionState.value === 'connected') {
        connectionState.value = 'initial';
        connectionError.value = '';
    }
});
</script>

<template>
    <div class="max-w-4xl mx-auto">
        <div class="py-8">
            <h2 class="text-3xl font-bold mb-4">Connect to Discord</h2>
            <p class="text-lg mb-6">
                PlexCord will display your Plex music activity on Discord using Rich Presence.
            </p>

            <!-- Discord Client Detection Notice -->
            <Message v-if="connectionState === 'initial'" severity="info" :closable="false" class="mb-6">
                <div class="flex items-center gap-3">
                    <i class="pi pi-discord text-xl"></i>
                    <span>Make sure Discord is running on your computer before connecting.</span>
                </div>
            </Message>

            <!-- Connection Status Section -->
            <div class="connection-section mb-6">
                <!-- Initial State: Show Connect Button -->
                <div v-if="connectionState === 'initial'" class="connection-initial">
                    <div class="flex flex-col items-center gap-4 p-8 border-2 border-dashed border-gray-600 rounded-lg">
                        <h3 class="text-xl font-semibold">Discord Rich Presence</h3>
                        <p class="text-center text-muted-color">
                            Connect to Discord to show your Plex music activity on your profile.
                        </p>
                        <Button
                            label="Connect to Discord"
                            icon="pi pi-link"
                            @click="connectToDiscord"
                            size="large"
                            class="mt-2"
                            :disabled="!isClientIdValid"
                        />
                    </div>
                </div>

                <!-- Connecting State -->
                <div v-if="connectionState === 'connecting'" class="connection-loading">
                    <div class="flex flex-col items-center gap-4 p-8 bg-surface-100 dark:bg-surface-800 rounded-lg">
                        <ProgressSpinner style="width: 50px; height: 50px" />
                        <h3 class="text-xl font-semibold">Connecting to Discord...</h3>
                        <p class="text-muted-color">Establishing connection with your Discord client</p>
                    </div>
                </div>

                <!-- Connected State -->
                <div v-if="connectionState === 'connected'" class="connection-success">
                    <div class="flex flex-col items-center gap-4 p-8 bg-green-50 dark:bg-green-900/20 border-2 border-green-500 rounded-lg">
                        <i class="pi pi-check-circle text-6xl text-green-500"></i>
                        <h3 class="text-xl font-semibold text-green-600 dark:text-green-400">Connected to Discord!</h3>
                        <p class="text-center text-muted-color">
                            PlexCord is now connected and will update your Discord status when you play music.
                        </p>
                        
                        <!-- Test Success Message -->
                        <Message v-if="testSuccess" severity="success" :closable="true" class="w-full">
                            Test presence sent successfully! Check your Discord profile.
                        </Message>
                        
                        <!-- Test Error Message -->
                        <Message v-if="testError" severity="error" :closable="true" class="w-full" @close="testError = ''">
                            {{ testError }}
                        </Message>
                        
                        <div class="flex gap-3 flex-wrap justify-center">
                            <Button
                                label="Send Test Presence"
                                icon="pi pi-send"
                                @click="testPresence"
                                :loading="isTesting"
                                severity="info"
                            />
                            <Button
                                label="Test Again"
                                icon="pi pi-refresh"
                                @click="retryConnection"
                                outlined
                                severity="secondary"
                            />
                            <Button
                                label="Disconnect"
                                icon="pi pi-times"
                                @click="disconnect"
                                outlined
                                severity="danger"
                            />
                        </div>
                    </div>
                </div>

                <!-- Error State -->
                <div v-if="connectionState === 'error'" class="connection-error">
                    <Message severity="error" :closable="false" class="mb-4">
                        <div class="flex flex-col gap-2">
                            <div class="flex items-center gap-2">
                                <i class="pi pi-times-circle"></i>
                                <span class="font-semibold">Connection Failed</span>
                            </div>
                            <p>{{ connectionError }}</p>
                        </div>
                    </Message>
                    <div class="flex flex-col gap-3">
                        <Button
                            label="Retry Connection"
                            icon="pi pi-refresh"
                            @click="retryConnection"
                            :loading="isConnecting"
                        />
                        <Button
                            label="Back to Setup"
                            icon="pi pi-arrow-left"
                            @click="disconnect"
                            outlined
                            severity="secondary"
                        />
                    </div>
                </div>
            </div>

            <!-- Advanced Configuration -->
            <div class="border-t border-surface-200 dark:border-surface-700 pt-8">
                <div class="flex items-center justify-between mb-4">
                    <h3 class="text-lg font-semibold">Advanced Configuration</h3>
                    <Button
                        :icon="showCustomClientId ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
                        @click="toggleCustomClientId"
                        text
                        size="small"
                        :label="showCustomClientId ? 'Hide' : 'Show'"
                    />
                </div>

                <div v-if="showCustomClientId" class="custom-client-id-section p-4 bg-surface-50 dark:bg-surface-900 rounded-lg">
                    <div class="mb-4">
                        <label class="block mb-2 font-medium">Custom Discord Application Client ID</label>
                        <p class="text-sm text-muted-color mb-3">
                            Use your own Discord application for custom branding or testing. Leave empty to use the default PlexCord application.
                        </p>
                        <div class="flex flex-col gap-2">
                            <InputText
                                v-model="customClientId"
                                placeholder="Enter Discord Application Client ID (17+ digits)"
                                @blur="handleClientIdChange"
                                @input="handleClientIdChange"
                                :class="{ 'p-invalid': !isClientIdValid }"
                                class="w-full"
                            />
                            <small v-if="clientIdError" class="text-red-500">{{ clientIdError }}</small>
                            <small v-else class="text-muted-color">
                                Default: {{ defaultClientId }}
                            </small>
                        </div>
                    </div>

                    <!-- Instructions Toggle -->
                    <div class="instructions-toggle">
                        <Button
                            label="How to create a Discord application"
                            icon="pi pi-question-circle"
                            @click="showInstructions = !showInstructions"
                            text
                            size="small"
                        />
                    </div>

                    <!-- Instructions Panel -->
                    <div v-if="showInstructions" class="instructions-panel mt-4 p-4 bg-surface-100 dark:bg-surface-800 rounded-lg">
                        <h4 class="font-semibold mb-3">Creating a Custom Discord Application:</h4>
                        <ol class="list-decimal list-inside space-y-2 text-sm">
                            <li>Go to the <a href="#" @click.prevent="openDeveloperPortal" class="text-primary hover:underline">Discord Developer Portal</a></li>
                            <li>Click "New Application" and give it a name</li>
                            <li>Copy the "Application ID" from the General Information page</li>
                            <li>Paste it in the field above</li>
                            <li>(Optional) Upload custom images in the Rich Presence Art Assets section</li>
                        </ol>
                        <Message severity="warn" :closable="false" class="mt-4">
                            <p class="text-sm">
                                Custom applications require additional setup. Most users should use the default PlexCord application.
                            </p>
                        </Message>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Tailwind animations for custom transitions */
@keyframes slideDown {
    from {
        opacity: 0;
        transform: translateY(-10px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.custom-client-id-section,
.instructions-panel {
    animation: slideDown 0.3s ease-out;
}
</style>
