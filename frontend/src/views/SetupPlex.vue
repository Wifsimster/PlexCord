<script setup>
import { ref, watch, computed, onMounted, onUnmounted } from 'vue';
import { useSetupStore } from '@/stores/setup';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { DiscoverPlexServers, ValidatePlexConnection, SavePlexToken, SaveServerURL, StartPlexPINAuth, CheckPlexPINAuth } from '../../wailsjs/go/main/App';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import ProgressSpinner from 'primevue/progressspinner';
import Message from 'primevue/message';
import ServerCard from '@/components/ServerCard.vue';

const setupStore = useSetupStore();

// PIN authentication state
const authStep = ref('initial'); // 'initial', 'waiting', 'success', 'error'
const pinCode = ref('');
const pinID = ref(null);
const authURL = ref('');
const authError = ref('');
const checkInterval = ref(null);

// Server discovery state
const isDiscovering = ref(false);
const discoveryError = ref('');
const hasDiscovered = ref(false);

// Manual server entry state
const showManualEntry = ref(setupStore.isManualEntry);
const manualServerUrl = ref(setupStore.isManualEntry ? setupStore.plexServerUrl : '');
const manualEntryError = ref('');
const isManualUrlValid = ref(false);

// Validation state
const isValidating = ref(false);
const validationError = ref('');
const validationAttempted = ref(false);

// Computed: Check if server URL is available for validation
const hasServerUrl = computed(() => {
    return setupStore.plexServerUrl && setupStore.plexServerUrl.trim().length > 0;
});

// Tooltip content
const discoveryTooltip = `Auto-discovery finds Plex servers on your LOCAL network only.

✓ Works for: Localhost servers, LAN servers with native installation
✗ Won't work for: Docker containers (bridge mode), remote servers, different subnets, blocked by firewall

If discovery fails, use "Enter Manually" instead.`;

const manualEntryTooltip = `Use manual entry when:

• Plex is in Docker container (e.g., http://192.168.0.237:32400)
• Remote/cloud server (e.g., https://plex.example.com:32400)
• Server on different subnet/VLAN
• Firewall blocks auto-discovery
• Localhost server (e.g., http://localhost:32400)`;

// Computed: Check if ready to validate (has token and server URL)
const canValidate = computed(() => {
    return setupStore.isPlexStepValid && hasServerUrl.value;
});

// Start PIN authentication
const startPINAuth = async () => {
    authError.value = '';
    authStep.value = 'loading';

    try {
        const result = await StartPlexPINAuth();
        pinCode.value = result.pinCode;
        pinID.value = result.pinID;
        authURL.value = result.authURL;
        authStep.value = 'waiting';

        // Start polling for authorization
        checkInterval.value = setInterval(checkPINStatus, 2000);

        // Open browser automatically
        BrowserOpenURL(authURL.value);
    } catch (error) {
        console.error('Failed to start PIN auth:', error);
        authError.value = error?.message || 'Failed to start authentication';
        authStep.value = 'error';
    }
};

// Check PIN status
const checkPINStatus = async () => {
    try {
        const result = await CheckPlexPINAuth(pinID.value);

        if (result.authorized) {
            // Success! Save the token
            setupStore.setPlexToken(result.authToken);
            authStep.value = 'success';

            // Stop polling
            if (checkInterval.value) {
                clearInterval(checkInterval.value);
                checkInterval.value = null;
            }
        } else if (result.expired) {
            authError.value = 'PIN expired. Please try again.';
            authStep.value = 'error';

            // Stop polling
            if (checkInterval.value) {
                clearInterval(checkInterval.value);
                checkInterval.value = null;
            }
        }
    } catch (error) {
        console.error('Error checking PIN status:', error);
    }
};

// Cancel PIN auth
const cancelPINAuth = () => {
    if (checkInterval.value) {
        clearInterval(checkInterval.value);
        checkInterval.value = null;
    }
    authStep.value = 'initial';
    pinCode.value = '';
    pinID.value = null;
    authURL.value = '';
    authError.value = '';
};

// Cleanup on unmount
onUnmounted(() => {
    if (checkInterval.value) {
        clearInterval(checkInterval.value);
    }
});

// Discover Plex servers
const discoverServers = async () => {
    isDiscovering.value = true;
    discoveryError.value = '';
    hasDiscovered.value = false;

    try {
        const servers = await DiscoverPlexServers();
        setupStore.setDiscoveredServers(servers);
        hasDiscovered.value = true;
    } catch (error) {
        console.error('Failed to discover servers:', error);
        discoveryError.value = 'Failed to discover servers. Please try again or enter manually.';
        setupStore.setDiscoveredServers([]);
    } finally {
        isDiscovering.value = false;
    }
};

// Handle server selection
const handleServerSelected = (server) => {
    setupStore.selectServer(server);
};

// Enter server manually
const enterManually = () => {
    // Switch to manual entry mode
    showManualEntry.value = true;
    setupStore.toggleManualEntry(true);
    hasDiscovered.value = false;
    manualEntryError.value = '';
};

// Switch back to discovery
const useDiscoveryInstead = () => {
    showManualEntry.value = false;
    setupStore.toggleManualEntry(false);
    manualServerUrl.value = '';
    manualEntryError.value = '';
    isManualUrlValid.value = false;
};

// Validate Plex server URL format
const validatePlexServerUrl = (url) => {
    if (!url || url.trim().length === 0) {
        return { valid: false, error: 'Server URL is required' };
    }

    const trimmedUrl = url.trim();

    // Check for protocol
    if (!trimmedUrl.startsWith('http://') && !trimmedUrl.startsWith('https://')) {
        return { valid: false, error: 'URL must start with http:// or https://' };
    }

    // Parse URL to validate structure
    try {
        const urlObj = new URL(trimmedUrl);

        // Validate port if present
        if (urlObj.port) {
            const portNum = parseInt(urlObj.port, 10);
            if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
                return { valid: false, error: 'Port must be between 1 and 65535' };
            }
        }

        // Validate hostname (IP or domain)
        if (!urlObj.hostname || urlObj.hostname.length === 0) {
            return { valid: false, error: 'Invalid hostname or IP address' };
        }

        return { valid: true, error: '' };
    } catch (e) {
        return { valid: false, error: 'Invalid URL format' };
    }
};

// Handle manual URL input change
const handleManualUrlChange = () => {
    const validation = validatePlexServerUrl(manualServerUrl.value);
    isManualUrlValid.value = validation.valid;
    manualEntryError.value = validation.error;

    if (validation.valid) {
        setupStore.setManualServerUrl(manualServerUrl.value.trim());
    }
};

// Handle manual URL blur (validate on blur)
const handleManualUrlBlur = () => {
    if (manualServerUrl.value.trim().length > 0) {
        handleManualUrlChange();
    }
};

// Validate Plex connection
const validateConnection = async () => {
    if (!canValidate.value) {
        return;
    }

    isValidating.value = true;
    validationError.value = '';
    validationAttempted.value = true;
    setupStore.clearValidation();

    try {
        // First save the token to keychain if not already saved
        await SavePlexToken(setupStore.plexToken);

        // Validate the connection
        const result = await ValidatePlexConnection(setupStore.plexServerUrl);
        
        // If no error was thrown, validation was successful
        setupStore.setValidationResult(result);
        validationError.value = '';

        // Save server URL to backend config after successful validation
        await SaveServerURL(setupStore.plexServerUrl);
    } catch (error) {
        console.error('Connection validation failed:', error);

        // Extract user-friendly error message
        let errorMessage = 'Failed to connect to Plex server';
        if (error && typeof error === 'string') {
            errorMessage = error;
        } else if (error && error.message) {
            errorMessage = error.message;
        }

        validationError.value = errorMessage;
        setupStore.clearValidation();
    } finally {
        isValidating.value = false;
    }
};

// Retry validation
const retryValidation = () => {
    validationError.value = '';
    validateConnection();
};

// Watch for server URL changes to clear validation
watch(
    () => setupStore.plexServerUrl,
    (newUrl, oldUrl) => {
        if (newUrl !== oldUrl && validationAttempted.value) {
            setupStore.clearValidation();
            validationAttempted.value = false;
            validationError.value = '';
        }
    }
);

// Watch for token changes to clear validation
watch(
    () => setupStore.plexToken,
    (newToken, oldToken) => {
        if (newToken !== oldToken && validationAttempted.value) {
            setupStore.clearValidation();
            validationAttempted.value = false;
            validationError.value = '';
        }
    }
);

// Restore validation state on mount
onMounted(() => {
    if (setupStore.isConnectionValidated && setupStore.validationResult) {
        validationAttempted.value = true;
    }
});
</script>

<template>
    <div class="max-w-4xl mx-auto">
        <div class="py-8">
            <h2 class="text-3xl font-bold mb-4 text-surface-900 dark:text-surface-0">Connect to Your Plex Account</h2>
            <p class="text-lg mb-6 text-surface-600 dark:text-surface-400">Authenticate with Plex using a secure PIN code - no password required!</p>

            <!-- PIN Authentication Section -->
            <div class="mb-6" :class="{ 'min-h-50': authStep !== 'success' }">
                <!-- Initial State: Show Connect Button -->
                <div v-if="authStep === 'initial'" class="animate-fadein">
                    <div class="flex flex-col items-center gap-4 p-8 border-2 border-dashed border-surface-300 dark:border-surface-600 rounded-lg">
                        <i class="pi pi-lock text-6xl text-primary-500"></i>
                        <h3 class="text-xl font-semibold text-surface-900 dark:text-surface-0">Secure Authentication</h3>
                        <p class="text-center text-surface-600 dark:text-surface-400">Click below to generate a PIN code, then authorize PlexCord in your browser.</p>
                        <Button label="Connect with Plex" icon="pi pi-sign-in" @click="startPINAuth" size="large" class="mt-2" />
                    </div>
                </div>

                <!-- Loading State -->
                <div v-if="authStep === 'loading'" class="animate-fadein">
                    <div class="flex flex-col items-center gap-4 p-8">
                        <ProgressSpinner style="width: 50px; height: 50px" />
                        <p class="text-surface-600 dark:text-surface-400">Generating PIN code...</p>
                    </div>
                </div>

                <!-- Waiting for Authorization -->
                <div v-if="authStep === 'waiting'" class="animate-fadein">
                    <div class="flex flex-col items-center gap-4 p-8 bg-surface-100 dark:bg-surface-800 rounded-lg">
                        <div class="flex items-center gap-4">
                            <ProgressSpinner style="width: 30px; height: 30px" />
                            <div>
                                <h3 class="text-2xl font-bold mb-2 text-surface-900 dark:text-surface-0">Waiting for authorization...</h3>
                                <p class="text-surface-600 dark:text-surface-400">A browser window has been opened. Please authorize PlexCord.</p>
                            </div>
                        </div>

                        <div class="text-center py-4">
                            <div class="text-sm text-surface-600 dark:text-surface-400 mb-2">Your PIN Code:</div>
                            <div class="text-3xl font-bold text-primary-500 font-mono break-all select-all px-4">{{ pinCode }}</div>
                            <div class="text-sm text-surface-600 dark:text-surface-400 mt-2">Enter this code at plex.tv/link</div>
                        </div>

                        <div class="flex gap-3 mt-4">
                            <Button label="Open Browser Again" icon="pi pi-external-link" @click="BrowserOpenURL(authURL)" outlined />
                            <Button label="Cancel" icon="pi pi-times" @click="cancelPINAuth" severity="secondary" outlined />
                        </div>
                    </div>
                </div>

                <!-- Success State -->
                <div v-if="authStep === 'success'" class="animate-fadein">
                    <Message severity="success" :closable="false">
                        <div class="flex items-center gap-3">
                            <i class="pi pi-check-circle text-3xl"></i>
                            <div>
                                <div class="font-semibold text-lg">Successfully authenticated!</div>
                                <div class="text-sm">You can now discover your Plex servers below.</div>
                            </div>
                        </div>
                    </Message>
                </div>

                <!-- Error State -->
                <div v-if="authStep === 'error'" class="animate-fadein">
                    <Message severity="error" :closable="false">
                        <div class="flex flex-col gap-3">
                            <div class="flex items-center gap-3">
                                <i class="pi pi-exclamation-circle text-2xl"></i>
                                <div class="font-semibold text-lg">Authentication failed</div>
                            </div>
                            <div class="text-sm">{{ authError }}</div>
                            <Button label="Try Again" icon="pi pi-refresh" @click="startPINAuth" size="small" class="w-fit" />
                        </div>
                    </Message>
                </div>
            </div>

            <!-- Server Discovery Section -->
            <div class="mt-8 pt-8 border-t border-surface-200 dark:border-surface-700">
                <div class="mb-6">
                    <h3 class="text-xl font-semibold mb-2 text-surface-900 dark:text-surface-0">{{ showManualEntry ? 'Enter Server Manually' : 'Discover Plex Servers' }}</h3>
                    <p class="text-sm text-surface-600 dark:text-surface-400">
                        {{ showManualEntry ? 'Enter your Plex server URL directly' : 'Automatically find Plex servers on your local network' }}
                    </p>
                </div>

                <!-- Manual Entry Form -->
                <div v-if="showManualEntry" class="space-y-4">
                    <div class="flex flex-col gap-2">
                        <label for="manual-server-url" class="block text-sm font-medium text-surface-700 dark:text-surface-200">
                            Plex Server URL
                            <span class="text-red-500">*</span>
                        </label>
                        <InputText
                            id="manual-server-url"
                            v-model="manualServerUrl"
                            @blur="handleManualUrlBlur"
                            @input="handleManualUrlChange"
                            placeholder="e.g., http://192.168.1.100:32400"
                            class="w-full"
                            :class="{ 'p-invalid': manualEntryError && manualServerUrl.trim().length > 0 }"
                        />

                        <!-- Validation Error -->
                        <small v-if="manualEntryError && manualServerUrl.trim().length > 0" class="flex items-center text-red-500">
                            <i class="pi pi-exclamation-circle mr-1"></i>
                            {{ manualEntryError }}
                        </small>

                        <!-- Success Message -->
                        <small v-if="isManualUrlValid && !manualEntryError" class="flex items-center text-green-600 dark:text-green-400">
                            <i class="pi pi-check-circle mr-1"></i>
                            Valid URL format
                        </small>

                        <!-- Helper Text -->
                        <div class="mt-4 p-4 bg-surface-50 dark:bg-surface-800 rounded-lg text-sm text-surface-600 dark:text-surface-400">
                            <p class="mb-2 font-medium">
                                <i class="pi pi-info-circle mr-1"></i>
                                Examples of valid URLs:
                            </p>
                            <ul class="list-disc list-inside ml-2 space-y-1 mb-3">
                                <li class="font-mono text-xs">http://192.168.1.100:32400</li>
                                <li class="font-mono text-xs">http://plex.local:32400</li>
                                <li class="font-mono text-xs">https://plex.example.com:32400</li>
                            </ul>
                            <p>
                                <i class="pi pi-lightbulb mr-1 text-yellow-500"></i>
                                <strong>Tip:</strong> The default Plex port is <code class="bg-surface-200 dark:bg-surface-700 px-1 rounded">32400</code>. Use HTTPS for remote connections.
                            </p>
                        </div>
                    </div>

                    <!-- Action Buttons -->
                    <div class="pt-2">
                        <Button label="Use Discovery Instead" icon="pi pi-search" @click="useDiscoveryInstead" severity="secondary" outlined class="w-full" />
                    </div>
                </div>

                <!-- Discovery UI (shown when NOT in manual entry mode) -->
                <div v-if="!showManualEntry">
                    <!-- Discover Button with Tooltip -->
                    <div class="flex items-center gap-2">
                        <Button v-if="!hasDiscovered && !isDiscovering" label="Discover Servers" icon="pi pi-search" @click="discoverServers" :disabled="!setupStore.isPlexStepValid" class="w-full" />
                        <i v-if="!hasDiscovered && !isDiscovering" class="pi pi-question-circle text-surface-600 dark:text-surface-400 cursor-help text-xl" v-tooltip.right="discoveryTooltip"></i>
                    </div>

                    <!-- Always show "Enter Manually" option with Tooltip -->
                    <div class="flex items-center gap-2 mt-3">
                        <Button v-if="!hasDiscovered && !isDiscovering" label="Enter Manually" icon="pi pi-pencil" @click="enterManually" severity="info" outlined class="w-full" />
                        <i v-if="!hasDiscovered && !isDiscovering" class="pi pi-question-circle text-surface-600 dark:text-surface-400 cursor-help text-xl" v-tooltip.right="manualEntryTooltip"></i>
                    </div>

                    <!-- Loading State -->
                    <div v-if="isDiscovering" class="flex flex-col items-center justify-center py-8">
                        <ProgressSpinner style="width: 50px; height: 50px" strokeWidth="4" fill="transparent" animationDuration="1s" />
                        <p class="text-surface-600 dark:text-surface-400 mt-3">Searching for Plex servers on your network...</p>
                    </div>

                    <!-- Discovery Error -->
                    <div v-if="discoveryError" class="mt-4 text-red-500 flex items-center justify-center">
                        <i class="pi pi-exclamation-triangle mr-2"></i>
                        <span>{{ discoveryError }}</span>
                    </div>

                    <!-- Discovered Servers -->
                    <div v-if="hasDiscovered && setupStore.discoveredServers.length > 0" class="mt-6">
                        <p class="text-sm text-surface-600 dark:text-surface-400 mb-4">Found {{ setupStore.discoveredServers.length }} server(s). Select one to continue:</p>
                        <div class="space-y-3">
                            <ServerCard v-for="server in setupStore.discoveredServers" :key="server.id" :server="server" :is-selected="setupStore.selectedServer?.id === server.id" @server-selected="handleServerSelected" />
                        </div>
                        <div class="mt-6">
                            <Button label="Search Again" icon="pi pi-refresh" @click="discoverServers" severity="secondary" outlined class="w-full sm:w-auto" />
                        </div>
                    </div>

                    <!-- No Servers Found -->
                    <div v-if="hasDiscovered && setupStore.discoveredServers.length === 0" class="mt-4 p-6 bg-surface-50 dark:bg-surface-800 rounded-lg text-center">
                        <i class="pi pi-info-circle text-3xl mb-3 text-blue-500"></i>
                        <h4 class="font-semibold mb-2 text-surface-900 dark:text-surface-0">No Servers Found</h4>
                        <p class="text-sm text-surface-600 dark:text-surface-400 mb-4">We couldn't find any Plex servers on your local network. Make sure your Plex Media Server is running and connected to the same network.</p>
                        <div class="flex flex-col sm:flex-row gap-3 justify-center">
                            <Button label="Search Again" icon="pi pi-refresh" @click="discoverServers" severity="secondary" outlined />
                            <Button label="Enter Manually" icon="pi pi-pencil" @click="enterManually" severity="info" outlined />
                        </div>
                    </div>

                    <!-- Helper Text -->
                    <small v-if="!setupStore.isPlexStepValid && !hasDiscovered" class="flex items-center justify-center mt-4 text-surface-500">
                        <i class="pi pi-info-circle mr-1"></i>
                        Enter your Plex token above to enable server discovery
                    </small>
                </div>
            </div>

            <!-- Connection Validation Section -->
            <div v-if="hasServerUrl" class="bg-surface-50 dark:bg-surface-800 p-6 rounded-lg border border-surface-200 dark:border-surface-700 mt-8 animate-fadein">
                <div class="mb-6">
                    <h3 class="text-xl font-semibold mb-2 text-surface-900 dark:text-surface-0">Verify Connection</h3>
                    <p class="text-sm text-surface-600 dark:text-surface-400">Test the connection to your Plex server before continuing</p>
                </div>

                <!-- Validation Button (shown when not validated and not validating) -->
                <div v-if="!setupStore.isConnectionValidated && !isValidating && !validationError">
                    <Button label="Validate Connection" icon="pi pi-check-circle" @click="validateConnection" :disabled="!canValidate" class="w-full" />
                    <small class="flex items-center justify-center mt-3 text-surface-500">
                        <i class="pi pi-info-circle mr-1"></i>
                        This will verify your token and server are working correctly
                    </small>
                </div>

                <!-- Validation Loading State -->
                <div v-if="isValidating" class="flex flex-col items-center justify-center py-4">
                    <ProgressSpinner style="width: 50px; height: 50px" strokeWidth="4" fill="transparent" animationDuration="1s" />
                    <p class="text-surface-600 dark:text-surface-400 mt-3">Validating connection...</p>
                </div>

                <!-- Validation Success -->
                <div v-if="setupStore.isConnectionValidated && setupStore.validationResult" class="animate-fadein">
                    <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-xl p-6">
                        <div class="flex items-center gap-4 mb-6">
                            <div class="w-12 h-12 rounded-full bg-green-100 dark:bg-green-800 flex items-center justify-center shrink-0">
                                <i class="pi pi-check text-2xl text-green-600 dark:text-green-400"></i>
                            </div>
                            <div>
                                <h4 class="text-xl font-bold text-green-900 dark:text-green-100">Successfully Connected</h4>
                                <p class="text-green-700 dark:text-green-300">
                                    Connected to <span class="font-semibold">{{ setupStore.validationResult.serverName || 'Plex Media Server' }}</span>
                                </p>
                            </div>
                        </div>
                        
                        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <div class="bg-white dark:bg-surface-800 p-4 rounded-lg border border-surface-200 dark:border-surface-700 flex items-center gap-3 shadow-sm">
                                <div class="w-10 h-10 rounded-full bg-surface-100 dark:bg-surface-700 flex items-center justify-center shrink-0">
                                    <i class="pi pi-server text-lg text-primary-500"></i>
                                </div>
                                <div>
                                    <div class="text-xs text-surface-500 uppercase font-bold tracking-wider">Version</div>
                                    <div class="font-mono text-sm font-medium text-surface-900 dark:text-surface-0">{{ setupStore.validationResult.serverVersion || 'Unknown' }}</div>
                                </div>
                            </div>
                            
                            <div class="bg-white dark:bg-surface-800 p-4 rounded-lg border border-surface-200 dark:border-surface-700 flex items-center gap-3 shadow-sm">
                                <div class="w-10 h-10 rounded-full bg-surface-100 dark:bg-surface-700 flex items-center justify-center shrink-0">
                                    <i class="pi pi-folder text-lg text-primary-500"></i>
                                </div>
                                <div>
                                    <div class="text-xs text-surface-500 uppercase font-bold tracking-wider">Content</div>
                                    <div class="font-mono text-sm font-medium text-surface-900 dark:text-surface-0">{{ setupStore.validationResult.libraryCount }} Libraries</div>
                                </div>
                            </div>
                        </div>

                        <div class="mt-6 flex justify-center sm:justify-end">
                            <Button label="Re-validate Connection" icon="pi pi-refresh" @click="retryValidation" severity="success" text size="small" />
                        </div>
                    </div>
                </div>

                <!-- Validation Error -->
                <div v-if="validationError && !isValidating" class="animate-fadein">
                    <Message severity="error" :closable="false" class="w-full">
                        <template #icon>
                            <i class="pi pi-times-circle text-2xl"></i>
                        </template>
                        <div class="w-full">
                            <h4 class="font-semibold mb-2">Connection Failed</h4>
                            <p class="text-sm">{{ validationError }}</p>
                        </div>
                    </Message>
                    <div class="mt-4 flex flex-col sm:flex-row gap-3 justify-center">
                        <Button label="Retry" icon="pi pi-refresh" @click="retryValidation" severity="danger" />
                        <Button
                            label="Modify Settings"
                            icon="pi pi-pencil"
                            @click="
                                showManualEntry = true;
                                setupStore.toggleManualEntry(true);
                            "
                            severity="secondary"
                            outlined
                        />
                    </div>
                    <small class="flex items-center justify-center mt-3 text-surface-600 dark:text-surface-400">
                        <i class="pi pi-lightbulb mr-1 text-yellow-500"></i>
                        Check that your Plex server is running and the token is correct
                    </small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.animate-fadein {
    animation: fadein 0.3s ease-in;
}

@keyframes fadein {
    from {
        opacity: 0;
        transform: translateY(-10px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
</style>
