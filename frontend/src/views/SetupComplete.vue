<script setup>
import DiscordPreview from '@/components/setup/DiscordPreview.vue';

// Note: The "Finish Setup" button is handled by SetupWizard.vue
// This component just displays the completion content
</script>

<template>
    <div class="setup-step">
        <div class="step-content">
            <!-- Success Header -->
            <div class="success-header text-center mb-6">
                <div class="success-icon mb-4">
                    <i class="pi pi-check-circle" style="font-size: 3rem; color: var(--green-500)"></i>
                </div>
                <h2 class="text-2xl font-bold mb-2">You're All Set!</h2>
                <p class="text-muted-color">
                    PlexCord is ready to display your Plex music activity on Discord.
                </p>
            </div>

            <!-- Discord Preview Section -->
            <div class="preview-section mb-6">
                <h3 class="text-lg font-semibold mb-4 text-center">Discord Status Preview</h3>
                <DiscordPreview />
            </div>

            <!-- What's Next Section -->
            <div class="next-steps">
                <h3 class="text-lg font-semibold mb-3">What happens next?</h3>
                <ul class="list-none p-0 m-0 space-y-3">
                    <li class="flex items-start gap-3">
                        <i class="pi pi-check-circle text-green-500 mt-1"></i>
                        <span>Your Discord status will automatically update when you play music on Plex</span>
                    </li>
                    <li class="flex items-start gap-3">
                        <i class="pi pi-check-circle text-green-500 mt-1"></i>
                        <span>PlexCord will run in the background and minimize to the system tray</span>
                    </li>
                    <li class="flex items-start gap-3">
                        <i class="pi pi-check-circle text-green-500 mt-1"></i>
                        <span>You can change settings anytime from the dashboard</span>
                    </li>
                </ul>
            </div>

            <!-- Tip Section -->
            <div class="tip-section mt-6">
                <div class="flex items-start gap-3">
                    <i class="pi pi-info-circle text-blue-500 mt-1"></i>
                    <div>
                        <span class="font-semibold">Tip:</span>
                        <span class="text-muted-color ml-1">
                            Start playing music on Plex now to see the live preview update above!
                        </span>
                    </div>
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
    padding: 1rem 0;
}

.success-icon {
    animation: scaleIn 0.3s ease-out;
}

@keyframes scaleIn {
    from {
        transform: scale(0);
    }
    to {
        transform: scale(1);
    }
}

.preview-section {
    background: var(--surface-ground);
    padding: 1.5rem;
    border-radius: 8px;
}

.next-steps {
    background: var(--surface-ground);
    padding: 1.5rem;
    border-radius: 8px;
}

.tip-section {
    background: var(--blue-50);
    padding: 1rem;
    border-radius: 8px;
    border-left: 4px solid var(--blue-500);
}

/* Dark mode adjustments */
:root[data-theme='dark'] .tip-section,
.dark .tip-section {
    background: rgba(59, 130, 246, 0.1);
}
</style>
