<template>
  <div class="data-analysis-page">
    <header class="data-analysis-header">
      <div class="data-analysis-header__inner">
        <div class="data-analysis-header__left">
          <button type="button" class="brand-link" aria-label="返回首页" @click="router.push('/')">
            <span class="brand-link__mark">
              <img :src="brandMarkIcon" alt="" />
            </span>
            <span class="brand-link__name">数字丝路</span>
          </button>

          <nav class="data-analysis-nav" aria-label="主导航">
            <button
              v-for="item in navItems"
              :key="item.label"
              type="button"
              class="data-analysis-nav__item"
              :class="{ 'data-analysis-nav__item--active': item.active }"
              :style="{ width: item.width }"
              :aria-current="item.active ? 'page' : undefined"
              @click="handleNavClick(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="data-analysis-header__right">
          <button type="button" class="header-icon-button" aria-label="通知">
            <img :src="bellIcon" alt="" />
            <span class="header-icon-button__dot"></span>
          </button>
        </div>
      </div>
    </header>

    <main class="data-analysis-main">
      <div class="data-analysis-shell">
        <section class="analytics-hero">
          <h1>数据分析中心</h1>
          <p>全链路数据追踪，AI智能优化建议</p>
        </section>

        <section class="metrics-grid">
          <article
            v-for="card in metricCards"
            :key="card.label"
            class="metric-card"
            :class="`metric-card--${card.tone}`"
          >
            <div class="metric-card__top">
              <span class="metric-card__icon" :class="`metric-card__icon--${card.tone}`" aria-hidden="true">
                <svg v-if="card.icon === 'eye'" viewBox="0 0 24 24" fill="none">
                  <path
                    d="M2.5 12C4.7 7.8 8 5.75 12 5.75S19.3 7.8 21.5 12c-2.2 4.2-5.5 6.25-9.5 6.25S4.7 16.2 2.5 12Z"
                    stroke="currentColor"
                    stroke-width="1.8"
                    stroke-linejoin="round"
                  />
                  <circle cx="12" cy="12" r="3.1" fill="currentColor" />
                </svg>
                <svg v-else-if="card.icon === 'spark'" viewBox="0 0 24 24" fill="none">
                  <path
                    d="m12 3.5 1.48 4.34L17.8 9.3l-4.32 1.46L12 15.1l-1.48-4.34L6.2 9.3l4.32-1.46L12 3.5Z"
                    fill="currentColor"
                  />
                  <path d="M18.6 4.8v2.4M17.4 6h2.4M5.4 15.8v2.4M4.2 17h2.4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                </svg>
                <svg v-else-if="card.icon === 'cart'" viewBox="0 0 24 24" fill="none">
                  <path
                    d="M4 5h1.5l1.5 8.25h9.5l2-5.75H8"
                    stroke="currentColor"
                    stroke-width="1.8"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                  <circle cx="10" cy="18" r="1.6" fill="currentColor" />
                  <circle cx="17" cy="18" r="1.6" fill="currentColor" />
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none">
                  <path d="M12 3v18M17 6.5c0-1.7-2-3-5-3S7 4.8 7 6.5 9 9.5 12 9.5s5 1.3 5 3-2 3-5 3-5-1.3-5-3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>

              <span class="metric-card__trend">
                <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
                  <path d="M3 10.5 6.5 7l2.6 2.6L13 5.7M10.1 5.7H13v2.9" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
                {{ card.trend }}
              </span>
            </div>

            <strong class="metric-card__value">{{ card.value }}</strong>
            <span class="metric-card__label">{{ card.label }}</span>
          </article>
        </section>

        <section class="analytics-chart-grid">
          <article class="analysis-panel">
            <div class="analysis-panel__title">流量趋势分析</div>

            <div class="line-chart">
              <div class="line-chart__surface">
                <div class="line-chart__y-axis">
                  <span v-for="label in trendYAxis" :key="label">{{ label }}</span>
                </div>

                <div class="line-chart__plot">
                  <div class="line-chart__grid" aria-hidden="true"></div>

                  <svg class="line-chart__svg" viewBox="0 0 370 236" preserveAspectRatio="none" aria-hidden="true">
                    <polyline
                      v-for="series in trendSeries"
                      :key="series.name"
                      :points="trendPoints"
                      fill="none"
                      :stroke="series.color"
                      :stroke-width="series.width"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      :opacity="series.opacity"
                    />
                    <g v-for="point in trendPointNodes" :key="point.x">
                      <circle :cx="point.x" :cy="point.y" r="5" fill="#ffffff" stroke="#ff7a18" stroke-width="2.25" />
                      <circle :cx="point.x" :cy="point.y" r="2.2" fill="#ff7a18" />
                    </g>
                  </svg>

                  <div class="line-chart__x-axis">
                    <span v-for="label in trendXAxis" :key="label">{{ label }}</span>
                  </div>
                </div>
              </div>

              <div class="chart-legend chart-legend--line">
                <span v-for="series in trendSeries" :key="series.name" class="chart-legend__item">
                  <i class="chart-legend__dot" :style="{ background: series.color }"></i>
                  {{ series.name }}
                </span>
              </div>
            </div>
          </article>

          <article class="analysis-panel">
            <div class="analysis-panel__title">市场分布</div>

            <div class="market-panel">
              <div class="market-panel__chart">
                <div class="market-donut">
                  <div class="market-donut__ring"></div>
                  <div class="market-donut__center"></div>
                </div>
              </div>

              <div class="market-panel__legend">
                <div v-for="item in marketShare" :key="item.name" class="market-panel__legend-item">
                  <i class="market-panel__legend-dot" :style="{ background: item.color }"></i>
                  <div class="market-panel__legend-copy">
                    <span>{{ item.name }}</span>
                    <strong>{{ item.value }}</strong>
                  </div>
                </div>
              </div>
            </div>
          </article>
        </section>

        <section class="analysis-panel analysis-panel--video">
          <div class="analysis-panel__title">视频表现对比</div>

          <div class="bar-chart">
            <div class="bar-chart__surface">
              <div class="bar-chart__y-axis">
                <span v-for="label in videoYAxis" :key="label">{{ label }}</span>
              </div>

              <div class="bar-chart__plot">
                <div class="bar-chart__grid" aria-hidden="true"></div>

                <div class="bar-chart__bars">
                  <div v-for="item in videoBars" :key="item.name" class="bar-chart__group">
                    <div class="bar-chart__pair">
                      <span
                        class="bar-chart__bar bar-chart__bar--view"
                        :style="{ height: `${getChartHeight(item.views, 60000)}px` }"
                      ></span>
                      <span
                        class="bar-chart__bar bar-chart__bar--conversion"
                        :style="{ height: `${getChartHeight(item.conversions, 60000)}px` }"
                      ></span>
                    </div>
                    <span class="bar-chart__label">{{ item.name }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="chart-legend chart-legend--bar">
              <span class="chart-legend__item">
                <i class="chart-legend__dot chart-legend__dot--view"></i>
                观看量
              </span>
              <span class="chart-legend__item">
                <i class="chart-legend__dot chart-legend__dot--conversion"></i>
                转化数
              </span>
            </div>
          </div>
        </section>

        <section class="insight-section">
          <h2>AI 智能洞察</h2>

          <div class="insight-grid">
            <article
              v-for="item in insights"
              :key="item.title"
              class="insight-card"
              :class="`insight-card--${item.tone}`"
            >
              <span class="insight-card__icon" :class="`insight-card__icon--${item.tone}`" aria-hidden="true">
                <svg v-if="item.icon === 'growth'" viewBox="0 0 24 24" fill="none">
                  <path d="M4 16.5 9 11.5l3 3 7-7" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
                  <path d="M14 7.5h5v5" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
                <svg v-else-if="item.icon === 'audience'" viewBox="0 0 24 24" fill="none">
                  <circle cx="9" cy="9" r="3" stroke="currentColor" stroke-width="1.8" />
                  <path d="M3.5 18c1.2-2.6 3-3.9 5.5-3.9s4.3 1.3 5.5 3.9" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
                  <circle cx="17.5" cy="8.5" r="2.5" stroke="currentColor" stroke-width="1.6" />
                  <path d="M15.5 17.3c.7-1.6 1.9-2.5 3.6-2.5.7 0 1.3.1 1.9.4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none">
                  <rect x="4.5" y="6" width="15" height="11" rx="2" stroke="currentColor" stroke-width="1.8" />
                  <path d="m10 10 4 2-4 2v-4Z" fill="currentColor" />
                </svg>
              </span>

              <h3>{{ item.title }}</h3>
              <p>{{ item.description }}</p>
            </article>
          </div>
        </section>

        <section class="suggestion-panel">
          <div class="suggestion-panel__title">
            <span class="suggestion-panel__title-icon" aria-hidden="true">
              <svg viewBox="0 0 32 32" fill="none">
                <path
                  d="M16 5.5c5.8 0 10.5 4.7 10.5 10.5 0 5.8-4.7 10.5-10.5 10.5S5.5 21.8 5.5 16C5.5 10.2 10.2 5.5 16 5.5Z"
                  stroke="currentColor"
                  stroke-width="2"
                />
                <path d="m10.5 14.5 3.4 3.4 7.6-7.6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </span>
            <h2>AI 优化建议</h2>
          </div>

          <div class="suggestion-grid">
            <article v-for="item in recommendations" :key="item.title" class="suggestion-card">
              <h3>{{ item.title }}</h3>
              <p>{{ item.description }}</p>
            </article>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import bellIcon from '@/assets/figma/product-entry/bell.svg'
import brandMarkIcon from '@/assets/figma/product-entry/brand-mark.svg'
import { useRouter } from 'vue-router'

type MetricIcon = 'eye' | 'spark' | 'cart' | 'coin'
type InsightIcon = 'growth' | 'audience' | 'video'
type AccentTone = 'blue' | 'purple' | 'orange' | 'green'

const router = useRouter()

const navItems = [
  { label: '工作台', path: '/dramas', active: false, width: '66px' },
  { label: '商品录入', path: '/dramas/create', active: false, width: '80px' },
  { label: '合规分析', path: '/compliance', active: false, width: '80px' },
  { label: '脚本/分镜', path: '/workspace/script', active: false, width: '92px' },
  { label: '内容创作', path: '/workspace/content', active: false, width: '80px' },
  { label: '视频剪辑', path: '/workspace/timeline', active: false, width: '80px' },
  { label: '数据分析', path: '/analytics', active: true, width: '80px' }
] as const

const metricCards: ReadonlyArray<{
  icon: MetricIcon
  label: string
  value: string
  trend: string
  tone: AccentTone
}> = [
  { icon: 'eye', label: '总曝光量', value: '125.4K', trend: '+24.5%', tone: 'blue' },
  { icon: 'spark', label: '点击率', value: '8.7%', trend: '+12.3%', tone: 'purple' },
  { icon: 'cart', label: '转化率', value: '4.2%', trend: '+18.7%', tone: 'orange' },
  { icon: 'coin', label: 'ROI', value: '3.8x', trend: '+25.4%', tone: 'green' }
]

const trendYAxis = ['26000', '19500', '13000', '6500', '0'] as const
const trendXAxis = ['04-12', '04-13', '04-14', '04-15', '04-16', '04-17', '04-18'] as const
const trendSeries = [
  { name: '曝光', color: '#19b9dc', width: 2.3, opacity: 0.7 },
  { name: '点击', color: '#7c3aed', width: 2.6, opacity: 0.72 },
  { name: '转化', color: '#ff7a18', width: 3.1, opacity: 1 }
] as const
const trendPointNodes = [
  { x: 0, y: 94.4 },
  { x: 61.67, y: 65.35 },
  { x: 123.33, y: 79.88 },
  { x: 185, y: 37.22 },
  { x: 246.67, y: 58.09 },
  { x: 308.33, y: 25.42 },
  { x: 370, y: 0 }
] as const
const trendPoints = trendPointNodes.map((point) => `${point.x},${point.y}`).join(' ')

const marketShare = [
  { name: '美国', value: '42%', color: '#18b5d5' },
  { name: '德国', value: '18%', color: '#7a3fe8' },
  { name: '日本', value: '15%', color: '#ff7a18' },
  { name: '英国', value: '12%', color: '#14b97f' },
  { name: '其他', value: '13%', color: '#64748b' }
] as const

const videoYAxis = ['60000', '45000', '30000', '15000', '0'] as const
const videoBars = [
  { name: '视频A', views: 47000, conversions: 45000 },
  { name: '视频B', views: 40500, conversions: 39000 },
  { name: '视频C', views: 33000, conversions: 32000 },
  { name: '视频D', views: 29500, conversions: 29000 }
] as const

const insights: ReadonlyArray<{
  icon: InsightIcon
  tone: AccentTone
  title: string
  description: string
}> = [
  {
    icon: 'growth',
    tone: 'green',
    title: '转化率提升显著',
    description: '相比上周提升18.7%，建议继续优化当前营销策略'
  },
  {
    icon: 'audience',
    tone: 'blue',
    title: '美国市场表现最佳',
    description: '占总转化的42%，建议增加美国市场投放预算'
  },
  {
    icon: 'video',
    tone: 'orange',
    title: '视频完播率待提升',
    description: '平均完播率62%，建议优化视频前3秒内容以提升留存'
  }
] as const

const recommendations = [
  {
    title: '📊 投放策略优化',
    description: '建议在美国市场增加30%预算，同时优化德国市场的内容本地化以提升转化率'
  },
  {
    title: '🎬 内容创意优化',
    description: '视频A表现最佳，建议制作相似风格的变体版本，并在前3秒突出核心卖点'
  }
] as const

const handleNavClick = (path: string) => {
  if (!path) {
    return
  }

  router.push(path)
}

const getChartHeight = (value: number, max: number) => Number(((value / max) * 236).toFixed(2))
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;700&family=Urbanist:wght@700&display=swap');

.data-analysis-page {
  min-height: 100vh;
  width: 100%;
  background: #f3f4f6;
  color: #0a2463;
  overflow-x: hidden;
}

.data-analysis-page,
.data-analysis-page :is(button, input, select) {
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.data-analysis-header {
  position: fixed;
  inset: 0 0 auto;
  z-index: 30;
  height: 65px;
  background: #ffffff;
  border-bottom: 1px solid #e5e7eb;
}

.data-analysis-header__inner {
  width: min(100%, 1075px);
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.data-analysis-header__left {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  gap: 32px;
}

.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  flex-shrink: 0;
}

.brand-link__mark {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0a2463 0%, #06b6d4 50%, #7c3aed 100%);
}

.brand-link__mark img {
  width: 20px;
  height: 20px;
  display: block;
}

.brand-link__name {
  color: #0a2463;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
  white-space: nowrap;
}

.data-analysis-nav {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  scrollbar-width: none;
}

.data-analysis-nav::-webkit-scrollbar {
  display: none;
}

.data-analysis-nav__item {
  height: 32px;
  padding: 0 12px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #45556c;
  font-size: 14px;
  line-height: 20px;
  font-weight: 400;
  cursor: pointer;
  flex: 0 0 auto;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.data-analysis-nav__item:hover {
  color: #0a2463;
  background: rgba(241, 245, 249, 0.9);
}

.data-analysis-nav__item--active {
  color: #0a2463;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.12) 0%, rgba(124, 58, 237, 0.12) 100%);
  box-shadow: inset 0 0 0 1px rgba(103, 116, 255, 0.08);
}

.data-analysis-header__right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.header-icon-button {
  position: relative;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 12px;
  background: transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.header-icon-button img {
  width: 20px;
  height: 20px;
  display: block;
}

.header-icon-button__dot {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #ff7a18;
}

.data-analysis-main {
  width: min(100%, 1075px);
  margin: 0 auto;
  padding: 104px 32px 48px;
}

.data-analysis-shell {
  width: min(100%, 1011px);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.analytics-hero {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.analytics-hero h1 {
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #0a2463;
}

.analytics-hero p {
  color: #45556c;
  font-size: 16px;
  line-height: 24px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 24px;
}

.metric-card,
.analysis-panel {
  background: #ffffff;
  border: 1px solid #e3e8f1;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.018);
}

.metric-card {
  min-height: 174px;
  padding: 25px;
  display: flex;
  flex-direction: column;
}

.metric-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.metric-card__icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
}

.metric-card__icon svg {
  width: 24px;
  height: 24px;
  display: block;
}

.metric-card__icon--blue {
  background: #1d9bf0;
}

.metric-card__icon--purple {
  background: #d946ef;
}

.metric-card__icon--orange {
  background: #ff5d1f;
}

.metric-card__icon--green {
  background: #10b981;
}

.metric-card__trend {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #14b97f;
  font-size: 13px;
  line-height: 20px;
  font-weight: 500;
  white-space: nowrap;
}

.metric-card__trend svg {
  width: 16px;
  height: 16px;
  display: block;
}

.metric-card__value {
  display: block;
  margin-top: auto;
  color: #163a76;
  font-size: 35px;
  line-height: 36px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.metric-card__label {
  margin-top: 4px;
  color: #45556c;
  font-size: 14px;
  line-height: 20px;
}

.analytics-chart-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 32px;
}

.analysis-panel {
  min-height: 402px;
  padding: 25px;
}

.analysis-panel--video {
  min-height: 402px;
}

.analysis-panel__title {
  color: #0a2463;
  font-size: 20px;
  line-height: 28px;
  font-weight: 700;
}

.line-chart {
  margin-top: 24px;
}

.line-chart__surface {
  display: grid;
  grid-template-columns: 46px minmax(0, 1fr);
  align-items: start;
  gap: 14px;
}

.line-chart__y-axis,
.bar-chart__y-axis {
  height: 236px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: flex-end;
  color: #94a3b8;
  font-size: 11px;
  line-height: 15px;
}

.line-chart__plot {
  position: relative;
  height: 260px;
  padding-bottom: 24px;
}

.line-chart__grid,
.bar-chart__grid {
  position: absolute;
  inset: 0 0 24px;
  border-left: 1px solid #94a3b8;
  border-bottom: 1px solid #94a3b8;
  background-image:
    repeating-linear-gradient(to right, rgba(203, 213, 225, 0.72) 0 1px, transparent 1px 61.666px),
    repeating-linear-gradient(to bottom, rgba(203, 213, 225, 0.72) 0 1px, transparent 1px 59px);
  background-position: left top;
  background-size: 100% 100%;
  opacity: 0.7;
}

.line-chart__svg {
  position: absolute;
  inset: 0 0 24px;
  width: 100%;
  height: 236px;
  overflow: visible;
}

.line-chart__x-axis {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  color: #94a3b8;
  font-size: 11px;
  line-height: 15px;
}

.line-chart__x-axis span {
  text-align: center;
}

.chart-legend {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 18px;
  color: #0a2463;
  font-size: 13px;
  line-height: 24px;
  margin-top: 10px;
}

.chart-legend__item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.chart-legend__dot {
  display: inline-block;
}

.chart-legend--line .chart-legend__dot {
  position: relative;
  width: 14px;
  height: 2px;
  border-radius: 999px;
}

.chart-legend--line .chart-legend__dot::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: inherit;
  transform: translate(-50%, -50%);
}

.chart-legend--bar .chart-legend__dot {
  width: 14px;
  height: 14px;
  border-radius: 4px;
}

.chart-legend__dot--view {
  background: #18b5d5;
}

.chart-legend__dot--conversion {
  background: #ff7a18;
}

.market-panel {
  margin-top: 24px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 56px;
  align-items: center;
  min-height: 250px;
}

.market-panel__chart {
  min-height: 250px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.market-donut {
  position: relative;
  width: 168px;
  height: 168px;
}

.market-donut__ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: conic-gradient(
    #18b5d5 0deg 151.2deg,
    #7a3fe8 151.2deg 216deg,
    #ff7a18 216deg 270deg,
    #14b97f 270deg 313.2deg,
    #64748b 313.2deg 360deg
  );
}

.market-donut__ring::after {
  content: '';
  position: absolute;
  inset: 24px;
  border-radius: 50%;
  background: #ffffff;
  box-shadow: inset 0 0 0 1px rgba(226, 232, 240, 0.9);
}

.market-donut__center {
  position: absolute;
  inset: 44px;
  border-radius: 50%;
  background: #ffffff;
}

.market-panel__legend {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.market-panel__legend-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.market-panel__legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  flex-shrink: 0;
}

.market-panel__legend-copy {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.market-panel__legend-copy span {
  color: #0a2463;
  font-size: 13px;
  line-height: 20px;
  font-weight: 600;
}

.market-panel__legend-copy strong {
  color: #64748b;
  font-size: 12px;
  line-height: 16px;
  font-weight: 500;
}

.bar-chart {
  margin-top: 24px;
}

.bar-chart__surface {
  display: grid;
  grid-template-columns: 46px minmax(0, 1fr);
  align-items: start;
  gap: 14px;
}

.bar-chart__plot {
  position: relative;
  height: 260px;
  padding-bottom: 24px;
}

.bar-chart__bars {
  position: absolute;
  inset: 0 0 24px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  align-items: end;
  padding: 0 18px;
}

.bar-chart__group {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  min-height: 236px;
}

.bar-chart__pair {
  position: relative;
  width: 87px;
  height: 236px;
}

.bar-chart__bar {
  position: absolute;
  left: 50%;
  bottom: 0;
  border-radius: 6px 6px 0 0;
  transform: translateX(-50%);
  transition: transform 0.18s ease, opacity 0.18s ease;
}

.bar-chart__bar--view {
  z-index: 1;
  width: 52px;
  background: linear-gradient(180deg, rgba(24, 181, 213, 0.46) 0%, rgba(24, 181, 213, 0.2) 100%);
  opacity: 0.32;
}

.bar-chart__bar--conversion {
  z-index: 2;
  width: 48px;
  background: linear-gradient(180deg, #ff7a18 0%, #ff6a00 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.bar-chart__group:hover .bar-chart__bar {
  transform: translateX(-50%) translateY(-2px);
}

.bar-chart__label {
  color: #64748b;
  font-size: 11px;
  line-height: 15px;
}

.chart-legend--bar {
  margin-top: 10px;
}

.insight-section {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.insight-section h2 {
  color: #0a2463;
  font-size: 24px;
  line-height: 28px;
  font-weight: 700;
}

.insight-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 24px;
}

.insight-card {
  min-height: 197.5px;
  padding: 26px;
  border-radius: 12px;
  border: 1px solid transparent;
}

.insight-card--green {
  background: linear-gradient(180deg, #eefcf2 0%, #f7fff9 100%);
  border-color: rgba(104, 224, 145, 0.7);
}

.insight-card--blue {
  background: linear-gradient(180deg, #eff6ff 0%, #f6faff 100%);
  border-color: rgba(118, 170, 255, 0.72);
}

.insight-card--orange {
  background: linear-gradient(180deg, #fff4eb 0%, #fffaf5 100%);
  border-color: rgba(255, 186, 112, 0.8);
}

.insight-card__icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
}

.insight-card__icon svg {
  width: 24px;
  height: 24px;
  display: block;
}

.insight-card__icon--green {
  background: #11b96d;
}

.insight-card__icon--blue {
  background: #1d9bf0;
}

.insight-card__icon--orange {
  background: #ff5d1f;
}

.insight-card h3 {
  margin-top: 16px;
  color: #0a2463;
  font-size: 24px;
  line-height: 28px;
  font-weight: 700;
}

.insight-card p {
  margin-top: 8px;
  color: #4b5a73;
  font-size: 14px;
  line-height: 22.75px;
}

.suggestion-panel {
  min-height: 253.5px;
  padding: 32px;
  border-radius: 16px;
  background: linear-gradient(98deg, #0b5e99 0%, #12b6d2 49%, #7c3aed 100%);
  box-shadow: 0 24px 52px -30px rgba(37, 99, 235, 0.46);
}

.suggestion-panel__title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.suggestion-panel__title h2 {
  color: #ffffff;
  font-size: 28px;
  line-height: 32px;
  font-weight: 700;
}

.suggestion-panel__title-icon {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
}

.suggestion-panel__title-icon svg {
  width: 32px;
  height: 32px;
  display: block;
}

.suggestion-grid {
  margin-top: 24px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 24px;
}

.suggestion-card {
  min-height: 133.5px;
  padding: 24px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.14) 0%, rgba(255, 255, 255, 0.08) 100%);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.suggestion-card:first-child {
  background: linear-gradient(135deg, rgba(20, 182, 209, 0.24) 0%, rgba(70, 196, 255, 0.14) 100%);
}

.suggestion-card:last-child {
  background: linear-gradient(135deg, rgba(124, 178, 255, 0.2) 0%, rgba(124, 58, 237, 0.18) 100%);
}

.suggestion-card h3 {
  color: #ffffff;
  font-size: 24px;
  line-height: 28px;
  font-weight: 700;
}

.suggestion-card p {
  margin-top: 12px;
  color: rgba(255, 255, 255, 0.86);
  font-size: 14px;
  line-height: 22.75px;
}

@media (max-width: 1120px) {
  .data-analysis-main {
    padding-inline: 20px;
  }
}

@media (max-width: 980px) {
  .metrics-grid,
  .analytics-chart-grid,
  .insight-grid,
  .suggestion-grid {
    grid-template-columns: 1fr;
  }

  .market-panel {
    grid-template-columns: 1fr;
    gap: 24px;
    justify-items: center;
  }

  .market-panel__legend {
    width: 100%;
    max-width: 220px;
  }
}

@media (max-width: 768px) {
  .data-analysis-header__inner {
    padding-inline: 16px;
    gap: 12px;
  }

  .data-analysis-header__left {
    gap: 18px;
  }

  .data-analysis-main {
    padding: 96px 16px 32px;
  }

  .metric-card,
  .analysis-panel,
  .insight-card,
  .suggestion-panel {
    padding-inline: 20px;
  }

  .bar-chart__pair {
    width: 72px;
  }

  .bar-chart__bar--conversion {
    width: 34px;
  }

  .bar-chart__bar--view {
    width: 38px;
  }

  .suggestion-panel__title h2 {
    font-size: 24px;
  }

  .suggestion-card h3,
  .insight-card h3 {
    font-size: 20px;
  }
}
</style>
