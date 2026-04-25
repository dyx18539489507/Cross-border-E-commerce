<template>
  <div class="app-shell">
    <main class="app-content">
      <router-view />
    </main>
    <footer v-if="showFooter" class="site-footer">
      <BeianFooter />
    </footer>
    <MobileAccessReminder />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { BeianFooter, MobileAccessReminder } from '@/components/common'

const route = useRoute()
const showFooter = computed(
  () =>
    !['Home', 'Product', 'Pricing', 'About', 'DramaList', 'DramaCreate', 'ComplianceAnalysis', 'DataAnalysis', 'EpisodeWorkflowNew', 'ScriptEdit', 'StoryboardEdit', 'DramaScriptStage', 'Generation', 'TimelineEditor'].includes(
      String(route.name ?? '')
    )
)
</script>

<style>
#app {
  width: 100%;
  min-height: var(--app-vh, 100vh);
  position: relative;
}

.app-shell {
  min-height: inherit;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
}

.app-shell::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background:
    radial-gradient(circle at 8% 12%, rgba(83, 180, 255, 0.2), transparent 24%),
    radial-gradient(circle at 88% 10%, rgba(255, 174, 112, 0.18), transparent 20%),
    radial-gradient(circle at 82% 88%, rgba(135, 102, 255, 0.12), transparent 18%),
    linear-gradient(rgba(139, 160, 196, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(139, 160, 196, 0.05) 1px, transparent 1px);
  background-size:
    auto,
    auto,
    auto,
    96px 96px,
    96px 96px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.46), transparent 88%);
}

.app-content {
  flex: 1 0 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
}

#app .app-content > .page-container {
  flex: 1 0 auto;
  min-height: 100%;
}

.site-footer {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  position: relative;
  z-index: 1;
  padding: 10px 16px calc(16px + env(safe-area-inset-bottom));
}

.dark .app-shell::before {
  background:
    radial-gradient(circle at 8% 12%, rgba(83, 180, 255, 0.14), transparent 24%),
    radial-gradient(circle at 88% 10%, rgba(255, 174, 112, 0.12), transparent 20%),
    radial-gradient(circle at 82% 88%, rgba(135, 102, 255, 0.11), transparent 18%),
    linear-gradient(rgba(139, 160, 196, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(139, 160, 196, 0.04) 1px, transparent 1px);
  background-size:
    auto,
    auto,
    auto,
    96px 96px,
    96px 96px;
}
</style>
