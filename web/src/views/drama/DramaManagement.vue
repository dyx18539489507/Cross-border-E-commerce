<template>
  <div class="management-page">
    <div class="management-page__backdrop" aria-hidden="true">
      <span class="management-page__glow management-page__glow--blue"></span>
      <span class="management-page__glow management-page__glow--violet"></span>
      <span class="management-page__glow management-page__glow--orange"></span>
      <span class="management-page__mesh"></span>
    </div>

    <div class="management-shell">
      <header class="workspace-topbar animate-fade-in">
        <div class="workspace-topbar__left">
          <button type="button" class="workspace-back" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
            <span>{{ $t('common.back') }}</span>
          </button>
          <div class="workspace-heading">
            <span class="workspace-marker">{{ $t('drama.currentFocusTitle') }}</span>
            <div class="workspace-heading__copy">
              <strong>{{ drama?.title || $t('drama.title') }}</strong>
              <span>{{ projectCountryText }}</span>
            </div>
          </div>
        </div>

        <div class="workspace-topbar__right">
          <div class="workspace-status">
            <span class="status-pill" :class="`status-pill--${getStatusTone(drama?.status)}`">
              {{ getStatusText(drama?.status) }}
            </span>
            <div class="workspace-status__meta">
              <span>{{ $t('common.updatedAt') }}</span>
              <strong>{{ formatDate(drama?.updated_at) }}</strong>
            </div>
          </div>

          <LanguageSwitcher />
        </div>
      </header>

      <main class="management-main" v-loading="loading">
        <section class="hero-grid animate-fade-in">
          <article class="hero-card hero-card--overview">
            <div class="hero-card__accent" aria-hidden="true"></div>

            <div class="hero-card__head">
              <span class="hero-card__eyebrow">{{ $t('drama.management.overview') }}</span>
              <span class="status-pill status-pill--hero" :class="`status-pill--${getStatusTone(drama?.status)}`">
                {{ getStatusText(drama?.status) }}
              </span>
            </div>

            <div class="hero-card__content">
              <h1>{{ drama?.title || $t('drama.title') }}</h1>
              <p>{{ projectSummary }}</p>
            </div>

            <div v-if="projectCountries.length > 0" class="project-stage__chips">
              <span
                v-for="country in projectCountries"
                :key="country"
                class="country-chip"
              >
                {{ formatCountryLabel(country) }}
              </span>
            </div>

            <div class="hero-card__actions">
              <el-button
                type="primary"
                size="large"
                :icon="Plus"
                class="hero-primary-action"
                @click="handleOverviewPrimaryAction"
              >
                {{ episodeSpotlight ? $t('drama.management.goToEdit') : $t('drama.management.createFirstEpisode') }}
              </el-button>
              <el-button
                plain
                size="large"
                :icon="EditPen"
                class="hero-secondary-action"
                @click="openEditDescription"
              >
                {{ $t('common.edit') }}
              </el-button>
            </div>

            <div class="hero-metrics">
              <article
                v-for="metric in heroMetrics"
                :key="metric.key"
                class="metric-card"
                :class="`metric-card--${metric.tone}`"
              >
                <div class="metric-card__icon">
                  <el-icon><component :is="metric.icon" /></el-icon>
                </div>
                <div class="metric-card__body">
                  <span>{{ metric.label }}</span>
                  <strong>{{ metric.value }}</strong>
                  <small>{{ metric.detail }}</small>
                </div>
              </article>
            </div>
          </article>

          <aside class="hero-side">
            <button
              type="button"
              class="focus-panel"
              @click="handleOverviewPrimaryAction"
            >
              <div class="focus-panel__head">
                <div class="focus-panel__copy">
                  <span class="focus-panel__eyebrow">
                    {{ episodeSpotlight ? $t('drama.management.episodes') : $t('drama.currentFocusTitle') }}
                  </span>
                  <h2>
                    {{
                      episodeSpotlight
                        ? getEpisodeTitle(episodeSpotlight)
                        : $t('drama.management.startFirstEpisode')
                    }}
                  </h2>
                </div>
                <span class="status-pill" :class="`status-pill--${episodeSpotlight ? getEpisodeTone(episodeSpotlight) : getStatusTone(drama?.status)}`">
                  {{ episodeSpotlight ? getEpisodeStatusText(episodeSpotlight) : getStatusText(drama?.status) }}
                </span>
              </div>

              <p>{{ episodeSpotlight ? getEpisodeSummary(episodeSpotlight) : $t('drama.management.clickToCreate') }}</p>

              <div class="focus-panel__stats">
                <article
                  v-for="stat in focusStats"
                  :key="stat.key"
                  class="focus-panel__stat"
                >
                  <span>{{ stat.label }}</span>
                  <strong>{{ stat.value }}</strong>
                </article>
              </div>
            </button>

            <div class="meta-grid">
              <article
                v-for="item in metaCards"
                :key="item.key"
                class="meta-card"
              >
                <span>{{ item.label }}</span>
                <strong :title="item.title || item.value">{{ item.value }}</strong>
              </article>
            </div>
          </aside>
        </section>

        <section class="workspace-panel animate-fade-in">
          <el-tabs v-model="activeTab" class="workspace-tabs">
            <el-tab-pane :label="$t('drama.management.overview')" name="overview">
              <div class="overview-pane">
                <div class="overview-stage">
                  <section class="overview-story">
                    <div class="overview-story__glow" aria-hidden="true"></div>

                    <div class="overview-story__header">
                      <div class="section-heading__copy section-heading__copy--light">
                        <span class="section-heading__eyebrow section-heading__eyebrow--soft">{{ $t('drama.management.projectInfo') }}</span>
                        <h2>{{ drama?.title || $t('drama.title') }}</h2>
                        <p>{{ projectCountryText }}</p>
                      </div>

                      <el-button
                        text
                        :icon="EditPen"
                        class="ghost-link ghost-link--light"
                        @click="openEditDescription"
                      >
                        {{ $t('common.edit') }}
                      </el-button>
                    </div>

                    <div class="overview-story__lead">
                      <span class="overview-story__label">{{ $t('drama.management.projectDesc') }}</span>
                      <p>{{ drama?.description || $t('drama.management.noDescription') }}</p>
                    </div>

                    <div class="overview-story__facts">
                      <article
                        v-for="fact in overviewFacts"
                        :key="fact.key"
                        class="story-fact"
                        :class="{ 'story-fact--wide': fact.wide }"
                      >
                        <span>{{ fact.label }}</span>
                        <strong :title="fact.value">{{ fact.value }}</strong>
                      </article>
                    </div>
                  </section>

                  <aside class="overview-flow">
                    <section class="overview-flow__panel">
                      <div class="section-heading section-heading--compact section-heading--light">
                        <div class="section-heading__copy section-heading__copy--light">
                          <span class="section-heading__eyebrow section-heading__eyebrow--soft">{{ $t('drama.currentFocusTitle') }}</span>
                          <h2>
                            {{
                              episodeSpotlight
                                ? getEpisodeTitle(episodeSpotlight)
                                : $t('drama.management.startFirstEpisode')
                            }}
                          </h2>
                          <p>{{ episodeSpotlight ? getEpisodeSummary(episodeSpotlight) : workflowHint }}</p>
                        </div>
                        <span class="status-pill" :class="`status-pill--${episodeSpotlight ? getEpisodeTone(episodeSpotlight) : getStatusTone(drama?.status)}`">
                          {{ episodeSpotlight ? getEpisodeStatusText(episodeSpotlight) : getStatusText(drama?.status) }}
                        </span>
                      </div>

                      <div class="signal-list">
                        <button
                          v-for="item in overviewSignals"
                          :key="item.key"
                          type="button"
                          class="signal-item"
                          :class="`signal-item--${item.tone}`"
                          @click="activateTab(item.key)"
                        >
                          <div class="signal-item__icon">
                            <el-icon><component :is="item.icon" /></el-icon>
                          </div>
                          <div class="signal-item__body">
                            <div class="signal-item__top">
                              <span>{{ item.label }}</span>
                              <strong>{{ item.count }}</strong>
                            </div>
                            <h3>{{ item.title }}</h3>
                            <p>{{ item.description }}</p>
                          </div>
                        </button>
                      </div>

                      <div class="overview-flow__footer">
                        <el-button
                          type="primary"
                          :icon="Plus"
                          class="overview-flow__action"
                          @click="handleOverviewPrimaryAction"
                        >
                          {{ episodeSpotlight ? $t('drama.management.goToEdit') : $t('drama.management.createFirstEpisode') }}
                        </el-button>

                        <div class="overview-flow__meta">
                          <span>{{ $t('common.updatedAt') }}</span>
                          <strong>{{ formatDate(drama?.updated_at) }}</strong>
                        </div>
                      </div>
                    </section>
                  </aside>
                </div>

                <el-dialog
                  v-model="editDescriptionDialogVisible"
                  :title="$t('drama.management.editDescriptionTitle')"
                  width="720px"
                  class="edit-desc-dialog dialog-form-safe"
                >
                  <el-form
                    ref="editDescriptionFormRef"
                    label-width="0"
                    class="long-form form-enter-flow"
                    @keydown.enter="handleFormEnterNavigation"
                  >
                    <el-form-item>
                      <el-input
                        v-model="editDescriptionValue"
                        type="textarea"
                        :autosize="{ minRows: 8, maxRows: 16 }"
                        :placeholder="$t('drama.management.projectDesc')"
                      />
                    </el-form-item>
                  </el-form>
                  <template #footer>
                    <el-button @click="editDescriptionDialogVisible = false">{{ $t('common.cancel') }}</el-button>
                    <el-button type="primary" :loading="savingDescription" @click="saveDescription">
                      {{ $t('common.save') }}
                    </el-button>
                  </template>
                </el-dialog>
              </div>
            </el-tab-pane>

            <el-tab-pane :label="$t('drama.management.episodes')" name="episodes">
              <div class="module-pane">
                <div class="module-header">
                  <div class="section-heading__copy">
                    <span class="section-heading__eyebrow">{{ $t('drama.management.episodeStats') }}</span>
                    <h2>{{ $t('drama.management.episodeList') }}</h2>
                    <p>{{ episodesCount > 0 ? $t('drama.management.episodesCreated') : $t('drama.management.noEpisodesYet') }}</p>
                  </div>

                  <el-button type="primary" :icon="Plus" @click="createNewEpisode">
                    {{ $t('drama.management.createNewEpisode') }}
                  </el-button>
                </div>

                <div class="module-summary">
                  <article class="mini-stat">
                    <span>{{ $t('drama.management.episodeStats') }}</span>
                    <strong>{{ episodesCount }}</strong>
                  </article>
                  <article class="mini-stat">
                    <span>{{ $t('drama.management.shotsCount') }}</span>
                    <strong>{{ totalShotCount }}</strong>
                  </article>
                  <article class="mini-stat">
                    <span>{{ $t('common.updatedAt') }}</span>
                    <strong>{{ episodeSpotlight ? formatDate(episodeSpotlight.updated_at || episodeSpotlight.created_at) : '-' }}</strong>
                  </article>
                </div>

                <div v-if="episodesCount === 0" class="compact-empty compact-empty--accent">
                  <div class="compact-empty__icon">
                    <el-icon><Document /></el-icon>
                  </div>
                  <div class="compact-empty__content">
                    <h3>{{ $t('drama.management.noEpisodes') }}</h3>
                    <p>{{ $t('drama.management.noEpisodesYet') }}</p>
                  </div>
                  <el-button type="primary" :icon="Plus" @click="createNewEpisode">
                    {{ $t('drama.management.createFirstEpisode') }}
                  </el-button>
                </div>

                <div v-else class="episode-list">
                  <article
                    v-for="episode in sortedEpisodes"
                    :key="episode.id || episode.episode_number"
                    class="episode-row"
                  >
                    <div class="episode-row__index">{{ formatEpisodeIndex(episode.episode_number) }}</div>
                    <div class="episode-row__body">
                      <div class="episode-row__title">
                        <h3>{{ episode.title || $t('drama.management.episodeNumber', { number: episode.episode_number }) }}</h3>
                        <span class="status-pill" :class="`status-pill--${getEpisodeTone(episode)}`">
                          {{ getEpisodeStatusText(episode) }}
                        </span>
                      </div>
                      <p>{{ getEpisodeSummary(episode) }}</p>
                      <div class="episode-row__meta">
                        <span>{{ $t('drama.management.shotsCount') }} {{ getEpisodeShotCount(episode) }}</span>
                        <span>{{ $t('common.createdAt') }} {{ formatDate(episode.created_at) }}</span>
                        <span>{{ $t('common.updatedAt') }} {{ formatDate(episode.updated_at || episode.created_at) }}</span>
                      </div>
                    </div>
                    <div class="episode-row__actions">
                      <el-button type="primary" @click="enterEpisodeWorkflow(episode)">
                        {{ $t('drama.management.goToEdit') }}
                      </el-button>
                      <el-button type="danger" plain @click="deleteEpisode(episode)">
                        {{ $t('common.delete') }}
                      </el-button>
                    </div>
                  </article>
                </div>
              </div>
            </el-tab-pane>

            <el-tab-pane :label="$t('drama.management.characters')" name="characters">
              <div class="module-pane">
                <div class="module-header">
                  <div class="section-heading__copy">
                    <span class="section-heading__eyebrow">{{ $t('drama.management.characterStats') }}</span>
                    <h2>{{ $t('drama.management.characterList') }}</h2>
                    <p>{{ charactersCount > 0 ? $t('drama.management.charactersCreated') : $t('drama.management.charactersTip') }}</p>
                  </div>

                  <el-button type="primary" :icon="Plus" @click="openAddCharacterDialog">
                    {{ $t('character.add') }}
                  </el-button>
                </div>

                <div class="module-summary">
                  <article class="mini-stat">
                    <span>{{ $t('drama.management.characterStats') }}</span>
                    <strong>{{ charactersCount }}</strong>
                  </article>
                  <article class="mini-stat">
                    <span>{{ $t('character.roles.main') }}</span>
                    <strong>{{ mainCharactersCount }}</strong>
                  </article>
                  <article class="mini-stat">
                    <span>{{ $t('common.updatedAt') }}</span>
                    <strong>{{ characterSpotlight ? formatDate(characterSpotlight.updated_at || characterSpotlight.created_at) : '-' }}</strong>
                  </article>
                </div>

                <div v-if="charactersList.length === 0" class="compact-empty compact-empty--success">
                  <div class="compact-empty__icon">
                    <el-icon><User /></el-icon>
                  </div>
                  <div class="compact-empty__content">
                    <h3>{{ $t('drama.management.noCharacters') }}</h3>
                    <p>{{ $t('drama.management.charactersTip') }}</p>
                  </div>
                  <el-button type="primary" :icon="Plus" @click="openAddCharacterDialog">
                    {{ $t('character.add') }}
                  </el-button>
                </div>

                <div v-else class="character-grid">
                  <article
                    v-for="character in charactersList"
                    :key="character.id"
                    class="character-card"
                  >
                    <div class="character-card__preview">
                      <img
                        v-if="character.image_url"
                        :src="fixImageUrl(character.image_url)"
                        :alt="character.name"
                      />
                      <el-avatar v-else :size="108" class="character-card__avatar">
                        {{ character.name?.[0] || '?' }}
                      </el-avatar>
                    </div>
                    <div class="character-card__body">
                      <div class="character-card__title">
                        <h3>{{ character.name }}</h3>
                        <el-tag :type="character.role === 'main' ? 'danger' : 'info'" size="small">
                          {{ formatCharacterRole(character.role) }}
                        </el-tag>
                      </div>
                      <p>{{ getCharacterSummary(character) }}</p>
                      <div class="character-card__meta">
                        <span>{{ $t('common.updatedAt') }} {{ formatDate(character.updated_at) }}</span>
                      </div>
                      <div class="character-card__actions">
                        <el-button @click="editCharacter(character)">{{ $t('common.edit') }}</el-button>
                        <el-button type="danger" plain @click="deleteCharacter(character)">
                          {{ $t('common.delete') }}
                        </el-button>
                      </div>
                    </div>
                  </article>
                </div>
              </div>
            </el-tab-pane>

            <el-tab-pane :label="$t('drama.management.sceneList')" name="scenes">
              <div class="module-pane">
                <div class="module-header">
                  <div class="section-heading__copy">
                    <span class="section-heading__eyebrow">{{ $t('drama.management.sceneStats') }}</span>
                    <h2>{{ $t('drama.management.sceneList') }}</h2>
                    <p>{{ scenesCount > 0 ? $t('drama.management.sceneLibraryCount') : $t('drama.management.scenesTip') }}</p>
                  </div>

                  <el-button type="primary" :icon="Plus" @click="openAddSceneDialog">
                    {{ $t('common.add') }}
                  </el-button>
                </div>

                <div class="module-summary">
                  <article class="mini-stat">
                    <span>{{ $t('drama.management.sceneStats') }}</span>
                    <strong>{{ scenesCount }}</strong>
                  </article>
                  <article class="mini-stat">
                    <span>{{ $t('drama.management.shotsCount') }}</span>
                    <strong>{{ totalSceneShotCount }}</strong>
                  </article>
                  <article class="mini-stat">
                    <span>{{ $t('common.updatedAt') }}</span>
                    <strong>{{ sceneSpotlight ? formatDate(sceneSpotlight.updated_at || sceneSpotlight.created_at) : '-' }}</strong>
                  </article>
                </div>

                <div v-if="scenes.length === 0" class="compact-empty compact-empty--warning">
                  <div class="compact-empty__icon">
                    <el-icon><Picture /></el-icon>
                  </div>
                  <div class="compact-empty__content">
                    <h3>{{ $t('drama.management.noScenes') }}</h3>
                    <p>{{ $t('drama.management.scenesTip') }}</p>
                  </div>
                  <el-button type="primary" :icon="Plus" @click="openAddSceneDialog">
                    {{ $t('common.add') }}
                  </el-button>
                </div>

                <div v-else class="scene-grid">
                  <article
                    v-for="scene in scenes"
                    :key="scene.id"
                    class="scene-card"
                  >
                    <div class="scene-card__preview">
                      <img
                        v-if="scene.image_url"
                        :src="fixImageUrl(scene.image_url)"
                        :alt="getSceneTitle(scene)"
                      />
                      <div v-else class="scene-card__placeholder">
                        <el-icon><Picture /></el-icon>
                      </div>
                    </div>
                    <div class="scene-card__body">
                      <div class="scene-card__title">
                        <div>
                          <h3>{{ getSceneTitle(scene) }}</h3>
                          <p>{{ getSceneDescription(scene) }}</p>
                        </div>
                        <span v-if="scene.time" class="scene-chip">{{ scene.time }}</span>
                      </div>
                      <div class="scene-card__meta">
                        <span>{{ $t('drama.management.shotsCount') }} {{ scene.storyboard_count || 0 }}</span>
                        <span>{{ $t('common.updatedAt') }} {{ formatDate(scene.updated_at) }}</span>
                      </div>
                      <div class="scene-card__actions">
                        <el-button @click="editScene(scene)">{{ $t('common.edit') }}</el-button>
                        <el-button type="danger" plain @click="deleteScene(scene)">
                          {{ $t('common.delete') }}
                        </el-button>
                      </div>
                    </div>
                  </article>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </section>

        <el-dialog v-model="addCharacterDialogVisible" :title="$t('character.add')" width="600px" class="dialog-form-safe">
          <el-form
            ref="characterFormRef"
            :model="newCharacter"
            label-width="100px"
            class="long-form form-enter-flow"
            @keydown.enter="handleFormEnterNavigation"
          >
            <el-form-item :label="$t('character.name')">
              <el-input v-model="newCharacter.name" :placeholder="$t('character.name')" />
            </el-form-item>
            <el-form-item :label="$t('character.role')">
              <el-select v-model="newCharacter.role" :placeholder="$t('common.pleaseSelect')">
                <el-option :label="$t('character.roles.main')" value="main" />
                <el-option :label="$t('character.roles.supporting')" value="supporting" />
                <el-option :label="$t('character.roles.minor')" value="minor" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('character.appearance')">
              <el-input v-model="newCharacter.appearance" type="textarea" :rows="3" :placeholder="$t('character.appearance')" />
            </el-form-item>
            <el-form-item :label="$t('character.personality')">
              <el-input v-model="newCharacter.personality" type="textarea" :rows="3" :placeholder="$t('character.personality')" />
            </el-form-item>
            <el-form-item :label="$t('character.description')">
              <el-input v-model="newCharacter.description" type="textarea" :rows="3" :placeholder="$t('common.description')" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="addCharacterDialogVisible = false">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="addCharacter">{{ $t('common.confirm') }}</el-button>
          </template>
        </el-dialog>

        <el-dialog v-model="addSceneDialogVisible" :title="$t('common.add')" width="600px" class="dialog-form-safe">
          <el-form
            ref="sceneFormRef"
            :model="newScene"
            label-width="100px"
            class="long-form form-enter-flow"
            @keydown.enter="handleFormEnterNavigation"
          >
            <el-form-item :label="$t('common.name')">
              <el-input v-model="newScene.location" :placeholder="$t('common.name')" />
            </el-form-item>
            <el-form-item :label="$t('common.description')">
              <el-input v-model="newScene.prompt" type="textarea" :rows="4" :placeholder="$t('common.description')" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="addSceneDialogVisible = false">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="addScene">{{ $t('common.confirm') }}</el-button>
          </template>
        </el-dialog>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Clock, Document, EditPen, Picture, Plus, User } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { characterLibraryAPI } from '@/api/character-library'
import type { Drama } from '@/types/drama'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import { ALL_COUNTRIES } from '@/constants/countries'
import { handleFormEnterNavigation } from '@/utils/formFocus'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()

const drama = ref<Drama>()
const loading = ref(false)
const activeTab = ref((route.query.tab as string) || 'overview')
const scenes = ref<any[]>([])

const editDescriptionDialogVisible = ref(false)
const editDescriptionValue = ref('')
const savingDescription = ref(false)
const editDescriptionFormRef = ref<{ $el?: HTMLElement } | null>(null)

const addCharacterDialogVisible = ref(false)
const addSceneDialogVisible = ref(false)
const characterFormRef = ref<{ $el?: HTMLElement } | null>(null)
const sceneFormRef = ref<{ $el?: HTMLElement } | null>(null)

const newCharacter = ref({
  name: '',
  role: 'supporting',
  appearance: '',
  personality: '',
  description: ''
})

const newScene = ref({
  location: '',
  prompt: ''
})

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

const episodesCount = computed(() => drama.value?.episodes?.length || 0)
const charactersList = computed(() => drama.value?.characters || [])
const charactersCount = computed(() => charactersList.value.length)
const scenesCount = computed(() => scenes.value.length)
const projectCountries = computed(() => normalizeTargetCountries(drama.value?.target_country as string | string[] | undefined))
const projectCountryText = computed(() => {
  if (projectCountries.value.length === 0) return '-'
  return projectCountries.value.map((country) => formatCountryLabel(country)).join(' / ')
})
const projectSummary = computed(() => {
  if (!drama.value) return t('drama.management.noDescription')
  return (
    drama.value.description ||
    drama.value.marketing_selling_points ||
    drama.value.material_composition ||
    t('drama.management.noDescription')
  )
})
const workflowHint = computed(() => {
  if (episodesCount.value === 0) return t('drama.management.noEpisodesYet')
  if (charactersCount.value === 0) return t('drama.management.charactersTip')
  if (scenesCount.value === 0) return t('drama.management.scenesTip')
  return `${t('drama.management.shotsCount')} ${totalShotCount.value}`
})

const sortedEpisodes = computed(() => {
  if (!drama.value?.episodes) return []
  return [...drama.value.episodes].sort((a, b) => a.episode_number - b.episode_number)
})
const episodeSpotlight = computed(() => {
  if (sortedEpisodes.value.length === 0) return null
  return [...sortedEpisodes.value].sort((a, b) => {
    return new Date(b.updated_at || b.created_at).getTime() - new Date(a.updated_at || a.created_at).getTime()
  })[0]
})
const totalShotCount = computed(() => {
  return sortedEpisodes.value.reduce((sum, episode) => sum + getEpisodeShotCount(episode), 0)
})
const mainCharactersCount = computed(() => {
  return charactersList.value.filter((character) => String(character.role || '').trim().toLowerCase() === 'main').length
})
const characterSpotlight = computed(() => {
  if (charactersList.value.length === 0) return null
  return [...charactersList.value].sort((a, b) => {
    return new Date(b.updated_at || b.created_at).getTime() - new Date(a.updated_at || a.created_at).getTime()
  })[0]
})
const totalSceneShotCount = computed(() => {
  return scenes.value.reduce((sum, scene) => sum + Number(scene?.storyboard_count || 0), 0)
})
const sceneSpotlight = computed(() => {
  if (scenes.value.length === 0) return null
  return [...scenes.value].sort((a, b) => {
    return new Date(b.updated_at || b.created_at).getTime() - new Date(a.updated_at || a.created_at).getTime()
  })[0]
})
const heroMetrics = computed(() => [
  {
    key: 'episodes',
    label: t('drama.management.episodeStats'),
    value: episodesCount.value,
    detail: t('drama.management.episodesCreated'),
    icon: Document,
    tone: 'accent'
  },
  {
    key: 'characters',
    label: t('drama.management.characterStats'),
    value: charactersCount.value,
    detail: t('drama.management.charactersCreated'),
    icon: User,
    tone: 'success'
  },
  {
    key: 'scenes',
    label: t('drama.management.sceneStats'),
    value: scenesCount.value,
    detail: t('drama.management.sceneLibraryCount'),
    icon: Picture,
    tone: 'warning'
  },
  {
    key: 'shots',
    label: t('drama.management.shotsCount'),
    value: totalShotCount.value,
    detail: workflowHint.value,
    icon: Clock,
    tone: 'neutral'
  }
])
const focusStats = computed(() => heroMetrics.value.slice(0, 3))
const overviewFacts = computed(() => [
  {
    key: 'status',
    label: t('common.status'),
    value: getStatusText(drama.value?.status)
  },
  {
    key: 'country',
    label: t('drama.targetCountry'),
    value: projectCountryText.value
  },
  {
    key: 'material',
    label: t('drama.materialComposition'),
    value: drama.value?.material_composition || '-',
    wide: true
  },
  {
    key: 'sellingPoints',
    label: t('drama.marketingSellingPoints'),
    value: drama.value?.marketing_selling_points || '-',
    wide: true
  }
])
const metaCards = computed(() => [
  {
    key: 'status',
    label: t('common.status'),
    value: getStatusText(drama.value?.status)
  },
  {
    key: 'country',
    label: t('drama.targetCountry'),
    value: projectCountryText.value,
    title: projectCountryText.value
  },
  {
    key: 'createdAt',
    label: t('common.createdAt'),
    value: formatDate(drama.value?.created_at)
  },
  {
    key: 'updatedAt',
    label: t('common.updatedAt'),
    value: formatDate(drama.value?.updated_at)
  }
])
const overviewSignals = computed(() => [
  {
    key: 'episodes',
    label: t('drama.management.episodes'),
    title: episodeSpotlight.value ? getEpisodeTitle(episodeSpotlight.value) : t('drama.management.noEpisodes'),
    description: episodeSpotlight.value ? getEpisodeSummary(episodeSpotlight.value) : t('drama.management.noEpisodesYet'),
    count: episodesCount.value,
    icon: Document,
    tone: 'accent'
  },
  {
    key: 'characters',
    label: t('drama.management.characters'),
    title: characterSpotlight.value ? characterSpotlight.value.name : t('drama.management.noCharacters'),
    description: characterSpotlight.value ? getCharacterSummary(characterSpotlight.value) : t('drama.management.charactersTip'),
    count: charactersCount.value,
    icon: User,
    tone: 'success'
  },
  {
    key: 'scenes',
    label: t('drama.management.sceneList'),
    title: sceneSpotlight.value ? getSceneTitle(sceneSpotlight.value) : t('drama.management.noScenes'),
    description: sceneSpotlight.value ? getSceneDescription(sceneSpotlight.value) : t('drama.management.scenesTip'),
    count: scenesCount.value,
    icon: Picture,
    tone: 'warning'
  }
])

const formatCountryLabel = (code: string) => countryLabelMap.get(code) || code

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push('/dramas')
}

const getStatusTone = (status?: string) => {
  if (status === 'completed') return 'completed'
  if (status === 'planning' || status === 'production' || status === 'in_progress' || status === 'generating') return 'live'
  return 'draft'
}

const getEpisodeTone = (episode: any) => {
  if (getEpisodeShotCount(episode) > 0) return 'completed'
  if (episode?.script_content || episode?.description) return 'live'
  return 'draft'
}

const getEpisodeShotCount = (episode: any) => {
  return (
    episode?.storyboards?.length ??
    episode?.shots?.length ??
    episode?.storyboard_count ??
    episode?.shot_count ??
    0
  )
}

const formatEpisodeIndex = (episodeNumber?: number) => `EP ${String(episodeNumber || 0).padStart(2, '0')}`
const getEpisodeTitle = (episode: any) => {
  return episode?.title || t('drama.management.episodeNumber', { number: episode?.episode_number || 0 })
}

const getEpisodeSummary = (episode: any) => {
  const summary = String(episode?.description || episode?.script_content || '').trim()
  return summary || t('drama.management.clickToCreate')
}

const getCharacterSummary = (character: any) => {
  return String(character?.appearance || character?.description || character?.personality || t('character.empty')).trim()
}

const loadDramaData = async () => {
  loading.value = true
  try {
    const data = await dramaAPI.get(route.params.id as string)
    drama.value = data
    loadScenes()
  } catch (error: any) {
    ElMessage.error(error.message || t('drama.management.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

const openEditDescription = () => {
  editDescriptionValue.value = drama.value?.description || ''
  editDescriptionDialogVisible.value = true
}

const saveDescription = async () => {
  if (!drama.value) return

  try {
    savingDescription.value = true
    const updated = await dramaAPI.update(String(drama.value.id), {
      description: editDescriptionValue.value.trim()
    })
    drama.value = { ...drama.value, description: updated.description }
    ElMessage.success(t('drama.management.messages.descriptionUpdated'))
    editDescriptionDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.message || t('drama.management.messages.updateFailed'))
  } finally {
    savingDescription.value = false
  }
}

const loadScenes = () => {
  if (drama.value?.scenes) {
    scenes.value = drama.value.scenes
    return
  }
  scenes.value = []
}

const getStatusText = (status?: string) => {
  const map: Record<string, string> = {
    draft: t('drama.status.draft'),
    in_progress: t('drama.status.production'),
    planning: t('drama.status.production'),
    production: t('drama.status.production'),
    generating: t('common.generating'),
    completed: t('drama.status.completed')
  }
  return map[status || 'draft'] || t('drama.status.draft')
}

const getEpisodeStatusText = (episode: any) => {
  if (getEpisodeShotCount(episode) > 0) return t('drama.management.episodeStatus.split')
  if (episode?.script_content) return t('drama.management.episodeStatus.created')
  return t('drama.management.episodeStatus.draft')
}

const formatDate = (date?: string) => {
  if (!date) return '-'
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return date
  return parsed.toLocaleString(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US')
}

const fixImageUrl = (url?: string | null): string => {
  const value = (url || '').trim()
  if (!value) return ''
  if (value.startsWith('blob:') || value.startsWith('data:')) return value
  if (value.startsWith('/api/v1/media/proxy')) {
    try {
      const parsed = new URL(value, window.location.origin)
      const raw = parsed.searchParams.get('url')
      if (raw) return fixImageUrl(decodeURIComponent(raw))
    } catch {
      // Keep original value on parse failure.
    }
    return value
  }
  if (value.startsWith('http://') || value.startsWith('https://')) {
    try {
      const parsed = new URL(value)
      const isTunnelHost =
        parsed.hostname.endsWith('.loca.lt') ||
        parsed.hostname.includes('ngrok') ||
        parsed.hostname.endsWith('.trycloudflare.com')
      if (isTunnelHost && parsed.pathname.startsWith('/static/')) {
        return `${parsed.pathname}${parsed.search}`
      }
    } catch {
      // Keep original value on parse failure.
    }
    return value
  }
  if (value.startsWith('/static/')) return value
  if (value.startsWith('/data/')) return `/static${value}`
  if (value.startsWith('data/')) return `/static/${value}`
  if (value.startsWith('/')) return value
  return `/static/${value}`
}

const getSceneTitle = (scene: any) => {
  const location = String(scene?.location || scene?.name || '').trim()
  const sceneTime = String(scene?.time || '').trim()

  if (location && sceneTime) return `${location} · ${sceneTime}`
  if (location) return location
  if (sceneTime) return sceneTime
  return t('drama.management.sceneUnnamed')
}

const getSceneDescription = (scene: any) => {
  const description = String(scene?.prompt || scene?.description || '').trim()
  return description || t('drama.management.sceneNoDescription')
}

const formatCharacterRole = (role?: string) => {
  const normalized = String(role || '').trim().toLowerCase()
  if (normalized === 'main') return t('character.roles.main')
  if (normalized === 'supporting') return t('character.roles.supporting')
  if (normalized === 'minor') return t('character.roles.minor')
  return role || '-'
}

const createNewEpisode = () => {
  const nextEpisodeNumber = episodesCount.value + 1
  router.push({
    name: 'EpisodeWorkflowNew',
    params: {
      id: route.params.id,
      episodeNumber: nextEpisodeNumber
    }
  })
}

const handleOverviewPrimaryAction = () => {
  if (episodeSpotlight.value) {
    enterEpisodeWorkflow(episodeSpotlight.value)
    return
  }
  createNewEpisode()
}

const activateTab = (tab: string) => {
  activeTab.value = tab
}

const enterEpisodeWorkflow = (episode: any) => {
  router.push({
    name: 'EpisodeWorkflowNew',
    params: {
      id: route.params.id,
      episodeNumber: episode.episode_number
    }
  })
}

const deleteEpisode = async (episode: any) => {
  try {
    await ElMessageBox.confirm(
      t('drama.management.messages.deleteEpisodeConfirm', { number: episode.episode_number }),
      t('common.confirmDelete'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )

    const existingEpisodes = drama.value?.episodes || []
    const updatedEpisodes = existingEpisodes
      .filter((item) => item.episode_number !== episode.episode_number)
      .map((item) => ({
        episode_number: item.episode_number,
        title: item.title,
        script_content: item.script_content,
        description: item.description,
        duration: item.duration,
        status: item.status
      }))

    await dramaAPI.saveEpisodes(drama.value!.id, updatedEpisodes)

    ElMessage.success(t('drama.management.messages.episodeDeleted', { number: episode.episode_number }))
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || t('drama.messages.deleteFailed'))
    }
  }
}

const openAddCharacterDialog = () => {
  newCharacter.value = {
    name: '',
    role: 'supporting',
    appearance: '',
    personality: '',
    description: ''
  }
  addCharacterDialogVisible.value = true
}

const addCharacter = async () => {
  if (!newCharacter.value.name.trim()) {
    ElMessage.warning(t('character.messages.enterName'))
    return
  }

  try {
    const existingCharacters = drama.value?.characters || []
    const allCharacters = [
      ...existingCharacters.map((character) => ({
        name: character.name,
        role: character.role,
        appearance: character.appearance,
        personality: character.personality,
        description: character.description
      })),
      newCharacter.value
    ]

    await dramaAPI.saveCharacters(drama.value!.id, allCharacters)
    ElMessage.success(t('character.messages.added'))
    addCharacterDialogVisible.value = false
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.message || t('character.messages.addFailed'))
  }
}

const editCharacter = (_character: any) => {
  ElMessage.info(t('character.messages.editInDevelopment'))
}

const deleteCharacter = async (character: any) => {
  if (character.library_id) {
    ElMessage.warning(t('character.messages.libraryDeleteHint'))
    return
  }

  if (!character.id) {
    ElMessage.error(t('character.messages.missingId'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('character.messages.deleteConfirm', { name: character.name }),
      t('common.confirmDelete'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )

    await characterLibraryAPI.deleteCharacter(character.id)
    ElMessage.success(t('character.messages.deleted'))
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除角色失败:', error)
      ElMessage.error(error.message || t('drama.messages.deleteFailed'))
    }
  }
}

const openAddSceneDialog = () => {
  newScene.value = {
    location: '',
    prompt: ''
  }
  addSceneDialogVisible.value = true
}

const addScene = async () => {
  if (!newScene.value.location.trim()) {
    ElMessage.warning(t('drama.management.messages.enterSceneName'))
    return
  }

  try {
    ElMessage.success(t('drama.management.messages.sceneAdded'))
    addSceneDialogVisible.value = false
    loadScenes()
  } catch (error: any) {
    ElMessage.error(error.message || t('character.messages.addFailed'))
  }
}

const editScene = (scene: any) => {
  newScene.value = {
    location: scene.location || scene.name || '',
    prompt: scene.prompt || scene.description || ''
  }
  addSceneDialogVisible.value = true
}

const deleteScene = async (scene: any) => {
  if (!scene.id) {
    ElMessage.error(t('drama.management.messages.missingSceneId'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('drama.management.messages.deleteSceneConfirm', { name: getSceneTitle(scene) }),
      t('common.confirmDelete'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )

    const sceneId = String(scene.id)
    await dramaAPI.deleteScene(sceneId)

    scenes.value = scenes.value.filter((item) => String(item.id) !== sceneId)
    if (drama.value?.scenes) {
      drama.value.scenes = drama.value.scenes.filter((item) => String(item.id) !== sceneId)
    }

    ElMessage.success(t('drama.management.messages.sceneDeleted'))
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除场景失败:', error)
      ElMessage.error(error.message || t('drama.messages.deleteFailed'))
    }
  }
}

onMounted(() => {
  loadDramaData()

  if (route.query.tab) {
    activeTab.value = route.query.tab as string
  }
})
</script>

<style scoped>
.management-page {
  --management-surface: rgba(255, 255, 255, 0.72);
  --management-surface-strong: rgba(255, 255, 255, 0.9);
  --management-panel: rgba(255, 255, 255, 0.94);
  --management-border: rgba(124, 146, 185, 0.2);
  --management-border-strong: rgba(89, 116, 170, 0.28);
  --management-shadow: 0 36px 90px -54px rgba(31, 54, 104, 0.34);
  --management-shadow-soft: 0 28px 56px -40px rgba(31, 54, 104, 0.24);
  --management-text: #18366d;
  --management-muted: #617ba2;
  --management-indigo: #4d69ee;
  --management-success: #1db56b;
  --management-warning: #f59f0b;
  --management-danger: #d9534f;
  --management-hero: linear-gradient(135deg, #12234f 0%, #254289 52%, #6f8df8 100%);
  position: relative;
  min-height: 100%;
  background:
    radial-gradient(circle at 8% 16%, rgba(104, 179, 255, 0.2), transparent 28%),
    radial-gradient(circle at 82% 18%, rgba(255, 194, 133, 0.2), transparent 22%),
    radial-gradient(circle at 76% 82%, rgba(126, 139, 255, 0.14), transparent 20%),
    linear-gradient(180deg, #fbfdff 0%, #f3f8ff 48%, #eef4ff 100%);
  color: var(--management-text);
  overflow: hidden;
}

.dark .management-page {
  --management-surface: rgba(11, 18, 34, 0.78);
  --management-surface-strong: rgba(15, 24, 46, 0.9);
  --management-panel: rgba(15, 24, 46, 0.92);
  --management-border: rgba(124, 150, 209, 0.16);
  --management-border-strong: rgba(142, 169, 232, 0.24);
  --management-shadow: 0 42px 100px -60px rgba(0, 0, 0, 0.72);
  --management-shadow-soft: 0 28px 56px -40px rgba(0, 0, 0, 0.48);
  --management-text: #eef4ff;
  --management-muted: #b1c3e8;
  --management-indigo: #85a1ff;
  --management-success: #55d28e;
  --management-warning: #ffbe57;
  --management-danger: #ff8b84;
  --management-hero: linear-gradient(135deg, #0c1738 0%, #183068 48%, #2f56a4 100%);
  background:
    radial-gradient(circle at 12% 20%, rgba(52, 183, 232, 0.16), transparent 28%),
    radial-gradient(circle at 78% 22%, rgba(255, 138, 38, 0.14), transparent 24%),
    radial-gradient(circle at 82% 88%, rgba(139, 92, 246, 0.16), transparent 20%),
    linear-gradient(180deg, #0e1426 0%, #0b1120 42%, #0d1629 100%);
}

.management-page__backdrop {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.management-page__glow {
  position: absolute;
  border-radius: 999px;
  filter: blur(18px);
  opacity: 0.95;
}

.management-page__glow--blue {
  top: -180px;
  left: -120px;
  width: 460px;
  height: 460px;
  background: radial-gradient(circle, rgba(52, 183, 232, 0.22) 0%, rgba(52, 183, 232, 0) 74%);
}

.management-page__glow--violet {
  top: 140px;
  right: 12%;
  width: 420px;
  height: 420px;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.14) 0%, rgba(139, 92, 246, 0) 74%);
}

.management-page__glow--orange {
  bottom: -180px;
  right: -100px;
  width: 520px;
  height: 520px;
  background: radial-gradient(circle, rgba(255, 138, 38, 0.16) 0%, rgba(255, 138, 38, 0) 72%);
}

.management-page__mesh {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(139, 160, 196, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(139, 160, 196, 0.07) 1px, transparent 1px);
  background-size: 96px 96px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.34), transparent 82%);
}

.dark .management-page__mesh {
  background-image:
    linear-gradient(rgba(139, 160, 196, 0.07) 1px, transparent 1px),
    linear-gradient(90deg, rgba(139, 160, 196, 0.06) 1px, transparent 1px);
}

.management-shell {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  max-width: 1520px;
  margin: 0 auto;
  padding: 24px 24px 18px;
}

.workspace-topbar {
  position: sticky;
  top: 18px;
  z-index: 14;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 16px 22px;
  border: 1px solid var(--management-border-strong);
  border-radius: 28px;
  background: var(--management-surface);
  box-shadow: var(--management-shadow);
  backdrop-filter: blur(22px);
}

.workspace-topbar__left,
.workspace-topbar__right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.workspace-topbar__left {
  min-width: 0;
}

.workspace-back {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 48px;
  padding: 0 16px;
  border: 1px solid rgba(152, 173, 206, 0.24);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  color: var(--management-text);
  font-size: 0.96rem;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.22s ease, border-color 0.22s ease, box-shadow 0.22s ease;
}

.workspace-back:hover {
  transform: translateY(-1px);
  border-color: rgba(85, 104, 239, 0.24);
  box-shadow: 0 18px 36px -28px rgba(34, 62, 109, 0.34);
}

.workspace-heading {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.workspace-heading__copy {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.workspace-heading__copy strong {
  color: var(--management-text);
  font-size: 0.98rem;
  font-weight: 800;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.workspace-heading__copy span {
  color: var(--management-muted);
  font-size: 0.82rem;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.workspace-marker {
  display: inline-flex;
  align-items: center;
  min-height: 42px;
  padding: 0 16px;
  border-radius: 999px;
  border: 1px solid rgba(85, 104, 239, 0.18);
  background: linear-gradient(135deg, rgba(52, 183, 232, 0.1), rgba(139, 92, 246, 0.08));
  color: var(--management-indigo);
  font-size: 0.9rem;
  font-weight: 800;
}

.workspace-status {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  padding: 10px 14px;
  border: 1px solid rgba(152, 173, 206, 0.22);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.6);
}

.workspace-status__meta {
  display: grid;
  gap: 2px;
}

.workspace-status__meta span {
  color: var(--management-muted);
  font-size: 0.8rem;
  font-weight: 700;
}

.workspace-status__meta strong {
  color: var(--management-text);
  font-size: 0.92rem;
  font-weight: 800;
}

.management-page :deep(.language-switcher) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  border-radius: 18px;
  border: 1px solid var(--management-border);
  background: rgba(255, 255, 255, 0.72);
  color: var(--management-text);
  box-shadow: 0 16px 36px -28px rgba(34, 62, 109, 0.34);
  gap: 10px;
  padding: 0 16px;
}

.management-page :deep(.lang-text) {
  color: var(--management-text);
  font-weight: 700;
}

.dark .management-page :deep(.language-switcher),
.dark .workspace-back,
.dark .workspace-status {
  background: rgba(18, 29, 54, 0.78);
}

.management-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 24px;
  padding-top: 24px;
}

.hero-grid {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(340px, 0.75fr);
  gap: 24px;
}

.hero-card,
.focus-panel,
.meta-card,
.workspace-panel,
.section-card,
.detail-strip,
.compact-empty,
.mini-stat,
.episode-row,
.character-card,
.scene-card {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--management-border);
  border-radius: 30px;
  box-shadow: var(--management-shadow-soft);
}

.hero-card::before,
.focus-panel::before,
.meta-card::before,
.workspace-panel::before,
.section-card::before,
.detail-strip::before,
.compact-empty::before,
.mini-stat::before,
.episode-row::before,
.character-card::before,
.scene-card::before {
  content: '';
  position: absolute;
  inset: 1px;
  border-radius: inherit;
  border: 1px solid rgba(255, 255, 255, 0.36);
  pointer-events: none;
}

.dark .hero-card::before,
.dark .focus-panel::before,
.dark .meta-card::before,
.dark .workspace-panel::before,
.dark .section-card::before,
.dark .detail-strip::before,
.dark .compact-empty::before,
.dark .mini-stat::before,
.dark .episode-row::before,
.dark .character-card::before,
.dark .scene-card::before {
  border-color: rgba(255, 255, 255, 0.04);
}

.hero-card {
  display: grid;
  gap: 24px;
  padding: 32px;
}

.hero-card--overview {
  border-color: rgba(255, 255, 255, 0.08);
  background: var(--management-hero);
  box-shadow: 0 44px 88px -50px rgba(17, 36, 84, 0.52);
}

.hero-card__accent {
  position: absolute;
  right: -64px;
  top: -72px;
  width: 240px;
  height: 240px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.26), transparent 70%);
  filter: blur(6px);
  pointer-events: none;
}

.hero-card__head,
.focus-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.hero-card__eyebrow,
.focus-panel__eyebrow {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  font-size: 0.84rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.hero-card__eyebrow {
  color: rgba(255, 255, 255, 0.74);
}

.hero-card__content {
  display: grid;
  gap: 16px;
}

.hero-card__content h1 {
  margin: 0;
  max-width: 12ch;
  font-size: clamp(2.7rem, 4.6vw, 4.6rem);
  line-height: 0.98;
  letter-spacing: -0.05em;
  color: #ffffff;
}

.hero-card__content p {
  margin: 0;
  max-width: 80ch;
  color: rgba(238, 244, 255, 0.82);
  font-size: 1rem;
  line-height: 1.95;
}

.project-stage__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero-primary-action {
  min-height: 52px;
  padding: 0 22px;
  border: none;
  border-radius: 18px;
  background: #ffffff;
  color: #173267;
  font-weight: 800;
  box-shadow: 0 20px 40px -28px rgba(12, 27, 66, 0.58);
}

.hero-primary-action:hover {
  background: rgba(255, 255, 255, 0.96);
  color: #173267;
}

.hero-secondary-action {
  min-height: 52px;
  padding: 0 20px;
  border-radius: 18px;
  border-color: rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
  font-weight: 800;
  backdrop-filter: blur(14px);
}

.hero-secondary-action:hover {
  border-color: rgba(255, 255, 255, 0.26);
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
}

.hero-metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.metric-card {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr);
  gap: 14px;
  align-items: center;
  padding: 18px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(18px);
}

.metric-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 18px;
  font-size: 1.4rem;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.12);
}

.metric-card__body {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.metric-card__body span,
.metric-card__body small {
  color: rgba(238, 244, 255, 0.76);
  font-weight: 700;
}

.metric-card__body span {
  font-size: 0.84rem;
}

.metric-card__body strong {
  color: #ffffff;
  font-size: 1.72rem;
  line-height: 1;
  font-weight: 900;
}

.metric-card__body small {
  font-size: 0.78rem;
  line-height: 1.5;
}

.metric-card--accent .metric-card__icon {
  background: rgba(115, 176, 255, 0.16);
}

.metric-card--success .metric-card__icon {
  background: rgba(60, 221, 160, 0.16);
}

.metric-card--warning .metric-card__icon {
  background: rgba(255, 196, 97, 0.16);
}

.metric-card--neutral .metric-card__icon {
  background: rgba(255, 255, 255, 0.14);
}

.hero-side {
  display: grid;
  gap: 18px;
  align-content: start;
}

.focus-panel,
.meta-card,
.workspace-panel,
.section-card,
.detail-strip,
.mini-stat,
.compact-empty,
.episode-row,
.character-card,
.scene-card {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(247, 250, 255, 0.84)),
    radial-gradient(circle at top right, rgba(86, 114, 239, 0.08), transparent 26%);
}

.dark .focus-panel,
.dark .meta-card,
.dark .workspace-panel,
.dark .section-card,
.dark .detail-strip,
.dark .mini-stat,
.dark .compact-empty,
.dark .episode-row,
.dark .character-card,
.dark .scene-card {
  background:
    linear-gradient(180deg, rgba(15, 24, 46, 0.94), rgba(11, 18, 34, 0.84)),
    radial-gradient(circle at top right, rgba(86, 114, 239, 0.16), transparent 26%);
}

.focus-panel {
  display: grid;
  gap: 18px;
  padding: 24px;
  text-align: left;
  color: inherit;
  cursor: pointer;
  transition: transform 0.22s ease, border-color 0.22s ease, box-shadow 0.22s ease;
}

.focus-panel:hover,
.module-nav-card:hover,
.episode-row:hover,
.character-card:hover,
.scene-card:hover {
  transform: translateY(-2px);
  border-color: rgba(85, 104, 239, 0.28);
  box-shadow: 0 32px 62px -40px rgba(34, 62, 109, 0.32);
}

.focus-panel__copy {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.focus-panel__eyebrow {
  color: var(--management-indigo);
}

.focus-panel__copy h2 {
  margin: 0;
  color: var(--management-text);
  font-size: 1.56rem;
  line-height: 1.16;
}

.focus-panel p {
  margin: 0;
  color: var(--management-muted);
  line-height: 1.78;
}

.focus-panel__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.focus-panel__stat,
.meta-card {
  display: grid;
  gap: 8px;
}

.focus-panel__stat {
  padding: 14px 16px;
  border-radius: 20px;
  background: rgba(77, 105, 238, 0.06);
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.meta-card {
  padding: 18px 20px;
}

.meta-card span,
.detail-strip span,
.mini-stat span,
.focus-panel__stat span {
  color: var(--management-muted);
  font-size: 0.86rem;
  font-weight: 700;
}

.meta-card strong,
.detail-strip strong,
.mini-stat strong,
.focus-panel__stat strong {
  color: var(--management-text);
  font-size: 1.08rem;
  line-height: 1.4;
}

.meta-card strong {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.workspace-panel {
  padding: 28px;
  backdrop-filter: blur(26px);
}

:deep(.workspace-tabs > .el-tabs__header) {
  margin: 0 0 28px;
}

:deep(.workspace-tabs .el-tabs__nav-wrap::after) {
  display: none;
}

:deep(.workspace-tabs .el-tabs__nav-scroll) {
  display: flex;
}

:deep(.workspace-tabs .el-tabs__nav) {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 8px;
  border-radius: 24px;
  border: 1px solid rgba(152, 173, 206, 0.16);
  background: rgba(255, 255, 255, 0.6);
}

.dark :deep(.workspace-tabs .el-tabs__nav) {
  background: rgba(16, 26, 48, 0.68);
}

:deep(.workspace-tabs .el-tabs__active-bar) {
  display: none;
}

:deep(.workspace-tabs .el-tabs__item) {
  height: auto;
  padding: 12px 18px;
  border-radius: 16px;
  color: var(--management-muted);
  font-size: 0.96rem;
  font-weight: 800;
  line-height: 1;
  transition: all 0.22s ease;
}

:deep(.workspace-tabs .el-tabs__item:hover) {
  color: var(--management-text);
}

:deep(.workspace-tabs .el-tabs__item.is-active) {
  color: var(--management-text);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(247, 250, 255, 0.82)),
    radial-gradient(circle at top right, rgba(85, 104, 239, 0.08), transparent 34%);
  box-shadow: 0 24px 48px -36px rgba(34, 62, 109, 0.42);
}

.overview-pane,
.module-pane {
  display: grid;
  gap: 22px;
}

.overview-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.14fr) minmax(340px, 0.86fr);
  gap: 18px;
}

.overview-stage {
  display: grid;
  grid-template-columns: minmax(0, 1.18fr) minmax(360px, 0.82fr);
  gap: 18px;
}

.overview-stack {
  display: grid;
  gap: 18px;
}

.overview-story,
.overview-flow__panel {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(132, 159, 220, 0.22);
  border-radius: 32px;
  box-shadow: 0 30px 70px -44px rgba(22, 43, 91, 0.42);
}

.overview-story::before,
.overview-flow__panel::before {
  content: '';
  position: absolute;
  inset: 1px;
  border-radius: inherit;
  border: 1px solid rgba(255, 255, 255, 0.14);
  pointer-events: none;
}

.overview-story {
  display: grid;
  gap: 22px;
  padding: 28px;
  background: linear-gradient(135deg, rgba(17, 35, 79, 0.98), rgba(36, 64, 132, 0.95) 58%, rgba(100, 129, 234, 0.92));
}

.overview-story__glow {
  position: absolute;
  right: -72px;
  bottom: -96px;
  width: 260px;
  height: 260px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.22), transparent 72%);
  filter: blur(10px);
  pointer-events: none;
}

.overview-story__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.overview-story .section-heading__copy--light h2,
.overview-story .section-heading__copy--light p,
.overview-flow__panel .section-heading__copy--light h2,
.overview-flow__panel .section-heading__copy--light p {
  color: #ffffff;
}

.overview-story .section-heading__copy--light p,
.overview-flow__panel .section-heading__copy--light p {
  color: rgba(235, 241, 255, 0.76);
}

.overview-story .section-heading__eyebrow--soft,
.overview-flow__panel .section-heading__eyebrow--soft {
  color: rgba(235, 241, 255, 0.7);
}

.overview-story .ghost-link--light,
.overview-flow__panel .ghost-link--light {
  color: #ffffff;
}

.overview-story .ghost-link--light:hover,
.overview-flow__panel .ghost-link--light:hover {
  color: rgba(255, 255, 255, 0.92);
}

.overview-story__lead {
  display: grid;
  gap: 10px;
}

.overview-story__label {
  color: rgba(235, 241, 255, 0.68);
  font-size: 0.84rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.overview-story__lead p {
  margin: 0;
  color: rgba(243, 247, 255, 0.92);
  font-size: 1.08rem;
  line-height: 1.95;
}

.overview-story__facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.story-fact {
  display: grid;
  gap: 8px;
  padding: 18px 20px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(14px);
}

.story-fact--wide {
  grid-column: span 2;
}

.story-fact span {
  color: rgba(235, 241, 255, 0.68);
  font-size: 0.84rem;
  font-weight: 700;
}

.story-fact strong {
  color: #ffffff;
  font-size: 1rem;
  line-height: 1.8;
  font-weight: 800;
}

.overview-flow {
  display: grid;
  gap: 18px;
}

.overview-flow__panel {
  display: grid;
  gap: 18px;
  padding: 24px;
  background:
    linear-gradient(180deg, rgba(14, 27, 61, 0.96), rgba(19, 38, 84, 0.9)),
    radial-gradient(circle at top right, rgba(120, 154, 255, 0.14), transparent 28%);
}

.section-heading--light {
  align-items: flex-start;
}

.section-heading--light .status-pill {
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
}

.signal-list {
  display: grid;
  gap: 14px;
}

.signal-item {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr);
  gap: 14px;
  align-items: flex-start;
  padding: 18px 20px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.08);
  color: inherit;
  text-align: left;
  transition: transform 0.22s ease, border-color 0.22s ease, background 0.22s ease;
}

button.signal-item {
  cursor: pointer;
}

button.signal-item:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.1);
}

.signal-item__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 18px;
  font-size: 1.35rem;
  background: rgba(255, 255, 255, 0.08);
}

.signal-item__body {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.signal-item__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.signal-item__top span,
.signal-item__body p {
  color: rgba(235, 241, 255, 0.74);
}

.signal-item__top strong,
.signal-item__body h3 {
  color: #ffffff;
}

.overview-flow__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 4px;
}

.overview-flow__action {
  min-height: 48px;
  padding: 0 20px;
  border: none;
  border-radius: 18px;
  background: #ffffff;
  color: #173267;
  font-weight: 800;
}

.overview-flow__action:hover {
  background: rgba(255, 255, 255, 0.94);
  color: #173267;
}

.overview-flow__meta {
  display: grid;
  gap: 4px;
  justify-items: end;
}

.overview-flow__meta span {
  color: rgba(235, 241, 255, 0.68);
  font-size: 0.8rem;
  font-weight: 700;
}

.overview-flow__meta strong {
  color: #ffffff;
  font-size: 0.94rem;
  font-weight: 800;
}

.module-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.overview-actions {
  align-content: start;
}

.section-card {
  display: grid;
  gap: 18px;
  padding: 24px;
}

.section-card--primary {
  padding: 26px;
}

.section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.section-heading--compact {
  align-items: center;
}

.section-heading__copy {
  display: grid;
  gap: 8px;
}

.section-heading__eyebrow {
  color: var(--management-indigo);
  font-size: 0.82rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.section-heading__copy h2 {
  margin: 0;
  color: var(--management-text);
  font-size: 1.48rem;
  line-height: 1.2;
}

.section-heading__copy p,
.detail-text,
.module-nav-card p,
.episode-row__body p,
.character-card__body p,
.scene-card__title p,
.compact-empty__content p {
  margin: 0;
  color: var(--management-muted);
  line-height: 1.8;
}

.detail-text--lead {
  font-size: 1.02rem;
  line-height: 1.9;
}

.ghost-link {
  min-height: auto;
  padding: 0;
  color: var(--management-indigo);
  box-shadow: none;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.detail-strip,
.mini-stat {
  display: grid;
  gap: 10px;
  padding: 18px 20px;
}

button.module-nav-card {
  cursor: pointer;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.module-nav {
  display: grid;
  gap: 14px;
}

.module-nav-card {
  display: grid;
  grid-template-columns: 68px minmax(0, 1fr);
  gap: 16px;
  align-items: center;
  padding: 18px 20px;
  color: inherit;
  text-align: left;
}

.module-nav-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 68px;
  height: 68px;
  border-radius: 22px;
  font-size: 1.7rem;
}

.module-nav-card--accent .module-nav-card__icon {
  color: var(--management-indigo);
  background: rgba(77, 105, 238, 0.1);
}

.module-nav-card--success .module-nav-card__icon {
  color: var(--management-success);
  background: rgba(29, 181, 107, 0.1);
}

.module-nav-card--warning .module-nav-card__icon {
  color: var(--management-warning);
  background: rgba(245, 159, 11, 0.12);
}

.module-nav-card__body {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.module-nav-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.module-nav-card__top span {
  color: var(--management-text);
  font-size: 0.98rem;
  font-weight: 800;
}

.module-nav-card__top strong {
  color: var(--management-indigo);
  font-size: 1.3rem;
  font-weight: 900;
}

.signal-item__body h3,
.episode-row__title h3,
.character-card__title h3,
.scene-card__title h3,
.compact-empty__content h3 {
  margin: 0;
  color: var(--management-text);
  font-size: 1.18rem;
  line-height: 1.24;
}

.signal-item--accent .signal-item__icon,
.module-nav-card--accent .module-nav-card__icon {
  color: var(--management-indigo);
  background: rgba(77, 105, 238, 0.1);
}

.signal-item--success .signal-item__icon,
.module-nav-card--success .module-nav-card__icon {
  color: var(--management-success);
  background: rgba(29, 181, 107, 0.1);
}

.signal-item--warning .signal-item__icon,
.module-nav-card--warning .module-nav-card__icon {
  color: var(--management-warning);
  background: rgba(245, 159, 11, 0.12);
}

.character-card__title,
.scene-card__title {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.country-chip {
  display: inline-flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
  font-size: 0.88rem;
  font-weight: 800;
}

.scene-chip {
  display: inline-flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(77, 105, 238, 0.08);
  color: var(--management-indigo);
  font-size: 0.88rem;
  font-weight: 800;
}

.module-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
}

.compact-empty {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 24px;
}

.compact-empty__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 76px;
  height: 76px;
  border-radius: 24px;
  font-size: 2rem;
  flex-shrink: 0;
}

.compact-empty--accent .compact-empty__icon {
  color: var(--management-indigo);
  background: rgba(77, 105, 238, 0.1);
}

.compact-empty--success .compact-empty__icon {
  color: var(--management-success);
  background: rgba(29, 181, 107, 0.1);
}

.compact-empty--warning .compact-empty__icon {
  color: var(--management-warning);
  background: rgba(245, 159, 11, 0.12);
}

.compact-empty__content {
  display: grid;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.episode-list {
  display: grid;
  gap: 14px;
}

.episode-row {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr) auto;
  gap: 18px;
  align-items: center;
  padding: 22px;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.episode-row__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 64px;
  padding: 0 14px;
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(52, 183, 232, 0.12), rgba(85, 104, 239, 0.14));
  color: var(--management-indigo);
  font-size: 0.96rem;
  font-weight: 900;
  letter-spacing: 0.08em;
}

.episode-row__body {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.episode-row__title {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.episode-row__meta,
.character-card__meta,
.scene-card__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px;
  color: var(--management-muted);
  font-size: 0.9rem;
  font-weight: 700;
}

.episode-row__actions,
.character-card__actions,
.scene-card__actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.character-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.character-card {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.character-card__preview {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 220px;
  background:
    linear-gradient(135deg, rgba(52, 183, 232, 0.9), rgba(85, 104, 239, 0.84) 55%, rgba(255, 138, 38, 0.82));
  overflow: hidden;
}

.character-card__preview img,
.scene-card__preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.22s ease;
}

.character-card:hover .character-card__preview img,
.scene-card:hover .scene-card__preview img {
  transform: scale(1.04);
}

.character-card__avatar {
  background: rgba(255, 255, 255, 0.18);
  color: rgba(255, 255, 255, 0.96);
  font-size: 2rem;
  font-weight: 800;
}

.character-card__body {
  display: grid;
  gap: 14px;
  padding: 22px;
}

.scene-grid {
  display: grid;
  gap: 16px;
}

.scene-card {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  overflow: hidden;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.scene-card__preview {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 220px;
  background:
    linear-gradient(135deg, rgba(255, 138, 38, 0.88), rgba(85, 104, 239, 0.72));
  overflow: hidden;
}

.scene-card__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.82);
  font-size: 2.5rem;
}

.scene-card__body {
  display: grid;
  gap: 16px;
  padding: 24px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 7px 12px;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 800;
  white-space: nowrap;
}

.status-pill--hero {
  background: rgba(255, 255, 255, 0.14);
  color: #ffffff;
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

:deep(.el-dialog) {
  border: 1px solid rgba(152, 173, 206, 0.22);
  border-radius: 28px;
  background: var(--management-surface-strong);
  box-shadow: var(--management-shadow);
  backdrop-filter: blur(24px);
}

:deep(.el-dialog__title),
:deep(.el-form-item__label),
:deep(.el-input__inner),
:deep(.el-textarea__inner) {
  color: var(--management-text);
}

:deep(.el-input__wrapper),
:deep(.el-textarea__inner),
:deep(.el-select__wrapper) {
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 0 0 1px rgba(152, 173, 206, 0.22) inset;
}

.dark :deep(.el-input__wrapper),
.dark :deep(.el-textarea__inner),
.dark :deep(.el-select__wrapper) {
  background: rgba(16, 26, 48, 0.82);
  box-shadow: 0 0 0 1px rgba(130, 161, 214, 0.18) inset;
}

:deep(.edit-desc-dialog .el-textarea__inner) {
  min-height: 220px;
}

:deep(.el-button--danger.is-plain) {
  background: rgba(232, 93, 84, 0.08);
  border-color: rgba(232, 93, 84, 0.18);
  color: #cf514a;
}

@media (max-width: 1280px) {
  .hero-grid,
  .overview-grid,
  .overview-stage {
    grid-template-columns: 1fr;
  }

  .character-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .scene-card {
    grid-template-columns: 220px minmax(0, 1fr);
  }
}

@media (max-width: 900px) {
  .management-shell {
    padding: 16px 16px 12px;
  }

  .workspace-topbar {
    top: 12px;
    padding: 14px 16px;
    border-radius: 24px;
    flex-direction: column;
    align-items: stretch;
  }

  .workspace-topbar__left,
  .workspace-topbar__right,
  .module-header,
  .section-heading,
  .compact-empty {
    flex-direction: column;
    align-items: stretch;
  }

  .hero-card,
  .workspace-panel {
    border-radius: 28px;
    padding: 22px;
  }

  .meta-grid,
  .module-summary,
  .focus-panel__stats,
  .character-grid,
  .scene-card,
  .episode-row,
  .module-nav-card,
  .hero-metrics,
  .detail-grid,
  .overview-story__facts {
    grid-template-columns: 1fr;
  }

  .story-fact--wide {
    grid-column: span 1;
  }

  .scene-card__preview {
    min-height: 180px;
  }
}

@media (max-width: 640px) {
  .workspace-back,
  .workspace-status,
  .management-page :deep(.language-switcher),
  .section-heading :deep(.el-button),
  .compact-empty :deep(.el-button),
  .module-header :deep(.el-button),
  .hero-card__actions :deep(.el-button) {
    width: 100%;
  }

  .hero-card__content h1 {
    font-size: 1.9rem;
  }

  .hero-card__content p,
  .section-heading__copy p,
  .detail-text,
  .focus-panel p,
  .module-nav-card p,
  .episode-row__body p,
  .character-card__body p,
  .scene-card__title p,
  .compact-empty__content p,
  .signal-item__body p {
    font-size: 0.95rem;
  }

  .section-card,
  .detail-strip,
  .meta-card,
  .mini-stat,
  .focus-panel,
  .module-nav-card,
  .episode-row,
  .character-card__body,
  .scene-card__body,
  .compact-empty,
  .signal-item {
    padding: 20px;
  }

  .episode-row__actions,
  .character-card__actions,
  .scene-card__actions,
  .module-nav-card__top,
  .episode-row__title,
  .character-card__title,
  .scene-card__title,
  .hero-card__head,
  .focus-panel__head,
  .signal-item__top,
  .workspace-heading,
  .overview-story__header,
  .overview-flow__footer {
    flex-direction: column;
    align-items: stretch;
  }

  .overview-flow__meta {
    justify-items: start;
  }
}
</style>
