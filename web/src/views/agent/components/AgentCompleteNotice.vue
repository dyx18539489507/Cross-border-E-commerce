<!--
/**
 * 模块说明：丝路 Agent 完成提示组件。
 * 业务场景：任务链全部完成后，需要给用户明确反馈并承接自动进入结果页的等待时间。
 * 核心职责：以轻量提示展示“已完成 + 跳转中”状态，不参与任何业务数据修改。
 */
-->
<template>
  <section class="complete-notice" aria-live="polite">
    <div class="complete-notice__message">
      <span class="complete-notice__icon-wrap" aria-hidden="true">
        <Check class="complete-notice__icon" />
      </span>
      <span>{{ message }}</span>
    </div>

    <div class="complete-notice__status">
      <span class="complete-notice__dots" aria-hidden="true">
        <span></span>
        <span></span>
        <span></span>
      </span>
      <span>自动跳转中</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { Check } from '@element-plus/icons-vue'

withDefaults(
  defineProps<{
    message?: string
  }>(),
  {
    message: '所有任务节点已完成，正在进入生成结果页……'
  }
)
</script>

<style scoped>
.complete-notice {
  min-height: 57px;
  padding: 13px 17px;
  border: 1px solid #b9f8cf;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  background: linear-gradient(90deg, #f0fdf4 0%, rgba(236, 254, 255, 0.6) 100%);
  opacity: 0;
  animation: completeNoticeIn 520ms ease 560ms forwards;
}

.complete-notice__message {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #016630;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.complete-notice__icon-wrap {
  width: 28px;
  height: 28px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #05df72 0%, #00bc7d 100%);
  box-shadow: 0 1px 3px rgba(0, 201, 80, 0.3), 0 1px 2px rgba(0, 201, 80, 0.3);
  color: #ffffff;
}

.complete-notice__icon {
  width: 16px;
  height: 16px;
}

.complete-notice__status {
  min-height: 31px;
  padding: 7px 13px;
  border: 1px solid rgba(185, 248, 207, 0.7);
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.7);
  color: #008236;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 11px;
  line-height: 17px;
  font-weight: 700;
  white-space: nowrap;
}

.complete-notice__dots {
  display: inline-flex;
  gap: 4px;
}

.complete-notice__dots span {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: #00bc7d;
  animation: completeDot 900ms ease-in-out infinite;
}

.complete-notice__dots span:nth-child(2) {
  animation-delay: 120ms;
}

.complete-notice__dots span:nth-child(3) {
  animation-delay: 240ms;
}

@keyframes completeNoticeIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes completeDot {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.45;
  }

  40% {
    transform: translateY(-2px);
    opacity: 1;
  }
}

@media (max-width: 640px) {
  .complete-notice {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
  }
}
</style>
