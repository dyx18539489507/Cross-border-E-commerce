<template>
  <div class="page-container video-editor-page">
    <header class="video-editor-header">
      <div class="video-editor-header__inner">
        <div class="video-editor-header__left">
          <button type="button" class="brand-link" @click="router.push('/')">
            <span class="brand-link__mark">
              <img :src="brandIcon" alt="" />
            </span>
            <span class="brand-link__name">{{ t('app.name') }}</span>
          </button>

          <nav class="video-editor-nav" aria-label="主导航">
            <button
              v-for="item in navItems"
              :key="item.label"
              type="button"
              class="video-editor-nav__item"
              :class="{ 'video-editor-nav__item--active': item.active }"
              :style="{ width: item.width }"
              @click="handleNavClick(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="video-editor-header__right">
          <button type="button" class="header-icon-button" aria-label="通知">
            <img :src="bellIcon" alt="" />
            <span class="header-icon-button__dot"></span>
          </button>
        </div>
      </div>
    </header>

    <main class="video-editor-main">
      <div class="video-editor-shell">
        <section class="video-editor-hero">
          <h1>视频剪辑与成片输出</h1>
          <p>智能拼接图片、视频、数字人内容，AI优化剪辑，一键输出成片</p>
        </section>

        <section v-if="hasLiveTimeline" class="video-editor-live" v-loading="liveLoading">
          <div class="video-editor-live__header">
            <div class="video-editor-live__copy">
              <span class="video-editor-live__eyebrow">已接入真实剪辑流程</span>
              <h2>第{{ flowContext?.episodeNumber || 1 }}集时间线</h2>
              <p>
                {{
                  liveTimelineError
                    || `已加载 ${liveStoryboards.length} 个分镜与 ${videoAssets.length} 条视频素材，可直接拖入时间线完成合成。`
                }}
              </p>
            </div>

            <div class="video-editor-live__actions">
              <button
                type="button"
                class="video-editor-live__button video-editor-live__button--ghost"
                @click="goBackToGeneration"
              >
                返回内容创作
              </button>

              <button
                type="button"
                class="video-editor-live__button video-editor-live__button--primary"
                @click="goToProfessionalEditor"
              >
                进入专业制作台
              </button>
            </div>
          </div>

          <div class="video-editor-live__panel">
            <VideoTimelineEditor
              v-if="flowContext"
              :scenes="liveStoryboards"
              :episode-id="flowContext.episodeId"
              :drama-id="flowContext.dramaId"
              :assets="videoAssets"
              @asset-deleted="loadLiveTimelineData"
              @merge-completed="handleLiveMergeCompleted"
            />
          </div>
        </section>

        <section v-else class="video-editor-grid">
          <aside class="video-editor-sidebar">
            <article class="editor-card editor-card--template">
              <div class="editor-card__title">
                <span class="editor-card__title-icon editor-card__title-icon--violet">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M6 6L18 18" />
                    <path d="M18 6L6 18" />
                    <path d="M9 3L10.5 6.5L14 8L10.5 9.5L9 13L7.5 9.5L4 8L7.5 6.5L9 3Z" />
                    <path d="M17 11L18 13.5L20.5 14.5L18 15.5L17 18L16 15.5L13.5 14.5L16 13.5L17 11Z" />
                  </svg>
                </span>
                <h2>模板风格</h2>
              </div>

              <div class="template-list">
                <button
                  v-for="template in templates"
                  :key="template.id"
                  type="button"
                  class="template-card"
                  :class="{ 'template-card--active': selectedTemplateId === template.id }"
                  @click="handleSelectTemplate(template.id)"
                >
                  <span class="template-card__emoji">{{ template.emoji }}</span>
                  <div class="template-card__copy">
                    <strong>{{ template.title }}</strong>
                    <span>{{ template.subtitle }}</span>
                  </div>
                </button>
              </div>
            </article>

            <article class="editor-card editor-card--music">
              <div class="editor-card__title">
                <span class="editor-card__title-icon editor-card__title-icon--orange">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M5 15V9" />
                    <path d="M9 18V6" />
                    <path d="M13 14V10" />
                    <path d="M17 16V8" />
                    <path d="M20 13C20 15.761 17.761 18 15 18H14" />
                  </svg>
                </span>
                <h2>背景音乐</h2>
              </div>

              <div class="music-list">
                <button
                  v-for="track in musicTracks"
                  :key="track.id"
                  type="button"
                  class="music-card"
                  :class="{ 'music-card--active': selectedMusicId === track.id }"
                  @click="handleSelectMusic(track.id)"
                >
                  <div class="music-card__top">
                    <strong>{{ track.title }}</strong>
                    <span class="music-card__play">
                      <svg viewBox="0 0 20 20" fill="none">
                        <path d="M7 5.667V14.333L13.667 10L7 5.667Z" />
                      </svg>
                    </span>
                  </div>
                  <div class="music-card__bottom">
                    <span>{{ track.subtitle }}</span>
                    <span>{{ track.duration }}</span>
                  </div>
                </button>
              </div>
            </article>

            <article class="editor-card editor-card--materials">
              <div class="editor-card__title">
                <span class="editor-card__title-icon editor-card__title-icon--cyan">
                  <svg viewBox="0 0 24 24" fill="none">
                    <rect x="4" y="5" width="16" height="14" rx="2.5" />
                    <circle cx="9" cy="10" r="1.5" />
                    <path d="M6.5 16L10.5 12L13.5 15L16.5 12L18 13.5" />
                  </svg>
                </span>
                <h2>素材库</h2>
              </div>

              <div class="material-grid">
                <button
                  v-for="material in materials"
                  :key="material.id"
                  type="button"
                  class="material-tile"
                  :class="{ 'material-tile--active': selectedMaterialId === material.id }"
                  @click="handleSelectMaterial(material.id)"
                >
                  <svg viewBox="0 0 24 24" fill="none">
                    <rect x="4" y="5" width="16" height="14" rx="2.5" />
                    <circle cx="9" cy="10" r="1.5" />
                    <path d="M6.5 16L10.5 12L13.5 15L16.5 12L18 13.5" />
                  </svg>
                </button>
              </div>

              <button type="button" class="upload-material-button" @click="handleUploadMaterial">
                + 上传素材
              </button>
            </article>

            <article class="editor-card editor-card--tools">
              <div class="editor-card__title">
                <span class="editor-card__title-icon editor-card__title-icon--purple">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M12 3L13.8 7.2L18 9L13.8 10.8L12 15L10.2 10.8L6 9L10.2 7.2L12 3Z" />
                    <path d="M18 14L18.9 16.1L21 17L18.9 17.9L18 20L17.1 17.9L15 17L17.1 16.1L18 14Z" />
                    <path d="M6 14L6.9 16.1L9 17L6.9 17.9L6 20L5.1 17.9L3 17L5.1 16.1L6 14Z" />
                  </svg>
                </span>
                <h2>AI 智能工具</h2>
              </div>

              <div class="tool-list">
                <button
                  v-for="tool in aiTools"
                  :key="tool.id"
                  type="button"
                  class="tool-button"
                  :class="{ 'tool-button--active': selectedToolId === tool.id }"
                  @click="handleSelectTool(tool.id)"
                >
                  {{ tool.label }}
                </button>
              </div>
            </article>
          </aside>

          <div class="video-editor-workspace">
            <article class="workspace-card workspace-card--preview">
              <div class="preview-stage">
                <div class="preview-stage__surface"></div>

                <button type="button" class="preview-stage__play" @click="togglePlayback">
                  <svg viewBox="0 0 40 40" fill="none">
                    <path d="M13 10.5V29.5L29 20L13 10.5Z" />
                  </svg>
                </button>

                <div class="preview-stage__footer">
                  <span class="preview-stage__time">{{ currentPreviewTime }}</span>
                  <button type="button" class="preview-stage__volume" @click="toggleMute" aria-label="切换静音">
                    <svg v-if="!isMuted" viewBox="0 0 16 16" fill="none">
                      <path d="M2.667 6H5.333L8.667 3.333V12.667L5.333 10H2.667V6Z" />
                      <path d="M11.333 5.333C12.438 6.438 12.438 9.562 11.333 10.667" />
                    </svg>
                    <svg v-else viewBox="0 0 16 16" fill="none">
                      <path d="M2.667 6H5.333L8.667 3.333V12.667L5.333 10H2.667V6Z" />
                      <path d="M10.667 5.333L13.333 10.667" />
                      <path d="M13.333 5.333L10.667 10.667" />
                    </svg>
                  </button>
                </div>
              </div>

              <div class="preview-controls">
                <button type="button" class="preview-controls__button" @click="handleTrim">
                  <svg viewBox="0 0 20 20" fill="none">
                    <path d="M8 3V17" />
                    <path d="M12 3V17" />
                    <path d="M7 8L12 13" />
                    <path d="M7 13L12 8" />
                  </svg>
                </button>

                <button type="button" class="preview-controls__button" @click="togglePause">
                  <svg viewBox="0 0 20 20" fill="none">
                    <rect x="6.5" y="5" width="2.5" height="10" rx="1" />
                    <rect x="11" y="5" width="2.5" height="10" rx="1" />
                  </svg>
                </button>

                <button
                  type="button"
                  class="preview-controls__button preview-controls__button--play"
                  :class="{ 'preview-controls__button--active': isPlaying }"
                  @click="togglePlayback"
                >
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M8.5 6.5V17.5L17.5 12L8.5 6.5Z" />
                  </svg>
                </button>
              </div>

              <div class="preview-progress">
                <div class="preview-progress__track"></div>
                <div class="preview-progress__fill" :style="{ width: `${previewProgress}%` }"></div>
                <div class="preview-progress__thumb" :style="{ left: `${previewProgress}%` }"></div>
              </div>
            </article>

            <article class="workspace-card workspace-card--timeline">
              <div class="workspace-card__title-row">
                <h2>时间轴</h2>
                <span>总时长: 45秒</span>
              </div>

              <div class="timeline-list">
                <div
                  v-for="segment in timelineSegments"
                  :key="segment.id"
                  class="timeline-row"
                >
                  <span class="timeline-row__duration">{{ segment.duration }}</span>
                  <button
                    type="button"
                    class="timeline-row__segment"
                    :class="`timeline-row__segment--${segment.tone}`"
                    @click="handleSelectSegment(segment.id)"
                  >
                    {{ segment.title }}
                  </button>
                </div>
              </div>

              <button type="button" class="timeline-add-button" @click="handleAddSegment">
                + 添加片段
              </button>
            </article>

            <article class="workspace-card workspace-card--export">
              <h2>导出设置</h2>

              <div class="export-grid">
                <label class="export-field">
                  <span>分辨率</span>
                  <button type="button" class="export-select" @click="cycleResolution">
                    {{ selectedResolution }}
                  </button>
                </label>

                <label class="export-field">
                  <span>格式</span>
                  <button type="button" class="export-select" @click="cycleFormat">
                    {{ selectedFormat }}
                  </button>
                </label>

                <label class="export-field">
                  <span>质量</span>
                  <button type="button" class="export-select" @click="cycleQuality">
                    {{ selectedQuality }}
                  </button>
                </label>
              </div>
            </article>

            <section class="workspace-actions">
              <button type="button" class="workspace-actions__primary" @click="handleExportVideo">
                <svg viewBox="0 0 20 20" fill="none">
                  <path d="M10 3.333V11.667" />
                  <path d="M6.667 8.333L10 11.667L13.333 8.333" />
                  <path d="M4 14.167V15.833H16V14.167" />
                </svg>
                <span>导出视频</span>
              </button>

              <button type="button" class="workspace-actions__secondary" @click="handlePublish">
                <span>发布并查看数据</span>
                <img :src="ctaArrowIcon" alt="" />
              </button>

              <button type="button" class="workspace-actions__ghost" @click="handleSaveProject">
                保存项目
              </button>
            </section>
          </div>
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
import { assetAPI } from '@/api/asset'
import { dramaAPI } from '@/api/drama'
import bellIcon from '@/assets/product-entry/bell.svg'
import brandIcon from '@/assets/product-entry/brand-icon.svg'
import ctaArrowIcon from '@/assets/product-entry/arrow-right.svg'
import VideoTimelineEditor from '@/components/editor/VideoTimelineEditor.vue'
import type { Asset } from '@/types/asset'
import type { Storyboard } from '@/types/drama'
import {
  buildEpisodeStagePath,
  buildProfessionalEditorPath,
  resolveEpisodeWorkflowContext,
  saveEpisodeWorkflowContext,
  type EpisodeWorkflowContext
} from '@/utils/episodeWorkflowContext'

type TemplateId = 'minimal' | 'dynamic' | 'lifestyle' | 'future'
type MusicId = 'corporate' | 'innovation' | 'jazz'
type MaterialId = 'm1' | 'm2' | 'm3' | 'm4'
type ToolId = 'palette' | 'subtitle' | 'transition' | 'balance'
type SegmentId = 'intro' | 'main' | 'showcase' | 'cta'

interface TemplateItem {
  id: TemplateId
  emoji: string
  title: string
  subtitle: string
}

interface MusicTrack {
  id: MusicId
  title: string
  subtitle: string
  duration: string
}

interface MaterialTile {
  id: MaterialId
}

interface AiTool {
  id: ToolId
  label: string
}

interface TimelineSegment {
  id: SegmentId
  title: string
  duration: string
  tone: 'blue' | 'purple' | 'orange' | 'green'
}

interface WorkspaceDraft {
  selectedTemplateId?: TemplateId | ''
  selectedMusicId?: MusicId | ''
  selectedMaterialId?: MaterialId | ''
  selectedToolId?: ToolId | ''
  selectedSegmentId?: SegmentId
  selectedResolution?: string
  selectedFormat?: string
  selectedQuality?: string
  previewProgress?: number
  isMuted?: boolean
  isPlaying?: boolean
}

const WORKSPACE_STORAGE_PREFIX = 'drama:video-editor:workspace:'

const templates: TemplateItem[] = [
  { id: 'minimal', emoji: '📱', title: '极简风格', subtitle: '简洁、专业' },
  { id: 'dynamic', emoji: '⚡', title: '动感炫酷', subtitle: '活力、时尚' },
  { id: 'lifestyle', emoji: '🏠', title: '温馨生活', subtitle: '温暖、亲和' },
  { id: 'future', emoji: '🚀', title: '科技未来', subtitle: '科技、创新' }
]

const musicTracks: MusicTrack[] = [
  { id: 'corporate', title: 'Upbeat Corporate', subtitle: '积极、专业', duration: '2:30' },
  { id: 'innovation', title: 'Tech Innovation', subtitle: '科技、未来', duration: '2:45' },
  { id: 'jazz', title: 'Smooth Jazz', subtitle: '轻松、优雅', duration: '3:00' }
]

const materials: MaterialTile[] = [{ id: 'm1' }, { id: 'm2' }, { id: 'm3' }, { id: 'm4' }]

const aiTools: AiTool[] = [
  { id: 'palette', label: '🎨 智能配色' },
  { id: 'subtitle', label: '✨ 自动字幕' },
  { id: 'transition', label: '🎬 场景转场' },
  { id: 'balance', label: '🔊 音量平衡' }
]

const timelineSegments: TimelineSegment[] = [
  { id: 'intro', title: '开场', duration: '3s', tone: 'blue' },
  { id: 'main', title: '主要内容', duration: '25s', tone: 'purple' },
  { id: 'showcase', title: '产品展示', duration: '12s', tone: 'orange' },
  { id: 'cta', title: 'CTA', duration: '5s', tone: 'green' }
]

const resolutionOptions = ['1080P', '2K', '4K']
const formatOptions = ['MP4', 'MOV', 'WEBM']
const qualityOptions = ['高', '超清', '标准']

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const flowContext = ref<EpisodeWorkflowContext | null>(null)
const liveLoading = ref(false)
const liveTimelineError = ref('')
const liveStoryboards = ref<Storyboard[]>([])
const videoAssets = ref<Asset[]>([])

const episodeId = computed(() => String(route.params.id || 'draft'))
const workspaceStorageKey = computed(() => `${WORKSPACE_STORAGE_PREFIX}${episodeId.value}`)

const selectedTemplateId = ref<TemplateId | ''>('')
const selectedMusicId = ref<MusicId | ''>('')
const selectedMaterialId = ref<MaterialId | ''>('')
const selectedToolId = ref<ToolId | ''>('')
const selectedSegmentId = ref<SegmentId>('main')
const selectedResolution = ref('1080P')
const selectedFormat = ref('MP4')
const selectedQuality = ref('高')
const previewProgress = ref(33.33)
const isMuted = ref(false)
const isPlaying = ref(true)
const hasLiveTimeline = computed(
  () => Boolean(flowContext.value?.dramaId && flowContext.value?.episodeId && flowContext.value?.episodeNumber)
)

const navItems = computed(() => [
  { label: '工作台', path: '/dramas', active: false, width: '66px' },
  { label: '商品录入', path: '/dramas/create', active: false, width: '80px' },
  { label: '合规分析', path: '/compliance', active: false, width: '80px' },
  {
    label: '脚本/分镜',
    path: flowContext.value ? buildEpisodeStagePath('script', flowContext.value) : '/workspace/script',
    active: false,
    width: '80px'
  },
  {
    label: '内容创作',
    path: flowContext.value ? buildEpisodeStagePath('generation', flowContext.value) : '/workspace/content',
    active: false,
    width: '80px'
  },
  { label: '视频剪辑', path: '/workspace/timeline', active: true, width: '80px' },
  { label: '数据分析', path: '/analytics', active: false, width: '80px' }
])

const currentPreviewTime = computed(() => {
  const totalSeconds = 45
  const currentSeconds = Math.round((previewProgress.value / 100) * totalSeconds)
  return `00:${String(currentSeconds).padStart(2, '0')} / 00:45`
})

const persistWorkspace = () => {
  if (typeof window === 'undefined') return

  const payload: WorkspaceDraft = {
    selectedTemplateId: selectedTemplateId.value,
    selectedMusicId: selectedMusicId.value,
    selectedMaterialId: selectedMaterialId.value,
    selectedToolId: selectedToolId.value,
    selectedSegmentId: selectedSegmentId.value,
    selectedResolution: selectedResolution.value,
    selectedFormat: selectedFormat.value,
    selectedQuality: selectedQuality.value,
    previewProgress: previewProgress.value,
    isMuted: isMuted.value,
    isPlaying: isPlaying.value
  }

  window.sessionStorage.setItem(workspaceStorageKey.value, JSON.stringify(payload))
}

const restoreWorkspace = () => {
  if (typeof window === 'undefined') return

  const raw = window.sessionStorage.getItem(workspaceStorageKey.value)
  if (!raw) return

  try {
    const parsed = JSON.parse(raw) as WorkspaceDraft
    selectedTemplateId.value = (parsed.selectedTemplateId || '') as TemplateId | ''
    selectedMusicId.value = (parsed.selectedMusicId || '') as MusicId | ''
    selectedMaterialId.value = (parsed.selectedMaterialId || '') as MaterialId | ''
    selectedToolId.value = (parsed.selectedToolId || '') as ToolId | ''
    selectedSegmentId.value = (parsed.selectedSegmentId || 'main') as SegmentId
    selectedResolution.value = parsed.selectedResolution || '1080P'
    selectedFormat.value = parsed.selectedFormat || 'MP4'
    selectedQuality.value = parsed.selectedQuality || '高'
    previewProgress.value = typeof parsed.previewProgress === 'number' ? parsed.previewProgress : 33.33
    isMuted.value = Boolean(parsed.isMuted)
    isPlaying.value = typeof parsed.isPlaying === 'boolean' ? parsed.isPlaying : true
  } catch {
    window.sessionStorage.removeItem(workspaceStorageKey.value)
  }
}

const cycleOption = (value: string, options: string[]) => {
  const currentIndex = options.indexOf(value)
  return options[(currentIndex + 1) % options.length]
}

const loadLiveTimelineData = async () => {
  if (!flowContext.value?.dramaId) return

  liveLoading.value = true
  liveTimelineError.value = ''

  try {
    const [storyboardsRes, assetsRes] = await Promise.all([
      dramaAPI.getStoryboards(flowContext.value.episodeId),
      assetAPI.listAssets({
        drama_id: flowContext.value.dramaId,
        episode_id: Number(flowContext.value.episodeId),
        type: 'video',
        page: 1,
        page_size: 100
      })
    ])

    liveStoryboards.value = storyboardsRes?.storyboards || []
    videoAssets.value = assetsRes.items || []

    if (liveStoryboards.value.length === 0) {
      liveTimelineError.value = '当前章节还没有分镜，请先回到内容创作或专业制作台生成分镜素材。'
    }
  } catch (error: any) {
    liveTimelineError.value = error?.message || '加载真实时间线数据失败，已保留当前页面配置。'
  } finally {
    liveLoading.value = false
  }
}

const handleNavClick = (path: string) => {
  if (!path) return
  router.push(path)
}

const goBackToGeneration = () => {
  if (!flowContext.value) return
  router.push(buildEpisodeStagePath('generation', flowContext.value))
}

const goToProfessionalEditor = () => {
  if (!flowContext.value) return
  router.push(buildProfessionalEditorPath(flowContext.value))
}

const handleSelectTemplate = (id: TemplateId) => {
  selectedTemplateId.value = selectedTemplateId.value === id ? '' : id
  persistWorkspace()
}

const handleSelectMusic = (id: MusicId) => {
  selectedMusicId.value = selectedMusicId.value === id ? '' : id
  persistWorkspace()
}

const handleSelectMaterial = (id: MaterialId) => {
  selectedMaterialId.value = selectedMaterialId.value === id ? '' : id
  persistWorkspace()
}

const handleSelectTool = (id: ToolId) => {
  selectedToolId.value = selectedToolId.value === id ? '' : id
  persistWorkspace()
  ElMessage.success('AI 工具已加入当前剪辑方案')
}

const handleSelectSegment = (id: SegmentId) => {
  selectedSegmentId.value = id
  persistWorkspace()
}

const handleUploadMaterial = () => {
  ElMessage.info('素材上传入口将在媒体工作流接通后启用，当前先保留高保真界面')
}

const togglePlayback = () => {
  isPlaying.value = !isPlaying.value
  persistWorkspace()
}

const toggleMute = () => {
  isMuted.value = !isMuted.value
  persistWorkspace()
}

const togglePause = () => {
  isPlaying.value = false
  persistWorkspace()
}

const handleTrim = () => {
  ElMessage.success('已进入片段裁剪模式')
}

const handleAddSegment = () => {
  ElMessage.info('添加片段功能将在素材编排接入后开放')
}

const cycleResolution = () => {
  selectedResolution.value = cycleOption(selectedResolution.value, resolutionOptions)
  persistWorkspace()
}

const cycleFormat = () => {
  selectedFormat.value = cycleOption(selectedFormat.value, formatOptions)
  persistWorkspace()
}

const cycleQuality = () => {
  selectedQuality.value = cycleOption(selectedQuality.value, qualityOptions)
  persistWorkspace()
}

const handleExportVideo = () => {
  persistWorkspace()
  ElMessage.success('导出任务已提交，当前先保留高保真页面与配置状态')
}

const handlePublish = () => {
  ElMessage.info('发布与数据链路将在成片工作流接通后启用')
}

const handleSaveProject = () => {
  persistWorkspace()
  ElMessage.success('视频剪辑项目已保存')
}

const handleLiveMergeCompleted = async () => {
  ElMessage.success('时间线合成任务已提交')
  await loadLiveTimelineData()
}

onMounted(async () => {
  const resolved = resolveEpisodeWorkflowContext({
    episodeId: route.params.id,
    dramaId: route.query.dramaId,
    episodeNumber: route.query.episodeNumber
  })

  if (resolved) {
    flowContext.value = resolved
    saveEpisodeWorkflowContext(resolved)
    await loadLiveTimelineData()
  }

  restoreWorkspace()
  persistWorkspace()
})
</script>

<style scoped>
.page-container.video-editor-page {
  --video-heading-font: 'Urbanist', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  --video-body-font: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  min-height: var(--app-vh, 100vh);
  padding: 0 !important;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%) !important;
  color: #0a2463;
  font-family: var(--video-body-font);
}

.video-editor-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
  height: 65px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.video-editor-header__inner {
  width: min(100%, 1075px);
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.video-editor-header__left,
.video-editor-header__right {
  display: flex;
  align-items: center;
}

.video-editor-header__left {
  gap: 32px;
  min-width: 0;
  flex: 1;
}

.brand-link {
  border: none;
  background: transparent;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0;
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
}

.brand-link__name {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
}

.video-editor-nav {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.video-editor-nav__item {
  height: 32px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #45556c;
  font-size: 14px;
  line-height: 20px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease;
}

.video-editor-nav__item--active {
  color: #0a2463;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
}

.video-editor-nav__item:not(.video-editor-nav__item--active):hover {
  color: #0a2463;
  background: rgba(241, 245, 249, 0.95);
}

.video-editor-header__right {
  gap: 16px;
  flex-shrink: 0;
}

.header-icon-button {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  position: relative;
  border-radius: 12px;
  cursor: pointer;
}

.header-icon-button img {
  width: 20px;
  height: 20px;
}

.header-icon-button__dot {
  position: absolute;
  top: 6px;
  right: 5px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #f97316;
}

.video-editor-main {
  padding-top: 104px;
  padding-bottom: 72px;
}

.video-editor-shell {
  width: min(100%, 1075px);
  margin: 0 auto;
  padding: 0 32px;
}

.video-editor-hero {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 32px;
}

.video-editor-hero h1 {
  margin: 0;
  font-family: var(--video-heading-font);
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #0a2463;
}

.video-editor-hero p {
  margin: 0;
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.video-editor-live {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.video-editor-live__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  padding: 28px 32px;
  border-radius: 24px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.98), rgba(240, 249, 255, 0.96));
  box-shadow: 0 22px 50px rgba(15, 23, 42, 0.08);
}

.video-editor-live__copy {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.video-editor-live__eyebrow {
  font-size: 13px;
  line-height: 20px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #0ea5e9;
}

.video-editor-live__copy h2 {
  margin: 0;
  font-family: var(--video-heading-font);
  font-size: 28px;
  line-height: 34px;
  font-weight: 700;
  color: #0a2463;
}

.video-editor-live__copy p {
  margin: 0;
  max-width: 720px;
  font-size: 15px;
  line-height: 24px;
  color: #45556c;
}

.video-editor-live__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
}

.video-editor-live__button {
  height: 44px;
  padding: 0 18px;
  border-radius: 14px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
}

.video-editor-live__button:hover {
  transform: translateY(-1px);
}

.video-editor-live__button--ghost {
  border: 1px solid rgba(148, 163, 184, 0.4);
  background: rgba(255, 255, 255, 0.92);
  color: #0a2463;
}

.video-editor-live__button--primary {
  border: none;
  background: linear-gradient(135deg, #0ea5e9, #2563eb);
  color: #ffffff;
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.24);
}

.video-editor-live__panel {
  padding: 24px;
  border-radius: 24px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  background: #ffffff;
  box-shadow: 0 22px 50px rgba(15, 23, 42, 0.08);
}

.video-editor-grid {
  display: grid;
  grid-template-columns: 315.664px minmax(0, 1fr);
  gap: 32px;
  align-items: start;
}

.video-editor-sidebar,
.video-editor-workspace {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.editor-card,
.workspace-card {
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
}

.editor-card {
  padding: 25px 25px 24px;
}

.editor-card__title,
.workspace-card__title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-card__title {
  margin-bottom: 24px;
}

.editor-card__title h2,
.workspace-card h2,
.workspace-card__title-row h2 {
  margin: 0;
  font-family: var(--video-heading-font);
  font-size: 18px;
  line-height: 28px;
  font-weight: 600;
  color: #0a2463;
}

.editor-card__title-icon {
  width: 24px;
  height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.editor-card__title-icon svg {
  width: 24px;
  height: 24px;
  fill: none;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.editor-card__title-icon--violet svg {
  stroke: #7c3aed;
}

.editor-card__title-icon--orange svg {
  stroke: #f97316;
}

.editor-card__title-icon--cyan svg {
  stroke: #06b6d4;
}

.editor-card__title-icon--purple svg {
  stroke: #7c3aed;
}

.template-list,
.music-list,
.tool-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.template-card,
.music-card {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: linear-gradient(165.24deg, #f8fafc 0%, #ffffff 100%);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.template-card {
  height: 70px;
  padding: 17px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.template-card:hover,
.music-card:hover,
.material-tile:hover,
.tool-button:hover,
.timeline-row__segment:hover,
.export-select:hover,
.workspace-actions__primary:hover,
.workspace-actions__secondary:hover,
.workspace-actions__ghost:hover,
.upload-material-button:hover,
.timeline-add-button:hover {
  transform: translateY(-1px);
}

.template-card--active {
  border-color: #7c3aed;
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.08);
}

.template-card__emoji {
  width: 41px;
  font-size: 30px;
  line-height: 36px;
  display: inline-flex;
  justify-content: center;
}

.template-card__copy,
.music-card__top,
.music-card__bottom {
  width: 100%;
}

.template-card__copy {
  display: flex;
  flex-direction: column;
}

.template-card__copy strong,
.music-card__top strong {
  font-size: 14px;
  line-height: 20px;
  font-weight: 700;
  color: #0a2463;
}

.template-card__copy span,
.music-card__bottom span {
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.music-card {
  height: 80px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.music-card--active {
  border-color: #f97316;
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.08);
}

.music-card__top,
.music-card__bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.music-card__play {
  width: 24px;
  height: 24px;
  border-radius: 12px;
  background: #f97316;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.music-card__play svg {
  width: 12px;
  height: 12px;
  stroke: #ffffff;
  fill: none;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.editor-card--materials {
  padding-bottom: 24px;
}

.material-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.material-tile {
  aspect-ratio: 1 / 1;
  border: none;
  border-radius: 16px;
  background: #e2e8f0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, outline 0.2s ease;
}

.material-tile svg {
  width: 32px;
  height: 32px;
  stroke: #94a3b8;
  fill: none;
  stroke-width: 1.65;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.material-tile--active {
  outline: 2px solid #06b6d4;
  box-shadow: 0 12px 24px rgba(6, 182, 212, 0.16);
}

.upload-material-button {
  width: 100%;
  height: 36px;
  border: none;
  border-radius: 12px;
  background: #f1f5f9;
  font-size: 12px;
  line-height: 20px;
  color: #0a2463;
  cursor: pointer;
}

.editor-card--tools {
  border-color: #e9d5ff;
  background: linear-gradient(180deg, rgba(250, 245, 255, 0.95) 0%, rgba(255, 255, 255, 0.98) 100%);
}

.tool-button {
  height: 36px;
  border: none;
  border-radius: 12px;
  background: #ffffff;
  text-align: left;
  padding: 0 16px;
  font-size: 14px;
  line-height: 20px;
  color: #314158;
  cursor: pointer;
  box-shadow: inset 0 0 0 1px rgba(233, 213, 255, 0.25);
  transition: transform 0.2s ease, box-shadow 0.2s ease, color 0.2s ease;
}

.tool-button--active {
  color: #6d28d9;
  box-shadow: inset 0 0 0 1px rgba(124, 58, 237, 0.28), 0 8px 18px rgba(124, 58, 237, 0.08);
}

.workspace-card {
  padding: 25px;
}

.workspace-card--preview {
  padding-bottom: 25px;
}

.preview-stage {
  position: relative;
  height: 345px;
  border-radius: 16px;
  overflow: hidden;
  background: linear-gradient(150.64deg, #111d34 0%, #1f2a44 100%);
}

.preview-stage__surface {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 12% 14%, rgba(255, 255, 255, 0.06), transparent 20%),
    linear-gradient(135deg, rgba(6, 182, 212, 0.03), transparent 35%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.02), transparent 62%);
}

.preview-stage__play {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 80px;
  height: 80px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 24px 42px rgba(249, 115, 22, 0.3);
}

.preview-stage__play svg {
  width: 40px;
  height: 40px;
  stroke: #ffffff;
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.preview-stage__footer {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.preview-stage__time {
  height: 32px;
  border-radius: 12px;
  padding: 0 12px;
  display: inline-flex;
  align-items: center;
  background: rgba(15, 23, 42, 0.8);
  font-size: 12px;
  line-height: 20px;
  color: #ffffff;
}

.preview-stage__volume {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 16px;
  background: rgba(15, 23, 42, 0.8);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.preview-stage__volume svg {
  width: 16px;
  height: 16px;
  stroke: #ffffff;
  fill: none;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.preview-controls {
  height: 56px;
  margin-top: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.preview-controls__button {
  width: 44px;
  height: 44px;
  border: none;
  border-radius: 16px;
  background: #f1f5f9;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.preview-controls__button svg {
  width: 20px;
  height: 20px;
  stroke: #62748e;
  fill: none;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.preview-controls__button--play {
  width: 56px;
  height: 56px;
  border-radius: 18px;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
}

.preview-controls__button--play svg {
  width: 24px;
  height: 24px;
  stroke: #ffffff;
}

.preview-controls__button--active {
  box-shadow: 0 18px 28px rgba(249, 115, 22, 0.2);
}

.preview-progress {
  position: relative;
  height: 8px;
  margin-top: 8px;
}

.preview-progress__track,
.preview-progress__fill {
  position: absolute;
  left: 0;
  top: 0;
  height: 8px;
  border-radius: 999px;
}

.preview-progress__track {
  width: 100%;
  background: #e2e8f0;
}

.preview-progress__fill {
  background: linear-gradient(90deg, #06b6d4 0%, #7c3aed 100%);
}

.preview-progress__thumb {
  position: absolute;
  top: 50%;
  width: 16px;
  height: 16px;
  transform: translate(-50%, -50%);
  border-radius: 999px;
  background: #ffffff;
  border: 3px solid #06b6d4;
  box-shadow: 0 8px 20px rgba(6, 182, 212, 0.18);
}

.workspace-card--timeline {
  padding-bottom: 24px;
}

.workspace-card__title-row {
  justify-content: space-between;
  margin-bottom: 24px;
}

.workspace-card__title-row span {
  font-size: 14px;
  line-height: 20px;
  color: #0a2463;
}

.timeline-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.timeline-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.timeline-row__duration {
  width: 64px;
  font-size: 14px;
  line-height: 20px;
  color: #45556c;
  text-align: right;
}

.timeline-row__segment {
  flex: 1;
  height: 48px;
  border: none;
  border-radius: 16px;
  padding: 0 16px;
  text-align: left;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #ffffff;
  cursor: pointer;
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s ease, box-shadow 0.2s ease, filter 0.2s ease;
}

.timeline-row__segment--blue {
  background: #2b7fff;
}

.timeline-row__segment--purple {
  background: linear-gradient(90deg, #ad46ff 0%, #d946ef 100%);
}

.timeline-row__segment--orange {
  background: #ff6900;
}

.timeline-row__segment--green {
  background: #00c950;
}

.timeline-row__segment:hover {
  filter: brightness(1.02);
}

.timeline-add-button {
  width: 100%;
  height: 48px;
  border: none;
  border-radius: 16px;
  background: #f1f5f9;
  font-size: 16px;
  line-height: 24px;
  color: #0a2463;
  cursor: pointer;
}

.workspace-card--export {
  border-color: #bedbff;
  background: linear-gradient(166.19deg, #eff6ff 0%, #ecfeff 100%);
}

.workspace-card--export h2 {
  margin-bottom: 16px;
}

.export-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.export-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.export-field span {
  font-size: 14px;
  line-height: 20px;
  color: #314158;
}

.export-select {
  height: 41px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  padding: 0 14px;
  text-align: left;
  font-size: 14px;
  line-height: 20px;
  color: #0a2463;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.export-select:hover {
  border-color: #bedbff;
  box-shadow: 0 10px 18px rgba(190, 219, 255, 0.28);
}

.workspace-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.workspace-actions__primary,
.workspace-actions__secondary,
.workspace-actions__ghost {
  height: 60px;
  border: none;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.workspace-actions__primary {
  flex: 1;
  min-width: 0;
  background: linear-gradient(90deg, #10b981 0%, #34d399 100%);
  box-shadow: 0 14px 26px rgba(16, 185, 129, 0.18);
}

.workspace-actions__primary svg {
  width: 20px;
  height: 20px;
  stroke: #ffffff;
  fill: none;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.workspace-actions__primary span,
.workspace-actions__secondary span {
  color: #ffffff;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
}

.workspace-actions__secondary {
  width: 188px;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
  box-shadow: 0 14px 26px rgba(249, 115, 22, 0.16);
}

.workspace-actions__secondary img {
  width: 20px;
  height: 20px;
  filter: brightness(0) invert(1);
}

.workspace-actions__ghost {
  width: 116px;
  border: 2px solid #e2e8f0;
  background: #ffffff;
  color: #0a2463;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
}

@media (max-width: 1140px) {
  .video-editor-header__inner,
  .video-editor-shell {
    padding-left: 20px;
    padding-right: 20px;
  }

  .video-editor-nav {
    display: none;
  }

  .video-editor-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 820px) {
  .video-editor-header__right {
    display: none;
  }

  .export-grid {
    grid-template-columns: 1fr;
  }

  .workspace-actions {
    flex-direction: column;
  }

  .workspace-actions__primary,
  .workspace-actions__secondary,
  .workspace-actions__ghost {
    width: 100%;
  }
}

@media (max-width: 680px) {
  .video-editor-main {
    padding-top: 92px;
    padding-bottom: 48px;
  }

  .video-editor-shell {
    padding-left: 16px;
    padding-right: 16px;
  }

  .video-editor-hero h1 {
    font-size: 26px;
    line-height: 32px;
  }

  .editor-card,
  .workspace-card {
    padding: 20px;
  }

  .preview-stage {
    height: 260px;
  }

  .preview-controls {
    gap: 12px;
  }

  .timeline-row {
    flex-direction: column;
    align-items: stretch;
  }

  .timeline-row__duration {
    width: auto;
    text-align: left;
  }
}
</style>
