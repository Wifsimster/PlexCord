import { defineStore } from 'pinia';
import { SavePlexToken, IsDiscordConnected, ConnectDiscord, StartSessionPolling, CompleteSetup } from '../../wailsjs/go/main/App';

export const useSetupStore = defineStore('setup', {
    state: () => ({
        currentStep: 0,
        completedSteps: [],
        setupComplete: false,

        plexToken: '',
        plexServerUrl: '',
        plexUserId: '',
        discoveredServers: [],
        selectedServer: null,
        isManualEntry: false,
        isConnectionValidated: false,
        validationResult: null,

        plexUsers: [],
        selectedPlexUser: null,

        discordClientId: '',
        discordConnected: false,
        discordSkipped: false,

        isFinishing: false,
        finishError: ''
    }),

    getters: {
        isStepCompleted: (state) => (stepIndex) => {
            return state.completedSteps.includes(stepIndex);
        },

        isPlexStepValid: (state) => {
            return state.plexToken && state.plexToken.trim().length > 0;
        },

        canGoNext: (state) => {
            return state.currentStep < 4;
        },

        canGoBack: (state) => {
            return state.currentStep > 0;
        },

        totalSteps: () => {
            return 5;
        },

        isServerSelected: (state) => {
            return state.selectedServer !== null;
        },

        isConnectionValid: (state) => {
            return state.isConnectionValidated && state.validationResult !== null;
        },

        isUserSelected: (state) => {
            return state.selectedPlexUser !== null;
        },

        isDiscordStepSatisfied: (state) => {
            return state.discordConnected || state.discordSkipped;
        },

        plexServerSummary: (state) => {
            if (!state.plexServerUrl) {
                return '';
            }
            return state.plexServerUrl.replace(/^https?:\/\//, '').replace(/\/+$/, '');
        }
    },

    actions: {
        nextStep() {
            if (this.canGoNext) {
                if (!this.completedSteps.includes(this.currentStep)) {
                    this.completedSteps.push(this.currentStep);
                }
                this.currentStep++;
                this.saveState();
            }
        },

        previousStep() {
            if (this.canGoBack) {
                this.currentStep--;
                this.saveState();
            }
        },

        goToStep(stepIndex) {
            if (stepIndex <= this.currentStep || this.isStepCompleted(stepIndex)) {
                this.currentStep = stepIndex;
                this.saveState();
            }
        },

        setPlexToken(token) {
            this.plexToken = token;
            this.saveState();
        },

        setDiscoveredServers(servers) {
            this.discoveredServers = servers;
            this.saveState();
        },

        selectServer(server) {
            this.selectedServer = server;
            if (server) {
                this.plexServerUrl = `http://${server.address}:${server.port}`;
                this.isManualEntry = false;
            }
            this.saveState();
        },

        setManualServerUrl(url) {
            this.plexServerUrl = url;
            this.isManualEntry = true;
            this.selectedServer = null;
            this.saveState();
        },

        toggleManualEntry(enabled) {
            this.isManualEntry = enabled;
            if (!enabled) {
                this.plexServerUrl = this.selectedServer ? `http://${this.selectedServer.address}:${this.selectedServer.port}` : '';
            }
            this.saveState();
        },

        setValidationResult(result) {
            this.validationResult = result;
            this.isConnectionValidated = result.success === true;
            this.saveState();
        },

        clearValidation() {
            this.validationResult = null;
            this.isConnectionValidated = false;
            this.saveState();
        },

        setPlexUsers(users) {
            this.plexUsers = users;
            this.saveState();
        },

        selectPlexUser(user) {
            this.selectedPlexUser = user;
            if (user) {
                this.plexUserId = user.id;
            } else {
                this.plexUserId = '';
            }
            this.saveState();
        },

        clearSelectedUser() {
            this.selectedPlexUser = null;
            this.plexUserId = '';
            this.plexUsers = [];
            this.saveState();
        },

        setDiscordConnected(connected) {
            this.discordConnected = connected;
            if (connected) {
                this.discordSkipped = false;
            }
            this.saveState();
        },

        setDiscordSkipped(skipped) {
            this.discordSkipped = skipped;
            this.saveState();
        },

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
                if (this.plexToken) {
                    await SavePlexToken(this.plexToken);
                }

                if (!this.discordSkipped) {
                    const connected = await IsDiscordConnected();
                    if (!connected) {
                        await ConnectDiscord('');
                    }
                    this.discordConnected = true;
                }

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

        saveState() {
            const state = {
                currentStep: this.currentStep,
                completedSteps: this.completedSteps,
                setupComplete: this.setupComplete,
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

        completeSetup() {
            this.setupComplete = true;
            this.saveState();
            localStorage.removeItem('plexcord-setup-wizard');
        },

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
