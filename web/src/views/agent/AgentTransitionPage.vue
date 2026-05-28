<!--
/**
 * 模块说明：丝路 Agent 过渡页设备分流。
 * 业务场景：同一次 Agent 分析在桌面端和移动端使用不同的信息密度与动效布局。
 * 核心职责：根据视口宽度选择桌面或移动过渡页，不改变 Agent 请求和结果缓存逻辑。
 */
-->
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

/**
 * 功能：同步当前视口是否应使用移动端 Agent 过渡页。
 * 参数：无；内部读取 matchMedia 或窗口宽度。
 * 返回：无返回值；更新 isMobileViewport 以切换对应页面组件。
 */
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
