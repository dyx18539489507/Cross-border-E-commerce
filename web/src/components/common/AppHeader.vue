<template>
  <div class="app-header-wrapper">
    <header class="app-header" :class="{ 'header-fixed': fixed }">
      <div class="header-content">
        <!-- Left section: Logo + Left slot -->
        <div class="header-left">
          <router-link v-if="showLogo" to="/dramas" class="logo">
            <span class="logo-text">🎬 {{ t('app.name') }}</span>
          </router-link>
          <!-- Left slot for business content | 左侧插槽用于业务内容 -->
          <slot name="left" />
        </div>

        <!-- Center section: Center slot -->
        <div class="header-center">
          <slot name="center" />
        </div>

        <!-- Right section: Actions + Right slot -->
        <div class="header-right">
          <!-- Language Switcher | 语言切换 -->
          <LanguageSwitcher v-if="showLanguage" />

          <!-- Right slot for business content (before actions) | 右侧插槽（在操作按钮前） -->
          <slot name="right" />
        </div>
      </div>
    </header>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'

/**
 * AppHeader - Global application header component
 * 应用顶部头组件
 * 
 * Features | 功能:
 * - Fixed position at top | 固定在顶部
 * - Language switch | 语言切换
 * - Slots support for business content | 支持插槽放置业务内容
 * 
 * Slots | 插槽:
 * - left: Content after logo | logo 右侧内容
 * - center: Center content | 中间内容
 * - right: Content before actions | 操作按钮左侧内容
 */

interface Props {
  /** Fixed position at top | 是否固定在顶部 */
  fixed?: boolean
  /** Show logo | 是否显示 logo */
  showLogo?: boolean
  /** Show language switcher | 是否显示语言切换 */
  showLanguage?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  fixed: true,
  showLogo: true,
  showLanguage: true
})

const { t } = useI18n()
</script>

<style scoped>
.app-header {
  border: 1px solid var(--border-primary);
  border-radius: 28px;
  background: var(--bg-card);
  box-shadow: var(--shadow-card);
  backdrop-filter: blur(24px);
  z-index: 1000;
}

.app-header.header-fixed {
  position: fixed;
  top: 18px;
  left: 24px;
  right: 24px;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  min-height: 78px;
  padding: 14px 22px;
  max-width: 100%;
  margin: 0 auto;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.header-center {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 700;
  font-size: 2rem;
  transition: opacity var(--transition-fast);
}

.logo:hover {
  opacity: 0.8;
}

.logo-text {
  background: linear-gradient(135deg, var(--theme-text) 0%, var(--theme-indigo) 58%, var(--theme-orange) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-btn {
  border-radius: var(--radius-lg);
  font-weight: 700;
}

.header-btn .btn-text {
  margin-left: 4px;
}

/* Dark mode adjustments | 深色模式适配 */
.dark .app-header {
  background: var(--bg-card);
}

/* ========================================
   Common Slot Styles / 插槽通用样式
   ======================================== */

/* Back Button | 返回按钮 */
:deep(.back-btn) {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

:deep(.back-btn:hover) {
  color: var(--text-primary);
  background: var(--bg-hover);
}

:deep(.back-btn .el-icon) {
  font-size: 1rem;
}

/* Page Title | 页面标题 */
:deep(.page-title) {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

:deep(.page-title h1),
:deep(.header-title),
:deep(.drama-title) {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.3;
}

:deep(.page-title .subtitle) {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

/* Episode Title | 章节标题 */
:deep(.episode-title) {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

/* Responsive | 响应式 */
@media (max-width: 768px) {
  .app-header.header-fixed {
    top: 12px;
    left: 16px;
    right: 16px;
    border-radius: 24px;
  }

  .header-content {
    min-height: 70px;
    padding: 12px 16px;
  }
  
  .btn-text {
    display: none;
  }
  
  .header-btn {
    padding: 8px;
  }

  :deep(.page-title h1),
  :deep(.header-title),
  :deep(.drama-title) {
    font-size: 1rem;
  }

  :deep(.back-btn span) {
    display: none;
  }
}
</style>
