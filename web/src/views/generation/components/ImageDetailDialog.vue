<template>
  <el-dialog
    v-model="visible"
    :title="$t('image.detailTitle')"
    width="900px"
    @close="handleClose"
  >
    <div v-if="image" class="image-detail">
      <el-row :gutter="20">
        <el-col :span="14">
          <div class="image-preview">
            <el-image
              v-if="image.status === 'completed' && image.image_url"
              :src="image.image_url"
              fit="contain"
              class="preview-image"
              :preview-src-list="[image.image_url]"
            >
              <template #error>
                <div class="image-error">
                  <el-icon><PictureFilled /></el-icon>
                  <span>{{ $t('image.loadFailed') }}</span>
                </div>
              </template>
            </el-image>

            <div v-else-if="image.status === 'processing'" class="image-status">
              <el-icon class="loading-icon"><Loading /></el-icon>
              <span>{{ $t('image.generatingWait') }}</span>
            </div>

            <div v-else-if="image.status === 'failed'" class="image-status error">
              <el-icon><CircleClose /></el-icon>
              <span>{{ $t('image.generateFailed') }}</span>
              <div class="error-message">{{ image.error_msg }}</div>
            </div>
          </div>
        </el-col>

        <el-col :span="10">
          <div class="image-info">
            <el-descriptions :column="1" border>
              <el-descriptions-item :label="$t('common.status')">
                <el-tag :type="getStatusType(image.status)">
                  {{ getStatusText(image.status) }}
                </el-tag>
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.provider')">
                {{ image.provider }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.model')" v-if="image.model">
                {{ image.model }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.size')" v-if="image.size">
                {{ image.size }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.resolution')" v-if="image.width && image.height">
                {{ image.width }} × {{ image.height }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.quality')" v-if="image.quality">
                {{ image.quality }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.style')" v-if="image.style">
                {{ image.style }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.steps')" v-if="image.steps">
                {{ image.steps }}
              </el-descriptions-item>

              <el-descriptions-item label="CFG Scale" v-if="image.cfg_scale">
                {{ image.cfg_scale }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.seed')" v-if="image.seed">
                {{ image.seed }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.createdAt')">
                {{ formatDateTime(image.created_at) }}
              </el-descriptions-item>

              <el-descriptions-item :label="$t('image.completedAt')" v-if="image.completed_at">
                {{ formatDateTime(image.completed_at) }}
              </el-descriptions-item>
            </el-descriptions>

            <el-divider />

            <div class="prompt-section">
              <h4>{{ $t('image.prompt') }}</h4>
              <div class="prompt-text">{{ image.prompt }}</div>
            </div>

            <div v-if="image.negative_prompt" class="prompt-section">
              <h4>{{ $t('image.negativePrompt') }}</h4>
              <div class="prompt-text">{{ image.negative_prompt }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <template #footer>
      <el-button @click="handleClose">{{ $t('image.close') }}</el-button>
      <el-button
        v-if="image?.status === 'completed' && image?.image_url"
        type="warning"
        @click="openDistributionDialog"
      >
        {{ $t('image.distribute') }}
      </el-button>
      <el-button
        v-if="image?.status === 'completed' && image?.image_url"
        type="primary"
        @click="downloadImage"
      >
        <el-icon><Download /></el-icon>
        {{ $t('image.downloadImage') }}
      </el-button>
      <el-button
        v-if="image?.status === 'completed'"
        type="success"
        @click="regenerate"
      >
        <el-icon><Refresh /></el-icon>
        {{ $t('image.regenerate') }}
      </el-button>
    </template>

    <DistributionDialog
      v-if="image"
      v-model="distributionDialogVisible"
      content-type="image"
      source-type="image_generation"
      :source-ref="image.id"
      :media-url="image.image_url"
      :initial-title="distributionTitle"
      :initial-body="image.prompt"
      :dialog-title="$t('image.distribute')"
    />
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  PictureFilled, Loading, CircleClose,
  Download, Refresh
} from '@element-plus/icons-vue'
import { imageAPI } from '@/api/image'
import type { ImageGeneration, ImageStatus } from '@/types/image'
import DistributionDialog from '@/components/distribution/DistributionDialog.vue'

interface Props {
  modelValue: boolean
  image?: ImageGeneration
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  regenerate: [image: ImageGeneration]
}>()
const { t, locale } = useI18n()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})
const distributionDialogVisible = ref(false)
const distributionTitle = computed(() => {
  const prompt = props.image?.prompt?.trim() || ''
  if (!prompt) {
    return t('image.defaultTitle')
  }
  return prompt.length > 40 ? `${prompt.slice(0, 40)}...` : prompt
})

const getStatusType = (status: ImageStatus) => {
  const types: Record<ImageStatus, any> = {
    pending: 'info',
    processing: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return types[status]
}

const getStatusText = (status: ImageStatus) => {
  return t(`image.status.${status}`)
}

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleString(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US')
}

const downloadImage = () => {
  if (!props.image?.image_url) return
  window.open(props.image.image_url, '_blank')
}

const openDistributionDialog = () => {
  if (!props.image?.image_url) {
    ElMessage.warning(t('image.noDistributableAsset'))
    return
  }
  distributionDialogVisible.value = true
}

const regenerate = () => {
  if (!props.image) return
  emit('regenerate', props.image)
  handleClose()
}

const handleClose = () => {
  visible.value = false
}
</script>

<style scoped>
.image-detail {
  min-height: 400px;
}

.image-preview {
  width: 100%;
  height: 600px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-soft);
  border-radius: 8px;
  overflow: hidden;
}

.preview-image {
  width: 100%;
  height: 100%;
}

.image-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--text-muted);
}

.image-status .el-icon {
  font-size: 64px;
}

.image-status.error {
  color: #f56c6c;
}

.loading-icon {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.error-message {
  margin-top: 8px;
  padding: 12px;
  background: var(--error-light);
  border: 1px solid rgba(248, 113, 113, 0.24);
  border-radius: 4px;
  font-size: 14px;
  color: var(--error);
  max-width: 300px;
  word-wrap: break-word;
}

.image-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
}

.image-error .el-icon {
  font-size: 48px;
}

.image-info {
  height: 600px;
  overflow-y: auto;
}

.prompt-section {
  margin-bottom: 20px;
}

.prompt-section h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.prompt-text {
  padding: 12px;
  background: var(--bg-soft);
  border-radius: 4px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>
