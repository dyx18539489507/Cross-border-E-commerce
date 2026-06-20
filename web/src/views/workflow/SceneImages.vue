<template>
  <div class="scene-images-container">
    <el-page-header @back="goBack" title="返回项目">
      <template #content>
        <h2>营销场景图片生成</h2>
      </template>
    </el-page-header>

    <el-card shadow="never" class="main-card" v-loading="loading">
      <el-alert
        v-if="loadError"
        class="load-alert"
        type="error"
        :title="loadError"
        show-icon
        :closable="false"
      />

      <el-empty
        v-if="!loading && episodes.length === 0"
        description="暂无可生成图片的营销分镜，请先在脚本工作台生成分镜"
      />

      <el-tabs v-else v-model="activeEpisode">
        <el-tab-pane 
          v-for="episode in episodes" 
          :key="episode.id"
          :label="`内容 ${episode.episode_number}`"
          :name="episode.id"
        >
          <el-empty
            v-if="!episode.scenes || episode.scenes.length === 0"
            description="该内容暂无营销场景"
          />
          <el-row :gutter="20">
            <el-col :span="8" v-for="scene in episode.scenes" :key="scene.id">
              <el-card shadow="hover" class="scene-card" :class="{ 'has-image': scene.image_url }">
                <template #header>
                  <div class="scene-header">
                    <span class="scene-number">场景 {{ scene.storyboard_number || scene.id }}</span>
                    <el-tag size="small">{{ scene.location }}</el-tag>
                  </div>
                </template>

                <div class="scene-preview">
                  <img v-if="scene.image_url" :src="scene.image_url" :alt="scene.title" />
                  <div v-else class="placeholder">
                    <el-icon :size="48"><Picture /></el-icon>
                    <p>未生成</p>
                  </div>
                </div>

                <div class="scene-info">
                  <h4>{{ scene.title || scene.location || '营销场景' }}</h4>
                  <p class="description">{{ scene.description || scene.prompt }}</p>
                </div>

                <el-button 
                  type="primary" 
                  @click="generateImage(scene)"
                  :loading="generatingId === scene.id"
                  :disabled="!!generatingId && generatingId !== scene.id"
                  style="width: 100%"
                >
                  {{ scene.image_url ? '重新生成' : '生成图片' }}
                </el-button>
              </el-card>
            </el-col>
          </el-row>
        </el-tab-pane>
      </el-tabs>

      <div class="actions">
        <el-button type="success" size="large" @click="goToNextStep" :disabled="!allImagesGenerated">
          下一步：营销视频生成
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { imageAPI } from '@/api/image'
import type { Episode, Scene } from '@/types/drama'

const route = useRoute()
const router = useRouter()
const dramaId = route.params.id as string

const episodes = ref<Episode[]>([])
const activeEpisode = ref<string>()
const generatingId = ref<string>()
const loading = ref(false)
const loadError = ref('')

const allImagesGenerated = computed(() => {
  return episodes.value.length > 0 && episodes.value.every(ep =>
    ep.scenes?.every(s => s.image_url)
  )
})

const goBack = () => {
  router.push(`/projects/${dramaId}`)
}

const generateImage = async (scene: Scene) => {
  generatingId.value = scene.id
  try {
    // 构建场景提示词
    let prompt = `${scene.location || 'product marketing scene'}, ${scene.time || 'daytime'}`
    if (scene.description || scene.prompt) {
      prompt += `, ${scene.description || scene.prompt}`
    }
    prompt += ', cross-border ecommerce marketing visual, clean product-focused composition, high quality'
    
    const result = await imageAPI.generateImage({
      drama_id: dramaId,
      scene_id: Number(scene.id),
      image_type: 'scene',
      prompt: prompt
    })

    if (result?.image_url) {
      scene.image_url = result.image_url
    }
    scene.image_generation_status = result?.status || 'pending'
    
    ElMessage.success('场景图片生成任务已提交')
  } catch (error: any) {
    ElMessage.error(error.message || '生成失败')
  } finally {
    generatingId.value = undefined
  }
}

const goToNextStep = () => {
  router.push(`/projects/${dramaId}/videos`)
}

const normalizeEpisodes = (rawEpisodes?: Episode[]): Episode[] => {
  return (rawEpisodes || [])
    .map((episode) => ({
      ...episode,
      scenes: episode.scenes || []
    }))
    .filter((episode) => (episode.scenes?.length || 0) > 0)
}

const loadScenes = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const project = await dramaAPI.get(dramaId)
    episodes.value = normalizeEpisodes(project.episodes)
    activeEpisode.value = episodes.value[0]?.id
  } catch (error: any) {
    episodes.value = []
    activeEpisode.value = undefined
    loadError.value = error.message || '加载营销场景失败'
  } finally {
    loading.value = false
  }
}

onMounted(loadScenes)
</script>

<style scoped>
.scene-images-container {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  color: var(--text-primary);
}

.main-card {
  margin-top: 20px;
}

.load-alert {
  margin-bottom: 16px;
}

.scene-card {
  margin-bottom: 20px;
}

.scene-card.has-image {
  border-color: #67c23a;
}

.scene-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.scene-number {
  font-weight: 500;
}

.scene-preview {
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-soft);
  border-radius: 8px;
  margin-bottom: 16px;
}

.scene-preview img {
  max-width: 100%;
  max-height: 200px;
  border-radius: 8px;
}

.placeholder {
  text-align: center;
  color: var(--text-muted);
}

.placeholder p {
  margin-top: 8px;
}

.scene-info h4 {
  margin: 8px 0;
}

.scene-info .description {
  color: var(--text-secondary);
  font-size: 13px;
  margin: 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.actions {
  margin-top: 30px;
  text-align: center;
}
</style>
