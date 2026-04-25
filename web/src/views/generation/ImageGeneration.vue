<template>
  <div class="page-container content-creation-page">
    <header class="creation-header">
      <div class="creation-header__inner">
        <div class="creation-header__left">
          <button type="button" class="brand-link" @click="router.push('/')">
            <span class="brand-link__mark">
              <img :src="brandIcon" alt="" />
            </span>
            <span class="brand-link__name">{{ t('app.name') }}</span>
          </button>

          <nav class="creation-nav" aria-label="主导航">
            <button
              v-for="item in navItems"
              :key="item.label"
              type="button"
              class="creation-nav__item"
              :class="{ 'creation-nav__item--active': item.active }"
              :style="{ width: item.width }"
              @click="handleNavClick(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="creation-header__right">
          <button type="button" class="header-icon-button" aria-label="通知">
            <img :src="bellIcon" alt="" />
            <span class="header-icon-button__dot"></span>
          </button>
        </div>
      </div>
    </header>

    <main class="content-creation-main">
      <div class="content-creation-shell">
        <section class="creation-hero">
          <h1>多模态内容创作工作台</h1>
          <p>AI生成商品图片、短视频片段、数字人口播，自由组合输出成片</p>
        </section>

        <section class="creation-mode-row" aria-label="创作模式">
          <button
            v-for="mode in creationModes"
            :key="mode.key"
            type="button"
            class="creation-mode-card"
            :class="[
              `creation-mode-card--${mode.key}`,
              { 'creation-mode-card--active': selectedMode === mode.key }
            ]"
            @click="handleSelectMode(mode.key)"
          >
            <span class="creation-mode-card__icon" aria-hidden="true">
              <svg v-if="mode.key === 'combo'" viewBox="0 0 32 32" fill="none">
                <path d="M8.667 11.333L16 7.333L23.333 11.333L16 15.333L8.667 11.333Z" />
                <path d="M8.667 16L16 20L23.333 16" />
                <path d="M8.667 20.667L16 24.667L23.333 20.667" />
              </svg>
              <svg v-else-if="mode.key === 'image'" viewBox="0 0 32 32" fill="none">
                <rect x="6.667" y="7.333" width="18.667" height="17.333" rx="2.667" />
                <circle cx="12" cy="13.333" r="2" />
                <path d="M8.667 20L13.333 15.333L17.333 19.333L20 16.667L23.333 20" />
              </svg>
              <svg v-else-if="mode.key === 'video'" viewBox="0 0 32 32" fill="none">
                <rect x="6.667" y="9.333" width="13.333" height="13.333" rx="2.667" />
                <path d="M20 13.333L25.333 10.667V21.333L20 18.667V13.333Z" />
              </svg>
              <svg v-else viewBox="0 0 32 32" fill="none">
                <path d="M16 17.333C19.682 17.333 22.667 14.348 22.667 10.667C22.667 6.985 19.682 4 16 4C12.318 4 9.333 6.985 9.333 10.667C9.333 14.348 12.318 17.333 16 17.333Z" />
                <path d="M6.667 26.667C6.667 21.512 10.845 17.333 16 17.333C21.155 17.333 25.333 21.512 25.333 26.667" />
              </svg>
            </span>
            <strong>{{ mode.label }}</strong>
            <span>{{ mode.description }}</span>
          </button>
        </section>

        <section class="workspace-grid">
          <div class="workspace-main">
            <article class="panel-card panel-card--selection">
              <div class="panel-card__header">
                <h2>选择内容形式</h2>
              </div>

              <div class="content-form-grid">
                <button
                  type="button"
                  class="content-form-card content-form-card--image"
                  :class="{ 'content-form-card--active': selectedFormats.includes('image') }"
                  @click="toggleFormat('image')"
                >
                  <div class="content-form-card__top">
                    <span class="content-form-card__icon" aria-hidden="true">
                      <svg viewBox="0 0 32 32" fill="none">
                        <rect x="6.667" y="7.333" width="18.667" height="17.333" rx="2.667" />
                        <circle cx="12" cy="13.333" r="2" />
                        <path d="M8.667 20L13.333 15.333L17.333 19.333L20 16.667L23.333 20" />
                      </svg>
                    </span>
                    <span class="selection-check" :class="{ 'selection-check--visible': selectedFormats.includes('image') }">
                      <svg viewBox="0 0 24 24" fill="none">
                        <circle cx="12" cy="12" r="9" />
                        <path d="M8.667 12.333L10.667 14.333L15.333 9.667" />
                      </svg>
                    </span>
                  </div>
                  <strong>AI 图片生成</strong>
                  <span>商品图/场景图</span>
                </button>

                <button
                  type="button"
                  class="content-form-card content-form-card--video"
                  :class="{ 'content-form-card--active': selectedFormats.includes('video') }"
                  @click="toggleFormat('video')"
                >
                  <div class="content-form-card__top">
                    <span class="content-form-card__icon" aria-hidden="true">
                      <svg viewBox="0 0 32 32" fill="none">
                        <rect x="6.667" y="9.333" width="13.333" height="13.333" rx="2.667" />
                        <path d="M20 13.333L25.333 10.667V21.333L20 18.667V13.333Z" />
                      </svg>
                    </span>
                    <span class="selection-check" :class="{ 'selection-check--visible': selectedFormats.includes('video') }">
                      <svg viewBox="0 0 24 24" fill="none">
                        <circle cx="12" cy="12" r="9" />
                        <path d="M8.667 12.333L10.667 14.333L15.333 9.667" />
                      </svg>
                    </span>
                  </div>
                  <strong>视频片段</strong>
                  <span>产品展示视频</span>
                </button>

                <button
                  type="button"
                  class="content-form-card content-form-card--avatar"
                  :class="{ 'content-form-card--active': selectedFormats.includes('avatar') }"
                  @click="toggleFormat('avatar')"
                >
                  <div class="content-form-card__top">
                    <span class="content-form-card__icon" aria-hidden="true">
                      <svg viewBox="0 0 32 32" fill="none">
                        <path d="M16 17.333C19.682 17.333 22.667 14.348 22.667 10.667C22.667 6.985 19.682 4 16 4C12.318 4 9.333 6.985 9.333 10.667C9.333 14.348 12.318 17.333 16 17.333Z" />
                        <path d="M6.667 26.667C6.667 21.512 10.845 17.333 16 17.333C21.155 17.333 25.333 21.512 25.333 26.667" />
                      </svg>
                    </span>
                    <span class="selection-check" :class="{ 'selection-check--visible': selectedFormats.includes('avatar') }">
                      <svg viewBox="0 0 24 24" fill="none">
                        <circle cx="12" cy="12" r="9" />
                        <path d="M8.667 12.333L10.667 14.333L15.333 9.667" />
                      </svg>
                    </span>
                  </div>
                  <strong>数字人口播</strong>
                  <span>品牌讲解</span>
                </button>
              </div>
            </article>

            <article class="panel-card panel-card--plan">
              <div class="panel-card__header panel-card__header--between">
                <h2>智能组合方案</h2>
                <button type="button" class="ai-optimize-button" @click="handleOptimizePlan">
                  <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
                    <path d="M8 1.333L8.933 4.067L11.667 5L8.933 5.933L8 8.667L7.067 5.933L4.333 5L7.067 4.067L8 1.333Z" />
                    <path d="M12.667 8L13.2 9.467L14.667 10L13.2 10.533L12.667 12L12.133 10.533L10.667 10L12.133 9.467L12.667 8Z" />
                    <path d="M4 9.333L4.533 10.8L6 11.333L4.533 11.867L4 13.333L3.467 11.867L2 11.333L3.467 10.8L4 9.333Z" />
                  </svg>
                  <span>AI 优化</span>
                </button>
              </div>

              <div class="plan-step-list">
                <article v-for="step in planSteps" :key="step.index" class="plan-step-card">
                  <div class="plan-step-card__index">{{ step.index }}</div>

                  <div class="plan-step-card__body">
                    <div class="plan-step-card__title-row">
                      <strong>{{ step.title }}</strong>
                      <span>{{ step.duration }}</span>
                    </div>

                    <div class="plan-step-card__meta">
                      <span class="plan-tag">{{ step.tag }}</span>
                      <span :class="['plan-detail', { 'plan-detail--highlight': step.highlight }]">
                        {{ step.detail }}
                      </span>
                    </div>
                  </div>
                </article>
              </div>

              <div class="plan-summary">
                <span>总时长预估</span>
                <strong>{{ totalDuration }}</strong>
              </div>
            </article>
          </div>

          <aside class="preview-panel">
            <article class="preview-panel__card">
              <h2>内容预览</h2>

              <div class="preview-stage">
                <div class="preview-stage__overlay"></div>
                <button type="button" class="preview-play-button" aria-label="播放预览">
                  <svg viewBox="0 0 32 32" fill="none">
                    <path d="M11.333 7.933V24.067L23.333 16L11.333 7.933Z" />
                  </svg>
                </button>
              </div>

              <div class="preview-metrics">
                <div class="preview-metric">
                  <span>图片数量</span>
                  <strong :class="{ 'preview-metric__value--muted': previewImageCount === 0 }">
                    {{ previewImageCount }}张
                  </strong>
                </div>
                <div class="preview-metric">
                  <span>视频片段</span>
                  <strong :class="{ 'preview-metric__value--muted': previewVideoCount === 0 }">
                    {{ previewVideoCount }}段
                  </strong>
                </div>
                <div class="preview-metric">
                  <span>数字人</span>
                  <strong :class="{ 'preview-metric__value--muted': !hasAvatar }">
                    {{ hasAvatar ? '已选择' : '未选择' }}
                  </strong>
                </div>
              </div>

              <div class="preview-actions">
                <button type="button" class="preview-actions__primary" @click="handleGenerateAll">
                  <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
                    <path d="M10 2.5L11.167 5.833L14.5 7L11.167 8.167L10 11.5L8.833 8.167L5.5 7L8.833 5.833L10 2.5Z" />
                    <path d="M15.5 9.5L16.167 11.333L18 12L16.167 12.667L15.5 14.5L14.833 12.667L13 12L14.833 11.333L15.5 9.5Z" />
                    <path d="M5 11.5L5.667 13.333L7.5 14L5.667 14.667L5 16.5L4.333 14.667L2.5 14L4.333 13.333L5 11.5Z" />
                  </svg>
                  <span>一键生成全部内容</span>
                </button>

                <button type="button" class="preview-actions__secondary" @click="handleEnterTimeline">
                  <span>进入视频剪辑台</span>
                  <img :src="ctaArrowIcon" alt="" />
                </button>

                <button type="button" class="preview-actions__ghost" @click="handleSaveConfig">
                  保存配置
                </button>
              </div>
            </article>
          </aside>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { imageAPI } from '@/api/image'
import { videoAPI } from '@/api/video'
import bellIcon from '@/assets/product-entry/bell.svg'
import brandIcon from '@/assets/product-entry/brand-icon.svg'
import ctaArrowIcon from '@/assets/product-entry/arrow-right.svg'
import {
  buildEpisodeStagePath,
  buildProfessionalEditorPath,
  resolveEpisodeWorkflowContext,
  saveEpisodeWorkflowContext,
  type EpisodeWorkflowContext
} from '@/utils/episodeWorkflowContext'

type CreationMode = 'combo' | 'image' | 'video' | 'avatar'
type FormatKey = 'image' | 'video' | 'avatar'

interface CreationModeItem {
  key: CreationMode
  label: string
  description: string
}

interface PlanStep {
  index: number
  title: string
  duration: string
  tag: string
  detail: string
  highlight?: boolean
}

interface WorkspaceDraft {
  selectedMode?: CreationMode
  selectedFormats?: FormatKey[]
}

const WORKSPACE_STORAGE_PREFIX = 'drama:content-creation:workspace:'

const creationModes: CreationModeItem[] = [
  { key: 'combo', label: '组合创作', description: '自由组合多种内容形式' },
  { key: 'image', label: '图片生成', description: '商品图/场景图/广告图' },
  { key: 'video', label: '视频片段', description: '产品展示/场景视频' },
  { key: 'avatar', label: '数字人', description: '口播讲解/品牌表达' }
]

const planSteps: PlanStep[] = [
  {
    index: 1,
    title: '开场问候与品牌介绍',
    duration: '5s',
    tag: '数字人',
    detail: '数字人: Sarah',
    highlight: true
  },
  {
    index: 2,
    title: '产品展示与功能演示',
    duration: '20s',
    tag: '图片+视频',
    detail: '3张图 + 2段视频'
  },
  {
    index: 3,
    title: '使用场景展示',
    duration: '12s',
    tag: '视频片段',
    detail: '1段视频'
  },
  {
    index: 4,
    title: 'CTA与购买引导',
    duration: '8s',
    tag: '数字人',
    detail: '数字人: Sarah',
    highlight: true
  }
]

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const flowContext = ref<EpisodeWorkflowContext | null>(null)
const remoteImageCount = ref(0)
const remoteVideoCount = ref(0)

const selectedMode = ref<CreationMode>('combo')
const selectedFormats = ref<FormatKey[]>(['image', 'video', 'avatar'])

const episodeId = computed(() => String(route.params.id || route.query.id || 'draft'))
const workspaceStorageKey = computed(() => `${WORKSPACE_STORAGE_PREFIX}${episodeId.value}`)

const navItems = computed(() => [
  { label: '工作台', path: '/dramas', active: false, width: '66px' },
  { label: '商品录入', path: '/dramas/create', active: false, width: '80px' },
  { label: '合规分析', path: '/compliance', active: false, width: '80px' },
  {
    label: '脚本/分镜',
    path: flowContext.value ? buildEpisodeStagePath('script', flowContext.value) : '/workspace/script',
    active: false,
    width: '86px'
  },
  { label: '内容创作', path: '/workspace/content', active: true, width: '80px' },
  {
    label: '视频剪辑',
    path: flowContext.value ? buildEpisodeStagePath('timeline', flowContext.value) : '/workspace/timeline',
    active: false,
    width: '80px'
  },
  { label: '数据分析', path: '/analytics', active: false, width: '80px' }
])

const previewImageCount = computed(() => {
  if (remoteImageCount.value > 0) return remoteImageCount.value
  return selectedFormats.value.includes('image') ? 3 : 0
})
const previewVideoCount = computed(() => {
  if (remoteVideoCount.value > 0) return remoteVideoCount.value
  return selectedFormats.value.includes('video') ? 2 : 0
})
const hasAvatar = computed(() => selectedFormats.value.includes('avatar'))
const totalDuration = computed(() => '45秒')

const persistWorkspace = () => {
  if (typeof window === 'undefined') return

  const payload: WorkspaceDraft = {
    selectedMode: selectedMode.value,
    selectedFormats: [...selectedFormats.value]
  }

  window.sessionStorage.setItem(workspaceStorageKey.value, JSON.stringify(payload))
}

const restoreWorkspace = () => {
  if (typeof window === 'undefined') return

  const raw = window.sessionStorage.getItem(workspaceStorageKey.value)
  if (!raw) return

  try {
    const parsed = JSON.parse(raw) as WorkspaceDraft
    if (parsed.selectedMode) {
      selectedMode.value = parsed.selectedMode
    }
    if (Array.isArray(parsed.selectedFormats) && parsed.selectedFormats.length > 0) {
      selectedFormats.value = parsed.selectedFormats.filter((item): item is FormatKey =>
        ['image', 'video', 'avatar'].includes(item)
      )
    }
  } catch {
    window.sessionStorage.removeItem(workspaceStorageKey.value)
  }
}

const handleNavClick = (path: string) => {
  if (!path) return
  router.push(path)
}

const handleSelectMode = (mode: CreationMode) => {
  selectedMode.value = mode

  if (mode === 'combo') {
    selectedFormats.value = ['image', 'video', 'avatar']
  } else if (mode === 'image') {
    selectedFormats.value = ['image']
  } else if (mode === 'video') {
    selectedFormats.value = ['video']
  } else {
    selectedFormats.value = ['avatar']
  }

  persistWorkspace()
}

const toggleFormat = (format: FormatKey) => {
  const exists = selectedFormats.value.includes(format)

  if (exists && selectedFormats.value.length === 1) {
    ElMessage.info('至少保留一种内容形式')
    return
  }

  selectedFormats.value = exists
    ? selectedFormats.value.filter((item) => item !== format)
    : [...selectedFormats.value, format]

  if (selectedFormats.value.length === 3) {
    selectedMode.value = 'combo'
  } else if (selectedFormats.value.length === 1) {
    selectedMode.value =
      selectedFormats.value[0] === 'image'
        ? 'image'
        : selectedFormats.value[0] === 'video'
          ? 'video'
          : 'avatar'
  }

  persistWorkspace()
}

const handleOptimizePlan = () => {
  ElMessage.success('AI 已按当前组合方案完成镜头节奏优化')
}

const handleGenerateAll = () => {
  persistWorkspace()

  if (flowContext.value) {
    router.push(buildProfessionalEditorPath(flowContext.value))
    ElMessage.success('已进入专业制作台，可继续生成图片、视频与数字人素材')
    return
  }

  ElMessage.info('尚未绑定剧集上下文，请先从脚本与分镜阶段进入')
}

const handleEnterTimeline = () => {
  if (!flowContext.value) {
    ElMessage.info('尚未绑定剧集上下文，请先从脚本与分镜阶段进入')
    return
  }

  router.push(buildEpisodeStagePath('timeline', flowContext.value))
}

const handleSaveConfig = () => {
  persistWorkspace()
  ElMessage.success('内容创作配置已保存')
}

const loadRemoteCounts = async () => {
  if (!flowContext.value?.dramaId) return

  try {
    const [images, videos] = await Promise.all([
      imageAPI.listImages({
        drama_id: flowContext.value.dramaId,
        page: 1,
        page_size: 100
      }),
      videoAPI.listVideos({
        drama_id: flowContext.value.dramaId,
        page: 1,
        page_size: 100
      })
    ])

    remoteImageCount.value = images.items?.filter((item) => item.status === 'completed').length || 0
    remoteVideoCount.value = videos.items?.filter((item) => item.status === 'completed').length || 0
  } catch (error) {
    remoteImageCount.value = 0
    remoteVideoCount.value = 0
  }
}

onMounted(async () => {
  const resolved = resolveEpisodeWorkflowContext({
    episodeId: route.params.id || route.query.id,
    dramaId: route.query.dramaId,
    episodeNumber: route.query.episodeNumber
  })

  if (resolved) {
    flowContext.value = resolved
    saveEpisodeWorkflowContext(resolved)
    await loadRemoteCounts()
  }

  restoreWorkspace()
  persistWorkspace()
})
</script>

<style scoped>
.page-container.content-creation-page {
  --creation-heading-font: 'Urbanist', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  --creation-body-font: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  min-height: var(--app-vh, 100vh);
  padding: 0 !important;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%) !important;
  color: #0a2463;
  font-family: var(--creation-body-font);
}

.creation-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
  height: 72px;
  backdrop-filter: blur(18px);
  background: rgba(255, 255, 255, 0.82);
  border-bottom: 1px solid rgba(226, 232, 240, 0.86);
}

.creation-header__inner {
  width: min(100%, 1384px);
  height: 100%;
  margin: 0 auto;
  padding: 0 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.creation-header__left,
.creation-header__right {
  display: flex;
  align-items: center;
}

.creation-header__left {
  gap: 28px;
}

.creation-header__right {
  gap: 12px;
}

.brand-link {
  border: none;
  background: transparent;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0;
  cursor: pointer;
}

.brand-link__mark {
  width: 34px;
  height: 34px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.18), rgba(124, 58, 237, 0.22));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.brand-link__mark img {
  width: 21px;
  height: 21px;
}

.brand-link__name {
  font-family: var(--creation-heading-font);
  font-size: 18px;
  font-weight: 700;
  color: #0a2463;
}

.creation-nav {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.9);
  border: 1px solid rgba(226, 232, 240, 0.85);
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.04);
}

.creation-nav__item {
  height: 36px;
  border: none;
  border-radius: 14px;
  background: transparent;
  color: #62748e;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.22s ease;
}

.creation-nav__item--active {
  color: #ffffff;
  font-weight: 700;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  box-shadow: 0 10px 24px rgba(99, 102, 241, 0.24);
}

.creation-nav__item:not(.creation-nav__item--active):hover {
  color: #0a2463;
  background: rgba(226, 232, 240, 0.65);
}

.header-icon-button {
  border: 1px solid rgba(226, 232, 240, 0.9);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
}

.header-icon-button {
  position: relative;
  width: 44px;
  height: 44px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.header-icon-button img {
  width: 20px;
  height: 20px;
}

.header-icon-button__dot {
  position: absolute;
  top: 11px;
  right: 11px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: linear-gradient(135deg, #fb7185, #f97316);
  box-shadow: 0 0 0 4px rgba(251, 113, 133, 0.12);
}

.content-creation-main {
  padding: 104px 0 72px;
}

.content-creation-shell {
  width: min(100%, 1384px);
  margin: 0 auto;
  padding: 0 32px;
}

.creation-hero {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 32px;
}

.creation-hero h1 {
  margin: 0;
  font-family: var(--creation-heading-font);
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #0a2463;
}

.creation-hero p {
  margin: 0;
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.creation-mode-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.creation-mode-card {
  height: 140px;
  border-radius: 16px;
  border: 2px solid #e2e8f0;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 24px;
  cursor: pointer;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.creation-mode-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.08);
}

.creation-mode-card strong {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.creation-mode-card span:last-child {
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
  text-align: center;
}

.creation-mode-card__icon {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.creation-mode-card__icon svg {
  width: 32px;
  height: 32px;
  stroke: #94a3b8;
  stroke-width: 1.85;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.creation-mode-card--active {
  box-shadow: 0 10px 15px rgba(15, 23, 42, 0.1), 0 4px 6px rgba(15, 23, 42, 0.1);
}

.creation-mode-card--combo.creation-mode-card--active {
  border-color: #06b6d4;
  background: linear-gradient(149.82deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
}

.creation-mode-card--combo.creation-mode-card--active .creation-mode-card__icon svg {
  stroke: #06b6d4;
}

.creation-mode-card--image.creation-mode-card--active {
  border-color: #06b6d4;
  background: linear-gradient(149.82deg, rgba(6, 182, 212, 0.08) 0%, rgba(59, 130, 246, 0.06) 100%);
}

.creation-mode-card--image.creation-mode-card--active .creation-mode-card__icon svg {
  stroke: #06b6d4;
}

.creation-mode-card--video.creation-mode-card--active {
  border-color: #7c3aed;
  background: linear-gradient(149.82deg, rgba(124, 58, 237, 0.08) 0%, rgba(236, 72, 153, 0.06) 100%);
}

.creation-mode-card--video.creation-mode-card--active .creation-mode-card__icon svg {
  stroke: #7c3aed;
}

.creation-mode-card--avatar.creation-mode-card--active {
  border-color: #f97316;
  background: linear-gradient(149.82deg, rgba(249, 115, 22, 0.08) 0%, rgba(251, 146, 60, 0.06) 100%);
}

.creation-mode-card--avatar.creation-mode-card--active .creation-mode-card__icon svg {
  stroke: #f97316;
}

.workspace-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 313px;
  gap: 24px;
  align-items: start;
}

.workspace-main {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.panel-card {
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  padding: 25px 25px 24px;
}

.panel-card__header {
  margin-bottom: 24px;
}

.panel-card__header--between {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.panel-card__header h2,
.preview-panel__card h2 {
  margin: 0;
  font-family: var(--creation-heading-font);
  font-size: 18px;
  line-height: 28px;
  font-weight: 600;
  color: #0a2463;
}

.content-form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.content-form-card {
  height: 140px;
  border-radius: 16px;
  border: 2px solid #e2e8f0;
  background: #ffffff;
  padding: 24px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  cursor: pointer;
  transition: transform 0.22s ease, box-shadow 0.22s ease, opacity 0.22s ease;
}

.content-form-card:hover {
  transform: translateY(-2px);
}

.content-form-card__top {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.content-form-card__icon {
  width: 32px;
  height: 32px;
  display: inline-flex;
}

.content-form-card__icon svg {
  width: 32px;
  height: 32px;
  stroke-width: 1.85;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.content-form-card strong {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.content-form-card span:last-child {
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.content-form-card--image {
  border-color: #06b6d4;
  background: linear-gradient(144.15deg, rgba(6, 182, 212, 0.05) 0%, rgba(124, 58, 237, 0.05) 100%);
  box-shadow: 0 10px 15px rgba(15, 23, 42, 0.1), 0 4px 6px rgba(15, 23, 42, 0.1);
}

.content-form-card--image .content-form-card__icon svg {
  stroke: #06b6d4;
}

.content-form-card--video {
  border-color: #7c3aed;
  background: linear-gradient(144.15deg, rgba(124, 58, 237, 0.05) 0%, rgba(236, 72, 153, 0.05) 100%);
  box-shadow: 0 10px 15px rgba(15, 23, 42, 0.1), 0 4px 6px rgba(15, 23, 42, 0.1);
}

.content-form-card--video .content-form-card__icon svg {
  stroke: #7c3aed;
}

.content-form-card--avatar {
  border-color: #f97316;
  background: linear-gradient(144.15deg, rgba(249, 115, 22, 0.05) 0%, rgba(251, 146, 60, 0.05) 100%);
  box-shadow: 0 10px 15px rgba(15, 23, 42, 0.1), 0 4px 6px rgba(15, 23, 42, 0.1);
}

.content-form-card--avatar .content-form-card__icon svg {
  stroke: #f97316;
}

.content-form-card:not(.content-form-card--active) {
  opacity: 0.55;
}

.content-form-card:not(.content-form-card--active):hover {
  opacity: 0.82;
}

.selection-check {
  width: 24px;
  height: 24px;
  opacity: 0;
  transform: scale(0.92);
  transition: all 0.18s ease;
}

.selection-check svg {
  width: 24px;
  height: 24px;
  stroke: #10b981;
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.selection-check--visible {
  opacity: 1;
  transform: scale(1);
}

.ai-optimize-button {
  height: 36px;
  padding: 0 16px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  box-shadow: 0 10px 24px rgba(99, 102, 241, 0.22);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.ai-optimize-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 30px rgba(99, 102, 241, 0.28);
}

.ai-optimize-button svg {
  width: 16px;
  height: 16px;
  stroke: #ffffff;
  fill: none;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.ai-optimize-button span {
  color: #ffffff;
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
}

.plan-step-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 16px;
}

.plan-step-card {
  min-height: 96px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: linear-gradient(90deg, #f8fafc 0%, #ffffff 100%);
  padding: 17px;
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.plan-step-card__index {
  width: 48px;
  height: 48px;
  border-radius: 16px;
  flex: 0 0 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #06b6d4 0%, #7c3aed 100%);
  color: #ffffff;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
}

.plan-step-card__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.plan-step-card__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.plan-step-card__title-row strong {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.plan-step-card__title-row span {
  font-size: 14px;
  line-height: 20px;
  color: #62748e;
  white-space: nowrap;
}

.plan-step-card__meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.plan-tag {
  height: 30px;
  padding: 0 12px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  display: inline-flex;
  align-items: center;
  font-size: 14px;
  line-height: 20px;
  color: #45556c;
}

.plan-detail {
  font-size: 14px;
  line-height: 20px;
  color: #62748e;
}

.plan-detail--highlight {
  color: #06b6d4;
}

.plan-summary {
  min-height: 60px;
  border-radius: 16px;
  background: linear-gradient(90deg, #eff6ff 0%, #ecfeff 100%);
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.plan-summary span {
  font-size: 14px;
  line-height: 20px;
  color: #314158;
}

.plan-summary strong {
  font-family: var(--creation-heading-font);
  font-size: 18px;
  line-height: 28px;
  font-weight: 700;
  color: #0a2463;
}

.preview-panel {
  position: sticky;
  top: 104px;
}

.preview-panel__card {
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  padding: 24px;
}

.preview-stage {
  position: relative;
  margin-top: 16px;
  height: 149px;
  border-radius: 16px;
  overflow: hidden;
  background: linear-gradient(150.64deg, #0f172b 0%, #1d293d 100%);
}

.preview-stage::before {
  content: '';
  position: absolute;
  inset: 0;
  opacity: 0.45;
  background:
    radial-gradient(circle at 0 0, rgba(51, 65, 85, 0.56), transparent 36%),
    linear-gradient(135deg, rgba(59, 130, 246, 0.08), transparent 42%),
    repeating-linear-gradient(
      135deg,
      rgba(148, 163, 184, 0.12) 0,
      rgba(148, 163, 184, 0.12) 1px,
      transparent 1px,
      transparent 10px
    );
  transform: translate(-26px, -24px);
}

.preview-stage__overlay {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 48%, rgba(255, 255, 255, 0.06), transparent 40%);
}

.preview-play-button {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 64px;
  height: 64px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 22px 34px rgba(249, 115, 22, 0.24);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.preview-play-button:hover {
  transform: translate(-50%, -50%) scale(1.04);
  box-shadow: 0 28px 40px rgba(249, 115, 22, 0.28);
}

.preview-play-button svg {
  width: 32px;
  height: 32px;
  stroke: #ffffff;
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.preview-metrics {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 24px;
}

.preview-metric {
  min-height: 48px;
  border-radius: 12px;
  background: #f8fafc;
  padding: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.preview-metric span {
  font-size: 14px;
  line-height: 20px;
  color: #45556c;
}

.preview-metric strong {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.preview-metric__value--muted {
  color: #94a3b8;
}

.preview-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 24px;
}

.preview-actions__primary,
.preview-actions__secondary,
.preview-actions__ghost {
  width: 100%;
  border: none;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.preview-actions__primary {
  height: 48px;
  border-radius: 16px;
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow: 0 12px 26px rgba(99, 102, 241, 0.2);
}

.preview-actions__primary:hover,
.preview-actions__secondary:hover,
.preview-actions__ghost:hover {
  transform: translateY(-1px);
}

.preview-actions__primary svg {
  width: 20px;
  height: 20px;
  stroke: #ffffff;
  fill: none;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.preview-actions__primary span,
.preview-actions__secondary span,
.preview-actions__ghost {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
}

.preview-actions__primary span {
  color: #ffffff;
}

.preview-actions__secondary {
  height: 48px;
  border-radius: 16px;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  box-shadow: 0 12px 24px rgba(249, 115, 22, 0.18);
}

.preview-actions__secondary span {
  color: #ffffff;
}

.preview-actions__secondary img {
  width: 18px;
  height: 18px;
  filter: brightness(0) invert(1);
}

.preview-actions__ghost {
  height: 52px;
  border-radius: 16px;
  border: 2px solid #e2e8f0;
  background: #ffffff;
  color: #0a2463;
}

@media (max-width: 1180px) {
  .creation-header__inner,
  .content-creation-shell {
    padding-left: 20px;
    padding-right: 20px;
  }

  .creation-nav {
    display: none;
  }

  .workspace-grid {
    grid-template-columns: 1fr;
  }

  .preview-panel {
    position: static;
  }
}

@media (max-width: 920px) {
  .creation-mode-row,
  .content-form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .creation-header__inner {
    gap: 12px;
  }

  .creation-header__right {
    display: none;
  }

  .content-creation-main {
    padding-top: 92px;
    padding-bottom: 48px;
  }

  .creation-hero h1 {
    font-size: 26px;
    line-height: 32px;
  }

  .creation-mode-row,
  .content-form-grid {
    grid-template-columns: 1fr;
  }

  .panel-card,
  .preview-panel__card {
    padding: 20px;
  }

  .panel-card__header--between {
    align-items: flex-start;
    flex-direction: column;
  }

  .plan-step-card__title-row,
  .plan-step-card__meta {
    flex-direction: column;
    align-items: flex-start;
  }

  .plan-summary {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
