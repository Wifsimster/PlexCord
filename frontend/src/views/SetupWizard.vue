<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick, provide } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useSetupStore } from '@/stores/setup';
import { SavePlexToken, SkipSetup } from '../../wailsjs/go/main/App';
import { useToast } from 'primevue/usetoast';
import BrandMark from '@/components/BrandMark.vue';
import DrawnCheck from '@/components/setup/DrawnCheck.vue';

const toast = useToast();
const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const setupStore = useSetupStore();

setupStore.loadState();
const savedStep = setupStore.currentStep;

const steps = [
    { labelKey: 'wizard.stepWelcome', route: '/setup/welcome' },
    { labelKey: 'wizard.stepPlex', route: '/setup/plex', brand: 'plex' },
    { labelKey: 'wizard.stepUser', route: '/setup/user' },
    { labelKey: 'wizard.stepDiscord', route: '/setup/discord', brand: 'discord' },
    { labelKey: 'wizard.stepDone', route: '/setup/complete' }
];
const lastIndex = steps.length - 1;

const activeIndex = computed(() => {
    const idx = steps.findIndex((s) => s.route === route.path);
    return idx === -1 ? 0 : idx;
});

watch(
    activeIndex,
    (idx) => {
        if (idx !== setupStore.currentStep) {
            setupStore.currentStep = idx;
        }
    },
    { immediate: true }
);

const stepTransition = ref('pc-step-next');
watch(activeIndex, (next, prev) => {
    stepTransition.value = next >= prev ? 'pc-step-next' : 'pc-step-prev';
});

const isStepDone = (index) => index !== activeIndex.value && (index < activeIndex.value || setupStore.isStepCompleted(index));

const canNavigateTo = (index) => index <= activeIndex.value || setupStore.isStepCompleted(index);

const stepSummary = (index) => {
    if (!isStepDone(index)) {
        return '';
    }
    switch (index) {
        case 1:
            return setupStore.plexServerSummary;
        case 2:
            return setupStore.selectedPlexUser?.name || '';
        case 3:
            return setupStore.discordSkipped ? t('wizard.summarySkipped') : setupStore.discordConnected ? t('wizard.summaryConnected') : '';
        default:
            return '';
    }
};

const railList = ref(null);
const progressHeight = ref(0);
const updateProgress = async () => {
    await nextTick();
    const el = railList.value?.querySelector('[aria-current="step"]');
    if (el) {
        progressHeight.value = el.offsetTop + el.offsetHeight;
    }
};
watch([activeIndex, () => steps.map((s, i) => stepSummary(i)).join('|')], updateProgress);

const onRailClick = (index) => {
    if (index === activeIndex.value || !canNavigateTo(index)) {
        return;
    }
    setupStore.goToStep(index);
    router.push(steps[index].route);
};

const goToNextStep = () => {
    if (activeIndex.value >= lastIndex) {
        return;
    }
    setupStore.nextStep();
    const nextRoute = steps[setupStore.currentStep]?.route;
    if (nextRoute) {
        router.push(nextRoute);
    }
};

const goToPreviousStep = () => {
    if (activeIndex.value <= 0) {
        return;
    }
    setupStore.previousStep();
    const prevRoute = steps[setupStore.currentStep]?.route;
    if (prevRoute) {
        router.push(prevRoute);
    }
};

const gate = computed(() => {
    switch (activeIndex.value) {
        case 0:
            return { enabled: true, label: t('wizard.getStarted'), reason: '' };
        case 1:
            return {
                enabled: setupStore.isPlexStepValid && setupStore.isConnectionValidated,
                label: t('wizard.continue'),
                reason: t('wizard.reasonValidate')
            };
        case 2:
            return { enabled: setupStore.isUserSelected, label: t('wizard.continue'), reason: t('wizard.reasonSelectUser') };
        case 3:
            return { enabled: setupStore.isDiscordStepSatisfied, label: t('wizard.continue'), reason: t('wizard.reasonConnectDiscord') };
        default:
            return { enabled: !setupStore.isFinishing, label: t('wizard.finishSetup'), reason: '' };
    }
});

const isLastStep = computed(() => activeIndex.value === lastIndex);

const showDiscordEscape = computed(() => activeIndex.value === 3 && !gate.value.enabled);

const continueWithoutDiscord = () => {
    setupStore.setDiscordSkipped(true);
    goToNextStep();
};

const finishSetup = async () => {
    const ok = await setupStore.finishSetup();
    if (ok) {
        router.push('/');
    }
};

const continueAction = () => {
    if (!gate.value.enabled) {
        return;
    }
    if (isLastStep.value) {
        finishSetup();
    } else {
        goToNextStep();
    }
};

const continueButton = ref(null);
const focusContinue = () => {
    continueButton.value?.focus();
};

const stepPrimary = ref(null);
provide('setupWizard', {
    registerPrimary: (fn) => {
        stepPrimary.value = fn;
    },
    unregisterPrimary: (fn) => {
        if (stepPrimary.value === fn) {
            stepPrimary.value = null;
        }
    },
    next: goToNextStep,
    focusContinue
});

const showSkipLink = computed(() => activeIndex.value > 0 && activeIndex.value < lastIndex);
const isSkipping = ref(false);

const skipSetup = async () => {
    if (isSkipping.value) {
        return;
    }
    isSkipping.value = true;

    try {
        setupStore.saveState();

        if (setupStore.plexToken) {
            await SavePlexToken(setupStore.plexToken);
        }

        await SkipSetup();

        router.push('/');
    } catch (error) {
        console.error('Failed to skip setup:', error);
        toast.add({
            severity: 'error',
            summary: t('wizard.skipFailed'),
            detail: error?.message || t('wizard.skipFailedDetail'),
            life: 8000
        });
    } finally {
        isSkipping.value = false;
    }
};

const handleKeydown = (event) => {
    if (event.ctrlKey || event.metaKey || event.altKey) {
        return;
    }
    if (event.target instanceof Element && event.target.closest('input, textarea, [contenteditable], .p-inputtext')) {
        return;
    }

    if (event.key === 'ArrowRight') {
        if (gate.value.enabled && !isLastStep.value) {
            goToNextStep();
        }
    } else if (event.key === 'ArrowLeft') {
        goToPreviousStep();
    } else if (event.key === 'Enter') {
        if (event.target instanceof Element && event.target.closest('button, a, [role="button"]')) {
            return;
        }
        if (typeof stepPrimary.value === 'function' && stepPrimary.value()) {
            return;
        }
        continueAction();
    }
};

onMounted(() => {
    if (route.path === steps[0].route && savedStep > 0 && savedStep <= lastIndex) {
        setupStore.currentStep = savedStep;
        router.replace(steps[savedStep].route);
    }

    window.addEventListener('keydown', handleKeydown);
    updateProgress();
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
    <div class="wizard">
        <aside class="wizard-rail">
            <div class="rail-header">
                <BrandMark :suffix="$t('wizard.suffix')" />
            </div>

            <nav ref="railList" class="rail-steps" :aria-label="$t('wizard.stepsAria')">
                <div class="rail-progress" :style="{ height: progressHeight + 'px' }" aria-hidden="true"></div>
                <button
                    v-for="(step, i) in steps"
                    :key="step.route"
                    type="button"
                    class="rail-step"
                    :class="{ 'rail-step--current': i === activeIndex, 'rail-step--done': isStepDone(i), 'rail-step--locked': !canNavigateTo(i) }"
                    :aria-current="i === activeIndex ? 'step' : undefined"
                    :aria-disabled="!canNavigateTo(i) || undefined"
                    :tabindex="canNavigateTo(i) ? 0 : -1"
                    v-tooltip.right="canNavigateTo(i) ? null : $t('wizard.lockedTooltip')"
                    @click="onRailClick(i)"
                >
                    <span class="rail-glyph">
                        <DrawnCheck v-if="isStepDone(i)" :size="14" />
                        <span v-else class="rail-dot" :class="i === activeIndex ? 'rail-dot--current' : 'rail-dot--locked'"></span>
                        <span v-if="step.brand" class="rail-tick" :class="`rail-tick--${step.brand}`" aria-hidden="true"></span>
                    </span>
                    <span class="rail-texts">
                        <span class="rail-label">{{ $t(step.labelKey) }}</span>
                        <span v-if="stepSummary(i)" class="rail-summary">{{ stepSummary(i) }}</span>
                    </span>
                </button>
            </nav>

            <div class="rail-foot">
                <a v-if="showSkipLink" href="#" class="rail-skip" :class="{ 'rail-skip--busy': isSkipping }" @click.prevent="skipSetup">
                    <span v-if="isSkipping">{{ $t('wizard.skipping') }}</span>
                    <span v-else>{{ $t('wizard.skip') }}</span>
                </a>
            </div>
        </aside>

        <div class="wizard-main">
            <main class="wizard-content">
                <router-view v-slot="{ Component }">
                    <Transition :name="stepTransition" mode="out-in">
                        <component :is="Component" class="wizard-step" />
                    </Transition>
                </router-view>
            </main>

            <footer class="wizard-footer">
                <button type="button" class="pc-btn pc-btn--ghost" :disabled="activeIndex === 0 || setupStore.isFinishing" @click="goToPreviousStep"><i class="pi pi-arrow-left" aria-hidden="true"></i> {{ $t('wizard.back') }}</button>

                <span class="wizard-footer-gap"></span>

                <span v-if="!gate.enabled && gate.reason" class="wizard-gate-reason" role="status">{{ gate.reason }}</span>
                <a v-if="showDiscordEscape" href="#" class="wizard-escape-link" @click.prevent="continueWithoutDiscord">{{ $t('wizard.continueWithoutDiscord') }}</a>

                <button ref="continueButton" type="button" class="pc-btn pc-btn--primary pc-btn--lg wizard-continue" :disabled="!gate.enabled" @click="continueAction">
                    <i v-if="setupStore.isFinishing && isLastStep" class="pi pi-spin pi-spinner" aria-hidden="true"></i>
                    {{ gate.label }}
                    <i v-if="!isLastStep" class="pi pi-arrow-right" aria-hidden="true"></i>
                    <i v-else-if="!setupStore.isFinishing" class="pi pi-check" aria-hidden="true"></i>
                </button>
            </footer>
        </div>
    </div>
</template>

<style scoped>
.wizard {
    display: grid;
    grid-template-columns: 240px 1fr;
    height: 100vh;
    overflow: hidden;
    background: var(--pc-bg);
}

.wizard-rail {
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: var(--pc-overlay);
    border-right: 1px solid var(--pc-border);
    padding: 16px 0 12px;
}
.rail-header {
    padding: 4px 16px 20px;
}
.rail-steps {
    position: relative;
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding-right: 12px;
}
.rail-progress {
    position: absolute;
    top: 0;
    left: 0;
    width: 2px;
    border-radius: 1px;
    background: var(--pc-accent);
    transition: height var(--pc-dur-3) var(--pc-ease-inout);
}
.rail-step {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    min-height: 40px;
    padding: 10px 8px 10px 20px;
    background: none;
    border: none;
    border-radius: 0 var(--pc-radius-sm) var(--pc-radius-sm) 0;
    text-align: left;
    cursor: pointer;
    transition: background-color var(--pc-dur-1) var(--pc-ease-out);
}
.rail-step:hover:not(.rail-step--locked):not(.rail-step--current) {
    background: var(--pc-raised);
}
.rail-step--locked {
    opacity: 0.4;
    cursor: default;
}
.rail-glyph {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 18px;
    flex: none;
}
.rail-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: var(--pc-radius-full);
}
.rail-dot--current {
    border: 2px solid var(--pc-accent);
    background: transparent;
}
.rail-dot--locked {
    border: 1px solid var(--pc-border-strong);
    background: transparent;
}
.rail-tick {
    position: absolute;
    right: -6px;
    top: 1px;
    width: 6px;
    height: 2px;
    border-radius: 1px;
}
.rail-tick--plex {
    background: var(--pc-plex);
}
.rail-tick--discord {
    background: var(--pc-blurple);
}
.rail-texts {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
}
.rail-label {
    font-size: 13px;
    font-weight: 500;
    line-height: 1.4;
    color: var(--pc-text-secondary);
}
.rail-step--current .rail-label {
    font-weight: 600;
    color: var(--pc-text);
}
.rail-summary {
    font-family: var(--pc-font-mono);
    font-size: 12px;
    color: var(--pc-text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.rail-foot {
    padding: 12px 16px 4px;
}
.rail-skip {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
    text-decoration: none;
    transition: color var(--pc-dur-1) var(--pc-ease-out);
}
.rail-skip:hover {
    color: var(--pc-accent);
}
.rail-skip--busy {
    pointer-events: none;
    opacity: 0.5;
}

.wizard-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
}
.wizard-content {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 40px var(--pc-page-gutter) 32px;
}
.wizard-step {
    max-width: 560px;
}
.wizard-footer {
    flex: none;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px var(--pc-page-gutter);
    border-top: 1px solid var(--pc-border);
}
.wizard-footer-gap {
    flex: 1;
}
.wizard-gate-reason {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.wizard-escape-link {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    text-decoration: none;
}
.wizard-escape-link:hover {
    color: var(--pc-accent);
}
.wizard-continue {
    min-width: 128px;
}
</style>

<style>
.setup-title {
    margin: 0 0 10px;
    font-family: var(--pc-font-display);
    font-size: var(--pc-text-display);
    font-weight: 700;
    line-height: 1.1;
    letter-spacing: -0.03em;
    color: var(--pc-text);
}
.setup-lede {
    margin: 0 0 var(--pc-space-section);
    max-width: 520px;
    font-size: 15px;
    line-height: 1.5;
    color: var(--pc-text-secondary);
}
.setup-panels {
    display: flex;
    flex-direction: column;
    gap: var(--pc-space-panel-gap);
}
</style>
