<template>
  <el-dialog
    v-model="visible"
    append-to-body
    align-center
    width="min(92vw, 420px)"
    class="mobile-access-reminder-dialog"
    modal-class="mobile-access-reminder-overlay"
    :show-close="false"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :z-index="3000"
  >
    <div class="mobile-access-reminder">
      <p class="mobile-access-reminder__message">
        {{ $t('mobileAccessReminder.message') }}
      </p>
      <div class="mobile-access-reminder__actions">
        <el-button type="primary" class="mobile-access-reminder__button" @click="dismissReminder">
          {{ $t('mobileAccessReminder.confirm') }}
        </el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { useMobileAccessReminder } from '@/composables/useMobileAccessReminder'

const { visible, dismissReminder } = useMobileAccessReminder()
</script>

<style scoped>
:deep(.mobile-access-reminder-overlay) {
  background: rgba(15, 23, 42, 0.62);
  backdrop-filter: blur(6px);
}

:deep(.mobile-access-reminder-dialog) {
  margin: 0;
  overflow: hidden;
  border-radius: var(--radius-2xl);
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, var(--bg-card) 100%);
  box-shadow: var(--shadow-xl);
}

.dark :deep(.mobile-access-reminder-dialog) {
  border-color: rgba(148, 163, 184, 0.2);
  background: linear-gradient(180deg, rgba(30, 39, 48, 0.98) 0%, var(--bg-card) 100%);
}

:deep(.mobile-access-reminder-dialog .el-dialog__header) {
  display: none;
}

:deep(.mobile-access-reminder-dialog .el-dialog__body) {
  padding: 0;
}

.mobile-access-reminder {
  padding: 28px 24px 24px;
}

.mobile-access-reminder__message {
  margin: 0;
  color: var(--text-primary);
  font-size: 0.95rem;
  line-height: 1.75;
  text-align: center;
}

.mobile-access-reminder__actions {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

.mobile-access-reminder__button {
  min-width: 140px;
  padding-inline: 24px;
}

@media (max-width: 640px) {
  .mobile-access-reminder {
    padding: 24px 20px 20px;
  }

  .mobile-access-reminder__message {
    font-size: 0.9375rem;
  }

  .mobile-access-reminder__actions {
    width: 100%;
  }

  .mobile-access-reminder__button {
    width: 100%;
  }
}
</style>
