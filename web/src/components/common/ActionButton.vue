<template>
  <!-- Minimalist action button with icon and optional tooltip -->
  <!-- 简约操作按钮，带图标和可选提示 -->
  <el-tooltip
    v-if="tooltip"
    :content="tooltip"
    placement="top"
    :show-after="500"
  >
    <button
      :class="['action-button', variant, { disabled }]"
      :disabled="disabled"
      @click="$emit('click')"
    >
      <el-icon :size="size">
        <component :is="icon" />
      </el-icon>
    </button>
  </el-tooltip>
  <button
    v-else
    :class="['action-button', variant, { disabled }]"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <el-icon :size="size">
      <component :is="icon" />
    </el-icon>
  </button>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

/**
 * ActionButton - Minimalist icon button for actions
 * 操作按钮 - 简约图标按钮用于各种操作
 */
withDefaults(defineProps<{
  icon: Component
  tooltip?: string
  variant?: 'default' | 'primary' | 'danger'
  size?: number
  disabled?: boolean
}>(), {
  variant: 'default',
  size: 16,
  disabled: false
})

defineEmits<{
  click: []
}>()
</script>

<style scoped>
.action-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  padding: 0;
  border: 1px solid rgba(214, 224, 242, 0.88);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.78);
  color: var(--text-secondary);
  box-shadow: 0 18px 34px -30px rgba(34, 62, 109, 0.22);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-button:hover {
  background: rgba(255, 255, 255, 0.96);
  border-color: rgba(140, 183, 255, 0.86);
  color: var(--text-primary);
  transform: translateY(-1px);
}

.action-button:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

.action-button.primary:hover {
  background: rgba(85, 104, 239, 0.12);
  color: var(--theme-indigo);
}

.action-button.danger:hover {
  background: rgba(232, 93, 84, 0.12);
  color: #e85d54;
}

.dark .action-button.danger:hover {
  background: rgba(239, 68, 68, 0.15);
}

.action-button.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-button.disabled:hover {
  background: rgba(255, 255, 255, 0.78);
  color: var(--text-muted);
}
</style>
