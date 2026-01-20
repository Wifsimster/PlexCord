<script setup>
import { ref, watch, computed, onMounted, onUnmounted } from 'vue';
import { useSetupStore } from '@/stores/setup';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { DiscoverPlexServers, ValidatePlexConnection, SavePlexToken, SaveServerURL, StartPlexPINAuth, CheckPlexPINAuth } from '../../wailsjs/go/main/App';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import ProgressSpinner from 'primevue/progressspinner';
import Message from 'primevue/message';
import Tooltip from 'primevue/tooltip';
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
watch(() => setupStore.plexServerUrl, (newUrl, oldUrl) => {
    if (newUrl !== oldUrl && validationAttempted.value) {
        setupStore.clearValidation();
        validationAttempted.value = false;
        validationError.value = '';
    }
});

// Watch for token changes to clear validation
watch(() => setupStore.plexToken, (newToken, oldToken) => {
    if (newToken !== oldToken && validationAttempted.value) {
        setupStore.clearValidation();
        validationAttempted.value = false;
        validationError.value = '';
    }
});

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
            <h2 class="text-3xl font-bold mb-4">Connect to Your Plex Account</h2>
            <p class="text-lg mb-6">
                Authenticate with Plex using a secure PIN code - no password required!
            </p>

            <!-- PIN Authentication Section -->
            <div class="min-h-50 mb-6">
                <!-- Initial State: Show Connect Button -->
                <div v-if="authStep === 'initial'" class="auth-initial">
                    <div class="flex flex-col items-center gap-4 p-8 border-2 border-dashed border-gray-600 rounded-lg">
                        <i class="pi pi-lock text-6xl text-primary"></i>
                        <h3 class="text-xl font-semibold">Secure Authentication</h3>
                        <p class="text-center text-surface-600 dark:text-surface-400">
                            Click below to generate a PIN code, then authorize PlexCord in your browser.
                        </p>
                        <Button
                            label="Connect with Plex"
                            icon="pi pi-sign-in"
                            @click="startPINAuth"
                            size="large"
                            class="mt-2"
                        />
                    </div>
                </div>

                <!-- Loading State -->
                <div v-if="authStep === 'loading'" class="auth-loading">
                    <div class="flex flex-col items-center gap-4 p-8">
                        <ProgressSpinner style="width: 50px; height: 50px" />
                        <p>Generating PIN code...</p>
                    </div>
                </div>

                <!-- Waiting for Authorization -->
                <div v-if="authStep === 'waiting'" class="auth-waiting">
                    <div class="flex flex-col items-center gap-4 p-8 bg-surface-100 dark:bg-surface-800 rounded-lg">
                        <div class="flex items-center gap-4">
                            <ProgressSpinner style="width: 30px; height: 30px" />
                            <div>
                                <h3 class="text-2xl font-bold mb-2">Waiting for authorization...</h3>
                                <p class="text-surface-600 dark:text-surface-400">A browser window has been opened. Please authorize PlexCord.</p>
                            </div>
                        </div>
                        
                        <div class="pin-display">
                            <div class="text-sm text-surface-600 dark:text-surface-400 mb-2">Your PIN Code:</div>
                            <div class="text-6xl font-bold tracking-widest text-primary-500">{{ pinCode }}</div>
                            <div class="text-sm text-surface-600 dark:text-surface-400 mt-2">Enter this code at plex.tv/link</div>
                        </div>

                        <div class="flex gap-3 mt-4">
                            <Button
                                label="Open Browser Again"
                                icon="pi pi-external-link"
                                @click="BrowserOpenURL(authURL)"
                                outlined
                            />
                            <Button
                                label="Cancel"
                                icon="pi pi-times"
                                @click="cancelPINAuth"
                                severity="secondary"
                                outlined
                            />
                        </div>
                    </div>
                </div>

                <!-- Success State -->
                <div v-if="authStep === 'success'" class="auth-success">
                    <Message severity="success" :closable="false">
                        <div class="flex items-center gap-3">
                            <i class="pi pi-check-circle text-3xl"></i>
                            <div>
                                <div class="font-semibold">Successfully authenticated!</div>
                                <div class="text-sm">You can now discover your Plex servers below.</div>
                            </div>
                        </div>
                    </Message>
                </div>

                <!-- Error State -->
                <div v-if="authStep === 'error'" class="auth-error">
                    <Message severity="error" :closable="false">
                        <div class="flex flex-col gap-3">
                            <div class="flex items-center gap-3">
                                <i class="pi pi-exclamation-circle text-2xl"></i>
                                <div class="font-semibold">Authentication failed</div>
                            </div>
                            <div class="text-sm">{{ authError }}</div>
                            <Button
                                label="Try Again"
                                icon="pi pi-refresh"
                                @click="startPINAuth"
                                size="small"
                            />
                        </div>
                    </Message>
                </div>
            </div>

            <!-- Server Discovery Section -->
            <div class="mt-8">
                <div class="mb-6">
                    <h3 class="text-xl font-semibold mb-2">{{ showManualEntry ? 'Enter Server Manually' : 'Discover Plex Servers' }}</h3>
                    <p class="text-sm text-surface-600 dark:text-surface-400">
                        {{ showManualEntry ? 'Enter your Plex server URL directly' : 'Automatically find Plex servers on your local network' }}
                    </p>
                </div>

                <!-- Manual Entry Form -->
                <div v-if="showManualEntry" class="manual-entry-form">
                    <div class="manual-entry-input-wrapper">
                        <label for="manual-server-url" class="block text-sm font-medium mb-2">
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
                        <small v-if="manualEntryError && manualServerUrl.trim().length > 0" class="error-text">
                            <i class="pi pi-exclamation-circle mr-1"></i>
                            {{ manualEntryError }}
                        </small>

                        <!-- Success Message -->
                        <small v-if="isManualUrlValid && !manualEntryError" class="helper-text text-green-600">
                            <i class="pi pi-check-circle mr-1"></i>
                            Valid URL format
                        </small>

                        <!-- Helper Text -->
                        <div class="manual-entry-help mt-4">
                            <p class="text-sm text-surface-600 dark:text-surface-400 mb-2">
                                <i class="pi pi-info-circle mr-1"></i>
                                <strong>Examples of valid URLs:</strong>
                            </p>
                            <ul class="example-list">
                                <li><code>http://192.168.1.100:32400</code> - Local server with IP</li>
                                <li><code>http://plex.local:32400</code> - Local server with hostname</li>
                                <li><code>https://plex.example.com:32400</code> - Remote server with HTTPS</li>
                            </ul>
                            <p class="text-sm text-surface-600 dark:text-surface-400 mt-3">
                                <i class="pi pi-lightbulb mr-1"></i>
                                <strong>Tip:</strong> The default Plex port is <code>32400</code>. Use HTTPS for remote connections.
                            </p>
                        </div>
                    </div>

                    <!-- Action Buttons -->
                    <div class="manual-entry-actions mt-4">
                        <Button
                            label="Use Discovery Instead"
                            icon="pi pi-search"
                            @click="useDiscoveryInstead"
                            severity="secondary"
                            outlined
                            class="w-full"
                        />
                    </div>
                </div>

                <!-- Discovery UI (shown when NOT in manual entry mode) -->
                <div v-if="!showManualEntry">
                    <!-- Discover Button with Tooltip -->
                    <div class="discovery-button-wrapper">
                        <Button
                            v-if="!hasDiscovered && !isDiscovering"
                            label="Discover Servers"
                            icon="pi pi-search"
                            @click="discoverServers"
                            :disabled="!setupStore.isPlexStepValid"
                            class="w-full"
                        />
                        <i 
                            v-if="!hasDiscovered && !isDiscovering"
                            class="pi pi-question-circle ml-2 text-surface-600 dark:text-surface-400 cursor-help"
                            v-tooltip.right="discoveryTooltip"
                            style="font-size: 1.2rem; vertical-align: middle;"
                        ></i>
                    </div>

                    <!-- Always show "Enter Manually" option with Tooltip -->
                    <div class="manual-entry-button-wrapper mt-3">
                        <Button
                            v-if="!hasDiscovered && !isDiscovering"
                            label="Enter Manually"
                            icon="pi pi-pencil"
                            @click="enterManually"
                            severity="info"
                            outlined
                            class="w-full"
                        />
                        <i 
                            v-if="!hasDiscovered && !isDiscovering"
                            class="pi pi-question-circle ml-2 text-surface-600 dark:text-surface-400 cursor-help"
                            v-tooltip.right="manualEntryTooltip"
                            style="font-size: 1.2rem; vertical-align: middle;"
                        ></i>
                    </div>

                <!-- Loading State -->
                <div v-if="isDiscovering" class="discovery-loading">
                    <ProgressSpinner
                        style="width: 50px; height: 50px"
                        strokeWidth="4"
                        fill="transparent"
                        animationDuration="1s"
                    />
                    <p class="text-surface-600 dark:text-surface-400 mt-3">
                        Searching for Plex servers on your network...
                    </p>
                </div>

                <!-- Discovery Error -->
                <div v-if="discoveryError" class="discovery-error mt-4">
                    <i class="pi pi-exclamation-triangle mr-2"></i>
                    <span>{{ discoveryError }}</span>
                </div>

                <!-- Discovered Servers -->
                <div v-if="hasDiscovered && setupStore.discoveredServers.length > 0" class="discovered-servers mt-6">
                    <p class="text-sm text-surface-600 dark:text-surface-400 mb-4">
                        Found {{ setupStore.discoveredServers.length }} server(s). Select one to continue:
                    </p>
                    <div class="server-list space-y-3">
                        <ServerCard
                            v-for="server in setupStore.discoveredServers"
                            :key="server.id"
                            :server="server"
                            :is-selected="setupStore.selectedServer?.id === server.id"
                            @server-selected="handleServerSelected"
                        />
                    </div>
                    <div class="discovery-actions mt-6">
                        <Button
                            label="Search Again"
                            icon="pi pi-refresh"
                            @click="discoverServers"
                            severity="secondary"
                            outlined
                        />
                    </div>
                </div>

                <!-- No Servers Found -->
                <div v-if="hasDiscovered && setupStore.discoveredServers.length === 0" class="no-servers-found mt-4">
                    <div class="no-servers-message">
                        <i class="pi pi-info-circle text-3xl mb-3"></i>
                        <h4 class="font-semibold mb-2">No Servers Found</h4>
                        <p class="text-sm text-surface-600 dark:text-surface-400 mb-4">
                            We couldn't find any Plex servers on your local network.
                            Make sure your Plex Media Server is running and connected to the same network.
                        </p>
                        <div class="no-servers-actions">
                            <Button
                                label="Search Again"
                                icon="pi pi-refresh"
                                @click="discoverServers"
                                severity="secondary"
                                outlined
                                class="mr-2"
                            />
                            <Button
                                label="Enter Manually"
                                icon="pi pi-pencil"
                                @click="enterManually"
                                severity="info"
                                outlined
                            />
                        </div>
                    </div>
                </div>

                    <!-- Helper Text -->
                    <small v-if="!setupStore.isPlexStepValid && !hasDiscovered" class="helper-text text-surface-600 dark:text-surface-400 mt-3">
                        <i class="pi pi-info-circle mr-1"></i>
                        Enter your Plex token above to enable server discovery
                    </small>
                </div>
            </div>

            <!-- Connection Validation Section -->
            <div v-if="hasServerUrl" class="bg-surface-50 dark:bg-surface-900 p-6 rounded-lg border border-surface-200 dark:border-surface-700 mt-8">
                <div class="mb-6">
                    <h3 class="text-xl font-semibold mb-2">Verify Connection</h3>
                    <p class="text-sm text-surface-600 dark:text-surface-400">
                        Test the connection to your Plex server before continuing
                    </p>
                </div>

                <!-- Validation Button (shown when not validated and not validating) -->
                <div v-if="!setupStore.isConnectionValidated && !isValidating && !validationError" class="validation-trigger">
                    <Button
                        label="Validate Connection"
                        icon="pi pi-check-circle"
                        @click="validateConnection"
                        :disabled="!canValidate"
                        class="w-full"
                    />
                    <small class="helper-text text-surface-600 dark:text-surface-400 mt-3">
                        <i class="pi pi-info-circle mr-1"></i>
                        This will verify your token and server are working correctly
                    </small>
                </div>

                <!-- Validation Loading State -->
                <div v-if="isValidating" class="validation-loading">
                    <ProgressSpinner
                        style="width: 50px; height: 50px"
                        strokeWidth="4"
                        fill="transparent"
                        animationDuration="1s"
                    />
                    <p class="text-surface-600 dark:text-surface-400 mt-3">
                        Validating connection...
                    </p>
                </div>

                <!-- Validation Success -->
                <div v-if="setupStore.isConnectionValidated && setupStore.validationResult" class="validation-success">
                    <Message severity="success" :closable="false" class="w-full">
                        <template #icon>
                            <i class="pi pi-check-circle text-2xl"></i>
                        </template>
                        <div class="success-content">
                            <h4 class="font-semibold text-lg mb-4">
                                Connected to Plex Media Server
                            </h4>
                            <div class="server-details flex flex-col gap-3">
                                <span class="detail-item flex items-center gap-3 text-sm">
                                    <i class="pi pi-server text-base"></i>
                                    <span>Version {{ setupStore.validationResult.serverVersion }}</span>
                                </span>
                                <span class="detail-item flex items-center gap-3 text-sm">
                                    <i class="pi pi-folder text-base"></i>
                                    <span>{{ setupStore.validationResult.libraryCount }} libraries found</span>
                                </span>
                            </div>
                        </div>
                    </Message>
                    <div class="validation-actions mt-6">
                        <Button
                            label="Re-validate"
                            icon="pi pi-refresh"
                            @click="retryValidation"
                            severity="secondary"
                            outlined
                            size="small"
                        />
                    </div>
                </div>

                <!-- Validation Error -->
                <div v-if="validationError && !isValidating" class="validation-error">
                    <Message severity="error" :closable="false" class="w-full">
                        <template #icon>
                            <i class="pi pi-times-circle text-2xl"></i>
                        </template>
                        <div class="error-content">
                            <h4 class="font-semibold mb-2">Connection Failed</h4>
                            <p class="text-sm">{{ validationError }}</p>
                        </div>
                    </Message>
                    <div class="validation-actions mt-4">
                        <Button
                            label="Retry"
                            icon="pi pi-refresh"
                            @click="retryValidation"
                            severity="danger"
                            class="mr-2"
                        />
                        <Button
                            label="Modify Settings"
                            icon="pi pi-pencil"
                            @click="showManualEntry = true; setupStore.toggleManualEntry(true);"
                            severity="secondary"
                            outlined
                        />
                    </div>
                    <small class="helper-text text-surface-600 dark:text-surface-400 mt-3">
                        <i class="pi pi-lightbulb mr-1"></i>
                        Check that your Plex server is running and the token is correct
                    </small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(-10px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.auth-initial,
.auth-waiting,
.auth-success,
.auth-error {
    animation: fadeIn 0.3s ease-in;
}
</style>
