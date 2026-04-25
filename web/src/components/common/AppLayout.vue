<template>
  <div class="app-layout">
    <!-- Global Header -->
    <header class="app-header">
      <div class="header-content">
        <div class="header-left">
          <router-link to="/dramas" class="logo">
            <span class="logo-text">🎬 {{ t('app.name') }}</span>
          </router-link>
        </div>
        <div class="header-right">
          <LanguageSwitcher />
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="app-main">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'

const { t } = useI18n()
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 24px;
}

.app-header {
  position: sticky;
  top: 18px;
  z-index: 100;
  border: 1px solid var(--border-primary);
  border-radius: 28px;
  background: var(--bg-card);
  box-shadow: var(--shadow-card);
  backdrop-filter: blur(24px);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  min-height: 76px;
  padding: 14px 22px;
  max-width: 100%;
  margin: 0 auto;
}
.header-btn {
  border-radius: var(--radius-lg);
  font-weight: 500;
}

.header-btn.primary {
  background: linear-gradient(135deg, var(--accent) 0%, #0284c7 100%);
  border: none;
  box-shadow: 0 4px 14px rgba(14, 165, 233, 0.35);
}

.header-btn.primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(14, 165, 233, 0.45);
}
.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.logo {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 700;
  font-size: 1.125rem;
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
  font-size: 1.75rem;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.app-main {
  flex: 1;
  padding-top: 28px;
}

/* Dark mode adjustments */
.dark .app-header {
  background: var(--bg-card);
}
</style>
