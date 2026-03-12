<template>
  <div class="app-shell">
    <main class="app-content">
      <router-view />
    </main>
    <footer v-if="showFooter" class="site-footer">
      <BeianFooter />
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { BeianFooter } from '@/components/common'

const route = useRoute()
const showFooter = computed(() => !['DramaList', 'EpisodeWorkflowNew'].includes(String(route.name ?? '')))
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
</style>
