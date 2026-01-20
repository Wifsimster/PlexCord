<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useSetupStore } from '@/stores/setup';
import { SavePlexToken, CompleteSetup, SkipSetup } from '../../wailsjs/go/main/App';
import { useToast } from 'primevue/usetoast';
import Stepper from 'primevue/stepper';
import StepList from 'primevue/steplist';
import Step from 'primevue/step';
import StepPanels from 'primevue/steppanels';
import StepPanel from 'primevue/steppanel';
import Button from 'primevue/button';
import Card from 'primevue/card';

const toast = useToast();

const router = useRouter();
const route = useRoute();
const setupStore = useSetupStore();

// Step definitions with numeric values (1-based as per PrimeVue best practices)
const steps = ref([
    { label: 'Welcome', value: 1, route: '/setup/welcome' },
    { label: 'Plex Server', value: 2, route: '/setup/plex' },
    { label: 'Select User', value: 3, route: '/setup/user' },
    { label: 'Discord', value: 4, route: '/setup/discord' },
    { label: 'Complete', value: 5, route: '/setup/complete' }
]);

// Current active step value based on route (1-based)
const activeStepValue = computed(() => {
    const currentPath = route.path;
    const step = steps.value.find(s => s.route === currentPath);
    return step ? step.value : 1;
});

// Sync active step with store (convert from 1-based to 0-based for store)
watch(activeStepValue, (newStepValue) => {
    const storeIndex = newStepValue - 1;
    if (storeIndex !== setupStore.currentStep) {
        setupStore.currentStep = storeIndex;
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

// Navigate using step value (1-based)
const activateStep = (stepValue) => {
    const stepIndex = stepValue - 1;
    if (stepIndex >= 0 && stepIndex < steps.value.length) {
        setupStore.goToStep(stepIndex);
        const targetRoute = steps.value[stepIndex]?.route;
        if (targetRoute) {
            router.push(targetRoute);
        }
    }
};

// Handle step change from stepper clicks
const onStepChange = (newStepValue) => {
    activateStep(newStepValue);
};

// Show/hide buttons based on current step (use 0-based store index)
const currentStoreStep = computed(() => setupStore.currentStep);

const showBackButton = computed(() => {
    return setupStore.canGoBack && currentStoreStep.value !== steps.value.length - 1;
});

const showNextButton = computed(() => {
    if (currentStoreStep.value >= steps.value.length - 1) {
        return false; // Last step doesn't have Next button
    }

    // Check if we're on the Plex step (store index 1) and validate connection
    if (currentStoreStep.value === 1) {
        return setupStore.isPlexStepValid && setupStore.isConnectionValidated;
    }

    // Check if we're on the User step (store index 2) and validate user selection
    if (currentStoreStep.value === 2) {
        return setupStore.isUserSelected;
    }

    return setupStore.canGoNext;
});

const showFinishButton = computed(() => {
    return currentStoreStep.value === steps.value.length - 1;
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
    return currentStoreStep.value > 0 && currentStoreStep.value < steps.value.length - 1;
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
    if (setupStore.currentStep !== currentStoreStep.value) {
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
                    <div class="header-title-container">
                        <!-- PlexAmp Logo (Left) -->
                        <svg class="header-logo" width="40" height="40" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
                            <polyline fill="none" stroke="#CC7B19" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" points="4.5 24 23.444 24 12.808 9.342 16.883 9.342 27.519 24 16.883 38.658 20.957 38.658 31.594 24 20.957 9.342 25.032 9.342 35.668 24 25.032 38.658 29.107 38.658 39.743 24 43.5 24"/>
                        </svg>
                        
                        <h1 class="text-4xl font-bold text-center mb-2">PlexCord Setup</h1>
                        
                        <!-- Discord Logo (Right) -->
                        <svg class="header-logo" width="40" height="40" viewBox="0 0 71 55" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M60.1045 4.8978C55.5792 2.8214 50.7265 1.2916 45.6527 0.41542C45.5603 0.39851 45.468 0.440769 45.4204 0.525289C44.7963 1.6353 44.105 3.0834 43.6209 4.2216C38.1637 3.4046 32.7345 3.4046 27.3892 4.2216C26.905 3.0581 26.1886 1.6353 25.5617 0.525289C25.5141 0.443589 25.4218 0.40133 25.3294 0.41542C20.2584 1.2888 15.4057 2.8186 10.8776 4.8978C10.8384 4.9147 10.8048 4.9429 10.7825 4.9795C1.57795 18.7309 -0.943561 32.1443 0.293408 45.3914C0.299005 45.4562 0.335386 45.5182 0.385761 45.5576C6.45866 50.0174 12.3413 52.7249 18.1147 54.5195C18.2071 54.5477 18.305 54.5139 18.3638 54.4378C19.7295 52.5728 20.9469 50.6063 21.9907 48.5383C22.0523 48.4172 21.9935 48.2735 21.8676 48.2256C19.9366 47.4931 18.0979 46.6 16.3292 45.5858C16.1893 45.5041 16.1781 45.304 16.3068 45.2082C16.679 44.9293 17.0513 44.6391 17.4067 44.3461C17.471 44.2926 17.5606 44.2813 17.6362 44.3151C29.2558 49.6202 41.8354 49.6202 53.3179 44.3151C53.3935 44.2785 53.4831 44.2898 53.5502 44.3433C53.9057 44.6363 54.2779 44.9293 54.6529 45.2082C54.7816 45.304 54.7732 45.5041 54.6333 45.5858C52.8646 46.6197 51.0259 47.4931 49.0921 48.2228C48.9662 48.2707 48.9102 48.4172 48.9718 48.5383C50.038 50.6034 51.2554 52.5699 52.5959 54.435C52.6519 54.5139 52.7526 54.5477 52.845 54.5195C58.6464 52.7249 64.529 50.0174 70.6019 45.5576C70.6551 45.5182 70.6887 45.459 70.6943 45.3942C72.1747 30.0791 68.2147 16.7757 60.1968 4.9823C60.1772 4.9429 60.1437 4.9147 60.1045 4.8978ZM23.7259 37.3253C20.2276 37.3253 17.3451 34.1136 17.3451 30.1693C17.3451 26.225 20.1717 23.0133 23.7259 23.0133C27.308 23.0133 30.1626 26.2532 30.1066 30.1693C30.1066 34.1136 27.28 37.3253 23.7259 37.3253ZM47.3178 37.3253C43.8196 37.3253 40.9371 34.1136 40.9371 30.1693C40.9371 26.225 43.7636 23.0133 47.3178 23.0133C50.9 23.0133 53.7545 26.2532 53.6986 30.1693C53.6986 34.1136 50.9 37.3253 47.3178 37.3253Z" fill="#5865F2"/>
                        </svg>
                    </div>
                </div>
            </template>

            <template #content>
                <!-- Stepper Component -->
                <Stepper :value="activeStepValue" @update:value="onStepChange" class="basis-full">
                    <!-- Step Headers -->
                    <StepList>
                        <Step
                            v-for="step in steps"
                            :key="step.value"
                            :value="step.value"
                        >
                            {{ step.label }}
                        </Step>
                    </StepList>

                    <!-- Step Content Panels -->
                    <StepPanels>
                        <StepPanel
                            v-for="step in steps"
                            :key="step.value"
                            :value="step.value"
                        >
                            <!-- Step View Content -->
                            <div class="step-view-container">
                                <router-view />
                            </div>

                            <!-- Navigation Buttons -->
                            <div class="navigation-buttons">
                                <Button
                                    v-if="showBackButton && step.value === activeStepValue"
                                    label="Back"
                                    icon="pi pi-arrow-left"
                                    severity="secondary"
                                    @click="goToPreviousStep"
                                    class="mr-2"
                                />
                                <span class="grow"></span>
                                <Button
                                    v-if="showNextButton && step.value === activeStepValue"
                                    label="Next"
                                    icon="pi pi-arrow-right"
                                    iconPos="right"
                                    @click="goToNextStep"
                                />
                                <Button
                                    v-if="showFinishButton && step.value === activeStepValue"
                                    label="Finish Setup"
                                    icon="pi pi-check"
                                    iconPos="right"
                                    @click="finishSetup"
                                    :loading="isFinishing"
                                    :disabled="isFinishing"
                                />
                            </div>
                        </StepPanel>
                    </StepPanels>
                </Stepper>

                <!-- Skip Link -->
                <div v-if="showSkipLink" class="text-center mt-3">
                    <a
                        href="#"
                        @click.prevent="skipSetup"
                        class="text-surface-600 dark:text-surface-400 hover:text-primary-500 transition-colors"
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
    max-width: 1200px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.wizard-header {
    padding: 2rem 2rem 1rem 2rem;
    background: linear-gradient(180deg, var(--surface-card) 0%, var(--surface-ground) 100%);
}

.header-title-container {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1.5rem;
    margin-bottom: 0.5rem;
}

.header-logo {
    flex-shrink: 0;
}

.header-title-container h1 {
    margin: 0;
}

/* Responsive adjustments */
@media (max-width: 768px) {
    .setup-wizard-container {
        padding: 1rem;
    }

    .wizard-header {
        padding: 1.5rem 1rem 0.5rem 1rem;
    }

    .header-title-container {
        gap: 1rem;
    }

    .header-logo {
        width: 32px;
        height: 32px;
    }

    .header-title-container h1 {
        font-size: 2rem;
    }

    .step-view-container {
        padding: 1rem 0;
    }

    .navigation-buttons {
        padding-top: 1rem;
        margin-top: 1rem;
    }
}

.step-view-container {
    min-height: 300px;
    padding: 2rem 0;
}

.navigation-buttons {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 1.5rem;
    margin-top: 1.5rem;
    border-top: 1px solid var(--surface-border);
}

/* Ensure stepper takes full width */
:deep(.p-stepper) {
    width: 100%;
}

/* Dark mode compatibility */
:deep(.p-stepper),
:deep(.p-steplist),
:deep(.p-step) {
    background: transparent;
}
</style>
