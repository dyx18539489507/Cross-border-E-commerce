<!--
/**
 * 模块说明：丝路 Agent 结果页与追问编排页。
 * 业务场景：展示一次 Agent 分析的合规、本地化、脚本、数字人和投放结论，并允许用户继续追问。
 * 核心职责：读取会话中的 AgentResult，处理追问 SSE 返回，把可应用的增量更新同步回当前方案和商品录入草稿。
 */
-->
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
        <p>{{ missingResultMessage }}</p>
        <button v-if="!isHistoryRestoring" type="button" class="agent-button agent-button--primary" @click="goHome">
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

      <section class="agent-card agent-next-card" aria-label="沉淀 Agent 商品信息">
        <div>
          <strong>将本次分析沉淀为营销项目</strong>
          <p>Agent 已识别的商品、市场、合规边界、脚本、数字人和投放建议会自动带入工作台。</p>
        </div>
        <div class="agent-next-card__actions">
          <button type="button" class="agent-button agent-button--primary" :disabled="isCreatingProject" @click="createMarketingProject">
            <span>{{ isCreatingProject ? '创建中…' : '一键创建营销项目' }}</span>
            <ArrowRight class="agent-button__icon" aria-hidden="true" />
          </button>
          <button type="button" class="agent-button agent-button--ghost" @click="continueProductEntryFromAgent">
            <span>继续完善商品信息</span>
          </button>
        </div>
      </section>

      <section v-if="workflowResult || matchedComplianceRules.length" class="agent-card agent-workflow-card" aria-label="多 Agent 工作流结果">
        <div class="agent-card__header">
          <div class="agent-section-title">
            <span class="agent-title-icon agent-title-icon--violet">
              <Cpu aria-hidden="true" />
            </span>
            <h2>多 Agent 执行链</h2>
          </div>
          <div class="agent-card__header-actions">
            <span v-if="workflowResult?.revised" class="agent-applied-badge">已根据 Critic 二次修订</span>
            <span v-if="workflowResult?.workflow_status" class="agent-safe-badge">
              <CircleCheck class="agent-safe-badge__icon" aria-hidden="true" />
              <span>{{ workflowStatusLabel }}</span>
            </span>
          </div>
        </div>

        <div v-if="workflowTraces.length" class="agent-trace-list">
          <article v-for="trace in workflowTraces" :key="`${trace.agent_name}-${trace.stage}`" class="agent-trace-item">
            <div class="agent-trace-item__head">
              <strong>{{ getAgentDisplayName(trace.agent_name) }}</strong>
              <span :class="`agent-trace-item__status agent-trace-item__status--${trace.status}`">{{ getTraceStatusLabel(trace.status) }}</span>
            </div>
            <p>{{ getTraceOutputSummary(trace) }}</p>
            <small>{{ trace.stage }} · {{ trace.duration_ms || 0 }}ms</small>
          </article>
        </div>

        <div v-if="criticScoreItems.length" class="agent-critic-panel">
          <div class="agent-critic-panel__scores">
            <article v-for="item in criticScoreItems" :key="item.label">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}/5</strong>
            </article>
          </div>
          <div class="agent-critic-panel__feedback">
            <strong>Critic 评审</strong>
            <ul>
              <li v-for="problem in criticProblems" :key="problem">{{ problem }}</li>
            </ul>
          </div>
        </div>

        <div v-if="matchedComplianceRules.length" class="agent-rule-panel">
          <div class="agent-rule-panel__head">
            <strong>命中的合规知识</strong>
            <span>{{ matchedComplianceRules.length }} 条</span>
          </div>
          <article v-for="rule in matchedComplianceRules" :key="rule.id" class="agent-rule-item">
            <div>
              <strong>{{ rule.id }}</strong>
              <span>{{ rule.country }} · {{ rule.platform }} · {{ rule.category }}</span>
            </div>
            <p>{{ rule.rule_text }}</p>
          </article>
          <p v-if="result.compliance.disclaimer" class="agent-rule-panel__disclaimer">{{ result.compliance.disclaimer }}</p>
        </div>
      </section>

      <section class="agent-card agent-shortcut-card" aria-label="Agent 工作台快捷入口">
        <button type="button" class="agent-shortcut-button" @click="goCompliance">
          <DocumentChecked aria-hidden="true" />
          <span>进入合规分析</span>
        </button>
        <button type="button" class="agent-shortcut-button" @click="goScriptWorkspace">
          <VideoCamera aria-hidden="true" />
          <span>进入脚本工作台</span>
        </button>
        <button type="button" class="agent-shortcut-button" @click="startContentWorkflow('combo')">
          <MagicStick aria-hidden="true" />
          <span>进入内容生成</span>
        </button>
        <button type="button" class="agent-shortcut-button" @click="goTimelineWorkspace">
          <TrendCharts aria-hidden="true" />
          <span>进入剪辑时间线</span>
        </button>
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

            <div class="agent-card-action-row">
              <button type="button" class="agent-card-action-button agent-card-action-button--primary" @click="startContentWorkflow('video')">
                <span>生成短视频素材</span>
                <ArrowRight class="agent-card-action-button__icon" aria-hidden="true" />
              </button>
              <button type="button" class="agent-card-action-button" @click="startContentWorkflow('combo')">
                <span>进入组合创作</span>
              </button>
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

            <div class="agent-card-action-row agent-card-action-row--compact">
              <button type="button" class="agent-card-action-button agent-card-action-button--pink" @click="startContentWorkflow('avatar')">
                <span>制作数字人口播</span>
                <ArrowRight class="agent-card-action-button__icon" aria-hidden="true" />
              </button>
            </div>
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

            <div class="agent-card-action-row agent-card-action-row--compact">
              <button type="button" class="agent-card-action-button agent-card-action-button--green" @click="startContentWorkflow('promotion')">
                <span>生成投放素材</span>
                <ArrowRight class="agent-card-action-button__icon" aria-hidden="true" />
              </button>
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
                    <strong>{{ message.regenerating ? '丝路 Agent 正在重新编排当前方案' : '丝路 Agent 正在增量分析' }}</strong>
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
                    <span v-if="message.intent">{{ getFollowUpIntentLabel(message.intent) }}</span>
                    <span v-for="tag in message.tags" :key="tag">{{ tag }}</span>
                  </div>
                  <div v-if="message.missingFields?.length" class="agent-follow-up-missing">
                    <strong>还需要补充</strong>
                    <span v-for="field in message.missingFields" :key="field">{{ field }}</span>
                  </div>
                  <div v-if="getFollowUpDetailCards(message).length" class="agent-follow-up-details">
                    <article v-for="detail in getFollowUpDetailCards(message)" :key="detail.key">
                      <strong>{{ detail.title }}</strong>
                      <p>{{ detail.value }}</p>
                    </article>
                  </div>
                  <div class="agent-follow-up-card__actions">
                    <button
                      type="button"
                      class="agent-follow-up-card__button agent-follow-up-card__button--primary"
                      :disabled="message.applied || isPlanRegenerating"
                      @click="applyFollowUp(message)"
                    >
                      {{ getFollowUpActionLabel(message) }}
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
            :disabled="isFollowUpLoading || isPlanRegenerating"
            @keyup.enter="sendAgentMessage"
          />
          <button
            type="button"
            class="agent-reminder-input__button"
            :disabled="isFollowUpLoading || isPlanRegenerating || !agentMessage.trim()"
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
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
import { ElMessage } from 'element-plus'
import type { AgentInput, AgentResult, AgentTrace, ComplianceRule, WorkflowResult } from '@/types/agent'
import {
  readAgentInput,
  readAgentResult,
  readAgentWorkflowResult,
  saveAgentInput,
  saveAgentResult,
  saveAgentWorkflowResult
} from '@/utils/agentStorage'
import {
  getProductDraft,
  saveAgentResultAsProductDraft,
  saveProductDraft
} from '@/utils/productEntryDraft'
import {
  buildContentWorkflowIntentFromAgent,
  saveContentWorkflowIntent,
  type ContentWorkflowTarget
} from '@/utils/contentWorkflowIntent'
import { agentAPI } from '@/api/agent'

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
  intent?: string
  regenerating?: boolean
  thinkingSteps?: string[]
  visibleThinkingSteps?: string[]
  summary?: string
  tags?: string[]
  detailExpanded?: boolean
  applied?: boolean
  details?: FollowUpDetails
  cards?: FollowUpDynamicCard[]
  updatedFields?: FollowUpUpdatedFields
  missingFields?: string[]
}

type FollowUpDetails = {
  compliance: string
  contentStyle: string
  videoExpression: string
  promotion: string
}

type FollowUpDetailKey = keyof FollowUpDetails

type FollowUpDetailDefinition = {
  key: FollowUpDetailKey
  title: string
  moduleKeywords: string[]
  queryKeywords: string[]
}

type FollowUpDetailCard = {
  key: string
  title: string
  value: string
  type?: string
}

type FollowUpDynamicCard = {
  key: string
  type: string
  title: string
  value: string
}

type FollowUpFieldValue = string | string[]
type FollowUpUpdatedFields = Partial<Record<
  | 'productName'
  | 'category'
  | 'targetMarket'
  | 'targetPlatform'
  | 'targetAudience'
  | 'coreSellingPoints'
  | 'materialSpec'
  | 'usageScenario'
  | 'marketingGoal'
  | 'budgetPreference'
  | 'complianceHints'
  | 'localizationHints'
  | 'description',
  FollowUpFieldValue
>>

type FollowUpResult = {
  summary: string
  intent: string
  affectedModules: string[]
  details: FollowUpDetails
  cards: FollowUpDynamicCard[]
  updatedFields: FollowUpUpdatedFields
  missingFields: string[]
}

type FollowUpContext = {
  productName: string
  category: string
  targetMarket: string
  platform: string
  audience: string
  sellingPoints: string
  materialSpec: string
  usageScenario: string
  imageUnderstanding: string
  rawPrompt: string
  complianceResult: string
  contentStrategy: string
  digitalHumanPlan: string
  promotionAdvice: string
}

const router = useRouter()
const route = useRoute()
const reminderAction = ref('')
const agentMessage = ref('')
const storedResult = ref<AgentResult | null>(readAgentResult())
const storedWorkflow = ref<WorkflowResult | null>(readAgentWorkflowResult())
const storedAgentInput = ref<AgentInput | null>(readAgentInput())
const isHistoryRestoring = ref(false)
const messages = ref<FollowUpMessage[]>([])
const isFollowUpLoading = ref(false)
const isPlanRegenerating = ref(false)
const isCreatingProject = ref(false)
const appliedUpdateModules = ref<string[]>([])
const followUpChatRef = ref<HTMLElement | null>(null)

const thinkingSteps = [
  '正在读取当前商品信息……',
  '正在结合原方案判断影响模块……',
  '正在重新评估目标市场与内容风格……',
  '正在检查合规表达边界……',
  '正在整理可执行优化建议……'
]

const regenerationSteps = [
  '正在合并追问与当前商品资料……',
  '正在重新识别商品、市场与平台……',
  '正在重算合规边界和本地化方向……',
  '正在生成新的脚本、数字人与投放建议……',
  '正在同步到当前方案和商品资料……'
]

const followUpDetailDefinitions: FollowUpDetailDefinition[] = [
  {
    key: 'compliance',
    title: '合规变化',
    moduleKeywords: ['合规', '风险', '认证', '准入', '法规', '敏感', '材料', '成分'],
    queryKeywords: ['合规', '风险', '认证', '准入', '禁', '材料', '成分', '市场', '国家', '印尼', '印度尼西亚', '马来', '美国', '欧洲', '东南亚', '食品', '美妆', '母婴', '儿童', '医疗', '电子', '产品', '商品']
  },
  {
    key: 'contentStyle',
    title: '市场与内容调整',
    moduleKeywords: ['市场', '内容', '风格', '本地化', '人群', '语气'],
    queryKeywords: ['市场', '国家', '印尼', '印度尼西亚', '马来', '美国', '欧洲', '东南亚', '语气', '年轻', '本地化', '人群', '风格', '内容', '产品', '商品']
  },
  {
    key: 'videoExpression',
    title: '视频表达建议',
    moduleKeywords: ['视频', '表达', '脚本', '分镜', '镜头', '数字人', '口播'],
    queryKeywords: ['视频', '脚本', '分镜', '镜头', '口播', '数字人', '画面', '开头', '字幕', '图片', '素材', '产品', '商品']
  },
  {
    key: 'promotion',
    title: '投放建议',
    moduleKeywords: ['投放', '平台', '渠道', '指标', '预算', '推广'],
    queryKeywords: ['投放', '平台', 'TikTok', 'Amazon', 'Shopee', 'Lazada', 'Temu', '预算', '点击', '转化', '完播', '推广', '市场']
  }
]

const followUpIntentLabels: Record<string, string> = {
  change_product: '意图：更换商品',
  change_market: '意图：调整市场',
  change_platform: '意图：调整平台',
  change_audience: '意图：调整人群',
  add_product_info: '意图：补充商品信息',
  add_material_info: '意图：补充材质',
  add_usage_scenario: '意图：补充场景',
  add_image_info: '意图：补充图片信息',
  adjust_content_tone: '意图：调整语气',
  optimize_script: '意图：优化脚本',
  optimize_compliance: '意图：优化合规',
  optimize_promotion: '意图：优化投放',
  ask_clarification: '意图：需要补充',
  general_question: '意图：综合补充'
}

const followUpTypeTitles: Record<string, string> = {
  product: '商品资料',
  market: '市场策略',
  platform: '平台适配',
  audience: '人群定位',
  selling_point: '卖点调整',
  material: '材质/成分',
  scenario: '使用场景',
  compliance: '合规提醒',
  localization: '本地化内容',
  script: '脚本表达',
  digital_human: '数字人方案',
  promotion: '投放建议',
  clarification: '需要补充'
}

const followUpTypeModuleLabels: Record<string, string> = {
  product: '商品资料',
  market: '市场策略',
  platform: '平台策略',
  audience: '用户人群',
  selling_point: '核心卖点',
  material: '材质成分',
  scenario: '使用场景',
  compliance: '合规风险',
  localization: '本地化内容',
  script: '视频脚本',
  digital_human: '数字人方案',
  promotion: '投放建议',
  clarification: '信息补充'
}

const legacyDetailTypeMap: Record<FollowUpDetailKey, string> = {
  compliance: 'compliance',
  contentStyle: 'localization',
  videoExpression: 'script',
  promotion: 'promotion'
}

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
const workflowResult = computed(() => storedWorkflow.value)
const workflowTraces = computed(() => workflowResult.value?.traces || [])
const workflowStatusLabel = computed(() => {
  const status = workflowResult.value?.workflow_status || ''
  if (status === 'completed_with_fallback') return '含本地兜底'
  if (status === 'completed') return '执行完成'
  return status || '已生成'
})
const matchedComplianceRules = computed<ComplianceRule[]>(() => {
  return result.value.compliance.matchedRules?.length ? result.value.compliance.matchedRules : []
})
const criticScoreItems = computed(() => {
  const critic = workflowResult.value?.critic
  if (!critic) return []
  return [
    { label: '完整性', value: critic.completeness_score },
    { label: '合规性', value: critic.compliance_score },
    { label: '本地化', value: critic.localization_score },
    { label: '营销力', value: critic.marketing_score },
    { label: '总分', value: critic.overall_score }
  ]
})
const criticProblems = computed(() => {
  const critic = workflowResult.value?.critic
  if (!critic) return []
  const problems = critic.problems?.length ? critic.problems : ['未发现明显结构性问题，建议人工复核合规依据。']
  return problems.slice(0, 4)
})
const missingResultMessage = computed(() => isHistoryRestoring.value
  ? '正在从 Agent 历史记录恢复本次方案，请稍候。'
  : '请返回首页重新填写商品信息并启动丝路 Agent。生成结果会写入当前设备历史，刷新或从通知中心返回时也可以恢复。')

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

const restoreAgentResultFromHistory = async () => {
  if (storedResult.value && storedWorkflow.value) {
    return
  }

  const rawHistoryId = Array.isArray(route.query.historyId)
    ? route.query.historyId[0]
    : route.query.historyId
  const historyId = Number(rawHistoryId)
  if (!Number.isFinite(historyId) || historyId <= 0) {
    return
  }

  isHistoryRestoring.value = true
  try {
    const { item } = await agentAPI.getHistory(historyId)
    if (item?.input) {
      storedAgentInput.value = item.input
      saveAgentInput(item.input)
    }
    if (item?.workflow) {
      storedWorkflow.value = item.workflow
      storedResult.value = item.workflow.result
      saveAgentWorkflowResult(item.workflow)
      saveAgentResult(item.workflow.result)
      return
    }
    if (item?.result) {
      storedResult.value = item.result
      saveAgentResult(item.result)
    }
  } catch (error) {
    console.warn('Failed to restore Agent history result.', error)
  } finally {
    isHistoryRestoring.value = false
  }
}

const continueProductEntryFromAgent = () => {
  if (!storedResult.value) return
  // 结果页进入商品录入时，把 Agent 识别出的市场、卖点和合规提示沉淀为商品草稿，避免用户重复填写。
  saveAgentResultAsProductDraft(storedResult.value, storedAgentInput.value)
  router.push({ path: '/projects/create', query: { source: 'agent' } })
}

const getCurrentHistoryId = () => {
  const rawHistoryId = Array.isArray(route.query.historyId)
    ? route.query.historyId[0]
    : route.query.historyId
  const historyId = Number(rawHistoryId || storedWorkflow.value?.session_id)
  return Number.isFinite(historyId) && historyId > 0 ? historyId : undefined
}

const createMarketingProject = async () => {
  if (!storedResult.value || isCreatingProject.value) return
  isCreatingProject.value = true
  try {
    saveAgentResultAsProductDraft(storedResult.value, storedAgentInput.value)
    const created = await agentAPI.createProjectFromAgent(getCurrentHistoryId(), {
      result: storedResult.value,
      workflow: storedWorkflow.value || undefined
    })
    ElMessage.success(created.summary || '营销项目已创建')
    router.push(`/projects/${created.project_id}`)
  } catch (error) {
    ElMessage.error((error as Error).message || '创建营销项目失败')
  } finally {
    isCreatingProject.value = false
  }
}

const startContentWorkflow = (target: ContentWorkflowTarget) => {
  if (!storedResult.value) return

  // Agent 结果页的内容生产入口需要同时沉淀商品草稿和创作意图，后续页面才能复用同一套商品、市场、脚本和投放建议。
  saveAgentResultAsProductDraft(storedResult.value, storedAgentInput.value)
  saveContentWorkflowIntent(buildContentWorkflowIntentFromAgent(storedResult.value, storedAgentInput.value, target))
  router.push({
    path: target === 'digital-human' ? '/digital-human/create' : '/media/image',
    query: {
      source: 'agent',
      focus: target
    }
  })
}

const goCompliance = () => {
  if (storedResult.value) {
    saveAgentResultAsProductDraft(storedResult.value, storedAgentInput.value)
  }
  router.push({ path: '/compliance', query: { source: 'agent' } })
}

const goScriptWorkspace = () => {
  if (storedResult.value) {
    saveAgentResultAsProductDraft(storedResult.value, storedAgentInput.value)
  }
  router.push({ path: '/projects/create', query: { source: 'agent', next: 'script' } })
}

const goTimelineWorkspace = () => {
  if (storedResult.value) {
    saveAgentResultAsProductDraft(storedResult.value, storedAgentInput.value)
  }
  router.push({ path: '/projects', query: { source: 'agent', next: 'editor' } })
}

const getAgentDisplayName = (name: string) => {
  const labels: Record<string, string> = {
    PlanningAgent: '任务规划智能体',
    ProductAgent: '商品理解智能体',
    ComplianceAgent: '合规分析智能体',
    LocalizationAgent: '本地化智能体',
    ContentAgent: '内容生成智能体',
    CriticAgent: '评审反馈智能体'
  }
  return labels[name] || name
}

const getTraceStatusLabel = (status: string) => {
  if (status === 'fallback') return '兜底'
  if (status === 'completed') return '完成'
  if (status === 'running') return '执行中'
  return status || '完成'
}

const getTraceOutputSummary = (trace: AgentTrace) => {
  const output = trace.output as Record<string, any> | undefined
  if (!output || typeof output !== 'object') return '已完成该阶段处理。'
  if (trace.agent_name === 'PlanningAgent') {
    return toTraceSentence(output.task_chain || output.execution_steps, '已生成任务链和输出结构规划。')
  }
  if (trace.agent_name === 'ProductAgent') {
    return [output.product_name, output.category, toTraceSentence(output.core_selling_points, '')].filter(Boolean).join(' · ') || '已完成商品结构化理解。'
  }
  if (trace.agent_name === 'ComplianceAgent') {
    return [output.level, output.summary].filter(Boolean).join(' · ') || '已完成合规知识检索与风险判断。'
  }
  if (trace.agent_name === 'LocalizationAgent') {
    return [output.language_style, toTraceSentence(output.localized_selling_points, '')].filter(Boolean).join(' · ') || '已生成本地化策略。'
  }
  if (trace.agent_name === 'ContentAgent') {
    return toTraceSentence(output.marketing_titles, output.short_video_script?.title || '已生成营销内容初稿。')
  }
  if (trace.agent_name === 'CriticAgent') {
    const score = output.overall_score ? `总分 ${output.overall_score}/5` : ''
    const revise = output.need_revise ? '建议修订' : '可继续推进'
    return [score, revise].filter(Boolean).join(' · ') || '已完成方案评审。'
  }
  return JSON.stringify(output).slice(0, 120)
}

const toTraceSentence = (value: unknown, fallback: string) => {
  if (Array.isArray(value)) {
    return value.map((item) => String(item).trim()).filter(Boolean).slice(0, 4).join(' / ') || fallback
  }
  if (typeof value === 'string') return value.trim() || fallback
  return fallback
}

const markReminderAction = (action: string) => {
  reminderAction.value = action
}

const sendAgentMessage = () => {
  const message = agentMessage.value.trim()
  if (!message || isFollowUpLoading.value || isPlanRegenerating.value) return

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

/**
 * 功能：向后端发送一次 Agent 追问并消费 SSE 结果。
 * 参数：question 为用户追加的问题或修改要求；agentMessageId 对应页面上的“思考中”消息。
 * 返回：Promise；成功时把追问结果转成摘要卡片，失败时保留错误消息供用户重试。
 */
const requestFollowUp = async (question: string, agentMessageId: string) => {
  isFollowUpLoading.value = true
  followUpAbortController = new AbortController()
  let receivedResult = false
  let streamError = ''

  try {
    const response = await agentAPI.sendFollowUpStream({
      question,
      context: buildFollowUpContext()
    }, followUpAbortController.signal)

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
          updateAgentSummary(agentMessageId, normalizeFollowUpResult(event.data, question))
          receivedResult = true
        } else if (event.event === 'error') {
          streamError = event.data?.message || 'follow-up stream error'
        }
      }
    }

    if (buffer.trim()) {
      const event = parseSSEBlock(buffer)
      if (event?.event === 'result') {
        updateAgentSummary(agentMessageId, normalizeFollowUpResult(event.data, question))
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
  const initialSteps = message.thinkingSteps?.length ? message.thinkingSteps : thinkingSteps
  message.visibleThinkingSteps = [initialSteps[0]]

  let index = 1
  const timer = window.setInterval(() => {
    const target = findMessage(messageId)
    if (!target || target.type !== 'thinking') {
      stopThinkingSteps(messageId)
      return
    }
    const steps = target.thinkingSteps?.length ? target.thinkingSteps : thinkingSteps
    if (index < steps.length) {
      target.visibleThinkingSteps = steps.slice(0, index + 1)
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
    message.visibleThinkingSteps = [...(message.thinkingSteps?.length ? message.thinkingSteps : thinkingSteps)]
  }
  stopThinkingSteps(messageId)
}

const updateAgentSummary = (messageId: string, payload: FollowUpResult) => {
  finishThinkingSteps(messageId)
  const message = findMessage(messageId)
  if (!message) return

  message.type = 'summary'
  message.summary = payload.summary
  message.intent = payload.intent
  message.tags = payload.affectedModules
  message.details = payload.details
  message.cards = payload.cards
  message.updatedFields = payload.updatedFields
  message.missingFields = payload.missingFields
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
  message.intent = ''
  message.tags = []
  message.detailExpanded = false
  message.applied = false
  message.cards = []
  message.updatedFields = {}
  message.missingFields = []
}

const normalizeFollowUpResult = (data: any, question = ''): FollowUpResult => {
  const details = data?.details || {}
  const cards = normalizeFollowUpCards(data?.cards)
  const updatedFields = normalizeUpdatedFields(data?.updatedFields)
  const missingFields = toStringList(data?.missingFields).slice(0, 6)
  const modules = Array.isArray(data?.affectedModules)
    ? data.affectedModules.filter((item: unknown): item is string => typeof item === 'string' && item.trim() !== '')
    : []
  const inferredModules = inferAffectedModules(question, modules, cards)
  // 后端新协议使用动态 cards，旧协议使用 details；这里统一归一化，保证历史结果和新版追问都能展示。
  const normalizedDetails = {
    compliance: sanitizeDisplayText(details.compliance, '继续避免绝对化、医疗化和未经证实的认证表达，并以目标市场实际准入材料为准。'),
    contentStyle: sanitizeDisplayText(details.contentStyle, '内容语气应贴近目标人群日常表达，保留真实体验感，减少夸张承诺。'),
    videoExpression: sanitizeDisplayText(details.videoExpression, '前 3 秒突出核心场景和可见卖点，中段用真实使用画面承接。'),
    promotion: sanitizeDisplayText(details.promotion, '优先测试当前主平台短视频内容，结合完播率、点击率和评论问题继续迭代素材。')
  }
  const legacyCards = cards.length ? cards : legacyDetailsToCards(normalizedDetails, question, modules)

  return {
    summary: sanitizeDisplayText(data?.summary, '已基于当前商品和原方案，整理出适合继续优化的补充建议。'),
    intent: sanitizeIntent(data?.intent, question),
    affectedModules: modules.length ? modules.slice(0, 6) : inferredModules,
    details: normalizedDetails,
    cards: legacyCards,
    updatedFields,
    missingFields
  }
}

const inferAffectedModules = (question: string, modules: string[], cards: FollowUpDynamicCard[] = []) => {
  if (cards.length) {
    const labels = Array.from(new Set(cards.map((card) => followUpTypeModuleLabels[card.type] || card.title).filter(Boolean)))
    if (labels.length) return labels.slice(0, 6)
  }

  const keys = inferFollowUpDetailKeys(question, modules)
  const labels: Record<FollowUpDetailKey, string> = {
    compliance: '合规风险',
    contentStyle: '市场策略',
    videoExpression: '视频表达',
    promotion: '投放建议'
  }
  const inferred = Array.from(keys).map((key) => labels[key])
  return inferred.length ? inferred : ['市场策略', '合规风险']
}

const inferFollowUpDetailKeys = (question: string, modules: string[] = []) => {
  const normalizedQuestion = question.toLowerCase()
  const moduleText = modules.join(' ').toLowerCase()
  const questionMatchedKeys = followUpDetailDefinitions
    .filter((definition) =>
      definition.queryKeywords.some((keyword) => normalizedQuestion.includes(keyword.toLowerCase()))
    )
    .map((definition) => definition.key)

  if (questionMatchedKeys.length) {
    return new Set(questionMatchedKeys)
  }

  const moduleMatchedKeys = followUpDetailDefinitions
    .filter((definition) =>
      definition.moduleKeywords.some((keyword) => moduleText.includes(keyword.toLowerCase()))
    )
    .map((definition) => definition.key)

  return new Set(moduleMatchedKeys.length ? moduleMatchedKeys : (['contentStyle', 'compliance'] as FollowUpDetailKey[]))
}

const getFollowUpDetailCards = (message: FollowUpMessage): FollowUpDetailCard[] => {
  if (message.cards?.length) {
    return message.cards
  }
  if (!message.details) return []

  const selectedKeys = inferFollowUpDetailKeys(message.content, message.tags || [])
  return followUpDetailDefinitions
    .filter((definition) => selectedKeys.has(definition.key))
    .map((definition) => ({
      key: definition.key,
      title: definition.title,
      value: message.details?.[definition.key] || '',
      type: definition.key
    }))
    .filter((item) => item.value.trim())
}

const normalizeFollowUpCards = (value: unknown): FollowUpDynamicCard[] => {
  if (!Array.isArray(value)) return []

  return value
    .map((item, index) => {
      const raw = item && typeof item === 'object' ? item as Record<string, unknown> : {}
      const type = sanitizeCardType(raw.type)
      const title = sanitizeDisplayText(raw.title, followUpTypeTitles[type] || '补充建议')
      const content = sanitizeDisplayText(raw.content, '')
      return {
        key: `${type}-${index}-${title}`,
        type,
        title,
        value: content
      }
    })
    .filter((card) => card.value.trim())
}

const legacyDetailsToCards = (details: FollowUpDetails, question: string, modules: string[]) => {
  const selectedKeys = inferFollowUpDetailKeys(question, modules)
  return followUpDetailDefinitions
    .filter((definition) => selectedKeys.has(definition.key))
    .map((definition) => ({
      key: definition.key,
      type: legacyDetailTypeMap[definition.key],
      title: definition.title,
      value: details[definition.key] || ''
    }))
    .filter((card) => card.value.trim())
}

const normalizeUpdatedFields = (value: unknown): FollowUpUpdatedFields => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
  const allowedKeys = new Set<keyof FollowUpUpdatedFields>([
    'productName',
    'category',
    'targetMarket',
    'targetPlatform',
    'targetAudience',
    'coreSellingPoints',
    'materialSpec',
    'usageScenario',
    'marketingGoal',
    'budgetPreference',
    'complianceHints',
    'localizationHints',
    'description'
  ])
  const out: FollowUpUpdatedFields = {}

  Object.entries(value as Record<string, unknown>).forEach(([key, rawValue]) => {
    if (!allowedKeys.has(key as keyof FollowUpUpdatedFields)) return
    if (Array.isArray(rawValue)) {
      const list = toStringList(rawValue)
      if (list.length) {
        out[key as keyof FollowUpUpdatedFields] = list
      }
      return
    }
    const text = sanitizeOptionalText(rawValue)
    if (text) {
      out[key as keyof FollowUpUpdatedFields] = text
    }
  })

  return out
}

const toStringList = (value: unknown): string[] => {
  if (Array.isArray(value)) {
    return value.map((item) => sanitizeOptionalText(item)).filter(Boolean)
  }
  if (typeof value === 'string') {
    return value
      .split(/[，,、;；\n]/)
      .map((item) => sanitizeOptionalText(item))
      .filter(Boolean)
  }
  return []
}

const sanitizeDisplayText = (value: unknown, fallback: string) => {
  if (typeof value !== 'string') return fallback
  const cleaned = value.replace(/reasoning_content|思维链|chain-of-thought|内部推理/gi, '').replace(/\s+/g, ' ').trim()
  return cleaned || fallback
}

const sanitizeOptionalText = (value: unknown) => {
  return sanitizeDisplayText(value, '')
}

const sanitizeIntent = (value: unknown, question = '') => {
  if (typeof value === 'string' && value.trim()) return value.trim()
  if (question.includes('换') && question.includes('市场')) return 'change_market'
  if (question.includes('平台')) return 'change_platform'
  if (question.includes('换') && (question.includes('产品') || question.includes('商品'))) return 'change_product'
  return 'general_question'
}

const sanitizeCardType = (value: unknown) => {
  const type = typeof value === 'string' ? value.trim() : ''
  return followUpTypeTitles[type] ? type : 'product'
}

const getFollowUpIntentLabel = (intent: string) => {
  return followUpIntentLabels[intent] || '意图：综合补充'
}

/**
 * 功能：构造追问接口需要的当前方案上下文。
 * 参数：无；读取当前结果、原始 Agent 输入和已生成模块。
 * 返回：FollowUpContext，帮助后端判断追问会影响商品、合规、脚本、数字人还是投放模块。
 */
const buildFollowUpContext = (): FollowUpContext => {
  const current = result.value
  const input = storedAgentInput.value
  return {
    productName: current.recognizedInfo.productName,
    category: current.recognizedInfo.category || input?.category || '',
    targetMarket: current.recognizedInfo.targetMarket,
    platform: current.recognizedInfo.targetPlatform,
    audience: current.recognizedInfo.targetAudience,
    sellingPoints: current.recognizedInfo.coreSellingPoints.join(' / '),
    materialSpec: input?.materialSpec || '',
    usageScenario: input?.usageScenario || current.localization.sceneSuggestions.join(' / '),
    imageUnderstanding: current.recognizedInfo.imageUnderstanding,
    rawPrompt: input?.rawPrompt || '',
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

const applyFollowUp = async (message: FollowUpMessage) => {
  if (message.type !== 'summary' || message.applied || isPlanRegenerating.value) return
  const nextResult = applyFollowUpToResult(result.value, message)

  if (shouldRegeneratePlan(message)) {
    // 换商品、换市场等会改变合规边界，必须重新生成完整方案，不能只在页面局部替换文案。
    await regenerateCurrentPlan(message, nextResult)
  } else {
    storedResult.value = nextResult
    saveAgentResult(nextResult)
    persistFollowUpProductDraft(message, nextResult)
  }

  message.applied = true
  const nextModules = new Set(appliedUpdateModules.value)
  ;(message.tags || []).forEach((tag) => nextModules.add(tag))
  if (shouldRegeneratePlan(message)) {
    nextModules.add('方案重编排')
  }
  appliedUpdateModules.value = Array.from(nextModules)
}

const regenerateCurrentPlan = async (message: FollowUpMessage, baseResult: AgentResult) => {
  const thinkingMessage: FollowUpMessage = {
    id: createMessageId('agent-regeneration'),
    role: 'agent',
    type: 'thinking',
    content: message.content,
    intent: message.intent,
    regenerating: true,
    thinkingSteps: [...regenerationSteps],
    visibleThinkingSteps: []
  }
  messages.value.push(thinkingMessage)
  startThinkingSteps(thinkingMessage.id)
  isPlanRegenerating.value = true
  scrollFollowUpChat()

  try {
    const input = buildRegenerationInput(baseResult, message)
    storedAgentInput.value = input
    saveAgentInput(input)
    const regenerated = await agentAPI.generate(input)
    const nextResult = mergeRegeneratedResult(regenerated, message)
    storedResult.value = nextResult
    saveAgentResult(nextResult)
    saveAgentResultAsProductDraft(nextResult, input)
    persistFollowUpProductDraft(message, nextResult)
    updateRegenerationSummary(thinkingMessage.id, message, nextResult)
  } catch {
    storedResult.value = baseResult
    saveAgentResult(baseResult)
    persistFollowUpProductDraft(message, baseResult)
    updateRegenerationFallback(thinkingMessage.id, message)
  } finally {
    stopThinkingSteps(thinkingMessage.id)
    isPlanRegenerating.value = false
    scrollFollowUpChat()
  }
}

const shouldRegeneratePlan = (message: FollowUpMessage) => {
  if (message.intent === 'ask_clarification' || message.missingFields?.length) {
    return false
  }

  const regenerationIntents = new Set([
    'change_product',
    'change_market',
    'change_platform',
    'change_audience',
    'add_product_info',
    'add_material_info',
    'add_usage_scenario',
    'add_image_info',
    'adjust_content_tone',
    'optimize_script',
    'optimize_compliance',
    'optimize_promotion'
  ])
  if (message.intent && regenerationIntents.has(message.intent)) {
    return true
  }

  const fieldKeys = Object.keys(message.updatedFields || {})
  return fieldKeys.some((key) =>
    ['productName', 'category', 'targetMarket', 'targetPlatform', 'targetAudience', 'coreSellingPoints', 'materialSpec', 'usageScenario'].includes(key)
  )
}

const buildRegenerationInput = (baseResult: AgentResult, message: FollowUpMessage): AgentInput => {
  const previous = storedAgentInput.value
  const draft = getProductDraft()
  const fields = message.updatedFields || {}
  const productName = fieldAsText(fields.productName) || baseResult.recognizedInfo.productName || previous?.productName || draft?.productName || ''
  const category = fieldAsText(fields.category) || baseResult.recognizedInfo.category || previous?.category || draft?.category || ''
  const targetMarket = fieldAsText(fields.targetMarket) || baseResult.recognizedInfo.targetMarket || previous?.targetMarket || draft?.targetMarket || ''
  const targetPlatform = fieldAsText(fields.targetPlatform) || baseResult.recognizedInfo.targetPlatform || previous?.targetPlatform || draft?.targetPlatform || ''
  const targetAudience = fieldAsText(fields.targetAudience) || baseResult.recognizedInfo.targetAudience || previous?.targetAudience || draft?.targetAudience || ''
  const coreSellingPoints = mergeUnique(
    fieldAsList(fields.coreSellingPoints),
    baseResult.recognizedInfo.coreSellingPoints,
    previous?.coreSellingPoints,
    draft?.coreSellingPoints
  )
  const materialSpec = fieldAsText(fields.materialSpec) || previous?.materialSpec || draft?.materialSpec || ''
  const usageScenario =
    fieldAsText(fields.usageScenario) ||
    previous?.usageScenario ||
    draft?.usageScenario ||
    baseResult.localization.sceneSuggestions.join('、')

  return {
    ...previous,
    productName,
    category,
    targetMarket,
    targetPlatform,
    targetAudience,
    coreSellingPoints,
    materialSpec,
    usageScenario,
    rawPrompt: mergeText(previous?.rawPrompt || draft?.description || '', `追问更新：${message.content}`),
    imageDataUrl: previous?.imageDataUrl || draft?.productImage || ''
  }
}

const mergeRegeneratedResult = (regenerated: AgentResult, message: FollowUpMessage): AgentResult => {
  const next = cloneAgentResult(regenerated)
  next.agentMessage.summary = mergeText(next.agentMessage.summary, message.summary || '')
  if (message.missingFields?.length) {
    next.compliance.missingInfo = mergeUnique(message.missingFields, next.compliance.missingInfo)
    next.agentMessage.missingInfoNotice = `建议继续补充：${message.missingFields.join('、')}。`
  }
  return next
}

const updateRegenerationSummary = (messageId: string, sourceMessage: FollowUpMessage, nextResult: AgentResult) => {
  finishThinkingSteps(messageId)
  const message = findMessage(messageId)
  if (!message) return

  message.type = 'summary'
  message.summary = `当前方案已根据“${sourceMessage.content}”重新编排，主结果区和商品资料已同步更新。`
  message.intent = sourceMessage.intent
  message.tags = mergeUnique(['方案已重新编排'], sourceMessage.tags || [])
  message.cards = buildRegenerationResultCards(nextResult)
  message.updatedFields = sourceMessage.updatedFields
  message.missingFields = []
  message.applied = true
  scrollFollowUpChat()
}

const updateRegenerationFallback = (messageId: string, sourceMessage: FollowUpMessage) => {
  finishThinkingSteps(messageId)
  const message = findMessage(messageId)
  if (!message) return

  message.type = 'summary'
  message.summary = '完整方案重新编排暂时未完成，已先将本次追问能确定的字段同步到当前方案。'
  message.intent = sourceMessage.intent
  message.tags = mergeUnique(['已应用字段更新'], sourceMessage.tags || [])
  message.cards = [
    {
      key: `fallback-${messageId}`,
      type: 'clarification',
      title: '重新编排未完成',
      value: '当前已保留追问更新；可稍后重新发送本次要求，让 Agent 再次完整生成方案。'
    }
  ]
  message.updatedFields = sourceMessage.updatedFields
  message.missingFields = sourceMessage.missingFields || []
  message.applied = true
  scrollFollowUpChat()
}

const buildRegenerationResultCards = (nextResult: AgentResult): FollowUpDynamicCard[] => [
  {
    key: 'regenerated-product',
    type: 'product',
    title: '商品资料',
    value: [nextResult.recognizedInfo.productName, nextResult.recognizedInfo.category].filter(Boolean).join(' / ') || '商品资料已更新'
  },
  {
    key: 'regenerated-market',
    type: 'market',
    title: '市场与平台',
    value: [nextResult.recognizedInfo.targetMarket, nextResult.recognizedInfo.targetPlatform].filter(Boolean).join(' / ') || nextResult.overview.marketStrategy
  },
  {
    key: 'regenerated-compliance',
    type: 'compliance',
    title: '合规结论',
    value: nextResult.compliance.summary || nextResult.overview.complianceRiskLevel
  },
  {
    key: 'regenerated-script',
    type: 'script',
    title: '内容脚本',
    value: nextResult.script.opening.content || nextResult.overview.recommendedVideoStyle
  }
].filter((card) => card.value.trim())

const fieldAsText = (value: FollowUpFieldValue | undefined) => {
  const text = Array.isArray(value) ? value[0] : value
  return sanitizeOptionalText(text)
}

const fieldAsList = (value: FollowUpFieldValue | undefined) => toStringList(value)

const getFollowUpActionLabel = (message: FollowUpMessage) => {
  if (message.applied) return '已应用'
  if (isPlanRegenerating.value) return '重新编排中'
  if (shouldRegeneratePlan(message)) return '应用并重新编排方案'
  if (message.intent === 'ask_clarification') return '记录补充需求'
  return '应用到当前方案'
}

const applyFollowUpToResult = (current: AgentResult, message: FollowUpMessage): AgentResult => {
  const next = cloneAgentResult(current)
  const fields = message.updatedFields || {}

  setStringField(fields.productName, (value) => {
    next.recognizedInfo.productName = value
  })
  setStringField(fields.category, (value) => {
    next.recognizedInfo.category = value
  })
  setStringField(fields.targetMarket, (value) => {
    next.recognizedInfo.targetMarket = value
    next.localization.direction = `${value}市场本地化方向`
    next.overview.marketStrategy = `围绕${value}市场重新校准商品表达、合规提示和投放节奏。`
  })
  setStringField(fields.targetPlatform, (value) => {
    next.recognizedInfo.targetPlatform = value
    next.promotion.platforms = mergeUnique([value], next.promotion.platforms)
  })
  setStringField(fields.targetAudience, (value) => {
    next.recognizedInfo.targetAudience = value
  })
  setListField(fields.coreSellingPoints, (values) => {
    next.recognizedInfo.coreSellingPoints = mergeUnique(values, next.recognizedInfo.coreSellingPoints)
  })
  setStringField(fields.materialSpec, (value) => {
    next.recognizedInfo.imageUnderstanding = mergeText(next.recognizedInfo.imageUnderstanding, `材质/成分：${value}`)
    next.compliance.missingInfo = next.compliance.missingInfo.filter((item) => !item.includes('材质') && !item.includes('成分'))
  })
  setStringField(fields.usageScenario, (value) => {
    next.localization.sceneSuggestions = mergeUnique([value], next.localization.sceneSuggestions)
  })
  setListField(fields.marketingGoal, (values) => {
    next.promotion.focusMetrics = mergeUnique(values, next.promotion.focusMetrics)
  })
  setStringField(fields.budgetPreference, (value) => {
    next.promotion.optimizationAdvice = mergeText(next.promotion.optimizationAdvice, `预算偏好：${value}`)
  })
  setListField(fields.complianceHints, (values) => {
    next.compliance.suggestions = mergeUnique(values, next.compliance.suggestions)
  })
  setListField(fields.localizationHints, (values) => {
    next.localization.keywords = mergeUnique(values, next.localization.keywords)
  })
  setStringField(fields.description, (value) => {
    next.agentMessage.summary = mergeText(next.agentMessage.summary, value)
  })

  applyFollowUpCardsToResult(next, message.cards || [])
  if (message.summary) {
    next.agentMessage.summary = mergeText(next.agentMessage.summary, message.summary)
  }
  if (message.missingFields?.length) {
    next.compliance.missingInfo = mergeUnique(message.missingFields, next.compliance.missingInfo)
    next.agentMessage.missingInfoNotice = `建议继续补充：${message.missingFields.join('、')}。`
  }

  return next
}

const applyFollowUpCardsToResult = (next: AgentResult, cards: FollowUpDynamicCard[]) => {
  cards.forEach((card) => {
    if (card.type === 'compliance') {
      next.compliance.suggestions = mergeUnique([card.value], next.compliance.suggestions)
      next.compliance.summary = mergeText(next.compliance.summary, card.value)
    } else if (['market', 'localization', 'audience', 'scenario'].includes(card.type)) {
      next.localization.sceneSuggestions = mergeUnique([card.value], next.localization.sceneSuggestions)
      next.localization.reason = mergeText(next.localization.reason, card.value)
    } else if (card.type === 'script') {
      next.script.opening.content = mergeText(next.script.opening.content, card.value)
      next.overview.recommendedVideoStyle = card.title
    } else if (card.type === 'digital_human') {
      next.digitalHuman.tone = mergeText(next.digitalHuman.tone, card.value)
    } else if (card.type === 'promotion' || card.type === 'platform') {
      next.promotion.optimizationAdvice = mergeText(next.promotion.optimizationAdvice, card.value)
    }
  })
}

const persistFollowUpProductDraft = (message: FollowUpMessage, nextResult: AgentResult) => {
  if (!getProductDraft()) {
    saveAgentResultAsProductDraft(nextResult, storedAgentInput.value)
  }

  const currentDraft = getProductDraft()
  const fields = message.updatedFields || {}
  const complianceHints = mergeUnique(
    [
      ...(currentDraft?.complianceHints || []),
      ...toStringList(fields.complianceHints),
      ...getCardsByType(message.cards || [], ['compliance']).map((card) => card.value)
    ]
  )
  const localizationHints = mergeUnique(
    [
      ...(currentDraft?.localizationHints || []),
      ...toStringList(fields.localizationHints),
      ...getCardsByType(message.cards || [], ['market', 'localization', 'audience', 'scenario', 'script', 'digital_human', 'promotion', 'platform']).map((card) => card.value)
    ]
  )
  const patch: Parameters<typeof saveProductDraft>[0] = {
    source: 'agent',
    agentSummary: mergeText(currentDraft?.agentSummary || '', message.summary || ''),
    complianceHints,
    localizationHints
  }

  setDraftTextField(patch, 'productName', fields.productName)
  setDraftTextField(patch, 'category', fields.category)
  setDraftTextField(patch, 'targetMarket', fields.targetMarket)
  setDraftTextField(patch, 'targetPlatform', fields.targetPlatform)
  setDraftTextField(patch, 'targetAudience', fields.targetAudience)
  setDraftTextField(patch, 'materialSpec', fields.materialSpec)
  setDraftTextField(patch, 'usageScenario', fields.usageScenario)
  setDraftTextField(patch, 'budgetPreference', fields.budgetPreference)
  setDraftTextField(patch, 'description', fields.description)
  setDraftListField(patch, 'coreSellingPoints', fields.coreSellingPoints)
  setDraftListField(patch, 'marketingGoal', fields.marketingGoal)

  saveProductDraft(patch)
}

const cloneAgentResult = (value: AgentResult): AgentResult => JSON.parse(JSON.stringify(value)) as AgentResult

const setDraftTextField = (
  patch: Parameters<typeof saveProductDraft>[0],
  key: 'productName' | 'category' | 'targetMarket' | 'targetPlatform' | 'targetAudience' | 'materialSpec' | 'usageScenario' | 'budgetPreference' | 'description',
  value: FollowUpFieldValue | undefined
) => {
  const text = Array.isArray(value) ? value[0] : value
  if (typeof text === 'string' && text.trim()) {
    patch[key] = text.trim()
  }
}

const setDraftListField = (
  patch: Parameters<typeof saveProductDraft>[0],
  key: 'coreSellingPoints' | 'marketingGoal',
  value: FollowUpFieldValue | undefined
) => {
  const list = toStringList(value)
  if (list.length) {
    patch[key] = list
  }
}

const setStringField = (value: FollowUpFieldValue | undefined, setter: (value: string) => void) => {
  const text = Array.isArray(value) ? value[0] : value
  if (typeof text === 'string' && text.trim()) {
    setter(text.trim())
  }
}

const setListField = (value: FollowUpFieldValue | undefined, setter: (value: string[]) => void) => {
  const list = toStringList(value)
  if (list.length) {
    setter(list)
  }
}

const mergeUnique = (...groups: Array<Array<string | undefined> | undefined>) => {
  const seen = new Set<string>()
  const merged: string[] = []
  groups.flat().forEach((item) => {
    const text = sanitizeOptionalText(item)
    if (!text || seen.has(text)) return
    seen.add(text)
    merged.push(text)
  })
  return merged
}

const mergeText = (base: string | undefined, addition: string | undefined) => {
  const current = sanitizeOptionalText(base)
  const next = sanitizeOptionalText(addition)
  if (!next) return current
  if (!current) return next
  if (current.includes(next)) return current
  return `${current}；${next}`
}

const getCardsByType = (cards: FollowUpDynamicCard[], types: string[]) => {
  const wanted = new Set(types)
  return cards.filter((card) => wanted.has(card.type))
}

const regenerateFollowUp = (message: FollowUpMessage) => {
  if (isFollowUpLoading.value || isPlanRegenerating.value || !message.content) return
  message.type = 'thinking'
  message.thinkingSteps = [...thinkingSteps]
  message.visibleThinkingSteps = []
  message.summary = ''
  message.intent = ''
  message.tags = []
  message.details = undefined
  message.cards = []
  message.updatedFields = {}
  message.missingFields = []
  message.detailExpanded = false
  message.applied = false
  startThinkingSteps(message.id)
  void requestFollowUp(message.content, message.id)
}

type AppliedModuleKey = 'compliance' | 'video' | 'localization' | 'digitalHuman' | 'promotion'

const hasAppliedUpdate = (module: AppliedModuleKey) => {
  const modules = appliedUpdateModules.value.join(' ')
  if (modules.includes('方案重编排')) {
    return true
  }
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

onMounted(() => {
  void restoreAgentResultFromHistory()
})

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

.agent-button:disabled {
  cursor: not-allowed;
  opacity: 0.68;
  transform: none;
}

.agent-card-action-row {
  margin-top: 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.agent-card-action-row--compact {
  margin-top: 14px;
}

.agent-card-action-button {
  min-height: 34px;
  border: 1px solid #dbeafe;
  border-radius: 999px;
  padding: 8px 13px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  background: #ffffff;
  color: #0a2463;
  font-family: inherit;
  font-size: 11px;
  line-height: 16px;
  font-weight: 800;
  cursor: pointer;
  transition: transform 180ms ease, box-shadow 180ms ease, border-color 180ms ease, background 180ms ease;
}

.agent-card-action-button:hover {
  transform: translateY(-1px);
  border-color: rgba(6, 182, 212, 0.45);
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.1);
}

.agent-card-action-button--primary {
  border-color: transparent;
  background: linear-gradient(135deg, #06b6d4 0%, #2563eb 100%);
  color: #ffffff;
}

.agent-card-action-button--pink {
  border-color: transparent;
  background: linear-gradient(135deg, #d946ef 0%, #7c3aed 100%);
  color: #ffffff;
}

.agent-card-action-button--green {
  border-color: transparent;
  background: linear-gradient(135deg, #10b981 0%, #0891b2 100%);
  color: #ffffff;
}

.agent-card-action-button__icon {
  width: 13px;
  height: 13px;
  flex: 0 0 auto;
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

.agent-next-card {
  margin-top: 16px;
  padding: 18px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.agent-next-card strong {
  color: #0a2463;
  font-size: 16px;
  line-height: 24px;
}

.agent-next-card p {
  margin: 4px 0 0;
  color: #45556c;
  font-size: 14px;
  line-height: 20px;
}

.agent-next-card__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.agent-workflow-card,
.agent-shortcut-card {
  margin-top: 16px;
  padding: 22px;
}

.agent-trace-list {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.agent-trace-item {
  min-height: 132px;
  padding: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.agent-trace-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.agent-trace-item__head strong {
  color: #0a2463;
  font-size: 14px;
  line-height: 20px;
}

.agent-trace-item p {
  margin: 0;
  color: #45556c;
  font-size: 13px;
  line-height: 20px;
  flex: 1 1 auto;
}

.agent-trace-item small {
  color: #90a1b9;
  font-size: 11px;
  line-height: 16px;
}

.agent-trace-item__status {
  min-width: 44px;
  padding: 3px 8px;
  border-radius: 999px;
  text-align: center;
  font-size: 11px;
  line-height: 16px;
  font-weight: 800;
}

.agent-trace-item__status--completed {
  background: #dcfce7;
  color: #15803d;
}

.agent-trace-item__status--fallback {
  background: #fef3c7;
  color: #92400e;
}

.agent-trace-item__status--running {
  background: #dbeafe;
  color: #1d4ed8;
}

.agent-critic-panel {
  margin-top: 16px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
  gap: 14px;
}

.agent-critic-panel__scores {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

.agent-critic-panel__scores article {
  min-height: 62px;
  padding: 10px;
  border-radius: 10px;
  background: #f8fafc;
  text-align: center;
}

.agent-critic-panel__scores span,
.agent-rule-panel__head span {
  color: #64748b;
  font-size: 12px;
  line-height: 16px;
}

.agent-critic-panel__scores strong {
  display: block;
  margin-top: 4px;
  color: #0a2463;
  font-size: 18px;
  line-height: 24px;
}

.agent-critic-panel__feedback {
  padding: 12px 14px;
  border-radius: 12px;
  background: #f8fafc;
}

.agent-critic-panel__feedback strong,
.agent-rule-panel__head strong {
  color: #0a2463;
  font-size: 14px;
  line-height: 20px;
}

.agent-critic-panel__feedback ul {
  margin: 8px 0 0;
  padding-left: 18px;
  color: #45556c;
  font-size: 13px;
  line-height: 20px;
}

.agent-rule-panel {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.agent-rule-panel__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.agent-rule-item {
  margin-top: 10px;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.agent-rule-item div {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px;
}

.agent-rule-item strong {
  color: #0a2463;
  font-size: 13px;
  line-height: 18px;
}

.agent-rule-item span {
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
}

.agent-rule-item p,
.agent-rule-panel__disclaimer {
  margin: 6px 0 0;
  color: #45556c;
  font-size: 13px;
  line-height: 20px;
}

.agent-rule-panel__disclaimer {
  color: #92400e;
}

.agent-shortcut-card {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.agent-shortcut-button {
  min-height: 46px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 10px 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: #ffffff;
  color: #0a2463;
  font-family: inherit;
  font-size: 13px;
  line-height: 18px;
  font-weight: 800;
  cursor: pointer;
  transition: transform 180ms ease, box-shadow 180ms ease, border-color 180ms ease;
}

.agent-shortcut-button svg {
  width: 16px;
  height: 16px;
}

.agent-shortcut-button:hover {
  transform: translateY(-1px);
  border-color: rgba(6, 182, 212, 0.42);
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.09);
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

.agent-follow-up-missing {
  margin-top: 10px;
  padding: 10px 11px;
  border: 1px solid rgba(249, 115, 22, 0.22);
  border-radius: 12px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 7px;
  background: rgba(255, 247, 237, 0.72);
}

.agent-follow-up-missing strong,
.agent-follow-up-missing span {
  font-size: 11px;
  line-height: 17px;
}

.agent-follow-up-missing strong {
  color: #9a3412;
}

.agent-follow-up-missing span {
  padding: 3px 8px;
  border-radius: 999px;
  background: #ffedd5;
  color: #c2410c;
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

  .agent-next-card {
    align-items: stretch;
    flex-direction: column;
  }

  .agent-next-card__actions {
    justify-content: flex-start;
  }

  .agent-trace-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .agent-critic-panel {
    grid-template-columns: 1fr;
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

  .agent-shortcut-card {
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
    .agent-next-card,
    .agent-workflow-card,
    .agent-shortcut-card,
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

  .agent-trace-list,
  .agent-shortcut-card,
  .agent-critic-panel__scores {
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
