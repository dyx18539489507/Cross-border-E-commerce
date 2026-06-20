<template>
  <nav class="stage-nav" aria-label="项目功能导航">
    <router-link v-for="item in items" :key="item.path" :to="item.path">{{ item.label }}</router-link>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ projectId: string | number }>()
const items = computed(() => {
  const base = `/projects/${props.projectId}`
  return [
    { label: '项目总览', path: base },
    { label: '合规分析', path: `${base}/compliance` },
    { label: '营销脚本', path: `${base}/script` },
    { label: '营销素材', path: `${base}/assets` },
    { label: '视频剪辑', path: `${base}/editor` },
    { label: '生成任务', path: `${base}/tasks` }
  ]
})
</script>

<style scoped>
.stage-nav{display:flex;gap:8px;overflow-x:auto;padding:7px;border:1px solid var(--border-primary);border-radius:16px;background:var(--bg-card);margin-bottom:24px}.stage-nav a{white-space:nowrap;padding:10px 15px;border-radius:11px;text-decoration:none;color:var(--text-secondary);font-weight:700}.stage-nav a.router-link-exact-active,.stage-nav a:hover{color:var(--accent);background:var(--accent-light)}
</style>
