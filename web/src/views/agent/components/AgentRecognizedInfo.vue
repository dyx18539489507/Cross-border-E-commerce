<!--
/**
 * 模块说明：丝路 Agent 已识别信息展示组件。
 * 业务场景：结果生成前后都需要让用户确认商品、类目、目标市场、平台、人群和卖点是否被正确理解。
 * 核心职责：优先展示后端 AgentResult，缺失时回退到用户输入，避免识别中状态出现空卡片。
 */
-->
<template>
  <section class="recognized-section" aria-labelledby="recognized-title">
    <div class="recognized-section__header">
      <h2 id="recognized-title">
        <Cpu class="recognized-section__title-icon" aria-hidden="true" />
        <span>Agent 已识别信息</span>
      </h2>

      <span class="recognized-section__badge" :class="{ 'is-pending': !recognized }">
        <CircleCheck class="recognized-section__badge-icon" aria-hidden="true" />
        <span>{{ recognized ? '识别完成' : '识别中' }}</span>
      </span>
    </div>

    <div class="recognized-grid">
      <article v-for="item in recognizedItems" :key="item.label" class="recognized-card">
        <div class="recognized-card__label">
          <component :is="item.icon" class="recognized-card__icon" aria-hidden="true" />
          <span>{{ item.label }}</span>
        </div>
        <strong>{{ item.value }}</strong>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed } from 'vue'
import {
  CircleCheck,
  CollectionTag,
  Cpu,
  Goods,
  Location,
  Platform,
  PriceTag,
  User
} from '@element-plus/icons-vue'
import type { AgentInput, AgentResult } from '@/types/agent'

type RecognizedItem = {
  label: string
  value: string
  icon: Component
}

const props = defineProps<{
  input?: AgentInput | null
  result?: AgentResult | null
}>()

const info = computed(() => props.result?.recognizedInfo)
const recognized = computed(() => Boolean(props.result))

/**
 * 功能：为识别字段提供业务占位。
 * 参数：value 为后端结果或用户输入字段；fallbackValue 为页面兜底文案。
 * 返回：清洗后的字段值，缺失时返回兜底文案。
 */
const fallback = (value: string | undefined, fallbackValue: string) => {
  const trimmed = value?.trim()
  return trimmed || fallbackValue
}

const recognizedItems = computed<RecognizedItem[]>(() => [
  { label: '商品', value: fallback(info.value?.productName, fallback(props.input?.productName, '待识别')), icon: Goods },
  { label: '商品类目', value: fallback(info.value?.category, fallback(props.input?.category, '待识别')), icon: CollectionTag },
  { label: '目标市场', value: fallback(info.value?.targetMarket, fallback(props.input?.targetMarket, '待识别')), icon: Location },
  { label: '目标平台', value: fallback(info.value?.targetPlatform, fallback(props.input?.targetPlatform, '待识别')), icon: Platform },
  { label: '目标人群', value: fallback(info.value?.targetAudience, fallback(props.input?.targetAudience, '待识别')), icon: User },
  {
    label: '核心卖点',
    value: (info.value?.coreSellingPoints?.length ? info.value.coreSellingPoints : props.input?.coreSellingPoints)?.join(' / ') || '待识别',
    icon: PriceTag
  }
])
</script>

<style scoped>
.recognized-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recognized-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.recognized-section__header h2 {
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

.recognized-section__title-icon {
  width: 16px;
  height: 16px;
  color: #06b6d4;
}

.recognized-section__badge {
  min-height: 21px;
  padding: 3px 8px;
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(6, 182, 212, 0.1);
  color: #0891b2;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  white-space: nowrap;
}

.recognized-section__badge-icon {
  width: 12px;
  height: 12px;
}

.recognized-section__badge.is-pending {
  border-color: rgba(124, 58, 237, 0.22);
  background: rgba(124, 58, 237, 0.08);
  color: #7c3aed;
}

.recognized-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.recognized-card {
  min-height: 62px;
  padding: 12px 13px 11px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: linear-gradient(166deg, #f8fafc 0%, #ffffff 100%);
  box-shadow: 0 8px 20px -20px rgba(15, 23, 42, 0.3);
}

.recognized-card__label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 15px;
  color: #90a1b9;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.recognized-card__icon {
  width: 12px;
  height: 12px;
  color: #7c3aed;
}

.recognized-card strong {
  color: #0a2463;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 18px;
  font-weight: 700;
}

@media (max-width: 760px) {
  .recognized-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 520px) {
  .recognized-section__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .recognized-grid {
    grid-template-columns: 1fr;
  }
}
</style>
