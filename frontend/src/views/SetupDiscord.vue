<script setup>
import { ref, computed, inject, onMounted, onUnmounted } from 'vue';
import { useSetupStore } from '@/stores/setup';
import { ConnectDiscord, IsDiscordConnected, GetDefaultDiscordClientID, GetDiscordClientID, SaveDiscordClientID, ValidateDiscordClientID, TestDiscordPresence } from '../../wailsjs/go/main/App';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import InputText from 'primevue/inputtext';
import DrawnCheck from '@/components/setup/DrawnCheck.vue';

const setupStore = useSetupStore();
const wizard = inject('setupWizard', null);

// Connection state machine
const connectionState = ref(setupStore.discordConnected ? 'connected' : 'initial'); // 'initial' | 'connecting' | 'connected' | 'error'
const connectionError = ref('');
const isConnecting = ref(false);

// Test presence state (M15 — transient inline confirmation)
const isTesting = ref(false);
const testSent = ref(false);
const testError = ref('');
let testSentTimer = null;

// Custom Client ID (Advanced) — the value persists when the section is
// collapsed (F31); a CUSTOM ID badge marks it while hidden.
const showAdvanced = ref(false);
const customClientId = ref('');
const defaultClientId = ref('');
const clientIdError = ref('');
const isClientIdValid = ref(true);
const showInstructions = ref(false);

const loadClientId = async () => {
    try {
        defaultClientId.value = await GetDefaultDiscordClientID();
        const currentId = await GetDiscordClientID();
        if (currentId && currentId !== defaultClientId.value) {
            customClientId.value = currentId;
        }
    } catch (error) {
        console.error('Failed to load Discord client ID:', error);
    }
};

const hasCustomClientId = computed(() => customClientId.value.trim().length > 0);

const activeClientId = computed(() => (hasCustomClientId.value ? customClientId.value.trim() : defaultClientId.value));

// ---- Client ID validation -------------------------------------------------
const validateClientId = async () => {
    if (!hasCustomClientId.value) {
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

const handleClientIdChange = async () => {
    await validateClientId();
};

const saveCustomClientId = async () => {
    if (!hasCustomClientId.value) {
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

// ---- Connect ---------------------------------------------------------------
const connectToDiscord = async () => {
    if (isConnecting.value) {
        return;
    }

    // Save client ID first if using custom
    const saved = await saveCustomClientId();
    if (!saved && hasCustomClientId.value) {
        showAdvanced.value = true;
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
            setupStore.setDiscordConnected(true);
        } else {
            throw new Error('Connection verification failed');
        }
    } catch (error) {
        console.error('Discord connection failed:', error);
        connectionState.value = 'error';
        setupStore.setDiscordConnected(false);

        // Parse error message for user-friendly display
        let errorMessage = 'Failed to connect to Discord';
        if (error && typeof error === 'string') {
            errorMessage = error;
        } else if (error && error.message) {
            errorMessage = error.message;
        }

        if (errorMessage.includes('not running')) {
            connectionError.value = 'Discord is not running. Start Discord, then try again.';
        } else if (errorMessage.includes('invalid') || errorMessage.includes('Client ID')) {
            connectionError.value = 'Invalid Discord Client ID. Check your configuration below.';
        } else {
            connectionError.value = errorMessage;
        }
    } finally {
        isConnecting.value = false;
    }
};

const retryConnection = () => {
    connectionError.value = '';
    connectToDiscord();
};

// ---- Test presence (M15) ----------------------------------------------------
const testPresence = async () => {
    isTesting.value = true;
    testError.value = '';
    testSent.value = false;

    try {
        await TestDiscordPresence();
        testSent.value = true;
        if (testSentTimer) {
            clearTimeout(testSentTimer);
        }
        // M15: hold 1600ms, then fade out via <Transition>
        testSentTimer = setTimeout(() => {
            testSent.value = false;
            testSentTimer = null;
        }, 1600);
    } catch (error) {
        console.error('Test presence failed:', error);
        testError.value = error?.message || 'Failed to send test presence';
    } finally {
        isTesting.value = false;
    }
};

// Collapsing keeps the entered value (F31)
const toggleAdvanced = () => {
    showAdvanced.value = !showAdvanced.value;
};

const openDeveloperPortal = () => {
    BrowserOpenURL('https://discord.com/developers/applications');
};

// ---- Wizard integration ------------------------------------------------------
// Enter submits this step's primary action while unconnected (§5.4)
const primaryAction = () => {
    if (connectionState.value === 'initial' || connectionState.value === 'error') {
        connectToDiscord();
        return true;
    }
    return false;
};

onMounted(async () => {
    await loadClientId();
    wizard?.registerPrimary(primaryAction);

    // Restore live connection state (e.g. connected in an earlier visit)
    try {
        const connected = await IsDiscordConnected();
        if (connected) {
            connectionState.value = 'connected';
            setupStore.setDiscordConnected(true);
        } else if (connectionState.value === 'connected') {
            connectionState.value = 'initial';
            setupStore.setDiscordConnected(false);
        }
    } catch (error) {
        console.error('Failed to check Discord connection state:', error);
    }
});

onUnmounted(() => {
    wizard?.unregisterPrimary(primaryAction);
    if (testSentTimer) {
        clearTimeout(testSentTimer);
        testSentTimer = null;
    }
});
</script>

<template>
    <div>
        <h1 class="setup-title">Connect to Discord</h1>
        <p class="setup-lede">PlexCord will show your Plex music activity on your Discord profile using Rich Presence.</p>

        <div class="setup-panels">
            <!-- Connection panel -->
            <section class="pc-panel">
                <span class="pc-eyebrow">Discord Rich Presence</span>

                <Transition name="pc-state" mode="out-in">
                    <!-- initial -->
                    <div v-if="connectionState === 'initial'" key="initial" class="discord-initial">
                        <p class="discord-notice">Discord must be running on this computer.</p>
                        <button type="button" class="pc-btn pc-btn--primary pc-btn--lg" :disabled="!isClientIdValid" @click="connectToDiscord">Connect to Discord</button>
                    </div>

                    <!-- connecting -->
                    <div v-else-if="connectionState === 'connecting'" key="connecting" class="discord-connecting">
                        <i class="pi pi-spin pi-spinner discord-spinner" aria-hidden="true"></i>
                        <span>Connecting to Discord…</span>
                    </div>

                    <!-- connected -->
                    <div v-else-if="connectionState === 'connected'" key="connected" class="discord-done">
                        <p class="discord-done-title">
                            <DrawnCheck :size="14" />
                            <span>Connected to Discord</span>
                        </p>
                        <p class="discord-done-caption">Your profile will update when you play music on Plex.</p>
                        <div class="discord-done-actions">
                            <button type="button" class="pc-btn pc-btn--secondary pc-btn--sm" :disabled="isTesting" @click="testPresence">
                                <i v-if="isTesting" class="pi pi-spin pi-spinner discord-spinner" aria-hidden="true"></i>
                                Send test presence
                            </button>
                            <Transition name="pc-fade">
                                <span v-if="testSent" class="discord-test-ok pc-fade-ok"><i class="pi pi-check" aria-hidden="true"></i> Sent — check your Discord profile</span>
                            </Transition>
                            <span v-if="testError" class="discord-test-error" role="alert">{{ testError }}</span>
                        </div>
                        <p class="discord-reconnect">
                            Connection trouble?
                            <a href="#" @click.prevent="retryConnection">Reconnect</a>
                        </p>
                    </div>

                    <!-- error -->
                    <div v-else key="error" class="discord-error" role="alert">
                        <p class="discord-error-title"><i class="pi pi-times-circle" aria-hidden="true"></i> Connection failed</p>
                        <p class="discord-error-text">{{ connectionError }}</p>
                        <button type="button" class="pc-btn pc-btn--secondary pc-btn--sm" :disabled="isConnecting" @click="retryConnection">
                            <i v-if="isConnecting" class="pi pi-spin pi-spinner discord-spinner" aria-hidden="true"></i>
                            Retry
                        </button>
                    </div>
                </Transition>
            </section>

            <!-- Advanced: custom Client ID (persistent value — F31) -->
            <section class="pc-panel">
                <button type="button" class="advanced-toggle" :aria-expanded="showAdvanced" @click="toggleAdvanced">
                    <span class="pc-eyebrow">Advanced</span>
                    <span v-if="hasCustomClientId" class="pc-badge pc-badge--accent">Custom ID</span>
                    <i class="pi advanced-chevron" :class="showAdvanced ? 'pi-chevron-up' : 'pi-chevron-down'" aria-hidden="true"></i>
                </button>

                <div class="pc-collapse" :class="{ 'pc-collapse--open': showAdvanced }">
                    <div>
                        <div class="advanced-body">
                            <label class="advanced-label" for="custom-client-id">Custom Discord application Client ID</label>
                            <p class="advanced-caption">Use your own Discord application for custom branding or testing. Leave empty to use the default PlexCord application.</p>
                            <InputText
                                id="custom-client-id"
                                v-model="customClientId"
                                class="advanced-input"
                                placeholder="17–20 digit application ID"
                                :invalid="!isClientIdValid"
                                aria-describedby="client-id-feedback"
                                @input="handleClientIdChange"
                                @blur="handleClientIdChange"
                            />
                            <small v-if="clientIdError" id="client-id-feedback" class="advanced-feedback advanced-feedback--danger" role="alert"><i class="pi pi-exclamation-circle" aria-hidden="true"></i> {{ clientIdError }}</small>
                            <small v-else id="client-id-feedback" class="advanced-feedback">
                                Default: <span class="pc-chip-mono">{{ defaultClientId }}</span>
                            </small>

                            <p class="advanced-instructions-link">
                                <a href="#" @click.prevent="showInstructions = !showInstructions">How to create a Discord application</a>
                            </p>

                            <div class="pc-collapse" :class="{ 'pc-collapse--open': showInstructions }">
                                <div>
                                    <div class="advanced-instructions">
                                        <ol class="advanced-steps">
                                            <li>
                                                Open the
                                                <a href="#" @click.prevent="openDeveloperPortal">Discord Developer Portal</a>
                                            </li>
                                            <li>Click "New Application" and give it a name</li>
                                            <li>Copy the "Application ID" from the General Information page</li>
                                            <li>Paste it in the field above</li>
                                            <li>(Optional) Upload custom images under Rich Presence Art Assets</li>
                                        </ol>
                                        <p class="advanced-note">Custom applications require additional setup. Most users should use the default PlexCord application.</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    </div>
</template>

<style scoped>
/* ---- Connection panel states ---- */
.discord-initial {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    margin-top: 12px;
}
.discord-notice {
    margin: 0;
    padding: 8px 12px;
    border-radius: var(--pc-radius-md);
    background: var(--pc-raised);
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
}
.discord-connecting {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
    font-size: var(--pc-text-body);
    color: var(--pc-text-secondary);
}
.discord-spinner {
    font-size: 12px;
}
.discord-done {
    margin-top: 12px;
}
.discord-done-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 4px;
    font-size: var(--pc-text-body);
    color: var(--pc-text);
}
.discord-done-caption {
    margin: 0 0 12px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.discord-done-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
}
.discord-test-ok {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: var(--pc-text-caption);
    color: var(--pc-success);
}
.discord-test-error {
    font-size: var(--pc-text-caption);
    color: var(--pc-danger);
}
.discord-reconnect {
    margin: 12px 0 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.discord-reconnect a {
    color: var(--pc-accent);
    text-decoration: none;
}
.discord-reconnect a:hover {
    text-decoration: underline;
}
.discord-error {
    margin-top: 12px;
    padding: 12px;
    border: 1px solid color-mix(in srgb, var(--pc-danger) 40%, transparent);
    border-radius: var(--pc-radius-md);
    background: var(--pc-danger-dim);
}
.discord-error-title {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 4px;
    font-size: var(--pc-text-body);
    font-weight: 600;
    color: var(--pc-danger);
}
.discord-error-text {
    margin: 0 0 10px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    overflow-wrap: anywhere;
}

/* ---- Advanced section ---- */
.advanced-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 0;
    background: none;
    border: none;
    cursor: pointer;
    text-align: left;
}
.advanced-chevron {
    margin-left: auto;
    font-size: 12px;
    color: var(--pc-text-muted);
}
.advanced-body {
    padding-top: 12px;
}
.advanced-label {
    display: block;
    margin-bottom: 4px;
    font-size: var(--pc-text-caption);
    font-weight: 500;
    color: var(--pc-text-secondary);
}
.advanced-caption {
    margin: 0 0 8px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.advanced-input {
    width: 100%;
    font-family: var(--pc-font-mono);
    font-size: var(--pc-text-mono);
}
.advanced-feedback {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.advanced-feedback--danger {
    color: var(--pc-danger);
}
.advanced-instructions-link {
    margin: 12px 0 0;
    font-size: var(--pc-text-caption);
}
.advanced-instructions-link a,
.advanced-steps a {
    color: var(--pc-accent);
    text-decoration: none;
}
.advanced-instructions-link a:hover,
.advanced-steps a:hover {
    text-decoration: underline;
}
.advanced-instructions {
    margin-top: 8px;
    padding: 12px;
    border-radius: var(--pc-radius-md);
    background: var(--pc-raised);
}
.advanced-steps {
    margin: 0;
    padding-left: 18px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
}
.advanced-note {
    margin: 10px 0 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
</style>
