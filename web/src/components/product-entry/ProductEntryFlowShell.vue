<template>
  <div class="product-entry-flow-page">
    <header class="product-entry-flow-header">
      <div class="product-entry-flow-header__inner">
        <div class="product-entry-flow-header__left">
          <button type="button" class="product-entry-flow-brand" aria-label="返回首页" @click="router.push('/')">
            <span class="product-entry-flow-brand__mark">
              <img :src="brandLogo" alt="" />
            </span>
            <span class="product-entry-flow-brand__copy">
              <strong>数字丝路</strong>
              <small>Digital Silk Road</small>
            </span>
          </button>

          <nav class="product-entry-flow-nav" aria-label="主导航">
            <button
              v-for="item in navItems"
              :key="item.label"
              type="button"
              class="product-entry-flow-nav__item"
              :class="{ 'product-entry-flow-nav__item--active': item.active }"
              :style="{ width: item.width }"
              :aria-current="item.active ? 'page' : undefined"
              @pointerenter="preloadRouteByPath(item.path)"
              @focus="preloadRouteByPath(item.path)"
              @click="handleNavClick(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="product-entry-flow-header__right">
          <NotificationBell />
        </div>
      </div>
    </header>

    <main class="product-entry-flow-main">
      <div class="product-entry-flow-shell">
        <div class="product-entry-flow-layout">
          <section class="product-entry-flow-head">
            <h1 class="product-entry-flow-head__title">商品信息录入</h1>
            <p class="product-entry-flow-head__subtitle">填写商品基本信息，开启智能合规检测与内容生成流程</p>
          </section>

          <section v-if="hasAgentPrefill" class="product-entry-flow-prefill">
            已根据丝路 Agent 分析结果预填，您可以继续补充或修改
          </section>

          <section class="product-entry-flow-steps" aria-label="步骤进度">
            <div
              v-for="(step, index) in steps"
              :key="step.key"
              class="product-entry-flow-step"
              :class="{ 'product-entry-flow-step--last': index === steps.length - 1 }"
            >
              <div class="product-entry-flow-step__lead">
                <span
                  class="product-entry-flow-step__icon"
                  :class="{
                    'product-entry-flow-step__icon--active': index <= activeStepIndex,
                    'product-entry-flow-step__icon--current': index === activeStepIndex
                  }"
                >
                  <img :src="step.icon" alt="" />
                </span>
                <span
                  class="product-entry-flow-step__label"
                  :class="{ 'product-entry-flow-step__label--active': index <= activeStepIndex }"
                >
                  {{ step.label }}
                </span>
              </div>

              <span v-if="index !== steps.length - 1" class="product-entry-flow-step__line" aria-hidden="true">
                <span
                  class="product-entry-flow-step__line-fill"
                  :class="{ 'product-entry-flow-step__line-fill--active': index < activeStepIndex }"
                ></span>
              </span>
            </div>
          </section>

          <slot />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { preloadRouteByPath } from '@/router'
import NotificationBell from '@/components/common/NotificationBell.vue'
import { getProductDraft } from '@/utils/productEntryDraft'
import stepBasicIcon from '@/assets/figma/product-entry/step-basic.svg'
import stepCompleteIcon from '@/assets/figma/product-entry/step-complete.svg'
import stepDetailIcon from '@/assets/figma/product-entry/step-detail.svg'
import stepMarketIcon from '@/assets/figma/product-entry/step-market.svg'

type ProductEntryStepKey = 'basic' | 'market' | 'details' | 'complete'

const props = defineProps<{
  activeStep: ProductEntryStepKey
}>()

const router = useRouter()
const brandLogo = '/logo_circle.png'
const hasAgentPrefill = computed(() => getProductDraft()?.source === 'agent')

const navItems = [
  { label: '工作台', path: '/dramas', active: false, width: '66px' },
  { label: '商品录入', path: '/dramas/create', active: true, width: '80px' },
  { label: '合规分析', path: '/compliance', active: false, width: '80px' },
  { label: '脚本/分镜', path: '/workspace/script', active: false, width: '92px' },
  { label: '内容创作', path: '/workspace/content', active: false, width: '80px' },
  { label: '视频剪辑', path: '/workspace/timeline', active: false, width: '80px' },
  { label: '数据分析', path: '/analytics', active: false, width: '80px' }
] as const

const steps = [
  { key: 'basic', label: '基本信息', icon: stepBasicIcon },
  { key: 'market', label: '目标市场', icon: stepMarketIcon },
  { key: 'details', label: '商品详情', icon: stepDetailIcon },
  { key: 'complete', label: '完成', icon: stepCompleteIcon }
] as const

const activeStepIndex = computed(() => {
  const index = steps.findIndex((step) => step.key === props.activeStep)
  return index >= 0 ? index : 0
})

const handleNavClick = (path: string) => {
  if (!path || path === router.currentRoute.value.path) {
    return
  }

  preloadRouteByPath(path)
  router.push(path)
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;700&family=Urbanist:wght@700&display=swap');

.product-entry-flow-page {
  min-height: 100vh;
  width: 100%;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
  overflow-x: hidden;
}

.product-entry-flow-page,
.product-entry-flow-page :is(button, input, select, textarea) {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.product-entry-flow-header {
  position: fixed;
  inset: 0 0 auto;
  z-index: 30;
  height: 65px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.product-entry-flow-header__inner {
  width: min(100%, 1075px);
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.product-entry-flow-header__left {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  gap: 32px;
}

.product-entry-flow-brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0;
  border: none;
  background: transparent;
  color: #0a2463;
  cursor: pointer;
  flex-shrink: 0;
}

.product-entry-flow-brand__mark {
  width: 44px;
  height: 44px;
  padding: 4px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(226, 232, 240, 0.92);
  box-shadow: 0 12px 28px -18px rgba(15, 23, 42, 0.34);
}

.product-entry-flow-brand__mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 999px;
  display: block;
}

.product-entry-flow-brand__copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.product-entry-flow-brand__copy strong {
  font-size: 16px;
  font-weight: 700;
  line-height: 22px;
  white-space: nowrap;
}

.product-entry-flow-brand__copy small {
  color: #62748e;
  font-size: 11px;
  line-height: 14px;
  white-space: nowrap;
}

.product-entry-flow-nav {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  scrollbar-width: none;
}

.product-entry-flow-nav::-webkit-scrollbar {
  display: none;
}

.product-entry-flow-nav__item {
  height: 32px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #45556c;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  cursor: pointer;
  transition:
    background-color 180ms ease,
    color 180ms ease,
    transform 180ms ease;
  white-space: nowrap;
}

.product-entry-flow-nav__item:hover {
  color: #0a2463;
  background: rgba(241, 245, 249, 0.92);
}

.product-entry-flow-nav__item--active {
  color: #0a2463;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
}

.product-entry-flow-header__right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  flex-shrink: 0;
}

.product-entry-flow-main,
.product-entry-flow-shell {
  width: 100%;
}

.product-entry-flow-shell {
  width: min(100%, 1075px);
  margin: 0 auto;
}

.product-entry-flow-layout {
  padding: 96px 46px 40px;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

.product-entry-flow-head,
.product-entry-flow-prefill,
.product-entry-flow-steps {
  width: 960px;
  margin-inline: auto;
}

.product-entry-flow-head__title {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 30px;
  font-weight: 700;
  line-height: 36px;
}

.product-entry-flow-head__subtitle {
  margin: 8px 0 0;
  color: #45556c;
  font-size: 16px;
  font-weight: 400;
  line-height: 24px;
}

.product-entry-flow-prefill {
  margin-top: 16px;
  padding: 12px 16px;
  border: 1px solid rgba(6, 182, 212, 0.24);
  border-radius: 12px;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1), rgba(124, 58, 237, 0.08));
  color: #0a2463;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

.product-entry-flow-steps {
  height: 76px;
  margin-top: 28px;
  display: flex;
  align-items: flex-start;
}

.product-entry-flow-step {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
  flex: 1 1 0;
}

.product-entry-flow-step--last {
  flex: 0 0 auto;
}

.product-entry-flow-step__lead {
  width: 56px;
  height: 76px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.product-entry-flow-step__icon {
  width: 48px;
  height: 48px;
  border-radius: 999px;
  background: #e2e8f0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  transition:
    background 180ms ease,
    box-shadow 180ms ease,
    transform 180ms ease;
}

.product-entry-flow-step__icon--active {
  background: linear-gradient(135deg, #06b6d4 0%, #6382e2 50%, #7c3aed 100%);
}

.product-entry-flow-step__icon--current {
  box-shadow:
    0 10px 15px 0 rgba(0, 0, 0, 0.1),
    0 4px 6px 0 rgba(0, 0, 0, 0.1);
}

.product-entry-flow-step__icon img {
  width: 24px;
  height: 24px;
  display: block;
  transition:
    filter 180ms ease,
    opacity 180ms ease;
}

.product-entry-flow-step__icon--active img {
  filter: brightness(0) invert(1);
}

.product-entry-flow-step__label {
  color: #90a1b9;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  white-space: nowrap;
}

.product-entry-flow-step__label--active {
  color: #0a2463;
}

.product-entry-flow-step__line {
  width: auto;
  min-width: 0;
  height: 2px;
  background: #e2e8f0;
  flex: 1 1 auto;
  margin-right: 16px;
  overflow: hidden;
}

.product-entry-flow-step__line-fill {
  display: block;
  width: 0;
  height: 100%;
  background: linear-gradient(90deg, #06b6d4 0%, #6382e2 50%, #7c3aed 100%);
  transition: width 220ms ease;
}

.product-entry-flow-step__line-fill--active {
  width: 100%;
}

@media (max-width: 1120px) {
  .product-entry-flow-header__inner,
  .product-entry-flow-shell {
    width: 100%;
  }

  .product-entry-flow-layout {
    padding-inline: 20px;
  }

  .product-entry-flow-head,
  .product-entry-flow-prefill,
  .product-entry-flow-steps {
    width: 100%;
  }
}

@media (max-width: 900px) {
  .product-entry-flow-header {
    height: auto;
  }

  .product-entry-flow-header__inner {
    height: auto;
    padding-block: 12px;
    align-items: flex-start;
    flex-direction: column;
  }

  .product-entry-flow-header__left,
  .product-entry-flow-header__right {
    width: 100%;
  }

  .product-entry-flow-header__right {
    justify-content: flex-end;
  }

  .product-entry-flow-layout {
    padding-top: 140px;
  }

  .product-entry-flow-steps {
    display: grid;
    height: auto;
    gap: 20px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .product-entry-flow-step__line {
    display: none;
  }
}

@media (max-width: 640px) {
  .product-entry-flow-layout {
    padding: 148px 16px 28px;
  }

  .product-entry-flow-head__title {
    font-size: 28px;
    line-height: 34px;
  }
}
</style>
