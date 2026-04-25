<template>
  <div class="home-page">
    <div class="home-page__backdrop" aria-hidden="true">
      <span class="home-page__glow home-page__glow--blue"></span>
      <span class="home-page__glow home-page__glow--violet"></span>
      <span class="home-page__glow home-page__glow--orange"></span>
      <span class="home-page__mesh"></span>
    </div>

    <div class="home-shell">
      <header class="home-header">
        <button type="button" class="brand-lockup" @click="scrollToHero">
          <span class="brand-lockup__mark">
            <img src="/logo_circle.png" alt="" class="brand-lockup__logo" />
          </span>
          <span class="brand-lockup__copy">
            <strong>{{ t('app.name') }}</strong>
            <small>Digital Silk Road</small>
          </span>
        </button>

        <div class="home-header__actions">
          <LanguageSwitcher />
          <el-button class="home-header__cta" type="primary" @click="handleCreate">
            {{ t('drama.homePrimaryAction') }}
          </el-button>
        </div>
      </header>

      <main class="home-main">
        <section ref="heroSection" class="hero-section section-anchor">
          <div class="hero-section__inner">
            <div class="hero-section__badge">{{ t('drama.homeKicker') }}</div>

            <h1 class="hero-section__title">
              <span class="hero-section__title-top">{{ t('app.name') }}</span>
              <span class="hero-section__title-bottom">{{ t('drama.homeHeadline') }}</span>
            </h1>

            <p class="hero-section__description">
              <span>{{ heroDescriptionLines.lead }}</span>
              <span v-if="heroDescriptionLines.tail" class="hero-section__description-tail">
                {{ heroDescriptionLines.tail }}
              </span>
            </p>

            <div class="hero-section__actions">
              <el-button class="hero-button hero-button--primary" type="primary" size="large" @click="handleCreate">
                {{ t('drama.homePrimaryAction') }}
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>

          <button
            v-if="latestProject"
            type="button"
            class="hero-focus"
            @click="viewDrama(latestProject.id)"
          >
            <div class="hero-focus__top">
              <span class="hero-focus__label">{{ t('drama.overviewTitle') }}</span>
              <span class="status-pill" :class="`status-pill--${getStatusTone(latestProject.status)}`">
                {{ getStatusLabel(latestProject.status) }}
              </span>
            </div>
            <div class="hero-focus__body">
              <div class="hero-focus__main">
                <h2>{{ latestProject.title }}</h2>
                <p>{{ getProjectSummary(latestProject) }}</p>
              </div>
              <div class="hero-focus__meta">
                <strong>{{ t('drama.homeEpisodeCount', { count: latestProject.total_episodes || 0 }) }}</strong>
                <span>{{ formatDate(latestProject.updated_at) }}</span>
              </div>
            </div>
          </button>

          <div v-else class="hero-focus hero-focus--empty">
            <h2>{{ t('drama.emptyStateTitle') }}</h2>
          </div>
        </section>

        <section v-if="loading || dramas.length > 0" ref="projectsSection" class="projects-section section-anchor">
          <div class="section-header">
            <div class="section-copy">
              <h2>{{ t('drama.homeShelfTitle') }}</h2>
              <p>{{ t('drama.homeLedgerDescription') }}</p>
            </div>

            <el-button class="section-header__button" type="primary" @click="handleCreate">
              <el-icon><Plus /></el-icon>
              {{ t('drama.createNew') }}
            </el-button>
          </div>

          <div v-loading="loading" class="project-grid" :class="{ 'project-grid--empty': !loading && dramas.length === 0 }">
            <template v-if="loading || dramas.length > 0">
              <article
                v-for="(drama, index) in dramas"
                :key="drama.id"
                class="project-card"
                :class="`project-card--${getStatusTone(drama.status)}`"
                tabindex="0"
                @click="viewDrama(drama.id)"
                @keydown.enter="viewDrama(drama.id)"
                @keydown.space.prevent="viewDrama(drama.id)"
              >
                <div class="project-card__top">
                  <span class="project-card__index">{{ formatSequence(index) }}</span>
                  <div class="project-card__actions" @click.stop>
                    <el-button circle text :icon="Edit" @click.stop="editDrama(drama.id)" />
                    <el-popconfirm
                      :title="$t('drama.deleteConfirm')"
                      :confirm-button-text="$t('common.confirm')"
                      :cancel-button-text="$t('common.cancel')"
                      @confirm="deleteDrama(drama.id)"
                    >
                      <template #reference>
                        <el-button circle text :icon="Delete" @click.stop />
                      </template>
                    </el-popconfirm>
                  </div>
                </div>

                <div class="project-card__status-row">
                  <span class="status-pill" :class="`status-pill--${getStatusTone(drama.status)}`">
                    {{ getStatusLabel(drama.status) }}
                  </span>
                  <span class="project-card__episodes">
                    {{ t('drama.homeEpisodeCount', { count: drama.total_episodes || 0 }) }}
                  </span>
                </div>

                <h3>{{ drama.title }}</h3>
                <p>{{ getProjectSummary(drama) }}</p>

                <div
                  v-if="getProjectCountries(drama.target_country).length > 0"
                  class="project-card__countries"
                >
                  <span
                    v-for="country in getProjectCountries(drama.target_country)"
                    :key="country"
                    class="country-chip"
                  >
                    {{ formatCountryLabel(country) }}
                  </span>
                </div>

                <div class="project-card__meta">
                  <div class="project-card__meta-item">
                    <span>{{ t('common.updatedAt') }}</span>
                    <strong>{{ formatDate(drama.updated_at) }}</strong>
                  </div>
                  <div class="project-card__meta-item">
                    <span>{{ t('drama.targetCountry') }}</span>
                    <strong>{{ getProjectCountries(drama.target_country).length }}</strong>
                  </div>
                </div>

                <div class="project-card__footer">
                  <span>{{ t('drama.homeCardFootnote') }}</span>
                  <span class="project-card__cta">
                    {{ t('common.view') }}
                    <el-icon><ArrowRight /></el-icon>
                  </span>
                </div>
              </article>
            </template>
          </div>

          <div v-if="total > 0" class="pagination-panel">
            <div class="pagination-panel__summary">
              <span>{{ t('drama.totalProjects', { count: total }) }}</span>
            </div>
            <div class="pagination-panel__controls">
              <el-pagination
                v-model:current-page="queryParams.page"
                v-model:page-size="queryParams.page_size"
                :total="total"
                :page-sizes="[12, 24, 36, 48]"
                :pager-count="5"
                layout="prev, pager, next"
                @size-change="loadDramas"
                @current-change="loadDramas"
              />
              <div class="pagination-panel__size">
                <span>{{ t('common.perPage') }}</span>
                <el-select v-model="queryParams.page_size" size="small" @change="loadDramas">
                  <el-option :value="12" label="12" />
                  <el-option :value="24" label="24" />
                  <el-option :value="36" label="36" />
                  <el-option :value="48" label="48" />
                </el-select>
              </div>
            </div>
          </div>
        </section>

        <BeianFooter class="page-beian" />
      </main>

      <el-dialog
        v-model="editDialogVisible"
        :title="$t('drama.editProject')"
        width="640px"
        :close-on-click-modal="false"
        class="edit-dialog dialog-form-safe"
      >
        <el-form
          ref="editFormRef"
          :model="editForm"
          :rules="editRules"
          label-position="top"
          v-loading="editLoading"
          class="edit-form long-form form-enter-flow"
          @submit.prevent="saveEdit"
          @keydown.enter="handleFormEnterNavigation"
        >
          <el-form-item :label="$t('drama.projectName')" prop="title" required>
            <el-input
              v-model="editForm.title"
              :placeholder="$t('drama.projectNamePlaceholder')"
              size="large"
              maxlength="50"
              show-word-limit
            />
          </el-form-item>

          <el-form-item :label="$t('drama.projectDesc')" prop="description" required>
            <el-input
              v-model="editForm.description"
              type="textarea"
              :rows="4"
              :placeholder="$t('drama.projectDescPlaceholder')"
              maxlength="500"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <el-form-item :label="$t('drama.targetCountry')" prop="target_country" required>
            <el-select
              v-model="editForm.target_country"
              size="large"
              multiple
              filterable
              :reserve-keyword="false"
              :filter-method="handleEditCountryFilter"
              @change="handleEditCountryChange"
              @visible-change="handleEditCountryVisibleChange"
              :placeholder="$t('drama.targetCountryPlaceholder')"
              :class="['country-select', { 'has-value': (editForm.target_country?.length || 0) > 0 }]"
            >
              <el-option
                v-for="country in filteredEditCountries"
                :key="country.code"
                :label="country.label"
                :value="country.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="$t('drama.materialComposition')" prop="material_composition">
            <el-input
              v-model="editForm.material_composition"
              type="textarea"
              :rows="3"
              :placeholder="$t('drama.materialCompositionPlaceholder')"
              maxlength="200"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <el-form-item :label="$t('drama.marketingSellingPoints')" prop="marketing_selling_points">
            <el-input
              v-model="editForm.marketing_selling_points"
              type="textarea"
              :rows="3"
              :placeholder="$t('drama.marketingSellingPointsPlaceholder')"
              maxlength="200"
              show-word-limit
              resize="none"
            />
          </el-form-item>
        </el-form>

        <template #footer>
          <div class="dialog-footer">
            <el-button size="large" @click="editDialogVisible = false">
              {{ $t('common.cancel') }}
            </el-button>
            <el-button type="primary" size="large" :loading="editLoading" @click="saveEdit">
              {{ $t('common.save') }}
            </el-button>
          </div>
        </template>
      </el-dialog>

      <CreateDramaDialog v-model="createDialogVisible" @created="loadDramas" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { ArrowRight, Delete, Edit, Plus } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import type { Drama, DramaListQuery } from '@/types/drama'
import CreateDramaDialog from '@/components/common/CreateDramaDialog.vue'
import BeianFooter from '@/components/common/BeianFooter.vue'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import { handleFormEnterNavigation } from '@/utils/formFocus'
import { ALL_COUNTRIES } from '@/constants/countries'

const router = useRouter()
const { t, locale } = useI18n()

const heroSection = ref<HTMLElement | null>(null)
const projectsSection = ref<HTMLElement | null>(null)

const loading = ref(false)
const dramas = ref<Drama[]>([])
const total = ref(0)

const queryParams = ref<DramaListQuery>({
  page: 1,
  page_size: 12
})

const createDialogVisible = ref(false)

const countryCodeSet = new Set(ALL_COUNTRIES.map((item) => item.value))
const countryLabelMap = new Map(ALL_COUNTRIES.map((item) => [item.value, item.name]))

const normalizeTargetCountries = (value: string | string[] | undefined): string[] => {
  const rawList = Array.isArray(value)
    ? value
    : String(value || '')
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item.length > 0)

  const normalized: string[] = []
  for (const item of rawList) {
    const trimmed = item.trim()
    if (!trimmed) continue

    let code = trimmed.toUpperCase()
    if (!countryCodeSet.has(code)) {
      const bracketMatch = trimmed.match(/\(([A-Za-z]{2})\)/)
      const tailCodeMatch = trimmed.match(/\b([A-Za-z]{2})\b$/)
      code = (bracketMatch?.[1] || tailCodeMatch?.[1] || '').toUpperCase()
    }

    if (countryCodeSet.has(code) && !normalized.includes(code)) {
      normalized.push(code)
    }
  }
  return normalized
}

const getProjectCountries = (value: string | string[] | undefined) => normalizeTargetCountries(value).slice(0, 3)

const formatCountryLabel = (code: string) => countryLabelMap.get(code) || code

const formatDate = (value?: string) => {
  if (!value) return t('drama.homeNoRecentUpdate')

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  return new Intl.DateTimeFormat(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

const getProjectSummary = (drama: Drama) => {
  return drama.description || drama.marketing_selling_points || drama.material_composition || t('drama.noDescription')
}

const heroDescriptionLines = computed(() => {
  const description = t('drama.homeDescription').trim()

  if (locale.value === 'zh-CN') {
    const tail = 'AI 解决方案。'
    if (description.includes(tail)) {
      return {
        lead: description.replace(tail, '').trim(),
        tail
      }
    }
  }

  return {
    lead: description,
    tail: ''
  }
})

const latestProject = computed(() => {
  return [...dramas.value].sort((a, b) => {
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
  })[0] || null
})

const pageStart = computed(() => ((queryParams.value.page || 1) - 1) * (queryParams.value.page_size || 12))

const formatSequence = (index: number) => String(pageStart.value + index + 1).padStart(2, '0')

const getStatusTone = (status?: string) => {
  if (status === 'completed') return 'completed'
  if (status === 'planning' || status === 'production' || status === 'generating') return 'live'
  return 'draft'
}

const getStatusLabel = (status?: string) => {
  switch (status) {
    case 'planning':
    case 'production':
      return t('drama.status.production')
    case 'generating':
      return t('common.generating')
    case 'completed':
      return t('drama.status.completed')
    case 'draft':
    default:
      return t('drama.status.draft')
  }
}

const scrollToHero = () => {
  heroSection.value?.scrollIntoView({
    behavior: 'smooth',
    block: 'start'
  })
}

const loadDramas = async () => {
  loading.value = true
  try {
    const result = await dramaAPI.list(queryParams.value)
    dramas.value = result.items || []
    total.value = result.pagination?.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || t('drama.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  createDialogVisible.value = true
}

const viewDrama = (id: string) => router.push(`/dramas/${id}`)

const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editCountryKeyword = ref('')
const editForm = ref({
  id: '',
  title: '',
  description: '',
  target_country: [] as string[],
  material_composition: '',
  marketing_selling_points: ''
})

const filteredEditCountries = computed(() => {
  const keyword = editCountryKeyword.value.trim().toLowerCase()
  if (!keyword) return ALL_COUNTRIES
  return ALL_COUNTRIES.filter((country) => country.searchText.includes(keyword))
})

const handleEditCountryFilter = (keyword: string) => {
  editCountryKeyword.value = keyword
}

const handleEditCountryVisibleChange = (visible: boolean) => {
  if (!visible) editCountryKeyword.value = ''
}

const handleEditCountryChange = () => {
  editCountryKeyword.value = ''
}

const editRules = computed<FormRules>(() => ({
  title: [
    { required: true, message: t('validation.projectNameRequired'), trigger: 'blur' },
    { min: 1, max: 50, message: t('validation.projectNameLength'), trigger: 'blur' }
  ],
  description: [
    { required: true, message: t('validation.projectDescRequired'), trigger: 'blur' },
    { min: 1, max: 500, message: t('validation.projectDescLength'), trigger: 'blur' }
  ],
  target_country: [
    { type: 'array', required: true, min: 1, message: t('validation.targetCountryRequired'), trigger: 'change' }
  ],
  material_composition: [
    { max: 200, message: t('validation.materialLength'), trigger: 'blur' }
  ],
  marketing_selling_points: [
    { max: 200, message: t('validation.marketingLength'), trigger: 'blur' }
  ]
}))

const editDrama = async (id: string) => {
  editLoading.value = true
  editDialogVisible.value = true
  try {
    const drama = await dramaAPI.get(id)
    editForm.value = {
      id: drama.id,
      title: drama.title,
      description: drama.description || '',
      target_country: normalizeTargetCountries(drama.target_country as string | string[] | undefined),
      material_composition: drama.material_composition || '',
      marketing_selling_points: drama.marketing_selling_points || ''
    }
  } catch (error: any) {
    ElMessage.error(error.message || t('drama.messages.loadFailed'))
    editDialogVisible.value = false
  } finally {
    editLoading.value = false
  }
}

const saveEdit = async () => {
  if (!editFormRef.value) return

  const valid = await editFormRef.value.validate().then(() => true).catch(() => false)
  if (!valid) return

  editLoading.value = true
  try {
    await dramaAPI.update(editForm.value.id, {
      title: editForm.value.title.trim(),
      description: editForm.value.description.trim(),
      target_country: editForm.value.target_country.map((item) => item.trim()).filter((item) => item.length > 0),
      material_composition: editForm.value.material_composition.trim(),
      marketing_selling_points: editForm.value.marketing_selling_points.trim()
    })
    ElMessage.success(t('drama.messages.saved'))
    editDialogVisible.value = false
    loadDramas()
  } catch (error: any) {
    ElMessage.error(error.message || t('drama.messages.saveFailed'))
  } finally {
    editLoading.value = false
  }
}

const deleteDrama = async (id: string) => {
  try {
    await dramaAPI.delete(id)
    ElMessage.success(t('drama.messages.deleted'))
    loadDramas()
  } catch (error: any) {
    ElMessage.error(error.message || t('drama.messages.deleteFailed'))
  }
}

onMounted(() => {
  loadDramas()
})
</script>

<style scoped>
.home-page {
  --home-bg: #f8fbff;
  --home-surface: rgba(255, 255, 255, 0.82);
  --home-surface-strong: rgba(255, 255, 255, 0.92);
  --home-border: rgba(116, 139, 168, 0.2);
  --home-shadow: 0 32px 80px -56px rgba(34, 62, 109, 0.35);
  --home-text: #18366d;
  --home-muted: #54709b;
  --home-blue: #34b7e8;
  --home-indigo: #5568ef;
  --home-violet: #8b5cf6;
  --home-orange: #ff8a26;
  --home-orange-strong: #ff6c1e;
  position: relative;
  min-height: 100%;
  background:
    radial-gradient(circle at 10% 18%, rgba(102, 173, 255, 0.18), transparent 28%),
    radial-gradient(circle at 78% 22%, rgba(255, 182, 111, 0.2), transparent 24%),
    radial-gradient(circle at 82% 88%, rgba(129, 140, 248, 0.16), transparent 20%),
    linear-gradient(180deg, #fdfefe 0%, #f7fbff 42%, #f3f8ff 100%);
  color: var(--home-text);
  overflow: hidden;
}

.dark .home-page {
  --home-bg: #0e1528;
  --home-surface: rgba(12, 21, 40, 0.8);
  --home-surface-strong: rgba(15, 24, 46, 0.92);
  --home-border: rgba(130, 161, 214, 0.18);
  --home-shadow: 0 40px 100px -62px rgba(0, 0, 0, 0.7);
  --home-text: #eef4ff;
  --home-muted: #adc0e7;
  background:
    radial-gradient(circle at 12% 20%, rgba(52, 183, 232, 0.16), transparent 28%),
    radial-gradient(circle at 78% 22%, rgba(255, 138, 38, 0.14), transparent 24%),
    radial-gradient(circle at 82% 88%, rgba(139, 92, 246, 0.16), transparent 20%),
    linear-gradient(180deg, #0e1426 0%, #0b1120 42%, #0d1629 100%);
}

.home-page__backdrop {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.home-page__glow {
  position: absolute;
  border-radius: 999px;
  filter: blur(18px);
  opacity: 0.95;
}

.home-page__glow--blue {
  top: -180px;
  left: -120px;
  width: 460px;
  height: 460px;
  background: radial-gradient(circle, rgba(52, 183, 232, 0.22) 0%, rgba(52, 183, 232, 0) 74%);
}

.home-page__glow--violet {
  top: 140px;
  right: 12%;
  width: 420px;
  height: 420px;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.14) 0%, rgba(139, 92, 246, 0) 74%);
}

.home-page__glow--orange {
  bottom: -180px;
  right: -100px;
  width: 520px;
  height: 520px;
  background: radial-gradient(circle, rgba(255, 138, 38, 0.16) 0%, rgba(255, 138, 38, 0) 72%);
}

.home-page__mesh {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(139, 160, 196, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(139, 160, 196, 0.07) 1px, transparent 1px);
  background-size: 96px 96px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.34), transparent 82%);
}

.dark .home-page__mesh {
  background-image:
    linear-gradient(rgba(139, 160, 196, 0.07) 1px, transparent 1px),
    linear-gradient(90deg, rgba(139, 160, 196, 0.06) 1px, transparent 1px);
}

.home-shell {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  max-width: 1480px;
  margin: 0 auto;
  padding: 24px 24px 16px;
}

.home-header {
  position: sticky;
  top: 18px;
  z-index: 12;
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: center;
  gap: 24px;
  padding: 16px 24px;
  border: 1px solid var(--home-border);
  border-radius: 28px;
  background: var(--home-surface);
  box-shadow: var(--home-shadow);
  backdrop-filter: blur(22px);
}

.brand-lockup {
  display: inline-flex;
  align-items: center;
  gap: 16px;
  padding: 0;
  border: none;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.brand-lockup__mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(52, 183, 232, 0.18), rgba(139, 92, 246, 0.18));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.brand-lockup__logo {
  width: 84%;
  height: 84%;
  object-fit: contain;
}

.brand-lockup__copy {
  display: grid;
  gap: 4px;
}

.brand-lockup__copy strong {
  font-size: 2rem;
  font-weight: 800;
  letter-spacing: -0.04em;
  color: var(--home-text);
}

.brand-lockup__copy small {
  font-size: 1rem;
  color: var(--home-muted);
}

.home-header__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

.home-page :deep(.language-switcher) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  border-radius: 18px;
  border: 1px solid var(--home-border);
  background: rgba(255, 255, 255, 0.72);
  color: var(--home-text);
  box-shadow: 0 16px 36px -28px rgba(34, 62, 109, 0.34);
}

.home-page :deep(.language-switcher) {
  gap: 10px;
  padding: 0 16px;
}

.home-page :deep(.lang-text) {
  color: var(--home-text);
  font-weight: 700;
}

.dark .home-page :deep(.language-switcher) {
  background: rgba(18, 29, 54, 0.78);
}

.home-header__cta {
  min-height: 48px;
  padding-inline: 22px;
  border: none;
  border-radius: 18px;
  background: linear-gradient(135deg, var(--home-orange) 0%, var(--home-orange-strong) 100%);
  box-shadow: 0 20px 36px -24px rgba(255, 108, 30, 0.5);
}

.home-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 28px;
  padding-top: 28px;
}

.section-anchor {
  scroll-margin-top: 120px;
}

.hero-section,
.projects-section {
  position: relative;
  border: 1px solid var(--home-border);
  border-radius: 40px;
  background: var(--home-surface);
  box-shadow: var(--home-shadow);
  backdrop-filter: blur(24px);
  overflow: hidden;
}

.hero-section {
  padding: 40px 40px 30px;
  text-align: center;
}

.hero-section::before,
.projects-section::before {
  content: '';
  position: absolute;
  inset: 1px;
  border-radius: inherit;
  border: 1px solid rgba(255, 255, 255, 0.46);
  pointer-events: none;
}

.dark .hero-section::before,
.dark .projects-section::before {
  border-color: rgba(255, 255, 255, 0.04);
}

.hero-section__inner {
  position: relative;
  z-index: 1;
  max-width: 880px;
  margin: 0 auto;
}

.hero-section__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  padding: 12px 24px;
  border: 1px solid rgba(85, 104, 239, 0.18);
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(52, 183, 232, 0.1), rgba(139, 92, 246, 0.08));
  color: var(--home-indigo);
  font-size: 1rem;
  font-weight: 700;
}

.hero-section__title {
  display: grid;
  gap: 10px;
  margin: 0;
  letter-spacing: -0.06em;
}

.hero-section__title-top {
  font-size: clamp(4.2rem, 10vw, 7rem);
  font-weight: 900;
  line-height: 0.92;
  color: #173a7d;
}

.hero-section__title-bottom {
  font-size: clamp(4rem, 10vw, 6.6rem);
  font-weight: 900;
  line-height: 0.98;
  background: linear-gradient(90deg, var(--home-blue) 0%, var(--home-violet) 48%, var(--home-orange) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.dark .hero-section__title-top {
  color: #f3f7ff;
}

.hero-section__description {
  max-width: 980px;
  margin: 28px auto 0;
  color: var(--home-muted);
  font-size: clamp(1.06rem, 1.8vw, 1.45rem);
  line-height: 1.9;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.hero-section__description-tail {
  white-space: nowrap;
}

.hero-section__actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 34px;
  flex-wrap: wrap;
}

.hero-button {
  min-width: 188px;
  min-height: 62px;
  border-radius: 22px;
  font-size: 1.1rem;
  font-weight: 800;
}

.hero-button :deep(.el-icon) {
  margin-left: 8px;
}

.hero-button--primary {
  border: none;
  background: linear-gradient(135deg, var(--home-orange) 0%, var(--home-orange-strong) 100%);
  box-shadow: 0 26px 44px -26px rgba(255, 108, 30, 0.48);
}

.projects-section {
  padding: 34px;
}

.section-copy {
  display: grid;
  align-content: start;
  gap: 14px;
}

.section-copy h2 {
  margin: 0;
  font-size: clamp(2rem, 4vw, 3rem);
  line-height: 1.12;
  letter-spacing: -0.04em;
  color: var(--home-text);
}

.section-copy p {
  margin: 0;
  color: var(--home-muted);
  font-size: 1rem;
  line-height: 1.8;
}

.hero-focus {
  display: grid;
  gap: 16px;
  align-content: start;
  max-width: 960px;
  margin: 28px auto 0;
  padding: 24px 26px;
  border: 1px solid rgba(152, 173, 206, 0.22);
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.86), rgba(245, 250, 255, 0.78)),
    radial-gradient(circle at top right, rgba(85, 104, 239, 0.08), transparent 34%);
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.hero-focus:hover {
  transform: translateY(-3px);
  box-shadow: 0 28px 56px -36px rgba(34, 62, 109, 0.38);
  border-color: rgba(85, 104, 239, 0.28);
}

.dark .hero-focus {
  background:
    linear-gradient(180deg, rgba(16, 26, 48, 0.88), rgba(11, 18, 34, 0.82)),
    radial-gradient(circle at top right, rgba(85, 104, 239, 0.16), transparent 34%);
}

.hero-focus__top,
.hero-focus__body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.hero-focus__label,
.hero-focus__meta span {
  color: var(--home-muted);
  font-size: 0.95rem;
  font-weight: 600;
}

.hero-focus__main {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.hero-focus__main h2 {
  margin: 0;
  font-size: clamp(1.5rem, 3vw, 2.2rem);
  line-height: 1.12;
  letter-spacing: -0.03em;
  color: var(--home-text);
}

.hero-focus__main p {
  margin: 0;
  color: var(--home-muted);
  line-height: 1.8;
}

.hero-focus__meta {
  display: grid;
  gap: 8px;
  justify-items: end;
  text-align: right;
}

.hero-focus__meta strong {
  color: var(--home-indigo);
  font-size: 1rem;
  font-weight: 700;
}

.hero-focus--empty {
  cursor: default;
  justify-items: center;
  text-align: center;
}

.hero-focus--empty:hover {
  transform: none;
  box-shadow: none;
  border-color: rgba(152, 173, 206, 0.22);
}

.hero-focus--empty h2 {
  margin: 0;
  font-size: 1.5rem;
  color: var(--home-text);
}

.project-card__countries {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.country-chip {
  display: inline-flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(85, 104, 239, 0.08);
  color: var(--home-indigo);
  font-size: 0.92rem;
  font-weight: 700;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 7px 12px;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 800;
}

.status-pill--draft {
  color: #5a6c8b;
  background: rgba(131, 145, 171, 0.16);
}

.status-pill--live {
  color: #177474;
  background: rgba(34, 197, 154, 0.14);
}

.status-pill--completed {
  color: #6a45d1;
  background: rgba(139, 92, 246, 0.14);
}

.section-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 24px;
}

.section-header__button {
  min-height: 48px;
  border-radius: 18px;
}

.project-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.project-grid--empty {
  grid-template-columns: minmax(0, 1fr);
}

.empty-state {
  display: grid;
  justify-items: center;
  gap: 14px;
  padding: 68px 24px;
  border: 1px dashed rgba(152, 173, 206, 0.34);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.54);
  text-align: center;
}

.empty-state__icon {
  position: relative;
  width: 104px;
  height: 72px;
}

.empty-state__icon span {
  position: absolute;
  display: block;
  width: 68px;
  height: 48px;
  border-radius: 16px;
  border: 1px solid rgba(152, 173, 206, 0.28);
  background: rgba(255, 255, 255, 0.74);
}

.empty-state__icon span:first-child {
  left: 10px;
  top: 16px;
}

.empty-state__icon span:last-child {
  right: 8px;
  top: 0;
}

.empty-state h3 {
  margin: 0;
  font-size: 1.6rem;
  color: var(--home-text);
}

.empty-state p {
  max-width: 520px;
  margin: 0;
  color: var(--home-muted);
  line-height: 1.8;
}

.project-card {
  display: grid;
  gap: 18px;
  min-width: 0;
  padding: 24px;
  border: 1px solid rgba(152, 173, 206, 0.2);
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(247, 250, 255, 0.78)),
    radial-gradient(circle at top right, rgba(85, 104, 239, 0.08), transparent 28%);
  cursor: pointer;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.project-card:hover,
.project-card:focus-visible {
  transform: translateY(-3px);
  border-color: rgba(85, 104, 239, 0.28);
  box-shadow: 0 30px 58px -36px rgba(34, 62, 109, 0.36);
  outline: none;
}

.dark .project-card {
  background:
    linear-gradient(180deg, rgba(16, 26, 48, 0.88), rgba(11, 18, 34, 0.82)),
    radial-gradient(circle at top right, rgba(85, 104, 239, 0.16), transparent 28%);
}

.project-card__top,
.project-card__status-row,
.project-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.project-card__actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.project-card__index {
  font-size: 0.98rem;
  font-weight: 800;
  color: var(--home-indigo);
}

.project-card__episodes {
  color: var(--home-muted);
  font-size: 0.92rem;
  font-weight: 700;
}

.project-card h3 {
  margin: 0;
  font-size: 1.45rem;
  line-height: 1.2;
  letter-spacing: -0.03em;
  color: var(--home-text);
}

.project-card p {
  margin: 0;
  color: var(--home-muted);
  line-height: 1.8;
}

.project-card__meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  padding-top: 4px;
}

.project-card__meta-item {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.project-card__meta-item span,
.project-card__footer span:first-child {
  color: var(--home-muted);
  font-size: 0.9rem;
}

.project-card__meta-item strong {
  color: var(--home-text);
  font-size: 0.98rem;
}

.project-card__cta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--home-orange-strong);
  font-weight: 800;
}

.pagination-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-top: 24px;
  padding: 20px 22px;
  border: 1px solid rgba(152, 173, 206, 0.22);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.66);
}

.dark .pagination-panel {
  background: rgba(16, 26, 48, 0.62);
}

.pagination-panel__summary {
  color: var(--home-text);
  font-weight: 700;
}

.pagination-panel__controls {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 14px;
  flex-wrap: wrap;
}

.pagination-panel__size {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: var(--home-muted);
  font-size: 0.92rem;
}

.pagination-panel__size :deep(.el-select) {
  width: 88px;
}

.page-beian {
  display: flex;
  justify-content: center;
  width: 100%;
  margin-top: auto;
  padding-top: 20px;
  padding-bottom: 8px;
}

@media (max-width: 1280px) {
  .home-header {
    grid-template-columns: 1fr;
    justify-items: center;
  }

  .home-header__actions {
    justify-content: center;
    flex-wrap: wrap;
  }

  .project-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .home-shell {
    padding: 16px 16px 12px;
  }

  .hero-section {
    padding: 34px 22px 22px;
  }

  .projects-section {
    padding: 24px 20px;
  }

  .project-grid,
  .project-card__meta {
    grid-template-columns: 1fr;
  }

  .section-header,
  .pagination-panel,
  .pagination-panel__controls {
    align-items: stretch;
    flex-direction: column;
  }

  .pagination-panel__summary,
  .pagination-panel__controls {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .brand-lockup {
    width: 100%;
    justify-content: center;
  }

  .brand-lockup__copy strong {
    font-size: 1.5rem;
  }

  .home-header {
    top: 12px;
    padding: 14px 16px;
    border-radius: 24px;
  }

  .home-header__actions,
  .hero-section__actions {
    width: 100%;
  }

  .home-header__cta,
  .hero-button {
    width: 100%;
  }

  .hero-section__badge {
    width: 100%;
    padding-inline: 16px;
  }

  .hero-section__description,
  .section-copy p,
  .project-card p,
  .hero-focus__main p {
    font-size: 0.96rem;
  }

  .hero-focus,
  .project-card,
  .empty-state {
    padding: 20px;
    border-radius: 24px;
  }

  .hero-focus__top,
  .hero-focus__body,
  .project-card__top,
  .project-card__status-row,
  .project-card__footer {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
