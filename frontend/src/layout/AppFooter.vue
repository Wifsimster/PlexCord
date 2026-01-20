<script setup>
import { ref, onMounted } from 'vue';
import { GetVersion } from '../../wailsjs/go/main/App';

const versionInfo = ref(null);
const versionDisplay = ref('');

onMounted(async () => {
    try {
        versionInfo.value = await GetVersion();
        versionDisplay.value = versionInfo.value.version;
    } catch (error) {
        console.error('Failed to get version:', error);
        versionDisplay.value = 'Unknown';
    }
});
</script>

<template>
    <div class="layout-footer">
        <span 
            class="font-semibold" 
            :title="versionInfo ? `Commit: ${versionInfo.commit}\nBuild Date: ${versionInfo.buildDate}` : ''"
        >
            PlexCord {{ versionDisplay }}
        </span>
    </div>
</template>
