<script setup>
import { onMounted } from 'vue';
import Toast from 'primevue/toast';

onMounted(() => {
    // Initialize dark mode
    document.documentElement.classList.add('dark');
});
</script>

<template>
    <router-view />
    <!-- Single global Toast host so useToast() emissions from Settings,
         SetupWizard etc. actually render somewhere. AppLayout used to
         render `<Toast />` without importing it, which silently fell
         back to a no-op native element. -->
    <Toast />
</template>

<style scoped></style>
