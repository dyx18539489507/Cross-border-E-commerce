<template>
  <div class="workflow-container">
    <AppHeader :fixed="false" :show-logo="false">
      <template #left>
        <el-button text @click="goBack" class="back-btn">
          <el-icon><ArrowLeft /></el-icon>
          <span>{{ $t('dramaWorkflow.returnToList') }}</span>
        </el-button>
        <h2 class="drama-title">{{ drama?.title }}</h2>
        <el-tag :type="getStatusType(drama?.status)" size="small">{{ getStatusText(drama?.status) }}</el-tag>
      </template>
      <template #center>
        <!-- 步骤进度条 -->
        <div class="custom-steps">
          <div class="step-item" :class="{ active: currentStep >= 0, current: currentStep === 0 }">
            <div class="step-circle">1</div>
            <span class="step-text">{{ $t('dramaWorkflow.episodeScript', { number: currentEpisodeNumber }) }}</span>
          </div>
          <el-icon class="step-arrow"><ArrowRight /></el-icon>
          <div class="step-item" :class="{ active: currentStep >= 1, current: currentStep === 1 }">
            <div class="step-circle">2</div>
            <span class="step-text">{{ $t('dramaWorkflow.storyboardBreakdown') }}</span>
          </div>
          <el-icon class="step-arrow"><ArrowRight /></el-icon>
          <div class="step-item" :class="{ active: currentStep >= 2, current: currentStep === 2 }">
            <div class="step-circle">3</div>
            <span class="step-text">{{ $t('dramaWorkflow.characterImages') }}</span>
          </div>
        </div>
      </template>
    </AppHeader>

    <!-- 当前阶段内容区域 -->
    <div class="stage-area">
      <!-- 阶段 0: 剧本生成 -->
      <el-card v-show="currentStep === 0" shadow="never" class="stage-card stage-card-fullscreen">
        <div class="stage-body stage-body-fullscreen">
          <!-- 初始状态：显示创建第一章按钮 -->
          <div v-if="!hasScript && !showScriptInput" class="create-chapter-prompt">
            <el-empty :description="$t('dramaWorkflow.createChapterPrompt')">
              <el-button 
                type="primary" 
                size="large"
                @click="startCreateChapter"
                :icon="Document"
              >
                {{ $t('dramaWorkflow.createChapter', { number: currentEpisodeNumber }) }}
              </el-button>
            </el-empty>
          </div>

          <!-- 未生成剧本时显示表单 -->
          <div v-if="!hasScript && showScriptInput" class="generation-form">
            <div class="script-input-header">
              <el-button
                type="primary"
                :icon="MagicStick"
                @click="generateScriptByAI"
                :loading="generatingScript"
              >
                {{ generatingScript ? 'AI生成中...' : '随机生成' }}
              </el-button>
            </div>

            <el-input
              v-model="scriptContent"
              type="textarea"
              placeholder="请输入剧本内容..."
              class="script-textarea script-textarea-fullscreen"
              :disabled="generatingScript"
            />

            <div class="action-buttons-inline">
              <el-button 
                type="primary" 
                size="default" 
                @click="saveChapterScript"
                :disabled="!scriptContent.trim() || generatingScript"
              >
                <el-icon><Check /></el-icon>
                <span>保存章节</span>
              </el-button>
              <el-button
                type="success"
                plain
                size="default"
                @click="openDigitalHumanDialog"
              >
                <el-icon><VideoCamera /></el-icon>
                <span>试试数字人？</span>
              </el-button>
            </div>
          </div>

          <div v-if="hasScript" class="overview-section">
            <el-divider />
            
            <div class="episode-info">
              <h3>第{{ currentEpisodeNumber }}集剧本内容</h3>
              <el-tag type="success" size="large">当前正在制作</el-tag>
            </div>
            <div class="overview-content">
              <div class="overview-item script-content-display">
                <el-input 
                  v-model="currentEpisode.script_content"
                  type="textarea"
                  :rows="15"
                  readonly
                  class="script-display"
                />
              </div>
            </div>

            <el-divider />

            <div class="action-buttons">
              <el-button 
                type="success"
                size="large"
                @click="nextStep"
              >
                开始分镜拆解
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
      </el-card>

      <el-dialog
        v-model="digitalHumanDialogVisible"
        title="数字人制作"
        width="560px"
        class="digital-human-dialog dialog-form-safe"
        :lock-scroll="true"
        :append-to-body="true"
        @close="resetDigitalHumanForm"
      >
        <el-form
          ref="digitalHumanDialogFormRef"
          label-position="top"
          class="digital-human-form long-form form-enter-flow"
          @keydown.enter="handleFormEnterNavigation"
        >
          <el-form-item>
          <template #label>
            <span class="digital-human-required-label">
              上传角色
              <span class="digital-human-required-dot">*</span>
            </span>
          </template>
          <el-upload
            class="digital-human-upload"
            :auto-upload="false"
            :limit="1"
            accept="image/*"
            :show-file-list="false"
            :file-list="digitalHumanImageList"
            :before-upload="beforeDigitalHumanImageUpload"
            :on-change="handleDigitalHumanImageChange"
            :on-remove="handleDigitalHumanImageRemove"
          >
              <el-button
                v-if="digitalHumanImageList.length"
                type="primary"
                class="digital-human-upload-btn digital-human-role-btn"
                :title="digitalHumanImageList[0].name"
                @click.stop.prevent="openDigitalHumanImagePreview"
              >
                <span class="digital-human-role-btn-label">{{ digitalHumanImageList[0].name }}</span>
                <el-icon
                  class="digital-human-role-btn-clear"
                  @click.stop.prevent="handleDigitalHumanImageRemove"
                >
                  <Close />
                </el-icon>
              </el-button>
              <el-button v-else type="primary" class="digital-human-upload-btn digital-human-role-btn" :icon="Upload">
                选择角色
              </el-button>
          </el-upload>
          </el-form-item>
        <el-form-item>
          <template #label>
            <span class="digital-human-required-label">
              上传音色
              <span class="digital-human-required-dot">*</span>
            </span>
          </template>
          <div class="digital-human-audio-row">
            <el-popover
              v-model:visible="voiceLibraryVisible"
              trigger="click"
              placement="bottom-start"
              width="680"
              popper-class="voice-library-popover"
              :teleported="true"
              @show="handleVoicePopoverShow"
            >
              <div class="voice-library-panel">
                <div class="voice-library-toolbar">
                  <el-input
                    v-model="voiceLibrarySearch"
                    :placeholder="$t('workflow.voiceLibrary.searchPlaceholder')"
                    clearable
                  />
                </div>
                <div class="voice-library-filters">
                  <el-select
                    v-model="voiceGenderFilter"
                    :placeholder="$t('workflow.voiceLibrary.filters.gender')"
                    size="small"
                    class="voice-filter"
                    clearable
                    :teleported="false"
                  >
                    <el-option v-for="g in voiceGenderOptions" :key="g" :label="g" :value="g" />
                  </el-select>
                  <el-select
                    v-model="voiceAgeFilter"
                    :placeholder="$t('workflow.voiceLibrary.filters.age')"
                    size="small"
                    class="voice-filter"
                    clearable
                    :teleported="false"
                  >
                    <el-option v-for="a in voiceAgeOptions" :key="a" :label="a" :value="a" />
                  </el-select>
                  <el-select
                    v-model="voiceLanguageFilter"
                    :placeholder="$t('workflow.voiceLibrary.filters.language')"
                    size="small"
                    class="voice-filter"
                    clearable
                    :teleported="false"
                  >
                    <el-option v-for="l in voiceLanguageOptions" :key="l" :label="l" :value="l" />
                  </el-select>
                  <el-select
                    v-model="voiceCategoryFilter"
                    :placeholder="$t('workflow.voiceLibrary.filters.category')"
                    size="small"
                    class="voice-filter"
                    clearable
                    :teleported="false"
                  >
                    <el-option v-for="c in voiceCategoryOptions" :key="c" :label="c" :value="c" />
                  </el-select>
                </div>
                <div v-if="voiceLibraryLoading" class="voice-library-loading">
                  <el-icon class="is-loading"><Loading /></el-icon>
                  <span>{{ $t('common.loading') }}</span>
                </div>
                <div v-else-if="voiceLibraryError" class="voice-library-error">{{ voiceLibraryError }}</div>
                <el-scrollbar v-else height="320">
                  <div class="voice-library-grid">
                    <button class="voice-card voice-card-create" type="button" :disabled="creatingCustomVoice" @click="openCreateVoice">
                      <el-icon v-if="creatingCustomVoice"><Loading /></el-icon>
                      <el-icon v-else><Plus /></el-icon>
                      <span>{{ creatingCustomVoice ? '创建中...' : '上传音色' }}</span>
                    </button>
                    <button
                      v-for="voice in filteredVoiceLibrary"
                      :key="voice.id"
                      class="voice-card"
                      :class="{
                        'is-trial-playing': voiceTrialPlayingId === voice.id,
                        'is-selected': !!selectedVoice && selectedVoice.id === voice.id
                      }"
                      type="button"
                      @click="selectVoice(voice)"
                    >
                      <span
                        class="voice-card-play"
                        :class="{ 'is-playing': voiceTrialPlayingId === voice.id }"
                        @click.stop="toggleVoiceTrial(voice)"
                      >
                        <el-icon v-if="voiceTrialPlayingId === voice.id"><VideoPause /></el-icon>
                        <el-icon v-else><VideoPlay /></el-icon>
                      </span>
                      <div class="voice-card-text">
                        <div class="voice-card-name">{{ voice.name }}</div>
                      </div>
                    </button>
                  </div>
                </el-scrollbar>
                <audio
                  ref="voiceTrialAudioRef"
                  class="voice-trial-audio"
                  :src="voiceTrialUrl"
                  preload="none"
                  @ended="stopVoiceTrial"
                />
              </div>
              <template #reference>
                <el-button
                  type="primary"
                  class="digital-human-upload-btn digital-human-voice-btn"
                  :icon="selectedVoice ? undefined : Upload"
                  :title="selectedVoice ? `${selectedVoice.name}${selectedVoice.voice_type ? ` (${selectedVoice.voice_type})` : ''}` : ''"
                >
                  <span class="digital-human-voice-btn-label">
                    {{ selectedVoice ? selectedVoice.name : '选择音色' }}
                  </span>
                  <el-icon
                    v-if="selectedVoice"
                    class="digital-human-voice-btn-clear"
                    @click.stop.prevent="clearSelectedVoice"
                  >
                    <Close />
                  </el-icon>
                </el-button>
              </template>
            </el-popover>
            <input
              ref="digitalHumanUploadVoiceInputRef"
              class="digital-human-hidden-input"
              type="file"
              accept="audio/*,.mp3,.wav,.m4a,.ogg,.aac,.flac,.mp4"
              @change="handleUploadVoiceFromCard"
            />
            <!-- <div class="digital-human-hint-inline">音频时长需小于60秒，支持 mp3/wav/m4a 等格式</div> -->
          </div>
          <el-upload
            class="digital-human-upload"
            :auto-upload="false"
            :limit="1"
            accept="audio/*"
            :show-file-list="false"
            :file-list="digitalHumanAudioList"
            :before-upload="beforeDigitalHumanAudioUpload"
            :on-change="handleDigitalHumanAudioChange"
            :on-remove="handleDigitalHumanAudioRemove"
          >
            <el-button type="info" plain class="digital-human-upload-btn digital-human-upload-secondary" :icon="Upload">
              上传音色音频
            </el-button>
          </el-upload>
          <div
            v-if="digitalHumanAudioList.length"
            class="digital-human-file-name"
            role="button"
            tabindex="0"
            @click="openDigitalHumanAudioPreview"
          >
            {{ digitalHumanAudioList[0].name }}
          </div>
          <div v-if="digitalHumanAudioPreviewVisible && digitalHumanAudioPreview" class="digital-human-audio">
            <audio
              ref="digitalHumanAudioRef"
              :key="digitalHumanAudioPreview"
              :src="digitalHumanAudioPreview"
              controls
              preload="metadata"
            />
          </div>
          </el-form-item>
        <el-form-item>
          <template #label>
            <span class="digital-human-required-label">
              说话内容
              <span class="digital-human-required-dot">*</span>
            </span>
          </template>
          <el-input
            v-model="digitalHumanForm.speechText"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 3 }"
            placeholder="请输入你希望角色说出的内容（将自动转语音）"
            class="digital-human-textarea"
          />
        </el-form-item>
        <el-form-item label="动作描述（可选）">
          <el-input
            v-model="digitalHumanForm.motionText"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 3 }"
            placeholder="添加动作描述和镜头语言，如人物展现专业播报姿态，手势自然摆动，镜头保持中景平视，眼神专注且自信。"
            class="digital-human-textarea"
          />
        </el-form-item>
        </el-form>

        <div v-if="digitalHumanResultUrl" class="digital-human-result">
          <div class="digital-human-result-title">生成结果</div>
          <video :src="digitalHumanPlayableResultUrl" controls preload="metadata" />
          <el-button type="primary" plain size="small" @click="downloadDigitalHumanVideo">下载视频</el-button>
        </div>

        <template #footer>
          <el-button @click="digitalHumanDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="digitalHumanLoading"
            :disabled="!digitalHumanCanGenerate"
            @click="submitDigitalHuman"
          >
            开始生成
          </el-button>
      </template>
    </el-dialog>
      <el-image-viewer
        v-if="digitalHumanImagePreviewVisible && digitalHumanImagePreview"
        :url-list="[digitalHumanImagePreview]"
        :teleported="true"
        @close="digitalHumanImagePreviewVisible = false"
      />

      <!-- 阶段 1: 分镜拆解 -->
      <el-card v-show="currentStep === 1" shadow="never" class="stage-card">
        <template #header>
          <div class="stage-header">
            <div class="header-left">
              <el-icon :size="48" color="#409eff"><Film /></el-icon>
              <div class="header-info">
                <h2>分镜拆解</h2>
                <p>将第{{ currentEpisodeNumber }}集剧本拆分为多个镜头</p>
              </div>
            </div>
            <el-tag v-if="currentEpisode?.shots?.length" type="success" size="large">
              已拆分 {{ currentEpisode.shots.length }} 个镜头
            </el-tag>
          </div>
        </template>

        <div class="stage-body">
          <!-- 分镜列表 -->
          <div v-if="currentEpisode?.shots && currentEpisode.shots.length > 0" class="shots-list">
            <div class="shots-header">
              <h3>镜头列表</h3>
              <el-button 
                type="primary"
                @click="parseShotsToCharacters"
                :loading="parsingCharacters"
                :icon="User"
              >
                解析角色
              </el-button>
            </div>
            
            <el-table :data="currentEpisode.shots" border stripe style="margin-top: 16px;">
              <el-table-column type="index" label="镜头" width="80" />
              <el-table-column prop="content" label="镜头内容" show-overflow-tooltip />
              <el-table-column label="时长" width="100">
                <template #default="{ row }">
                  {{ row.duration || '-' }}秒
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row, $index }">
                  <el-button 
                    type="primary" 
                    size="small"
                    @click="editShot(row, $index)"
                  >
                    划分
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            
            <div class="action-buttons" style="margin-top: 24px;">
              <el-button 
                @click="regenerateShots"
                :icon="MagicStick"
              >
                {{ $t('dramaWorkflow.reGenerateShots') }}
              </el-button>
              <el-button 
                type="success"
                @click="nextStep"
                :disabled="!hasCharacters"
              >
                {{ $t('dramaWorkflow.nextStepCharacterImages') }}
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>
          
          <!-- 未拆分时显示 -->
          <div v-else class="empty-shots">
            <el-empty :description="$t('dramaWorkflow.splitStoryboardFirst')">
              <el-button 
                type="primary" 
                @click="generateShots"
                :loading="generatingShots"
                :icon="MagicStick"
              >
                {{ generatingShots ? $t('dramaWorkflow.aiSplitting') : $t('dramaWorkflow.aiAutoSplit') }}
              </el-button>
            </el-empty>
          </div>
        </div>
      </el-card>

      <!-- 阶段 2: 角色图片 -->
      <el-card v-show="currentStep === 2" shadow="never" class="stage-card stage-card-fullscreen">
        <div class="stage-body stage-body-fullscreen">

          <div class="batch-toolbar-compact">
            <div class="toolbar-left">
              <el-checkbox v-model="selectAllCharacters" @change="handleSelectAllCharacters" :indeterminate="isCharacterIndeterminate">
                {{ $t('common.selectAll') }}
              </el-checkbox>
              <span class="selection-info">{{ $t('dramaWorkflow.selected') }} {{ selectedCharacterIds.length }}/{{ drama?.characters?.length || 0 }}</span>
            </div>
            <div class="toolbar-right">
              <span class="stats-compact">{{ $t('dramaWorkflow.characterCount') }}: {{ drama?.characters?.length || 0 }} | {{ $t('dramaWorkflow.generated') }}: {{ characterImagesCount || 0 }}</span>
              <el-button 
                type="primary" 
                size="small"
                :disabled="selectedCharacterIds.length === 0"
                :loading="batchGenerating"
                @click="batchGenerateCharacterImages"
                :icon="MagicStick"
              >
                {{ $t('dramaWorkflow.batchGenerate') }}({{ selectedCharacterIds.length }})
              </el-button>
              <el-button 
                type="success"
                size="small"
                @click="nextStep" 
                :disabled="!allCharactersHaveImages"
              >
                {{ $t('dramaWorkflow.nextStep') }}
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>

          <div class="character-cards-area-fullscreen">
            <el-row :gutter="16">
              <el-col :span="4" v-for="character in drama?.characters" :key="character.id">
                <el-card shadow="hover" class="character-card" :class="{ 'has-image': character.image_url, 'selected': isCharacterSelected(character.id) }">
                  <el-checkbox 
                    class="card-checkbox" 
                    :model-value="isCharacterSelected(character.id)" 
                    @change="toggleCharacterSelection(character.id)"
                  />
                  <div class="character-preview">
                    <img v-if="character.image_url" :src="fixImageUrl(character.image_url)" :alt="character.name" />
                    <el-avatar v-else :size="120">{{ character.name[0] }}</el-avatar>
                  </div>
                  
                  <div class="character-info">
                    <h4>{{ character.name }}</h4>
                    <el-tag :type="character.role === 'main' ? 'danger' : 'info'" size="small">
                      {{ character.role === 'main' ? '主角' : character.role === 'supporting' ? '配角' : '次要' }}
                    </el-tag>
                    <p class="desc">{{ character.appearance || character.description }}</p>
                    <el-button 
                      size="small" 
                      text 
                      type="primary"
                      @click="editCharacterDescription(character)"
                      :icon="Edit"
                    >
                      编辑描述
                    </el-button>
                  </div>

                  <div v-if="character.image_generation_status === 'failed'" class="error-hint" style="margin-bottom: 10px;">
                    <el-alert type="error" :closable="false" show-icon>
                      <template #title>
                        生成失败
                      </template>
                      <template #default v-if="character.image_generation_error">
                        {{ character.image_generation_error }}
                      </template>
                    </el-alert>
                  </div>

                  <div class="character-actions">
                    <el-button 
                      type="primary" 
                      size="small"
                      :loading="generatingCharacterIds.includes(character.id)"
                      @click="generateCharacterImage(character)"
                      :icon="MagicStick"
                    >
                      <span v-if="generatingCharacterIds.includes(character.id)">生成中...</span>
                      <span v-else-if="character.image_generation_status === 'failed'">重新生成</span>
                      <span v-else>AI生成形象</span>
                    </el-button>
                    <el-button 
                      size="small"
                      @click="openUploadDialog(character)"
                      :icon="UploadFilled"
                    >
                      上传图片
                    </el-button>
                    <el-button 
                      size="small"
                      @click="openCharacterLibrary(character)"
                      :icon="FolderOpened"
                    >
                      从角色库选择
                    </el-button>
                    <el-button 
                      v-if="character.image_url"
                      size="small"
                      type="success"
                      plain
                      @click="addToCharacterLibrary(character)"
                      :icon="Plus"
                    >
                      添加到角色库
                    </el-button>
                    <el-button 
                      size="small"
                      type="danger"
                      plain
                      @click="deleteCharacter(character)"
                      :icon="Delete"
                    >
                      删除角色
                    </el-button>
                  </div>
                </el-card>
              </el-col>
              
              <!-- 添加角色卡片 -->
              <el-col :span="4">
                <el-card shadow="hover" class="character-card add-character-card" @click="openAddCharacterDialog">
                  <div class="add-character-content">
                    <el-icon :size="40" color="#909399"><Plus /></el-icon>
                    <span class="add-text">添加角色</span>
                  </div>
                </el-card>
              </el-col>
            </el-row>
          </div>

        </div>
      </el-card>

      <!-- 阶段 3: 剧集制作 -->
      <el-card v-show="currentStep === 3" shadow="never" class="stage-card">
        <template #header>
          <div class="stage-header">
            <div class="header-left">
              <el-icon :size="48" color="#409eff"><Film /></el-icon>
              <div class="header-info">
                <h2>剧集制作</h2>
                <p>对每一集进行分镜、图片、视频、剪辑</p>
              </div>
            </div>
            <el-tag v-if="completedEpisodesCount > 0" type="success" size="large">
              {{ completedEpisodesCount }}/{{ drama?.episodes?.length || 0 }} 已完成
            </el-tag>
          </div>
        </template>

        <div class="stage-body">
          <div class="stats-row">
            <div class="stat-box">
              <div class="stat-label">总剧集数</div>
              <div class="stat-value">{{ drama?.episodes?.length || 0 }}</div>
            </div>
            <div class="stat-box">
              <div class="stat-label">已完成</div>
              <div class="stat-value">{{ completedEpisodesCount || 0 }}</div>
            </div>
            <div class="stat-box">
              <div class="stat-label">总进度</div>
              <div class="stat-value">{{ overallProgress }}%</div>
            </div>
          </div>

          <el-divider />

          <h3>剧集列表</h3>
          <el-table :data="sortedEpisodes" border size="small" max-height="400" style="margin-bottom: 24px;">
            <el-table-column prop="episode_number" label="集数" width="80" sortable />
            <el-table-column prop="title" label="标题" width="200" />
            <el-table-column prop="description" label="简介" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'completed' ? 'success' : 'info'" size="small">
                  {{ row.status === 'completed' ? '已完成' : '制作中' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="时长" width="100">
              <template #default="{ row }">
                {{ row.duration ? `${row.duration}秒` : '-' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button 
                  type="primary" 
                  size="small"
                  @click="goToEpisodeDetail(row.id)"
                >
                  进入制作
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="action-area">
            <h3>操作</h3>
            <p class="hint-text">点击进入剧集列表，对每一集进行分镜、背景、合成、视频、剪辑</p>
            <el-button type="primary" size="large" @click="goToEpisodeList" class="main-action-btn">
              <el-icon :size="20"><Film /></el-icon>
              <span>进入剧集制作</span>
            </el-button>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 编辑角色描述对话框 -->
    <el-dialog v-model="editDescDialogVisible" title="编辑角色描述" width="600px" class="dialog-form-safe">
      <el-form
        v-if="editingCharacter"
        ref="editDescFormRef"
        label-width="100px"
        class="long-form form-enter-flow"
        @keydown.enter="handleFormEnterNavigation"
      >
        <el-form-item label="角色名称">
          <el-input v-model="editingCharacter.name" disabled />
        </el-form-item>
        <el-form-item label="外貌描述">
          <el-input 
            v-model="editingCharacter.appearance" 
            type="textarea" 
            :rows="4"
            placeholder="描述角色的外貌特征，如身高、体型、发型、穿着等"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="性格描述">
          <el-input 
            v-model="editingCharacter.personality" 
            type="textarea" 
            :rows="3"
            placeholder="描述角色的性格特点"
            maxlength="300"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="角色简介">
          <el-input 
            v-model="editingCharacter.description" 
            type="textarea" 
            :rows="3"
            placeholder="角色的背景故事或简介"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDescDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveCharacterDescription" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加角色对话框 -->
    <el-dialog v-model="addCharacterDialogVisible" title="添加新角色" width="600px" class="dialog-form-safe">
      <el-form
        ref="addCharacterFormRef"
        :model="newCharacter"
        label-width="80px"
        class="long-form form-enter-flow"
        @keydown.enter="handleFormEnterNavigation"
      >
        <el-form-item label="角色名称" required>
          <el-input v-model="newCharacter.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色类型">
          <el-select v-model="newCharacter.role" placeholder="请选择角色类型">
            <el-option label="主角" value="main" />
            <el-option label="配角" value="supporting" />
            <el-option label="次要" value="minor" />
          </el-select>
        </el-form-item>
        <el-form-item label="外貌描述">
          <el-input 
            v-model="newCharacter.appearance" 
            type="textarea"
            :rows="3"
            placeholder="描述角色的外貌特征，如身高、体型、发型、穿着等"
          />
        </el-form-item>
        <el-form-item label="性格特点">
          <el-input 
            v-model="newCharacter.personality" 
            type="textarea"
            :rows="3"
            placeholder="描述角色的性格特点、行为习惯等"
          />
        </el-form-item>
        <el-form-item label="角色描述" required>
          <el-input 
            v-model="newCharacter.description" 
            type="textarea"
            :rows="4"
            placeholder="请输入角色的详细描述，包括背景故事、角色关系等"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addCharacterDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addCharacter" :loading="saving">添加</el-button>
      </template>
    </el-dialog>

    <!-- 角色库选择对话框 -->
    <el-dialog v-model="libraryDialogVisible" title="从角色库选择" width="800px">
      <div class="library-grid" v-if="characterLibrary.length > 0">
        <el-row :gutter="16">
          <el-col :span="6" v-for="item in characterLibrary" :key="item.id">
            <el-card 
              shadow="hover" 
              class="library-item" 
              @click="selectFromLibrary(item)"
              :body-style="{ padding: '10px' }"
            >
              <img :src="fixImageUrl(item.image_url)" :alt="item.name" class="library-image" />
              <div class="library-info">
                <div class="library-name">{{ item.name }}</div>
                <el-tag size="small">{{ item.category || '未分类' }}</el-tag>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
      <el-empty v-else description="角色库为空，生成形象后可添加到角色库" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, reactive, nextTick, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadProps, UploadUserFile } from 'element-plus'
import {
  MagicStick,
  Film,
  User,
  Picture,
  ArrowLeft,
  ArrowRight,
  Edit,
  Document,
  ArrowDown,
  Upload,
  UploadFilled,
  FolderOpened,
  Plus,
  Loading,
  WarningFilled,
  InfoFilled,
  Check,
  Delete,
  VideoCamera,
  Close,
  VideoPlay,
  VideoPause
} from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { generationAPI } from '@/api/generation'
import { characterLibraryAPI } from '@/api/character-library'
import { digitalHumanAPI } from '@/api/digital-human'
import { voiceLibraryAPI } from '@/api/voice-library'
import type { VoiceLibraryItem } from '@/api/voice-library'
import request from '@/utils/request'
import { handleFormEnterNavigation } from '@/utils/formFocus'
import type { Drama, DramaStatus } from '@/types/drama'
import { AppHeader } from '@/components/common'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const drama = ref<Drama>()
const currentStep = ref(0)
const currentEpisodeNumber = ref(1) // 当前正在创作的集数
const generatingCharacterIds = ref<(number | string)[]>([])
const batchGenerating = ref(false)
const selectedCharacterIds = ref<(number | string)[]>([])
const selectAllCharacters = ref(false)
const generatingScript = ref(false)
const scriptContent = ref('')
const showScriptInput = ref(false) // 控制是否显示剧本输入框

const digitalHumanDialogVisible = ref(false)
const digitalHumanDialogFormRef = ref<{ $el?: HTMLElement } | null>(null)
const digitalHumanLoading = ref(false)
const digitalHumanResultUrl = ref('')
const digitalHumanImageList = ref<UploadUserFile[]>([])
const digitalHumanAudioList = ref<UploadUserFile[]>([])
const digitalHumanImagePreview = ref('')
const digitalHumanAudioPreview = ref('')
const digitalHumanImagePreviewVisible = ref(false)
const digitalHumanAudioPreviewVisible = ref(false)
const digitalHumanAudioRef = ref<HTMLAudioElement | null>(null)
const digitalHumanUploadVoiceInputRef = ref<HTMLInputElement | null>(null)
const voiceLibraryVisible = ref(false)
const voiceLibraryLoading = ref(false)
const voiceLibraryList = ref<VoiceLibraryItem[]>([])
const voiceLibrarySearch = ref('')
const voiceLibraryError = ref('')
const selectedVoice = ref<VoiceLibraryItem | null>(null)
const voiceGenderFilter = ref<string | null>(null)
const voiceAgeFilter = ref<string | null>(null)
const voiceLanguageFilter = ref<string | null>(null)
const voiceCategoryFilter = ref<string | null>(null)
const voiceTrialAudioRef = ref<HTMLAudioElement | null>(null)
const voiceTrialUrl = ref('')
const voiceTrialPlayingId = ref('')
const creatingCustomVoice = ref(false)
const digitalHumanForm = reactive({
  imageFile: null as File | null,
  audioFile: null as File | null,
  speechText: '',
  motionText: ''
})

const getUploadRawFile = (file: any): File | null => {
  const raw = file?.raw ?? file?.originFileObj ?? null
  return raw instanceof File ? raw : null
}

const selectedDigitalHumanImageFile = computed(() => {
  return digitalHumanForm.imageFile || getUploadRawFile(digitalHumanImageList.value?.[0])
})

const digitalHumanCanGenerate = computed(() => {
  if (!selectedDigitalHumanImageFile.value) return false
  if (!digitalHumanForm.speechText.trim()) return false
  return !!selectedVoice.value || !!digitalHumanForm.audioFile
})

const toPlayableMediaUrl = (url: string): string => {
  const value = (url || '').trim()
  if (!value) return ''
  if (value.startsWith('blob:') || value.startsWith('data:')) return value
  if (value.startsWith('/api/v1/media/proxy')) return value
  if (value.startsWith('http://') || value.startsWith('https://')) {
    return `/api/v1/media/proxy?url=${encodeURIComponent(value)}`
  }
  return value
}

const digitalHumanPlayableResultUrl = computed(() => {
  return toPlayableMediaUrl(digitalHumanResultUrl.value)
})

// 分镜相关状态
const generatingShots = ref(false)
const parsingCharacters = ref(false)

const isCharacterIndeterminate = computed(() => {
  const selectedCount = selectedCharacterIds.value.length
  const totalCount = drama.value?.characters?.length || 0
  return selectedCount > 0 && selectedCount < totalCount
})

const getVoiceLanguage = (voiceType?: string) => {
  const vt = (voiceType || '').toLowerCase()
  if (vt.startsWith('zh_') || vt.includes('_zh_')) return '中文'
  if (vt.startsWith('en_') || vt.includes('_en_')) return '英语'
  if (vt.startsWith('jp_') || vt.includes('_jp_')) return '日语'
  if (vt.startsWith('kr_') || vt.includes('_kr_')) return '韩语'
  return '其他'
}

const voiceGenderOptions = computed(() => {
  const set = new Set<string>()
  for (const item of voiceLibraryList.value) {
    if (item.gender) set.add(item.gender)
  }
  return Array.from(set)
})

const voiceAgeOptions = computed(() => {
  const set = new Set<string>()
  for (const item of voiceLibraryList.value) {
    if (item.age) set.add(item.age)
  }
  return Array.from(set)
})

const voiceLanguageOptions = computed(() => {
  const set = new Set<string>()
  for (const item of voiceLibraryList.value) {
    set.add(getVoiceLanguage(item.voice_type))
  }
  const ordered = ['中文', '英语', '日语', '韩语']
  const list = Array.from(set)
  return [
    ...ordered.filter(v => list.includes(v)),
    ...list.filter(v => !ordered.includes(v) && v !== '其他'),
    ...((list.includes('其他') ? ['其他'] : []) as string[])
  ]
})

const voiceCategoryOptions = computed(() => {
  const set = new Set<string>()
  for (const item of voiceLibraryList.value) {
    for (const cat of item.categories || []) set.add(cat)
  }
  return Array.from(set)
})

const filteredVoiceLibrary = computed(() => {
  const keyword = voiceLibrarySearch.value.trim().toLowerCase()
  return voiceLibraryList.value.filter(item => {
    if (keyword) {
      const matched =
        item.name?.toLowerCase().includes(keyword) ||
        item.voice_type?.toLowerCase().includes(keyword) ||
        item.gender?.toLowerCase().includes(keyword) ||
        item.age?.toLowerCase().includes(keyword)
      if (!matched) return false
    }

    if (voiceGenderFilter.value && item.gender !== voiceGenderFilter.value) return false
    if (voiceAgeFilter.value && item.age !== voiceAgeFilter.value) return false
    if (voiceLanguageFilter.value && getVoiceLanguage(item.voice_type) !== voiceLanguageFilter.value) return false
    if (voiceCategoryFilter.value && !(item.categories || []).includes(voiceCategoryFilter.value)) return false

    return true
  })
})

watch(digitalHumanDialogVisible, (visible) => {
  if (typeof document === 'undefined') return
  if (visible) {
    document.body.classList.add('digital-human-dialog-open')
    document.documentElement.classList.add('digital-human-dialog-open')
  } else {
    document.body.classList.remove('digital-human-dialog-open')
    document.documentElement.classList.remove('digital-human-dialog-open')
  }
})

watch(voiceLibraryVisible, (visible) => {
  if (!visible) {
    stopVoiceTrial()
  }
})

onBeforeUnmount(() => {
  if (typeof document === 'undefined') return
  document.body.classList.remove('digital-human-dialog-open')
  document.documentElement.classList.remove('digital-human-dialog-open')
})

const isCharacterSelected = (id: number | string) => {
  return selectedCharacterIds.value.includes(id)
}

const toggleCharacterSelection = (id: number | string) => {
  const index = selectedCharacterIds.value.indexOf(id)
  if (index > -1) {
    selectedCharacterIds.value.splice(index, 1)
  } else {
    selectedCharacterIds.value.push(id)
  }
  updateSelectAllCharactersState()
}

const handleSelectAllCharacters = (val: boolean) => {
  if (val && drama.value?.characters) {
    selectedCharacterIds.value = drama.value.characters.map(c => c.id)
  } else {
    selectedCharacterIds.value = []
  }
}

const updateSelectAllCharactersState = () => {
  const totalCount = drama.value?.characters?.length || 0
  selectAllCharacters.value = selectedCharacterIds.value.length === totalCount && totalCount > 0
}
const libraryDialogVisible = ref(false)
const selectedCharacter = ref<any>(null)
const characterLibrary = ref<any[]>([])
const editDescDialogVisible = ref(false)
const editingCharacter = ref<any>(null)
const saving = ref(false)
const addCharacterDialogVisible = ref(false)
const editDescFormRef = ref<{ $el?: HTMLElement } | null>(null)
const addCharacterFormRef = ref<{ $el?: HTMLElement } | null>(null)
const newCharacter = ref({
  name: '',
  role: 'supporting',
  appearance: '',
  personality: '',
  description: ''
})

// 各阶段完成状态
// 判断当前集是否已有剧本
const hasScript = computed(() => {
  if (!drama.value?.episodes || drama.value.episodes.length === 0) {
    return false
  }
  // 检查当前集是否存在
  const currentEpisode = drama.value.episodes.find(
    ep => ep.episode_number === currentEpisodeNumber.value
  )
  return currentEpisode && currentEpisode.script_content && currentEpisode.script_content.length > 0
})

// 获取当前集
const currentEpisode = computed(() => {
  if (!drama.value?.episodes) return null
  return drama.value.episodes.find(
    ep => ep.episode_number === currentEpisodeNumber.value
  )
})

// 判断是否有角色
const hasCharacters = computed(() => {
  return drama.value?.characters && drama.value.characters.length > 0
})
const episodesCount = computed(() => drama.value?.episodes?.length || 0)
const sortedEpisodes = computed(() => {
  if (!drama.value?.episodes) return []
  return [...drama.value.episodes].sort((a, b) => a.episode_number - b.episode_number)
})
const charactersCount = computed(() => drama.value?.characters?.length || 0)
const characterImagesCount = computed(() => {
  return drama.value?.characters?.filter(c => c.image_url).length || 0
})
const allCharactersHaveImages = computed(() => {
  if (!drama.value?.characters || drama.value.characters.length === 0) {
    return false
  }
  return drama.value.characters.every(c => c.image_url && c.image_url.length > 0)
})
const completedEpisodesCount = computed(() => {
  return 0
})
const overallProgress = computed(() => {
  return 0
})

// 修复图片URL协议问题
const fixImageUrl = (url: string | undefined | null): string => {
  if (!url) return ''
  // 如果是blob URL，直接返回
  if (url.startsWith('blob:')) return url
  return url
}

// 状态标签
const getStatusType = (status?: DramaStatus) => {
  const types: Partial<Record<DramaStatus, string>> = {
    draft: 'info',
    planning: 'primary',
    production: 'warning',
    generating: 'warning',
    completed: 'success',
    archived: 'info',
    error: 'danger'
  }
  return status ? types[status] : 'info'
}

const getStatusText = (status?: DramaStatus) => {
  const texts: Partial<Record<DramaStatus, string>> = {
    draft: '草稿',
    planning: '策划中',
    production: '制作中',
    generating: '生成中',
    completed: '已完成',
    archived: '已归档',
    error: '错误'
  }
  return status ? texts[status] : '未知'
}

// 导航
const goBack = () => {
  router.push('/dramas')
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
    updateUrlState()
  }
}

const nextStep = () => {
  if (currentStep.value < 2) {
    currentStep.value++
    updateUrlState()
  }
}

// 更新URL状态，保存当前步骤
const updateUrlState = () => {
  router.replace({
    query: {
      ...route.query,
      step: currentStep.value.toString()
    }
  })
}

// 页面跳转
const goToScriptGeneration = () => {
  router.push(`/dramas/${drama.value?.id}/script`)
}

// AI流式生成剧本
const generateScriptByAI = async () => {
  if (!drama.value?.title) {
    ElMessage.warning('项目标题不存在')
    return
  }

  generatingScript.value = true
  scriptContent.value = ''
  
  try {
    const response = await fetch('/api/v1/ai/generate-script-stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        drama_title: drama.value.title,
        drama_id: drama.value.id
      })
    })

    if (!response.ok) {
      throw new Error('生成失败')
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    if (!reader) {
      throw new Error('无法读取响应流')
    }

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      
      const chunk = decoder.decode(value, { stream: true })
      scriptContent.value += chunk
    }

    ElMessage.success('剧本生成完成')
  } catch (error: any) {
    ElMessage.error(error.message || '生成失败')
    scriptContent.value = ''
  } finally {
    generatingScript.value = false
  }
}

// 保存章节剧本（不解析角色）
const saveChapterScript = async () => {
  if (!scriptContent.value.trim()) {
    ElMessage.warning('请输入章节内容')
    return
  }

  generatingScript.value = true
  try {
    ElMessage.info('正在保存章节...')
    
    // 保存当前章节内容，不进行角色解析
    const existingEpisodes = drama.value?.episodes || []
    const episodeIndex = existingEpisodes.findIndex(
      ep => ep.episode_number === currentEpisodeNumber.value
    )
    
    const currentEpisodeData = {
      episode_number: currentEpisodeNumber.value,
      title: `第${currentEpisodeNumber.value}章`,
      script_content: scriptContent.value,
      description: '',
      duration: 0,
      status: 'draft'
    }
    
    let episodesToSave
    if (episodeIndex > -1) {
      // 更新现有章节
      episodesToSave = [...existingEpisodes]
      episodesToSave[episodeIndex] = {
        ...existingEpisodes[episodeIndex],
        ...currentEpisodeData
      }
    } else {
      // 添加新章节
      episodesToSave = [
        ...existingEpisodes.map(ep => ({
          episode_number: ep.episode_number,
          title: ep.title,
          script_content: ep.script_content,
          description: ep.description,
          duration: ep.duration,
          status: ep.status
        })),
        currentEpisodeData
      ]
    }
    
    await dramaAPI.saveEpisodes(drama.value!.id, episodesToSave)
    
    ElMessage.success(`第${currentEpisodeNumber.value}章保存成功`)
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    generatingScript.value = false
  }
}

const openDigitalHumanDialog = () => {
  digitalHumanDialogVisible.value = true
}

const loadVoiceLibrary = async () => {
  voiceLibraryLoading.value = true
  voiceLibraryError.value = ''
  try {
    const data = await voiceLibraryAPI.list()
    voiceLibraryList.value = data || []
  } catch (error: any) {
    voiceLibraryError.value = error?.message || '获取音色库失败'
    ElMessage.error(voiceLibraryError.value)
  } finally {
    voiceLibraryLoading.value = false
  }
}

const openVoiceLibrary = async () => {
  voiceLibraryVisible.value = true
  if (!voiceLibraryList.value.length) {
    await loadVoiceLibrary()
  }
}

const handleVoicePopoverShow = async () => {
  voiceLibrarySearch.value = ''
  voiceGenderFilter.value = null
  voiceAgeFilter.value = null
  voiceLanguageFilter.value = null
  voiceCategoryFilter.value = null
  if (!voiceLibraryList.value.length) {
    await loadVoiceLibrary()
  }
}

const selectVoice = (voice: VoiceLibraryItem) => {
  selectedVoice.value = voice
  stopVoiceTrial()
  voiceLibraryVisible.value = false
}

const openCreateVoice = () => {
  digitalHumanUploadVoiceInputRef.value?.click()
}

const stopVoiceTrial = () => {
  voiceTrialPlayingId.value = ''
  voiceTrialUrl.value = ''
  const audioEl = voiceTrialAudioRef.value
  if (audioEl) {
    audioEl.pause()
    audioEl.currentTime = 0
  }
}

const toggleVoiceTrial = async (voice: VoiceLibraryItem) => {
  if (!voice.trial_url) {
    ElMessage.warning('该音色暂无试听')
    return
  }

  const audioEl = voiceTrialAudioRef.value
  if (!audioEl) {
    window.open(voice.trial_url, '_blank')
    return
  }

  try {
    if (voiceTrialPlayingId.value === voice.id && !audioEl.paused) {
      audioEl.pause()
      voiceTrialPlayingId.value = ''
      return
    }

    // Stop any previous trial before switching.
    audioEl.pause()

    voiceTrialPlayingId.value = voice.id
    voiceTrialUrl.value = voice.trial_url
    await nextTick()
    audioEl.currentTime = 0
    await audioEl.play()
  } catch (error) {
    stopVoiceTrial()
    window.open(voice.trial_url, '_blank')
  }
}

const clearSelectedVoice = () => {
  selectedVoice.value = null
}

const handleUploadVoiceFromCard = async (event: Event) => {
  const inputEl = event.target as HTMLInputElement | null
  const file = inputEl?.files?.[0]
  if (!file) {
    return
  }

  const allowed = await Promise.resolve(beforeDigitalHumanAudioUpload?.(file as any))
  if (allowed !== false) {
    const rawUID = Date.now()
    digitalHumanAudioList.value = [
      {
        name: file.name,
        size: file.size,
        status: 'success',
        uid: rawUID,
        raw: Object.assign(file, { uid: rawUID })
      } as UploadUserFile
    ]
    digitalHumanForm.audioFile = file
    digitalHumanAudioPreviewVisible.value = false
    if (digitalHumanAudioPreview.value && digitalHumanAudioPreview.value.startsWith('blob:')) {
      URL.revokeObjectURL(digitalHumanAudioPreview.value)
    }
    digitalHumanAudioPreview.value = URL.createObjectURL(file)
    selectedVoice.value = null
  }

  if (inputEl) {
    inputEl.value = ''
  }
}

const downloadDigitalHumanVideo = async () => {
  const playableURL = digitalHumanPlayableResultUrl.value
  if (!playableURL) {
    ElMessage.warning('暂无可下载视频')
    return
  }

  try {
    const loadingMsg = ElMessage.info({
      message: '正在准备下载...',
      duration: 0
    })
    const response = await fetch(playableURL)
    if (!response.ok) {
      throw new Error(`下载失败: ${response.status}`)
    }
    const blob = await response.blob()
    const blobURL = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = blobURL
    link.download = `digital_human_${Date.now()}.mp4`
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    setTimeout(() => {
      document.body.removeChild(link)
      URL.revokeObjectURL(blobURL)
    }, 100)
    loadingMsg.close()
    ElMessage.success('视频下载已开始')
  } catch (error: any) {
    ElMessage.error(error?.message || '视频下载失败，请稍后重试')
  }
}

const resetDigitalHumanForm = () => {
  digitalHumanForm.imageFile = null
  digitalHumanForm.audioFile = null
  digitalHumanForm.speechText = ''
  digitalHumanForm.motionText = ''
  digitalHumanImageList.value = []
  digitalHumanAudioList.value = []
  digitalHumanResultUrl.value = ''
  digitalHumanImagePreviewVisible.value = false
  digitalHumanAudioPreviewVisible.value = false
  selectedVoice.value = null
  if (digitalHumanImagePreview.value && digitalHumanImagePreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(digitalHumanImagePreview.value)
    digitalHumanImagePreview.value = ''
  }
  if (digitalHumanAudioPreview.value && digitalHumanAudioPreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(digitalHumanAudioPreview.value)
    digitalHumanAudioPreview.value = ''
  }
}

const beforeDigitalHumanImageUpload: UploadProps['beforeUpload'] = (file) => {
  if (!file.type.startsWith('image/')) {
    ElMessage.error('请上传图片文件')
    return false
  }
  if (file.size > 10 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过10MB')
    return false
  }
  return true
}

const beforeDigitalHumanAudioUpload: UploadProps['beforeUpload'] = (file) => {
  if (!(file.type.startsWith('audio/') || file.type === 'video/mp4')) {
    ElMessage.error('请上传音频文件')
    return false
  }
  if (file.size > 20 * 1024 * 1024) {
    ElMessage.error('音频大小不能超过20MB')
    return false
  }
  return true
}

const handleDigitalHumanImageChange: UploadProps['onChange'] = (file, fileList) => {
  digitalHumanImageList.value = fileList.slice(-1)
  const rawFile = getUploadRawFile(file) || getUploadRawFile(digitalHumanImageList.value?.[0])
  digitalHumanForm.imageFile = rawFile
  digitalHumanImagePreviewVisible.value = false
  if (digitalHumanImagePreview.value && digitalHumanImagePreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(digitalHumanImagePreview.value)
  }
  if (rawFile) {
    digitalHumanImagePreview.value = URL.createObjectURL(rawFile)
  } else if (file.url) {
    digitalHumanImagePreview.value = file.url
  } else {
    digitalHumanImagePreview.value = ''
  }
}

const handleDigitalHumanAudioChange: UploadProps['onChange'] = (file, fileList) => {
  digitalHumanAudioList.value = fileList.slice(-1)
  const rawFile = (file.raw ?? (file as any).originFileObj) as File | undefined
  digitalHumanForm.audioFile = rawFile || null
  digitalHumanAudioPreviewVisible.value = false
  if (digitalHumanAudioPreview.value && digitalHumanAudioPreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(digitalHumanAudioPreview.value)
  }
  if (rawFile) {
    digitalHumanAudioPreview.value = URL.createObjectURL(rawFile)
  } else if (file.url) {
    digitalHumanAudioPreview.value = file.url
  } else {
    digitalHumanAudioPreview.value = ''
  }
}

const handleDigitalHumanImageRemove: UploadProps['onRemove'] = () => {
  digitalHumanForm.imageFile = null
  digitalHumanImageList.value = []
  digitalHumanImagePreviewVisible.value = false
  if (digitalHumanImagePreview.value && digitalHumanImagePreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(digitalHumanImagePreview.value)
    digitalHumanImagePreview.value = ''
  }
}

const handleDigitalHumanAudioRemove: UploadProps['onRemove'] = () => {
  digitalHumanForm.audioFile = null
  digitalHumanAudioList.value = []
  digitalHumanAudioPreviewVisible.value = false
  if (digitalHumanAudioPreview.value && digitalHumanAudioPreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(digitalHumanAudioPreview.value)
    digitalHumanAudioPreview.value = ''
  }
}

const openDigitalHumanImagePreview = () => {
  if (!digitalHumanImagePreview.value) {
    return
  }
  digitalHumanImagePreviewVisible.value = true
}

const openDigitalHumanAudioPreview = async () => {
  if (!digitalHumanAudioPreview.value) {
    return
  }
  digitalHumanAudioPreviewVisible.value = true
  await nextTick()
  const audioEl = digitalHumanAudioRef.value
  if (!audioEl) {
    return
  }
  try {
    audioEl.currentTime = 0
    await audioEl.play()
  } catch (error) {
    ElMessage.warning('浏览器阻止了自动播放，请手动点击播放')
  }
}

const submitDigitalHuman = async () => {
  const speechText = digitalHumanForm.speechText.trim()
  const motionText = digitalHumanForm.motionText.trim()
  const hasAudio = !!digitalHumanForm.audioFile
  const hasSelectedVoice = !!selectedVoice.value
  const imageFile = selectedDigitalHumanImageFile.value
  if (!imageFile) {
    ElMessage.warning('请先上传角色图片')
    return
  }
  if (!speechText) {
    ElMessage.warning('请先填写说话内容')
    return
  }
  if (!hasAudio && !hasSelectedVoice) {
    ElMessage.warning('请先选择音色或上传音色音频')
    return
  }

  digitalHumanLoading.value = true
  digitalHumanResultUrl.value = ''
  try {
    ElMessage.info('数字人视频生成中，请稍候...')
    const formData = new FormData()
    formData.append('image', imageFile)
    if (digitalHumanForm.audioFile) {
      formData.append('audio', digitalHumanForm.audioFile)
    }
    if (speechText) formData.append('speech_text', speechText)
    if (motionText) formData.append('motion_text', motionText)
    if (selectedVoice.value) {
      formData.append('voice_id', selectedVoice.value.id)
      if (selectedVoice.value.voice_type) {
        formData.append('voice_type', selectedVoice.value.voice_type)
      }
    }

    const result = await digitalHumanAPI.generate(formData)
    if (!result.video_url) {
      throw new Error('未获取到视频链接')
    }
    digitalHumanResultUrl.value = result.video_url
    ElMessage.success('数字人视频生成完成')
  } catch (error: any) {
    const raw = String(error?.message || '')
    if (raw.includes('DIGITAL_HUMAN_TTS_NOT_ENABLED') || raw.includes('未开通文本配音能力')) {
      ElMessage.error('当前账号未开通文本配音能力：请上传音色音频后生成，或联系管理员开通文本驱动配音')
    } else {
      ElMessage.error(raw || '生成失败')
    }
  } finally {
    digitalHumanLoading.value = false
  }
}

// 编辑角色描述
const editCharacterDescription = (character: any) => {
  editingCharacter.value = { ...character }
  editDescDialogVisible.value = true
}

// 保存角色描述
const saveCharacterDescription = async () => {
  if (!editingCharacter.value) return
  
  saving.value = true
  try {
    await characterLibraryAPI.updateCharacter(editingCharacter.value.id, {
      appearance: editingCharacter.value.appearance,
      personality: editingCharacter.value.personality,
      description: editingCharacter.value.description
    })
    
    ElMessage.success('角色描述已更新')
    editDescDialogVisible.value = false
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

// 集数切换
const switchEpisode = (episodeNumber: number) => {
  currentEpisodeNumber.value = episodeNumber
  // 加载该集的剧本内容
  const episode = drama.value?.episodes?.find(ep => ep.episode_number === episodeNumber)
  if (episode && episode.script_content) {
    scriptContent.value = episode.script_content
  } else {
    scriptContent.value = ''
  }
}

// 开始创建章节
const startCreateChapter = () => {
  showScriptInput.value = true
}

// 创建下一集
const createNextEpisode = () => {
  currentEpisodeNumber.value = episodesCount.value + 1
  scriptContent.value = ''
  showScriptInput.value = true // 显示输入框
  currentStep.value = 0
}

// 编辑当前集剧本
const editCurrentEpisodeScript = () => {
  if (currentEpisode.value?.script_content) {
    scriptContent.value = currentEpisode.value.script_content
  }
}

// AI自动拆分分镜
const generateShots = async () => {
  if (!currentEpisode.value?.script_content) {
    ElMessage.warning(t('dramaWorkflow.pleaseWriteScript'))
    return
  }
  
  generatingShots.value = true
  try {
    ElMessage.info('AI正在拆分镜头...')
    
    // 调用分镜拆分API
    const result = await generationAPI.generateStoryboard(currentEpisode.value.id.toString())
    ElMessage.success(result.message || '分镜拆分任务已提交')
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '拆分失败')
  } finally {
    generatingShots.value = false
  }
}

// 重新拆分分镜
const regenerateShots = async () => {
  await ElMessageBox.confirm(
    t('dramaWorkflow.reGenerateShotsConfirm'),
    t('dramaWorkflow.reGenerateShots'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  )
  await generateShots()
}

// 编辑镜头
const editShot = (shot: any, index: number) => {
  // TODO: 打开镜头编辑对话框
  ElMessage.info('镜头编辑功能开发中')
}

// 从分镜解析角色
const parseShotsToCharacters = async () => {
  if (!currentEpisode.value?.shots || currentEpisode.value.shots.length === 0) {
    ElMessage.warning('请先进行分镜拆分')
    return
  }
  
  parsingCharacters.value = true
  try {
    ElMessage.info('正在解析角色...')
    
    // 从所有镜头内容中提取角色
    const shotsContent = currentEpisode.value.shots.map((s: any) => s.content).join('\n')
    
    const parseResult = await (generationAPI as any).parseScript({
      drama_id: drama.value!.id,
      script_content: shotsContent,
      auto_split: false
    })
    
    if (parseResult.characters && parseResult.characters.length > 0) {
      const existingCharacters = drama.value?.characters || []
      const existingNames = new Set(existingCharacters.map(c => c.name))
      
      // 只添加新角色
      const newCharacters = parseResult.characters.filter(
        (c: any) => !existingNames.has(c.name)
      )
      
      if (newCharacters.length > 0) {
        const allCharacters = [
          ...existingCharacters.map(c => ({
            name: c.name,
            role: c.role,
            appearance: c.appearance,
            personality: c.personality,
            description: c.description
          })),
          ...newCharacters
        ]
        await dramaAPI.saveCharacters(drama.value!.id, allCharacters)
        ElMessage.success(`成功解析 ${newCharacters.length} 个新角色`)
      } else {
        ElMessage.info('未发现新角色')
      }
    } else {
      ElMessage.warning('未解析到角色信息')
    }
    
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '解析失败')
  } finally {
    parsingCharacters.value = false
  }
}

// 打开添加角色对话框
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

// 添加角色
const addCharacter = async () => {
  if (!newCharacter.value.name.trim()) {
    ElMessage.warning('请输入角色名称')
    return
  }
  
  if (!newCharacter.value.description.trim()) {
    ElMessage.warning('请输入角色描述')
    return
  }
  
  saving.value = true
  try {
    // 将新角色添加到现有角色列表中，而不是覆盖
    const existingCharacters = drama.value?.characters || []
    const allCharacters = [
      ...existingCharacters.map(c => ({
        name: c.name,
        role: c.role,
        appearance: c.appearance,
        personality: c.personality,
        description: c.description
      })),
      {
        name: newCharacter.value.name,
        role: newCharacter.value.role,
        appearance: newCharacter.value.appearance,
        personality: newCharacter.value.personality,
        description: newCharacter.value.description
      }
    ]
    
    await dramaAPI.saveCharacters(drama.value!.id, allCharacters)
    
    ElMessage.success('角色添加成功')
    addCharacterDialogVisible.value = false
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '添加失败')
  } finally {
    saving.value = false
  }
}

// 删除角色
const deleteCharacter = async (character: any) => {
  try {
    // 检查角色是否在角色库中
    if (character.library_id) {
      ElMessage.warning('该角色来自角色库，请到角色库中删除')
      return
    }
    
    await ElMessageBox.confirm(
      `确定要删除角色 "${character.name}" 吗？此操作不可恢复。`,
      '删除角色',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    saving.value = true
    // 从现有角色列表中移除该角色，然后保存
    const remainingCharacters = drama.value!.characters!
      .filter(c => c.id !== character.id)
      .map(c => ({
        name: c.name,
        role: c.role,
        appearance: c.appearance,
        personality: c.personality,
        description: c.description
      }))
    
    await dramaAPI.saveCharacters(drama.value!.id, remainingCharacters)
    
    ElMessage.success('角色已删除')
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  } finally {
    saving.value = false
  }
}

const generateCharacterImage = async (character: any) => {
  if (generatingCharacterIds.value.includes(character.id)) return
  
  generatingCharacterIds.value.push(character.id)
  try {
    const res = await characterLibraryAPI.generateCharacterImage(character.id)
    ElMessage.success(`${character.name}的形象生成成功`)
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || `${character.name}生成失败`)
  } finally {
    const index = generatingCharacterIds.value.indexOf(character.id)
    if (index > -1) {
      generatingCharacterIds.value.splice(index, 1)
    }
  }
}

const batchGenerateCharacterImages = async () => {
  if (selectedCharacterIds.value.length === 0) {
    ElMessage.warning('请选择要生成的角色')
    return
  }

  if (selectedCharacterIds.value.length > 10) {
    ElMessage.warning('单次最多生成10个角色')
    return
  }

  batchGenerating.value = true
  generatingCharacterIds.value = [...selectedCharacterIds.value]
  
  try {
    await characterLibraryAPI.batchGenerateCharacterImages(
      selectedCharacterIds.value.map(id => String(id))
    )
    
    ElMessage.success(`批量生成任务已提交，正在后台生成 ${selectedCharacterIds.value.length} 个角色形象`)
    
    // 轮询检查生成状态
    startCharacterPolling()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '批量生成失败')
    batchGenerating.value = false
    generatingCharacterIds.value = []
  }
}

let characterPollingTimer: number | null = null

const startCharacterPolling = () => {
  if (characterPollingTimer) return
  
  characterPollingTimer = window.setInterval(async () => {
    try {
      await loadDramaData()
      
      if (!drama.value?.characters) return
      
      // 检查每个选中角色的状态
      let completedCount = 0
      let failedCount = 0
      const failedCharacters: string[] = []
      
      selectedCharacterIds.value.forEach(id => {
        const char = drama.value?.characters?.find(c => c.id === id)
        if (char) {
          if (char.image_url) {
            completedCount++
          } else if (char.image_generation_status === 'failed') {
            failedCount++
            failedCharacters.push(char.name)
          }
        }
      })
      
      // 如果所有任务都完成（成功或失败），停止轮询
      if (completedCount + failedCount === selectedCharacterIds.value.length) {
        stopCharacterPolling()
        
        if (failedCount > 0) {
          ElMessage.warning(`批量生成完成：${completedCount}个成功，${failedCount}个失败（${failedCharacters.join('、')}）`)
        } else {
          ElMessage.success('批量生成完成')
        }
      }
    } catch (error) {
      console.error('轮询错误:', error)
    }
  }, 5000) // 每5秒检查一次
}

const stopCharacterPolling = () => {
  if (characterPollingTimer) {
    clearInterval(characterPollingTimer)
    characterPollingTimer = null
  }
  batchGenerating.value = false
  generatingCharacterIds.value = []
  selectedCharacterIds.value = []
  selectAllCharacters.value = false
}

const openUploadDialog = (character: any) => {
  selectedCharacter.value = character
  
  // 创建临时文件输入框
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/jpeg,image/png,image/jpg'
  
  input.onchange = async (e: any) => {
    const file = e.target.files?.[0]
    if (!file) return
    
    // 验证文件大小（10MB）
    if (file.size > 10 * 1024 * 1024) {
      ElMessage.error('图片大小不能超过10MB')
      return
    }
    
    try {
      // 创建FormData上传文件
      const formData = new FormData()
      formData.append('file', file)
      
      ElMessage.info('正在上传图片...')
      
      // 上传到后端MinIO（后端会自动更新数据库）
      await request.post<{ url: string }>(`/characters/${selectedCharacter.value.id}/upload-image`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
      
      ElMessage.success('图片上传成功')
      await loadDramaData()
    } catch (error: any) {
      ElMessage.error(error.message || '上传失败')
    }
  }
  
  // 触发文件选择
  input.click()
}

const openCharacterLibrary = async (character: any) => {
  selectedCharacter.value = character
  try {
    const res = await characterLibraryAPI.list({ page: 1, page_size: 100 })
    characterLibrary.value = res.items || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载角色库失败')
    characterLibrary.value = []
  }
  libraryDialogVisible.value = true
}

const selectFromLibrary = async (libraryItem: any) => {
  try {
    await characterLibraryAPI.applyFromLibrary(selectedCharacter.value.id, libraryItem.id)
    ElMessage.success('已应用角色库形象')
    libraryDialogVisible.value = false
    await loadDramaData()
  } catch (error: any) {
    ElMessage.error(error.message || '应用失败')
  }
}

const addToCharacterLibrary = async (character: any) => {
  try {
    await characterLibraryAPI.addCharacterToLibrary(character.id)
    ElMessage.success(`${character.name}已添加到角色库`)
  } catch (error: any) {
    ElMessage.error(error.message || '添加失败')
  }
}

const goToEpisodeList = () => {
  router.push(`/dramas/${drama.value?.id}/episodes`)
}

const goToEpisodeDetail = (episodeId: string) => {
  router.push(`/dramas/${drama.value?.id}/episodes/${episodeId}`)
}

const loadDramaData = async () => {
  const dramaId = route.params.id as string
  try {
    drama.value = await dramaAPI.get(dramaId)
  } catch (error: any) {
    ElMessage.error(error.message || '获取剧本信息失败')
    router.push('/dramas')
  }
}

onMounted(() => {
  loadDramaData()
})
</script>

<style scoped>
.workflow-container {
  min-height: var(--app-vh, 100vh);
  background: #f8fafc;
  transition: background var(--transition-normal);
}

.dark .workflow-container {
  background: #0f172a;
}

.workflow-header {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-primary);
  padding: 10px 24px;
  margin-bottom: 0;
}

.header-single-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 32px;
}

.header-left-section {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.back-btn {
  color: var(--text-secondary);
  padding: 0;
  margin-right: 4px;
}

.back-btn:hover {
  color: #0ea5e9;
}

.drama-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
}

.steps-inline {
  flex: 1;
  display: flex;
  justify-content: center;
  min-width: 0;
}

.custom-steps {
  display: flex;
  align-items: center;
  gap: 16px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.step-circle {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  background: #e4e7ed;
  color: #909399;
  transition: all 0.3s;
}

.step-item.active .step-circle {
  background: #409eff;
  color: #ffffff;
}

.step-item.current .step-circle {
  background: #409eff;
  color: #ffffff;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.2);
}

.step-text {
  font-size: 13px;
  font-weight: 500;
  color: #909399;
  transition: color 0.3s;
}

.step-item.active .step-text {
  color: #303133;
}

.step-item.current .step-text {
  color: #409eff;
  font-weight: 600;
}

.step-arrow {
  font-size: 16px;
  color: #c0c4cc;
}

.stage-area {
  padding: 0;
}

.stage-card {
  min-height: 400px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.stage-card-fullscreen {
  min-height: calc(100vh - 70px);
  display: flex;
  flex-direction: column;
  margin: 0;
  border: none;
  border-radius: 0;
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-info h2 {
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
}

.header-info p {
  margin: 0;
  font-size: 13px;
  color: #909399;
}

.stage-body {
  padding: 20px 0;
}

.stats-row {
  display: flex;
  gap: 24px;
  justify-content: center;
  margin: 24px 0;
}

.stat-box {
  text-align: center;
  min-width: 140px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  transition: all 0.3s ease;
}

.stat-box:hover {
  background: #ecf5ff;
  border-color: #c6e2ff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.1);
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
  font-weight: 500;
}

.stat-value {
  font-size: 32px;
  font-weight: 600;
  color: #409eff;
}

.stage-body-fullscreen {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.action-buttons-inline {
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.digital-human-dialog {
  :global(body.digital-human-dialog-open),
  :global(html.digital-human-dialog-open) {
    overflow: hidden !important;
    padding-right: 0 !important;
    height: 100% !important;
    overscroll-behavior: none;
  }

  :global(body.digital-human-dialog-open #app) {
    height: 100% !important;
    overflow: hidden !important;
  }

  :global(body.digital-human-dialog-open .el-overlay),
  :global(body.digital-human-dialog-open .el-overlay-dialog),
  :global(body.digital-human-dialog-open .el-dialog__wrapper) {
    overflow: hidden !important;
  }

  :deep(.el-dialog) {
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  :deep(.el-dialog__body) {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .digital-human-form {
    :deep(.el-form-item__label) {
      font-weight: 600;
    }
  }

  .digital-human-upload {
    width: 100%;
  }

  .digital-human-audio-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .digital-human-audio-row .digital-human-upload {
    width: auto;
  }

  .digital-human-audio-row :deep(.el-upload) {
    display: inline-flex;
  }

  .digital-human-upload-btn {
    width: auto;
    height: 28px;
    padding: 0 8px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.1px;
    box-shadow: 0 3px 8px rgba(64, 158, 255, 0.16);
  }

  .digital-human-role-btn {
    max-width: 240px;
  }

  .digital-human-role-btn-label {
    display: inline-block;
    max-width: 170px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    vertical-align: bottom;
  }

  .digital-human-role-btn-clear {
    margin-left: 6px;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.92);
  }

  .digital-human-voice-btn {
    max-width: 240px;
  }

  .digital-human-voice-btn-label {
    display: inline-block;
    max-width: 160px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    vertical-align: bottom;
  }

  .digital-human-voice-btn-clear {
    margin-left: 6px;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.92);
  }

  .digital-human-upload-secondary {
    box-shadow: none;
  }

  .digital-human-file-name {
    margin-top: 8px;
    font-size: 8px;
    color: #606266;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 0 6px;
    border-radius: 6px;
    border: 1px solid #dcdfe6;
    background: #f5f7fa;
    transition: all 0.2s ease;
    max-width: 180px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .digital-human-clear-icon {
    margin-left: 6px;
    font-size: 12px;
    color: #c0c4cc;
  }

  .digital-human-file-name:hover {
    color: #409eff;
    border-color: #409eff;
    background: #ecf5ff;
  }

  .digital-human-preview {
    margin-top: 10px;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid #e4e7ed;
    background: #f5f7fa;

    :deep(.el-image) {
      width: 100%;
      height: 200px;
    }
  }

  .digital-human-audio {
    margin-top: 10px;
    max-width: 100%;
    audio {
      width: 100%;
      max-width: 100%;
      display: block;
    }
  }

  .digital-human-hint {
    margin-top: 6px;
    font-size: 12px;
    color: #909399;
  }

  .digital-human-hint-inline {
    font-size: 11px;
    color: #909399;
  }

  .digital-human-result {
    margin-top: 12px;
    padding: 12px;
    border-radius: 8px;
    background: #f5f7fa;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .digital-human-result-title {
    font-size: 13px;
    font-weight: 600;
    color: #303133;
  }

  .digital-human-result video {
    width: 100%;
    border-radius: 8px;
    background: #000;
  }

  .digital-human-textarea :deep(textarea) {
    resize: none;
    overflow: auto;
  }
}

.voice-library-popover {
  padding: 12px;
}

.digital-human-required-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.digital-human-required-dot {
  color: #f56c6c;
  font-weight: 700;
  line-height: 1;
}

.digital-human-hidden-input {
  display: none;
}

.voice-library-panel {
  width: 100%;
}

.voice-library-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.voice-library-toolbar :deep(.el-input) {
  flex: 1;
}

.voice-library-filters {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.voice-filter {
  width: 148px;
  max-width: 100%;
}

.voice-library-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #909399;
  padding: 12px 0;
}

.voice-library-error {
  margin-top: 8px;
  font-size: 12px;
  color: #f56c6c;
}

.voice-library-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}

.voice-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid #e4e7ed;
  background: #fafafa;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.voice-card.is-trial-playing {
  border-color: #409eff;
  background: #eaf4ff;
  box-shadow: 0 6px 14px rgba(64, 158, 255, 0.18);
}

.voice-card.is-selected {
  border-color: rgba(64, 158, 255, 0.65);
  background: #f1f8ff;
}

.voice-card.is-selected.is-trial-playing {
  border-color: #409eff;
  background: #eaf4ff;
  box-shadow: 0 6px 14px rgba(64, 158, 255, 0.18);
}

.voice-card:hover {
  border-color: #c6e2ff;
  background: #ecf5ff;
}

.voice-card-create {
  border-style: dashed;
  color: #409eff;
  justify-content: center;
}

.voice-card-play {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #eef5ff;
  color: #409eff;
  font-size: 14px;
  flex-shrink: 0;
}

.voice-card-play.is-playing {
  background: #409eff;
  color: #ffffff;
}

.voice-card-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.voice-card-name {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.voice-card-meta {
  font-size: 11px;
  color: #909399;
  display: flex;
  gap: 6px;
}

.voice-trial-audio {
  display: none;
}

.action-area {
  text-align: center;
  padding: 20px 0;
  flex-shrink: 0;
}

.action-area h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
}

.main-action-btn {
  width: 100%;
  height: 50px;
  font-size: 16px;
}

.hint-text {
  color: #909399;
  font-size: 13px;
  text-align: center;
  margin: 0 0 16px 0;
  line-height: 1.6;
}

.warning-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  margin-bottom: 16px;
  background-color: #fef0f0;
  border: 1px solid #fbc4c4;
  border-radius: 8px;
  color: #f56c6c;
  font-size: 14px;
}

/* 角色卡片区域 */
.character-cards-area {
  margin: 24px 0;
}

.character-card {
  margin-bottom: 16px;
  border-radius: 12px;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
  border: 2px solid transparent;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.character-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #409eff, #67c23a, #e6a23c);
  opacity: 0;
  transition: opacity 0.3s;
}

.character-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  border-color: #e4e7ed;
}

.character-card:hover::before {
  opacity: 1;
}

.character-card.has-image {
  background: linear-gradient(135deg, #f0f9ff 0%, #e8f4f8 100%);
  border-color: #67c23a;
}

.character-card.has-image::before {
  background: linear-gradient(90deg, #67c23a, #85ce61);
  opacity: 1;
}

.character-card.selected {
  background: linear-gradient(135deg, #ecf5ff 0%, #d9ecff 100%);
  border-color: #409eff;
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.25);
}

.character-card.selected::before {
  background: linear-gradient(90deg, #409eff, #66b1ff);
  opacity: 1;
}

.card-checkbox {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 2;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(4px);
  padding: 4px;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.batch-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  margin-bottom: 20px;
  background: #ecf5ff;
  border-radius: 8px;
  border: 1px solid #d9ecff;
}

.batch-toolbar-compact {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #f5f7fa;
  border-radius: 6px;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.stats-compact {
  color: #909399;
  font-size: 13px;
  padding-right: 12px;
  border-right: 1px solid #dcdfe6;
}

.selection-info {
  color: #606266;
  font-size: 13px;
}

.character-cards-area-fullscreen {
  flex: 1;
  overflow-y: auto;
  padding-right: 8px;
}

.character-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 220px;
  margin: -2px -2px 12px -2px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8eaf0 100%);
  position: relative;
  overflow: hidden;
}

.character-preview::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.05) 0%, rgba(103, 194, 58, 0.05) 100%);
  pointer-events: none;
}

.character-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.character-info {
  margin-bottom: 10px;
  padding: 0 4px;
  text-align: center;
}

.character-info h4 {
  margin: 0 0 6px 0;
  font-size: 14px;
  font-weight: 700;
  color: #303133;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.character-info .desc {
  font-size: 12px;
  color: #606266;
  margin: 8px 0 0 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  background: rgba(245, 247, 250, 0.5);
  padding: 6px 8px;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.character-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 10px;
  padding: 0 4px 4px;
}

.character-actions .el-button {
  width: 100%;
  border-radius: 8px;
  font-weight: 500;
  transition: all 0.3s;
}

.character-actions .el-button--primary {
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
  border: none;
}

.character-actions .el-button--primary:hover {
  background: linear-gradient(135deg, #66b1ff 0%, #409eff 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4);
}

.character-actions .el-button:not(.el-button--primary) {
  background: #ffffff;
  border: 1px solid #dcdfe6;
}

.character-actions .el-button:not(.el-button--primary):hover {
  background: #f5f7fa;
  border-color: #409eff;
  color: #409eff;
  transform: translateY(-1px);
}

/* 添加角色卡片样式 */
.add-character-card {
  cursor: pointer;
  border: 2px dashed #dcdfe6;
  background: linear-gradient(135deg, #fafbfc 0%, #f5f7fa 100%);
  transition: all 0.3s;
}

.add-character-card:hover {
  border-color: #409eff;
  background: linear-gradient(135deg, #ecf5ff 0%, #e6f2ff 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.15);
}

.add-character-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 400px;
  gap: 16px;
}

.add-character-content .add-text {
  font-size: 16px;
  font-weight: 600;
  color: #606266;
  transition: color 0.3s;
}

.add-character-card:hover .add-text {
  color: #409eff;
}

.add-character-card:hover .el-icon {
  color: #409eff !important;
}

/* 角色库样式 */
.library-grid {
  max-height: 500px;
  overflow-y: auto;
}

.library-item {
  cursor: pointer;
  margin-bottom: 16px;
  transition: all 0.3s;
}

.library-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  border-color: #409eff;
}

.library-image {
  width: 100%;
  height: 150px;
  object-fit: cover;
  border-radius: 4px;
  margin-bottom: 8px;
}

.library-info {
  text-align: center;
}

.library-name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.navigation-buttons {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin: 40px 0 20px;
}

/* 概览区域样式 */
.episode-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.episode-info h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.script-content-display {
  width: 100%;
}

.script-display :deep(.el-textarea__inner) {
  background: #fafafa;
  border: 1px solid #e4e7ed;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  line-height: 1.8;
}

.overview-section {
  margin-top: 24px;
}

.overview-section h3 {
  margin: 16px 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.action-buttons {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin: 20px 0;
}

/* 分镜列表样式 */
.shots-list {
  width: 100%;
}

.shots-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.shots-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.empty-shots {
  padding: 60px 0;
}

/* 创建章节提示 */
.create-chapter-prompt {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 500px;
}

.overview-content {
  background: #fafafa;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.overview-item {
  margin-bottom: 12px;
  line-height: 1.8;
}

.overview-item:last-child {
  margin-bottom: 0;
}

.overview-item .label {
  font-weight: 600;
  color: #606266;
  margin-right: 8px;
}

.overview-item .value {
  color: #303133;
}

.character-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.character-tag {
  padding: 8px 16px;
}

.action-buttons {
  display: flex;
  gap: 16px;
  align-items: center;
  justify-content: center;
}

/* 剧本生成表单样式 */
.generation-form {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 12px 16px;
  gap: 10px;
}

.script-input-header {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  flex-shrink: 0;
}


.script-textarea {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.script-textarea-fullscreen {
  flex: 1;
  display: flex;
  flex-direction: column;
}

:deep(.script-textarea-fullscreen .el-textarea) {
  height: 100%;
  display: flex;
  flex-direction: column;
}

:deep(.script-textarea-fullscreen .el-textarea__inner) {
  flex: 1;
  height: 100% !important;
  min-height: 700px !important;
  resize: none;
}

:deep(.script-textarea .el-textarea__inner) {
  background: #ffffff;
  color: #303133;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  padding: 16px;
  font-size: 15px;
  line-height: 1.8;
}

:deep(.script-textarea .el-textarea__inner:focus) {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.1);
}
</style>
