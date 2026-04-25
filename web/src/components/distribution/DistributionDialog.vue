<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="840px"
    top="6vh"
    destroy-on-close
  >
    <div class="distribution-dialog">
      <div class="dialog-hero">
        <div>
          <h3>{{ heroTitle }}</h3>
          <p>{{ heroDescription }}</p>
        </div>
        <div class="hero-tags">
          <el-tag effect="plain">{{ contentTypeLabel }}</el-tag>
          <el-tag v-if="sourceLabel" type="info" effect="plain">{{ sourceLabel }}</el-tag>
        </div>
      </div>

      <el-card shadow="never" class="section-card">
        <template #header>
          <div class="section-title">
            <span>账号与目标</span>
            <div class="section-actions">
              <el-button size="small" @click="loadTargets">刷新</el-button>
              <el-button size="small" :loading="syncingProfile" @click="syncProfile">刷新连接状态</el-button>
            </div>
          </div>
        </template>

        <div class="platform-grid">
          <div class="platform-card">
            <div class="platform-card__header">
              <span>Pinterest</span>
              <el-tag :type="getConnectionTagType('pinterest')">
                {{ hasConnectedPlatform('pinterest') ? '已连接' : '未连接' }}
              </el-tag>
            </div>
            <p class="platform-card__meta">
              {{ props.contentType === 'text' ? '文本内容不支持 Pinterest。' : pinterestBoards.length ? `已加载 ${pinterestBoards.length} 个 board` : '连接后可拉取并选择 board。' }}
            </p>
            <div class="platform-card__actions">
              <el-button
                size="small"
                type="primary"
                plain
                :loading="connectingPlatform === 'pinterest'"
                @click="openConnectLink('pinterest')"
              >
                连接 Pinterest
              </el-button>
              <el-button
                size="small"
                :disabled="!hasConnectedPlatform('pinterest')"
                :loading="loadingBoards"
                @click="refreshBoards"
              >
                拉取 Boards
              </el-button>
            </div>
          </div>

          <div class="platform-card">
            <div class="platform-card__header">
              <span>Reddit</span>
              <el-tag :type="getConnectionTagType('reddit')">
                {{ hasConnectedPlatform('reddit') ? '已连接' : '未连接' }}
              </el-tag>
            </div>
            <p class="platform-card__meta">
              {{ defaultRedditTarget ? `默认目标：${defaultRedditTarget.name || defaultRedditTarget.identifier}` : '支持默认 subreddit，发布时也可以覆盖。' }}
            </p>
            <div class="platform-card__actions">
              <el-button
                size="small"
                type="primary"
                plain
                :loading="connectingPlatform === 'reddit'"
                @click="openConnectLink('reddit')"
              >
                连接 Reddit
              </el-button>
            </div>
          </div>

          <div class="platform-card">
            <div class="platform-card__header">
              <span>Discord</span>
              <el-tag :type="discordTargets.length ? 'success' : 'warning'">
                {{ discordTargets.length ? '已配置' : '未配置' }}
              </el-tag>
            </div>
            <p class="platform-card__meta">
              {{ defaultDiscordTarget ? `默认目标：${defaultDiscordTarget.name || defaultDiscordTarget.identifier}` : '请先在设置页保存官方 webhook 目标。' }}
            </p>
            <div class="platform-card__actions">
              <el-select
                v-model="form.discordTargetId"
                size="small"
                class="platform-target-select"
                placeholder="选择 Discord 目标"
                clearable
              >
                <el-option
                  v-for="target in discordTargets"
                  :key="target.id"
                  :label="target.name || target.identifier"
                  :value="target.id"
                />
              </el-select>
            </div>
          </div>
        </div>
      </el-card>

      <el-card shadow="never" class="section-card">
        <template #header>
          <div class="section-title">
            <span>发布内容</span>
            <el-tag type="info" effect="plain">{{ contentTypeLabel }}</el-tag>
          </div>
        </template>

        <el-form label-position="top">
          <el-form-item label="分发平台">
            <el-checkbox-group v-model="form.selectedPlatforms" class="platform-selector">
              <div
                v-for="platform in platformOptions"
                :key="platform.value"
                class="platform-option"
                :class="{ 'is-disabled': !platform.enabled }"
              >
                <el-checkbox
                  :label="platform.value"
                  :disabled="!platform.enabled"
                >
                  {{ platform.label }}
                </el-checkbox>
                <span class="platform-option__hint">{{ platform.hint }}</span>
              </div>
            </el-checkbox-group>
          </el-form-item>

          <el-form-item label="标题">
            <el-input
              v-model="form.title"
              maxlength="120"
              show-word-limit
              placeholder="给这次分发填写标题"
            />
          </el-form-item>

          <el-form-item label="正文 / 描述">
            <el-input
              v-model="form.body"
              type="textarea"
              :rows="props.contentType === 'text' ? 5 : 4"
              maxlength="2000"
              show-word-limit
              placeholder="输入发布正文、描述或补充说明"
            />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :xs="24" :md="12">
              <el-form-item label="发布模式">
                <el-radio-group v-model="form.publishMode">
                  <el-radio-button label="immediate">立即发布</el-radio-button>
                  <el-radio-button label="schedule">定时发布</el-radio-button>
                </el-radio-group>
              </el-form-item>
            </el-col>
            <el-col v-if="form.publishMode === 'schedule'" :xs="24" :md="12">
              <el-form-item label="计划时间">
                <el-date-picker
                  v-model="form.scheduledAt"
                  type="datetime"
                  class="full-width"
                  placeholder="选择发布时间"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <div v-if="form.selectedPlatforms.includes('reddit')" class="platform-fields">
            <div class="platform-fields__title">Reddit</div>
            <el-row :gutter="16">
              <el-col :xs="24" :md="12">
                <el-form-item label="Subreddit">
                  <el-input
                    v-model="form.redditSubreddit"
                    placeholder="留空则使用默认 subreddit"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :md="12">
                <el-form-item label="First Comment（可选）">
                  <el-input
                    v-model="form.redditFirstComment"
                    placeholder="可作为首条评论补充信息"
                  />
                </el-form-item>
              </el-col>
            </el-row>
          </div>

          <div v-if="form.selectedPlatforms.includes('pinterest')" class="platform-fields">
            <div class="platform-fields__title">Pinterest</div>
            <el-form-item label="Board">
              <el-select
                v-model="form.pinterestBoardId"
                class="full-width"
                placeholder="留空则使用默认 board"
                clearable
                filterable
              >
                <el-option
                  v-for="board in pinterestBoards"
                  :key="board.id"
                  :label="board.name || board.identifier"
                  :value="board.identifier"
                />
              </el-select>
            </el-form-item>
            <el-alert
              v-if="!pinterestBoards.length"
              title="还没有可用 Pinterest board，请先在上方拉取 boards 或在设置页完成默认 board 配置。"
              type="warning"
              :closable="false"
              show-icon
            />
          </div>

          <div v-if="form.selectedPlatforms.includes('discord')" class="platform-fields">
            <div class="platform-fields__title">Discord</div>
            <el-form-item label="频道目标">
              <el-select
                v-model="form.discordTargetId"
                class="full-width"
                placeholder="留空则使用默认 Discord 目标"
                clearable
              >
                <el-option
                  v-for="target in discordTargets"
                  :key="target.id"
                  :label="target.name || target.identifier"
                  :value="target.id"
                />
              </el-select>
            </el-form-item>
            <el-alert
              v-if="!discordTargets.length"
              title="请先在项目设置页新增 Discord webhook 目标。"
              type="warning"
              :closable="false"
              show-icon
            />
          </div>
        </el-form>
      </el-card>

      <div ref="recordsRef" class="records-section">
        <div class="section-title section-title--spaced">
          <span>分发记录 / 状态</span>
          <el-button size="small" :loading="jobsLoading" @click="loadJobs">刷新记录</el-button>
        </div>

        <el-empty v-if="!filteredJobs.length && !jobsLoading" description="还没有当前内容的分发记录" />

        <div v-else class="job-list">
          <el-card
            v-for="job in filteredJobs"
            :key="job.id"
            shadow="never"
            class="job-card"
          >
            <div class="job-card__header">
              <div>
                <div class="job-title">
                  {{ job.title || heroTitle }}
                  <el-tag size="small" :type="getJobStatusType(job.status)">
                    {{ getJobStatusText(job.status) }}
                  </el-tag>
                </div>
                <div class="job-meta">
                  Job #{{ job.id }} · {{ formatDateTime(job.created_at) }}
                </div>
              </div>
              <el-button
                v-if="job.results.some(item => item.status === 'failed')"
                size="small"
                @click="retryJob(job.id)"
              >
                重试失败项
              </el-button>
            </div>

            <div class="result-list">
              <div
                v-for="result in job.results"
                :key="result.id"
                class="result-row"
              >
                <div class="result-row__main">
                  <div class="result-row__title">
                    <span>{{ getPlatformLabel(result.platform) }}</span>
                    <el-tag size="small" :type="getResultStatusType(result.status)">
                      {{ getResultStatusText(result.status) }}
                    </el-tag>
                  </div>
                  <div class="result-row__meta">
                    {{ getResultTargetLabel(result) }}
                  </div>
                  <div v-if="result.error_msg" class="result-row__error">
                    {{ result.error_msg }}
                  </div>
                </div>
                <div class="result-row__side">
                  <el-link
                    v-if="result.published_url"
                    type="primary"
                    :underline="false"
                    @click="openPublishedUrl(result.published_url)"
                  >
                    查看结果
                  </el-link>
                  <span v-else-if="result.request_id || result.job_id_external" class="pending-id">
                    {{ result.request_id || result.job_id_external }}
                  </span>
                </div>
              </div>
            </div>
          </el-card>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">关闭</el-button>
        <el-button type="primary" :loading="submitting" @click="submitDistribution">
          提交分发
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { distributionAPI } from '@/api/distribution'
import type {
  CreateDistributionRequest,
  DistributionContentType,
  DistributionJob,
  DistributionJobStatus,
  DistributionPlatform,
  DistributionResult,
  DistributionResultStatus,
  DistributionTarget,
  DistributionTargetsView
} from '@/types/distribution'

interface Props {
  modelValue: boolean
  contentType: DistributionContentType
  sourceType?: string
  sourceRef?: string | number
  mediaUrl?: string
  initialTitle?: string
  initialBody?: string
  dialogTitle?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submitted: [job: DistributionJob]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loadingTargets = ref(false)
const jobsLoading = ref(false)
const submitting = ref(false)
const syncingProfile = ref(false)
const loadingBoards = ref(false)
const connectingPlatform = ref<DistributionPlatform | ''>('')

const targetsView = ref<DistributionTargetsView>({
  uploadPostProfile: undefined,
  targets: []
})
const jobs = ref<DistributionJob[]>([])
const currentJobId = ref<number>()
const recordsRef = ref<HTMLElement | null>(null)
let pollingTimer: number | null = null

const form = reactive({
  selectedPlatforms: [] as DistributionPlatform[],
  title: '',
  body: '',
  publishMode: 'immediate' as 'immediate' | 'schedule',
  scheduledAt: null as Date | null,
  redditSubreddit: '',
  redditFirstComment: '',
  pinterestBoardId: '',
  discordTargetId: undefined as number | undefined
})

const sourceRefValue = computed(() => {
  if (props.sourceRef === undefined || props.sourceRef === null) {
    return ''
  }
  return String(props.sourceRef)
})

const dialogTitle = computed(() => props.dialogTitle || '一键分发')

const heroTitle = computed(() => {
  const fallbackMap: Record<DistributionContentType, string> = {
    text: '文本分发',
    image: '图片分发',
    video: '视频分发'
  }
  return props.initialTitle?.trim() || fallbackMap[props.contentType]
})

const heroDescription = computed(() => {
  if (props.contentType === 'text') {
    return '选择平台并填写目标参数，系统会按平台能力自动路由。'
  }
  return props.mediaUrl
    ? '媒体已就绪，提交后会异步分发并持续刷新各平台状态。'
    : '将复用当前内容源中的媒体资源进行分发。'
})

const contentTypeLabel = computed(() => {
  const labels: Record<DistributionContentType, string> = {
    text: 'Text',
    image: 'Image',
    video: 'Video'
  }
  return labels[props.contentType]
})

const sourceLabel = computed(() => {
  if (!props.sourceType || !sourceRefValue.value) {
    return ''
  }
  return `${props.sourceType}#${sourceRefValue.value}`
})

const pinterestBoards = computed(() =>
  targetsView.value.targets.filter(target => target.platform === 'pinterest' && target.status !== 'disabled')
)

const defaultPinterestBoard = computed(() =>
  pinterestBoards.value.find(target => target.is_default)
)

const redditTargets = computed(() =>
  targetsView.value.targets.filter(target => target.platform === 'reddit' && target.status !== 'disabled')
)

const defaultRedditTarget = computed(() =>
  redditTargets.value.find(target => target.is_default)
)

const discordTargets = computed(() =>
  targetsView.value.targets.filter(target => target.platform === 'discord' && target.status !== 'disabled')
)

const defaultDiscordTarget = computed(() =>
  discordTargets.value.find(target => target.is_default)
)

const connectedPlatforms = computed(() => {
  return Array.isArray(targetsView.value.uploadPostProfile?.connected_platforms)
    ? targetsView.value.uploadPostProfile?.connected_platforms || []
    : []
})

const platformOptions = computed(() => {
  return [
    {
      value: 'pinterest' as DistributionPlatform,
      label: 'Pinterest',
      enabled: props.contentType !== 'text' && hasConnectedPlatform('pinterest'),
      hint: props.contentType === 'text'
        ? 'Pinterest 不支持 text-only'
        : hasConnectedPlatform('pinterest')
          ? '图片 / 视频通过 Upload-Post 分发'
          : '请先连接 Pinterest'
    },
    {
      value: 'reddit' as DistributionPlatform,
      label: 'Reddit',
      enabled: hasConnectedPlatform('reddit'),
      hint: hasConnectedPlatform('reddit')
        ? '支持文本、图片、视频'
        : '请先连接 Reddit'
    },
    {
      value: 'discord' as DistributionPlatform,
      label: 'Discord',
      enabled: discordTargets.value.length > 0,
      hint: discordTargets.value.length
        ? '通过官方 webhook 发送'
        : '请先配置 Discord 目标'
    }
  ]
})

const filteredJobs = computed(() => {
  const sourceType = props.sourceType?.trim()
  const sourceRef = sourceRefValue.value
  return jobs.value.filter((job) => {
    if (sourceType && job.source_type !== sourceType) {
      return false
    }
    if (sourceRef && String(job.source_ref || '') !== sourceRef) {
      return false
    }
    return true
  })
})

const initializeForm = () => {
  form.selectedPlatforms = platformOptions.value
    .filter(item => item.enabled)
    .map(item => item.value)
  form.title = props.initialTitle?.trim() || ''
  form.body = props.initialBody?.trim() || ''
  form.publishMode = 'immediate'
  form.scheduledAt = null
  form.redditSubreddit = defaultRedditTarget.value?.identifier || ''
  form.redditFirstComment = ''
  form.pinterestBoardId = defaultPinterestBoard.value?.identifier || ''
  form.discordTargetId = defaultDiscordTarget.value?.id
}

const loadTargets = async () => {
  loadingTargets.value = true
  try {
    targetsView.value = await distributionAPI.listTargets()
    initializeForm()
  } catch (error: any) {
    ElMessage.error(error.message || '加载分发目标失败')
  } finally {
    loadingTargets.value = false
  }
}

const loadJobs = async () => {
  jobsLoading.value = true
  try {
    const result = await distributionAPI.listJobs({ page: 1, page_size: 20 })
    jobs.value = result.jobs
    const latestPending = filteredJobs.value.find(job => isJobPending(job.status))
    if (latestPending) {
      currentJobId.value = latestPending.id
      startPolling()
    } else if (!currentJobId.value) {
      stopPolling()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载分发记录失败')
  } finally {
    jobsLoading.value = false
  }
}

const syncProfile = async () => {
  syncingProfile.value = true
  try {
    await distributionAPI.syncUploadPostProfile()
    await loadTargets()
    ElMessage.success('连接状态已刷新')
  } catch (error: any) {
    ElMessage.error(error.message || '刷新连接状态失败')
  } finally {
    syncingProfile.value = false
  }
}

const openConnectLink = async (platform: DistributionPlatform) => {
  connectingPlatform.value = platform
  try {
    if (!targetsView.value.uploadPostProfile) {
      await distributionAPI.ensureUploadPostProfile()
    }
    const result = await distributionAPI.generateUploadPostConnectLink()
    window.open(result.access_url, '_blank', 'noopener')
    ElMessage.success(`已打开 ${getPlatformLabel(platform)} 连接页`)
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '打开连接页失败')
  } finally {
    connectingPlatform.value = ''
  }
}

const refreshBoards = async () => {
  loadingBoards.value = true
  try {
    await distributionAPI.listPinterestBoards()
    await loadTargets()
    ElMessage.success('Pinterest boards 已更新')
  } catch (error: any) {
    ElMessage.error(error.message || '更新 boards 失败')
  } finally {
    loadingBoards.value = false
  }
}

const submitDistribution = async () => {
  if (!form.selectedPlatforms.length) {
    ElMessage.warning('请至少选择一个平台')
    return
  }

  if (form.publishMode === 'schedule' && !form.scheduledAt) {
    ElMessage.warning('请选择计划发布时间')
    return
  }

  if (form.selectedPlatforms.includes('reddit') && !form.title.trim()) {
    ElMessage.warning('Reddit 分发需要标题')
    return
  }

  if (form.selectedPlatforms.includes('pinterest') && props.contentType === 'text') {
    ElMessage.warning('Pinterest 不支持 text-only 分发')
    return
  }

  const payload: CreateDistributionRequest = {
    sourceType: props.sourceType,
    sourceRef: sourceRefValue.value || undefined,
    contentType: props.contentType,
    title: form.title.trim(),
    body: form.body.trim() || undefined,
    mediaUrl: props.mediaUrl || undefined,
    selectedPlatforms: [...form.selectedPlatforms],
    platformOptions: {
      reddit: {
        subreddit: form.redditSubreddit.trim() || undefined,
        firstComment: form.redditFirstComment.trim() || undefined
      },
      pinterest: {
        boardId: form.pinterestBoardId || undefined
      },
      discord: {
        targetId: form.discordTargetId
      }
    },
    publishMode: form.publishMode,
    scheduledAt: form.scheduledAt ? new Date(form.scheduledAt).toISOString() : undefined
  }

  submitting.value = true
  try {
    const job = await distributionAPI.createDistribution(payload)
    upsertJob(job)
    currentJobId.value = job.id
    emit('submitted', job)
    ElMessage.success('分发任务已提交')
    startPolling()
    await nextTick()
    recordsRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  } catch (error: any) {
    ElMessage.error(error.message || '提交分发失败')
  } finally {
    submitting.value = false
  }
}

const retryJob = async (jobId: number) => {
  try {
    const job = await distributionAPI.retryJob(jobId)
    upsertJob(job)
    currentJobId.value = job.id
    ElMessage.success('已重新提交失败项')
    startPolling()
  } catch (error: any) {
    ElMessage.error(error.message || '重试失败')
  }
}

const upsertJob = (job: DistributionJob) => {
  const next = jobs.value.filter(item => item.id !== job.id)
  next.unshift(job)
  jobs.value = next.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
}

const startPolling = () => {
  if (pollingTimer || !currentJobId.value) {
    return
  }

  pollingTimer = window.setInterval(async () => {
    if (!currentJobId.value) {
      stopPolling()
      return
    }

    try {
      const job = await distributionAPI.getJob(currentJobId.value)
      upsertJob(job)
      if (!isJobPending(job.status)) {
        stopPolling()
      }
    } catch (error) {
      console.error('poll distribution job failed', error)
    }
  }, 4000)
}

const stopPolling = () => {
  if (pollingTimer) {
    window.clearInterval(pollingTimer)
    pollingTimer = null
  }
}

const hasConnectedPlatform = (platform: DistributionPlatform) => {
  return connectedPlatforms.value.includes(platform)
}

const getConnectionTagType = (platform: DistributionPlatform) => {
  return hasConnectedPlatform(platform) ? 'success' : 'warning'
}

const isJobPending = (status: DistributionJobStatus) => {
  return status === 'pending' || status === 'processing' || status === 'scheduled'
}

const getJobStatusText = (status: DistributionJobStatus) => {
  switch (status) {
    case 'pending':
      return '待处理'
    case 'scheduled':
      return '已排期'
    case 'processing':
      return '处理中'
    case 'completed':
      return '全部成功'
    case 'partially_failed':
      return '部分失败'
    case 'failed':
      return '失败'
    default:
      return status
  }
}

const getJobStatusType = (status: DistributionJobStatus) => {
  switch (status) {
    case 'completed':
      return 'success'
    case 'partially_failed':
      return 'warning'
    case 'failed':
      return 'danger'
    case 'pending':
    case 'scheduled':
    case 'processing':
      return 'info'
    default:
      return 'info'
  }
}

const getResultStatusText = (status: DistributionResultStatus) => {
  switch (status) {
    case 'pending':
      return '待处理'
    case 'scheduled':
      return '已排期'
    case 'processing':
      return '处理中'
    case 'success':
      return '成功'
    case 'failed':
      return '失败'
    default:
      return status
  }
}

const getResultStatusType = (status: DistributionResultStatus) => {
  switch (status) {
    case 'success':
      return 'success'
    case 'failed':
      return 'danger'
    case 'pending':
    case 'scheduled':
    case 'processing':
      return 'warning'
    default:
      return 'info'
  }
}

const getPlatformLabel = (platform: string) => {
  const labels: Record<string, string> = {
    pinterest: 'Pinterest',
    reddit: 'Reddit',
    discord: 'Discord'
  }
  return labels[platform] || platform
}

const getResultTargetLabel = (result: DistributionResult) => {
  if (result.target?.name) {
    return result.target.name
  }
  if (result.target?.identifier) {
    return result.target.identifier
  }

  const snapshot = result.target_snapshot as Record<string, any> | undefined
  return String(snapshot?.name || snapshot?.identifier || '当前目标')
}

const formatDateTime = (value?: string) => {
  if (!value) return '未知时间'
  return new Date(value).toLocaleString('zh-CN')
}

const openPublishedUrl = (url: string) => {
  window.open(url, '_blank', 'noopener')
}

watch(
  () => props.modelValue,
  async (value) => {
    if (!value) {
      stopPolling()
      return
    }

    await loadTargets()
    await loadJobs()
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<style scoped>
.distribution-dialog {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dialog-hero {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 20px;
  border-radius: 18px;
  background:
    linear-gradient(135deg, rgba(12, 74, 110, 0.1), rgba(14, 116, 144, 0.02)),
    var(--el-fill-color-extra-light);
  border: 1px solid rgba(14, 116, 144, 0.16);
}

.dialog-hero h3 {
  margin: 0 0 8px;
  font-size: 20px;
}

.dialog-hero p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.hero-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: flex-start;
}

.section-card {
  border-radius: 16px;
}

.section-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.section-title--spaced {
  margin-bottom: 12px;
}

.section-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.platform-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.platform-card {
  padding: 14px;
  border-radius: 14px;
  background: var(--el-fill-color-extra-light);
  border: 1px solid var(--el-border-color-lighter);
}

.platform-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.platform-card__meta {
  min-height: 40px;
  margin: 10px 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.platform-card__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.platform-target-select {
  width: 100%;
}

.platform-selector {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  width: 100%;
}

.platform-option {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 14px;
  border-radius: 14px;
  border: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color-page);
}

.platform-option.is-disabled {
  opacity: 0.6;
}

.platform-option__hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}

.platform-fields {
  margin-top: 6px;
  padding: 14px 16px;
  border-radius: 14px;
  background: var(--el-fill-color-extra-light);
  border: 1px solid var(--el-border-color-lighter);
}

.platform-fields__title {
  margin-bottom: 12px;
  font-weight: 600;
}

.full-width {
  width: 100%;
}

.records-section {
  padding-top: 4px;
}

.job-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.job-card {
  border-radius: 16px;
}

.job-card__header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.job-title {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  font-weight: 600;
}

.job-meta {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.result-list {
  margin-top: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.result-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  background: var(--el-fill-color-extra-light);
}

.result-row__title {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  font-weight: 500;
}

.result-row__meta {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.result-row__error {
  margin-top: 8px;
  color: var(--el-color-danger);
  font-size: 12px;
  line-height: 1.6;
}

.result-row__side {
  display: flex;
  align-items: center;
}

.pending-id {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 900px) {
  .platform-grid,
  .platform-selector {
    grid-template-columns: 1fr;
  }

  .dialog-hero,
  .job-card__header,
  .result-row {
    flex-direction: column;
  }
}
</style>
