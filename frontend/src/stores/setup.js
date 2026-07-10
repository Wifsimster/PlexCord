import { defineStore } from 'pinia';
import { SavePlexToken, IsDiscordConnected, ConnectDiscord, StartSessionPolling, CompleteSetup } from '../../wailsjs/go/main/App';

/**
 * Setup Wizard Store
 * Manages the state and navigation for the multi-step setup wizard
 */
export const useSetupStore = defineStore('setup', {
    state: () => ({
        // Wizard navigation
        currentStep: 0,
        completedSteps: [],
        setupComplete: false,

        // Plex configuration
        plexToken: '',
        plexServerUrl: '',
        plexUserId: '',
        discoveredServers: [],
        selectedServer: null,
        isManualEntry: false,
        isConnectionValidated: false,
        validationResult: null,

        // Plex user selection
        plexUsers: [],
        selectedPlexUser: null,

        // Discord configuration
        discordClientId: '',
        discordConnected: false,
        discordSkipped: false,

        // Finish action (spec §5.4 step 5 / F32) — side effects run once,
        // behind these flags; failures surface on the Complete step.
        isFinishing: false,
        finishError: ''
    }),

    getters: {
        /**
         * Check if a step has been completed
         * @param {number} stepIndex - The step index to check
         * @returns {boolean}
         */
        isStepCompleted: (state) => (stepIndex) => {
            return state.completedSteps.includes(stepIndex);
        },

        /**
         * Check if the Plex step is valid (token entered)
         * @returns {boolean}
         */
        isPlexStepValid: (state) => {
            return state.plexToken && state.plexToken.trim().length > 0;
        },

        /**
         * Check if navigation to next step is allowed
         * @returns {boolean}
         */
        canGoNext: (state) => {
            // For now, allow navigation to any step (validation will be added in future stories)
            return state.currentStep < 4; // 5 steps total (0-4)
        },

        /**
         * Check if navigation to previous step is allowed
         * @returns {boolean}
         */
        canGoBack: (state) => {
            return state.currentStep > 0;
        },

        /**
         * Get total number of steps
         * @returns {number}
         */
        totalSteps: () => {
            return 5; // Welcome, Plex, User, Discord, Complete
        },

        /**
         * Check if a server has been selected
         * @returns {boolean}
         */
        isServerSelected: (state) => {
            return state.selectedServer !== null;
        },

        /**
         * Check if connection has been validated
         * @returns {boolean}
         */
        isConnectionValid: (state) => {
            return state.isConnectionValidated && state.validationResult !== null;
        },

        /**
         * Check if a Plex user has been selected
         * @returns {boolean}
         */
        isUserSelected: (state) => {
            return state.selectedPlexUser !== null;
        },

        /**
         * Check if the Discord step gate is satisfied (connected or
         * explicitly skipped via the escape hatch — F29)
         * @returns {boolean}
         */
        isDiscordStepSatisfied: (state) => {
            return state.discordConnected || state.discordSkipped;
        },

        /**
         * Mono host summary of the accumulated Plex result for the wizard
         * rail (spec §5.4), e.g. 'plex.local:32400'
         * @returns {string}
         */
        plexServerSummary: (state) => {
            if (!state.plexServerUrl) {
                return '';
            }
            return state.plexServerUrl.replace(/^https?:\/\//, '').replace(/\/+$/, '');
        }
    },

    actions: {
        /**
         * Navigate to the next step
         */
        nextStep() {
            if (this.canGoNext) {
                // Mark current step as completed
                if (!this.completedSteps.includes(this.currentStep)) {
                    this.completedSteps.push(this.currentStep);
                }
                this.currentStep++;
                this.saveState();
            }
        },

        /**
         * Navigate to the previous step
         */
        previousStep() {
            if (this.canGoBack) {
                this.currentStep--;
                this.saveState();
            }
        },

        /**
         * Jump to a specific step (if completed or current)
         * @param {number} stepIndex - The step to jump to
         */
        goToStep(stepIndex) {
            if (stepIndex <= this.currentStep || this.isStepCompleted(stepIndex)) {
                this.currentStep = stepIndex;
                this.saveState();
            }
        },

        /**
         * Set the Plex authentication token
         * @param {string} token - The Plex token to store
         */
        setPlexToken(token) {
            this.plexToken = token;
            this.saveState();
        },

        /**
         * Set discovered Plex servers
         * @param {Array} servers - Array of discovered servers
         */
        setDiscoveredServers(servers) {
            this.discoveredServers = servers;
            this.saveState();
        },

        /**
         * Select a Plex server
         * @param {Object} server - The server to select
         */
        selectServer(server) {
            this.selectedServer = server;
            // Also set the server URL for config
            if (server) {
                this.plexServerUrl = `http://${server.address}:${server.port}`;
                this.isManualEntry = false;
            }
            this.saveState();
        },

        /**
         * Set manually entered server URL
         * @param {string} url - The server URL to store
         */
        setManualServerUrl(url) {
            this.plexServerUrl = url;
            this.isManualEntry = true;
            this.selectedServer = null; // Clear discovery selection
            this.saveState();
        },

        /**
         * Toggle manual entry mode
         * @param {boolean} enabled - True to enable manual entry, false for discovery
         */
        toggleManualEntry(enabled) {
            this.isManualEntry = enabled;
            if (!enabled) {
                // Switching back to discovery mode
                this.plexServerUrl = this.selectedServer ? `http://${this.selectedServer.address}:${this.selectedServer.port}` : '';
            }
            this.saveState();
        },

        /**
         * Set validation result after connection test
         * @param {Object} result - Validation result from backend
         */
        setValidationResult(result) {
            this.validationResult = result;
            this.isConnectionValidated = result.success === true;
            this.saveState();
        },

        /**
         * Clear validation state (for retry scenarios)
         */
        clearValidation() {
            this.validationResult = null;
            this.isConnectionValidated = false;
            this.saveState();
        },

        /**
         * Set the list of available Plex users
         * @param {Array} users - Array of PlexUser objects from backend
         */
        setPlexUsers(users) {
            this.plexUsers = users;
            this.saveState();
        },

        /**
         * Select a Plex user to monitor
         * @param {Object} user - The PlexUser object to select
         */
        selectPlexUser(user) {
            this.selectedPlexUser = user;
            if (user) {
                this.plexUserId = user.id;
            } else {
                this.plexUserId = '';
            }
            this.saveState();
        },

        /**
         * Clear the selected user (for retry/reset scenarios)
         */
        clearSelectedUser() {
            this.selectedPlexUser = null;
            this.plexUserId = '';
            this.plexUsers = [];
            this.saveState();
        },

        /**
         * Record the Discord connection state reached during the wizard
         * @param {boolean} connected
         */
        setDiscordConnected(connected) {
            this.discordConnected = connected;
            if (connected) {
                this.discordSkipped = false;
            }
            this.saveState();
        },

        /**
         * Record that the user chose "Continue without Discord" (F29);
         * the flag is surfaced as a warn panel on the Complete step.
         * @param {boolean} skipped
         */
        setDiscordSkipped(skipped) {
            this.discordSkipped = skipped;
            this.saveState();
        },

        /**
         * Finish setup (spec §5.4 step 5 / F32): connect Discord if needed →
         * StartSessionPolling → CompleteSetup. Side effects live here (not in
         * onMounted), idempotent behind the isFinishing/setupComplete flags;
         * failures are surfaced via finishError on the Complete step.
         * @returns {Promise<boolean>} true when setup completed
         */
        async finishSetup() {
            if (this.isFinishing) {
                return false;
            }
            if (this.setupComplete) {
                return true;
            }

            this.isFinishing = true;
            this.finishError = '';

            try {
                // Save Plex token to OS keychain
                if (this.plexToken) {
                    await SavePlexToken(this.plexToken);
                }

                // Connect Discord if needed (unless explicitly skipped)
                if (!this.discordSkipped) {
                    const connected = await IsDiscordConnected();
                    if (!connected) {
                        await ConnectDiscord('');
                    }
                    this.discordConnected = true;
                }

                // Start polling, then persist completion (also starts backend services)
                await StartSessionPolling();
                await CompleteSetup();

                this.completeSetup();
                return true;
            } catch (error) {
                console.error('Failed to complete setup:', error);
                this.finishError = typeof error === 'string' ? error : error?.message || 'Failed to complete setup. Please try again.';
                return false;
            } finally {
                this.isFinishing = false;
            }
        },

        /**
         * Save wizard state to localStorage
         */
        saveState() {
            const state = {
                currentStep: this.currentStep,
                completedSteps: this.completedSteps,
                setupComplete: this.setupComplete,
                // Never persist the token to localStorage - it's stored securely in the OS keychain
                plexServerUrl: this.plexServerUrl,
                plexUserId: this.plexUserId,
                discoveredServers: this.discoveredServers,
                selectedServer: this.selectedServer,
                isManualEntry: this.isManualEntry,
                isConnectionValidated: this.isConnectionValidated,
                validationResult: this.validationResult,
                plexUsers: this.plexUsers,
                selectedPlexUser: this.selectedPlexUser,
                discordClientId: this.discordClientId,
                discordConnected: this.discordConnected,
                discordSkipped: this.discordSkipped
            };
            localStorage.setItem('plexcord-setup-wizard', JSON.stringify(state));
        },

        /**
         * Load wizard state from localStorage
         */
        loadState() {
            const saved = localStorage.getItem('plexcord-setup-wizard');
            if (saved) {
                try {
                    const state = JSON.parse(saved);
                    this.currentStep = state.currentStep || 0;
                    this.completedSteps = state.completedSteps || [];
                    this.setupComplete = state.setupComplete || false;
                    this.plexToken = state.plexToken || '';
                    this.plexServerUrl = state.plexServerUrl || '';
                    this.plexUserId = state.plexUserId || '';
                    this.discoveredServers = state.discoveredServers || [];
                    this.selectedServer = state.selectedServer || null;
                    this.isManualEntry = state.isManualEntry || false;
                    this.isConnectionValidated = state.isConnectionValidated || false;
                    this.validationResult = state.validationResult || null;
                    this.plexUsers = state.plexUsers || [];
                    this.selectedPlexUser = state.selectedPlexUser || null;
                    this.discordClientId = state.discordClientId || '';
                    this.discordConnected = state.discordConnected || false;
                    this.discordSkipped = state.discordSkipped || false;
                } catch (error) {
                    console.error('Failed to load setup wizard state:', error);
                }
            }
        },

        /**
         * Complete the setup wizard and clear state
         */
        completeSetup() {
            this.setupComplete = true;
            this.saveState();
            // Clear wizard state after completion
            localStorage.removeItem('plexcord-setup-wizard');
        },

        /**
         * Reset wizard to initial state
         */
        resetWizard() {
            this.currentStep = 0;
            this.completedSteps = [];
            this.setupComplete = false;
            this.plexToken = '';
            this.plexServerUrl = '';
            this.plexUserId = '';
            this.discoveredServers = [];
            this.selectedServer = null;
            this.isManualEntry = false;
            this.isConnectionValidated = false;
            this.validationResult = null;
            this.plexUsers = [];
            this.selectedPlexUser = null;
            this.discordClientId = '';
            this.discordConnected = false;
            this.discordSkipped = false;
            this.isFinishing = false;
            this.finishError = '';
            localStorage.removeItem('plexcord-setup-wizard');
        }
    }
});
