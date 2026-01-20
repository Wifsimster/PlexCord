<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useSetupStore } from '@/stores/setup';
import { SavePlexToken, CompleteSetup, SkipSetup } from '../../wailsjs/go/main/App';
import { useToast } from 'primevue/usetoast';
import Steps from 'primevue/steps';
import Button from 'primevue/button';
import Card from 'primevue/card';

const toast = useToast();

const router = useRouter();
const route = useRoute();
const setupStore = useSetupStore();

// Step definitions
const steps = ref([
    { label: 'Welcome', route: '/setup/welcome' },
    { label: 'Plex Server', route: '/setup/plex' },
    { label: 'Select User', route: '/setup/user' },
    { label: 'Discord', route: '/setup/discord' },
    { label: 'Complete', route: '/setup/complete' }
]);

// Current active step based on route
const activeStep = computed(() => {
    const currentPath = route.path;
    const index = steps.value.findIndex(step => step.route === currentPath);
    return index >= 0 ? index : 0;
});

// Sync active step with store
watch(activeStep, (newStep) => {
    if (newStep !== setupStore.currentStep) {
        setupStore.currentStep = newStep;
    }
});

// Navigation methods
const goToNextStep = () => {
    setupStore.nextStep();
    const nextRoute = steps.value[setupStore.currentStep]?.route;
    if (nextRoute) {
        router.push(nextRoute);
    }
};

const goToPreviousStep = () => {
    setupStore.previousStep();
    const prevRoute = steps.value[setupStore.currentStep]?.route;
    if (prevRoute) {
        router.push(prevRoute);
    }
};

const goToStep = (index) => {
    if (index <= setupStore.currentStep || setupStore.isStepCompleted(index)) {
        setupStore.goToStep(index);
        const targetRoute = steps.value[index]?.route;
        if (targetRoute) {
            router.push(targetRoute);
        }
    }
};

// Show/hide buttons based on current step
const showBackButton = computed(() => {
    return setupStore.canGoBack && activeStep.value !== steps.value.length - 1;
});

const showNextButton = computed(() => {
    if (activeStep.value >= steps.value.length - 1) {
        return false; // Last step doesn't have Next button
    }

    // Check if we're on the Plex step (index 1) and validate connection
    if (activeStep.value === 1) {
        return setupStore.isPlexStepValid && setupStore.isConnectionValidated;
    }

    // Check if we're on the User step (index 2) and validate user selection
    if (activeStep.value === 2) {
        return setupStore.isUserSelected;
    }

    return setupStore.canGoNext;
});

const showFinishButton = computed(() => {
    return activeStep.value === steps.value.length - 1;
});

const isFinishing = ref(false);

const finishSetup = async () => {
    isFinishing.value = true;

    try {
        // Save Plex token to OS keychain
        if (setupStore.plexToken) {
            await SavePlexToken(setupStore.plexToken);
        }

        // Mark setup as complete in backend (persists to config.json and starts polling)
        await CompleteSetup();

        // Also update frontend store state
        setupStore.completeSetup();

        // Navigate to dashboard
        router.push('/');

        toast.add({
            severity: 'success',
            summary: 'Setup Complete',
            detail: 'PlexCord has been configured successfully',
            life: 3000
        });
    } catch (error) {
        console.error('Failed to complete setup:', error);
        toast.add({
            severity: 'error',
            summary: 'Setup Failed',
            detail: error?.message || 'Failed to complete setup. Please try again.',
            life: 5000
        });
    } finally {
        isFinishing.value = false;
    }
};

// Skip setup functionality
const showSkipLink = computed(() => {
    // Show skip link on all steps except welcome (0) and complete (last)
    return activeStep.value > 0 && activeStep.value < steps.value.length - 1;
});

const isSkipping = ref(false);

const skipSetup = async () => {
    isSkipping.value = true;

    try {
        // Save current progress via store (already persisted to localStorage)
        setupStore.saveState();

        // Save Plex token to OS keychain if available
        if (setupStore.plexToken) {
            await SavePlexToken(setupStore.plexToken);
        }

        // Mark setup as skipped in backend
        await SkipSetup();

        // Navigate to dashboard
        router.push('/');

        toast.add({
            severity: 'info',
            summary: 'Setup Skipped',
            detail: 'You can complete setup later from Settings',
            life: 4000
        });
    } catch (error) {
        console.error('Failed to skip setup:', error);
        toast.add({
            severity: 'error',
            summary: 'Skip Failed',
            detail: error?.message || 'Failed to skip setup. Please try again.',
            life: 5000
        });
    } finally {
        isSkipping.value = false;
    }
};

// Keyboard navigation - respects same validation as UI buttons
const handleKeydown = (event) => {
    if (event.key === 'ArrowRight' && showNextButton.value) {
        goToNextStep();
    } else if (event.key === 'ArrowLeft' && showBackButton.value) {
        goToPreviousStep();
    }
};

// Load saved state and setup keyboard listeners on mount
onMounted(() => {
    setupStore.loadState();
    // Navigate to saved step if different from current
    if (setupStore.currentStep !== activeStep.value) {
        const savedRoute = steps.value[setupStore.currentStep]?.route;
        if (savedRoute && savedRoute !== route.path) {
            router.push(savedRoute);
        }
    }
    // Setup keyboard navigation
    window.addEventListener('keydown', handleKeydown);
});

// Cleanup on unmount
onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
    <div class="setup-wizard-container">
        <Card class="setup-wizard-card">
            <template #header>
                <div class="wizard-header">
                    <h1 class="text-4xl font-bold text-center mb-2">PlexCord Setup</h1>
                    <p class="text-center text-muted-color">Complete the steps below to get started</p>
                </div>
            </template>

            <template #content>
                <!-- Step Indicator -->
                <div class="steps-container mb-6">
                    <Steps
                        :model="steps"
                        :activeStep="activeStep"
                        :readonly="false"
                        @step-select="(event) => goToStep(event.index)"
                    >
                        <template #item="{ item, index }">
                            <span class="step-label">{{ item.label }}</span>
                        </template>
                    </Steps>
                </div>

                <!-- Step Content -->
                <div class="step-view-container">
                    <router-view />
                </div>

                <!-- Navigation Buttons -->
                <div class="navigation-buttons">
                    <Button
                        v-if="showBackButton"
                        label="Back"
                        icon="pi pi-arrow-left"
                        severity="secondary"
                        @click="goToPreviousStep"
                        class="mr-2"
                    />
                    <span class="flex-grow-1"></span>
                    <Button
                        v-if="showNextButton"
                        label="Next"
                        icon="pi pi-arrow-right"
                        iconPos="right"
                        @click="goToNextStep"
                    />
                    <Button
                        v-if="showFinishButton"
                        label="Finish Setup"
                        icon="pi pi-check"
                        iconPos="right"
                        @click="finishSetup"
                        :loading="isFinishing"
                        :disabled="isFinishing"
                    />
                </div>

                <!-- Keyboard Hint -->
                <div class="keyboard-hint text-center mt-4">
                    <small class="text-muted-color">
                        <i class="pi pi-info-circle mr-1"></i>
                        Use arrow keys to navigate between steps
                    </small>
                </div>

                <!-- Skip Link -->
                <div v-if="showSkipLink" class="skip-link text-center mt-3">
                    <a
                        href="#"
                        @click.prevent="skipSetup"
                        class="text-muted-color hover:text-primary"
                        :class="{ 'pointer-events-none opacity-50': isSkipping }"
                    >
                        <small>
                            <i v-if="isSkipping" class="pi pi-spin pi-spinner mr-1"></i>
                            {{ isSkipping ? 'Skipping...' : 'Skip for now' }}
                        </small>
                    </a>
                </div>
            </template>
        </Card>
    </div>
</template>

<style scoped>
.setup-wizard-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    background: var(--surface-ground);
}

.setup-wizard-card {
    width: 100%;
    max-width: 900px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.wizard-header {
    padding: 2rem 2rem 1rem 2rem;
    background: linear-gradient(180deg, var(--surface-card) 0%, var(--surface-ground) 100%);
}

.steps-container {
    padding: 0 2rem;
}

.step-view-container {
    min-height: 300px;
    padding: 2rem;
}

.navigation-buttons {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 2rem 1rem 2rem;
    border-top: 1px solid var(--surface-border);
    padding-top: 1.5rem;
}

.keyboard-hint {
    padding: 0 2rem 0.5rem 2rem;
}

/* Responsive adjustments */
@media (max-width: 768px) {
    .setup-wizard-container {
        padding: 1rem;
    }

    .wizard-header {
        padding: 1.5rem 1rem 0.5rem 1rem;
    }

    .wizard-header h1 {
        font-size: 2rem;
    }

    .steps-container {
        padding: 0 1rem;
    }

    .step-view-container {
        padding: 1rem;
    }

    .navigation-buttons {
        padding: 0 1rem 0.5rem 1rem;
    }
}

/* Dark mode compatibility */
:deep(.p-steps) {
    background: transparent;
}
</style>
