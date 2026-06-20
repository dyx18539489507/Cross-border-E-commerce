<template>
  <div class="storyboard-generation">
    <el-empty description="请从营销脚本页生成视频分镜" />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  dramaId: string
  episodeId: string
}>()

defineEmits<{
  storyboardGenerated: []
}>()
</script>

<style scoped lang="scss">
.storyboard-generation {
  padding: 24px;
}
</style>
