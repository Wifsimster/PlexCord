<script setup>
import { ref, watch, computed, inject, onMounted, onUnmounted } from 'vue';
import { useSetupStore } from '@/stores/setup';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { DiscoverPlexServers, ValidatePlexConnection, SavePlexToken, SaveServerURL, StartPlexPINAuth, CheckPlexPINAuth } from '../../wailsjs/go/main/App';
import { validatePlexServerUrl, PLEX_URL_PLACEHOLDER } from '@/utils/plexUrl';
import InputText from 'primevue/inputtext';
import DrawnCheck from '@/components/setup/DrawnCheck.vue';

const setupStore = useSetupStore();
const wizard = inject('setupWizard', null);

// ---- PIN authentication state machine (kept — spec §5.4 step 2.1) ------
const authStep = ref(setupStore.isPlexStepValid ? 'success' : 'initial'); // 'initial' | 'loading' | 'waiting' | 'success' | 'error'
const pinCode = ref('');
const pinID = ref(null);
const authURL = ref('');
const authError = ref('');
const checkInterval = ref(null);
const pinPollFailures = ref(0);
const MAX_PIN_POLL_FAILURES = 5;

// ---- Server discovery state ---------------------------------------------
const isDiscovering = ref(false);
const discoveryError = ref('');
const hasDiscovered = ref(setupStore.discoveredServers.length > 0);

// ---- Manual server entry state ------------------------------------------
const showManualEntry = ref(setupStore.isManualEntry);
const manualServerUrl = ref(setupStore.isManualEntry ? setupStore.plexServerUrl : '');
const manualEntryError = ref('');
const isManualUrlValid = ref(setupStore.isManualEntry && validatePlexServerUrl(setupStore.plexServerUrl).valid);

// ---- Validation state ----------------------------------------------------
const isValidating = ref(false);
const validationError = ref('');
const validationAttempted = ref(false);

const signedIn = computed(() => authStep.value === 'success' || setupStore.isPlexStepValid);
const hasServerUrl = computed(() => setupStore.plexServerUrl && setupStore.plexServerUrl.trim().length > 0);
const canValidate = computed(() => setupStore.isPlexStepValid && hasServerUrl.value);

const libraryLabel = computed(() => {
    const count = setupStore.validationResult?.libraryCount ?? 0;
    return `${count} ${count === 1 ? 'library' : 'libraries'}`;
});

// ---- PIN auth ------------------------------------------------------------
const startPINAuth = async () => {
    authError.value = '';
    authStep.value = 'loading';

    try {
        const result = await StartPlexPINAuth();
        pinCode.value = result.pinCode;
        pinID.value = result.pinID;
        authURL.value = result.authURL;
        authStep.value = 'waiting';
        pinPollFailures.value = 0;

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

const stopPinPolling = () => {
    if (checkInterval.value) {
        clearInterval(checkInterval.value);
        checkInterval.value = null;
    }
};

const checkPINStatus = async () => {
    try {
        const result = await CheckPlexPINAuth(pinID.value);
        pinPollFailures.value = 0;

        if (result.authorized) {
            // Success! Save the token
            setupStore.setPlexToken(result.authToken);
            authStep.value = 'success';
            stopPinPolling();
        } else if (result.expired) {
            authError.value = 'PIN expired. Please try again.';
            authStep.value = 'error';
            stopPinPolling();
        }
    } catch (error) {
        console.error('Error checking PIN status:', error);
        // Bail out after repeated failures so the interval doesn't loop
        // forever on a persistent network/backend failure.
        pinPollFailures.value += 1;
        if (pinPollFailures.value >= MAX_PIN_POLL_FAILURES) {
            authError.value = error?.message || 'Failed to check PIN status. Please try again.';
            authStep.value = 'error';
            stopPinPolling();
        }
    }
};

const cancelPINAuth = () => {
    stopPinPolling();
    authStep.value = 'initial';
    pinCode.value = '';
    pinID.value = null;
    authURL.value = '';
    authError.value = '';
};

// ---- Discovery (auto-runs on auth success — F23) -------------------------
const discoverServers = async () => {
    isDiscovering.value = true;
    discoveryError.value = '';

    try {
        const servers = await DiscoverPlexServers();
        setupStore.setDiscoveredServers(servers || []);
        hasDiscovered.value = true;
    } catch (error) {
        console.error('Failed to discover servers:', error);
        discoveryError.value = 'Discovery failed. Try again, or enter your server address manually.';
        setupStore.setDiscoveredServers([]);
        hasDiscovered.value = true;
    } finally {
        isDiscovering.value = false;
    }
};

watch(
    signedIn,
    (isSignedIn) => {
        if (isSignedIn && !showManualEntry.value && !hasDiscovered.value && !isDiscovering.value && !setupStore.isConnectionValidated) {
            discoverServers();
        }
    },
    { immediate: true }
);

// ---- Validation (auto-runs on selection — F23) ---------------------------
const validateConnection = async () => {
    if (!canValidate.value || isValidating.value) {
        return;
    }

    isValidating.value = true;
    validationError.value = '';
    validationAttempted.value = true;
    setupStore.clearValidation();

    try {
        // Save the token to keychain first
        await SavePlexToken(setupStore.plexToken);

        // Validate the connection
        const result = await ValidatePlexConnection(setupStore.plexServerUrl);
        setupStore.setValidationResult(result);
        validationError.value = '';

        // Save server URL to backend config after successful validation
        await SaveServerURL(setupStore.plexServerUrl);
    } catch (error) {
        console.error('Connection validation failed:', error);

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

const serverUrlOf = (server) => `http://${server.address}:${server.port}`;

const isSelected = (server) => !showManualEntry.value && setupStore.selectedServer?.id === server.id;

// Selecting a server auto-validates it (spec §5.4 step 2.3)
const selectServer = async (server) => {
    if (isSelected(server) && setupStore.isConnectionValidated) {
        return;
    }
    setupStore.selectServer(server);
    await validateConnection();
};

const serverDotClass = (server) => {
    if (!isSelected(server)) {
        return 'pc-dot--idle';
    }
    if (setupStore.isConnectionValidated) {
        return 'pc-dot--success';
    }
    if (validationError.value) {
        return 'pc-dot--danger';
    }
    return 'pc-dot--idle';
};

// ---- Manual entry --------------------------------------------------------
const enterManually = () => {
    showManualEntry.value = true;
    setupStore.toggleManualEntry(true);
    manualEntryError.value = '';
};

const useDiscoveryInstead = () => {
    showManualEntry.value = false;
    setupStore.toggleManualEntry(false);
    manualServerUrl.value = '';
    manualEntryError.value = '';
    isManualUrlValid.value = false;
    if (!hasDiscovered.value && !isDiscovering.value) {
        discoverServers();
    }
};

const handleManualUrlChange = () => {
    const validation = validatePlexServerUrl(manualServerUrl.value);
    isManualUrlValid.value = validation.valid;
    manualEntryError.value = validation.error;

    if (validation.valid) {
        setupStore.setManualServerUrl(manualServerUrl.value.trim());
    }
};

const handleManualUrlBlur = () => {
    if (manualServerUrl.value.trim().length > 0) {
        handleManualUrlChange();
    }
};

// Valid manual URL + Enter auto-validates (spec §5.4 step 2.3)
const submitManualUrl = () => {
    handleManualUrlChange();
    if (isManualUrlValid.value) {
        validateConnection();
    }
};

const retryValidation = () => {
    validationError.value = '';
    validateConnection();
};

// ---- Clear stale validation when inputs change ---------------------------
watch(
    () => setupStore.plexServerUrl,
    (newUrl, oldUrl) => {
        if (newUrl !== oldUrl && validationAttempted.value && !isValidating.value) {
            setupStore.clearValidation();
            validationAttempted.value = false;
            validationError.value = '';
        }
    }
);

watch(
    () => setupStore.plexToken,
    (newToken, oldToken) => {
        if (newToken !== oldToken && validationAttempted.value && !isValidating.value) {
            setupStore.clearValidation();
            validationAttempted.value = false;
            validationError.value = '';
        }
    }
);

// ---- Wizard integration ---------------------------------------------------
// Enter submits this step's primary action while sign-in is pending (§5.4)
const primaryAction = () => {
    if (authStep.value === 'initial' || authStep.value === 'error') {
        startPINAuth();
        return true;
    }
    return false;
};

onMounted(() => {
    if (setupStore.isConnectionValidated && setupStore.validationResult) {
        validationAttempted.value = true;
    }
    wizard?.registerPrimary(primaryAction);
});

onUnmounted(() => {
    stopPinPolling();
    wizard?.unregisterPrimary(primaryAction);
});
</script>

<template>
    <div>
        <h1 class="setup-title">Connect to your Plex account</h1>
        <p class="setup-lede">Sign in with a secure PIN code, then pick the server PlexCord should watch.</p>

        <div class="setup-panels">
            <!-- Sign-in panel -->
            <section class="pc-panel">
                <span class="pc-eyebrow">Plex account</span>
                <Transition name="pc-state" mode="out-in">
                    <!-- initial / loading -->
                    <div v-if="authStep === 'initial' || authStep === 'loading'" key="initial" class="auth-initial">
                        <button type="button" class="pc-btn pc-btn--primary pc-btn--lg auth-signin" :disabled="authStep === 'loading'" @click="startPINAuth">
                            <i v-if="authStep === 'loading'" class="pi pi-spin pi-spinner" aria-hidden="true"></i>
                            Sign in with Plex
                        </button>
                        <span class="auth-caption">Opens plex.tv in your browser</span>
                    </div>

                    <!-- waiting for authorization -->
                    <div v-else-if="authStep === 'waiting'" key="waiting" class="auth-waiting">
                        <span class="auth-pin" aria-label="Your Plex PIN code">{{ pinCode }}</span>
                        <span class="auth-caption"><i class="pi pi-spin pi-spinner auth-spinner" aria-hidden="true"></i> Enter this code at plex.tv/link</span>
                        <div class="auth-actions">
                            <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="BrowserOpenURL(authURL)">Open browser again</button>
                            <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="cancelPINAuth">Cancel</button>
                        </div>
                    </div>

                    <!-- signed in (collapsed row) -->
                    <div v-else-if="authStep === 'success'" key="success" class="auth-done">
                        <DrawnCheck :size="14" />
                        <span>Signed in with Plex</span>
                    </div>

                    <!-- error -->
                    <div v-else key="error" class="auth-error" role="alert">
                        <p class="auth-error-text"><i class="pi pi-exclamation-circle" aria-hidden="true"></i> {{ authError }}</p>
                        <button type="button" class="pc-btn pc-btn--secondary pc-btn--sm" @click="startPINAuth">Try again</button>
                    </div>
                </Transition>
            </section>

            <!-- Server panel (appears once signed in; discovery auto-runs) -->
            <section v-if="signedIn" class="pc-panel">
                <div class="server-head">
                    <span class="pc-eyebrow">Plex server</span>
                    <button v-if="!showManualEntry && hasDiscovered && !isDiscovering" type="button" class="server-head-link" @click="discoverServers">Search again</button>
                </div>

                <!-- Discovery mode -->
                <template v-if="!showManualEntry">
                    <!-- skeleton rows while discovering (M20) -->
                    <div v-if="isDiscovering" class="server-skeletons" aria-label="Searching for Plex servers">
                        <div class="pc-skeleton server-skeleton"></div>
                        <div class="pc-skeleton server-skeleton"></div>
                        <div class="pc-skeleton server-skeleton"></div>
                    </div>

                    <!-- discovery error -->
                    <div v-else-if="discoveryError" class="server-note" role="alert">
                        <p class="server-note-text server-note-text--danger"><i class="pi pi-exclamation-triangle" aria-hidden="true"></i> {{ discoveryError }}</p>
                        <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="discoverServers">Search again</button>
                    </div>

                    <!-- selectable server rows -->
                    <ul v-else-if="setupStore.discoveredServers.length > 0" class="server-list">
                        <li v-for="server in setupStore.discoveredServers" :key="server.id">
                            <button
                                type="button"
                                class="server-row"
                                :class="{ 'server-row--selected': isSelected(server), 'server-row--invalid': isSelected(server) && validationError }"
                                :aria-pressed="isSelected(server)"
                                @click="selectServer(server)"
                            >
                                <span class="pc-dot" :class="serverDotClass(server)" aria-hidden="true"></span>
                                <span class="server-name">{{ server.name }}</span>
                                <span class="pc-chip-mono server-url">{{ serverUrlOf(server) }}</span>
                                <span class="pc-badge">{{ server.isLocal ? 'Local' : 'Remote' }}</span>
                                <i v-if="isSelected(server) && isValidating" class="pi pi-spin pi-spinner server-spinner" aria-hidden="true"></i>
                                <DrawnCheck v-else-if="isSelected(server) && setupStore.isConnectionValidated" :size="14" />
                            </button>
                        </li>
                    </ul>

                    <!-- no servers found -->
                    <div v-else-if="hasDiscovered" class="server-note">
                        <p class="server-note-text">No Plex servers found on your local network. Make sure your server is running, or enter its address manually below.</p>
                        <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="discoverServers">Search again</button>
                    </div>

                    <!-- persistent manual-entry hint (F25) -->
                    <p class="server-hint">
                        Docker or remote server?
                        <a href="#" @click.prevent="enterManually">Enter its address manually</a>
                    </p>
                </template>

                <!-- Manual entry mode -->
                <template v-else>
                    <div class="manual-entry">
                        <label class="manual-label" for="manual-server-url">Server address</label>
                        <InputText
                            id="manual-server-url"
                            v-model="manualServerUrl"
                            class="manual-input"
                            :placeholder="PLEX_URL_PLACEHOLDER"
                            :invalid="!!manualEntryError && manualServerUrl.trim().length > 0"
                            aria-describedby="manual-url-feedback"
                            @input="handleManualUrlChange"
                            @blur="handleManualUrlBlur"
                            @keydown.enter.prevent="submitManualUrl"
                        />
                        <small v-if="manualEntryError && manualServerUrl.trim().length > 0" id="manual-url-feedback" class="manual-feedback manual-feedback--danger" role="alert">
                            <i class="pi pi-exclamation-circle" aria-hidden="true"></i>
                            {{ manualEntryError }}
                        </small>
                        <small v-else-if="isManualUrlValid" id="manual-url-feedback" class="manual-feedback manual-feedback--success">
                            <i class="pi pi-check-circle" aria-hidden="true"></i>
                            Valid format — press Enter to test the connection
                        </small>
                        <small v-else id="manual-url-feedback" class="manual-feedback">The default Plex port is 32400. Use https for remote connections.</small>
                    </div>

                    <p class="server-hint">
                        On your local network?
                        <a href="#" @click.prevent="useDiscoveryInstead">Use auto-discovery instead</a>
                    </p>
                </template>

                <!-- Validation outcome -->
                <div v-if="isValidating && showManualEntry" class="validate-pending">
                    <i class="pi pi-spin pi-spinner server-spinner" aria-hidden="true"></i>
                    <span>Testing connection…</span>
                </div>

                <div v-if="setupStore.isConnectionValidated && setupStore.validationResult" class="validate-ok">
                    <DrawnCheck :size="14" />
                    <span>
                        Reachable — {{ libraryLabel }}
                        <template v-if="setupStore.validationResult.serverName"
                            >on <strong>{{ setupStore.validationResult.serverName }}</strong></template
                        >
                    </span>
                </div>

                <!-- Validation failure detail (M17) -->
                <div class="pc-collapse" :class="{ 'pc-collapse--open': validationError && !isValidating }">
                    <div>
                        <div class="validate-error" role="alert">
                            <p class="validate-error-title"><i class="pi pi-times-circle" aria-hidden="true"></i> Couldn't reach this server</p>
                            <p class="validate-error-text">{{ validationError }}</p>
                            <p class="validate-error-suggestion">Check that your Plex server is running and reachable from this computer.</p>
                            <button type="button" class="pc-btn pc-btn--ghost-danger pc-btn--sm" @click="retryValidation">Retry</button>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    </div>
</template>

<style scoped>
/* ---- Sign-in panel ---- */
.auth-initial {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 12px;
}
.auth-caption {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.auth-waiting {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    margin-top: 12px;
}
.auth-pin {
    display: inline-block;
    padding: 8px 16px;
    border-radius: var(--pc-radius-sm);
    background: var(--pc-raised);
    font-family: var(--pc-font-mono);
    font-size: 28px;
    letter-spacing: 0.12em;
    color: var(--pc-text);
    user-select: all;
}
.auth-spinner,
.server-spinner {
    font-size: 12px;
}
.auth-actions {
    display: flex;
    gap: 8px;
}
.auth-done {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
    font-size: var(--pc-text-body);
    color: var(--pc-text);
}
.auth-error {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    margin-top: 12px;
}
.auth-error-text {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0;
    font-size: var(--pc-text-body);
    color: var(--pc-danger);
}

/* ---- Server panel ---- */
.server-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
}
.server-head-link {
    background: none;
    border: none;
    padding: 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    cursor: pointer;
}
.server-head-link:hover {
    color: var(--pc-accent);
}
.server-skeletons {
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.server-skeleton {
    height: 44px;
    border-radius: var(--pc-radius-md);
}
.server-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.server-row {
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
    transition:
        border-color var(--pc-dur-1) var(--pc-ease-out),
        background-color var(--pc-dur-1) var(--pc-ease-out);
}
.server-row:hover {
    border-color: var(--pc-border-strong);
}
.server-row--selected {
    border-color: var(--pc-accent);
    background: var(--pc-accent-dim);
}
.server-row--invalid {
    border-color: var(--pc-danger);
    background: var(--pc-danger-dim);
}
.server-name {
    font-size: var(--pc-text-body);
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.server-url {
    margin-left: auto;
}
.server-note {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 12px;
    background: var(--pc-raised);
    border-radius: var(--pc-radius-md);
}
.server-note-text {
    margin: 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
}
.server-note-text--danger {
    color: var(--pc-danger);
    display: flex;
    align-items: center;
    gap: 6px;
}
.server-hint {
    margin: 12px 0 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.server-hint a {
    color: var(--pc-accent);
    text-decoration: none;
}
.server-hint a:hover {
    text-decoration: underline;
}

/* ---- Manual entry ---- */
.manual-entry {
    display: flex;
    flex-direction: column;
    gap: 6px;
}
.manual-label {
    font-size: var(--pc-text-caption);
    font-weight: 500;
    color: var(--pc-text-secondary);
}
.manual-input {
    width: 100%;
    font-family: var(--pc-font-mono);
    font-size: var(--pc-text-mono);
}
.manual-feedback {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.manual-feedback--danger {
    color: var(--pc-danger);
}
.manual-feedback--success {
    color: var(--pc-success);
}

/* ---- Validation outcome ---- */
.validate-pending {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
}
.validate-ok {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
    font-size: var(--pc-text-body);
    color: var(--pc-text);
}
.validate-error {
    margin-top: 12px;
    padding: 12px;
    border: 1px solid color-mix(in srgb, var(--pc-danger) 40%, transparent);
    border-radius: var(--pc-radius-md);
    background: var(--pc-danger-dim);
}
.validate-error-title {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 4px;
    font-size: var(--pc-text-body);
    font-weight: 600;
    color: var(--pc-danger);
}
.validate-error-text {
    margin: 0 0 4px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    overflow-wrap: anywhere;
}
.validate-error-suggestion {
    margin: 0 0 10px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
</style>
