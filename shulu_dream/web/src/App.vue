<template>
  <div class="app-shell">
    <main class="app-content">
      <router-view />
    </main>
    <footer v-if="showFooter" class="site-footer">
      <a
        class="site-footer-link"
        href="https://beian.miit.gov.cn"
        target="_blank"
        rel="noopener noreferrer"
      >
        豫ICP备2026007932号-1
      </a>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const showFooter = computed(() => route.name !== 'DramaList')
</script>

<style>
#app {
  width: 100%;
  min-height: var(--app-vh, 100vh);
}

.app-shell {
  min-height: inherit;
  display: flex;
  flex-direction: column;
}

.app-content {
  flex: 1 0 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

#app .app-content > .page-container {
  flex: 1 0 auto;
  min-height: 100%;
}

.site-footer {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  padding: 8px 16px calc(10px + env(safe-area-inset-bottom));
}

.site-footer-link {
  color: var(--text-secondary, #6b7280);
  font-size: 13px;
  line-height: 1.3;
  text-align: center;
  text-decoration: none;
  transition: color 0.2s ease;
}

.site-footer-link:hover {
  color: var(--accent, #0ea5e9);
}
</style>
