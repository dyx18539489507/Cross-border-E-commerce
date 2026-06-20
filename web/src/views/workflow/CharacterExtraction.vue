<template>
  <div class="character-extraction-container">
    <el-page-header @back="goBack" :title="$t('character.backToProject')">
      <template #content>
        <h2>{{ $t('character.title') }}</h2>
      </template>
    </el-page-header>

    <el-card shadow="never" class="main-card">
      <template #header>
        <div class="card-header">
          <h3>{{ $t('character.list') }}</h3>
          <div class="header-actions">
            <el-button @click="addCharacter">
              <el-icon><Plus /></el-icon>
              {{ $t('character.add') }}
            </el-button>
            <el-button type="primary" @click="saveCharacters" :loading="saving">
              {{ $t('character.saveChanges') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-alert
        v-if="loadError"
        class="load-alert"
        type="error"
        :title="loadError"
        show-icon
        :closable="false"
      />

      <div v-loading="loading">
        <el-empty v-if="characters.length === 0" :description="$t('character.empty')" />

        <el-row :gutter="20" v-else>
          <el-col :span="8" v-for="character in characters" :key="character.id">
            <el-card shadow="hover" class="character-card">
              <template #header>
                <div class="character-header">
                  <el-avatar :size="60">{{ character.name[0] }}</el-avatar>
                  <div class="character-info">
                    <h4>{{ character.name }}</h4>
                    <el-tag size="small">{{ formatCharacterRole(character.role) }}</el-tag>
                  </div>
                </div>
              </template>

              <div class="character-details">
                <p><strong>{{ $t('character.personality') }}：</strong>{{ character.personality || '-' }}</p>
                <p><strong>{{ $t('character.appearance') }}：</strong>{{ character.appearance || '-' }}</p>
                <p><strong>{{ $t('character.background') }}：</strong>{{ character.background || character.description || '-' }}</p>
              </div>

              <template #footer>
                <el-button-group style="width: 100%">
                  <el-button size="small" @click="editCharacter(character)">{{ $t('common.edit') }}</el-button>
                  <el-button size="small" type="primary" @click="generateCharacterImage(character)">
                    {{ $t('character.generateImage') }}
                  </el-button>
                </el-button-group>
              </template>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <div class="actions" v-if="characters.length > 0">
        <el-button type="success" size="large" @click="goToNextStep">
          {{ $t('character.nextStep') }}
        </el-button>
      </div>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editDialogVisible" :title="$t('character.edit')" width="600px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item :label="$t('character.name')">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item :label="$t('character.role')">
          <el-input v-model="editForm.role" />
        </el-form-item>
        <el-form-item :label="$t('character.personality')">
          <el-input v-model="editForm.personality" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('character.appearance')">
          <el-input v-model="editForm.appearance" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveCharacter">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import type { Character } from '@/types/drama'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const dramaId = route.params.id as string

const characters = ref<Character[]>([])
const loading = ref(false)
const saving = ref(false)
const loadError = ref('')
const editingCharacterId = ref<number | null>(null)
const editDialogVisible = ref(false)
const editForm = reactive({
  name: '',
  role: '',
  personality: '',
  appearance: '',
  background: ''
})

const goBack = () => {
  router.push(`/projects/${dramaId}`)
}

const addCharacter = () => {
  editingCharacterId.value = null
  Object.assign(editForm, {
    name: '',
    role: '',
    personality: '',
    appearance: '',
    background: ''
  })
  editDialogVisible.value = true
}

const formatCharacterRole = (role?: string) => {
  const normalized = String(role || '').trim().toLowerCase()
  if (normalized === 'main') return t('character.roles.main')
  if (normalized === 'supporting') return t('character.roles.supporting')
  if (normalized === 'minor') return t('character.roles.minor')
  return role || '-'
}

const saveCharacters = async () => {
  saving.value = true
  try {
    await dramaAPI.saveCharacters(dramaId, characters.value.map(toSavePayload))
    await loadCharacters()
    ElMessage.success(t('character.messages.saveSuccess'))
  } catch (error: any) {
    ElMessage.error(error.message || t('character.messages.saveFailed'))
  } finally {
    saving.value = false
  }
}

const editCharacter = (character: Character) => {
  editingCharacterId.value = Number(character.id)
  Object.assign(editForm, character)
  editDialogVisible.value = true
}

const saveCharacter = () => {
  if (!editForm.name.trim()) {
    ElMessage.warning(t('character.messages.enterName'))
    return
  }

  const payload = {
    ...editForm,
    name: editForm.name.trim()
  }

  const currentId = editingCharacterId.value
  if (currentId === null) {
    characters.value.push({
      id: -Date.now(),
      drama_id: dramaId,
      ...payload,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Character)
  } else {
    const index = characters.value.findIndex((item) => Number(item.id) === currentId)
    if (index >= 0) {
      characters.value[index] = {
        ...characters.value[index],
        ...payload,
        updated_at: new Date().toISOString()
      }
    }
  }

  editingCharacterId.value = null
  editDialogVisible.value = false
  saveCharacters()
}

const generateCharacterImage = (character: Character) => {
  router.push(`/projects/${dramaId}/images/characters?character=${character.id}`)
}

const goToNextStep = () => {
  router.push(`/projects/${dramaId}/images/characters`)
}

const toSavePayload = (character: Character) => ({
  name: character.name,
  role: character.role || 'supporting',
  appearance: character.appearance || '',
  personality: character.personality || '',
  description: character.description || character.background || '',
  background: character.background || character.description || '',
  image_url: character.image_url
})

const loadCharacters = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const result = await dramaAPI.getCharacters(dramaId)
    characters.value = Array.isArray(result) ? result as Character[] : []
  } catch (error: any) {
    characters.value = []
    loadError.value = error.message || t('character.messages.loadFailed')
  } finally {
    loading.value = false
  }
}

onMounted(loadCharacters)
</script>

<style scoped>
.character-extraction-container {
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
}

.character-card {
  margin-bottom: 20px;
}

.character-header {
  display: flex;
  gap: 16px;
  align-items: center;
}

.character-info h4 {
  margin: 0 0 8px 0;
}

.character-details p {
  margin: 8px 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.actions {
  margin-top: 30px;
  text-align: center;
}
</style>
