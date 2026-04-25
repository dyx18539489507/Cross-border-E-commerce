<template>
  <div class="drama-settings-container">
    <el-page-header @back="goBack" :title="$t('character.backToProject')">
      <template #content>
        <h2>{{ $t('dramaSettings.title') }}</h2>
      </template>
    </el-page-header>

    <el-card shadow="never" class="main-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="$t('dramaSettings.tabs.basic')" name="basic">
          <el-form :model="form" label-width="100px" style="max-width: 600px">
            <el-form-item :label="$t('dramaSettings.fields.title')">
              <el-input v-model="form.title" />
            </el-form-item>
            <el-form-item :label="$t('dramaSettings.fields.description')">
              <el-input v-model="form.description" type="textarea" :rows="4" />
            </el-form-item>
            <el-form-item :label="$t('dramaSettings.fields.genre')">
              <el-select v-model="form.genre">
                <el-option :label="$t('genres.urban')" value="都市" />
                <el-option :label="$t('genres.costume')" value="古装" />
                <el-option :label="$t('genres.mystery')" value="悬疑" />
                <el-option :label="$t('genres.romance')" value="爱情" />
                <el-option :label="$t('genres.comedy')" value="喜剧" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('dramaSettings.fields.status')">
              <el-select v-model="form.status">
                <el-option :label="$t('dramaSettings.status.draft')" value="draft" />
                <el-option :label="$t('dramaSettings.status.planning')" value="planning" />
                <el-option :label="$t('dramaSettings.status.production')" value="production" />
                <el-option :label="$t('dramaSettings.status.completed')" value="completed" />
                <el-option :label="$t('dramaSettings.status.archived')" value="archived" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveSettings">{{ $t('dramaSettings.actions.save') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane :label="$t('dramaSettings.tabs.distribution')" name="distribution">
          <DistributionSettingsPanel />
        </el-tab-pane>

        <el-tab-pane :label="$t('dramaSettings.tabs.danger')" name="danger">
          <el-alert
            :title="$t('dramaSettings.warningTitle')"
            type="warning"
            :description="$t('dramaSettings.dangerDescription')"
            :closable="false"
            show-icon
          />
          <div class="danger-zone">
            <el-button type="danger" @click="deleteProject">{{ $t('dramaSettings.actions.delete') }}</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { dramaAPI } from '@/api/drama'
import DistributionSettingsPanel from '@/components/distribution/DistributionSettingsPanel.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const dramaId = route.params.id as string

const activeTab = ref('basic')
const form = reactive({
  title: '',
  description: '',
  genre: '',
  status: 'draft' as any
})

const goBack = () => {
  router.push(`/dramas/${dramaId}`)
}

const saveSettings = async () => {
  try {
    await dramaAPI.update(dramaId, form)
    ElMessage.success(t('dramaSettings.messages.saved'))
  } catch (error: any) {
    ElMessage.error(error.message || t('dramaSettings.messages.saveFailed'))
  }
}

const deleteProject = async () => {
  try {
    await ElMessageBox.confirm(
      t('dramaSettings.messages.deleteConfirmMessage'),
      t('dramaSettings.messages.deleteConfirmTitle'),
      {
        confirmButtonText: t('dramaSettings.messages.deleteConfirmButton'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    )
    
    await dramaAPI.delete(dramaId)
    ElMessage.success(t('dramaSettings.messages.deleted'))
    router.push('/dramas')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || t('dramaSettings.messages.deleteFailed'))
    }
  }
}

onMounted(async () => {
  try {
    const drama = await dramaAPI.get(dramaId)
    Object.assign(form, drama)
  } catch (error: any) {
    ElMessage.error(error.message || t('dramaSettings.messages.loadFailed'))
  }
})
</script>

<style scoped>
.drama-settings-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.main-card {
  margin-top: 20px;
}

.danger-zone {
  margin-top: 20px;
  padding: 20px;
  text-align: center;
}
</style>
