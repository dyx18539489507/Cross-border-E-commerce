<!--
/**
 * 模块说明：丝路 Agent 自动任务链展示组件。
 * 业务场景：用户需要理解 Agent 从商品理解到投放优化的拆解顺序，而不只是等待一个最终结果。
 * 核心职责：根据完成节点数展示六个业务任务的进度，并在失败或进行中状态下给出稳定视觉反馈。
 */
-->
<template>
  <section class="task-section" aria-labelledby="task-chain-title">
    <div class="task-section__header">
      <h2 id="task-chain-title">
        <Lightning class="task-section__title-icon" aria-hidden="true" />
        <span>Agent 自动任务链</span>
      </h2>

      <span class="task-section__progress">
        进度
        <strong>{{ completedCount }}/{{ tasks.length }}</strong>
      </span>
    </div>

    <div class="task-chain" aria-label="Agent 自动任务链完成进度">
      <div class="task-chain__line" aria-hidden="true"></div>
      <ol class="task-chain__list">
        <li
          v-for="(task, index) in tasks"
          :key="task"
          class="task-chain__item"
          :class="{ 'is-active': index < completedCount, 'is-running': index === completedCount && !failed }"
          :style="{ '--delay': `${index * 90}ms` }"
        >
          <span class="task-chain__node">
            <Check class="task-chain__check" aria-hidden="true" />
          </span>
          <span class="task-chain__label">{{ task }}</span>
        </li>
      </ol>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Check, Lightning } from '@element-plus/icons-vue'

const tasks = [
  '商品理解',
  '合规风险识别',
  '本地化方向',
  '短视频脚本',
  '数字人方案',
  '投放优化'
]

const props = withDefaults(
  defineProps<{
    activeCount?: number
    failed?: boolean
  }>(),
  {
    activeCount: 6,
    failed: false
  }
)

const completedCount = computed(() => Math.max(0, Math.min(props.activeCount, tasks.length)))
</script>

<style scoped>
.task-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.task-section__header h2 {
  margin: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 18px;
  line-height: 27px;
  font-weight: 700;
}

.task-section__title-icon {
  width: 16px;
  height: 16px;
  color: #7c3aed;
}

.task-section__progress {
  color: #62748e;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 11px;
  line-height: 17px;
  white-space: nowrap;
}

.task-section__progress strong {
  color: #7c3aed;
  font-weight: 700;
}

.task-chain {
  position: relative;
  min-height: 54px;
  padding: 0 4px;
}

.task-chain__line {
  position: absolute;
  left: 16px;
  right: 16px;
  top: 16px;
  height: 2px;
  border-radius: 999px;
  background: #e2e8f0;
}

.task-chain__list {
  position: relative;
  z-index: 1;
  min-width: 0;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  list-style: none;
}

.task-chain__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  animation: taskNodeIn 360ms ease both;
  animation-delay: var(--delay);
}

.task-chain__node {
  width: 32px;
  height: 32px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f8fafc;
  box-shadow: inset 0 0 0 1px #cbd5e1;
  color: #94a3b8;
}

.task-chain__check {
  width: 16px;
  height: 16px;
  opacity: 0;
}

.task-chain__item.is-active .task-chain__node {
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  box-shadow: 0 4px 6px rgba(124, 58, 237, 0.3), 0 2px 4px rgba(124, 58, 237, 0.3);
  color: #ffffff;
}

.task-chain__item.is-active .task-chain__check {
  opacity: 1;
}

.task-chain__item.is-running .task-chain__node {
  background: #ffffff;
  box-shadow: inset 0 0 0 2px #7c3aed, 0 0 0 6px rgba(124, 58, 237, 0.08);
  color: #7c3aed;
  animation: runningPulse 1100ms ease-in-out infinite;
}

.task-chain__label {
  color: #0a2463;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 11px;
  line-height: 14px;
  font-weight: 700;
  text-align: center;
  white-space: nowrap;
}

@keyframes taskNodeIn {
  from {
    opacity: 0;
    transform: translateY(6px) scale(0.9);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes runningPulse {
  0%,
  100% {
    transform: scale(1);
  }

  50% {
    transform: scale(1.08);
  }
}

@media (max-width: 720px) {
  .task-chain {
    overflow-x: auto;
    padding-bottom: 6px;
  }

  .task-chain__line,
  .task-chain__list {
    min-width: 680px;
  }
}

@media (max-width: 520px) {
  .task-section__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }
}
</style>
