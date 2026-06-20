<template>
  <div class="digital-human-center">
    <AppHeader />

    <main class="digital-human-center__main">
      <section class="digital-human-center__hero">
        <div>
          <span class="eyebrow">Digital Human Marketing</span>
          <h1>数字人营销任务中心</h1>
          <p>查看跨境商品讲解视频、数字人口播任务状态和生成结果。</p>
        </div>
        <div class="hero-actions">
          <el-select v-model="statusFilter" placeholder="任务状态" clearable class="status-select" @change="loadTasks">
            <el-option label="全部" value="" />
            <el-option label="排队中" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已完成" value="completed" />
            <el-option label="失败" value="failed" />
          </el-select>
          <el-button type="primary" :loading="loading" @click="loadTasks">刷新</el-button>
        </div>
      </section>

      <el-alert
        v-if="errorMessage"
        class="state-alert"
        type="error"
        :title="errorMessage"
        show-icon
        :closable="false"
      />

      <section v-loading="loading" class="task-panel">
        <el-empty
          v-if="!loading && tasks.length === 0"
          description="暂无数字人营销任务，请先在营销项目中创建数字人口播视频"
        />

        <div v-else class="task-grid">
          <article v-for="task in tasks" :key="task.id" class="task-card">
            <div class="task-card__head">
              <div>
                <span class="task-card__type">跨境商品讲解视频</span>
                <strong>{{ task.result?.speech_text || task.message || '数字人口播任务' }}</strong>
              </div>
              <el-tag :type="statusTagType(task.status)">{{ formatStatus(task.status) }}</el-tag>
            </div>

            <el-progress
              :percentage="Math.max(0, Math.min(100, Number(task.progress || 0)))"
              :status="task.status === 'failed' ? 'exception' : task.status === 'completed' ? 'success' : undefined"
            />

            <p v-if="task.error" class="task-card__error">{{ task.error }}</p>
            <p v-else class="task-card__meta">更新时间：{{ formatTime(task.updated_at) }}</p>

            <video
              v-if="task.video_url"
              class="task-card__video"
              :src="toPlayableUrl(task.video_url)"
              controls
              preload="metadata"
            />

            <div class="task-card__actions">
              <el-button size="small" @click="openResult(task)" :disabled="!task.video_url && task.status !== 'completed'">
                查看结果
              </el-button>
              <el-button size="small" type="danger" plain @click="deleteTask(task)">
                删除
              </el-button>
            </div>
          </article>
        </div>
      </section>
    </main>

    <el-dialog v-model="resultDialogVisible" title="数字人营销结果" width="720px">
      <template v-if="selectedTask">
        <video
          v-if="selectedTask.video_url"
          class="result-video"
          :src="toPlayableUrl(selectedTask.video_url)"
          controls
          preload="metadata"
        />
        <el-descriptions :column="1" border>
          <el-descriptions-item label="任务状态">{{ formatStatus(selectedTask.status) }}</el-descriptions-item>
          <el-descriptions-item label="上游任务">{{ selectedTask.upstream_task_id || selectedTask.task_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="口播文本">{{ selectedTask.result?.speech_text || '-' }}</el-descriptions-item>
          <el-descriptions-item label="使用场景">{{ selectedTask.result?.marketing_use_case || '跨境商品数字人口播视频' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AppHeader } from '@/components/common'
import { digitalHumanAPI } from '@/api/digital-human'
import type { DigitalHumanTask } from '@/api/digital-human'
import { mediaAPI } from '@/api/media'

const loading = ref(false)
const errorMessage = ref('')
const statusFilter = ref('')
const tasks = ref<DigitalHumanTask[]>([])
const selectedTask = ref<DigitalHumanTask | null>(null)
const resultDialogVisible = ref(false)

const loadTasks = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await digitalHumanAPI.getDigitalHumanTasks({
      page: 1,
      page_size: 24,
      status: statusFilter.value || undefined
    })
    tasks.value = result.items || []
  } catch (error: any) {
    tasks.value = []
    errorMessage.value = error.message || '数字人任务加载失败'
  } finally {
    loading.value = false
  }
}

const formatStatus = (status?: string) => {
  const map: Record<string, string> = {
    pending: '排队中',
    processing: '处理中',
    completed: '已完成',
    failed: '失败'
  }
  return map[status || ''] || status || '-'
}

const statusTagType = (status?: string) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'processing') return 'warning'
  return 'info'
}

const formatTime = (value?: string) => {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

const toPlayableUrl = (url: string) => mediaAPI.getMediaProxyUrl(url)

const openResult = async (task: DigitalHumanTask) => {
  selectedTask.value = task
  resultDialogVisible.value = true
  if (!task.result && task.status === 'completed') {
    try {
      const result = await digitalHumanAPI.getDigitalHumanResult(task.id)
      selectedTask.value = {
        ...task,
        result: 'video_url' in result ? result : task.result,
        video_url: 'video_url' in result ? result.video_url : task.video_url
      }
    } catch (error: any) {
      ElMessage.error(error.message || '结果加载失败')
    }
  }
}

const deleteTask = async (task: DigitalHumanTask) => {
  try {
    await ElMessageBox.confirm('删除后不会影响已导出的本地文件，确认删除该数字人任务吗？', '删除数字人任务', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
    await digitalHumanAPI.deleteDigitalHumanTask(task.id)
    ElMessage.success('已删除数字人任务')
    await loadTasks()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

onMounted(loadTasks)
</script>

<style scoped>
.digital-human-center {
  min-height: 100vh;
  padding: 112px 24px 48px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.digital-human-center__main {
  width: min(1180px, 100%);
  margin: 0 auto;
}

.digital-human-center__hero {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: flex-end;
  margin-bottom: 24px;
}

.eyebrow {
  display: block;
  margin-bottom: 8px;
  color: var(--theme-indigo);
  font-size: 13px;
  font-weight: 700;
  text-transform: uppercase;
}

h1 {
  margin: 0;
  font-size: 34px;
  line-height: 1.18;
}

p {
  margin: 8px 0 0;
  color: var(--text-secondary);
}

.hero-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.status-select {
  width: 148px;
}

.state-alert {
  margin-bottom: 18px;
}

.task-panel {
  min-height: 320px;
}

.task-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.task-card {
  padding: 18px;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background: var(--bg-card);
}

.task-card__head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.task-card__head strong {
  display: -webkit-box;
  overflow: hidden;
  color: var(--text-primary);
  font-size: 15px;
  line-height: 1.45;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.task-card__type,
.task-card__meta,
.task-card__error {
  display: block;
  margin-bottom: 4px;
  color: var(--text-secondary);
  font-size: 12px;
}

.task-card__error {
  color: var(--danger);
}

.task-card__video,
.result-video {
  width: 100%;
  margin-top: 12px;
  border-radius: 8px;
  background: #111827;
}

.task-card__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 14px;
}

.result-video {
  max-height: 420px;
  margin-bottom: 18px;
}

@media (max-width: 720px) {
  .digital-human-center {
    padding: 96px 16px 32px;
  }

  .digital-human-center__hero,
  .hero-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .status-select {
    width: 100%;
  }
}
</style>
