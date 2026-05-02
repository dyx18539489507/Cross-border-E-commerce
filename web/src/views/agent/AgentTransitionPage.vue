<template>
  <MobileAgentTransitionPage v-if="isMobileViewport" />
  <main v-else class="agent-transition-page" aria-labelledby="agent-transition-title">
    <section class="agent-transition-card">
      <div class="agent-transition-card__top-line" aria-hidden="true"></div>
      <div class="agent-transition-card__glow" aria-hidden="true"></div>

      <header class="agent-transition-header">
        <div class="agent-transition-header__copy">
          <span class="agent-transition-header__icon-wrap" aria-hidden="true">
            <Cpu class="agent-transition-header__icon" />
            <span class="agent-transition-header__dot"></span>
          </span>

          <div>
            <h1 id="agent-transition-title">{{ titleText }}</h1>
            <p>{{ subtitleText }}</p>
          </div>
        </div>

        <span class="agent-transition-header__badge" :class="{ 'is-error': status === 'error' }">
          <CircleCheck class="agent-transition-header__badge-icon" aria-hidden="true" />
          <span>{{ statusBadge }}</span>
        </span>
      </header>

      <AgentRecognizedInfo :input="agentInput" :result="agentResult" />
      <AgentTaskChain :active-count="activeTaskCount" :failed="status === 'error'" />
      <AgentCompleteNotice v-if="status === 'success'" />

      <section v-else-if="status === 'error'" class="agent-error-card" aria-live="assertive">
        <strong>生成失败</strong>
        <p>{{ errorMessage }}</p>
        <div class="agent-error-card__actions">
          <button type="button" @click="retryGenerate">重新生成</button>
          <button type="button" class="agent-error-card__secondary" @click="goHome">返回首页</button>
        </div>
      </section>

      <section v-else class="agent-running-notice" aria-live="polite">
        <span class="agent-running-notice__spinner" aria-hidden="true"></span>
        <span>丝路 Agent 正在分析商品、市场与合规风险，请稍候……</span>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CircleCheck, Cpu } from '@element-plus/icons-vue'
import { agentAPI } from '@/api/agent'
import type { AgentInput, AgentResult } from '@/types/agent'
import { readAgentInput, saveAgentResult } from '@/utils/agentStorage'
import AgentCompleteNotice from './components/AgentCompleteNotice.vue'
import AgentRecognizedInfo from './components/AgentRecognizedInfo.vue'
import AgentTaskChain from './components/AgentTaskChain.vue'
import MobileAgentTransitionPage from './MobileAgentTransitionPage.vue'

const router = useRouter()
let redirectTimer: number | undefined
let progressTimer: number | undefined
let mediaQuery: MediaQueryList | undefined
const status = ref<'running' | 'success' | 'error'>('running')
const activeTaskCount = ref(0)
const errorMessage = ref('')
const agentInput = ref<AgentInput | null>(null)
const agentResult = ref<AgentResult | null>(null)
const isMobileViewport = ref(false)

const titleText = computed(() => {
  if (status.value === 'success') return '丝路 Agent 已完成方案生成'
  if (status.value === 'error') return '丝路 Agent 生成遇到问题'
  return '丝路 Agent 正在生成出海营销方案'
})

const subtitleText = computed(() => {
  if (status.value === 'success') return '已完成商品理解、合规分析、本地化内容、数字人方案与投放建议。'
  if (status.value === 'error') return '你可以重新生成，或返回首页补充更多商品信息后再启动。'
  return '正在执行商品理解、合规识别、本地化方向、短视频脚本、数字人方案与投放优化。'
})

const statusBadge = computed(() => {
  if (status.value === 'success') return 'AI 分析完成'
  if (status.value === 'error') return '需要重试'
  return 'AI 分析中'
})

const startProgress = () => {
  window.clearInterval(progressTimer)
  activeTaskCount.value = 0
  progressTimer = window.setInterval(() => {
    if (status.value !== 'running') return
    activeTaskCount.value = Math.min(activeTaskCount.value + 1, 5)
  }, 700)
}

const finishAndRedirect = (result: AgentResult) => {
  agentResult.value = result
  saveAgentResult(result)
  status.value = 'success'
  activeTaskCount.value = 6
  window.clearInterval(progressTimer)
  redirectTimer = window.setTimeout(() => {
    router.push('/agent/result')
  }, 1100)
}

const generate = async () => {
  const input = readAgentInput()
  if (!input) {
    errorMessage.value = '没有找到本次 Agent 输入，请返回首页重新填写商品信息。'
    status.value = 'error'
    return
  }
  agentInput.value = input
  status.value = 'running'
  errorMessage.value = ''
  startProgress()

  try {
    const result = await agentAPI.generate(input)
    finishAndRedirect(result)
  } catch (error) {
    status.value = 'error'
    activeTaskCount.value = Math.max(activeTaskCount.value, 1)
    window.clearInterval(progressTimer)
    errorMessage.value = error instanceof Error ? error.message : '丝路 Agent 生成失败，请稍后重试。'
  }
}

const retryGenerate = () => {
  generate()
}

const goHome = () => {
  router.push('/')
}

const updateMobileViewport = () => {
  isMobileViewport.value = mediaQuery?.matches ?? window.innerWidth <= 768
}

onMounted(() => {
  mediaQuery = window.matchMedia('(max-width: 768px)')
  updateMobileViewport()
  mediaQuery.addEventListener('change', updateMobileViewport)
  if (!isMobileViewport.value) {
    generate()
  }
})

onBeforeUnmount(() => {
  mediaQuery?.removeEventListener('change', updateMobileViewport)
  if (redirectTimer) {
    window.clearTimeout(redirectTimer)
  }
  if (progressTimer) {
    window.clearInterval(progressTimer)
  }
})
</script>

<style scoped>
.agent-transition-page {
  position: relative;
  min-height: 100vh;
  padding: 64px 24px;
  display: grid;
  place-items: center;
  overflow: hidden;
  background:
    linear-gradient(144deg, #f8fafc 0%, #ffffff 50%, rgba(124, 58, 237, 0.05) 100%),
    linear-gradient(120deg, rgba(6, 182, 212, 0.08), rgba(249, 115, 22, 0.04));
  color: #0a2463;
}

.agent-transition-page::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(135deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0) 35%),
    linear-gradient(315deg, rgba(124, 58, 237, 0.14) 0%, rgba(255, 255, 255, 0) 42%);
}

.agent-transition-card {
  position: relative;
  z-index: 1;
  width: min(820px, 100%);
  padding: 32px;
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  animation: transitionCardIn 520ms ease both;
}

.agent-transition-card__top-line {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 4px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 50%, #f97316 100%);
}

.agent-transition-card__glow {
  position: absolute;
  right: -48px;
  top: -88px;
  width: 240px;
  height: 240px;
  pointer-events: none;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.14), rgba(124, 58, 237, 0.14));
  filter: blur(64px);
}

.agent-transition-header {
  position: relative;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.agent-transition-header__copy {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.agent-transition-header__icon-wrap {
  position: relative;
  width: 48px;
  height: 48px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  box-shadow: 0 10px 15px rgba(124, 58, 237, 0.3), 0 4px 6px rgba(124, 58, 237, 0.3);
  color: #ffffff;
}

.agent-transition-header__icon {
  width: 24px;
  height: 24px;
}

.agent-transition-header__dot {
  position: absolute;
  right: -4px;
  top: -4px;
  width: 14px;
  height: 14px;
  border: 2px solid #ffffff;
  border-radius: 999px;
  background: #05df72;
}

.agent-transition-header h1 {
  margin: 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 20px;
  line-height: 28px;
  font-weight: 700;
}

.agent-transition-header p {
  margin: 4px 0 0;
  color: #62748e;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 14px;
  line-height: 20px;
  white-space: normal;
}

.agent-transition-header__badge {
  min-height: 27px;
  padding: 5px 10px;
  border: 1px solid #b9f8cf;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex: 0 0 auto;
  background: #f0fdf4;
  color: #00a63e;
  font-family: 'IBM Plex Sans', 'Noto Sans SC', sans-serif;
  font-size: 11px;
  line-height: 17px;
  font-weight: 700;
  white-space: nowrap;
}

.agent-transition-header__badge-icon {
  width: 12px;
  height: 12px;
}

.agent-transition-header__badge.is-error {
  border-color: #fecaca;
  background: #fef2f2;
  color: #dc2626;
}

.agent-running-notice,
.agent-error-card {
  min-height: 57px;
  padding: 13px 17px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.agent-running-notice {
  border: 1px solid rgba(6, 182, 212, 0.22);
  background: linear-gradient(90deg, rgba(236, 254, 255, 0.86), rgba(245, 243, 255, 0.72));
  color: #0891b2;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.agent-running-notice__spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(6, 182, 212, 0.25);
  border-top-color: #06b6d4;
  border-radius: 999px;
  animation: agentSpin 900ms linear infinite;
  flex: 0 0 auto;
}

.agent-error-card {
  border: 1px solid #fecaca;
  align-items: flex-start;
  flex-direction: column;
  background: #fef2f2;
  color: #991b1b;
}

.agent-error-card strong {
  font-size: 14px;
  line-height: 20px;
}

.agent-error-card p {
  margin: 0;
  color: #b91c1c;
  font-size: 13px;
  line-height: 20px;
}

.agent-error-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.agent-error-card button {
  min-height: 34px;
  padding: 8px 14px;
  border: 0;
  border-radius: 12px;
  background: #dc2626;
  color: #ffffff;
  font-family: inherit;
  font-weight: 700;
  cursor: pointer;
}

.agent-error-card__secondary {
  background: #ffffff !important;
  color: #991b1b !important;
  border: 1px solid #fecaca !important;
}

@keyframes agentSpin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes transitionCardIn {
  from {
    opacity: 0;
    transform: translateY(14px) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@media (max-width: 720px) {
  .agent-transition-page {
    padding: 32px 16px;
    align-items: flex-start;
  }

  .agent-transition-card {
    padding: 24px;
    border-radius: 22px;
    gap: 22px;
  }

  .agent-transition-header {
    flex-direction: column;
  }
}

@media (max-width: 520px) {
  .agent-transition-card {
    padding: 22px 18px;
  }

  .agent-transition-header__copy {
    gap: 10px;
  }

  .agent-transition-header h1 {
    font-size: 18px;
    line-height: 26px;
  }

  .agent-transition-header p {
    font-size: 13px;
    line-height: 19px;
  }
}
</style>
