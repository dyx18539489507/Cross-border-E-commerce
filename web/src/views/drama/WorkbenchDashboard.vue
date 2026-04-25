<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import bellIcon from '../../assets/workbench/bell.svg'
import brandLogo from '../../assets/workbench/brand-logo.svg'
import overviewBagIcon from '../../assets/workbench/overview-bag.svg'
import overviewComplianceIcon from '../../assets/workbench/overview-compliance.svg'
import overviewMarketIcon from '../../assets/workbench/overview-market.svg'
import overviewTrendIcon from '../../assets/workbench/overview-trend.svg'
import overviewVideoIcon from '../../assets/workbench/overview-video.svg'
import quickBagIcon from '../../assets/workbench/quick-bag.svg'
import quickChartIcon from '../../assets/workbench/quick-chart.svg'
import quickComplianceIcon from '../../assets/workbench/quick-compliance.svg'
import quickVideoIcon from '../../assets/workbench/quick-video.svg'
import recentArrowIcon from '../../assets/workbench/recent-arrow.svg'
import taskDoneIcon from '../../assets/workbench/task-done.svg'
import taskPendingIcon from '../../assets/workbench/task-pending.svg'
import taskProgressIcon from '../../assets/workbench/task-progress.svg'
import trendNoteIcon from '../../assets/workbench/trend-note.svg'

type NavItem = {
  label: string
  path?: string
  width: number
}

type GradientCard = {
  label: string
  value: string
  trend: string
  icon: string
  gradient: string
}

type QuickAction = {
  label: string
  icon: string
  path?: string
  gradient: string
}

type TaskStatus = 'done' | 'progress' | 'pending'

type TaskItem = {
  title: string
  market: string
  status: string
  meta: string
  tone: TaskStatus
}

const router = useRouter()
const route = useRoute()

const navigationItems: NavItem[] = [
  { label: '工作台', path: '/dramas', width: 66 },
  { label: '商品录入', path: '/dramas/create', width: 80 },
  { label: '合规分析', path: '/compliance', width: 80 },
  { label: '脚本/分镜', path: '/workspace/script', width: 86 },
  { label: '内容创作', path: '/workspace/content', width: 80 },
  { label: '视频剪辑', path: '/workspace/timeline', width: 80 },
  { label: '数据分析', path: '/analytics', width: 80 }
]

const overviewCards: GradientCard[] = [
  {
    label: '待处理商品',
    value: '12',
    trend: '+3',
    icon: overviewBagIcon,
    gradient: 'linear-gradient(135deg, #2b7fff 0%, #00b8db 100%)'
  },
  {
    label: '合规检测完成',
    value: '156',
    trend: '+24',
    icon: overviewComplianceIcon,
    gradient: 'linear-gradient(135deg, #ad46ff 0%, #f6339a 100%)'
  },
  {
    label: '视频已生成',
    value: '89',
    trend: '+15',
    icon: overviewVideoIcon,
    gradient: 'linear-gradient(135deg, #ff6900 0%, #fb2c36 100%)'
  },
  {
    label: '覆盖市场',
    value: '23',
    trend: '+5',
    icon: overviewMarketIcon,
    gradient: 'linear-gradient(135deg, #00c950 0%, #00bc7d 100%)'
  }
]

const quickActions: QuickAction[] = [
  {
    label: '录入新商品',
    icon: quickBagIcon,
    path: '/dramas/create',
    gradient: 'linear-gradient(135deg, #2b7fff 0%, #00b8db 100%)'
  },
  {
    label: '合规检测',
    icon: quickComplianceIcon,
    path: '/compliance',
    gradient: 'linear-gradient(135deg, #ad46ff 0%, #f6339a 100%)'
  },
  {
    label: '内容创作',
    icon: quickVideoIcon,
    path: '/workspace/content',
    gradient: 'linear-gradient(135deg, #ff6900 0%, #fb2c36 100%)'
  },
  {
    label: '数据分析',
    icon: quickChartIcon,
    path: '/analytics',
    gradient: 'linear-gradient(135deg, #00c950 0%, #00bc7d 100%)'
  }
]

const weeklyActivity = [
  { label: '周一', value: 11 },
  { label: '周二', value: 18 },
  { label: '周三', value: 15 },
  { label: '周四', value: 24 },
  { label: '周五', value: 21 },
  { label: '周六', value: 17 },
  { label: '周日', value: 19 }
]

const conversionTrend = [
  { label: '1月', value: 2.4 },
  { label: '2月', value: 3.2 },
  { label: '3月', value: 4.1 },
  { label: '4月', value: 4.8 }
]

const recentTasks: TaskItem[] = [
  {
    title: '智能手表 Pro',
    market: '美国',
    status: '已完成',
    meta: '2小时前',
    tone: 'done'
  },
  {
    title: '无线耳机',
    market: '德国',
    status: '进行中',
    meta: '正在进行',
    tone: 'progress'
  },
  {
    title: '蓝牙音箱',
    market: '日本',
    status: '待处理',
    meta: '待开始',
    tone: 'pending'
  },
  {
    title: '充电宝',
    market: '英国',
    status: '已完成',
    meta: '1天前',
    tone: 'done'
  }
]

const barChart = computed(() => {
  const coordinates = [
    { x: 70.2857, y: 125, width: 19, height: 90, labelX: 91.4286 },
    { x: 123.1429, y: 72.5, width: 19, height: 142.5, labelX: 144.2857 },
    { x: 176, y: 102.5, width: 19, height: 112.5, labelX: 197.1429 },
    { x: 228.8571, y: 27.5, width: 19, height: 187.5, labelX: 250 },
    { x: 281.7143, y: 50, width: 19, height: 165, labelX: 302.8571 },
    { x: 334.5714, y: 80, width: 19, height: 135, labelX: 355.7143 },
    { x: 387.4286, y: 65, width: 19, height: 150, labelX: 408.5714 }
  ]

  return weeklyActivity.map((item, index) => ({
    ...item,
    ...coordinates[index]
  }))
})

const lineChartPoints = computed(() => {
  const coordinates = [
    { x: 65, y: 146 },
    { x: 188.3333, y: 125 },
    { x: 311.6667, y: 101.375 },
    { x: 435, y: 83 }
  ]

  return conversionTrend.map((item, index) => ({
    ...item,
    ...coordinates[index]
  }))
})

const linePath = computed(() =>
  lineChartPoints.value.map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x} ${point.y}`).join(' ')
)

const taskIcons: Record<TaskStatus, string> = {
  done: taskDoneIcon,
  progress: taskProgressIcon,
  pending: taskPendingIcon
}

function navigateTo(path?: string) {
  if (!path || path === route.path) {
    return
  }

  router.push(path)
}

function isActive(path?: string) {
  return path === route.path
}
</script>

<template>
  <div class="workbench-page">
    <div class="workbench-page__mesh" aria-hidden="true"></div>

    <header class="workbench-header">
      <div class="workbench-header__inner">
        <div class="workbench-header__left">
          <button type="button" class="brand-button" @click="navigateTo('/dramas')">
            <span class="brand-button__mark">
              <img :src="brandLogo" alt="" />
            </span>
            <span class="brand-button__text">数字丝路</span>
          </button>

          <nav class="workbench-nav" aria-label="工作台导航">
            <button
              v-for="item in navigationItems"
              :key="item.label"
              type="button"
              class="workbench-nav__item"
              :class="{ 'workbench-nav__item--active': isActive(item.path) }"
              :style="{ width: `${item.width}px` }"
              @click="navigateTo(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="workbench-header__actions">
          <button type="button" class="icon-button" aria-label="消息提醒">
            <img :src="bellIcon" alt="" />
            <span class="icon-button__dot"></span>
          </button>
        </div>
      </div>
    </header>

    <main class="workbench-main">
      <section class="page-title" data-node-id="1:6">
        <h1>工作台总览</h1>
        <p>欢迎回来！这是您的业务概览</p>
      </section>

      <section class="overview-grid" data-node-id="1:11">
        <article
          v-for="card in overviewCards"
          :key="card.label"
          class="overview-card"
        >
          <div class="overview-card__top">
            <span class="overview-card__icon" :style="{ backgroundImage: card.gradient }">
              <img :src="card.icon" alt="" />
            </span>
            <span class="overview-card__trend">
              <img :src="overviewTrendIcon" alt="" />
              {{ card.trend }}
            </span>
          </div>

          <strong class="overview-card__value">{{ card.value }}</strong>
          <p class="overview-card__label">{{ card.label }}</p>
        </article>
      </section>

      <section class="quick-section" data-node-id="1:80">
        <h2>快捷操作</h2>

        <div class="quick-grid">
          <button
            v-for="action in quickActions"
            :key="action.label"
            type="button"
            class="quick-card"
            @click="navigateTo(action.path)"
          >
            <span class="quick-card__icon" :style="{ backgroundImage: action.gradient }">
              <img :src="action.icon" alt="" />
            </span>
            <span class="quick-card__label">{{ action.label }}</span>
          </button>
        </div>
      </section>

      <section class="analytics-grid" data-node-id="1:116">
        <article class="chart-card">
          <h2>本周活动统计</h2>

          <div class="chart-frame">
            <svg viewBox="0 0 440 250" class="chart-svg" role="img" aria-label="本周活动统计柱状图">
              <defs>
                <linearGradient id="bar-card-gradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#8b5cf6" />
                  <stop offset="100%" stop-color="#6d28d9" />
                </linearGradient>
              </defs>

              <g stroke="#dbe8ff" stroke-dasharray="3 3">
                <line x1="64" y1="6" x2="434" y2="6" />
                <line x1="64" y1="58" x2="434" y2="58" />
                <line x1="64" y1="110" x2="434" y2="110" />
                <line x1="64" y1="162" x2="434" y2="162" />
                <line x1="64" y1="214" x2="434" y2="214" />

                <line v-for="bar in barChart" :key="`${bar.label}-grid`" :x1="bar.labelX" y1="6" :x2="bar.labelX" y2="214" />
              </g>

              <g stroke="#90a1b9" stroke-width="1.5">
                <line x1="64" y1="6" x2="64" y2="214" />
                <line x1="64" y1="214" x2="434" y2="214" />
              </g>

              <g fill="#64748b" font-size="12" font-family="'IBM Plex Sans', 'Noto Sans SC', sans-serif">
                <text x="56" y="218" text-anchor="end">0</text>
                <text x="56" y="166" text-anchor="end">7</text>
                <text x="56" y="114" text-anchor="end">14</text>
                <text x="56" y="62" text-anchor="end">21</text>
                <text x="56" y="10" text-anchor="end">28</text>
              </g>

              <g>
                <rect
                  v-for="bar in barChart"
                  :key="bar.label"
                  :x="bar.x"
                  :y="bar.y"
                  :width="bar.width"
                  :height="bar.height"
                  rx="10"
                  fill="url(#bar-card-gradient)"
                />
              </g>

              <g fill="#64748b" font-size="12" font-family="'IBM Plex Sans', 'Noto Sans SC', sans-serif">
                <text
                  v-for="bar in barChart"
                  :key="`${bar.label}-label`"
                  :x="bar.labelX"
                  y="230"
                  text-anchor="middle"
                >
                  {{ bar.label }}
                </text>
              </g>
            </svg>
          </div>
        </article>

        <article class="chart-card">
          <h2>转化率趋势</h2>

          <div class="chart-frame">
            <svg viewBox="0 0 440 250" class="chart-svg" role="img" aria-label="转化率趋势折线图">
              <g stroke="#dbe8ff" stroke-dasharray="3 3">
                <line x1="64" y1="6" x2="434" y2="6" />
                <line x1="64" y1="58" x2="434" y2="58" />
                <line x1="64" y1="110" x2="434" y2="110" />
                <line x1="64" y1="162" x2="434" y2="162" />
                <line x1="64" y1="214" x2="434" y2="214" />

                <line v-for="point in lineChartPoints" :key="`${point.label}-grid`" :x1="point.x" y1="6" :x2="point.x" y2="214" />
              </g>

              <g stroke="#90a1b9" stroke-width="1.5">
                <line x1="64" y1="6" x2="64" y2="214" />
                <line x1="64" y1="214" x2="434" y2="214" />
              </g>

              <g fill="#64748b" font-size="12" font-family="'IBM Plex Sans', 'Noto Sans SC', sans-serif">
                <text x="56" y="218" text-anchor="end">0</text>
                <text x="56" y="166" text-anchor="end">2</text>
                <text x="56" y="114" text-anchor="end">4</text>
                <text x="56" y="62" text-anchor="end">6</text>
                <text x="56" y="10" text-anchor="end">8</text>
              </g>

              <path :d="linePath" fill="none" stroke="#f97316" stroke-width="3" />

              <circle
                v-for="point in lineChartPoints"
                :key="point.label"
                :cx="point.x"
                :cy="point.y"
                r="7"
                fill="#f97316"
              />

              <g fill="#64748b" font-size="12" font-family="'IBM Plex Sans', 'Noto Sans SC', sans-serif">
                <text
                  v-for="point in lineChartPoints"
                  :key="`${point.label}-label`"
                  :x="point.x"
                  y="230"
                  text-anchor="middle"
                >
                  {{ point.label }}
                </text>
              </g>
            </svg>
          </div>

          <div class="trend-note">
            <img :src="trendNoteIcon" alt="" />
            <span>转化率提升 100% 相比上季度</span>
          </div>
        </article>
      </section>

      <section class="tasks-card" data-node-id="1:285">
        <div class="tasks-card__header">
          <h2>最近任务</h2>
          <button type="button" class="tasks-card__all">
            查看全部
            <img :src="recentArrowIcon" alt="" />
          </button>
        </div>

        <div class="tasks-list">
          <article v-for="task in recentTasks" :key="task.title" class="task-row">
            <div class="task-row__info">
              <img class="task-row__status-icon" :src="taskIcons[task.tone]" alt="" />
              <div class="task-row__copy">
                <h3>{{ task.title }}</h3>
                <p>目标市场: {{ task.market }}</p>
              </div>
            </div>

            <div class="task-row__meta">
              <strong>{{ task.status }}</strong>
              <span>{{ task.meta }}</span>
            </div>
          </article>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.workbench-page {
  --wb-navy: #0a2463;
  --wb-text: #45556c;
  --wb-text-strong: #314158;
  --wb-text-soft: #62748e;
  --wb-line: #e2e8f0;
  --wb-card: #ffffff;
  --wb-surface: #f8fafc;
  --wb-cyan: #06b6d4;
  --wb-orange: #f97316;
  --wb-green: #10b981;
  position: relative;
  min-height: 100vh;
  background:
    radial-gradient(circle at left top, rgba(6, 182, 212, 0.07), transparent 34%),
    radial-gradient(circle at right top, rgba(124, 58, 237, 0.06), transparent 28%),
    linear-gradient(180deg, #f8fbff 0%, #ffffff 34%);
  color: var(--wb-text-strong);
  font-family: "IBM Plex Sans", "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif;
}

.workbench-page__mesh {
  pointer-events: none;
  position: absolute;
  inset: 64px 0 0;
  background-image:
    linear-gradient(rgba(186, 230, 253, 0.32) 1px, transparent 1px),
    linear-gradient(90deg, rgba(186, 230, 253, 0.32) 1px, transparent 1px);
  background-position: center top;
  background-size: 76px 76px;
  mask-image: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(255, 255, 255, 0.28));
  opacity: 0.36;
}

.workbench-header {
  position: relative;
  z-index: 2;
  background: #ffffff;
  border-bottom: 1px solid var(--wb-line);
}

.workbench-header__inner {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  width: min(1075px, 100%);
  margin: 0 auto;
  padding: 0 24px;
}

.workbench-header__left {
  display: flex;
  align-items: center;
  gap: 32px;
  min-width: 0;
}

.brand-button {
  display: flex;
  align-items: center;
  gap: 12px;
  border: 0;
  background: transparent;
  padding: 0;
  cursor: pointer;
}

.brand-button__mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 12px;
  background: linear-gradient(135deg, #0a2463 0%, #06b6d4 50%, #7c3aed 100%);
  box-shadow: 0 8px 16px rgba(10, 36, 99, 0.18);
}

.brand-button__mark img {
  width: 20px;
  height: 20px;
}

.brand-button__text {
  color: var(--wb-navy);
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  white-space: nowrap;
}

.workbench-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.workbench-nav__item {
  min-width: 0;
  height: 32px;
  padding: 0 12px;
  border: 0;
  border-radius: 12px;
  background: transparent;
  color: var(--wb-text);
  font-size: 14px;
  line-height: 20px;
  font-weight: 500;
  white-space: nowrap;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.workbench-nav__item:hover {
  background: rgba(6, 182, 212, 0.08);
  color: var(--wb-navy);
}

.workbench-nav__item--active {
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
  color: var(--wb-navy);
}

.workbench-header__actions {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.icon-button {
  position: relative;
  width: 36px;
  height: 36px;
  border: 0;
  border-radius: 12px;
  background: transparent;
  cursor: pointer;
}

.icon-button img {
  width: 20px;
  height: 20px;
}

.icon-button__dot {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--wb-orange);
}

.workbench-main {
  position: relative;
  z-index: 1;
  box-sizing: border-box;
  width: min(1075px, 100%);
  margin: 0 auto;
  padding: 39px 32px 48px;
}

.page-title {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 32px;
}

.page-title h1 {
  margin: 0;
  color: var(--wb-navy);
  font-family: "Urbanist", "Noto Sans SC", "PingFang SC", sans-serif;
  font-size: 36px;
  line-height: 36px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.page-title p {
  margin: 0;
  color: var(--wb-text);
  font-size: 16px;
  line-height: 24px;
  font-weight: 400;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 24px;
  margin-bottom: 32px;
}

.overview-card,
.quick-card,
.chart-card,
.tasks-card {
  background: var(--wb-card);
  border: 1px solid var(--wb-line);
  box-shadow: 0 1px 1px rgba(15, 23, 42, 0.02);
}

.overview-card {
  height: 174px;
  border-radius: 16px;
  padding: 24px;
}

.overview-card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.overview-card__icon,
.quick-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 16px;
}

.overview-card__icon img,
.quick-card__icon img {
  width: 24px;
  height: 24px;
}

.overview-card__trend {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--wb-green);
  font-size: 14px;
  line-height: 20px;
}

.overview-card__trend img {
  width: 16px;
  height: 16px;
}

.overview-card__value {
  display: block;
  margin-bottom: 4px;
  color: var(--wb-navy);
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
}

.overview-card__label {
  margin: 0;
  color: var(--wb-text);
  font-size: 14px;
  line-height: 20px;
}

.quick-section {
  margin-bottom: 32px;
}

.quick-section h2,
.chart-card h2,
.tasks-card h2 {
  margin: 0;
  color: var(--wb-navy);
  font-family: "Urbanist", "Noto Sans SC", "PingFang SC", sans-serif;
  font-size: 20px;
  line-height: 28px;
  font-weight: 600;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.quick-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 16px;
  height: 138px;
  padding: 24px;
  border-radius: 16px;
  text-align: left;
  cursor: pointer;
}

.quick-card__label {
  color: var(--wb-navy);
  font-size: 16px;
  line-height: 24px;
  font-weight: 500;
}

.analytics-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 32px;
  margin-bottom: 32px;
}

.chart-card {
  display: flex;
  flex-direction: column;
  height: 420px;
  border-radius: 16px;
  padding: 25px 25px 25px;
}

.chart-card h2 {
  margin-bottom: 24px;
  font-size: 18px;
}

.chart-frame {
  width: 100%;
  height: 250px;
}

.chart-svg {
  display: block;
  width: 100%;
  height: 100%;
}

.trend-note {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 52px;
  padding: 16px;
  margin-top: 16px;
  border-radius: 16px;
  background: linear-gradient(90deg, #fff7ed 0%, #fef2f2 100%);
  color: var(--wb-orange);
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.trend-note img {
  width: 16px;
  height: 16px;
}

.tasks-card {
  min-height: 454px;
  border-radius: 16px;
  padding: 25px;
}

.tasks-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.tasks-card__all {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 0;
  background: transparent;
  padding: 0;
  color: var(--wb-cyan);
  font-size: 14px;
  line-height: 20px;
  font-weight: 500;
  cursor: pointer;
}

.tasks-card__all img {
  width: 16px;
  height: 16px;
}

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 76px;
  padding: 16px;
  border-radius: 16px;
  background: var(--wb-surface);
}

.task-row__info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.task-row__status-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.task-row__copy h3 {
  margin: 0;
  color: var(--wb-navy);
  font-size: 16px;
  line-height: 24px;
  font-weight: 500;
}

.task-row__copy p,
.task-row__meta span {
  margin: 0;
  color: var(--wb-text-soft);
  font-size: 14px;
  line-height: 20px;
}

.task-row__meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  min-width: 52px;
}

.task-row__meta strong {
  color: var(--wb-text-strong);
  font-size: 14px;
  line-height: 20px;
  font-weight: 500;
}

@media (max-width: 1160px) {
  .overview-grid,
  .quick-grid,
  .analytics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workbench-header__inner {
    gap: 20px;
  }

  .workbench-header__left {
    gap: 20px;
  }
}

@media (max-width: 860px) {
  .workbench-page__mesh {
    inset: 64px 0 0;
  }

  .workbench-header__inner {
    flex-wrap: wrap;
    height: auto;
    padding-top: 16px;
    padding-bottom: 16px;
  }

  .workbench-header__left,
  .workbench-nav {
    flex-wrap: wrap;
  }

  .workbench-main {
    width: 100%;
    padding: 32px 16px 40px;
  }

  .overview-grid,
  .quick-grid,
  .analytics-grid {
    grid-template-columns: 1fr;
  }

  .task-row {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
  }

  .task-row__meta {
    align-items: flex-start;
  }
}

@media (max-width: 640px) {
  .page-title h1 {
    font-size: 30px;
    line-height: 38px;
  }

  .chart-card,
  .tasks-card,
  .overview-card,
  .quick-card {
    padding-left: 18px;
    padding-right: 18px;
  }

  .chart-card {
    height: auto;
  }

  .chart-frame {
    overflow-x: auto;
  }

  .chart-svg {
    min-width: 440px;
  }
}
</style>
