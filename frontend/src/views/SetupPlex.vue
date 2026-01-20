<script setup>
import { ref, watch, computed, onMounted } from 'vue';
import { useSetupStore } from '@/stores/setup';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { DiscoverPlexServers, ValidatePlexConnection, SavePlexToken } from '../../wailsjs/go/main/App';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import ProgressSpinner from 'primevue/progressspinner';
import Message from 'primevue/message';
import ServerCard from '@/components/ServerCard.vue';

const setupStore = useSetupStore();

// Token input state
const showToken = ref(false);
const tokenInput = ref(setupStore.plexToken || '');

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

// Computed: Check if ready to validate (has token and server URL)
const canValidate = computed(() => {
    return setupStore.isPlexStepValid && hasServerUrl.value;
});

// Watch for changes and update store
watch(tokenInput, (newToken) => {
    setupStore.setPlexToken(newToken);
});

// Toggle password visibility
const toggleTokenVisibility = () => {
    showToken.value = !showToken.value;
};

// Open Plex token page in external browser
const openPlexTokenPage = () => {
    BrowserOpenURL('https://www.plex.tv/claim/');
};

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
    <div class="setup-step">
        <div class="step-content">
            <h2 class="text-3xl font-bold mb-4">Connect to Your Plex Account</h2>
            <p class="text-lg mb-6">
                To connect PlexCord to your Plex Media Server, you'll need an authentication token.
            </p>

            <!-- Security Note -->
            <div class="security-note mb-6">
                <i class="pi pi-lock mr-2"></i>
                <div>
                    <strong>Security:</strong> Your token will be stored securely using your operating system's credential manager in the next step. Never share your token with anyone.
                </div>
            </div>

            <!-- Instructions Card -->
            <div class="instructions-card mb-6">
                <ol class="instructions-list">
                    <li class="mb-6">
                        <div class="flex items-center gap-3">
                            <span class="step-number">1</span>
                            <Button
                                label="Get Token from Plex.tv"
                                icon="pi pi-external-link"
                                iconPos="right"
                                @click="openPlexTokenPage"
                                severity="info"
                                outlined
                            />
                            <span class="text-sm text-gray-400">Sign in to your Plex account if prompted and copy the token</span>
                        </div>
                    </li>
                    <li>
                        <div class="flex items-start gap-3">
                            <span class="step-number">2</span>
                            <div class="flex-1 token-input-section" style="margin-top: -4px;">
                                <label for="plex-token" class="block text-sm font-medium mb-2">
                                    Paste your token here
                                    <span class="text-red-500">*</span>
                                </label>

                                <div class="token-input-wrapper">
                                    <InputText
                                        id="plex-token"
                                        v-model="tokenInput"
                                        :type="showToken ? 'text' : 'password'"
                                        placeholder="Enter your 20-character token here"
                                        class="token-input"
                                        :class="{ 'token-valid': tokenInput && tokenInput.trim().length > 0 }"
                                    />
                                    <Button
                                        :icon="showToken ? 'pi pi-eye-slash' : 'pi pi-eye'"
                                        @click="toggleTokenVisibility"
                                        class="toggle-visibility-btn"
                                        severity="secondary"
                                        text
                                        :aria-label="showToken ? 'Hide token' : 'Show token'"
                                    />
                                </div>

                                <small v-if="!tokenInput || tokenInput.trim().length === 0" class="helper-text text-muted-color">
                                    <i class="pi pi-info-circle mr-1"></i>
                                    Required field - enter your token to continue
                                </small>
                                <small v-else class="helper-text text-green-600">
                                    <i class="pi pi-check-circle mr-1"></i>
                                    Token entered - you can proceed to the next step
                                </small>
                            </div>
                        </div>
                    </li>
                </ol>
            </div>

            <!-- Server Discovery Section -->
            <div class="discovery-section mt-6">
                <div class="section-header mb-4">
                    <h3 class="text-xl font-semibold">{{ showManualEntry ? 'Enter Server Manually' : 'Discover Plex Servers' }}</h3>
                    <p class="text-sm text-muted-color mt-2">
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
                            <p class="text-sm text-muted-color mb-2">
                                <i class="pi pi-info-circle mr-1"></i>
                                <strong>Examples of valid URLs:</strong>
                            </p>
                            <ul class="example-list">
                                <li><code>http://192.168.1.100:32400</code> - Local server with IP</li>
                                <li><code>http://plex.local:32400</code> - Local server with hostname</li>
                                <li><code>https://plex.example.com:32400</code> - Remote server with HTTPS</li>
                            </ul>
                            <p class="text-sm text-muted-color mt-3">
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
                    <!-- Discover Button -->
                    <Button
                        v-if="!hasDiscovered && !isDiscovering"
                        label="Discover Servers"
                        icon="pi pi-search"
                        @click="discoverServers"
                        :disabled="!setupStore.isPlexStepValid"
                        class="w-full"
                        severity="success"
                    />

                    <!-- Always show "Enter Manually" option -->
                    <Button
                        v-if="!hasDiscovered && !isDiscovering"
                        label="Enter Manually"
                        icon="pi pi-pencil"
                        @click="enterManually"
                        severity="info"
                        outlined
                        class="w-full mt-3"
                    />

                <!-- Loading State -->
                <div v-if="isDiscovering" class="discovery-loading">
                    <ProgressSpinner
                        style="width: 50px; height: 50px"
                        strokeWidth="4"
                        fill="transparent"
                        animationDuration="1s"
                    />
                    <p class="text-muted-color mt-3">
                        Searching for Plex servers on your network...
                    </p>
                </div>

                <!-- Discovery Error -->
                <div v-if="discoveryError" class="discovery-error mt-4">
                    <i class="pi pi-exclamation-triangle mr-2"></i>
                    <span>{{ discoveryError }}</span>
                </div>

                <!-- Discovered Servers -->
                <div v-if="hasDiscovered && setupStore.discoveredServers.length > 0" class="discovered-servers mt-4">
                    <p class="text-sm text-muted-color mb-3">
                        Found {{ setupStore.discoveredServers.length }} server(s). Select one to continue:
                    </p>
                    <div class="server-list">
                        <ServerCard
                            v-for="server in setupStore.discoveredServers"
                            :key="server.id"
                            :server="server"
                            :is-selected="setupStore.selectedServer?.id === server.id"
                            @server-selected="handleServerSelected"
                        />
                    </div>
                    <div class="discovery-actions mt-4">
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
                        <p class="text-sm text-muted-color mb-4">
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
                    <small v-if="!setupStore.isPlexStepValid && !hasDiscovered" class="helper-text text-muted-color mt-3">
                        <i class="pi pi-info-circle mr-1"></i>
                        Enter your Plex token above to enable server discovery
                    </small>
                </div>
            </div>

            <!-- Connection Validation Section -->
            <div v-if="hasServerUrl" class="validation-section mt-6">
                <div class="section-header mb-4">
                    <h3 class="text-xl font-semibold">Verify Connection</h3>
                    <p class="text-sm text-muted-color mt-2">
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
                        severity="success"
                    />
                    <small class="helper-text text-muted-color mt-3">
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
                    <p class="text-muted-color mt-3">
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
                            <h4 class="font-semibold mb-2">
                                Connected to {{ setupStore.validationResult.serverName }}
                            </h4>
                            <div class="server-details">
                                <span class="detail-item">
                                    <i class="pi pi-server mr-1"></i>
                                    Version {{ setupStore.validationResult.serverVersion }}
                                </span>
                                <span class="detail-item">
                                    <i class="pi pi-folder mr-1"></i>
                                    {{ setupStore.validationResult.libraryCount }} libraries found
                                </span>
                            </div>
                        </div>
                    </Message>
                    <div class="validation-actions mt-4">
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
                    <small class="helper-text text-muted-color mt-3">
                        <i class="pi pi-lightbulb mr-1"></i>
                        Check that your Plex server is running and the token is correct
                    </small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.setup-step {
    max-width: 600px;
    margin: 0 auto;
}

.step-content {
    padding: 2rem 0;
}

.instructions-card {
    background: var(--surface-ground);
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid var(--surface-border);
}

.instructions-list {
    list-style: none;
    padding: 0;
    margin: 1rem 0 0 0;
}

.instructions-list li {
    display: flex;
    align-items: flex-start;
    margin-bottom: 0.75rem;
    padding-left: 0.5rem;
}

.step-number {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--primary-color);
    color: var(--primary-color-text);
    font-weight: bold;
    font-size: 0.875rem;
    margin-right: 0.75rem;
    flex-shrink: 0;
}

.token-input-section {
    margin-top: 2rem;
}

.token-input-wrapper {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}

.token-input {
    flex: 1;
    font-family: monospace;
    letter-spacing: 0.05em;
}

.token-input.token-valid {
    border-color: var(--green-500);
}

.toggle-visibility-btn {
    flex-shrink: 0;
}

.helper-text {
    display: flex;
    align-items: center;
    margin-top: 0.5rem;
    font-size: 0.875rem;
}

.security-note {
    display: flex;
    align-items: flex-start;
    background: var(--surface-ground);
    padding: 1rem;
    border-radius: 8px;
    border-left: 3px solid var(--primary-color);
    font-size: 0.875rem;
    color: var(--text-color-secondary);
}

.security-note i {
    color: var(--primary-color);
    margin-top: 0.125rem;
}

/* Discovery Section */
.discovery-section {
    background: var(--surface-ground);
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid var(--surface-border);
}

.section-header h3 {
    margin: 0;
}

.discovery-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
}

.discovery-error {
    display: flex;
    align-items: center;
    background: var(--red-50);
    color: var(--red-700);
    padding: 0.75rem;
    border-radius: 6px;
    border: 1px solid var(--red-200);
}

.discovery-error i {
    font-size: 1.25rem;
}

.server-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.discovery-actions {
    display: flex;
    justify-content: center;
}

.no-servers-found {
    text-align: center;
}

.no-servers-message {
    padding: 2rem 1rem;
    background: var(--surface-card);
    border-radius: 8px;
    border: 2px dashed var(--surface-border);
}

.no-servers-message i {
    color: var(--text-color-secondary);
}

.no-servers-actions {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 0.5rem;
}

/* Manual Entry Section */
.manual-entry-form {
    margin-top: 1rem;
}

.manual-entry-input-wrapper {
    margin-bottom: 1rem;
}

.error-text {
    display: flex;
    align-items: center;
    margin-top: 0.5rem;
    font-size: 0.875rem;
    color: var(--red-600);
}

.manual-entry-help {
    background: var(--surface-card);
    padding: 1rem;
    border-radius: 6px;
    border: 1px solid var(--surface-border);
}

.example-list {
    list-style: none;
    padding: 0;
    margin: 0.5rem 0;
}

.example-list li {
    padding: 0.25rem 0;
    font-size: 0.875rem;
    color: var(--text-color-secondary);
}

.example-list code {
    background: var(--surface-ground);
    padding: 0.125rem 0.375rem;
    border-radius: 3px;
    font-family: monospace;
    font-size: 0.8125rem;
    color: var(--primary-color);
}

.manual-entry-actions {
    display: flex;
    justify-content: center;
}

/* Validation Section */
.validation-section {
    background: var(--surface-ground);
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid var(--surface-border);
}

.validation-trigger {
    text-align: center;
}

.validation-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
}

.validation-success {
    text-align: left;
}

.validation-success .success-content {
    margin-left: 0.5rem;
}

.validation-success .server-details {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    margin-top: 0.5rem;
}

.validation-success .detail-item {
    display: inline-flex;
    align-items: center;
    font-size: 0.875rem;
    color: var(--text-color-secondary);
}

.validation-error {
    text-align: left;
}

.validation-error .error-content {
    margin-left: 0.5rem;
}

.validation-actions {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 0.5rem;
}

/* Responsive adjustments */
@media (max-width: 768px) {
    .step-content {
        padding: 1rem 0;
    }

    .instructions-card {
        padding: 1rem;
    }

    .token-input-wrapper {
        flex-direction: column;
        align-items: stretch;
    }

    .toggle-visibility-btn {
        width: 100%;
    }

    .discovery-section {
        padding: 1rem;
    }

    .no-servers-message {
        padding: 1.5rem 1rem;
    }

    .no-servers-actions {
        flex-direction: column;
    }

    .no-servers-actions button {
        width: 100%;
    }

    .validation-section {
        padding: 1rem;
    }

    .validation-actions {
        flex-direction: column;
    }

    .validation-actions button {
        width: 100%;
    }

    .validation-success .server-details {
        flex-direction: column;
        gap: 0.5rem;
    }
}
</style>
