<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useSetupStore } from '@/stores/setup';
import { GetPlexUsers, SavePlexUserSelection } from '../../wailsjs/go/main/App';
import Button from 'primevue/button';
import ProgressSpinner from 'primevue/progressspinner';
import Message from 'primevue/message';
import UserCard from '@/components/UserCard.vue';

const router = useRouter();
const setupStore = useSetupStore();

// Loading and error state
const isLoading = ref(false);
const error = ref('');
const autoSelected = ref(false);

// Computed properties
const users = computed(() => setupStore.plexUsers);
const selectedUser = computed(() => setupStore.selectedPlexUser);
const isUserSelected = computed(() => setupStore.isUserSelected);

// Fetch users from Plex server
const fetchUsers = async () => {
    if (!setupStore.plexServerUrl) {
        error.value = 'No server URL configured. Please go back and select a server.';
        return;
    }

    isLoading.value = true;
    error.value = '';
    autoSelected.value = false;

    try {
        const fetchedUsers = await GetPlexUsers(setupStore.plexServerUrl);
        setupStore.setPlexUsers(fetchedUsers);

        // Auto-select if only one user (AC3)
        if (fetchedUsers.length === 1) {
            setupStore.selectPlexUser(fetchedUsers[0]);
            autoSelected.value = true;

            // Persist auto-selection to Go config
            try {
                await SavePlexUserSelection(fetchedUsers[0].id, fetchedUsers[0].name);
            } catch (saveErr) {
                console.error('Failed to save auto-selected user to config:', saveErr);
            }
        }

        error.value = '';
    } catch (err) {
        console.error('Failed to fetch Plex users:', err);

        // Extract user-friendly error message
        let errorMessage = 'Failed to load users from Plex server';
        if (err && typeof err === 'string') {
            errorMessage = err;
        } else if (err && err.message) {
            errorMessage = err.message;
        }

        error.value = errorMessage;
        setupStore.setPlexUsers([]);
    } finally {
        isLoading.value = false;
    }
};

// Select a user and persist to config
const selectUser = async (user) => {
    setupStore.selectPlexUser(user);
    autoSelected.value = false; // Clear auto-select message if user manually selects

    // Persist selection to Go config for application restart persistence
    try {
        await SavePlexUserSelection(user.id, user.name);
    } catch (err) {
        console.error('Failed to save user selection to config:', err);
        // Don't show error to user - localStorage persistence still works
    }
};

// Navigation helper for error state
const goBack = () => {
    router.push('/setup/plex');
};

// Fetch users on mount if not already loaded
onMounted(() => {
    setupStore.loadState();

    // If we already have users and a selection, don't refetch
    if (users.value.length === 0 || !selectedUser.value) {
        fetchUsers();
    } else if (users.value.length === 1 && selectedUser.value) {
        // Restore auto-selected state
        autoSelected.value = true;
    }
});
</script>

<template>
    <div class="setup-step">
        <div class="step-header">
            <h2 class="text-2xl font-bold mb-2">Select User Account</h2>
            <p class="text-muted-color">
                Choose which Plex user account to monitor for playback activity
            </p>
        </div>

        <div class="step-content mt-6">
            <!-- Loading State -->
            <div v-if="isLoading" class="loading-container">
                <ProgressSpinner
                    style="width: 50px; height: 50px"
                    strokeWidth="4"
                    fill="transparent"
                    animationDuration="1s"
                />
                <p class="text-muted-color mt-3">Loading users...</p>
            </div>

            <!-- Error State -->
            <div v-else-if="error" class="error-container">
                <Message severity="error" :closable="false" class="w-full mb-4">
                    <template #icon>
                        <i class="pi pi-times-circle text-2xl"></i>
                    </template>
                    <div class="error-content">
                        <h4 class="font-semibold mb-2">Failed to Load Users</h4>
                        <p class="text-sm">{{ error }}</p>
                    </div>
                </Message>
                <div class="error-actions">
                    <Button
                        label="Retry"
                        icon="pi pi-refresh"
                        @click="fetchUsers"
                        severity="danger"
                        class="mr-2"
                    />
                    <Button
                        label="Go Back"
                        icon="pi pi-arrow-left"
                        @click="goBack"
                        severity="secondary"
                        outlined
                    />
                </div>
            </div>

            <!-- No Users Found -->
            <div v-else-if="users.length === 0 && !isLoading" class="no-users-container">
                <Message severity="warn" :closable="false" class="w-full mb-4">
                    <template #icon>
                        <i class="pi pi-info-circle text-2xl"></i>
                    </template>
                    <div>
                        <h4 class="font-semibold mb-2">No Users Found</h4>
                        <p class="text-sm">
                            No user accounts were found on this Plex server.
                            This may happen if the server is configured for admin-only access.
                        </p>
                    </div>
                </Message>
                <div class="no-users-actions">
                    <Button
                        label="Retry"
                        icon="pi pi-refresh"
                        @click="fetchUsers"
                        severity="secondary"
                    />
                </div>
            </div>

            <!-- User List -->
            <div v-else class="users-section">
                <!-- Auto-select info message -->
                <Message v-if="autoSelected" severity="info" :closable="false" class="w-full mb-4">
                    <template #icon>
                        <i class="pi pi-check-circle text-xl"></i>
                    </template>
                    Only one user found - automatically selected
                </Message>

                <div class="user-grid">
                    <UserCard
                        v-for="user in users"
                        :key="user.id"
                        :user="user"
                        :selected="selectedUser?.id === user.id"
                        @select="selectUser"
                    />
                </div>

                <!-- Selection confirmation -->
                <div v-if="selectedUser && !autoSelected" class="selection-info mt-4">
                    <Message severity="success" :closable="false" class="w-full">
                        <template #icon>
                            <i class="pi pi-user text-xl"></i>
                        </template>
                        Monitoring playback for: <strong>{{ selectedUser.name }}</strong>
                    </Message>
                </div>
            </div>
        </div>

    </div>
</template>

<style scoped>
.setup-step {
    max-width: 700px;
    margin: 0 auto;
    padding: 2rem;
}

.step-header {
    text-align: center;
}

.step-content {
    min-height: 300px;
}

.loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 1rem;
}

.error-container {
    text-align: center;
}

.error-content {
    text-align: left;
    margin-left: 0.5rem;
}

.error-actions {
    display: flex;
    justify-content: center;
    gap: 0.5rem;
    margin-top: 1rem;
}

.no-users-container {
    text-align: center;
}

.no-users-actions {
    display: flex;
    justify-content: center;
    margin-top: 1rem;
}

.users-section {
    display: flex;
    flex-direction: column;
}

.user-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 1rem;
    justify-content: center;
}

.selection-info {
    text-align: center;
}

/* Responsive adjustments */
@media (max-width: 600px) {
    .setup-step {
        padding: 1rem;
    }

    .user-grid {
        grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    }

    .error-actions {
        flex-direction: column;
    }

    .error-actions button {
        width: 100%;
    }
}
</style>
