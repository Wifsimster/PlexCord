<script setup>
import { ref, computed, inject, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useSetupStore } from '@/stores/setup';
import { GetPlexUsers, SavePlexUserSelection } from '../../wailsjs/go/main/App';
import DrawnCheck from '@/components/setup/DrawnCheck.vue';

const { t } = useI18n();
const router = useRouter();
const setupStore = useSetupStore();
const wizard = inject('setupWizard', null);

const isLoading = ref(false);
const error = ref('');
const autoSelected = ref(false);

const prefersReducedMotion = typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
const autoAdvancePending = ref(false);
let autoAdvanceTimer = null;

const users = computed(() => setupStore.plexUsers);
const selectedUser = computed(() => setupStore.selectedPlexUser);

const cancelAutoAdvance = () => {
    if (autoAdvanceTimer) {
        clearTimeout(autoAdvanceTimer);
        autoAdvanceTimer = null;
    }
    autoAdvancePending.value = false;
};

const scheduleAutoAdvance = () => {
    if (prefersReducedMotion || !wizard) {
        wizard?.focusContinue();
        return;
    }
    autoAdvancePending.value = true;
    autoAdvanceTimer = setTimeout(() => {
        autoAdvanceTimer = null;
        autoAdvancePending.value = false;
        wizard.next();
    }, 800);
};

const fetchUsers = async () => {
    if (!setupStore.plexServerUrl) {
        error.value = t('user.errNoServer');
        return;
    }

    isLoading.value = true;
    error.value = '';
    autoSelected.value = false;

    try {
        const fetchedUsers = await GetPlexUsers(setupStore.plexServerUrl);
        setupStore.setPlexUsers(fetchedUsers || []);

        if (fetchedUsers && fetchedUsers.length === 1) {
            setupStore.selectPlexUser(fetchedUsers[0]);
            autoSelected.value = true;

            try {
                await SavePlexUserSelection(fetchedUsers[0].id, fetchedUsers[0].name);
            } catch (saveErr) {
                console.error('Failed to save auto-selected user to config:', saveErr);
            }

            scheduleAutoAdvance();
        }

        error.value = '';
    } catch (err) {
        console.error('Failed to fetch Plex users:', err);

        let errorMessage = t('user.errLoad');
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

const selectUser = async (user) => {
    cancelAutoAdvance();
    setupStore.selectPlexUser(user);
    autoSelected.value = false;

    try {
        await SavePlexUserSelection(user.id, user.name);
    } catch (err) {
        console.error('Failed to save user selection to config:', err);
    }
};

const goBack = () => {
    router.push('/setup/plex');
};

onMounted(() => {
    if (users.value.length === 0 || !selectedUser.value) {
        fetchUsers();
    } else if (users.value.length === 1 && selectedUser.value) {
        autoSelected.value = true;
    }
});

onUnmounted(() => {
    cancelAutoAdvance();
});
</script>

<template>
    <div>
        <h1 class="setup-title">{{ $t('user.title') }}</h1>
        <p class="setup-lede">{{ $t('user.lede') }}</p>

        <div class="setup-panels">
            <div v-if="isLoading" class="user-grid" :aria-label="$t('user.loadingAria')">
                <div v-for="i in 4" :key="i" class="pc-skeleton user-skeleton"></div>
            </div>

            <section v-else-if="error" class="pc-panel user-error" role="alert">
                <p class="user-error-title"><i class="pi pi-times-circle" aria-hidden="true"></i> {{ $t('user.couldntLoadTitle') }}</p>
                <p class="user-error-text">{{ error }}</p>
                <div class="user-error-actions">
                    <button type="button" class="pc-btn pc-btn--secondary pc-btn--sm" @click="fetchUsers">{{ $t('common.retry') }}</button>
                    <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="goBack">{{ $t('common.back') }}</button>
                </div>
            </section>

            <section v-else-if="users.length === 0" class="pc-panel user-error">
                <p class="user-error-title user-error-title--warn"><i class="pi pi-exclamation-triangle" aria-hidden="true"></i> {{ $t('user.noUsersTitle') }}</p>
                <p class="user-error-text">{{ $t('user.noUsersText') }}</p>
                <div class="user-error-actions">
                    <button type="button" class="pc-btn pc-btn--secondary pc-btn--sm" @click="fetchUsers">{{ $t('common.retry') }}</button>
                </div>
            </section>

            <section v-else-if="users.length === 1 && selectedUser" class="pc-panel user-done">
                <p class="user-done-title">
                    <DrawnCheck :size="14" />
                    <span
                        >{{ $t('user.monitoring') }} <strong>{{ selectedUser.name || $t('user.fallbackName', { id: selectedUser.id }) }}</strong></span
                    >
                </p>
                <p class="user-done-caption">
                    {{ $t('user.onlyOneUser') }}<template v-if="autoAdvancePending">
                        — {{ $t('user.continuing') }}
                        <a href="#" class="user-stay" @click.prevent="cancelAutoAdvance">{{ $t('user.stay') }}</a>
                    </template>
                </p>
            </section>

            <div v-else class="user-grid" role="listbox" :aria-label="$t('user.usersAria')">
                <button v-for="user in users" :key="user.id" type="button" class="user-card" :class="{ 'user-card--selected': selectedUser?.id === user.id }" role="option" :aria-selected="selectedUser?.id === user.id" @click="selectUser(user)">
                    <span class="user-avatar">
                        <img v-if="user.thumb" :src="user.thumb" alt="" class="user-avatar-img" />
                        <i v-else class="pi pi-user" aria-hidden="true"></i>
                    </span>
                    <span class="user-name">{{ user.name || $t('user.fallbackName', { id: user.id }) }}</span>
                    <DrawnCheck v-if="selectedUser?.id === user.id" :size="14" class="user-check" />
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.user-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 12px;
}
.user-skeleton {
    height: 96px;
    border-radius: var(--pc-radius-md);
}

.user-card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 16px 12px;
    background: var(--pc-raised);
    border: 1px solid var(--pc-border);
    border-radius: var(--pc-radius-md);
    cursor: pointer;
    color: var(--pc-text);
    transition:
        border-color var(--pc-dur-2) var(--pc-ease-out),
        background-color var(--pc-dur-2) var(--pc-ease-out);
}
.user-card:hover {
    border-color: var(--pc-border-strong);
}
.user-card--selected {
    border-color: var(--pc-accent);
    background: var(--pc-accent-dim);
}
.user-avatar {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: var(--pc-radius-full);
    overflow: hidden;
    background: var(--pc-surface-700);
    color: var(--pc-text-muted);
}
:root:not(.dark) .user-avatar {
    background: var(--pc-surface-200);
}
.user-avatar-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}
.user-name {
    font-size: var(--pc-text-body);
    font-weight: 500;
    max-width: 100%;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.user-check {
    position: absolute;
    top: 8px;
    right: 8px;
}

.user-done-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 4px;
    font-size: var(--pc-text-body);
    color: var(--pc-text);
}
.user-done-caption {
    margin: 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.user-stay {
    color: var(--pc-accent);
    text-decoration: none;
}
.user-stay:hover {
    text-decoration: underline;
}

.user-error-title {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 4px;
    font-size: var(--pc-text-body);
    font-weight: 600;
    color: var(--pc-danger);
}
.user-error-title--warn {
    color: var(--pc-warn);
}
.user-error-text {
    margin: 0 0 12px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    overflow-wrap: anywhere;
}
.user-error-actions {
    display: flex;
    gap: 8px;
}
</style>
