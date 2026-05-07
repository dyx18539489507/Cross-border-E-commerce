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
              <div class="agent-card__header-actions">
                <span v-if="hasAppliedUpdate('compliance')" class="agent-applied-badge">已根据追问更新</span>
                <span class="agent-risk-badge" :class="riskBadgeClass">{{ result.overview.complianceRiskLevel }}</span>
              </div>
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
              <div class="agent-card__header-actions">
                <span v-if="hasAppliedUpdate('video')" class="agent-applied-badge">已根据追问更新</span>
                <span class="agent-safe-badge">
                  <CircleCheck class="agent-safe-badge__icon" aria-hidden="true" />
                  <span>已规避高风险功效表达</span>
                </span>
              </div>
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
              <span v-if="hasAppliedUpdate('localization')" class="agent-applied-badge">已根据追问更新</span>
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
              <span v-if="hasAppliedUpdate('digitalHuman')" class="agent-applied-badge">已根据追问更新</span>
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
              <span v-if="hasAppliedUpdate('promotion')" class="agent-applied-badge">已根据追问更新</span>
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

        <div ref="followUpChatRef" class="follow-up-chat" aria-live="polite">
          <article class="follow-up-message follow-up-message--agent follow-up-message--initial">
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
          </article>

          <article
            v-for="message in messages"
            :key="message.id"
            class="follow-up-message"
            :class="[`follow-up-message--${message.role}`, `follow-up-message--${message.type}`]"
          >
            <template v-if="message.role === 'user'">
              <div class="follow-up-user-bubble">{{ message.content }}</div>
            </template>

            <template v-else>
              <span class="agent-reminder-card__avatar follow-up-message__avatar" aria-hidden="true">
                <Cpu />
              </span>
              <div class="follow-up-agent-panel">
                <div v-if="message.type === 'thinking'" class="agent-thinking-card">
                  <div class="agent-thinking-card__header">
                    <span class="agent-thinking-card__pulse" aria-hidden="true"></span>
                    <strong>丝路 Agent 正在增量分析</strong>
                  </div>
                  <ul class="agent-thinking-card__steps">
                    <li
                      v-for="(step, index) in message.visibleThinkingSteps"
                      :key="step"
                      :class="{ 'is-active': index === (message.visibleThinkingSteps?.length || 0) - 1 }"
                    >
                      <span>{{ step }}</span>
                    </li>
                  </ul>
                </div>

                <div v-else-if="message.type === 'summary'" class="agent-follow-up-card">
                  <div class="agent-follow-up-card__bar" aria-hidden="true"></div>
                  <div class="agent-follow-up-card__header">
                    <div class="agent-follow-up-card__title">
                      <Cpu aria-hidden="true" />
                      <strong>Agent 补充建议</strong>
                      <span>| 根据追问生成</span>
                    </div>
                    <span v-if="message.applied" class="agent-follow-up-card__sync">
                      <CircleCheck aria-hidden="true" />
                      已同步到当前方案
                    </span>
                  </div>
                  <p class="agent-follow-up-card__summary">{{ message.summary }}</p>
                  <div class="agent-follow-up-card__tags">
                    <span v-for="tag in message.tags" :key="tag">{{ tag }}</span>
                  </div>
                  <div class="agent-follow-up-details">
                    <article>
                      <strong>合规变化</strong>
                      <p>{{ message.details?.compliance }}</p>
                    </article>
                    <article>
                      <strong>内容风格变化</strong>
                      <p>{{ message.details?.contentStyle }}</p>
                    </article>
                    <article>
                      <strong>视频表达建议</strong>
                      <p>{{ message.details?.videoExpression }}</p>
                    </article>
                    <article>
                      <strong>投放建议</strong>
                      <p>{{ message.details?.promotion }}</p>
                    </article>
                  </div>
                  <div class="agent-follow-up-card__actions">
                    <button
                      type="button"
                      class="agent-follow-up-card__button agent-follow-up-card__button--primary"
                      :disabled="message.applied"
                      @click="applyFollowUp(message)"
                    >
                      {{ message.applied ? '已应用' : '应用到当前方案' }}
                    </button>
                  </div>
                </div>

                <div v-else-if="message.type === 'error'" class="agent-follow-up-error">
                  <WarningFilled aria-hidden="true" />
                  <div>
                    <strong>Agent 暂时无法继续分析，请稍后重试。</strong>
                    <button type="button" @click="regenerateFollowUp(message)">重新生成</button>
                  </div>
                </div>
              </div>
            </template>
          </article>
        </div>

        <div class="agent-reminder-input" aria-label="继续与丝路 Agent 对话">
          <input
            v-model="agentMessage"
            class="agent-reminder-input__field"
            type="text"
            aria-label="继续告诉丝路 Agent"
            placeholder="输入补充要求，例如换成印尼市场、语气更年轻一点……"
            :disabled="isFollowUpLoading"
            @keyup.enter="sendAgentMessage"
          />
          <button
            type="button"
            class="agent-reminder-input__button"
            :disabled="isFollowUpLoading || !agentMessage.trim()"
            @click="sendAgentMessage"
          >
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
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
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

type FollowUpMessage = {
  id: string
  role: 'user' | 'agent'
  type: 'text' | 'thinking' | 'summary' | 'error'
  content: string
  thinkingSteps?: string[]
  visibleThinkingSteps?: string[]
  summary?: string
  tags?: string[]
  detailExpanded?: boolean
  applied?: boolean
  details?: FollowUpDetails
}

type FollowUpDetails = {
  compliance: string
  contentStyle: string
  videoExpression: string
  promotion: string
}

type FollowUpResult = {
  summary: string
  affectedModules: string[]
  details: FollowUpDetails
}

type FollowUpContext = {
  productName: string
  targetMarket: string
  platform: string
  audience: string
  sellingPoints: string
  complianceResult: string
  contentStrategy: string
  digitalHumanPlan: string
  promotionAdvice: string
}

const router = useRouter()
const reminderAction = ref('')
const agentMessage = ref('')
const storedResult = ref<AgentResult | null>(readAgentResult())
const messages = ref<FollowUpMessage[]>([])
const isFollowUpLoading = ref(false)
const appliedUpdateModules = ref<string[]>([])
const followUpChatRef = ref<HTMLElement | null>(null)

const thinkingSteps = [
  '正在读取当前商品信息……',
  '正在结合原方案判断影响模块……',
  '正在重新评估目标市场与内容风格……',
  '正在检查合规表达边界……',
  '正在整理可执行优化建议……'
]

const thinkingTimers = new Map<string, number>()
let followUpAbortController: AbortController | null = null

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
  if (!message || isFollowUpLoading.value) return

  reminderAction.value = `“${message}”`
  agentMessage.value = ''
  appendFollowUp(message)
}

const appendFollowUp = (question: string) => {
  const userMessage: FollowUpMessage = {
    id: createMessageId('user'),
    role: 'user',
    type: 'text',
    content: question
  }
  const agentReply: FollowUpMessage = {
    id: createMessageId('agent'),
    role: 'agent',
    type: 'thinking',
    content: question,
    thinkingSteps: [...thinkingSteps],
    visibleThinkingSteps: []
  }

  messages.value.push(userMessage, agentReply)
  startThinkingSteps(agentReply.id)
  scrollFollowUpChat()
  void requestFollowUp(question, agentReply.id)
}

const requestFollowUp = async (question: string, agentMessageId: string) => {
  isFollowUpLoading.value = true
  followUpAbortController = new AbortController()
  let receivedResult = false
  let streamError = ''

  try {
    const response = await fetch('/api/v1/agent/follow-up', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        question,
        context: buildFollowUpContext()
      }),
      signal: followUpAbortController.signal
    })

    if (!response.ok || !response.body) {
      throw new Error('follow-up stream unavailable')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      const blocks = buffer.split(/\n\n/)
      buffer = blocks.pop() || ''
      for (const block of blocks) {
        const event = parseSSEBlock(block)
        if (!event) continue
        if (event.event === 'result') {
          updateAgentSummary(agentMessageId, normalizeFollowUpResult(event.data))
          receivedResult = true
        } else if (event.event === 'error') {
          streamError = event.data?.message || 'follow-up stream error'
        }
      }
    }

    if (buffer.trim()) {
      const event = parseSSEBlock(buffer)
      if (event?.event === 'result') {
        updateAgentSummary(agentMessageId, normalizeFollowUpResult(event.data))
        receivedResult = true
      } else if (event?.event === 'error') {
        streamError = event.data?.message || 'follow-up stream error'
      }
    }

    if (streamError || !receivedResult) {
      throw new Error(streamError || 'follow-up result missing')
    }
  } catch (error) {
    if ((error as Error).name !== 'AbortError') {
      updateAgentError(agentMessageId)
    }
  } finally {
    stopThinkingSteps(agentMessageId)
    isFollowUpLoading.value = false
    followUpAbortController = null
    scrollFollowUpChat()
  }
}

const parseSSEBlock = (block: string) => {
  const lines = block.split(/\r?\n/)
  let event = 'message'
  const dataLines: string[] = []

  lines.forEach((line) => {
    if (line.startsWith('event:')) {
      event = line.slice(6).trim()
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trim())
    }
  })

  if (!dataLines.length) return null
  try {
    return { event, data: JSON.parse(dataLines.join('\n')) }
  } catch {
    return { event, data: {} }
  }
}

const startThinkingSteps = (messageId: string) => {
  const message = findMessage(messageId)
  if (!message) return
  message.visibleThinkingSteps = [thinkingSteps[0]]

  let index = 1
  const timer = window.setInterval(() => {
    const target = findMessage(messageId)
    if (!target || target.type !== 'thinking') {
      stopThinkingSteps(messageId)
      return
    }
    if (index < thinkingSteps.length) {
      target.visibleThinkingSteps = thinkingSteps.slice(0, index + 1)
      index += 1
      scrollFollowUpChat()
    }
  }, 760)
  thinkingTimers.set(messageId, timer)
}

const stopThinkingSteps = (messageId: string) => {
  const timer = thinkingTimers.get(messageId)
  if (timer) {
    window.clearInterval(timer)
    thinkingTimers.delete(messageId)
  }
}

const finishThinkingSteps = (messageId: string) => {
  const message = findMessage(messageId)
  if (message?.type === 'thinking') {
    message.visibleThinkingSteps = [...thinkingSteps]
  }
  stopThinkingSteps(messageId)
}

const updateAgentSummary = (messageId: string, payload: FollowUpResult) => {
  finishThinkingSteps(messageId)
  const message = findMessage(messageId)
  if (!message) return

  message.type = 'summary'
  message.summary = payload.summary
  message.tags = payload.affectedModules
  message.details = payload.details
  message.detailExpanded = true
  message.applied = false
  scrollFollowUpChat()
}

const updateAgentError = (messageId: string) => {
  finishThinkingSteps(messageId)
  const message = findMessage(messageId)
  if (!message) return
  message.type = 'error'
  message.summary = ''
  message.tags = []
  message.detailExpanded = false
  message.applied = false
}

const normalizeFollowUpResult = (data: any): FollowUpResult => {
  const details = data?.details || {}
  const modules = Array.isArray(data?.affectedModules)
    ? data.affectedModules.filter((item: unknown): item is string => typeof item === 'string' && item.trim() !== '')
    : []

  return {
    summary: sanitizeDisplayText(data?.summary, '已基于当前商品和原方案，整理出适合继续优化的补充建议。'),
    affectedModules: modules.length ? modules.slice(0, 6) : ['市场策略', '内容风格', '投放建议', '合规风险'],
    details: {
      compliance: sanitizeDisplayText(details.compliance, '继续避免绝对化、医疗化和未经证实的认证表达，并以目标市场实际准入材料为准。'),
      contentStyle: sanitizeDisplayText(details.contentStyle, '内容语气应贴近目标人群日常表达，保留真实体验感，减少夸张承诺。'),
      videoExpression: sanitizeDisplayText(details.videoExpression, '前 3 秒突出核心场景和可见卖点，中段用真实使用画面承接。'),
      promotion: sanitizeDisplayText(details.promotion, '优先测试当前主平台短视频内容，结合完播率、点击率和评论问题继续迭代素材。')
    }
  }
}

const sanitizeDisplayText = (value: unknown, fallback: string) => {
  if (typeof value !== 'string') return fallback
  const cleaned = value.replace(/reasoning_content|思维链|chain-of-thought|内部推理/gi, '').replace(/\s+/g, ' ').trim()
  return cleaned || fallback
}

const buildFollowUpContext = (): FollowUpContext => {
  const current = result.value
  return {
    productName: current.recognizedInfo.productName,
    targetMarket: current.recognizedInfo.targetMarket,
    platform: current.recognizedInfo.targetPlatform,
    audience: current.recognizedInfo.targetAudience,
    sellingPoints: current.recognizedInfo.coreSellingPoints.join(' / '),
    complianceResult: `${current.overview.complianceRiskLevel}；${current.compliance.summary}`,
    contentStrategy: [
      current.localization.direction,
      current.localization.reason,
      current.script.opening.content,
      current.script.middle.content,
      current.script.ending.content
    ].filter(Boolean).join('；'),
    digitalHumanPlan: [
      current.digitalHuman.persona,
      current.digitalHuman.tone,
      current.digitalHuman.visualStyle,
      current.digitalHuman.shootingStyle
    ].filter(Boolean).join('；'),
    promotionAdvice: [
      current.promotion.platforms.join(' / '),
      current.promotion.contentTags.join(' / '),
      current.promotion.optimizationAdvice
    ].filter(Boolean).join('；')
  }
}

const applyFollowUp = (message: FollowUpMessage) => {
  if (message.type !== 'summary' || message.applied) return
  message.applied = true
  const nextModules = new Set(appliedUpdateModules.value)
  ;(message.tags || []).forEach((tag) => nextModules.add(tag))
  appliedUpdateModules.value = Array.from(nextModules)
}

const regenerateFollowUp = (message: FollowUpMessage) => {
  if (isFollowUpLoading.value || !message.content) return
  message.type = 'thinking'
  message.thinkingSteps = [...thinkingSteps]
  message.visibleThinkingSteps = []
  message.summary = ''
  message.tags = []
  message.details = undefined
  message.detailExpanded = false
  message.applied = false
  startThinkingSteps(message.id)
  void requestFollowUp(message.content, message.id)
}

type AppliedModuleKey = 'compliance' | 'video' | 'localization' | 'digitalHuman' | 'promotion'

const hasAppliedUpdate = (module: AppliedModuleKey) => {
  const modules = appliedUpdateModules.value.join(' ')
  const rules: Record<AppliedModuleKey, string[]> = {
    compliance: ['合规', '风险'],
    video: ['视频', '表达', '内容风格'],
    localization: ['市场', '内容风格', '本地化'],
    digitalHuman: ['数字人'],
    promotion: ['投放', '平台']
  }
  return rules[module].some((keyword) => modules.includes(keyword))
}

const findMessage = (messageId: string) => messages.value.find((message) => message.id === messageId)

const createMessageId = (prefix: string) => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `${prefix}-${crypto.randomUUID()}`
  }
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

const scrollFollowUpChat = () => {
  void nextTick(() => {
    const el = followUpChatRef.value
    if (el) {
      el.scrollTop = el.scrollHeight
    }
  })
}

onBeforeUnmount(() => {
  followUpAbortController?.abort()
  thinkingTimers.forEach((timer) => window.clearInterval(timer))
  thinkingTimers.clear()
})
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

.agent-card__header-actions {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex: 0 0 auto;
}

.agent-applied-badge {
  min-height: 24px;
  padding: 4px 9px;
  border: 1px solid rgba(124, 58, 237, 0.24);
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.09), rgba(124, 58, 237, 0.1));
  color: #7c3aed;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  white-space: nowrap;
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

.follow-up-chat {
  min-height: 214px;
  max-height: 560px;
  margin-top: 16px;
  padding: 2px 2px 4px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(124, 58, 237, 0.24) transparent;
}

.follow-up-message {
  min-width: 0;
  display: flex;
  gap: 12px;
}

.follow-up-message--initial {
  align-items: flex-start;
}

.follow-up-message--user {
  justify-content: flex-end;
}

.follow-up-user-bubble {
  max-width: min(620px, 82%);
  padding: 11px 15px;
  border-radius: 18px 18px 6px 18px;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
  font-size: 14px;
  line-height: 22px;
  font-weight: 700;
  box-shadow: 0 18px 28px -22px rgba(124, 58, 237, 0.7);
  overflow-wrap: anywhere;
}

.follow-up-message__avatar {
  margin-top: 2px;
}

.follow-up-agent-panel {
  min-width: 0;
  width: min(768px, calc(100% - 48px));
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

.agent-thinking-card {
  width: min(520px, 100%);
  margin-left: 8px;
  padding: 15px 16px;
  border: 1px solid rgba(124, 58, 237, 0.18);
  border-radius: 16px;
  background: linear-gradient(135deg, rgba(248, 250, 252, 0.92) 0%, rgba(255, 255, 255, 0.96) 48%, rgba(124, 58, 237, 0.08) 100%);
  box-shadow: 0 12px 28px -26px rgba(15, 23, 42, 0.38);
}

.agent-thinking-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #0a2463;
  font-size: 13px;
  line-height: 18px;
}

.agent-thinking-card__pulse {
  position: relative;
  width: 9px;
  height: 9px;
  border-radius: 999px;
  background: #7c3aed;
  box-shadow: 0 0 0 0 rgba(124, 58, 237, 0.32);
  animation: agent-thinking-pulse 1.2s ease-in-out infinite;
}

.agent-thinking-card__steps {
  margin: 12px 0 0;
  padding: 0;
  display: grid;
  gap: 7px;
  list-style: none;
}

.agent-thinking-card__steps li {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
}

.agent-thinking-card__steps li::before {
  content: '';
  width: 5px;
  height: 5px;
  border-radius: 999px;
  flex: 0 0 auto;
  background: rgba(6, 182, 212, 0.52);
}

.agent-thinking-card__steps li.is-active span {
  background: linear-gradient(90deg, #0891b2 0%, #7c3aed 100%);
  background-clip: text;
  color: transparent;
  font-weight: 700;
}

.agent-thinking-card__steps li.is-active span::after {
  content: '';
  display: inline-block;
  width: 6px;
  height: 14px;
  margin-left: 3px;
  border-radius: 999px;
  background: #7c3aed;
  vertical-align: -2px;
  animation: agent-cursor-blink 0.86s steps(2, start) infinite;
}

.agent-follow-up-card {
  position: relative;
  width: min(768px, 100%);
  margin-left: 8px;
  padding: 16px;
  border: 1px solid rgba(124, 58, 237, 0.18);
  border-radius: 16px;
  overflow: hidden;
  background: linear-gradient(145deg, #ffffff 0%, #ffffff 48%, rgba(124, 58, 237, 0.09) 100%);
  box-shadow: 0 18px 34px -28px rgba(15, 23, 42, 0.5);
}

.agent-follow-up-card__bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 54%, #f97316 100%);
}

.agent-follow-up-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.agent-follow-up-card__title {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.agent-follow-up-card__title svg {
  width: 14px;
  height: 14px;
  color: #7c3aed;
  flex: 0 0 auto;
}

.agent-follow-up-card__title strong {
  color: #0a2463;
  font-size: 14px;
  line-height: 20px;
}

.agent-follow-up-card__title span {
  color: #90a1b9;
  font-size: 10px;
  line-height: 15px;
  white-space: nowrap;
}

.agent-follow-up-card__sync {
  min-height: 24px;
  padding: 4px 9px;
  border: 1px solid #b9f8cf;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: #f0fdf4;
  color: #008236;
  font-size: 10px;
  line-height: 15px;
  font-weight: 700;
  white-space: nowrap;
}

.agent-follow-up-card__sync svg {
  width: 12px;
  height: 12px;
}

.agent-follow-up-card__summary {
  margin: 13px 0 0;
  color: #314158;
  font-size: 14px;
  line-height: 23px;
}

.agent-follow-up-card__tags {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.agent-follow-up-card__tags span {
  min-height: 23px;
  padding: 3px 9px;
  border: 1px solid rgba(124, 58, 237, 0.25);
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1), rgba(124, 58, 237, 0.1));
  color: #0a2463;
  font-size: 11px;
  line-height: 17px;
  font-weight: 700;
}

.agent-follow-up-details {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.agent-follow-up-details article {
  min-width: 0;
  padding: 11px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: rgba(248, 250, 252, 0.74);
}

.agent-follow-up-details strong {
  display: block;
  color: #7c3aed;
  font-size: 10px;
  line-height: 15px;
  letter-spacing: 0.5px;
}

.agent-follow-up-details p {
  margin: 4px 0 0;
  color: #45556c;
  font-size: 12px;
  line-height: 19px;
}

.agent-follow-up-card__actions {
  margin-top: 13px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.agent-follow-up-card__button {
  min-height: 30px;
  padding: 7px 13px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  color: #0a2463;
  font-family: inherit;
  font-size: 12px;
  line-height: 16px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 180ms ease, border-color 180ms ease, opacity 180ms ease;
}

.agent-follow-up-card__button:hover:not(:disabled) {
  border-color: rgba(124, 58, 237, 0.34);
  transform: translateY(-1px);
}

.agent-follow-up-card__button--primary {
  border-color: transparent;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
}

.agent-follow-up-card__button:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.agent-follow-up-error {
  width: min(520px, 100%);
  margin-left: 8px;
  padding: 15px;
  border: 1px solid rgba(251, 44, 54, 0.18);
  border-radius: 16px;
  display: flex;
  gap: 10px;
  background: rgba(254, 242, 242, 0.78);
  color: #991b1b;
}

.agent-follow-up-error > svg {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
}

.agent-follow-up-error strong {
  display: block;
  font-size: 13px;
  line-height: 20px;
}

.agent-follow-up-error button {
  margin-top: 8px;
  min-height: 30px;
  padding: 7px 13px;
  border: 0;
  border-radius: 12px;
  background: #ffffff;
  color: #0a2463;
  font-family: inherit;
  font-size: 12px;
  line-height: 16px;
  font-weight: 700;
  cursor: pointer;
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
  transition: opacity 180ms ease, transform 180ms ease;
}

.agent-reminder-input__button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.agent-reminder-input__button:disabled,
.agent-reminder-input__field:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.agent-reminder-input__icon {
  width: 14px;
  height: 14px;
}

@keyframes agent-thinking-pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(124, 58, 237, 0.32);
  }
  70% {
    box-shadow: 0 0 0 8px rgba(124, 58, 237, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(124, 58, 237, 0);
  }
}

@keyframes agent-cursor-blink {
  0%,
  45% {
    opacity: 1;
  }
  46%,
  100% {
    opacity: 0;
  }
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

  .follow-up-message--agent {
    align-items: flex-start;
  }

  .follow-up-agent-panel,
  .agent-follow-up-card,
  .agent-thinking-card,
  .agent-follow-up-error {
    width: 100%;
    margin-left: 0;
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

  .agent-card__header-actions,
  .agent-follow-up-card__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .follow-up-chat {
    max-height: none;
  }

  .follow-up-user-bubble {
    max-width: 92%;
  }

  .agent-follow-up-details {
    grid-template-columns: 1fr;
  }

  .agent-reminder-input {
    align-items: stretch;
    flex-direction: column;
  }

  .agent-reminder-input__button {
    width: 100%;
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
