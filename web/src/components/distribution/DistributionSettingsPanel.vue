<template>
  <div class="distribution-settings" v-loading="loading">
    <div class="settings-header">
      <div>
        <h3>分发账号绑定</h3>
        <p>为当前设备身份绑定你自己的 Pinterest、Reddit 和 Discord 分发目标。</p>
      </div>
      <div class="settings-actions">
        <el-button @click="loadTargets">刷新</el-button>
        <el-button type="primary" :loading="ensuringProfile" @click="ensureProfile">
          创建 / 查询 Upload-Post Profile
        </el-button>
      </div>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="12">
        <el-card shadow="hover" class="settings-card">
          <template #header>
            <div class="card-title">
              <span>Upload-Post Profile</span>
              <el-tag :type="getProfileStatusType(targetsView.uploadPostProfile?.status)">
                {{ getProfileStatusText(targetsView.uploadPostProfile?.status) }}
              </el-tag>
            </div>
          </template>

          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="Username">
              {{ targetsView.uploadPostProfile?.username || '未创建' }}
            </el-descriptions-item>
            <el-descriptions-item label="已连接平台">
              <div class="tag-list">
                <el-tag
                  v-for="platform in connectedPlatforms"
                  :key="platform"
                  size="small"
                  effect="plain"
                >
                  {{ getPlatformLabel(platform) }}
                </el-tag>
                <span v-if="!connectedPlatforms.length" class="muted">尚未连接</span>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="最近同步">
              {{ formatDateTime(targetsView.uploadPostProfile?.last_sync_at) }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="action-group">
            <el-button
              type="primary"
              :disabled="!targetsView.uploadPostProfile"
              :loading="connectingPlatform === 'pinterest'"
              @click="openConnectLink('pinterest')"
            >
              连接 Pinterest
            </el-button>
            <el-button
              type="primary"
              plain
              :disabled="!targetsView.uploadPostProfile"
              :loading="connectingPlatform === 'reddit'"
              @click="openConnectLink('reddit')"
            >
              连接 Reddit
            </el-button>
            <el-button :loading="syncingProfile" @click="syncProfile">
              刷新连接状态
            </el-button>
          </div>

          <el-alert
            title="连接说明"
            type="info"
            :closable="false"
            show-icon
          >
            连接页会在新窗口打开。完成授权后回到这里点击“刷新连接状态”。
          </el-alert>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="12">
        <el-card shadow="hover" class="settings-card">
          <template #header>
            <div class="card-title">
              <span>Pinterest 默认 Board</span>
              <el-tag :type="getConnectionTagType('pinterest')">
                {{ hasConnectedPlatform('pinterest') ? '已连接' : '未连接' }}
              </el-tag>
            </div>
          </template>

          <el-alert
            v-if="!hasConnectedPlatform('pinterest')"
            title="请先连接 Pinterest"
            type="warning"
            :closable="false"
            show-icon
          />

          <template v-else>
            <div class="inline-actions">
              <el-button :loading="loadingBoards" @click="refreshBoards">
                拉取 Boards
              </el-button>
              <span class="muted">选择默认 board，发布时可覆盖。</span>
            </div>

            <el-select
              v-model="selectedBoardId"
              class="full-width"
              placeholder="请选择默认 board"
              filterable
              clearable
            >
              <el-option
                v-for="board in pinterestBoards"
                :key="board.id"
                :label="board.name || board.identifier"
                :value="board.id"
              />
            </el-select>

            <div class="action-group">
              <el-button
                type="primary"
                :disabled="!selectedBoardId"
                @click="saveDefaultBoard"
              >
                保存默认 Board
              </el-button>
            </div>

            <div class="target-list">
              <div
                v-for="board in pinterestBoards"
                :key="board.id"
                class="target-row"
              >
                <div>
                  <div class="target-name">{{ board.name || board.identifier }}</div>
                  <div class="target-meta">{{ board.identifier }}</div>
                </div>
                <el-tag v-if="board.is_default" type="success" effect="plain">默认</el-tag>
              </div>
              <el-empty v-if="!pinterestBoards.length" description="还没有可用 board" />
            </div>
          </template>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="12">
        <el-card shadow="hover" class="settings-card">
          <template #header>
            <div class="card-title">
              <span>Reddit 默认目标</span>
              <el-tag :type="getConnectionTagType('reddit')">
                {{ hasConnectedPlatform('reddit') ? '已连接' : '未连接' }}
              </el-tag>
            </div>
          </template>

          <el-form label-position="top">
            <el-form-item label="默认 Subreddit">
              <el-input v-model="redditForm.subreddit" placeholder="例如 movies 或 r/movies" />
            </el-form-item>
            <el-form-item label="Flair ID（可选）">
              <el-input v-model="redditForm.flairId" placeholder="发布时使用的 flair id" />
            </el-form-item>
          </el-form>

          <div class="action-group">
            <el-button type="primary" :loading="savingReddit" @click="saveRedditDefault">
              保存默认 subreddit
            </el-button>
          </div>

          <div class="target-list">
            <div
              v-for="target in redditTargets"
              :key="target.id"
              class="target-row"
            >
              <div>
                <div class="target-name">{{ target.name || target.identifier }}</div>
                <div class="target-meta">{{ target.identifier }}</div>
              </div>
              <el-tag v-if="target.is_default" type="success" effect="plain">默认</el-tag>
            </div>
            <el-empty v-if="!redditTargets.length" description="还没有保存 subreddit" />
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="12">
        <el-card shadow="hover" class="settings-card">
          <template #header>
            <div class="card-title">
              <span>Discord 频道目标</span>
              <el-tag type="info">{{ discordTargets.length }} 个目标</el-tag>
            </div>
          </template>

          <el-form label-position="top">
            <el-form-item label="Webhook URL">
              <el-input
                v-model="discordForm.webhookUrl"
                type="textarea"
                :rows="3"
                placeholder="粘贴 Discord 官方 webhook URL"
              />
            </el-form-item>
            <el-form-item label="显示名称（可选）">
              <el-input v-model="discordForm.name" placeholder="例如 主频道 / 社群公告" />
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="discordForm.isDefault">设为默认目标</el-checkbox>
            </el-form-item>
          </el-form>

          <div class="action-group">
            <el-button type="primary" :loading="savingDiscord" @click="saveDiscordTarget">
              保存 Discord 目标
            </el-button>
          </div>

          <div class="target-list">
            <div
              v-for="target in discordTargets"
              :key="target.id"
              class="target-row target-row--actions"
            >
              <div>
                <div class="target-name">{{ target.name || target.identifier }}</div>
                <div class="target-meta">
                  Guild {{ getConfigField(target, 'guild_id') || '-' }} · Channel {{ getConfigField(target, 'channel_id') || '-' }}
                </div>
              </div>
              <div class="target-actions">
                <el-tag v-if="target.is_default" type="success" effect="plain">默认</el-tag>
                <el-button
                  v-else
                  size="small"
                  @click="setDefaultTarget(target.id)"
                >
                  设为默认
                </el-button>
                <el-popconfirm title="确定删除该 Discord 目标吗？" @confirm="removeTarget(target.id)">
                  <template #reference>
                    <el-button size="small" type="danger" plain>删除</el-button>
                  </template>
                </el-popconfirm>
              </div>
            </div>
            <el-empty v-if="!discordTargets.length" description="还没有 Discord webhook 目标" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { distributionAPI } from '@/api/distribution'
import type {
  DistributionPlatform,
  DistributionTarget,
  DistributionTargetsView,
  UploadPostProfile
} from '@/types/distribution'

const loading = ref(false)
const ensuringProfile = ref(false)
const syncingProfile = ref(false)
const connectingPlatform = ref<DistributionPlatform | ''>('')
const loadingBoards = ref(false)
const savingReddit = ref(false)
const savingDiscord = ref(false)

const targetsView = ref<DistributionTargetsView>({
  uploadPostProfile: undefined,
  targets: []
})
const selectedBoardId = ref<number>()

const redditForm = reactive({
  subreddit: '',
  flairId: ''
})

const discordForm = reactive({
  webhookUrl: '',
  name: '',
  isDefault: true
})

const pinterestBoards = computed(() =>
  targetsView.value.targets.filter(target => target.platform === 'pinterest' && target.status !== 'disabled')
)

const redditTargets = computed(() =>
  targetsView.value.targets.filter(target => target.platform === 'reddit' && target.status !== 'disabled')
)

const discordTargets = computed(() =>
  targetsView.value.targets.filter(target => target.platform === 'discord' && target.status !== 'disabled')
)

const connectedPlatforms = computed(() => {
  return Array.isArray(targetsView.value.uploadPostProfile?.connected_platforms)
    ? targetsView.value.uploadPostProfile?.connected_platforms || []
    : []
})

const hydrateTargetState = () => {
  const defaultBoard = pinterestBoards.value.find(item => item.is_default)
  selectedBoardId.value = defaultBoard?.id

  const defaultReddit = redditTargets.value.find(item => item.is_default)
  redditForm.subreddit = defaultReddit?.identifier || ''
  redditForm.flairId = String(getConfigField(defaultReddit, 'flair_id') || '')
}

const loadTargets = async () => {
  loading.value = true
  try {
    targetsView.value = await distributionAPI.listTargets()
    hydrateTargetState()
  } catch (error: any) {
    ElMessage.error(error.message || '加载分发配置失败')
  } finally {
    loading.value = false
  }
}

const ensureProfile = async () => {
  ensuringProfile.value = true
  try {
    const profile = await distributionAPI.ensureUploadPostProfile()
    targetsView.value = {
      ...targetsView.value,
      uploadPostProfile: profile
    }
    ElMessage.success('Upload-Post profile 已就绪')
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '创建 profile 失败')
  } finally {
    ensuringProfile.value = false
  }
}

const syncProfile = async () => {
  syncingProfile.value = true
  try {
    const profile = await distributionAPI.syncUploadPostProfile()
    targetsView.value = {
      ...targetsView.value,
      uploadPostProfile: profile
    }
    ElMessage.success('连接状态已刷新')
    await loadTargets()
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
      await ensureProfile()
    }
    const result = await distributionAPI.generateUploadPostConnectLink()
    window.open(result.access_url, '_blank', 'noopener')
    ElMessage.success(`已打开 ${getPlatformLabel(platform)} 连接页，完成后请刷新状态`)
  } catch (error: any) {
    ElMessage.error(error.message || '生成连接链接失败')
  } finally {
    connectingPlatform.value = ''
  }
}

const refreshBoards = async () => {
  loadingBoards.value = true
  try {
    await distributionAPI.listPinterestBoards()
    ElMessage.success('Pinterest boards 已更新')
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '拉取 boards 失败')
  } finally {
    loadingBoards.value = false
  }
}

const saveDefaultBoard = async () => {
  if (!selectedBoardId.value) {
    ElMessage.warning('请先选择一个 board')
    return
  }

  try {
    await distributionAPI.setDefaultTarget(selectedBoardId.value)
    ElMessage.success('默认 Board 已保存')
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '保存默认 Board 失败')
  }
}

const saveRedditDefault = async () => {
  if (!redditForm.subreddit.trim()) {
    ElMessage.warning('请输入 subreddit')
    return
  }

  savingReddit.value = true
  try {
    await distributionAPI.saveRedditDefaultTarget({
      subreddit: redditForm.subreddit.trim(),
      flairId: redditForm.flairId.trim() || undefined
    })
    ElMessage.success('默认 subreddit 已保存')
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '保存 subreddit 失败')
  } finally {
    savingReddit.value = false
  }
}

const saveDiscordTarget = async () => {
  if (!discordForm.webhookUrl.trim()) {
    ElMessage.warning('请输入 webhook URL')
    return
  }

  savingDiscord.value = true
  try {
    await distributionAPI.upsertDiscordTarget({
      webhookUrl: discordForm.webhookUrl.trim(),
      name: discordForm.name.trim() || undefined,
      isDefault: discordForm.isDefault
    })
    ElMessage.success('Discord 目标已保存')
    discordForm.webhookUrl = ''
    discordForm.name = ''
    discordForm.isDefault = discordTargets.value.length === 0
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '保存 Discord 目标失败')
  } finally {
    savingDiscord.value = false
  }
}

const setDefaultTarget = async (targetId: number) => {
  try {
    await distributionAPI.setDefaultTarget(targetId)
    ElMessage.success('默认目标已更新')
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '更新默认目标失败')
  }
}

const removeTarget = async (targetId: number) => {
  try {
    await distributionAPI.deleteTarget(targetId)
    ElMessage.success('目标已删除')
    await loadTargets()
  } catch (error: any) {
    ElMessage.error(error.message || '删除目标失败')
  }
}

const hasConnectedPlatform = (platform: DistributionPlatform) => {
  return connectedPlatforms.value.includes(platform)
}

const getPlatformLabel = (platform: string) => {
  const labels: Record<string, string> = {
    pinterest: 'Pinterest',
    reddit: 'Reddit',
    discord: 'Discord'
  }
  return labels[platform] || platform
}

const getProfileStatusText = (status?: UploadPostProfile['status']) => {
  switch (status) {
    case 'active':
      return '已激活'
    case 'error':
      return '异常'
    case 'pending':
      return '待连接'
    default:
      return '未创建'
  }
}

const getProfileStatusType = (status?: UploadPostProfile['status']) => {
  switch (status) {
    case 'active':
      return 'success'
    case 'error':
      return 'danger'
    case 'pending':
      return 'warning'
    default:
      return 'info'
  }
}

const getConnectionTagType = (platform: DistributionPlatform) => {
  return hasConnectedPlatform(platform) ? 'success' : 'warning'
}

const getConfigField = (target?: DistributionTarget, field?: string) => {
  if (!target?.config || !field) return ''
  return (target.config as Record<string, any>)[field]
}

const formatDateTime = (value?: string) => {
  if (!value) return '未同步'
  return new Date(value).toLocaleString('zh-CN')
}

onMounted(() => {
  loadTargets()
})
</script>

<style scoped>
.distribution-settings {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.settings-header h3 {
  margin: 0 0 8px;
  font-size: 20px;
}

.settings-header p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.settings-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.settings-card {
  height: 100%;
}

.card-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.tag-list {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.action-group {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin: 16px 0;
}

.inline-actions {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.full-width {
  width: 100%;
}

.target-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 16px;
}

.target-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  background: var(--el-fill-color-extra-light);
}

.target-row--actions {
  align-items: center;
}

.target-name {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.target-meta {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.target-actions {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.muted {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

@media (max-width: 768px) {
  .settings-header {
    flex-direction: column;
  }
}
</style>
