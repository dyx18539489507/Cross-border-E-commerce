<template>
  <MobileAgentTransitionPage v-if="isMobileViewport" />
  <DesktopAgentTransitionPage v-else />
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import DesktopAgentTransitionPage from './DesktopAgentTransitionPage.vue'
import MobileAgentTransitionPage from './MobileAgentTransitionPage.vue'

let mediaQuery: MediaQueryList | undefined
const isMobileViewport = ref(typeof window !== 'undefined' ? window.innerWidth <= 768 : false)

const updateMobileViewport = () => {
  isMobileViewport.value = mediaQuery?.matches ?? window.innerWidth <= 768
}

onMounted(() => {
  mediaQuery = window.matchMedia('(max-width: 768px)')
  updateMobileViewport()
  mediaQuery.addEventListener('change', updateMobileViewport)
})

onBeforeUnmount(() => {
  mediaQuery?.removeEventListener('change', updateMobileViewport)
})
</script>
