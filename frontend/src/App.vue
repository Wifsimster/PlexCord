<script setup>
import ConfirmDialog from 'primevue/confirmdialog';
import Toast from 'primevue/toast';
</script>

<template>
    <router-view />
    <!-- Single global Toast host so useToast() emissions from Settings,
         SetupWizard etc. actually render somewhere. AppLayout used to
         render `<Toast />` without importing it, which silently fell
         back to a no-op native element. -->
    <Toast />
    <!-- Global ConfirmDialog host for useConfirm() (destructive confirms). -->
    <ConfirmDialog />
</template>

<style scoped></style>
