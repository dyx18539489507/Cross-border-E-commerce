<template>
  <div class="page-container script-storyboard-page">
    <header class="stage-header">
      <div class="stage-header__inner">
        <div class="stage-header__left">
          <button type="button" class="brand-link" @click="router.push('/')">
            <span class="brand-link__mark">
              <img :src="brandIcon" alt="" />
            </span>
            <span class="brand-link__name">{{ t('app.name') }}</span>
          </button>

          <nav class="stage-nav" aria-label="主导航">
            <button
              v-for="item in navItems"
              :key="item.label"
              type="button"
              class="stage-nav__item"
              :class="{ 'stage-nav__item--active': item.active }"
              :style="{ width: item.width }"
              @click="handleNavClick(item.path)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="stage-header__right">
          <button type="button" class="header-icon-button" aria-label="通知">
            <img :src="bellIcon" alt="" />
            <span class="header-icon-button__dot"></span>
          </button>
        </div>
      </div>
    </header>

    <main class="script-stage-main">
      <div class="script-stage-shell">
        <section class="script-stage-hero">
          <h1>营销脚本与分镜生成</h1>
          <p>AI生成多语种营销脚本与可视化分镜，为内容创作做好准备</p>
        </section>

        <section class="script-stage-grid">
          <aside class="language-panel">
            <div class="language-panel__title">
              <span class="language-panel__icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none">
                  <path d="M7 4v6" />
                  <path d="M4 7h6" />
                  <path d="M14 4h6" />
                  <path d="M17 4v2" />
                  <path d="M14.75 8.5c.9 2.4 2.4 4.6 4.25 6.2" />
                  <path d="M19 8.5c-.55 1.85-1.55 3.55-2.85 4.95" />
                  <path d="M9.75 13.5h5.5" />
                  <path d="M12.5 10.5v6" />
                  <path d="M10.25 16.5l2.25-6 2.25 6" />
                </svg>
              </span>
              <h2>语言版本</h2>
            </div>

            <div class="language-panel__list">
              <button
                v-for="version in languageVersions"
                :key="version.code"
                type="button"
                class="language-card"
                :class="{ 'language-card--active': version.code === selectedLanguageCode }"
                @click="handleSelectLanguage(version.code)"
              >
                <div class="language-card__row">
                  <div class="language-card__label-group">
                    <span class="language-card__flag">{{ version.flag }}</span>
                    <span class="language-card__label">{{ version.label }}</span>
                  </div>

                  <span
                    v-if="version.completed"
                    class="language-card__check"
                    aria-hidden="true"
                  >
                    <svg viewBox="0 0 20 20" fill="none">
                      <path d="M18.333 10a8.333 8.333 0 1 1-16.666 0 8.333 8.333 0 0 1 16.666 0Z" />
                      <path d="m6.667 10.417 2.083 2.083 4.583-5" />
                    </svg>
                  </span>
                </div>

                <div class="language-card__progress-track">
                  <span
                    class="language-card__progress-fill"
                    :style="{ width: `${version.progress}%` }"
                  ></span>
                </div>

                <span class="language-card__progress-text">{{ version.progress }}% 完成</span>
              </button>
            </div>

            <button type="button" class="generate-language-button" @click="handleGenerateMoreLanguages">
              <span class="generate-language-button__icon" aria-hidden="true">
                <svg viewBox="0 0 20 20" fill="none">
                  <path d="M10 2.5v4.167" />
                  <path d="M10 13.333V17.5" />
                  <path d="M2.5 10h4.167" />
                  <path d="M13.333 10H17.5" />
                  <path d="m5.833 5.833 2.083 2.084" />
                  <path d="m12.084 12.083 2.083 2.084" />
                </svg>
              </span>
              <span>生成更多语言</span>
            </button>
          </aside>

          <div class="script-stage-content">
            <section class="stage-card stage-card--title">
              <div class="stage-card__header">
                <h3>标题</h3>

                <div class="stage-card__actions">
                  <button
                    type="button"
                    class="plain-icon-button"
                    aria-label="刷新标题"
                    @click="handleRefreshSection('title')"
                  >
                    <svg viewBox="0 0 20 20" fill="none">
                      <path d="M16.667 3.333v4.167H12.5" />
                      <path d="M15 8.333A6.667 6.667 0 1 0 17 13.334" />
                    </svg>
                  </button>

                  <button
                    type="button"
                    class="plain-icon-button"
                    aria-label="复制标题"
                    @click="handleCopySection('title')"
                  >
                    <svg viewBox="0 0 20 20" fill="none">
                      <rect x="7.5" y="7.5" width="8.333" height="8.333" rx="1.5" />
                      <path d="M5.833 12.5H5a1.667 1.667 0 0 1-1.667-1.667V5A1.667 1.667 0 0 1 5 3.333h5.833A1.667 1.667 0 0 1 12.5 5v.833" />
                    </svg>
                  </button>
                </div>
              </div>

              <div class="origin-card origin-card--title">
                <span class="origin-card__label">原文</span>
                <p class="origin-card__text origin-card__text--single">{{ originalTitle }}</p>
              </div>

              <div class="editable-panel editable-panel--title">
                <textarea
                  v-model="titleOutput"
                  spellcheck="false"
                  aria-label="营销标题编辑区"
                ></textarea>
              </div>
            </section>

            <section class="stage-card stage-card--description">
              <div class="stage-card__header">
                <h3>商品描述</h3>

                <div class="stage-card__actions">
                  <button
                    type="button"
                    class="plain-icon-button"
                    aria-label="刷新商品描述"
                    @click="handleRefreshSection('description')"
                  >
                    <svg viewBox="0 0 20 20" fill="none">
                      <path d="M16.667 3.333v4.167H12.5" />
                      <path d="M15 8.333A6.667 6.667 0 1 0 17 13.334" />
                    </svg>
                  </button>

                  <button
                    type="button"
                    class="plain-icon-button"
                    aria-label="复制商品描述"
                    @click="handleCopySection('description')"
                  >
                    <svg viewBox="0 0 20 20" fill="none">
                      <rect x="7.5" y="7.5" width="8.333" height="8.333" rx="1.5" />
                      <path d="M5.833 12.5H5a1.667 1.667 0 0 1-1.667-1.667V5A1.667 1.667 0 0 1 5 3.333h5.833A1.667 1.667 0 0 1 12.5 5v.833" />
                    </svg>
                  </button>
                </div>
              </div>

              <div class="origin-card origin-card--description">
                <span class="origin-card__label">原文</span>
                <p class="origin-card__text">{{ originalDescription }}</p>
              </div>

              <div class="editable-panel editable-panel--description">
                <textarea
                  v-model="descriptionOutput"
                  spellcheck="false"
                  aria-label="营销描述编辑区"
                ></textarea>
              </div>
            </section>

            <section class="stage-card stage-card--features">
              <div class="stage-card__header">
                <h3>产品特点</h3>

                <button
                  type="button"
                  class="plain-icon-button"
                  aria-label="刷新产品特点"
                  @click="handleRefreshFeatures"
                >
                  <svg viewBox="0 0 20 20" fill="none">
                    <path d="M16.667 3.333v4.167H12.5" />
                    <path d="M15 8.333A6.667 6.667 0 1 0 17 13.334" />
                  </svg>
                </button>
              </div>

              <div class="feature-columns">
                <div class="feature-column">
                  <span class="feature-column__label">原文</span>
                  <ul class="feature-list feature-list--origin">
                    <li v-for="feature in originalFeatures" :key="feature">
                      <span class="feature-list__dot"></span>
                      <span>{{ feature }}</span>
                    </li>
                  </ul>
                </div>

                <div class="feature-column">
                  <span class="feature-column__label">译文</span>
                  <ul class="feature-list feature-list--translated">
                    <li v-for="feature in selectedLanguage.translatedFeatures" :key="feature">
                      <span class="feature-list__check">✓</span>
                      <span>{{ feature }}</span>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section class="insight-row">
              <article class="insight-card">
                <div class="insight-card__header">
                  <span class="insight-card__icon insight-card__icon--culture" aria-hidden="true">
                    <svg viewBox="0 0 20 20" fill="none">
                      <path d="M10 3.333a6.667 6.667 0 1 1 0 13.334" />
                      <path d="M10 6.667a3.333 3.333 0 0 1 0 6.666" />
                      <path d="M3.333 10h3.334" />
                      <path d="M13.333 10h3.334" />
                    </svg>
                  </span>
                  <h4>文化适配建议</h4>
                </div>
                <p>美国用户更关注健康数据的准确性和隐私保护，建议强调"medical-grade"和"privacy-first"</p>
              </article>

              <article class="insight-card">
                <div class="insight-card__header">
                  <span class="insight-card__icon insight-card__icon--seo" aria-hidden="true">
                    <svg viewBox="0 0 20 20" fill="none">
                      <path d="M10 2.5v4.167" />
                      <path d="M10 13.333V17.5" />
                      <path d="M2.5 10h4.167" />
                      <path d="M13.333 10H17.5" />
                      <path d="m5.833 5.833 2.083 2.084" />
                      <path d="m12.084 12.083 2.083 2.084" />
                    </svg>
                  </span>
                  <h4>SEO优化建议</h4>
                </div>
                <p>添加关键词: "fitness tracker", "health smartwatch", "activity monitor"以提升搜索排名</p>
              </article>
            </section>

            <section class="stage-card stage-card--preview">
              <h3>分镜预览</h3>

              <div class="storyboard-preview">
                <article v-for="shot in previewShots" :key="shot.title" class="storyboard-preview__card">
                  <span class="storyboard-preview__emoji">🎬</span>
                  <strong>{{ shot.title }}</strong>
                  <span>{{ shot.duration }}</span>
                </article>
              </div>
            </section>

            <section class="stage-actions">
              <button type="button" class="stage-actions__primary" @click="handleBeginCreation">
                <span>开始创作内容（图片/视频/数字人）</span>
                <img :src="ctaArrowIcon" alt="" />
              </button>

              <button type="button" class="stage-actions__secondary" @click="handleSaveDraft">
                保存草稿
              </button>

              <button type="button" class="stage-actions__ghost" @click="router.push('/dramas')">
                返回工作台
              </button>
            </section>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { dramaAPI } from '@/api/drama'
import bellIcon from '@/assets/product-entry/bell.svg'
import brandIcon from '@/assets/product-entry/brand-icon.svg'
import ctaArrowIcon from '@/assets/product-entry/arrow-right.svg'
import type { Drama, Episode } from '@/types/drama'
import {
  buildEpisodeStagePath,
  resolveEpisodeWorkflowContext,
  saveEpisodeWorkflowContext,
  type EpisodeWorkflowContext
} from '@/utils/episodeWorkflowContext'

type LanguageCode = 'en-US' | 'de-DE' | 'ja-JP' | 'es-ES' | 'fr-FR'
type EditableSection = 'title' | 'description'

interface ProductEntryBasicDraft {
  title?: string
  category?: string
  brand?: string
}

interface CreateDramaDraftPayload {
  title?: string
  description?: string
  marketing_selling_points?: string
}

interface LanguageVersion {
  code: LanguageCode
  flag: string
  label: string
  progress: number
  completed: boolean
  generatedTitle: string
  generatedDescription: string
  translatedFeatures: string[]
}

const PRODUCT_ENTRY_DRAFT_KEY = 'drama:create:product-entry:basic'
const CREATE_DRAMA_DRAFT_KEY = 'drama:create:draft'
const WORKSPACE_STORAGE_PREFIX = 'drama:script-storyboard:workspace:'

const DEFAULT_TITLE = '智能手表 Pro - 健康生活的智能伴侣'
const DEFAULT_DESCRIPTION =
  '全天候健康监测，精准运动追踪，智能生活助手。采用医疗级传感器，实时监测心率、血氧、睡眠质量。续航长达7天，防水等级IP68。'
const DEFAULT_FEATURES = [
  '24小时心率监测',
  '血氧饱和度检测',
  '睡眠质量分析',
  '50+运动模式',
  '7天超长续航',
  'IP68防水防尘'
]

const LANGUAGE_VERSIONS: LanguageVersion[] = [
  {
    code: 'en-US',
    flag: '🇺🇸',
    label: '英语 (美国)',
    progress: 100,
    completed: true,
    generatedTitle: 'Smartwatch Pro - Your Intelligent Partner for Healthy Living',
    generatedDescription:
      'Stay on top of your wellness with all-day health tracking, precise fitness insights, and a smart assistant built for daily life. Medical-grade sensors continuously monitor heart rate, SpO2, and sleep quality, while the 7-day battery and IP68 protection keep up with your routine.',
    translatedFeatures: [
      '24/7 Heart Rate Monitoring',
      'Blood Oxygen Level Detection',
      'Sleep Quality Analysis',
      '50+ Sports Modes',
      '7-Day Battery Life',
      'IP68 Water & Dust Resistance'
    ]
  },
  {
    code: 'de-DE',
    flag: '🇩🇪',
    label: '德语',
    progress: 100,
    completed: true,
    generatedTitle: 'Smartwatch Pro - Ihr intelligenter Begleiter fur ein gesundes Leben',
    generatedDescription:
      'Erleben Sie ganzheitliches Gesundheitsmonitoring, prazises Sport-Tracking und smarte Alltagshilfe in einem eleganten Gerat. Medizinische Sensoren uberwachen Herzfrequenz, Blutsauerstoff und Schlafqualitat in Echtzeit, wahrend 7 Tage Akkulaufzeit und IP68-Schutz fur Verlasslichkeit sorgen.',
    translatedFeatures: [
      '24/7 Herzfrequenzuberwachung',
      'Blutsauerstoffmessung',
      'Analyse der Schlafqualitat',
      'Uber 50 Sportmodi',
      '7 Tage Akkulaufzeit',
      'IP68 Wasser- und Staubschutz'
    ]
  },
  {
    code: 'ja-JP',
    flag: '🇯🇵',
    label: '日语',
    progress: 80,
    completed: false,
    generatedTitle: 'Smartwatch Pro - 健康な毎日を支えるスマートパートナー',
    generatedDescription:
      '24時間の健康モニタリング、精密な運動追跡、そして日常を助けるスマートアシスタントを一台に。医療グレードのセンサーが心拍数、血中酸素、睡眠の質をリアルタイムで計測します。',
    translatedFeatures: [
      '24時間心拍モニタリング',
      '血中酸素レベル測定',
      '睡眠品質分析',
      '50種類以上のスポーツモード',
      '7日間ロングバッテリー',
      'IP68防水防塵'
    ]
  },
  {
    code: 'es-ES',
    flag: '🇪🇸',
    label: '西班牙语',
    progress: 60,
    completed: false,
    generatedTitle: 'Smartwatch Pro - Tu companero inteligente para una vida saludable',
    generatedDescription:
      'Monitoreo integral de salud, seguimiento deportivo preciso y un asistente inteligente para el dia a dia. Sus sensores de nivel medico vigilan la frecuencia cardiaca, el oxigeno en sangre y la calidad del sueno.',
    translatedFeatures: [
      'Monitorizacion cardiaca 24/7',
      'Deteccion de oxigeno en sangre',
      'Analisis de la calidad del sueno',
      'Mas de 50 modos deportivos',
      'Bateria de 7 dias',
      'Resistencia IP68 al agua y polvo'
    ]
  },
  {
    code: 'fr-FR',
    flag: '🇫🇷',
    label: '法语',
    progress: 60,
    completed: false,
    generatedTitle: 'Smartwatch Pro - Votre partenaire intelligent pour une vie saine',
    generatedDescription:
      "Profitez d'un suivi sante continu, d'une analyse sportive precise et d'un assistant intelligent concu pour le quotidien. Les capteurs de grade medical surveillent en temps reel le rythme cardiaque, l'oxygene sanguin et la qualite du sommeil.",
    translatedFeatures: [
      'Suivi cardiaque 24h/24',
      "Mesure de l'oxygene sanguin",
      'Analyse de la qualite du sommeil',
      '50+ modes sportifs',
      'Autonomie de 7 jours',
      "Resistance a l'eau et a la poussiere IP68"
    ]
  }
]

const previewShots = [
  { title: '开场', duration: '3s' },
  { title: '产品特写', duration: '5s' },
  { title: '功能演示', duration: '7s' },
  { title: 'CTA', duration: '9s' }
]

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const flowContext = ref<EpisodeWorkflowContext | null>(null)
const dramaRecord = ref<Drama | null>(null)

const episodeId = computed(() => flowContext.value?.episodeId || String(route.params.id || route.query.id || ''))
const workspaceStorageKey = computed(
  () => `${WORKSPACE_STORAGE_PREFIX}${episodeId.value || String(route.params.id || route.name || 'draft')}`
)

const navItems = computed(() => [
  { label: '工作台', path: '/dramas', active: false, width: '66px' },
  { label: '商品录入', path: '/dramas/create', active: false, width: '80px' },
  { label: '合规分析', path: '/compliance', active: false, width: '80px' },
  {
    label: '脚本/分镜',
    path: flowContext.value ? buildEpisodeStagePath('script', flowContext.value) : '/workspace/script',
    active: true,
    width: '86px'
  },
  {
    label: '内容创作',
    path: flowContext.value ? buildEpisodeStagePath('generation', flowContext.value) : '/workspace/content',
    active: false,
    width: '80px'
  },
  {
    label: '视频剪辑',
    path: flowContext.value ? buildEpisodeStagePath('timeline', flowContext.value) : '/workspace/timeline',
    active: false,
    width: '80px'
  },
  { label: '数据分析', path: '/analytics', active: false, width: '80px' }
])

const originalTitle = ref(DEFAULT_TITLE)
const originalDescription = ref(DEFAULT_DESCRIPTION)
const originalFeatures = ref<string[]>([...DEFAULT_FEATURES])
const selectedLanguageCode = ref<LanguageCode>('en-US')

const editableOutputs = reactive<Record<LanguageCode, { title: string; description: string }>>({
  'en-US': { title: '', description: '' },
  'de-DE': { title: '', description: '' },
  'ja-JP': { title: '', description: '' },
  'es-ES': { title: '', description: '' },
  'fr-FR': { title: '', description: '' }
})

const languageVersions = LANGUAGE_VERSIONS

const selectedLanguage = computed(
  () => languageVersions.find((item) => item.code === selectedLanguageCode.value) || languageVersions[0]
)

const titleOutput = computed({
  get: () => editableOutputs[selectedLanguageCode.value].title,
  set: (value: string) => {
    editableOutputs[selectedLanguageCode.value].title = value
    persistWorkspace()
  }
})

const descriptionOutput = computed({
  get: () => editableOutputs[selectedLanguageCode.value].description,
  set: (value: string) => {
    editableOutputs[selectedLanguageCode.value].description = value
    persistWorkspace()
  }
})

const splitFeatureText = (value?: string) =>
  String(value || '')
    .split(/[\n,，;；、]/)
    .map((item) => item.trim())
    .filter((item) => item.length > 0)

const buildDisplayTitle = (rawTitle?: string) => {
  const cleaned = String(rawTitle || '').trim()
  if (!cleaned) return DEFAULT_TITLE
  if (/[-—]/.test(cleaned)) return cleaned
  return `${cleaned} - 健康生活的智能伴侣`
}

const readJsonSession = <T>(key: string): T | null => {
  if (typeof window === 'undefined') return null
  const raw = window.sessionStorage.getItem(key)
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    window.sessionStorage.removeItem(key)
    return null
  }
}

const restoreSourceDrafts = () => {
  const basicDraft = readJsonSession<ProductEntryBasicDraft>(PRODUCT_ENTRY_DRAFT_KEY)
  if (basicDraft?.title) {
    originalTitle.value = buildDisplayTitle(basicDraft.title)
  }

  const createDramaDraft = readJsonSession<CreateDramaDraftPayload>(CREATE_DRAMA_DRAFT_KEY)
  if (createDramaDraft?.title && !basicDraft?.title) {
    originalTitle.value = buildDisplayTitle(createDramaDraft.title)
  }

  if (createDramaDraft?.description && createDramaDraft.description.trim().length >= 18) {
    originalDescription.value = createDramaDraft.description.trim()
  }

  const draftFeatures = splitFeatureText(createDramaDraft?.marketing_selling_points)
  if (draftFeatures.length >= 3) {
    originalFeatures.value = draftFeatures.slice(0, 6)
  }
}

const applyRemoteDraft = (drama: Drama, episode?: Episode | null) => {
  if (drama.title?.trim()) {
    originalTitle.value = buildDisplayTitle(drama.title)
  }

  const description = String(episode?.description || drama.description || '').trim()
  if (description) {
    originalDescription.value = description
  }

  const featureList = splitFeatureText(drama.marketing_selling_points)
  if (featureList.length > 0) {
    originalFeatures.value = featureList.slice(0, 6)
  }
}

const persistWorkspace = () => {
  if (typeof window === 'undefined') return

  const payload = {
    selectedLanguageCode: selectedLanguageCode.value,
    editableOutputs: JSON.parse(JSON.stringify(editableOutputs)),
    originalTitle: originalTitle.value,
    originalDescription: originalDescription.value,
    originalFeatures: [...originalFeatures.value]
  }

  window.sessionStorage.setItem(workspaceStorageKey.value, JSON.stringify(payload))
}

const restoreWorkspace = () => {
  const workspace = readJsonSession<{
    selectedLanguageCode?: LanguageCode
    editableOutputs?: Partial<Record<LanguageCode, { title?: string; description?: string }>>
    originalTitle?: string
    originalDescription?: string
    originalFeatures?: string[]
  }>(workspaceStorageKey.value)

  if (!workspace) return

  if (workspace.selectedLanguageCode) {
    selectedLanguageCode.value = workspace.selectedLanguageCode
  }

  if (workspace.originalTitle?.trim()) {
    originalTitle.value = workspace.originalTitle.trim()
  }

  if (workspace.originalDescription?.trim()) {
    originalDescription.value = workspace.originalDescription.trim()
  }

  if (Array.isArray(workspace.originalFeatures) && workspace.originalFeatures.length > 0) {
    originalFeatures.value = workspace.originalFeatures.map((item) => String(item).trim()).filter(Boolean)
  }

  if (workspace.editableOutputs) {
    ;(Object.keys(editableOutputs) as LanguageCode[]).forEach((code) => {
      const item = workspace.editableOutputs?.[code]
      if (!item) return
      editableOutputs[code].title = typeof item.title === 'string' ? item.title : editableOutputs[code].title
      editableOutputs[code].description =
        typeof item.description === 'string' ? item.description : editableOutputs[code].description
    })
  }
}

const buildEpisodeScriptDocument = () => {
  const localizedTitle = titleOutput.value.trim() || selectedLanguage.value.generatedTitle || originalTitle.value
  const localizedDescription =
    descriptionOutput.value.trim() || selectedLanguage.value.generatedDescription || originalDescription.value
  const localizedFeatures =
    selectedLanguage.value.translatedFeatures.length > 0
      ? selectedLanguage.value.translatedFeatures
      : originalFeatures.value

  return [
    `# ${localizedTitle}`,
    `## 商品描述\n${localizedDescription}`,
    `## 产品特点\n${localizedFeatures.map((feature) => `- ${feature}`).join('\n')}`,
    `## 文化适配建议\n美国用户更关注健康数据的准确性和隐私保护，建议强调"medical-grade"和"privacy-first"`,
    `## SEO优化建议\n添加关键词: "fitness tracker", "health smartwatch", "activity monitor"`,
    `## 分镜预览\n${previewShots
      .map((shot, index) => `${index + 1}. ${shot.title}（${shot.duration}）`)
      .join('\n')}`
  ].join('\n\n')
}

const buildEpisodeSummary = () =>
  (descriptionOutput.value.trim() || selectedLanguage.value.generatedDescription || originalDescription.value).trim()

const buildEpisodeTitle = (episodeNumber: number) => {
  const contentTitle = (titleOutput.value.trim() || selectedLanguage.value.generatedTitle || originalTitle.value).trim()
  return contentTitle ? `第${episodeNumber}集 · ${contentTitle}` : `第${episodeNumber}集`
}

const ensureEpisodeContext = async (): Promise<EpisodeWorkflowContext | null> => {
  if (flowContext.value) return flowContext.value

  if (route.name === 'DramaScriptStage') {
    const dramaId = String(route.params.id || '').trim()
    if (!dramaId) return null

    const drama = await dramaAPI.get(dramaId)
    dramaRecord.value = drama

    let targetEpisode = drama.episodes?.find((episode) => episode.episode_number === 1) || drama.episodes?.[0]

    if (!targetEpisode) {
      await dramaAPI.saveEpisodes(dramaId, [
        {
          episode_number: 1,
          title: '第1集',
          description: drama.description || '',
          script_content: ''
        }
      ])

      const refreshedDrama = await dramaAPI.get(dramaId)
      dramaRecord.value = refreshedDrama
      targetEpisode =
        refreshedDrama.episodes?.find((episode) => episode.episode_number === 1) || refreshedDrama.episodes?.[0]
    }

    if (!targetEpisode?.id) return null

    flowContext.value = {
      dramaId,
      episodeId: String(targetEpisode.id),
      episodeNumber: targetEpisode.episode_number || 1
    }
    saveEpisodeWorkflowContext(flowContext.value)
    return flowContext.value
  }

  const resolved = resolveEpisodeWorkflowContext({
    episodeId: route.params.id,
    dramaId: route.query.dramaId,
    episodeNumber: route.query.episodeNumber
  })

  if (!resolved) return null

  flowContext.value = resolved
  saveEpisodeWorkflowContext(resolved)
  return resolved
}

const persistEpisodeScript = async () => {
  const context = await ensureEpisodeContext()
  if (!context?.dramaId) return 'local' as const

  try {
    const drama =
      dramaRecord.value && String(dramaRecord.value.id) === context.dramaId
        ? dramaRecord.value
        : await dramaAPI.get(context.dramaId)

    const existingEpisodes = [...(drama.episodes || [])]
    const hasProtectedProductionData = existingEpisodes.some((episode) => {
      return (
        Number(episode.storyboard_count || 0) > 0
        || Number(episode.scene_count || 0) > 0
        || Number(episode.video_count || 0) > 0
        || Number(episode.composition_count || 0) > 0
      )
    })

    if (existingEpisodes.length > 1 || hasProtectedProductionData) {
      return 'local' as const
    }

    const episodeNumber = context.episodeNumber || 1
    const summary = buildEpisodeSummary()
    const scriptContent = buildEpisodeScriptDocument()

    const episodePayloads = existingEpisodes.map((episode) => ({
      episode_number: episode.episode_number,
      title: episode.title || `第${episode.episode_number}集`,
      description: episode.description || '',
      script_content: episode.script_content || '',
      duration: episode.duration
    }))

    const targetIndex = episodePayloads.findIndex((episode) => episode.episode_number === episodeNumber)
    const nextPayload = {
      episode_number: episodeNumber,
      title: buildEpisodeTitle(episodeNumber),
      description: summary,
      script_content: scriptContent,
      duration: episodePayloads[targetIndex]?.duration
    }

    if (targetIndex >= 0) {
      episodePayloads[targetIndex] = {
        ...episodePayloads[targetIndex],
        ...nextPayload
      }
    } else {
      episodePayloads.push(nextPayload)
      episodePayloads.sort((left, right) => left.episode_number - right.episode_number)
    }

    await dramaAPI.saveEpisodes(context.dramaId, episodePayloads)

    const refreshedDrama = await dramaAPI.get(context.dramaId)
    dramaRecord.value = refreshedDrama

    const refreshedEpisode =
      refreshedDrama.episodes?.find((episode) => episode.episode_number === episodeNumber) ||
      refreshedDrama.episodes?.[0]

    if (refreshedEpisode?.id) {
      flowContext.value = {
        dramaId: context.dramaId,
        episodeId: String(refreshedEpisode.id),
        episodeNumber: refreshedEpisode.episode_number || episodeNumber
      }
      saveEpisodeWorkflowContext(flowContext.value)
    }

    return 'synced' as const
  } catch (error: any) {
    ElMessage.error(error?.message || '同步脚本到项目失败')
    return 'local' as const
  }
}

const handleNavClick = (path: string) => {
  if (!path) return
  router.push(path)
}

const handleSelectLanguage = (code: LanguageCode) => {
  selectedLanguageCode.value = code
  persistWorkspace()
}

const handleRefreshSection = (section: EditableSection) => {
  if (section === 'title') {
    editableOutputs[selectedLanguageCode.value].title = selectedLanguage.value.generatedTitle
    ElMessage.success(`已为${selectedLanguage.value.label}刷新营销标题`)
  } else {
    editableOutputs[selectedLanguageCode.value].description = selectedLanguage.value.generatedDescription
    ElMessage.success(`已为${selectedLanguage.value.label}刷新营销描述`)
  }

  persistWorkspace()
}

const handleRefreshFeatures = () => {
  ElMessage.success(`已同步${selectedLanguage.value.label}的产品特点翻译`)
}

const writeClipboard = async (value: string) => {
  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value)
    return
  }

  const input = document.createElement('textarea')
  input.value = value
  input.style.position = 'fixed'
  input.style.opacity = '0'
  document.body.appendChild(input)
  input.select()
  document.execCommand('copy')
  document.body.removeChild(input)
}

const handleCopySection = async (section: EditableSection) => {
  const text =
    section === 'title'
      ? editableOutputs[selectedLanguageCode.value].title.trim() || originalTitle.value
      : editableOutputs[selectedLanguageCode.value].description.trim() || originalDescription.value

  try {
    await writeClipboard(text)
    ElMessage.success(section === 'title' ? '标题内容已复制' : '商品描述已复制')
  } catch {
    ElMessage.error('复制失败，请稍后重试')
  }
}

const handleGenerateMoreLanguages = () => {
  ElMessage.info('更多语言生成能力将在下一步与翻译工作流接通，当前先保留高保真界面。')
}

const handleSaveDraft = async () => {
  persistWorkspace()
  const result = await persistEpisodeScript()
  ElMessage.success(
    result === 'synced'
      ? '脚本与分镜草稿已保存并同步到项目'
      : '脚本与分镜草稿已保存，当前项目将继续使用本地阶段草稿'
  )
}

const handleBeginCreation = async () => {
  persistWorkspace()
  await persistEpisodeScript()

  const context = await ensureEpisodeContext()
  if (context) {
    router.push(buildEpisodeStagePath('generation', context))
    return
  }

  ElMessage.info('内容创作流程将在剧集上下文接入后自动跳转，当前先保留页面与草稿状态。')
}

onMounted(async () => {
  restoreSourceDrafts()
  restoreWorkspace()
  const context = await ensureEpisodeContext()

  if (context?.dramaId) {
    try {
      const drama = dramaRecord.value && String(dramaRecord.value.id) === context.dramaId
        ? dramaRecord.value
        : await dramaAPI.get(context.dramaId)
      dramaRecord.value = drama

      const episode =
        drama.episodes?.find(
          (item) => item.episode_number === context.episodeNumber || String(item.id) === context.episodeId
        ) || null

      applyRemoteDraft(drama, episode)
    } catch (error: any) {
      ElMessage.warning(error?.message || '加载项目上下文失败，已回退到本地草稿')
    }
  }

  persistWorkspace()
})
</script>

<style scoped>
.page-container.script-storyboard-page {
  --stage-heading-font: 'Urbanist', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  --stage-body-font: 'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  min-height: var(--app-vh, 100vh);
  padding: 0 !important;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%) !important;
  color: #0a2463;
  font-family: var(--stage-body-font);
}

.stage-header {
  position: fixed;
  inset: 0 0 auto;
  z-index: 20;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.stage-header__inner {
  width: min(100%, 1075px);
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.stage-header__left {
  display: flex;
  align-items: center;
  gap: 32px;
  flex: 1;
  min-width: 0;
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
  background:
    linear-gradient(
      135deg,
      #0a2463 0%,
      #093873 7.1429%,
      #074c83 14.2857%,
      #066193 21.4286%,
      #0575a3 28.5714%,
      #048ab3 35.7143%,
      #05a0c3 42.8571%,
      #06b6d4 50%,
      #3aa8d9 57.1429%,
      #4f99dd 64.2857%,
      #5e8ae1 71.4286%,
      #6879e4 78.5714%,
      #7168e7 85.7143%,
      #7753ea 92.8571%,
      #7c3aed 100%
    );
}

.brand-link__mark img {
  width: 20px;
  height: 20px;
  display: block;
}

.brand-link__name {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  color: #0a2463;
  white-space: nowrap;
}

.stage-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.stage-nav::-webkit-scrollbar {
  display: none;
}

.stage-nav__item {
  height: 32px;
  padding: 0 12px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #45556c;
  font-size: 14px;
  line-height: 20px;
  font-weight: 400;
  white-space: nowrap;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.stage-nav__item:hover {
  color: #0a2463;
  background: rgba(241, 245, 249, 0.9);
}

.stage-nav__item--active {
  color: #0a2463;
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
}

.stage-header__right {
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
  background: #f97316;
}


.script-stage-main {
  width: min(100%, 1075px);
  margin: 0 auto;
  padding: 104px 32px 48px;
}

.script-stage-shell {
  width: min(100%, 1011px);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.script-stage-hero {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.script-stage-hero h1 {
  font-family: var(--stage-heading-font);
  font-size: 30px;
  line-height: 36px;
  font-weight: 700;
  color: #0a2463;
}

.script-stage-hero p {
  font-size: 16px;
  line-height: 24px;
  color: #45556c;
}

.script-stage-grid {
  display: grid;
  grid-template-columns: 228.75px minmax(0, 750.25px);
  gap: 32px;
  align-items: start;
}

.language-panel {
  min-height: 732px;
  padding: 25px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.language-panel__title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.language-panel__icon {
  width: 24px;
  height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #06b6d4;
}

.language-panel__icon svg {
  width: 24px;
  height: 24px;
  stroke: currentColor;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.language-panel__title h2 {
  font-family: var(--stage-heading-font);
  font-size: 18px;
  line-height: 28px;
  font-weight: 600;
  color: #0a2463;
}

.language-panel__list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.language-card {
  min-height: 102px;
  padding: 16px;
  border: 2px solid transparent;
  border-radius: 16px;
  background: #f8fafc;
  text-align: left;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.language-card:hover {
  transform: translateY(-1px);
}

.language-card--active {
  border-color: #06b6d4;
  background: linear-gradient(150.29deg, rgba(6, 182, 212, 0.1) 0%, rgba(124, 58, 237, 0.1) 100%);
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.1);
}

.language-card__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.language-card__label-group {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.language-card__flag {
  font-size: 24px;
  line-height: 32px;
}

.language-card__label {
  font-size: 14px;
  line-height: 20px;
  font-weight: 400;
  color: #0a2463;
  white-space: nowrap;
}

.language-card__check {
  width: 20px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #10b981;
  flex-shrink: 0;
}

.language-card__check svg {
  width: 20px;
  height: 20px;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.language-card__progress-track {
  width: 100%;
  height: 6px;
  margin-top: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: #e2e8f0;
}

.language-card__progress-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background:
    linear-gradient(
      90deg,
      #06b6d4 0%,
      #2aafd6 7.1429%,
      #3aa8d9 14.2857%,
      #46a0db 21.4286%,
      #4f99dd 28.5714%,
      #5791df 35.7143%,
      #5e8ae1 42.8571%,
      #6382e2 50%,
      #6879e4 57.1429%,
      #6d71e6 64.2857%,
      #7168e7 71.4286%,
      #745ee9 78.5714%,
      #7753ea 85.7143%,
      #7a48ec 92.8571%,
      #7c3aed 100%
    );
}

.language-card__progress-text {
  display: block;
  margin-top: 8px;
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.generate-language-button {
  width: 100%;
  min-height: 48px;
  border: none;
  border-radius: 16px;
  background:
    linear-gradient(
      90deg,
      #06b6d4 0%,
      #2aafd6 7.1429%,
      #3aa8d9 14.2857%,
      #46a0db 21.4286%,
      #4f99dd 28.5714%,
      #5791df 35.7143%,
      #5e8ae1 42.8571%,
      #6382e2 50%,
      #6879e4 57.1429%,
      #6d71e6 64.2857%,
      #7168e7 71.4286%,
      #745ee9 78.5714%,
      #7753ea 85.7143%,
      #7a48ec 92.8571%,
      #7c3aed 100%
    );
  color: #ffffff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.generate-language-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 16px 28px -18px rgba(12, 27, 78, 0.28);
}

.generate-language-button__icon {
  width: 20px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.generate-language-button__icon svg {
  width: 20px;
  height: 20px;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.script-stage-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.stage-card {
  width: 100%;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #ffffff;
}

.stage-card--title {
  min-height: 276px;
  padding: 24px;
}

.stage-card--description {
  min-height: 348px;
  padding: 24px;
}

.stage-card--features {
  min-height: 286px;
  padding: 25px;
}

.stage-card--preview {
  min-height: 210px;
  padding: 25px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.stage-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 36px;
}

.stage-card__header h3,
.stage-card--preview h3 {
  font-family: var(--stage-heading-font);
  font-size: 18px;
  line-height: 28px;
  font-weight: 600;
  color: #0a2463;
}

.stage-card__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.plain-icon-button {
  width: 36px;
  height: 36px;
  padding: 8px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: #45556c;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.plain-icon-button:hover {
  background: #f8fafc;
  color: #0a2463;
}

.plain-icon-button svg {
  width: 20px;
  height: 20px;
  display: block;
  stroke: currentColor;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.origin-card {
  margin-top: 16px;
  border-radius: 16px;
  background: #f8fafc;
  padding: 16px;
  overflow: hidden;
}

.origin-card--title {
  min-height: 80px;
}

.origin-card--description {
  min-height: 104px;
}

.origin-card__label,
.feature-column__label {
  display: block;
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.origin-card__text {
  margin-top: 8px;
  font-size: 16px;
  line-height: 24px;
  color: #314158;
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.origin-card__text--single {
  -webkit-line-clamp: 1;
  white-space: nowrap;
}

.editable-panel {
  margin-top: 12px;
  border: 2px solid rgba(6, 182, 212, 0.2);
  border-radius: 16px;
  background: linear-gradient(173.806deg, rgba(6, 182, 212, 0.05) 0%, rgba(124, 58, 237, 0.05) 100%);
  overflow: hidden;
}

.editable-panel--title {
  min-height: 76px;
}

.editable-panel--description {
  min-height: 124px;
  background: linear-gradient(169.958deg, rgba(6, 182, 212, 0.05) 0%, rgba(124, 58, 237, 0.05) 100%);
}

.editable-panel textarea {
  width: 100%;
  min-height: inherit;
  padding: 16px 18px;
  border: none;
  background: transparent;
  color: #0a2463;
  font-size: 16px;
  line-height: 24px;
  resize: none;
  outline: none;
  font-family: var(--stage-body-font);
}

.editable-panel textarea::placeholder {
  color: transparent;
}

.feature-columns {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.feature-column {
  min-height: 184px;
}

.feature-list {
  list-style: none;
  margin: 8px 0 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.feature-list li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.feature-list li span:last-child {
  font-size: 14px;
  line-height: 20px;
}

.feature-list--origin li span:last-child {
  color: #314158;
  font-weight: 400;
}

.feature-list__dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: #90a1b9;
  flex-shrink: 0;
  margin-top: 7px;
}

.feature-list--translated li span:last-child {
  color: #0a2463;
  font-weight: 500;
}

.feature-list__check {
  width: 16px;
  height: 16px;
  border: 1.5px solid #10b981;
  border-radius: 999px;
  color: #10b981;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  line-height: 1;
  font-weight: 700;
  flex-shrink: 0;
  margin-top: 2px;
}

.insight-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 24px;
}

.insight-card {
  min-height: 142px;
  padding: 25px;
  border: 1px solid #e9d4ff;
  border-radius: 16px;
  background: linear-gradient(158.642deg, #faf5ff 0%, #fdf2f8 100%);
}

.insight-card__header {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 40px;
}

.insight-card__icon {
  width: 40px;
  height: 40px;
  border-radius: 16px;
  color: #ffffff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(
      135deg,
      #7c3aed 0%,
      #833fe7 7.1429%,
      #8b43e2 14.2857%,
      #9346dc 21.4286%,
      #9b48d7 28.5714%,
      #a24ad1 35.7143%,
      #aa4ccb 42.8571%,
      #b24dc5 50%,
      #bb4ebf 57.1429%,
      #c34eb9 64.2857%,
      #cb4eb3 71.4286%,
      #d34dad 78.5714%,
      #db4ca6 85.7143%,
      #e44aa0 92.8571%,
      #ec4899 100%
    );
}

.insight-card__icon svg {
  width: 20px;
  height: 20px;
  stroke: currentColor;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.insight-card__header h4 {
  font-family: var(--stage-heading-font);
  font-size: 16px;
  line-height: 24px;
  font-weight: 600;
  color: #0a2463;
}

.insight-card p {
  margin-top: 12px;
  max-width: 314px;
  font-size: 14px;
  line-height: 20px;
  color: #314158;
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.storyboard-preview {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.storyboard-preview__card {
  min-height: 116px;
  border-radius: 16px;
  background: linear-gradient(144.573deg, #f1f5f9 0%, #e2e8f0 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  text-align: center;
}

.storyboard-preview__emoji {
  font-size: 30px;
  line-height: 36px;
  color: #0f172a;
}

.storyboard-preview__card strong {
  font-size: 14px;
  line-height: 20px;
  font-weight: 400;
  color: #0a2463;
}

.storyboard-preview__card span:last-child {
  font-size: 12px;
  line-height: 16px;
  color: #62748e;
}

.stage-actions {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding-top: 16px;
}

.stage-actions button {
  min-height: 60px;
  border: none;
  border-radius: 16px;
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  font-family: var(--stage-body-font);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.stage-actions button:hover {
  transform: translateY(-1px);
}

.stage-actions__primary {
  width: 474.25px;
  padding: 18px 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #ffffff;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
}

.stage-actions__primary img {
  width: 20px;
  height: 20px;
  display: block;
}

.stage-actions__secondary {
  width: 116px;
  background: #ffffff;
  border: 2px solid #e2e8f0 !important;
  color: #0a2463;
}

.stage-actions__ghost {
  width: 128px;
  background: #f1f5f9;
  color: #0a2463;
}

@media (max-width: 1080px) {
  .script-stage-main {
    padding-left: 24px;
    padding-right: 24px;
  }

  .script-stage-grid {
    grid-template-columns: 220px minmax(0, 1fr);
  }
}

@media (max-width: 980px) {
  .script-stage-grid {
    grid-template-columns: 1fr;
  }

  .language-panel {
    min-height: auto;
  }

  .stage-actions {
    flex-wrap: wrap;
  }

  .stage-actions__primary {
    width: 100%;
  }
}

@media (max-width: 720px) {
  .stage-header__inner {
    min-height: 64px;
    height: auto;
    padding: 10px 16px;
    flex-direction: column;
    align-items: flex-start;
  }

  .stage-header__left,
  .stage-header__right {
    width: 100%;
  }

  .stage-header__right {
    justify-content: space-between;
  }

  .script-stage-main {
    padding: 132px 16px 40px;
  }

  .feature-columns,
  .insight-row,
  .storyboard-preview {
    grid-template-columns: 1fr;
  }

  .stage-card--title,
  .stage-card--description,
  .stage-card--features,
  .stage-card--preview,
  .language-panel,
  .insight-card {
    padding: 20px;
  }

  .stage-actions__secondary,
  .stage-actions__ghost {
    width: 100%;
  }
}
</style>
