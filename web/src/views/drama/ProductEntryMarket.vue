<!--
/**
 * 模块说明：数字丝路商品录入第二步。
 * 业务场景：商品基础信息确定后，需要选择目标国家/市场、销售平台和营销目标。
 * 核心职责：把市场与平台写入统一商品草稿，为后续合规检查和本地化内容生成提供关键上下文。
 */
-->
<template>
  <ProductEntryFlowShell active-step="market">
    <section class="product-entry-card product-entry-card--market">
      <div class="product-entry-card__body">
        <div class="entry-section">
          <h2 class="entry-section__title">选择目标市场</h2>
          <div class="market-grid" role="radiogroup" aria-label="选择目标市场">
            <button
              v-for="market in marketOptions"
              :key="market.code"
              type="button"
              class="market-option"
              :class="{ 'market-option--selected': form.marketCode === market.code }"
              :aria-pressed="form.marketCode === market.code"
              @click="selectMarket(market)"
            >
              <span class="market-option__emoji">{{ market.emoji }}</span>
              <span class="market-option__copy">
                <strong>{{ market.name }}</strong>
                <small>{{ market.code }}</small>
              </span>
            </button>
          </div>
        </div>

        <div class="entry-section">
          <h2 class="entry-section__title">选择销售平台</h2>
          <div class="platform-grid" role="radiogroup" aria-label="选择销售平台">
            <button
              v-for="platform in platformOptions"
              :key="platform.name"
              type="button"
              class="platform-option"
              :class="{ 'platform-option--selected': form.platform === platform.name }"
              :aria-pressed="form.platform === platform.name"
              @click="selectPlatform(platform.name)"
            >
              <span class="platform-option__emoji">{{ platform.emoji }}</span>
              <strong>{{ platform.name }}</strong>
            </button>
          </div>
        </div>

        <div class="entry-section entry-section--compact">
          <div class="entry-section__title-row">
            <h2 class="entry-section__title">营销目标</h2>
            <span class="entry-section__hint">可多选</span>
          </div>
          <div class="goal-list" aria-label="选择营销目标">
            <button
              v-for="goal in marketingGoalOptions"
              :key="goal"
              type="button"
              class="goal-chip"
              :class="{ 'goal-chip--selected': form.marketingGoals.includes(goal) }"
              :aria-pressed="form.marketingGoals.includes(goal)"
              @click="toggleMarketingGoal(goal)"
            >
              {{ goal }}
            </button>
          </div>
        </div>
      </div>

      <p v-if="notice" class="entry-notice">{{ notice }}</p>

      <div class="product-entry-card__footer">
        <button type="button" class="footer-button footer-button--ghost" @click="goPrevious">上一步</button>

        <button type="button" class="footer-button footer-button--primary" @click="goNext">
          <span>下一步</span>
          <img :src="arrowRightIcon" alt="" />
        </button>
      </div>
    </section>
  </ProductEntryFlowShell>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import ProductEntryFlowShell from '@/components/product-entry/ProductEntryFlowShell.vue'
import arrowRightIcon from '@/assets/figma/product-entry/arrow-right.svg'
import { saveCreateDramaDraft } from '@/utils/createDramaDraft'
import {
  buildCreateDramaDraftFromProductEntry,
  readProductEntryDraft,
  writeProductEntryDraft
} from '@/utils/productEntryDraft'
import type { ProductEntryTargetMarket } from '@/utils/productEntryDraft'

interface MarketOption {
  code: string
  name: string
  emoji: string
}

const router = useRouter()
const notice = ref('')

const marketOptions: MarketOption[] = [
  { code: 'US', name: '美国', emoji: '🇺🇸' },
  { code: 'DE', name: '德国', emoji: '🇩🇪' },
  { code: 'JP', name: '日本', emoji: '🇯🇵' },
  { code: 'UK', name: '英国', emoji: '🇬🇧' },
  { code: 'FR', name: '法国', emoji: '🇫🇷' },
  { code: 'AU', name: '澳大利亚', emoji: '🇦🇺' }
]

const platformOptions = [
  { name: 'Amazon', emoji: '📦' },
  { name: 'eBay', emoji: '🛒' },
  { name: 'Shopify', emoji: '🏪' }
] as const

const marketingGoalOptions = ['提升曝光', '提升转化', '新品测试', '品牌种草'] as const

const form = reactive<ProductEntryTargetMarket>({
  marketCode: 'US',
  marketName: '美国',
  marketEmoji: '🇺🇸',
  platform: 'Amazon',
  marketingGoals: []
})

const persistDraft = () => {
  // 目标市场会影响准入法规和内容表达，必须在每次选择后立即沉淀，避免返回上一步时丢失。
  writeProductEntryDraft({
    targetMarket: form.marketName,
    targetPlatform: form.platform,
    marketingGoal: [...form.marketingGoals]
  })

  const createDraft = buildCreateDramaDraftFromProductEntry()
  if (createDraft) {
    saveCreateDramaDraft(createDraft)
  }
}

const selectMarket = (market: MarketOption) => {
  form.marketCode = market.code
  form.marketName = market.name
  form.marketEmoji = market.emoji
  notice.value = ''
  persistDraft()
}

const selectPlatform = (platform: string) => {
  form.platform = platform
  notice.value = ''
  persistDraft()
}

const toggleMarketingGoal = (goal: string) => {
  const index = form.marketingGoals.indexOf(goal)
  if (index >= 0) {
    form.marketingGoals.splice(index, 1)
  } else {
    form.marketingGoals.push(goal)
  }

  persistDraft()
}

const restoreDraft = () => {
  const draft = readProductEntryDraft()
  // Agent 可能返回“东南亚/印尼”等不在快捷选项里的市场，此时保留原始文本而不是强行映射。
  const savedMarket = marketOptions.find((market) => {
    return market.code === draft.targetMarket || market.name === draft.targetMarket
  })

  if (savedMarket) {
    form.marketCode = savedMarket.code
    form.marketName = savedMarket.name
    form.marketEmoji = savedMarket.emoji
  } else if (draft.targetMarket.trim()) {
    form.marketCode = draft.targetMarket
    form.marketName = draft.targetMarket
    form.marketEmoji = ''
  }

  if (draft.targetPlatform.trim()) {
    form.platform = draft.targetPlatform
  }

  if (Array.isArray(draft.marketingGoal)) {
    form.marketingGoals = draft.marketingGoal.filter((goal): goal is string => typeof goal === 'string')
  }

  persistDraft()
}

const validateStep = () => {
  if (!form.marketCode || !form.marketName) {
    notice.value = '请先选择一个目标市场'
    return false
  }

  if (!form.platform) {
    notice.value = '请选择主要销售平台'
    return false
  }

  notice.value = ''
  return true
}

const goPrevious = () => {
  persistDraft()
  router.push('/projects/create')
}

const goNext = () => {
  if (!validateStep()) {
    ElMessage.warning(notice.value)
    return
  }

  persistDraft()
  router.push('/product-entry/details')
}

onMounted(restoreDraft)
</script>

<style scoped>
.product-entry-card {
  width: 960px;
  margin: 36px auto 0;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #ffffff;
  box-shadow:
    0 10px 15px 0 rgba(0, 0, 0, 0.1),
    0 4px 6px 0 rgba(0, 0, 0, 0.1);
}

.product-entry-card__body {
  display: flex;
  flex-direction: column;
  gap: 28px;
  padding: 33px 33px 0;
}

.entry-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.entry-section--compact {
  gap: 12px;
}

.entry-section__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.entry-section__title {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 18px;
  font-weight: 700;
  line-height: 28px;
}

.entry-section__hint {
  color: #90a1b9;
  font-size: 13px;
  line-height: 20px;
}

.market-grid,
.platform-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.market-option,
.platform-option,
.goal-chip {
  border: 2px solid #e2e8f0;
  background: #ffffff;
  color: #0a2463;
  cursor: pointer;
  transition:
    border-color 180ms ease,
    background 180ms ease,
    box-shadow 180ms ease,
    transform 180ms ease;
}

.market-option:hover,
.platform-option:hover,
.goal-chip:hover {
  border-color: #06b6d4;
  transform: translateY(-1px);
}

.market-option--selected,
.platform-option--selected,
.goal-chip--selected {
  border-color: #06b6d4;
  background: linear-gradient(155deg, rgba(6, 182, 212, 0.05) 0%, rgba(124, 58, 237, 0.05) 100%);
  box-shadow:
    0 10px 15px 0 rgba(0, 0, 0, 0.1),
    0 4px 6px 0 rgba(0, 0, 0, 0.1);
}

.market-option:focus-visible,
.platform-option:focus-visible,
.goal-chip:focus-visible,
.footer-button:focus-visible {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.09);
}

.market-option {
  height: 76px;
  border-radius: 16px;
  padding: 18px;
  display: flex;
  align-items: center;
  gap: 12px;
  text-align: left;
}

.market-option__emoji {
  width: 34px;
  font-size: 30px;
  line-height: 36px;
}

.market-option__copy {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.market-option__copy strong {
  color: #0a2463;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.market-option__copy small {
  color: #62748e;
  font-size: 12px;
  line-height: 16px;
}

.platform-option {
  height: 124px;
  border-radius: 16px;
  padding: 26px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
}

.platform-option__emoji {
  font-size: 36px;
  line-height: 40px;
}

.platform-option strong {
  color: #0a2463;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.goal-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.goal-chip {
  min-height: 38px;
  border-radius: 999px;
  padding: 7px 16px;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

.entry-notice {
  margin: 18px 33px 0;
  color: #e11d48;
  font-size: 14px;
  line-height: 20px;
}

.product-entry-card__footer {
  margin: 24px 33px 28px;
  padding-top: 28px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.footer-button {
  height: 48px;
  border: none;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
  cursor: pointer;
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    opacity 180ms ease;
}

.footer-button img {
  width: 20px;
  height: 20px;
  display: block;
}

.footer-button--ghost {
  width: 96px;
  background: #f1f5f9;
  color: #0a2463;
}

.footer-button--ghost:hover {
  background: #e8eef6;
}

.footer-button--primary {
  width: 124px;
  color: #ffffff;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
}

.footer-button--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(99, 102, 241, 0.24);
}

.footer-button:active {
  transform: translateY(0);
}

@media (max-width: 1120px) {
  .product-entry-card {
    width: 100%;
  }
}

@media (max-width: 760px) {
  .market-grid,
  .platform-grid {
    grid-template-columns: 1fr;
  }

  .platform-option {
    height: 96px;
    align-items: flex-start;
    justify-content: center;
  }
}

@media (max-width: 640px) {
  .product-entry-card__body {
    padding: 24px 20px 0;
  }

  .entry-notice {
    margin-inline: 20px;
  }

  .product-entry-card__footer {
    margin: 24px 20px;
    flex-direction: column-reverse;
    gap: 12px;
  }

  .footer-button,
  .footer-button--ghost,
  .footer-button--primary {
    width: 100%;
  }
}
</style>
