<script setup>
import { onMounted } from 'vue';
import DiscordPreview from '@/components/setup/DiscordPreview.vue';
import { StartSessionPolling, ConnectDiscord, IsDiscordConnected } from '../../wailsjs/go/main/App';

// Note: The "Finish Setup" button is handled by SetupWizard.vue
// This component displays the completion content with live preview

// Start polling and ensure Discord is connected when component mounts
onMounted(async () => {
    try {
        // Ensure Discord is connected for live preview
        const isConnected = await IsDiscordConnected();
        if (!isConnected) {
            console.log('Discord not connected, connecting now...');
            await ConnectDiscord('');
        }

        await StartSessionPolling();
        console.log('Session polling started for preview');
    } catch (error) {
        // Silently handle error - polling might already be running
        console.log('Setup complete initialization:', error);
    }
});
</script>

<template>
    <div class="max-w-4xl mx-auto">
        <div class="py-4">
            <!-- Success Header -->
            <div class="text-center mb-6">
                <div class="mb-4 animate-[scaleIn_0.3s_ease-out]">
                    <i class="pi pi-check-circle text-5xl text-green-500"></i>
                </div>
                <h2 class="text-2xl font-bold mb-2">You're All Set!</h2>
                <p class="text-surface-600 dark:text-surface-400">PlexCord is ready to display your Plex music activity on Discord.</p>
            </div>

            <!-- Discord Preview Section -->
            <div class="mb-6">
                <h3 class="text-lg font-semibold mb-4 text-center">Discord Status Preview</h3>
                <div class="bg-surface-50 dark:bg-surface-900 p-6 rounded-lg">
                    <DiscordPreview />
                </div>
            </div>

            <!-- What's Next Section -->
            <div class="bg-surface-50 dark:bg-surface-900 p-6 rounded-lg">
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
            <div class="mt-6 bg-blue-50 dark:bg-blue-900/10 p-4 rounded-lg border-l-4 border-blue-500">
                <div class="flex items-start gap-3">
                    <i class="pi pi-info-circle text-blue-500 mt-1"></i>
                    <div>
                        <span class="font-semibold">Tip:</span>
                        <span class="text-surface-600 dark:text-surface-400 ml-1"> Start playing music on Plex now to see the live preview update above! </span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
@keyframes scaleIn {
    from {
        transform: scale(0);
    }
    to {
        transform: scale(1);
    }
}
</style>
