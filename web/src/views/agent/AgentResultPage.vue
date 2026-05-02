<template>
  <main class="agent-result-page" aria-labelledby="agent-result-title">
    <div class="agent-result-page__ambient agent-result-page__ambient--cyan" aria-hidden="true"></div>
    <div class="agent-result-page__ambient agent-result-page__ambient--orange" aria-hidden="true"></div>

    <div class="agent-result-shell">
      <header class="agent-result-header">
        <div class="agent-result-header__main">
          <button type="button" class="agent-icon-button agent-result-header__back" aria-label="返回首页" @click="goHome">
            <ArrowLeft class="agent-icon-button__icon" aria-hidden="true" />
          </button>

          <div class="agent-result-header__copy">
            <span class="agent-agent-badge">
              <Cpu class="agent-agent-badge__icon" aria-hidden="true" />
              <span>SILK ROAD AGENT</span>
            </span>
            <h1 id="agent-result-title">{{ pageTitle }}</h1>
            <p>{{ pageDescription }}</p>
          </div>
        </div>

        <div class="agent-result-header__status" aria-label="Agent 任务状态">
          <span class="agent-status-pill agent-status-pill--success">
            <CircleCheck class="agent-status-pill__icon" aria-hidden="true" />
            <span>分析完成</span>
          </span>
        </div>
      </header>

      <section v-if="!storedResult" class="agent-card agent-missing-card">
        <div class="agent-section-title">
          <WarningFilled class="agent-section-title__loose-icon agent-section-title__loose-icon--violet" aria-hidden="true" />
          <h2>没有找到本次 Agent 结果</h2>
        </div>
        <p>请返回首页重新填写商品信息并启动丝路 Agent。生成结果会临时保存在当前浏览器会话中，刷新结果页不会丢失。</p>
        <button type="button" class="agent-button agent-button--primary" @click="goHome">
          <span>返回首页</span>
          <ArrowRight class="agent-button__icon" aria-hidden="true" />
        </button>
      </section>

      <template v-else>
      <section class="agent-overview-card" aria-labelledby="agent-overview-title">
        <div class="agent-overview-card__content">
          <div class="agent-section-title">
            <MagicStick class="agent-section-title__loose-icon" aria-hidden="true" />
            <h2 id="agent-overview-title">本次 Agent 方案总览</h2>
          </div>

          <div class="agent-overview-card__items" aria-label="本次 Agent 核心结论">
            <article v-for="item in overviewItems" :key="item.label" class="agent-overview-item" :title="item.value">
              <span class="agent-overview-item__icon" :class="`agent-overview-item__icon--${item.tone}`">
                <component :is="item.icon" aria-hidden="true" />
              </span>
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </article>
          </div>
        </div>
      </section>

      <section class="agent-result-grid" aria-label="Agent 详细结果">
        <div class="agent-result-grid__main">
          <article class="agent-card compliance-result-card">
            <div class="agent-card__header">
              <div class="agent-section-title">
                <span class="agent-title-icon agent-title-icon--blue">
                  <DocumentChecked aria-hidden="true" />
                </span>
                <h2>合规分析结果</h2>
              </div>
              <span class="agent-risk-badge" :class="riskBadgeClass">{{ result.overview.complianceRiskLevel }}</span>
            </div>

            <p class="compliance-result-card__body">
              {{ result.compliance.summary }}
            </p>

            <div class="compliance-risk-box">
              <span class="compliance-risk-box__label">风险词 / 敏感表达</span>
              <div class="compliance-risk-box__tags">
                <span v-for="word in riskWords" :key="word">{{ word }}</span>
              </div>
              <p>
                <WarningFilled class="compliance-risk-box__icon" aria-hidden="true" />
                <span>建议使用更克制的安全表达，避免绝对化和医疗化承诺。</span>
              </p>
            </div>

            <ul class="agent-mini-list">
              <li v-for="item in result.compliance.suggestions" :key="item">{{ item }}</li>
            </ul>
          </article>

          <article class="agent-card video-script-card">
            <div class="agent-card__header">
              <div class="agent-section-title">
                <span class="agent-title-icon agent-title-icon--orange">
                  <VideoCamera aria-hidden="true" />
                </span>
                <h2>短视频脚本</h2>
              </div>
              <span class="agent-safe-badge">
                <CircleCheck class="agent-safe-badge__icon" aria-hidden="true" />
                <span>已规避高风险功效表达</span>
              </span>
            </div>

            <div class="video-script-card__timeline">
              <article v-for="part in scriptParts" :key="part.title" class="script-step" :class="`script-step--${part.tone}`">
                <div class="script-step__meta">
                  <span class="script-step__dot" aria-hidden="true"></span>
                  <div>
                    <strong>{{ part.title }}</strong>
                    <span>{{ part.time }}</span>
                  </div>
                </div>
                <p>{{ part.text }}</p>
              </article>
            </div>

            <div v-if="result.script.storyboard?.length" class="storyboard-list">
              <article v-for="shot in result.script.storyboard" :key="`${shot.shot}-${shot.subtitle}`" class="storyboard-list__item">
                <strong>{{ shot.shot }}</strong>
                <p>{{ shot.visual }}</p>
                <span>{{ shot.subtitle || shot.voiceover }}</span>
              </article>
            </div>
          </article>
        </div>

        <aside class="agent-result-grid__side" aria-label="Agent 辅助建议">
          <article class="agent-card localization-card">
            <div class="agent-card__header">
              <div class="agent-section-title">
                <span class="agent-title-icon agent-title-icon--violet">
                  <Location aria-hidden="true" />
                </span>
                <h2>本地化营销方向</h2>
              </div>
            </div>

            <div class="localization-card__focus">
              <span>推荐方向</span>
              <strong>{{ result.localization.direction }}</strong>
            </div>
            <p>{{ result.localization.reason }}</p>
            <div class="agent-chip-row localization-card__tags">
              <span v-for="keyword in result.localization.keywords" :key="keyword" class="agent-chip agent-chip--violet">{{ keyword }}</span>
            </div>
          </article>

          <article class="agent-card digital-human-card">
            <div class="agent-card__header">
              <div class="agent-section-title">
                <span class="agent-title-icon agent-title-icon--pink">
                  <User aria-hidden="true" />
                </span>
                <h2>数字人方案</h2>
              </div>
            </div>

            <dl class="agent-info-list">
              <div v-for="row in digitalHumanRows" :key="row.label">
                <dt>{{ row.label }}</dt>
                <dd>{{ row.value }}</dd>
              </div>
            </dl>
          </article>

          <article class="agent-card launch-suggestion-card">
            <div class="agent-card__header">
              <div class="agent-section-title">
                <span class="agent-title-icon agent-title-icon--green">
                  <TrendCharts aria-hidden="true" />
                </span>
                <h2>投放建议</h2>
              </div>
            </div>

            <div class="launch-suggestion-card__group">
              <span>推荐平台</span>
              <div class="agent-chip-row">
                <span v-for="platform in result.promotion.platforms" :key="platform" class="agent-chip agent-chip--platform">{{ platform }}</span>
              </div>
            </div>

            <div class="launch-suggestion-card__group launch-suggestion-card__group--metrics">
              <span>推荐内容</span>
              <div class="agent-chip-row">
                <span v-for="scene in launchScenes" :key="scene" class="agent-chip agent-chip--violet">{{ scene }}</span>
              </div>
            </div>

            <div class="launch-suggestion-card__group">
              <span>重点指标</span>
              <div class="agent-chip-row">
                <span v-for="metric in launchMetrics" :key="metric" class="agent-chip agent-chip--orange">{{ metric }}</span>
              </div>
            </div>

            <div class="launch-suggestion-card__advice">
              <span>优化建议</span>
              <p>{{ result.promotion.optimizationAdvice }}</p>
            </div>
          </article>
        </aside>
      </section>

      <section class="agent-card agent-reminder-card" aria-labelledby="agent-reminder-title">
        <div class="agent-card__header">
          <div class="agent-section-title">
            <ChatDotRound class="agent-section-title__loose-icon agent-section-title__loose-icon--violet" aria-hidden="true" />
            <h2 id="agent-reminder-title">与丝路 Agent 对话</h2>
          </div>
          <span class="agent-reminder-card__time">最近一次回复 · 刚刚</span>
        </div>

        <div class="agent-reminder-card__body">
          <span class="agent-reminder-card__avatar" aria-hidden="true">
            <Cpu />
          </span>
          <div class="agent-reminder-card__content">
            <div class="agent-gap-card">
              <div class="agent-gap-card__text">
                <span class="agent-gap-card__icon" aria-hidden="true">
                  <WarningFilled />
                </span>
                <div>
                  <h3>Agent 发现的关键信息缺口</h3>
                  <p>{{ result.agentMessage.missingInfoNotice }}</p>
                </div>
              </div>
              <div class="agent-gap-card__actions">
                <button
                  v-for="(action, index) in quickActions"
                  :key="action"
                  type="button"
                  class="agent-gap-card__button"
                  :class="{ 'agent-gap-card__button--primary': index === 0, 'agent-gap-card__button--quiet': index > 1 }"
                  @click="markReminderAction(action)"
                >
                  {{ action }}
                </button>
              </div>
            </div>

            <p v-if="reminderAction" class="agent-reminder-card__feedback">{{ reminderAction }}已记录，后续优化会基于该选择继续。</p>
          </div>
        </div>

        <div class="agent-reminder-input" aria-label="继续与丝路 Agent 对话">
          <input
            v-model="agentMessage"
            class="agent-reminder-input__field"
            type="text"
            aria-label="继续告诉丝路 Agent"
            placeholder="输入补充要求，例如换成印尼市场、语气更年轻一点……"
            @keyup.enter="sendAgentMessage"
          />
          <button type="button" class="agent-reminder-input__button" @click="sendAgentMessage">
            <Promotion class="agent-reminder-input__icon" aria-hidden="true" />
            <span>发送</span>
          </button>
        </div>
      </section>

      </template>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { Component } from 'vue'
import {
  ArrowLeft,
  ArrowRight,
  ChatDotRound,
  CircleCheck,
  Cpu,
  DocumentChecked,
  Goods,
  Location,
  MagicStick,
  Promotion,
  TrendCharts,
  User,
  VideoCamera,
  WarningFilled
} from '@element-plus/icons-vue'
import type { AgentResult } from '@/types/agent'
import { readAgentResult } from '@/utils/agentStorage'

defineOptions({
  name: 'AgentResultPage'
})

type OverviewItem = {
  label: string
  value: string
  tone: 'orange' | 'cyan' | 'pink' | 'violet'
  icon: Component
}

const router = useRouter()
const reminderAction = ref('')
const agentMessage = ref('')
const storedResult = ref<AgentResult | null>(readAgentResult())

const fallbackResult: AgentResult = {
  recognizedInfo: {
    productName: '待分析商品',
    category: '未填写类目',
    targetMarket: '目标市场待补充',
    targetPlatform: '目标平台待补充',
    targetAudience: '目标人群待补充',
    coreSellingPoints: ['核心卖点待补充'],
    imageUnderstanding: '未找到本次图片理解结果。'
  },
  overview: {
    complianceRiskLevel: '中风险',
    marketStrategy: '补充商品信息后再生成完整方案。',
    recommendedVideoStyle: '生活场景化竖屏短视频',
    recommendedDigitalHuman: '亲和型本地化讲解数字人'
  },
  compliance: {
    title: '合规分析结果',
    summary: '暂无本次 Agent 结果。',
    riskTags: ['信息缺失'],
    missingInfo: ['商品信息'],
    suggestions: ['返回首页重新启动丝路 Agent。'],
    forbiddenExpressions: ['100%安全', '治疗', '保证通过'],
    saferExpressions: ['适合日常使用', '建议查看材质说明']
  },
  localization: {
    direction: '待生成',
    reason: '暂无本次 Agent 结果。',
    keywords: ['lifestyle'],
    tone: '自然可信',
    sceneSuggestions: ['开箱展示']
  },
  script: {
    title: '短视频脚本',
    duration: '20-25s',
    opening: { time: '0-3s', content: '返回首页重新生成后展示开头脚本。' },
    middle: { time: '3-20s', content: '返回首页重新生成后展示中段脚本。' },
    ending: { time: '20-25s', content: '返回首页重新生成后展示结尾脚本。' },
    storyboard: []
  },
  digitalHuman: {
    persona: '待生成',
    tone: '待生成',
    videoRatio: '9:16',
    subtitleAdvice: '待生成',
    visualStyle: '待生成',
    shootingStyle: '待生成'
  },
  promotion: {
    platforms: ['目标平台待补充'],
    contentTags: ['内容方向待生成'],
    focusMetrics: ['重点指标待生成'],
    optimizationAdvice: '暂无本次 Agent 结果。'
  },
  agentMessage: {
    summary: '没有找到本次 Agent 结果，请返回首页重新生成。',
    missingInfoNotice: '暂无信息缺口。',
    quickActions: ['返回首页']
  }
}

const result = computed(() => storedResult.value ?? fallbackResult)

const pageTitle = computed(() => {
  const productName = result.value.recognizedInfo.productName
  return productName && productName !== '待分析商品'
    ? `丝路 Agent 已生成「${productName}」出海营销方案`
    : '丝路 Agent 已生成出海营销方案'
})

const pageDescription = computed(() => {
  const info = result.value.recognizedInfo
  return `基于${info.targetMarket || '目标市场'}、${info.targetPlatform || '目标平台'}与商品信息，自动完成合规判断、本地化内容、数字人方案与投放建议。`
})

const overviewItems = computed<OverviewItem[]>(() => [
  { label: '合规风险', value: result.value.overview.complianceRiskLevel || '待评估', tone: 'orange', icon: WarningFilled },
  { label: '推荐市场策略', value: toOverviewSnippet(result.value.overview.marketStrategy, '补齐信息后投放'), tone: 'cyan', icon: Goods },
  { label: '推荐视频形式', value: toOverviewSnippet(result.value.overview.recommendedVideoStyle, '生活场景竖屏短视频', 12), tone: 'pink', icon: VideoCamera }
])

const riskWords = computed(() => {
  const expressions = result.value.compliance.forbiddenExpressions?.length
    ? result.value.compliance.forbiddenExpressions
    : result.value.compliance.riskTags
  return expressions.length ? expressions : ['暂无高风险表达']
})

const riskBadgeClass = computed(() => {
  const risk = result.value.overview.complianceRiskLevel || ''
  if (risk.includes('高')) return 'agent-risk-badge--high'
  if (risk.includes('中')) return 'agent-risk-badge--medium'
  return 'agent-risk-badge--low'
})

const scriptParts = computed(() => [
  {
    title: '开头',
    time: result.value.script.opening.time || '0-3s',
    tone: 'blue',
    text: result.value.script.opening.content || '待生成开头脚本。'
  },
  {
    title: '中段',
    time: result.value.script.middle.time || '3-20s',
    tone: 'violet',
    text: result.value.script.middle.content || '待生成中段脚本。'
  },
  {
    title: '结尾',
    time: result.value.script.ending.time || '20-25s',
    tone: 'orange',
    text: result.value.script.ending.content || '待生成结尾脚本。'
  }
])

const digitalHumanRows = computed(() => [
  { label: '推荐数字人', value: result.value.digitalHuman.persona || '待生成' },
  { label: '口播风格', value: result.value.digitalHuman.tone || '待生成' },
  { label: '视频比例', value: result.value.digitalHuman.videoRatio || '9:16' },
  { label: '字幕建议', value: result.value.digitalHuman.subtitleAdvice || '待生成' },
  { label: '画面风格', value: result.value.digitalHuman.visualStyle || '待生成' },
  { label: '拍摄方式', value: result.value.digitalHuman.shootingStyle || '待生成' }
])

const launchScenes = computed(() => result.value.promotion.contentTags?.length ? result.value.promotion.contentTags : result.value.localization.sceneSuggestions)
const launchMetrics = computed(() => result.value.promotion.focusMetrics?.length ? result.value.promotion.focusMetrics : ['完播率', '点击率', '收藏率'])
const quickActions = computed(() => result.value.agentMessage.quickActions?.length ? result.value.agentMessage.quickActions : ['我来补充', '先按常见情况分析', '暂时忽略'])

const toOverviewSnippet = (value: string | undefined, fallback: string, maxLength = 16) => {
  const normalized = (value || fallback).replace(/\s+/g, '')
  const firstClause = normalized
    .split(/[，,。；;：:]/)
    .map((item) => item.trim())
    .find(Boolean) || fallback
  return firstClause.length > maxLength ? `${firstClause.slice(0, maxLength)}…` : firstClause
}

const goHome = () => {
  router.push('/')
}

const markReminderAction = (action: string) => {
  reminderAction.value = action
}

const sendAgentMessage = () => {
  const message = agentMessage.value.trim()
  if (!message) return

  reminderAction.value = `“${message}”`
  agentMessage.value = ''
}
</script>

<style scoped>
.agent-result-page {
  position: relative;
  min-height: 100vh;
  overflow: visible;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', sans-serif;
}

:global(.app-shell:has(.agent-result-page)) {
  overflow: visible;
}

:global(.app-content:has(.agent-result-page)) {
  display: block;
}

.agent-result-page__ambient {
  position: absolute;
  z-index: 0;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(64px);
}

.agent-result-page__ambient--cyan {
  top: 80px;
  left: calc(50% - 512px);
  width: 288px;
  height: 288px;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.1), rgba(124, 58, 237, 0.1));
}

.agent-result-page__ambient--orange {
  top: 240px;
  right: calc(50% - 512px);
  width: 384px;
  height: 384px;
  background: linear-gradient(135deg, rgba(249, 115, 22, 0.1), rgba(6, 182, 212, 0.1));
}

.agent-result-shell {
  position: relative;
  z-index: 1;
  width: min(1039px, calc(100% - 64px));
  margin: 0 auto;
  padding: 40px 0;
}

.agent-result-header {
  min-height: 136px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.agent-result-header__main {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.agent-icon-button {
  width: 40px;
  height: 40px;
  padding: 0;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: #ffffff;
  color: #45556c;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 180ms ease, border-color 180ms ease, color 180ms ease, box-shadow 180ms ease;
}

.agent-icon-button:hover {
  border-color: rgba(6, 182, 212, 0.45);
  color: #0891b2;
  transform: translateY(-1px);
  box-shadow: 0 12px 24px -18px rgba(15, 23, 42, 0.35);
}

.agent-icon-button__icon {
  width: 16px;
  height: 16px;
}

.agent-result-header__copy {
  min-width: 0;
}

.agent-agent-badge {
  min-height: 27px;
  padding: 5px 12px;
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1), rgba(124, 58, 237, 0.1));
  color: #06b6d4;
  font-size: 11px;
  line-height: 17px;
  font-weight: 700;
}

.agent-agent-badge__icon {
  width: 14px;
  height: 14px;
  color: #7c3aed;
}

.agent-result-header h1 {
  margin: 8px 0 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 36px;
  line-height: 45px;
  font-weight: 700;
}

.agent-result-header p {
  max-width: 660px;
  margin: 8px 0 0;
  color: #45556c;
  font-size: 16px;
  line-height: 24px;
}

.agent-result-header__status {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.agent-status-pill {
  min-height: 27px;
  padding: 5px 10px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  line-height: 17px;
  font-weight: 700;
  white-space: nowrap;
}

.agent-status-pill__icon {
  width: 12px;
  height: 12px;
}

.agent-status-pill--success {
  border: 1px solid #b9f8cf;
  background: #f0fdf4;
  color: #00a63e;
}

.agent-status-pill--cyan {
  border: 1px solid rgba(6, 182, 212, 0.3);
  background: rgba(6, 182, 212, 0.1);
  color: #0891b2;
}

.agent-card {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #ffffff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.1);
}

.agent-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.agent-section-title {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.agent-section-title h2 {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 18px;
  line-height: 27px;
  font-weight: 700;
}

.agent-section-title--large h2 {
  font-size: 20px;
  line-height: 30px;
}

.agent-section-title__loose-icon {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
  color: #7c3aed;
}

.agent-section-title__loose-icon--cyan {
  color: #06b6d4;
}

.agent-section-title__loose-icon--violet {
  color: #7c3aed;
}

.agent-title-icon {
  width: 28px;
  height: 28px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  color: #ffffff;
}

.agent-title-icon svg {
  width: 14px;
  height: 14px;
}

.agent-title-icon--gradient {
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
}

.agent-title-icon--blue {
  width: 32px;
  height: 32px;
  border-radius: 16px;
  background: #3b82f6;
}

.agent-title-icon--orange {
  width: 32px;
  height: 32px;
  border-radius: 16px;
  background: #f97316;
}

.agent-title-icon--violet {
  width: 28px;
  height: 28px;
  border-radius: 12px;
  background: #a855f7;
}

.agent-title-icon--pink {
  width: 28px;
  height: 28px;
  border-radius: 12px;
  background: #d946ef;
}

.agent-title-icon--green {
  width: 28px;
  height: 28px;
  border-radius: 12px;
  background: #10b981;
}

.agent-text-cyan {
  color: #0891b2;
  font-weight: 700;
}

.agent-text-violet {
  color: #7c3aed;
  font-weight: 700;
}

.agent-text-orange {
  color: #f97316;
  font-weight: 700;
}

.agent-button {
  min-height: 38px;
  border: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-family: inherit;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: transform 180ms ease, box-shadow 180ms ease, border-color 180ms ease, background 180ms ease;
}

.agent-button__icon {
  width: 14px;
  height: 14px;
}

.agent-button--ghost {
  width: 110px;
  padding: 8px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  color: #0a2463;
  font-size: 14px;
  line-height: 20px;
}

.agent-button--ghost:hover {
  border-color: rgba(6, 182, 212, 0.38);
  transform: translateY(-1px);
}

.agent-button--primary {
  min-height: 52px;
  padding: 14px 24px;
  border-radius: 16px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 50%, #f97316 100%);
  color: #ffffff;
  font-size: 16px;
  line-height: 24px;
  box-shadow: 0 18px 30px -18px rgba(124, 58, 237, 0.55);
}

.agent-button--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 22px 34px -18px rgba(124, 58, 237, 0.6);
}

.agent-summary-card {
  margin-top: 32px;
  min-height: 156px;
  padding: 25px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.agent-summary-card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.agent-summary-card__copy {
  min-width: 0;
  flex: 1 1 auto;
}

.agent-summary-card__prompt {
  margin: 8px 0 0 36px;
  color: #314158;
  font-size: 16px;
  line-height: 26px;
}

.agent-summary-card__tags {
  margin-left: 36px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.agent-info-tag {
  min-height: 26px;
  padding: 5px 13px;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(90deg, #f8fafc 0%, #ffffff 100%);
  color: #0a2463;
  font-size: 12px;
  line-height: 16px;
  font-weight: 700;
}

.agent-info-tag__label {
  color: #90a1b9;
}

.agent-overview-card {
  min-height: 112px;
  margin-top: 24px;
  padding: 25px;
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 16px;
  display: flex;
  align-items: stretch;
  overflow: hidden;
  background: linear-gradient(105deg, rgba(103, 232, 249, 0.35) 0%, rgba(255, 255, 255, 0.82) 54%, rgba(124, 58, 237, 0.28) 100%);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1);
}

.agent-overview-card__content {
  min-width: 0;
  width: 100%;
  flex: 1 1 auto;
}

.agent-overview-card__items {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.agent-overview-item {
  min-height: 64px;
  padding: 12px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr);
  grid-template-rows: auto auto;
  column-gap: 10px;
  row-gap: 4px;
  align-items: center;
  align-content: center;
  background: rgba(255, 255, 255, 0.82);
}

.agent-overview-item__icon {
  grid-row: 1 / span 2;
  width: 24px;
  height: 24px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
}

.agent-overview-item__icon svg {
  width: 12px;
  height: 12px;
}

.agent-overview-item__icon--orange {
  background: #f59e0b;
}

.agent-overview-item__icon--cyan {
  background: #06b6d4;
}

.agent-overview-item__icon--pink {
  background: #f43f5e;
}

.agent-overview-item__icon--violet {
  background: #8b5cf6;
}

.agent-overview-item span:not(.agent-overview-item__icon) {
  color: #90a1b9;
  font-size: 10px;
  line-height: 15px;
}

.agent-overview-item strong {
  min-width: 0;
  display: -webkit-box;
  overflow: hidden;
  color: #0a2463;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.agent-result-grid {
  margin-top: 24px;
  display: grid;
  grid-template-columns: minmax(0, 684.664fr) minmax(320px, 330.336fr);
  gap: 24px;
}

.agent-result-grid__main,
.agent-result-grid__side {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.compliance-result-card,
.video-script-card,
.localization-card,
.digital-human-card,
.launch-suggestion-card,
.agent-reminder-card {
  padding: 25px;
}

.localization-card,
.digital-human-card,
.launch-suggestion-card {
  padding: 21px;
}

.compliance-result-card {
  min-height: 245px;
}

.compliance-result-card__body {
  margin: 14px 0 18px;
  color: #314158;
  font-size: 14px;
  line-height: 24px;
  font-weight: 600;
}

.agent-risk-badge {
  min-height: 25px;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 16px;
  font-weight: 700;
}

.agent-risk-badge--low {
  border: 1px solid #b9f8cf;
  background: #f0fdf4;
  color: #008236;
}

.agent-risk-badge--medium {
  border: 1px solid #fee685;
  background: #fefce8;
  color: #ca8a04;
}

.agent-risk-badge--high {
  border: 1px solid #fecaca;
  background: #fef2f2;
  color: #dc2626;
}

.compliance-risk-box {
  min-height: 84px;
  padding: 17px;
  border: 1px solid rgba(251, 44, 54, 0.2);
  border-radius: 12px;
  background: rgba(254, 242, 242, 0.72);
}

.compliance-risk-box__label {
  display: block;
  color: #fb2c36;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
}

.compliance-risk-box__tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.compliance-risk-box__tags span {
  min-height: 22px;
  padding: 4px 10px;
  border-radius: 999px;
  background: #ffe2e2;
  color: #e7000b;
  font-size: 10px;
  line-height: 14px;
  font-weight: 700;
}

.compliance-risk-box p {
  margin: 8px 0 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #fb2c36;
  font-size: 11px;
  line-height: 17px;
}

.agent-mini-list {
  margin: 12px 0 0;
  padding: 0;
  display: grid;
  gap: 6px;
  list-style: none;
}

.agent-mini-list li {
  position: relative;
  padding-left: 14px;
  color: #45556c;
  font-size: 12px;
  line-height: 20px;
}

.agent-mini-list li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 9px;
  width: 5px;
  height: 5px;
  border-radius: 999px;
  background: #06b6d4;
}

.compliance-risk-box__icon {
  width: 12px;
  height: 12px;
  flex: 0 0 auto;
}

.video-script-card {
  min-height: 409px;
}

.agent-safe-badge {
  min-height: 25px;
  padding: 5px 10px;
  border: 1px solid #b9f8cf;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: #f0fdf4;
  color: #00a63e;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  white-space: nowrap;
}

.agent-safe-badge__icon {
  width: 12px;
  height: 12px;
}

.video-script-card__timeline {
  position: relative;
  margin-top: 20px;
  padding-left: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.video-script-card__timeline::before {
  content: '';
  position: absolute;
  left: 7px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  border-radius: 999px;
  background: linear-gradient(180deg, #06b6d4 0%, #d946ef 46%, #f97316 100%);
}

.script-step {
  position: relative;
  min-height: 82.5px;
  padding: 17px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  display: block;
  background: #f8fafc;
}

.script-step:nth-child(2) {
  min-height: 108.5px;
}

.script-step__meta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding-top: 0;
}

.script-step__meta > div {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.script-step__dot {
  position: absolute;
  z-index: 1;
  left: -17px;
  top: 8px;
  width: 12px;
  height: 12px;
  margin-top: 0;
  border: 2px solid #ffffff;
  border-radius: 999px;
}

.script-step--blue .script-step__dot {
  background: #06b6d4;
}

.script-step--violet .script-step__dot {
  background: #d946ef;
}

.script-step--orange .script-step__dot {
  background: #f97316;
}

.script-step__meta strong {
  display: inline;
  font-size: 12px;
  line-height: 16.5px;
  font-weight: 700;
}

.script-step--blue .script-step__meta strong {
  color: #06b6d4;
}

.script-step--violet .script-step__meta strong {
  color: #d946ef;
}

.script-step--orange .script-step__meta strong {
  color: #f97316;
}

.script-step__meta span:not(.script-step__dot) {
  display: inline;
  color: #90a1b9;
  font-size: 10px;
  line-height: 16.5px;
}

.script-step p {
  margin: 6px 0 0;
  min-height: 0;
  padding: 0;
  border: 0;
  border-radius: 0;
  display: block;
  background: transparent;
  color: #314158;
  font-size: 14px;
  line-height: 26px;
}

.storyboard-list {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.storyboard-list__item {
  min-width: 0;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.storyboard-list__item strong {
  color: #7c3aed;
  font-size: 11px;
  line-height: 16px;
}

.storyboard-list__item p {
  margin: 4px 0;
  color: #314158;
  font-size: 12px;
  line-height: 18px;
}

.storyboard-list__item span {
  color: #90a1b9;
  font-size: 11px;
  line-height: 16px;
}

.localization-card {
  min-height: 227px;
}

.localization-card__focus {
  margin-top: 12px;
  padding: 13px;
  border: 1px solid rgba(124, 58, 237, 0.2);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: linear-gradient(180deg, rgba(124, 58, 237, 0.05), rgba(6, 182, 212, 0.04));
}

.localization-card__focus span {
  color: #7c3aed;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
}

.localization-card__focus strong {
  color: #0a2463;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.localization-card p {
  margin: 12px 0 0;
  color: #45556c;
  font-size: 12px;
  line-height: 20px;
}

.localization-card__tags {
  margin-top: 10px;
}

.digital-human-card {
  min-height: 230px;
}

.agent-info-list {
  margin: 12px 0 0;
}

.agent-info-list div {
  min-height: 25px;
  border-bottom: 1px solid #f1f5f9;
  display: grid;
  grid-template-columns: 88px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
}

.agent-info-list div:last-child {
  border-bottom: 0;
}

.agent-info-list dt {
  color: #62748e;
  font-size: 11px;
  line-height: 17px;
}

.agent-info-list dd {
  margin: 0;
  color: #0a2463;
  font-size: 11px;
  line-height: 17px;
  font-weight: 700;
  text-align: right;
}

.launch-suggestion-card {
  min-height: 280px;
}

.launch-suggestion-card__group {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.launch-suggestion-card .agent-card__header + .launch-suggestion-card__group {
  margin-top: 12px;
}

.launch-suggestion-card__group > span,
.launch-suggestion-card__advice > span {
  color: #0891b2;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
}

.launch-suggestion-card__group--metrics > span {
  color: #fb2c36;
}

.agent-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.agent-chip {
  min-height: 22px;
  padding: 3px 9px;
  border-radius: 999px;
  font-size: 10px;
  line-height: 16px;
  font-weight: 700;
}

.agent-chip--platform {
  background: rgba(6, 182, 212, 0.1);
  color: #0891b2;
}

.agent-chip--violet {
  background: rgba(124, 58, 237, 0.1);
  color: #7c3aed;
}

.agent-chip--orange {
  background: rgba(249, 115, 22, 0.1);
  color: #ea580c;
}

.launch-suggestion-card__advice {
  margin-top: 10px;
  padding-top: 9px;
  border-top: 1px solid #f1f5f9;
}

.launch-suggestion-card__advice p {
  margin: 4px 0 0;
  color: #45556c;
  font-size: 12px;
  line-height: 20px;
}

.agent-reminder-card {
  margin-top: 24px;
  min-height: 379px;
  display: flex;
  flex-direction: column;
}

.agent-reminder-card__time {
  color: #90a1b9;
  font-size: 11px;
  line-height: 17px;
  white-space: nowrap;
}

.agent-reminder-card__body {
  margin-top: 16px;
  display: flex;
  gap: 12px;
  flex: 1 1 auto;
}

.agent-reminder-card__avatar {
  width: 36px;
  height: 36px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1), 0 2px 4px rgba(0, 0, 0, 0.1);
}

.agent-reminder-card__avatar svg {
  width: 16px;
  height: 16px;
}

.agent-reminder-card__content {
  min-width: 0;
  max-width: 768px;
}

.agent-gap-card {
  width: min(576px, 100%);
  margin: 0 0 0 8px;
  padding: 17px;
  border: 1px solid #fee685;
  border-radius: 16px;
  background: rgba(255, 251, 235, 0.65);
}

.agent-gap-card__text {
  display: flex;
  gap: 10px;
}

.agent-gap-card__icon {
  width: 28px;
  height: 28px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #ffb900 0%, #ff6900 100%);
  color: #ffffff;
}

.agent-gap-card__icon svg {
  width: 14px;
  height: 14px;
}

.agent-gap-card h3 {
  margin: 0;
  color: #7b3306;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.agent-gap-card p {
  margin: 2px 0 0;
  color: rgba(151, 60, 0, 0.9);
  font-size: 14px;
  line-height: 22px;
}

.agent-gap-card__actions {
  margin: 12px 0 0 36px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.agent-gap-card__button {
  min-height: 30px;
  padding: 7px 13px;
  border: 1px solid #fee685;
  border-radius: 12px;
  background: #ffffff;
  color: #973c00;
  font-family: inherit;
  font-size: 12px;
  line-height: 16px;
  cursor: pointer;
}

.agent-gap-card__button--primary {
  border-color: transparent;
  background: linear-gradient(90deg, #fe9a00 0%, #ff6900 100%);
  color: #ffffff;
  font-weight: 700;
}

.agent-gap-card__button--quiet {
  border-color: #e2e8f0;
  color: #45556c;
}

.agent-reminder-card__feedback {
  margin: 10px 0 0 8px;
  color: #0891b2;
  font-size: 12px;
  line-height: 18px;
}

.agent-reminder-input {
  min-height: 55px;
  margin-top: 16px;
  padding-top: 13px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  align-items: center;
  gap: 8px;
}

.agent-reminder-input__field {
  appearance: none;
  min-width: 0;
  height: 44px;
  padding: 10px 17px;
  border: 1px solid rgba(6, 182, 212, 0.28);
  border-radius: 16px;
  flex: 1 1 auto;
  overflow: hidden;
  background: linear-gradient(180deg, #ffffff 0%, rgba(240, 249, 255, 0.86) 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.92), 0 8px 18px -18px rgba(6, 182, 212, 0.6);
  color: #0a2463;
  font-family: inherit;
  font-size: 14px;
  line-height: 20px;
  white-space: nowrap;
  text-overflow: ellipsis;
  outline: none;
  cursor: text;
  transition: border-color 180ms ease, background 180ms ease, box-shadow 180ms ease;
}

.agent-reminder-input__field::placeholder {
  color: #64748b;
}

.agent-reminder-input__field:focus {
  border-color: rgba(124, 58, 237, 0.42);
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.08), 0 8px 20px -18px rgba(6, 182, 212, 0.7);
}

.agent-reminder-input__button {
  width: 80px;
  height: 40px;
  border: 0;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  flex: 0 0 auto;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
  font-family: inherit;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
  cursor: pointer;
}

.agent-reminder-input__icon {
  width: 14px;
  height: 14px;
}

.agent-optimize-card {
  position: relative;
  min-height: 151px;
  margin-top: 24px;
  padding: 24px;
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  overflow: hidden;
  background: linear-gradient(108deg, rgba(6, 182, 212, 0.18) 0%, #ffffff 50%, rgba(249, 115, 22, 0.18) 100%);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1);
}

.agent-optimize-card__glow {
  position: absolute;
  right: 16px;
  top: -80px;
  width: 288px;
  height: 288px;
  border-radius: 999px;
  background: rgba(124, 58, 237, 0.1);
  filter: blur(64px);
}

.agent-optimize-card__copy {
  position: relative;
  min-width: 0;
  flex: 1 1 auto;
}

.agent-optimize-card__copy > p {
  margin: 8px 0 0;
  color: #45556c;
  font-size: 14px;
  line-height: 20px;
}

.agent-optimize-card__prompts {
  margin-top: 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.agent-optimize-card__prompts button {
  min-height: 30px;
  padding: 7px 13px;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  background: #ffffff;
  color: #0a2463;
  font-family: inherit;
  font-size: 12px;
  line-height: 16px;
  cursor: pointer;
  transition: border-color 180ms ease, color 180ms ease, transform 180ms ease, background 180ms ease;
}

.agent-optimize-card__prompts button:hover,
.agent-optimize-card__prompts button.is-active {
  border-color: rgba(124, 58, 237, 0.32);
  background: rgba(255, 255, 255, 0.92);
  color: #7c3aed;
  transform: translateY(-1px);
}

.agent-optimize-card__button {
  position: relative;
  width: 172px;
  flex: 0 0 auto;
}

.agent-missing-card {
  margin-top: 24px;
  padding: 25px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.agent-missing-card p {
  margin: 0;
  color: #45556c;
  font-size: 14px;
  line-height: 22px;
}

.agent-missing-card .agent-button {
  width: 180px;
}

@media (min-width: 1041px) {
  .agent-result-header {
    min-height: 135.5px;
  }

  .agent-summary-card {
    min-height: 156px;
  }

  .agent-overview-card {
    min-height: 175.5px;
  }

  .compliance-result-card {
    min-height: 270.5px;
  }

  .video-script-card {
    min-height: 407.5px;
  }

  .localization-card {
    min-height: 227.25px;
  }

  .digital-human-card {
    min-height: 230px;
  }

  .launch-suggestion-card {
    min-height: 279.625px;
  }

  .agent-reminder-card {
    min-height: 378.75px;
  }

  .agent-optimize-card {
    min-height: 151px;
  }
}

@media (max-width: 1040px) {
  .agent-result-shell {
    width: min(1039px, calc(100% - 48px));
  }

  .agent-result-header {
    min-height: 0;
    flex-direction: column;
  }

  .agent-result-header__status {
    margin-left: 56px;
  }

  .agent-overview-card {
    align-items: stretch;
    flex-direction: column;
  }

  .agent-result-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 780px) {
  .agent-result-shell {
    width: calc(100% - 32px);
    padding: 32px 0;
  }

  .agent-result-header__main {
    gap: 12px;
  }

  .agent-result-header h1 {
    font-size: 30px;
    line-height: 38px;
  }

  .agent-result-header p {
    font-size: 14px;
    line-height: 22px;
  }

  .agent-summary-card__top {
    flex-direction: column;
  }

  .agent-summary-card__prompt,
  .agent-summary-card__tags {
    margin-left: 0;
  }

  .agent-button--ghost {
    width: 100%;
  }

  .agent-overview-card__items {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .storyboard-list {
    grid-template-columns: 1fr;
  }

  .agent-reminder-card__body {
    flex-direction: column;
  }

  .agent-reminder-card__content {
    max-width: none;
  }

  .agent-gap-card {
    margin-left: 0;
  }

  .agent-optimize-card {
    align-items: stretch;
    flex-direction: column;
  }

  .agent-optimize-card__button {
    width: 100%;
  }
}

@media (max-width: 560px) {
  .agent-result-page__ambient {
    display: none;
  }

  .agent-result-shell {
    padding: 24px 0;
  }

  .agent-result-header__main {
    flex-direction: column;
  }

  .agent-result-header__status {
    margin-left: 0;
    flex-wrap: wrap;
  }

  .agent-result-header h1 {
    font-size: 25px;
    line-height: 32px;
  }

  .agent-summary-card,
  .agent-overview-card,
  .compliance-result-card,
  .video-script-card,
  .localization-card,
  .digital-human-card,
  .launch-suggestion-card,
  .agent-reminder-card,
  .agent-optimize-card {
    padding: 18px;
  }

  .agent-section-title h2,
  .agent-section-title--large h2 {
    font-size: 17px;
    line-height: 25px;
  }

  .agent-summary-card__prompt {
    font-size: 14px;
    line-height: 23px;
  }

  .agent-overview-card__items {
    grid-template-columns: 1fr;
  }

  .agent-card__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .script-step {
    grid-template-columns: 1fr;
    gap: 6px;
  }

  .video-script-card__timeline::before {
    left: 5px;
  }

  .script-step__meta {
    padding-left: 0;
  }

  .agent-info-list div {
    grid-template-columns: 1fr;
    gap: 2px;
    padding: 6px 0;
  }

  .agent-info-list dd {
    text-align: left;
  }

  .agent-gap-card__text {
    flex-direction: column;
  }

  .agent-gap-card__actions {
    margin-left: 0;
  }
}
</style>
