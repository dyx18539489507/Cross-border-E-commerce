<template>
  <div class="professional-editor">
    <!-- 顶部工具栏 -->
    <AppHeader :fixed="false" :show-logo="false" @config-updated="loadVideoModels">
      <template #left>
        <el-button text @click="goBack" class="back-btn">
          <el-icon><ArrowLeft /></el-icon>
          <span>{{ $t('editor.backToEpisode') }}</span>
        </el-button>
        <span class="episode-title">{{ drama?.title }} - {{ $t('editor.episode', { number: episodeNumber }) }}</span>
      </template>
    </AppHeader>

    <!-- 主编辑区域 -->
    <div class="editor-main">
      <!-- 左侧分镜列表 -->
      <div class="storyboard-panel">
        <div class="panel-header">
          <h3>{{ $t('storyboard.scriptStructure') }}</h3>
          <el-button text :icon="Plus" @click="handleAddStoryboard">{{ $t('storyboard.add') }}</el-button>
        </div>

        <div class="storyboard-list">
          <div v-for="(shot, index) in storyboards" :key="shot.id" class="storyboard-item"
            :class="{ active: String(currentStoryboardId) === String(shot.id) }" @click="selectStoryboard(shot.id)">
            <div class="shot-content">
              <div class="shot-header">
                <div class="shot-title-row">
                  <span class="shot-number">{{ $t('storyboard.shotNumber', { number: shot.storyboard_number }) }}</span>
                  <span class="shot-title">{{ shot.title || $t('storyboard.untitled') }}</span>
                </div>
                <div class="shot-duration">{{ shot.duration }}s</div>
              </div>
              <div class="shot-action" v-if="shot.action">{{ shot.action }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 中间时间线编辑区域 -->
      <div class="timeline-area">
        <VideoTimelineEditor ref="timelineEditorRef" v-if="storyboards.length > 0" :scenes="storyboards"
          :episode-id="episodeId.toString()" :drama-id="dramaId.toString()" :assets="videoAssets"
          @select-scene="handleTimelineSelect" @asset-deleted="loadVideoAssets"
          @merge-completed="handleMergeCompleted" />
        <el-empty v-else :description="$t('storyboard.noStoryboard')" class="empty-timeline" />
      </div>

      <!-- 右侧编辑面板 -->
      <div class="edit-panel">
        <el-tabs v-model="activeTab" class="edit-tabs">
          <!-- 镜头属性标签 -->
          <el-tab-pane :label="$t('storyboard.shotProperties')" name="shot" v-if="currentStoryboard">
            <div v-if="currentStoryboard" class="shot-editor-new">
              <!-- 场景(Scene) -->
              <div class="scene-section">
                <div class="section-label">
                  {{ $t('storyboard.scene') }} (Scene)
                  <el-button size="small" text @click="showSceneSelector = true">{{ $t('storyboard.selectScene')
                    }}</el-button>
                </div>
                <div class="scene-preview" v-if="currentStoryboard.background?.image_url" @click="showSceneImage">
                  <img :src="fixMediaUrl(currentStoryboard.background.image_url)" alt="场景" style="cursor: pointer;" />
                  <div class="scene-info">
                    <div>{{ currentStoryboard.background.location }} · {{ currentStoryboard.background.time }}</div>
                    <div class="scene-id">{{ $t('editor.sceneId') }}: {{ currentStoryboard.scene_id || 'N/A' }}</div>
                  </div>
                </div>
                <div class="scene-preview-empty" v-else>
                  <el-icon :size="48" color="#666">
                    <Picture />
                  </el-icon>
                  <div>{{ currentStoryboard.background ? $t('editor.sceneGenerating') : $t('editor.noBackground') }}
                  </div>
                </div>
              </div>

              <!-- 登场角色(Cast) -->
              <div class="cast-section">
                <div class="section-label">
                  {{ $t('editor.cast') }} (Cast)
                  <el-button size="small" text :icon="Plus" @click="showCharacterSelector = true">{{
                    $t('editor.addCharacter') }}</el-button>
                </div>
                <div class="cast-list">
                  <div v-for="char in currentStoryboardCharacters" :key="char.id" class="cast-item active">
                    <div class="cast-avatar" @click="showCharacterImage(char)">
                      <img v-if="char.image_url" :src="fixMediaUrl(char.image_url)" :alt="char.name" />
                      <span v-else>{{ char.name?.[0] || '?' }}</span>
                    </div>
                    <div class="cast-name">{{ char.name }}</div>
                    <div class="cast-remove" @click.stop="toggleCharacterInShot(char.id)"
                      :title="$t('editor.removeCharacter')">
                      <el-icon :size="14">
                        <Close />
                      </el-icon>
                    </div>
                  </div>
                  <div v-if="!currentStoryboard?.characters || currentStoryboard.characters.length === 0"
                    class="cast-empty">
                    {{ $t('editor.noCharacters') }}
                  </div>
                </div>
              </div>

              <!-- 视效设置 -->
              <div class="settings-section">
                <div class="section-label">{{ $t('editor.visualSettings') }}</div>
                <div class="settings-grid">
                  <div class="setting-item">
                    <label>{{ $t('editor.shotType') }}</label>
                    <el-select v-model="currentStoryboard.shot_type" clearable
                      :placeholder="$t('editor.shotTypePlaceholder')" @change="saveStoryboardField('shot_type')">
                      <el-option label="大远景" value="大远景" />
                      <el-option label="远景" value="远景" />
                      <el-option label="全景" value="全景" />
                      <el-option label="中全景" value="中全景" />
                      <el-option label="中景" value="中景" />
                      <el-option label="中近景" value="中近景" />
                      <el-option label="近景" value="近景" />
                      <el-option label="特写" value="特写" />
                      <el-option label="大特写" value="大特写" />
                    </el-select>
                  </div>

                  <div class="setting-item">
                    <label>{{ $t('editor.movement') }}</label>
                    <el-select v-model="currentStoryboard.movement" clearable
                      :placeholder="$t('editor.movementPlaceholder')" @change="saveStoryboardField('movement')">
                      <el-option label="固定镜头" value="固定镜头" />
                      <el-option label="推镜" value="推镜" />
                      <el-option label="拉镜" value="拉镜" />
                      <el-option label="摇镜" value="摇镜" />
                      <el-option label="移镜" value="移镜" />
                      <el-option label="跟镜" value="跟镜" />
                      <el-option label="升降镜头" value="升降镜头" />
                      <el-option label="环绕" value="环绕" />
                      <el-option label="甩镜" value="甩镜" />
                      <el-option label="变焦" value="变焦" />
                      <el-option label="手持晃动" value="手持晃动" />
                      <el-option label="稳定器运动" value="稳定器运动" />
                      <el-option label="轨道推拉" value="轨道推拉" />
                      <el-option label="航拍" value="航拍" />
                    </el-select>
                  </div>

                  <div class="setting-item">
                    <label>{{ $t('editor.angle') }}</label>
                    <el-select v-model="currentStoryboard.angle" clearable
                      :placeholder="$t('editor.anglePlaceholder')" @change="saveStoryboardField('angle')">
                      <el-option label="平视" value="平视" />
                      <el-option label="俯视" value="俯视" />
                      <el-option label="仰视" value="仰视" />
                      <el-option label="大俯视（鸟瞰）" value="大俯视（鸟瞰）" />
                      <el-option label="大仰视" value="大仰视" />
                      <el-option label="正侧面" value="正侧面" />
                      <el-option label="斜侧面" value="斜侧面" />
                      <el-option label="背面" value="背面" />
                      <el-option label="倾斜（荷兰角）" value="倾斜（荷兰角）" />
                      <el-option label="主观视角" value="主观视角" />
                      <el-option label="过肩" value="过肩" />
                    </el-select>
                  </div>
                </div>
              </div>

              <!-- 叙事内容 -->
              <div class="narrative-section">
                <div class="section-label">{{ $t('editor.action') }} (Action)</div>
                <el-input v-model="currentStoryboard.action" type="textarea" :rows="3"
                  :placeholder="$t('editor.actionPlaceholder')" @blur="saveStoryboardField('action')" />
              </div>

              <div class="narrative-section">
                <div class="section-label">{{ $t('editor.result') }} (Result)</div>
                <el-input v-model="currentStoryboard.result" type="textarea" :rows="2"
                  :placeholder="$t('editor.resultPlaceholder')" @blur="saveStoryboardField('result')" />
              </div>

              <div class="dialogue-section">
                <div class="section-label">{{ $t('editor.dialogue') }} (Dialogue)</div>
                <el-input v-model="currentStoryboard.dialogue" type="textarea" :rows="3"
                  :placeholder="$t('editor.dialoguePlaceholder')" @blur="saveStoryboardField('dialogue')" />
              </div>

              <div class="narrative-section">
                <div class="section-label">{{ $t('editor.description') }} (Description)</div>
                <el-input v-model="currentStoryboard.description" type="textarea" :rows="3"
                  :placeholder="$t('editor.descriptionPlaceholder')" @blur="saveStoryboardField('description')" />
              </div>

              <!-- 音效设置 -->
              <div class="settings-section">
                <div class="section-label">{{ $t('editor.soundEffects') }}</div>
                <div class="audio-controls">
                  <el-input v-model="currentStoryboard.sound_effect" :placeholder="$t('editor.soundEffectsPlaceholder')"
                    size="small" type="textarea" :rows="2" @blur="saveStoryboardField('sound_effect')" />
                </div>
              </div>

              <!-- 配乐设置 -->
              <div class="settings-section">
                <div class="section-label">{{ $t('editor.bgmPrompt') }}</div>
                <div class="audio-controls">
                  <el-input v-model="currentStoryboard.bgm_prompt" :placeholder="$t('editor.bgmPromptPlaceholder')"
                    size="small" type="textarea" :rows="2" @blur="saveStoryboardField('bgm_prompt')" />
                </div>
              </div>

              <!-- 氛围设置 -->
              <div class="settings-section">
                <div class="section-label">{{ $t('editor.atmosphere') }}</div>
                <div class="audio-controls">
                  <el-input v-model="currentStoryboard.atmosphere" :placeholder="$t('editor.atmospherePlaceholder')"
                    size="small" type="textarea" :rows="2" @blur="saveStoryboardField('atmosphere')" />
                </div>
              </div>
            </div>
            <el-empty v-else :description="$t('editor.noShotSelected')" />
          </el-tab-pane>

          <!-- 图片生成标签 -->
          <el-tab-pane :label="$t('editor.shotImage')" name="image">
            <div class="tab-content" v-if="currentStoryboard">
              <div class="image-generation-section">
                <!-- 帧类型选择 -->
                <div class="frame-type-selector">
                  <div class="section-label">{{ $t('editor.selectFrameType') }}</div>
                  <el-radio-group v-model="selectedFrameType" size="small">
                    <el-radio-button label="first">{{ $t('editor.firstFrame') }}</el-radio-button>
                    <el-radio-button label="last">{{ $t('editor.lastFrame') }}</el-radio-button>
                    <el-radio-button label="panel">{{ $t('editor.panelFrame') }}</el-radio-button>
                    <el-radio-button label="action">{{ $t('editor.actionSequence') }}</el-radio-button>
                    <el-radio-button label="key">{{ $t('editor.keyFrame') }}</el-radio-button>
                  </el-radio-group>
                  <el-input-number v-if="selectedFrameType === 'panel'" v-model="panelCount" :min="2" :max="6"
                    size="small" class="panel-count-input" style="margin-left: 10px; margin-top: 12px;" />
                  <span v-if="selectedFrameType === 'panel'" class="panel-count-label">{{ $t('editor.panelCount')
                    }}</span>
                </div>

                <!-- 提示词区域 -->
                <div class="prompt-section">
                  <div class="section-label">
                    {{ $t('editor.prompt') }}
                    <el-button size="small" type="primary" :disabled="generatingPrompt" :loading="generatingPrompt"
                      @click="extractFramePrompt" style="margin-left: 10px;">
                      {{ $t('editor.extractPrompt') }}
                    </el-button>
                  </div>
                  <el-input v-model="currentFramePrompt" type="textarea" :rows="8"
                    :placeholder="$t('editor.promptPlaceholder')" />
                </div>

                <!-- 生成控制 -->
                <div class="generation-controls">
                  <el-button type="success" :icon="MagicStick" :loading="generatingImage"
                    :disabled="!currentFramePrompt" @click="generateFrameImage">
                    {{ generatingImage ? $t('editor.generating') : $t('editor.generateImage') }}
                  </el-button>
                  <el-button :icon="Upload" :loading="uploadingImage" @click="uploadImage">
                    {{ $t('editor.uploadImage') }}
                  </el-button>
                  <input
                    ref="uploadInputRef"
                    type="file"
                    accept="image/*"
                    style="display: none;"
                    @change="handleUploadImage"
                  />
                </div>

                <!-- 生成结果 -->
                <div class="generation-result" v-if="generatedImages.length > 0">
                  <div class="section-label">{{ $t('editor.generationResult') }} ({{ generatedImages.length }})</div>
                  <div class="image-grid image-grid--generated">
                    <div
                      v-for="img in generatedImages"
                      :key="img.id"
                      class="image-item"
                      :class="{ 'has-image': !!img.image_url }"
                      :style="img.image_url ? { '--image-url': `url('${fixMediaUrl(img.image_url)}')` } : undefined"
                    >
                      <el-image v-if="img.image_url" :src="fixMediaUrl(img.image_url)"
                        :preview-src-list="generatedImages.filter(i => i.image_url).map(i => fixMediaUrl(i.image_url!))"
                        :initial-index="generatedImages.filter(i => i.image_url).findIndex(i => i.id === img.id)"
                        fit="contain" preview-teleported />
                      <div v-else class="image-placeholder">
                        <el-icon :size="32">
                          <Picture />
                        </el-icon>
                        <p>生成中...</p>
                      </div>
                      <div class="image-info">
                        <el-tag :type="getStatusType(img.status)" size="small">{{ getStatusText(img.status) }}</el-tag>
                        <span v-if="img.frame_type" class="frame-type-tag">{{ getFrameTypeText(img.frame_type) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <el-empty v-else description="未选择镜头" />
          </el-tab-pane>

          <!-- 视频生成标签 -->
          <el-tab-pane :label="$t('video.videoGeneration')" name="video">
            <div class="tab-content" v-if="currentStoryboard">
              <div class="video-generation-section">
                <!-- 生成提示词展示 -->
                <div class="video-prompt-box">
                  {{ currentStoryboard.video_prompt || '暂无提示词' }}
                </div>

                <!-- 视频参数设置 -->
                <div class="video-params-section">
                  <!--
                  <div class="param-row">
                    <span class="param-label">{{ $t('video.model') }}</span>
                    <el-select v-model="selectedVideoModel" :placeholder="$t('video.selectVideoModel')" size="default"
                      style="flex: 1;">
                      <el-option v-for="model in videoModelCapabilities" :key="model.id" :label="model.name"
                        :value="model.id">
                        <div style="display: flex; justify-content: space-between; align-items: center;">
                          <span>{{ model.name }}</span>
                          <div class="model-tags">
                            <el-tag v-if="model.supportMultipleImages" size="small" type="success"
                              style="margin-left: 4px;">多图</el-tag>
                            <el-tag v-if="model.supportFirstLastFrame" size="small" type="primary"
                              style="margin-left: 4px;">首尾帧</el-tag>
                            <el-tag size="small" type="info" style="margin-left: 4px;">最多{{ model.maxImages }}张</el-tag>
                          </div>
                        </div>
                      </el-option>
                    </el-select>
                  </div>
                  -->

                  <!-- 参考图模式选择 -->
                  <div v-if="selectedVideoModel && availableReferenceModes.length > 0" class="param-row">
                    <span class="param-label">参考图</span>
                    <el-select v-model="selectedReferenceMode" placeholder="请选择参考图模式" size="default" style="flex: 1;">
                      <el-option v-for="mode in availableReferenceModes" :key="mode.value" :label="mode.label"
                        :value="mode.value">
                        <div style="display: flex; justify-content: space-between; align-items: center;">
                          <span>{{ mode.label }}</span>
                          <span v-if="mode.description" class="mode-description">{{ mode.description }}</span>
                        </div>
                      </el-option>
                    </el-select>
                  </div>

                  <div class="param-row">
                    <span class="param-label">{{ $t('professionalEditor.duration') }}</span>
                    <div style="flex: 1; display: flex; align-items: center;">
                      <el-slider v-model="videoDuration" :min="4" :max="10" :step="1" show-stops style="flex: 1;" />
                      <span style="margin-left: 10px; min-width: 40px;">{{ videoDuration }}{{
                        $t('professionalEditor.seconds') }}</span>
                    </div>
                  </div>
                </div>

                <!-- 选择参考图片 -->
                <div v-if="selectedReferenceMode && selectedReferenceMode !== 'none'" class="reference-images-section"
                  style="margin-top: 0;">
                  <div class="frame-type-buttons" style="text-align: center; margin-bottom: 8px;">
                    <el-radio-group v-model="selectedVideoFrameType" size="default">
                      <el-radio-button label="first">首帧</el-radio-button>
                      <el-radio-button label="last">尾帧</el-radio-button>
                      <el-radio-button label="panel">分镜板</el-radio-button>
                      <el-radio-button label="action">动作序列</el-radio-button>
                      <el-radio-button label="key">关键帧</el-radio-button>
                    </el-radio-group>
                  </div>

                  <div class="frame-type-content">
                    <!-- 首帧 -->
                    <div v-show="selectedVideoFrameType === 'first'" class="image-scroll-container"
                      style="max-height: 280px; overflow-y: auto; overflow-x: hidden;">
                      
                      <!-- 上一镜头尾帧推荐（紧凑版） -->
                      <div v-if="previousStoryboardLastFrames.length > 0" class="previous-frame-section">
                        <div style="display: flex; align-items: center; gap: 6px; margin-bottom: 6px;">
                          <el-tag size="small" type="primary">
                            上一镜头 #{{ previousStoryboard?.storyboard_number }} 尾帧
                          </el-tag>
                          <span class="hint-text">点击添加为首帧参考</span>
                        </div>
                        <div style="display: flex; gap: 8px; flex-wrap: wrap;">
                          <div v-for="img in previousStoryboardLastFrames" :key="'prev-' + img.id" 
                            class="reference-item"
                            :class="{ selected: selectedImagesForVideo.includes(img.id) }"
                            style="position: relative; border: 2px solid #1890ff; border-radius: 4px; overflow: hidden; cursor: pointer;"
                            @click="selectPreviousLastFrame(img)">
                            <el-image :src="img.image_url" fit="cover"
                              style="width: 60px; height: 40px; display: block; pointer-events: none;" />
                            <div v-if="selectedImagesForVideo.includes(img.id)" 
                              style="position: absolute; top: 0; right: 0; background: #52c41a; color: #fff; font-size: 10px; padding: 1px 4px;">
                              ✓
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <!-- 当前镜头首帧列表 -->
                      <div class="reference-grid"
                        style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; max-width: 600px;">
                        <div
                          v-for="img in videoReferenceImages.filter(i => i.status === 'completed' && i.image_url && i.frame_type === 'first')"
                          :key="img.id" class="reference-item"
                          :class="{ selected: selectedImagesForVideo.includes(img.id) }" style="position: relative;"
                          @click="handleImageSelect(img.id)">
                          <el-image :src="fixMediaUrl(img.image_url)" fit="cover"
                            class="reference-image" />
                          <div class="preview-icon" @click.stop="previewImage(img.image_url)"
                            style="position: absolute; top: 4px; right: 4px; width: 24px; height: 24px; background: rgba(0,0,0,0.6); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; z-index: 10;">
                            <el-icon :size="14" color="#fff">
                              <ZoomIn />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                      <el-empty
                        v-if="!videoReferenceImages.some(i => i.status === 'completed' && i.image_url && i.frame_type === 'first') && previousStoryboardLastFrames.length === 0"
                        description="暂无首帧图片" size="small" />
                    </div>

                    <!-- 关键帧 -->
                    <div v-show="selectedVideoFrameType === 'key'" class="image-scroll-container"
                      style="max-height: 280px; overflow-y: auto; overflow-x: hidden;">
                      <div class="reference-grid"
                        style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; max-width: 600px;">
                        <div
                          v-for="img in videoReferenceImages.filter(i => i.status === 'completed' && i.image_url && i.frame_type === 'key')"
                          :key="img.id" class="reference-item"
                          :class="{ selected: selectedImagesForVideo.includes(img.id) }" style="position: relative;"
                          @click="handleImageSelect(img.id)">
                          <el-image :src="fixMediaUrl(img.image_url)" fit="cover"
                            class="reference-image" />
                          <div class="preview-icon" @click.stop="previewImage(img.image_url)"
                            style="position: absolute; top: 4px; right: 4px; width: 24px; height: 24px; background: rgba(0,0,0,0.6); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; z-index: 10;">
                            <el-icon :size="14" color="#fff">
                              <ZoomIn />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                      <el-empty
                        v-if="!videoReferenceImages.some(i => i.status === 'completed' && i.image_url && i.frame_type === 'key')"
                        description="暂无关键帧图片" size="small" />
                    </div>

                    <!-- 尾帧 -->
                    <div v-show="selectedVideoFrameType === 'last'" class="image-scroll-container"
                      style="max-height: 280px; overflow-y: auto; overflow-x: hidden;">
                      <div class="reference-grid"
                        style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; max-width: 600px;">
                        <div
                          v-for="img in videoReferenceImages.filter(i => i.status === 'completed' && i.image_url && i.frame_type === 'last')"
                          :key="img.id" class="reference-item"
                          :class="{ selected: selectedImagesForVideo.includes(img.id) }" style="position: relative;"
                          @click="handleImageSelect(img.id)">
                          <el-image :src="fixMediaUrl(img.image_url)" fit="cover"
                            class="reference-image" />
                          <div class="preview-icon" @click.stop="previewImage(img.image_url)"
                            style="position: absolute; top: 4px; right: 4px; width: 24px; height: 24px; background: rgba(0,0,0,0.6); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; z-index: 10;">
                            <el-icon :size="14" color="#fff">
                              <ZoomIn />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                      <el-empty
                        v-if="!videoReferenceImages.some(i => i.status === 'completed' && i.image_url && i.frame_type === 'last')"
                        description="暂无尾帧图片" size="small" />
                    </div>

                    <!-- 分镜板 -->
                    <div v-show="selectedVideoFrameType === 'panel'" class="image-scroll-container"
                      style="max-height: 280px; overflow-y: auto; overflow-x: hidden;">
                      <div class="reference-grid"
                        style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; max-width: 600px;">
                        <div
                          v-for="img in videoReferenceImages.filter(i => i.status === 'completed' && i.image_url && i.frame_type === 'panel')"
                          :key="img.id" class="reference-item"
                          :class="{ selected: selectedImagesForVideo.includes(img.id) }" style="position: relative;"
                          @click="handleImageSelect(img.id)">
                          <el-image :src="fixMediaUrl(img.image_url)" fit="cover"
                            class="reference-image" />
                          <div class="preview-icon" @click.stop="previewImage(img.image_url)"
                            style="position: absolute; top: 4px; right: 4px; width: 24px; height: 24px; background: rgba(0,0,0,0.6); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; z-index: 10;">
                            <el-icon :size="14" color="#fff">
                              <ZoomIn />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                      <el-empty
                        v-if="!videoReferenceImages.some(i => i.status === 'completed' && i.image_url && i.frame_type === 'panel')"
                        description="暂无分镜板图片" size="small" />
                    </div>

                    <!-- 动作序列 -->
                    <div v-show="selectedVideoFrameType === 'action'" class="image-scroll-container"
                      style="max-height: 280px; overflow-y: auto; overflow-x: hidden;">
                      <div class="reference-grid"
                        style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; max-width: 600px;">
                        <div
                          v-for="img in videoReferenceImages.filter(i => i.status === 'completed' && i.image_url && i.frame_type === 'action')"
                          :key="img.id" class="reference-item"
                          :class="{ selected: selectedImagesForVideo.includes(img.id) }" style="position: relative;"
                          @click="handleImageSelect(img.id)">
                          <el-image :src="fixMediaUrl(img.image_url)" fit="cover"
                            class="reference-image" />
                          <div class="preview-icon" @click.stop="previewImage(img.image_url)"
                            style="position: absolute; top: 4px; right: 4px; width: 24px; height: 24px; background: rgba(0,0,0,0.6); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; z-index: 10;">
                            <el-icon :size="14" color="#fff">
                              <ZoomIn />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                      <el-empty
                        v-if="!videoReferenceImages.some(i => i.status === 'completed' && i.image_url && i.frame_type === 'action')"
                        description="暂无动作序列图片" size="small" />
                    </div>
                  </div>
                </div>

                <!-- 参考图片设置 -->
                <div v-if="selectedReferenceMode && selectedReferenceMode !== 'none'" class="reference-config-section"
                  style="margin-top: 24px;">
                  <!-- 图片框配置区 -->
                  <div class="image-slots-container" style="margin-top: 16px; margin-bottom: 24px;">
                    <!-- 单图模式 -->
                    <div v-if="selectedReferenceMode === 'single'" style="text-align: center;">
                      <div class="reference-mode-title">单图参考</div>
                      <div style="display: inline-block;">
                        <div class="image-slot"
                          @click="selectedImagesForVideo.length > 0 && removeSelectedImage(selectedImagesForVideo[0])">
                          <img v-if="selectedImageObjects[0]" :src="fixMediaUrl(selectedImageObjects[0].image_url)" alt=""
                            style="width: 100%; height: 100%; object-fit: contain;" />
                          <div v-else class="image-slot-placeholder">
                            <el-icon :size="32" color="#c0c4cc">
                              <Plus />
                            </el-icon>
                            <div class="slot-hint">点击上方选择图片</div>
                          </div>
                          <div v-if="selectedImageObjects[0]" class="image-slot-remove">
                            <el-icon :size="16" color="#fff">
                              <Close />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                    </div>

                    <!-- 首尾帧模式 -->
                    <div v-else-if="selectedReferenceMode === 'first_last'" style="text-align: center;">
                      <div class="reference-mode-title">首尾帧</div>
                      <div style="display: flex; gap: 20px; justify-content: center; align-items: center;">
                        <div>
                          <div class="frame-label">首帧</div>
                          <div class="image-slot"
                            @click="firstFrameSlotImage && removeSelectedImage(firstFrameSlotImage.id)">
                            <img v-if="firstFrameSlotImage" :src="fixMediaUrl(firstFrameSlotImage.image_url)" alt=""
                              style="width: 100%; height: 100%; object-fit: contain;" />
                            <div v-else class="image-slot-placeholder">
                              <el-icon :size="32" color="#c0c4cc">
                                <Plus />
                              </el-icon>
                              <div class="slot-hint">选择首帧</div>
                            </div>
                            <div v-if="firstFrameSlotImage" class="image-slot-remove">
                              <el-icon :size="16" color="#fff">
                                <Close />
                              </el-icon>
                            </div>
                          </div>
                        </div>
                        <el-icon :size="24" color="#909399">
                          <Right />
                        </el-icon>
                        <div>
                          <div class="frame-label">尾帧</div>
                          <div class="image-slot"
                            @click="lastFrameSlotImage && removeSelectedImage(lastFrameSlotImage.id)">
                            <img v-if="lastFrameSlotImage" :src="fixMediaUrl(lastFrameSlotImage.image_url)" alt=""
                              style="width: 100%; height: 100%; object-fit: contain;" />
                            <div v-else class="image-slot-placeholder">
                              <el-icon :size="32" color="#c0c4cc">
                                <Plus />
                              </el-icon>
                              <div class="slot-hint">选择尾帧</div>
                            </div>
                            <div v-if="lastFrameSlotImage" class="image-slot-remove">
                              <el-icon :size="16" color="#fff">
                                <Close />
                              </el-icon>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>

                    <!-- 多图模式 -->
                    <div v-else-if="selectedReferenceMode === 'multiple'" style="text-align: center;">
                      <div class="reference-mode-title">
                        多图参考 ({{ selectedImagesForVideo.length }}/{{ currentModelCapability?.maxImages || 6 }})
                      </div>
                      <div style="display: flex; gap: 12px; justify-content: center; flex-wrap: wrap;">
                        <div v-for="index in (currentModelCapability?.maxImages || 6)" :key="index"
                          class="image-slot image-slot-small"
                          @click="selectedImageObjects[index - 1] && removeSelectedImage(selectedImageObjects[index - 1].id)">
                          <img v-if="selectedImageObjects[index - 1]" :src="fixMediaUrl(selectedImageObjects[index - 1].image_url)"
                            alt="" style="width: 100%; height: 100%; object-fit: contain;" />
                          <div v-else class="image-slot-placeholder">
                            <el-icon :size="20" color="#c0c4cc">
                              <Plus />
                            </el-icon>
                            <div class="image-slot-index">{{ index }}</div>
                          </div>
                          <div v-if="selectedImageObjects[index - 1]" class="image-slot-remove">
                            <el-icon :size="14" color="#fff">
                              <Close />
                            </el-icon>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 生成控制 -->
                <div class="generation-controls" style="margin-top: 32px; text-align: center;">
                  <el-button type="primary" :icon="VideoCamera" :loading="generatingVideo"
                    :disabled="!selectedVideoModel || (selectedReferenceMode !== 'none' && selectedImagesForVideo.length === 0)"
                    @click="generateVideo">
                    {{ generatingVideo ? '生成中...' : '生成视频' }}
                  </el-button>
                </div>

                <!-- 生成的视频列表 -->
                <div class="generation-result" v-if="generatedVideos.length > 0" style="margin-top: 24px;">
                  <div class="section-label"
                    style="font-size: 13px; font-weight: 600; margin-bottom: 12px; display: flex; align-items: center; gap: 6px;">
                    <span></span>
                    生成结果 ({{ generatedVideos.length }})
                  </div>
                  <div class="image-grid"
                    style="display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 10px;">
                    <div v-for="video in generatedVideos" :key="video.id" class="image-item video-item">
                      <div class="video-thumbnail" v-if="video.video_url"
                    @mouseenter="(e) => ((e.currentTarget as HTMLElement).querySelector('.play-overlay') as HTMLElement).style.opacity = '1'"
                    @mouseleave="(e) => ((e.currentTarget as HTMLElement).querySelector('.play-overlay') as HTMLElement).style.opacity = '0'"
                        @click="playVideo(video)">
                        <video :src="resolveVideoUrl(video.video_url)"
                          :poster="video.first_frame_url ? fixMediaUrl(video.first_frame_url) : undefined"
                          preload="metadata"
                          style="width: 100%; height: 100%; object-fit: cover; display: block; pointer-events: none;" />
                        <div class="play-overlay"
                          style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: center; background: rgba(0, 0, 0, 0.3); opacity: 0; transition: opacity 0.2s;">
                          <el-icon :size="32" color="#fff" style="filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.3));">
                            <VideoPlay />
                          </el-icon>
                        </div>
                      </div>
                      <div v-else class="image-placeholder">
                        <el-icon v-if="video.status === 'failed'" :size="32" color="#f56c6c">
                          <WarningFilled />
                        </el-icon>
                        <el-icon v-else :size="32">
                          <VideoCamera />
                        </el-icon>
                        <p style="margin: 0; font-size: 11px;">
                          {{ video.status === 'failed' ? '生成失败' : '生成中...' }}
                        </p>
                      </div>
                      <div class="image-info">
                        <div style="display: flex; align-items: center; gap: 4px;">
                          <el-tag :type="getStatusType(video.status)" size="small"
                            style="font-size: 10px; height: 20px; padding: 0 6px;">{{ getStatusText(video.status)
                            }}</el-tag>
                        </div>
                        <div style="display: flex; gap: 4px;">
                          <el-button v-if="video.status === 'completed' && video.video_url" type="success" size="small"
                            :loading="addingToAssets.has(video.id)" @click.stop="addVideoToAssets(video)">
                            {{ addingToAssets.has(video.id) ? '添加中...' : (isVideoInAssetLibrary(video) ? '已在素材库' : '添加到素材库') }}
                          </el-button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div ref="videoGenerationBottomRef"></div>
              </div>
            </div>
            <el-empty v-else description="未选择镜头" />
          </el-tab-pane>

          <!-- 音效与配乐标签 -->
          <el-tab-pane :label="$t('video.soundAndMusicTab')" name="audio">
            <div class="tab-content sound-music-tab-content">
              <div class="sound-music-panel">
                <div class="sound-music-header">
                  <div class="sound-music-title">{{ $t('video.soundMusicTitle') }}</div>
                  <div class="sound-music-subtitle">{{ $t('video.soundMusicDesc') }}</div>
                  <!--
                  <div v-if="douyinMusicUpdatedAt" class="sound-music-meta">
                    {{ $t('video.soundMusicUpdatedAt') }} {{ douyinMusicUpdatedAt }}
                  </div>
                  -->
                </div>

                <div class="sound-music-filters">
                  <el-radio-group v-model="audioMode" size="small">
                    <el-radio-button label="music">{{ $t('video.soundMusicMusicTab') }}</el-radio-button>
                    <el-radio-button label="sfx">{{ $t('video.soundMusicSfxTab') }}</el-radio-button>
                  </el-radio-group>
                  <el-input
                    v-model="audioSearch"
                    size="small"
                    clearable
                    :placeholder="$t('video.soundMusicSearchPlaceholder')"
                    class="audio-search"
                  />
                  <div v-if="audioMode === 'sfx'" class="sfx-categories">
                    <el-radio-group v-model="sfxCategory" size="small">
                      <el-radio-button v-for="cat in sfxCategories" :key="cat" :label="cat">{{ cat }}</el-radio-button>
                    </el-radio-group>
                  </div>
                  <!--
                  <div v-if="audioMode === 'sfx'" class="sfx-ai-generate">
                    <el-input
                      v-model="sfxPrompt"
                      size="small"
                      clearable
                      class="sfx-ai-input"
                      placeholder="输入想要生成的音效，如：开门、爆炸"
                      @keyup.enter="generateSfx"
                    />
                    <el-button type="primary" size="small" :loading="generatingSfx" @click="generateSfx">
                      生成音效
                    </el-button>
                  </div>
                  -->
                  <!--
                  <el-select
                    v-model="audioCategory"
                    size="small"
                    class="audio-category"
                  >
                    <el-option
                      v-for="option in audioCategoryOptions"
                      :key="option.value"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                  -->
                  <!--
                  <el-switch
                    v-model="audioHotOnly"
                    :active-text="$t('video.soundMusicHotOnly')"
                  />
                  -->
                </div>

                <div ref="audioListRef" class="audio-list" v-loading="showAudioListOverlay">
                  <div v-if="audioSearch.trim() && audioSearchLoading" class="audio-search-status">
                    <el-icon class="rotating"><Loading /></el-icon>
                    <span>正在加载中...</span>
                  </div>
                  <el-empty
                    v-if="filteredAudioAssets.length === 0 && !audioSearchLoading"
                    :description="$t('video.soundMusicEmpty')"
                  />
                  <div v-else class="audio-grid">
                    <div
                      v-for="asset in pagedAudioAssets"
                      :key="asset.id"
                      class="audio-card"
                      :class="{
                        previewing: previewingAudioId === asset.id,
                        'preview-loading': previewLoadingAudioId === asset.id
                      }"
                    >
                      <div class="audio-card-main">
                        <div class="audio-icon">
                          <img v-if="asset.cover" :src="asset.cover" alt="cover" class="audio-cover" />
                          <el-icon v-else><Headset /></el-icon>
                        </div>
                        <div class="audio-info">
                          <div class="audio-name">{{ asset.name }}</div>
                          <div v-if="asset.artist && asset.source !== 'sfx'" class="audio-artist">{{ asset.artist }}</div>
                          <div class="audio-meta">
                            <el-tag v-if="asset.category" size="small">{{ asset.category }}</el-tag>
                            <el-tag v-if="isDouyinHot(asset)" size="small" type="danger">抖音热门</el-tag>
                            <span v-if="asset.duration" class="audio-duration">{{ formatDuration(asset.duration) }}</span>
                            <span v-if="asset.view_count" class="audio-views">🔥 {{ asset.view_count }}</span>
                          </div>
                        </div>
                      </div>
                      <div class="audio-actions">
                        <el-button
                          size="small"
                          :loading="previewLoadingAudioId === asset.id"
                          :disabled="unsupportedPreviewAudioIds.has(asset.id) && previewingAudioId !== asset.id"
                          @click="toggleAudioPreview(asset)"
                        >
                          <el-icon v-if="previewLoadingAudioId !== asset.id">
                            <VideoPause v-if="previewingAudioId === asset.id && previewLoadingAudioId !== asset.id" />
                            <VideoPlay v-else />
                          </el-icon>
                          {{
                            previewLoadingAudioId === asset.id
                              ? '加载中...'
                              : (unsupportedPreviewAudioIds.has(asset.id) && previewingAudioId !== asset.id
                                  ? '暂不支持试听'
                                  : (previewingAudioId === asset.id ? $t('video.soundMusicStop') : $t('video.soundMusicPreview')))
                          }}
                        </el-button>
                        <el-button type="primary" size="small" @click="addAudioToTimeline(asset)">
                          <el-icon><Plus /></el-icon>
                          {{ $t('video.soundMusicAddToTrack') }}
                        </el-button>
                      </div>
                    </div>
                  </div>
                  <div
                    v-if="isAudioScrollLoading"
                    class="audio-scroll-loading"
                  >
                    <el-icon class="rotating"><Loading /></el-icon>
                    <span>正在加载中...</span>
                  </div>
                  <div
                    v-else-if="showAudioLazyTip"
                    class="audio-lazy-tip"
                  >
                    下滑加载更多
                  </div>
                  <div v-if="audioMode === 'music' && audioSearch.trim() && neteaseSearchTotal > 0" class="audio-pagination">
                    <span class="audio-total">共搜索到 {{ neteaseSearchTotal }} 条</span>
                    <el-pagination
                      v-model:current-page="audioSearchPage"
                      :page-size="audioSearchPageSize"
                      :total="neteaseSearchTotal"
                      layout="prev, pager, next"
                      @current-change="handleAudioSearchPageChange"
                    />
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>

          <!-- 视频合成列表标签 -->
          <el-tab-pane :label="$t('video.videoMerge')" name="merges">
            <div class="tab-content">
              <div class="merges-list" v-loading="loadingMerges && mergeListInitialized && videoMerges.length > 0">
                <div v-if="loadingMerges && !videoMerges.length" class="merge-loading-state">
                  <el-icon class="rotating" :size="28"><Loading /></el-icon>
                  <div class="merge-loading-text">正在加载合成视频...</div>
                </div>
                <el-empty v-else-if="videoMerges.length === 0" :description="$t('video.noMergeRecords')" :image-size="120">
                  <template #description>
                    <div style="color: var(--text-muted); font-size: 14px; margin-top: 12px;">
                      <p style="margin: 0;">{{ $t('video.noMergeYet') }}</p>
                      <p style="margin: 8px 0 0 0; font-size: 12px;">{{ $t('video.mergeInstructions') }}</p>
                    </div>
                  </template>
                </el-empty>
                <div v-else class="merge-items">
                  <div v-for="merge in videoMerges" :key="merge.id" class="merge-item"
                    :class="'merge-status-' + merge.status">
                    <!-- 状态指示条 -->
                    <div class="status-indicator"></div>

                    <!-- 主要内容区域 -->
                    <div class="merge-content">
                      <!-- 标题和状态 -->
                      <div class="merge-header">
                        <div class="title-section">
                          <el-icon :size="20" class="title-icon">
                            <VideoCamera v-if="merge.status === 'completed'" />
                            <Loading v-else-if="merge.status === 'processing'" class="rotating" />
                            <WarningFilled v-else-if="merge.status === 'failed'" />
                            <Clock v-else />
                          </el-icon>
                          <h3 class="merge-title">{{ merge.title }}</h3>
                        </div>
                        <el-tag
                          :type="merge.status === 'completed' ? 'success' : merge.status === 'failed' ? 'danger' : 'warning'"
                          effect="dark" size="large" round>
                          {{ merge.status === 'pending' ? '等待中' : merge.status === 'processing' ? '合成中' : merge.status
                            === 'completed' ?
                            '已完成' : '失败' }}
                        </el-tag>
                      </div>

                      <!-- 详细信息网格 -->
                      <div class="merge-details">
                        <div class="detail-item">
                          <div class="detail-icon">
                            <el-icon :size="16">
                              <Timer />
                            </el-icon>
                          </div>
                          <div class="detail-content">
                            <div class="detail-label">{{ $t('professionalEditor.videoDuration') }}</div>
                            <div class="detail-value">{{ merge.duration ? `${merge.duration}
                              ${$t('professionalEditor.seconds')}` : '-'
                              }}</div>
                          </div>
                        </div>
                        <div class="detail-item">
                          <div class="detail-icon">
                            <el-icon :size="16">
                              <Calendar />
                            </el-icon>
                          </div>
                          <div class="detail-content">
                            <div class="detail-label">创建时间</div>
                            <div class="detail-value">{{ formatDateTime(merge.created_at) }}</div>
                          </div>
                        </div>
                        <div class="detail-item" v-if="merge.completed_at">
                          <div class="detail-icon">
                            <el-icon :size="16">
                              <Check />
                            </el-icon>
                          </div>
                          <div class="detail-content">
                            <div class="detail-label">完成时间</div>
                            <div class="detail-value">{{ formatDateTime(merge.completed_at) }}</div>
                          </div>
                        </div>
                      </div>

                      <!-- 错误提示 -->
                      <div class="merge-error" v-if="merge.status === 'failed' && merge.error_msg">
                        <el-alert type="error" :closable="false" show-icon>
                          <template #title>
                            <div style="font-size: 13px; line-height: 1.5;">{{ merge.error_msg }}</div>
                          </template>
                        </el-alert>
                      </div>

                      <!-- 分发状态 -->
                      <div
                        v-if="merge.status === 'completed' && getMergeDistributionSummary(merge.id).length"
                        class="distribution-summary"
                      >
                        <span class="distribution-summary-label">分发状态：</span>
                        <div class="distribution-summary-list">
                          <el-tag
                            v-for="item in getMergeDistributionSummary(merge.id)"
                            :key="`${merge.id}-${item.platform}`"
                            :type="getDistributionStatusType(item.status)"
                            size="small"
                            effect="plain"
                            class="distribution-tag"
                            :class="{ 'is-link': !!item.published_url }"
                            @click="openDistributionRecord(item)"
                          >
                            {{ getPlatformLabel(item.platform) }} · {{ getDistributionStatusText(item.status) }}
                          </el-tag>
                        </div>
                      </div>

                      <!-- 操作按钮 -->
                      <div
                        class="merge-actions"
                        :class="{ 'single-action': !(merge.status === 'completed' && merge.merged_url) }"
                      >
                        <template v-if="merge.status === 'completed' && merge.merged_url">
                          <el-button type="primary" :icon="VideoCamera"
                            @click="downloadVideo(merge.merged_url, merge.title)" round>
                            下载视频
                          </el-button>
                          <el-button class="preview-button" :icon="View" @click="previewMergedVideo(merge.merged_url)" round>
                            在线预览
                          </el-button>
                          <el-button type="success" :icon="Connection" @click="openDistributionDialog(merge)" round>
                            一键分发
                          </el-button>
                        </template>
                        <el-button type="danger" :icon="Delete"
                          @click="deleteMerge(merge.id)" round>
                          删除
                        </el-button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <!-- 角色选择器对话框 -->
    <el-dialog v-model="showCharacterImagePreview" :title="previewCharacter?.name" width="600px">
      <div class="character-image-preview" v-if="previewCharacter">
        <img v-if="previewCharacter.image_url" :src="fixMediaUrl(previewCharacter.image_url)" :alt="previewCharacter.name" />
        <el-empty v-else description="暂无图片" />
      </div>
      <!-- ... -->
    </el-dialog>

    <!-- 场景大图预览对话框 -->
    <el-dialog v-model="showSceneImagePreview"
      :title="currentStoryboard?.background ? `${currentStoryboard.background.location} · ${currentStoryboard.background.time}` : '场景预览'"
      width="800px">
      <div class="scene-image-preview" v-if="currentStoryboard?.background?.image_url">
        <img :src="fixMediaUrl(currentStoryboard.background.image_url)" alt="场景" />
      </div>
    </el-dialog>

    <!-- 角色选择对话框 -->
    <el-dialog v-model="showCharacterSelector" title="添加角色到镜头" width="800px">
      <div class="character-selector-grid">
        <div v-for="char in availableCharacters" :key="char.id" class="character-card"
          :class="{ selected: isCharacterInCurrentShot(char.id) }" @click="toggleCharacterInShot(char.id)">
          <div class="character-avatar-large">
            <img v-if="char.image_url" :src="fixMediaUrl(char.image_url)" :alt="char.name" />
            <span v-else>{{ char.name?.[0] || '?' }}</span>
          </div>
          <div class="character-info">
            <div class="character-name">{{ char.name }}</div>
            <div class="character-role">{{ char.role || '角色' }}</div>
          </div>
          <div class="character-check" v-if="isCharacterInCurrentShot(char.id)">
            <el-icon color="#409eff" :size="24">
              <Check />
            </el-icon>
          </div>
        </div>
        <div v-if="availableCharacters.length === 0" class="empty-characters">
          <el-empty description="暂无角色，请先在剧集中创建角色" />
        </div>
      </div>
      <template #footer>
        <el-button @click="showCharacterSelector = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 场景选择对话框 -->
    <el-dialog v-model="showSceneSelector" title="选择场景背景" width="800px">
      <div class="scene-selector-grid">
        <div v-for="scene in availableScenes" :key="scene.id" class="scene-card"
          :class="{ selected: currentStoryboard?.scene_id === scene.id }" @click="selectScene(scene.id)">
          <div class="scene-image">
            <img v-if="scene.image_url" :src="fixMediaUrl(scene.image_url)" :alt="scene.location" />
            <el-icon v-else :size="48" color="#ccc">
              <Picture />
            </el-icon>
          </div>
          <div class="scene-info">
            <div class="scene-location">{{ scene.location }}</div>
            <div class="scene-time">{{ scene.time }}</div>
          </div>
        </div>
        <div v-if="availableScenes.length === 0" class="empty-scenes">
          <el-empty description="暂无可用场景" />
        </div>
      </div>
    </el-dialog>

    <DistributionDialog
      v-if="distributionTargetMerge"
      v-model="distributionDialogVisible"
      content-type="video"
      source-type="video_merge"
      :source-ref="distributionTargetMerge.id"
      :media-url="distributionTargetMerge.merged_url"
      :initial-title="getDefaultDistributionTitle(distributionTargetMerge)"
      dialog-title="视频一键分发"
    />

    <!-- 视频预览对话框 -->
    <el-dialog v-model="showVideoPreview" title="视频预览" width="800px" :close-on-click-modal="true" destroy-on-close>
      <div class="video-preview-container" v-if="previewVideo">
        <video v-if="previewVideo.video_url" :src="resolveVideoUrl(previewVideo.video_url)" controls autoplay
          style="width: 100%; max-height: 70vh; display: block; background: #000; border-radius: 8px;" />
        <div v-else style="text-align: center; padding: 40px;">
          <el-icon :size="48" color="#ccc">
            <VideoCamera />
          </el-icon>
          <p style="margin-top: 16px; color: var(--text-muted);">视频生成中...</p>
        </div>
        <div class="video-meta">
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <el-tag :type="getStatusType(previewVideo.status)" size="small">{{ getStatusText(previewVideo.status)
                }}</el-tag>
              <span v-if="previewVideo.duration" style="margin-left: 12px; color: var(--text-secondary); font-size: 14px;">{{
                $t('professionalEditor.duration') }}: {{ previewVideo.duration }}{{ $t('professionalEditor.seconds')
                }}</span>
            </div>
            <el-button v-if="previewVideo.video_url" size="small"
              @click="openVideoUrl(previewVideo.video_url)">
              {{ $t('professionalEditor.downloadVideo') }}
            </el-button>
          </div>
          <div v-if="previewVideo.prompt" style="margin-top: 12px; font-size: 12px; color: var(--text-secondary); line-height: 1.6;">
            <strong>提示词：</strong>{{ previewVideo.prompt }}
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onActivated, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, Plus, Picture, VideoPlay, VideoPause, View, Setting,
  Upload, MagicStick, VideoCamera, ZoomIn, ZoomOut, Top, Bottom, Check, Close, Right,
  Timer, Calendar, Clock, Loading, WarningFilled, Delete, Connection, Headset
} from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { generateFramePrompt, type FrameType } from '@/api/frame'
import { imageAPI } from '@/api/image'
import { videoAPI } from '@/api/video'
import { aiAPI } from '@/api/ai'
import { assetAPI } from '@/api/asset'
import {
  videoMergeAPI,
  type DistributionPlatform,
  type VideoDistribution,
  type VideoDistributionStatus,
  type VideoMerge
} from '@/api/videoMerge'
import {
  socialBindingAPI,
  type SocialAccountBinding,
  type SocialBindingPlatform
} from '@/api/socialBinding'
import type { ImageGeneration } from '@/types/image'
import type { VideoGeneration } from '@/types/video'
import type { AIServiceConfig } from '@/types/ai'
import type { Asset } from '@/types/asset'
import VideoTimelineEditor from '@/components/editor/VideoTimelineEditor.vue'
import DistributionDialog from '@/components/distribution/DistributionDialog.vue'
import type { Drama, Episode, Storyboard } from '@/types/drama'
import { AppHeader } from '@/components/common'

const route = useRoute()
const router = useRouter()
const { t: $t } = useI18n()

const dramaId = Number(route.params.dramaId)
const episodeNumber = Number(route.params.episodeNumber)
const episodeId = ref<string>('')

const drama = ref<Drama | null>(null)
const episode = ref<Episode | null>(null)
const storyboards = ref<Storyboard[]>([])
const characters = ref<any[]>([])
const availableScenes = ref<any[]>([])

const currentStoryboardId = ref<string | null>(null)
const activeTab = ref('shot')
const showSceneSelector = ref(false)
const showCharacterSelector = ref(false)
const showCharacterImagePreview = ref(false)
const previewCharacter = ref<any>(null)
const showSceneImagePreview = ref(false)
const showSettings = ref(false)
const showVideoPreview = ref(false)
const previewVideo = ref<VideoGeneration | null>(null)
const addingToAssets = ref<Set<number>>(new Set())

const currentPlayState = ref<'playing' | 'paused'>('paused')
const currentTime = ref(0)
const totalDuration = computed(() => {
  if (!Array.isArray(storyboards.value)) return 0
  return storyboards.value.reduce((sum, s) => sum + (s.duration || 0), 0)
})

const selectedCharacters = ref<number[]>([])
const narrativeTab = ref('shot-prompt')

// 图片生成相关状态
const selectedFrameType = ref<FrameType>('first')
const panelCount = ref(3)
const generatingPrompt = ref(false)
const framePrompts = ref<Record<string, string>>({
  key: '',
  first: '',
  last: '',
  panel: ''
})
const currentFramePrompt = ref('')
const generatingImage = ref(false)
const generatedImages = ref<ImageGeneration[]>([])
const isSwitchingFrameType = ref(false) // 标志位：是否正在切换帧类型
const loadingImages = ref(false)
let pollingTimer: any = null
let pollingFrameType: FrameType | null = null // 记录正在轮询的帧类型

// 视频生成相关状态
const videoDuration = ref(5)  // 默认5秒，会根据镜头duration自动更新
const selectedVideoFrameType = ref<FrameType>('first')
const selectedImagesForVideo = ref<number[]>([])
const selectedLastImageForVideo = ref<number | null>(null)
const generatingVideo = ref(false)
const generatedVideos = ref<VideoGeneration[]>([])
const videoAssets = ref<Asset[]>([])
const loadingVideos = ref(false)
const audioAssets = ref<Asset[]>([])
const loadingAudioAssets = ref(false)
const douyinMusicAssets = ref<AudioListItem[]>([])
const loadingDouyinMusic = ref(false)
const douyinMusicUpdatedAt = ref<string | null>(null)
const sfxAssets = ref<AudioListItem[]>([])
const loadingSfx = ref(false)
const sfxCategory = ref('热门')
const sfxPage = ref(1)
const sfxHasMore = ref(true)
const sfxLoadingMore = ref(false)
const neteaseSearchResults = ref<AudioListItem[]>([])
const neteaseSearchTotal = ref(0)
const loadingNeteaseSearch = ref(false)
const neteaseSearchError = ref<string | null>(null)
const audioMode = ref<'music' | 'sfx'>('music')
const audioSearch = ref('')
const audioSearchPage = ref(1)
const audioSearchPageSize = ref(10)
const audioCategory = ref('all')
const audioHotOnly = ref(true)
const audioListRef = ref<HTMLElement | null>(null)
const hotMusicVisibleCount = ref(20)
const previewingAudioId = ref<string | null>(null)
const previewLoadingAudioId = ref<string | null>(null)
const previewAudioPlayer = ref<HTMLAudioElement | null>(null)
const unsupportedPreviewAudioIds = ref<Set<string>>(new Set())
const timelineEditorRef = ref<InstanceType<typeof VideoTimelineEditor> | null>(null)
const videoReferenceImages = ref<ImageGeneration[]>([])
const DEFAULT_VIDEO_MODEL_ID = 'doubao-seedance-1-5-pro-251215'
const DEFAULT_REFERENCE_MODE = 'single'
const selectedVideoModel = ref<string>(DEFAULT_VIDEO_MODEL_ID)
const selectedReferenceMode = ref<string>(DEFAULT_REFERENCE_MODE)  // 参考图模式：single, first_last, multiple, none
const previewImageUrl = ref<string>('')  // 预览大图的URL
const videoModelCapabilities = ref<VideoModelCapability[]>([])
let videoPollingTimer: any = null
let mergePollingTimer: any = null  // 视频合成列表轮询定时器
const videoGenerationBottomRef = ref<HTMLElement | null>(null)

// 视频合成列表
const videoMerges = ref<VideoMerge[]>([])
const loadingMerges = ref(false)
const mergeListInitialized = ref(false)

type ManagedDistributionPlatform = SocialBindingPlatform

const distributionPlatforms: Array<{
  value: ManagedDistributionPlatform
  label: string
  placeholder: string
}> = [
  { value: 'discord', label: 'Discord', placeholder: '请输入 Discord 用户名、频道号或 webhook 标识' },
  { value: 'reddit', label: 'Reddit', placeholder: '请输入 Reddit 用户名或版块标识' },
  { value: 'pinterest', label: 'Pinterest', placeholder: '请输入 Pinterest 用户名或画板标识' }
]
const distributionDialogVisible = ref(false)
const submittingDistribution = ref(false)
const loadingSocialBindings = ref(false)
const distributionTargetMerge = ref<VideoMerge | null>(null)
const socialBindings = ref<Record<string, SocialAccountBinding>>({})
const bindingPlatformLoading = ref<Record<string, boolean>>({})
const distributionForm = ref<{
  mergeId: number
  platforms: DistributionPlatform[]
  title: string
  description: string
  hashtagsText: string
}>({
  mergeId: 0,
  platforms: [],
  title: '',
  description: '',
  hashtagsText: ''
})
const mergeDistributions = ref<Record<number, VideoDistribution[]>>({})
let distributionPollingTimer: any = null

// 视频模型能力配置
interface VideoModelCapability {
  id: string
  name: string
  supportMultipleImages: boolean  // 支持多张图片
  supportFirstLastFrame: boolean  // 支持首尾帧
  supportSingleImage: boolean     // 支持单图
  supportTextOnly: boolean        // 支持纯文本
  maxImages: number  // 最多支持几张图片
}

type AudioListItem = {
  id: string
  name: string
  url: string
  category?: string
  duration?: number
  view_count?: number
  artist?: string
  cover?: string
  description?: string
  tags?: Array<{ name: string }>
  source?: 'asset' | 'douyin' | 'netease' | 'sfx' | 'qq' | 'kugou' | 'migu' | 'baidu'
  sourceId?: string
  sourceSongUrl?: string
  sourceMid?: string
  sourceHash?: string
  sourceContentId?: string
  assetId?: number
  rank?: number
  updatedAt?: string
  isFavorite?: boolean
}

type HotMusicFallbackCursor = {
  keyword: string
  page: number
  exhausted: boolean
}


// 模型能力默认配置（作为后备）
const defaultModelCapabilities: Record<string, Omit<VideoModelCapability, 'id' | 'name'>> = {
  'kling': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 1
  },
  'runway': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: true,
    supportTextOnly: true,
    maxImages: 2
  },
  'pika': {
    supportSingleImage: true,
    supportMultipleImages: true,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 6
  },
  'doubao-seedance-1-5-pro-251215': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: true,
    supportTextOnly: true,
    maxImages: 2
  },
  'doubao-seedance-1-0-lite-i2v-250428': {
    supportSingleImage: true,
    supportMultipleImages: true,
    supportFirstLastFrame: true,
    supportTextOnly: false,
    maxImages: 6
  },
  'doubao-seedance-1-0-lite-t2v-250428': {
    supportSingleImage: false,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 0
  },
  'doubao-seedance-1-0-pro-250528': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: true,
    supportTextOnly: true,
    maxImages: 2
  },
  'doubao-seedance-1-0-pro-fast-251015': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 1
  },
  'sora-2': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 1
  },
  'sora-2-pro': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: true,
    supportTextOnly: true,
    maxImages: 2
  },
  'MiniMax-Hailuo-2.3': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 1
  },
  'MiniMax-Hailuo-2.3-Fast': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 1
  },
  'MiniMax-Hailuo-02': {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 1
  }
}

const getVideoModelCapability = (modelName: string): VideoModelCapability => {
  const capability = defaultModelCapabilities[modelName] || {
    supportSingleImage: true,
    supportMultipleImages: false,
    supportFirstLastFrame: false,
    supportTextOnly: true,
    maxImages: 1
  }

  return {
    id: modelName,
    name: modelName,
    ...capability
  }
}

// 从模型名称提取provider
const extractProviderFromModel = (modelName: string): string => {
  if (modelName.startsWith('doubao-') || modelName.startsWith('seedance')) {
    return 'doubao'
  }
  if (modelName.startsWith('runway')) {
    return 'runway'
  }
  if (modelName.startsWith('pika')) {
    return 'pika'
  }
  if (modelName.startsWith('MiniMax-') || modelName.toLowerCase().startsWith('minimax') || modelName.startsWith('hailuo')) {
    return 'minimax'
  }
  if (modelName.startsWith('sora')) {
    return 'openai'
  }
  if (modelName.startsWith('kling')) {
    return 'kling'
  }

  // 默认返回doubao
  return 'doubao'
}

// 加载视频AI配置
const loadVideoModels = async () => {
  try {
    const configs = await aiAPI.list('video')

    // 只显示启用的配置
    const activeConfigs = configs.filter(c => c.is_active)

    // 展开模型列表并去重
    const allModels = activeConfigs.flatMap(config => {
      const models = Array.isArray(config.model) ? config.model : [config.model]
      return models.map(modelName => ({
        modelName,
        configName: config.name,
        priority: config.priority || 0
      }))
    }).sort((a, b) => b.priority - a.priority)

    // 按模型名称去重
    const modelMap = new Map<string, { configName: string, priority: number }>()
    allModels.forEach(model => {
      if (!modelMap.has(model.modelName)) {
        modelMap.set(model.modelName, { configName: model.configName, priority: model.priority })
      }
    })

    // 构建模型能力列表
    videoModelCapabilities.value = Array.from(modelMap.keys()).map(getVideoModelCapability)

    // 默认优先选择豆包模型
    if (videoModelCapabilities.value.length > 0) {
      const hasCurrent = videoModelCapabilities.value.some(m => m.id === selectedVideoModel.value)
      if (!selectedVideoModel.value || !hasCurrent) {
        const preferredModel = videoModelCapabilities.value.find(m => {
          const id = m.id.toLowerCase()
          return id.startsWith('doubao') || id.startsWith('seedance') || id.includes('doubao') || id.includes('seedance')
        })
        selectedVideoModel.value = (preferredModel || videoModelCapabilities.value[0]).id
      }
    }
  } catch (error: any) {
    console.error('加载视频模型配置失败:', error)
    ElMessage.error('加载视频模型失败')
  }
}

// 加载视频素材库
const loadVideoAssets = async () => {
  try {
    const result = await assetAPI.listAssets({
      drama_id: dramaId.toString(),
      episode_id: episodeId.value ? Number(episodeId.value) : undefined,
      type: 'video',
      page: 1,
      page_size: 100
    })
    // 检查数据结构并正确赋值
    videoAssets.value = result.items || []
  } catch (error: any) {
    console.error('加载视频素材库失败:', error)
  }
}

// 加载音频素材库
const loadAudioAssets = async () => {
  loadingAudioAssets.value = true
  try {
    const result = await assetAPI.listAssets({
      drama_id: dramaId.toString(),
      episode_id: episodeId.value ? Number(episodeId.value) : undefined,
      type: 'audio',
      page: 1,
      page_size: 200
    })
    audioAssets.value = result.items || []
  } catch (error: any) {
    console.error('加载音频素材库失败:', error)
  } finally {
    loadingAudioAssets.value = false
  }
}

const DOUYIN_MUSIC_RAW_SOURCE = 'https://raw.githubusercontent.com/lonnyzhang423/douyin-hot-hub/main/README.md'
const DOUYIN_MUSIC_SOURCE_CANDIDATES = [
  `/api/v1/media/proxy?url=${encodeURIComponent(DOUYIN_MUSIC_RAW_SOURCE)}`,
  DOUYIN_MUSIC_RAW_SOURCE
]

const FRONTEND_FETCH_TIMEOUT_MS = 45000
const MUSIC_SEARCH_TIMEOUT_MS = 60000
const MUSIC_PREVIEW_PROBE_TIMEOUT_MS = 25000
const AUDIO_METADATA_TIMEOUT_MS = 30000
const SFX_FETCH_TIMEOUT_MS = 45000
const SEARCH_PREVIEW_FALLBACK_TIMEOUT_MS = 30000
const HOT_MUSIC_FALLBACK_TIMEOUT_MS = 45000
const DOUYIN_MUSIC_FETCH_TIMEOUT_MS = 45000
const HOT_MUSIC_MAX_ITEMS = 500
const HOT_MUSIC_FALLBACK_PAGE_SIZE = 50
const HOT_MUSIC_LAZY_BATCH_SIZE = 20
const HOT_MUSIC_FALLBACK_MAX_PAGES = 30
const SFX_LAZY_PAGE_SIZE = 20

const createTimeoutSignal = (timeoutMs: number, externalSignal?: AbortSignal) => {
  const timeoutController = new AbortController()
  const timeoutId = window.setTimeout(() => timeoutController.abort(), timeoutMs)

  const onExternalAbort = () => timeoutController.abort()
  if (externalSignal) {
    if (externalSignal.aborted) {
      timeoutController.abort()
    } else {
      externalSignal.addEventListener('abort', onExternalAbort, { once: true })
    }
  }

  const cleanup = () => {
    window.clearTimeout(timeoutId)
    if (externalSignal) {
      externalSignal.removeEventListener('abort', onExternalAbort)
    }
  }

  return { signal: timeoutController.signal, cleanup }
}

const fetchJsonWithTimeout = async (url: string, timeoutMs = FRONTEND_FETCH_TIMEOUT_MS, externalSignal?: AbortSignal) => {
  const { signal, cleanup } = createTimeoutSignal(timeoutMs, externalSignal)
  try {
    const response = await fetch(url, { cache: 'no-store', signal })
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    const json = await response.json()
    if (json && typeof json === 'object' && 'success' in json && 'data' in json) {
      return json.data
    }
    return json
  } catch (error: any) {
    if (error?.name === 'AbortError') {
      throw new Error(externalSignal?.aborted ? 'cancelled' : 'timeout')
    }
    throw error
  } finally {
    cleanup()
  }
}

const fetchTextWithTimeout = async (url: string, timeoutMs = FRONTEND_FETCH_TIMEOUT_MS, externalSignal?: AbortSignal) => {
  const { signal, cleanup } = createTimeoutSignal(timeoutMs, externalSignal)
  try {
    const response = await fetch(url, { cache: 'no-store', signal })
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    return response.text()
  } catch (error: any) {
    if (error?.name === 'AbortError') {
      throw new Error(externalSignal?.aborted ? 'cancelled' : 'timeout')
    }
    throw error
  } finally {
    cleanup()
  }
}

let neteaseSearchRequestId = 0
const neteaseSearchCache = new Map<string, { items: any[]; total: number }>()
const neteaseSearchTotalCache = new Map<string, number>()
const audioPreviewCache = new Map<string, string>()
const audioPreviewCandidatesCache = new Map<string, string[]>()
const audioPlayableUrlCache = new Map<string, string>()
const audioDurationCache = new Map<string, number>()
const audioDurationLoading = new Set<string>()
let previewRequestSeq = 0
const sfxCategories = ['热门', '转场', '笑声', '尴尬', '震惊']

const parseDurationToSeconds = (value: any) => {
  if (value === null || value === undefined) return undefined
  if (typeof value === 'number') {
    return value > 10000 ? Math.round(value / 1000) : value
  }
  const text = String(value).trim()
  if (!text) return undefined
  if (/^\d+$/.test(text)) return Number(text)
  const parts = text.split(':').map(p => Number(p))
  if (parts.some(Number.isNaN)) return undefined
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2]
  if (parts.length === 2) return parts[0] * 60 + parts[1]
  return undefined
}

const containsChinese = (text: string) => /[\u4E00-\u9FFF]/.test(text)

const normalizeSfxDisplayName = (name: string) => {
  return (name || '')
    .replace(/\.(wav|mp3|ogg|flac|m4a|aac)$/i, '')
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

const localizeSfxName = (rawName: string, category: string, rank: number) => {
  const cleaned = normalizeSfxDisplayName(rawName)
  if (!cleaned) return `${category}音效 ${rank}`
  if (containsChinese(cleaned)) return cleaned

  const exactMap: Record<string, string> = {
    'cash register': '收银机提示音',
    'cash register purchase': '收银成交提示音',
    'cha ching': '到账提示音',
    'till with bell': '收银铃声',
    'cash register fake': '收银机模拟提示音',
    'whoosh': '转场呼啸音效',
    'laugh': '欢笑音效',
    'record scratch': '尴尬打断音效',
    'dramatic hit': '震惊重击音效'
  }

  const normalized = cleaned.toLowerCase()
  if (exactMap[normalized]) return exactMap[normalized]

  const replacementRules: Array<{ pattern: RegExp, replacement: string }> = [
    { pattern: /\bcash\s*register\b/gi, replacement: '收银机' },
    { pattern: /\bregister\b/gi, replacement: '收银' },
    { pattern: /\bpurchase\b/gi, replacement: '成交' },
    { pattern: /\bcheckout\b/gi, replacement: '结账' },
    { pattern: /\bcha\s*ching\b/gi, replacement: '到账提示' },
    { pattern: /\btill\b/gi, replacement: '收银台' },
    { pattern: /\bbell\b/gi, replacement: '铃声' },
    { pattern: /\bcoin\b/gi, replacement: '硬币' },
    { pattern: /\bding\b/gi, replacement: '提示' },
    { pattern: /\bbeep\b/gi, replacement: '提示音' },
    { pattern: /\bclick\b/gi, replacement: '点击声' },
    { pattern: /\bwhoosh\b/gi, replacement: '转场呼啸' },
    { pattern: /\blaugh(ter)?\b/gi, replacement: '欢笑' },
    { pattern: /\brecord\s*scratch\b/gi, replacement: '尴尬打断' },
    { pattern: /\bawkward\b/gi, replacement: '尴尬' },
    { pattern: /\bdramatic\b/gi, replacement: '震撼' },
    { pattern: /\bhit\b/gi, replacement: '重击' },
    { pattern: /\bshock\b/gi, replacement: '震惊' },
    { pattern: /\bsurprise\b/gi, replacement: '惊讶' },
    { pattern: /\bfake\b/gi, replacement: '模拟' },
    { pattern: /\bsound\s*effect(s)?\b/gi, replacement: '音效' },
    { pattern: /\bsfx\b/gi, replacement: '音效' }
  ]

  let localized = normalized
  replacementRules.forEach(({ pattern, replacement }) => {
    localized = localized.replace(pattern, replacement)
  })

  localized = localized
    .replace(/\b(with|and|the|a|an|of|for|to|in|on)\b/gi, ' ')
    .replace(/[^0-9\u4E00-\u9FFF\s]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  if (containsChinese(localized)) {
    return localized.replace(/\s+/g, '')
  }

  return `${category}音效 ${rank}`
}

const formatDuration = (seconds?: number) => {
  if (!seconds || seconds <= 0) return ''
  const s = Math.round(seconds)
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${sec.toString().padStart(2, '0')}`
}

const fixMediaUrl = (url?: string | null) => {
  if (!url) return ''
  if (url.startsWith('blob:')) return url
  if (url.startsWith('data:')) return url
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  // If backend stores absolute filesystem-like paths, map them to /static
  if (url.startsWith('/static/')) return url
  if (url.startsWith('/data/')) return `/static${url}`
  if (url.startsWith('data/')) return `/static/${url}`
  if (url.startsWith('/')) return url
  return `/static/${url}`
}

const getApiBase = () => {
  const raw = (import.meta.env.VITE_API_BASE_URL as string | undefined) || window.location.origin
  return raw.replace(/\/$/, '')
}

const isTunnelStaticUrl = (url: URL) => {
  return (
    url.pathname.startsWith('/static/') &&
    (
      url.hostname.endsWith('.loca.lt') ||
      url.hostname.includes('ngrok') ||
      url.hostname.endsWith('.trycloudflare.com')
    )
  )
}

const toBackendMediaUrl = (url?: string | null) => {
  if (!url) return ''
  if (url.startsWith('blob:') || url.startsWith('data:')) return url
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  if (url.startsWith('/')) return `${getApiBase()}${url}`
  return `${getApiBase()}/${url}`
}

const resolveVideoUrl = (url?: string | null) => {
  if (!url) return ''
  const fixed = fixMediaUrl(url)
  if (!fixed) return ''
  if (fixed.startsWith('blob:') || fixed.startsWith('data:')) return fixed
  if (fixed.startsWith('/api/v1/media/proxy')) return fixed
  if (fixed.startsWith('http://') || fixed.startsWith('https://')) {
    try {
      const parsed = new URL(fixed)
      if (parsed.pathname === '/api/v1/media/proxy') {
        return `${parsed.pathname}${parsed.search}`
      }
      if (isTunnelStaticUrl(parsed)) {
        return toBackendMediaUrl(`${parsed.pathname}${parsed.search}`)
      }
    } catch {
      // noop
    }
    return `/api/v1/media/proxy?url=${encodeURIComponent(fixed)}`
  }
  return toBackendMediaUrl(fixed)
}

const resolveAudioUrl = (url?: string | null) => {
  if (!url) return ''
  if (url.startsWith('data:') || url.startsWith('blob:')) return url
  if (url.startsWith('/api/v1/music/stream') || url.startsWith('/api/v1/music/netease/stream')) return url
  if (url.startsWith('/api/v1/media/proxy')) return url
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return `/api/v1/media/proxy?url=${encodeURIComponent(url)}`
  }
  return fixMediaUrl(url)
}

const resolveListPayload = (data: any) => {
  const candidates = [data, data?.data, data?.data?.data, data?.result]
  for (const candidate of candidates) {
    if (candidate && Array.isArray(candidate.items)) {
      return candidate
    }
    if (candidate && Array.isArray(candidate.songs)) {
      return { items: candidate.songs, total: candidate.songCount || candidate.total || candidate.songs.length }
    }
    if (candidate?.result && Array.isArray(candidate.result.songs)) {
      return { items: candidate.result.songs, total: candidate.result.songCount || candidate.result.total || candidate.result.songs.length }
    }
  }
  if (Array.isArray(data)) {
    return { items: data, total: data.length }
  }
  return { items: [], total: 0 }
}

const resolveSongSource = (song: any) => {
  const raw = `${song?.source || song?.platform || ''}`.toLowerCase()
  if (raw) return raw
  if (song?.mid || song?.songmid || song?.songMid) return 'qq'
  if (song?.hash) return 'kugou'
  if (song?.content_id || song?.contentId) return 'migu'
  const url = song?.song_url || song?.url || ''
  if (typeof url === 'string') {
    if (url.includes('qqmusic') || url.includes('qq.com')) return 'qq'
    if (url.includes('kugou')) return 'kugou'
    if (url.includes('migu')) return 'migu'
    if (url.includes('music.163') || url.includes('music.126')) return 'netease'
  }
  return 'netease'
}

const normalizeSongText = (text?: string) => {
  if (!text) return ''
  return text
    .toLowerCase()
    .replace(/\([^)]*\)|（[^）]*）|\[[^\]]*]|【[^】]*】/g, '')
    .replace(/[^\p{L}\p{N}]+/gu, '')
}

const scoreSongMatch = (song: any, targetName: string, targetArtist: string) => {
  const songName = normalizeSongText(song?.title || song?.name || '')
  const songArtist = normalizeSongText(song?.artist || '')
  const expectedName = normalizeSongText(targetName)
  const expectedArtist = normalizeSongText(targetArtist)
  let score = 0

  if (songName && expectedName) {
    if (songName === expectedName) score += 120
    else if (songName.includes(expectedName) || expectedName.includes(songName)) score += 80
    else {
      const overlap = Array.from(new Set(expectedName.split(''))).filter(ch => songName.includes(ch)).length
      score += overlap
    }
  }
  if (songArtist && expectedArtist) {
    if (songArtist === expectedArtist) score += 60
    else if (songArtist.includes(expectedArtist) || expectedArtist.includes(songArtist)) score += 40
  }

  const source = resolveSongSource(song)
  if (source !== 'netease') score += 10
  if (song?.song_url) score += 6
  return score
}

const searchNeteaseSongs = async (keywords: string) => {
  const query = keywords.trim()
  if (!query || audioMode.value !== 'music') {
    neteaseSearchResults.value = []
    neteaseSearchTotal.value = 0
    if (neteaseSearchAbortController) {
      neteaseSearchAbortController.abort()
      neteaseSearchAbortController = null
    }
    return
  }

  if (neteaseSearchAbortController) {
    neteaseSearchAbortController.abort()
  }
  const abortController = new AbortController()
  neteaseSearchAbortController = abortController

  const currentRequestId = ++neteaseSearchRequestId
  loadingNeteaseSearch.value = true
  neteaseSearchError.value = null
  try {
    const pageLimit = audioSearchPageSize.value
    const backendUrl = `/api/v1/music/search?keywords=${encodeURIComponent(query)}&page=${audioSearchPage.value}&page_size=${pageLimit}`

    const cacheKey = `${query}::${audioSearchPage.value}::${pageLimit}`
    if (neteaseSearchCache.has(cacheKey)) {
      const cached = neteaseSearchCache.get(cacheKey)!
      neteaseSearchResults.value = cached.items
      if (neteaseSearchTotalCache.has(query)) {
        neteaseSearchTotal.value = neteaseSearchTotalCache.get(query) || cached.total
      } else {
        neteaseSearchTotal.value = cached.total
        if (cached.total) {
          neteaseSearchTotalCache.set(query, cached.total)
        }
      }
      void fillMissingDurations(neteaseSearchResults.value)
      return
    }

    const backendData = await fetchJsonWithTimeout(backendUrl, MUSIC_SEARCH_TIMEOUT_MS, abortController.signal)
    const { items: backendItems, total: backendTotal } = resolveListPayload(backendData)

    if (currentRequestId !== neteaseSearchRequestId) return
    const totalCacheKey = query
    if (audioSearchPage.value === 1 || !neteaseSearchTotalCache.has(totalCacheKey)) {
      const resolvedTotal = backendTotal || backendItems.length
      if (resolvedTotal) {
        neteaseSearchTotalCache.set(totalCacheKey, resolvedTotal)
      }
    }
    neteaseSearchTotal.value = neteaseSearchTotalCache.get(totalCacheKey) || backendTotal || backendItems.length
    const mappedResults = (backendItems || []).map((song: any) => {
      const source = resolveSongSource(song)
      const platformMap: Record<string, string> = {
        netease: '网易云音乐',
        qq: 'QQ音乐',
        kugou: '酷狗音乐',
        migu: '咪咕音乐',
        baidu: '百度音乐'
      }
      const platform = platformMap[source] || song.platform || '网易云音乐'
      const artist = song.artist || (song.ar ? (song.ar || []).map((a: any) => a.name).filter(Boolean).join('/') : '')
      const album = song.album || song.al?.name || ''
      const cover = song.cover || song.cover_url || song.al?.picUrl || song.album?.picUrl || ''
      const meta = [artist, album].filter(Boolean).join(' · ')
      const streamUrl = buildMusicStreamUrl(song, song.title || '')
      const rawDuration = song.duration
        ?? song.dt
        ?? song.interval
        ?? song.time
        ?? song.timelength
        ?? song.timeLength
        ?? song.time_length
        ?? song.length
        ?? song.length_ms
        ?? song.play_time
        ?? song.playTime
      return {
        id: `${source || 'music'}-${song.id || song.mid || song.hash || Math.random()}`,
        sourceId: song.id ? String(song.id) : (song.song_id ? String(song.song_id) : (song.songId ? String(song.songId) : undefined)),
        name: song.title || song.name || '',
        url: streamUrl,
        category: platform,
        duration: parseDurationToSeconds(rawDuration),
        view_count: song.view_count || 0,
        artist: artist || undefined,
        cover: cover || undefined,
        description: meta || undefined,
        tags: [{ name: platform }],
        source: (source || 'netease') as AudioListItem['source'],
        sourceSongUrl: song.song_url || undefined,
        sourceMid: song.mid ? String(song.mid) : (song.songmid ? String(song.songmid) : (song.songMid ? String(song.songMid) : undefined)),
        sourceHash: song.hash ? String(song.hash) : (song.fileHash ? String(song.fileHash) : undefined),
        sourceContentId: song.content_id ? String(song.content_id) : (song.contentId ? String(song.contentId) : undefined)
      }
    })
    mappedResults.sort((a, b) => {
      const scoreA = scoreSongMatch(
        { title: a.name, artist: a.artist, source: a.source, song_url: a.sourceSongUrl },
        query,
        ''
      )
      const scoreB = scoreSongMatch(
        { title: b.name, artist: b.artist, source: b.source, song_url: b.sourceSongUrl },
        query,
        ''
      )
      return scoreB - scoreA
    })
    neteaseSearchResults.value = mappedResults
    void fillMissingDurations(neteaseSearchResults.value)
    neteaseSearchCache.set(cacheKey, { items: neteaseSearchResults.value, total: neteaseSearchTotal.value })
  } catch (error: any) {
    if (currentRequestId !== neteaseSearchRequestId) return
    if (error?.message === 'cancelled') return
    console.error('网易云搜索失败:', error)
    neteaseSearchError.value = error?.message || '搜索失败'
    neteaseSearchResults.value = []
    neteaseSearchTotal.value = 0
  } finally {
    if (neteaseSearchAbortController === abortController) {
      neteaseSearchAbortController = null
    }
    if (currentRequestId === neteaseSearchRequestId) {
      loadingNeteaseSearch.value = false
    }
  }
}

const mapSfxItems = (items: any[], fallbackCategory: string, rankOffset = 0): AudioListItem[] => {
  const now = Date.now()
  const mappedRaw: Array<AudioListItem | null> = items
    .map((item: any, index: number) => {
      const sourceName = `${item.source || ''}`.trim().toLowerCase()
      const sourceLabel = sourceName === 'pixabay' ? 'Pixabay' : (sourceName === 'freesound' ? 'Freesound' : '')
      const category = item.category || fallbackCategory || '热门音效'
      const rawRank = Number(item.rank)
      const rank = rankOffset > 0
        ? (rankOffset + index + 1)
        : ((Number.isFinite(rawRank) && rawRank > 0) ? rawRank : (index + 1))
      const url = resolveAudioUrl(item.url || item.audio_url || item.file_url || item.file_path || item.preview_url)
      if (!url) {
        return null
      }

      const tags = [category]
      if (sourceLabel) {
        tags.push(sourceLabel)
      }

      return {
        id: item.id ? String(item.id) : `sfx-${sourceName || 'mix'}-${category}-${index}-${now}`,
        name: localizeSfxName(item.name || item.title || '', category, rank),
        url,
        category,
        duration: parseDurationToSeconds(item.duration ?? item.length),
        view_count: Number(item.view_count || item.downloads || item.plays || 0),
        artist: item.artist || item.user || '',
        cover: item.cover || item.image || '',
        description: item.description || (sourceLabel ? `来源：${sourceLabel}` : ''),
        tags: tags.map((name) => ({ name })),
        source: 'sfx' as const,
        rank
      } as AudioListItem
    })

  const mapped = mappedRaw.filter((asset): asset is AudioListItem => asset !== null)

  void fillMissingDurations(mapped)
  return mapped
}

const loadSfx = async (append = false) => {
  const query = audioSearch.value.trim()
  const isAppendMode = append && !query
  if (isAppendMode) {
    if (sfxLoadingMore.value || !sfxHasMore.value) return
    sfxLoadingMore.value = true
  } else {
    loadingSfx.value = true
    sfxPage.value = 1
    sfxHasMore.value = true
  }

  const targetPage = isAppendMode ? (sfxPage.value + 1) : 1
  try {
    const params = new URLSearchParams({ limit: String(SFX_LAZY_PAGE_SIZE), page: String(targetPage) })
    if (query) {
      params.set('keywords', query)
    } else {
      params.set('category', sfxCategory.value)
    }

    const data = await fetchJsonWithTimeout(`/api/v1/sfx?${params.toString()}`, SFX_FETCH_TIMEOUT_MS)
    const { items } = resolveListPayload(data)
    const mapped = mapSfxItems(items, sfxCategory.value, isAppendMode ? sfxAssets.value.length : 0)
    const hasMoreFromResponse = typeof data?.has_more === 'boolean'
      ? data.has_more
      : mapped.length >= SFX_LAZY_PAGE_SIZE

    if (isAppendMode) {
      const existingKeys = new Set(
        sfxAssets.value.map((asset) => `${asset.source || ''}|${asset.id}|${asset.url || ''}`)
      )
      const appended = mapped.filter((asset) => {
        const key = `${asset.source || ''}|${asset.id}|${asset.url || ''}`
        if (existingKeys.has(key)) return false
        existingKeys.add(key)
        return true
      })
      sfxAssets.value = [...sfxAssets.value, ...appended]
      sfxPage.value = targetPage
      sfxHasMore.value = hasMoreFromResponse && mapped.length > 0
    } else {
      sfxAssets.value = mapped
      sfxPage.value = targetPage
      sfxHasMore.value = hasMoreFromResponse && mapped.length > 0
    }

    const warnings = Array.isArray(data?.warnings) ? data.warnings : []
    if (warnings.length > 0) {
      console.warn('音效部分数据源不可用:', warnings)
    }
  } catch (error: any) {
    console.error('加载音效失败:', error)
    if (!isAppendMode) {
      sfxAssets.value = []
    }
    sfxHasMore.value = false
  } finally {
    if (isAppendMode) {
      sfxLoadingMore.value = false
    } else {
      loadingSfx.value = false
    }
  }
}

const buildNeteaseStreamUrl = (id: string) => {
  return `/api/v1/music/netease/stream?id=${encodeURIComponent(id)}`
}

const buildResolverStreamUrl = (song: any, fallbackTitle?: string) => {
  const source = resolveSongSource(song)
  if (!source || source === 'netease') return ''
  return `/api/v1/music/stream?source=${encodeURIComponent(source)}&id=${encodeURIComponent(song.id || '')}&mid=${encodeURIComponent(song.mid || '')}&hash=${encodeURIComponent(song.hash || '')}&content_id=${encodeURIComponent(song.content_id || '')}&title=${encodeURIComponent(song.title || fallbackTitle || '')}&artist=${encodeURIComponent(song.artist || '')}`
}

const buildMusicStreamUrl = (song: any, fallbackTitle?: string) => {
  if (!song) return ''
  const source = resolveSongSource(song)
  if (source === 'netease') {
    return buildNeteaseStreamUrl(String(song.id || ''))
  }
  const resolverUrl = buildResolverStreamUrl(song, fallbackTitle)
  if (song.song_url) {
    return `/api/v1/music/stream?url=${encodeURIComponent(song.song_url)}`
  }
  if (resolverUrl) {
    return resolverUrl
  }
  return ''
}

const resolveSearchPreviewCandidates = async (asset: AudioListItem) => {
  const cacheKey = asset.id || asset.name || ''
  if (cacheKey && audioPreviewCandidatesCache.has(cacheKey)) {
    return audioPreviewCandidatesCache.get(cacheKey) || []
  }
  const keyword = `${asset.name || ''} ${asset.artist || ''}`.trim()
  if (!keyword) return []
  const searchUrl = `/api/v1/music/search?keywords=${encodeURIComponent(keyword)}&page=1&page_size=20`
  const data = await fetchJsonWithTimeout(searchUrl, SEARCH_PREVIEW_FALLBACK_TIMEOUT_MS)
  const { items } = resolveListPayload(data)
  if (!items.length) return []

  const sorted = [...items].sort((a: any, b: any) => {
    const scoreA = scoreSongMatch(a, asset.name || '', asset.artist || '')
      + (a.song_url ? 40 : 0)
      + (resolveSongSource(a) !== 'netease' ? 15 : 0)
    const scoreB = scoreSongMatch(b, asset.name || '', asset.artist || '')
      + (b.song_url ? 40 : 0)
      + (resolveSongSource(b) !== 'netease' ? 15 : 0)
    return scoreB - scoreA
  })
  const candidates: string[] = []
  sorted.slice(0, 20).forEach((song: any) => {
    if (song.song_url) {
      candidates.push(`/api/v1/music/stream?url=${encodeURIComponent(song.song_url)}`)
    }
    const primary = buildMusicStreamUrl(song, keyword)
    if (primary) candidates.push(primary)
    if (resolveSongSource(song) === 'netease' && song.id) {
      candidates.push(buildNeteaseStreamUrl(String(song.id)))
    }
  })
  const deduped = [...new Set(candidates.filter(Boolean))]
  if (cacheKey && deduped.length > 0) {
    audioPreviewCandidatesCache.set(cacheKey, deduped)
  }
  return deduped
}

const resolveSearchPreviewUrl = async (asset: AudioListItem) => {
  const cacheKey = asset.id || asset.name || ''
  if (cacheKey && audioPreviewCache.has(cacheKey)) {
    return audioPreviewCache.get(cacheKey) || ''
  }
  const candidates = await resolveSearchPreviewCandidates(asset)
  return candidates[0] || ''
}

const parseDouyinMusic = (content: string) => {
  const lines = content.split('\n')
  const items: AudioListItem[] = []
  let currentListName = ''
  let inMusicSection = false
  let index = 0
  let updatedAt: string | null = null
  const seenUrls = new Set<string>()

  const withPrefix = /^\s*\d+\.\s*([^:]+):\s*\[([^\]]+)\]\(([^)]+)\)\s*\|\s*([0-9.]+)\s*\|\s*([0-9-:\s]+)/
  const withoutPrefix = /^\s*\d+\.\s*\[([^\]]+)\]\(([^)]+)\)\s*\|\s*([0-9.]+)\s*\|\s*([0-9-:\s]+)/
  const markdownList = /^\s*\d+\.\s*\[([^\]]+)\]\(([^)]+)\)\s*(?:-\s*(.+))?/

  lines.forEach((line) => {
    const updateMatch = line.match(/更新时间：\s*([0-9-:\s+]+)\s*/i)
    if (updateMatch) {
      updatedAt = updateMatch[1].trim()
    }

    const sectionMatch = line.match(/^##\s+(.*)/)
    if (sectionMatch) {
      const sectionTitle = sectionMatch[1].trim()
      if (sectionTitle.includes('音乐榜')) {
        inMusicSection = true
        currentListName = sectionTitle
      } else {
        if (inMusicSection) {
          inMusicSection = false
        }
        currentListName = ''
      }
      return
    }

    if (!inMusicSection) return
    if (/暂无数据/.test(line)) return

    let match = line.match(withPrefix)
    let listName = ''
    let title = ''
    let url = ''
    let hotValue = ''
    let dateValue = ''
    let artist = ''

    if (match) {
      listName = match[1].trim()
      title = match[2].trim()
      url = match[3].trim()
      hotValue = match[4].trim()
      dateValue = match[5].trim()
    } else {
      match = line.match(withoutPrefix)
      if (match) {
        listName = currentListName
        title = match[1].trim()
        url = match[2].trim()
        hotValue = match[3].trim()
        dateValue = match[4].trim()
      } else {
        match = line.match(markdownList)
        if (match) {
          listName = currentListName || '抖音音乐榜'
          title = match[1].trim()
          url = match[2].trim()
          artist = (match[3] || '').trim()
        }
      }
    }

    if (!match || !title || !url) return
    if (seenUrls.has(url)) return
    seenUrls.add(url)

    const hotNumber = Number(hotValue.replace(/[^\d.]/g, '')) || 0
    const listLabel = listName || '抖音音乐榜'
    const tags = [{ name: listLabel }, { name: '抖音音乐榜' }]
    if (artist) tags.push({ name: artist })

    const item: AudioListItem = {
      id: `douyin-${index}-${url}`,
      name: title,
      url,
      category: listLabel,
      view_count: hotNumber || Math.max(0, 1000 - index),
      artist: artist || undefined,
      description: artist ? `歌手: ${artist}` : (dateValue ? `更新: ${dateValue}` : undefined),
      tags,
      source: 'douyin',
      rank: index + 1,
      updatedAt: dateValue
    }

    items.push(item)
    index += 1
  })

  return { items, updatedAt }
}

const HOT_MUSIC_FALLBACK_KEYWORDS = ['抖音热歌', '热门歌曲', '热歌榜']

const buildHotMusicDedupeKey = (asset: AudioListItem) => {
  return [
    asset.source || 'unknown',
    asset.sourceId || '',
    asset.sourceMid || '',
    asset.sourceHash || '',
    asset.sourceContentId || '',
    normalizeSongText(asset.name || ''),
    normalizeSongText(asset.artist || '')
  ].join('|')
}

const resetHotMusicFallbackCursors = () => {
  hotMusicFallbackCursors = HOT_MUSIC_FALLBACK_KEYWORDS.map((keyword) => ({
    keyword,
    page: 1,
    exhausted: false
  }))
  hotMusicFallbackCursorIndex = 0
}

const hasMoreHotMusicFallbackPages = () => {
  return hotMusicFallbackCursors.some(cursor => !cursor.exhausted && cursor.page <= HOT_MUSIC_FALLBACK_MAX_PAGES)
}

const pickNextHotMusicFallbackCursor = (): HotMusicFallbackCursor | null => {
  if (!hasMoreHotMusicFallbackPages()) return null
  if (hotMusicFallbackCursors.length === 0) return null

  for (let i = 0; i < hotMusicFallbackCursors.length; i += 1) {
    const idx = (hotMusicFallbackCursorIndex + i) % hotMusicFallbackCursors.length
    const cursor = hotMusicFallbackCursors[idx]
    if (!cursor.exhausted && cursor.page <= HOT_MUSIC_FALLBACK_MAX_PAGES) {
      hotMusicFallbackCursorIndex = (idx + 1) % hotMusicFallbackCursors.length
      return cursor
    }
  }
  return null
}

const toFallbackHotMusicItem = (song: any, rank: number): AudioListItem | null => {
  const source = resolveSongSource(song)
  const name = song.title || song.name || ''
  const artist = song.artist || ''
  const sourceId = song.id ? String(song.id) : (song.song_id ? String(song.song_id) : (song.songId ? String(song.songId) : ''))
  const sourceMid = song.mid ? String(song.mid) : (song.songmid ? String(song.songmid) : (song.songMid ? String(song.songMid) : ''))
  const sourceHash = song.hash ? String(song.hash) : (song.fileHash ? String(song.fileHash) : '')
  const sourceContentID = song.content_id ? String(song.content_id) : (song.contentId ? String(song.contentId) : '')
  const streamUrl = buildMusicStreamUrl(song, name)
  if (!name || !streamUrl) return null

  return {
    id: `fallback-hot-${source || 'netease'}-${sourceId || sourceMid || sourceHash || rank}`,
    name,
    url: streamUrl,
    category: '热门配乐',
    duration: parseDurationToSeconds(song.duration),
    view_count: Math.max(0, 1000 - rank + 1),
    artist: artist || undefined,
    description: '抖音热门推荐',
    tags: [{ name: '热门配乐' }, { name: '抖音音乐榜' }],
    source: (source || 'netease') as AudioListItem['source'],
    sourceId: sourceId || undefined,
    sourceSongUrl: song.song_url || undefined,
    sourceMid: sourceMid || undefined,
    sourceHash: sourceHash || undefined,
    sourceContentId: sourceContentID || undefined,
    rank
  }
}

const loadFallbackHotMusic = async (seedItems: AudioListItem[] = [], requestCount = HOT_MUSIC_FALLBACK_KEYWORDS.length) => {
  if (hotMusicFallbackLoading.value) return
  if (hotMusicFallbackCursors.length === 0) {
    resetHotMusicFallbackCursors()
  }

  hotMusicFallbackLoading.value = true
  try {
    const merged = new Map<string, AudioListItem>()
    let rank = 0
    const baseItems = seedItems.length > 0 ? seedItems : douyinMusicAssets.value

    baseItems.forEach((asset) => {
      if (!asset?.name || !asset?.url) return
      const key = buildHotMusicDedupeKey(asset)
      if (merged.has(key)) return
      rank += 1
      merged.set(key, {
        ...asset,
        rank: asset.rank && asset.rank > 0 ? asset.rank : rank
      })
    })

    let requestsUsed = 0
    while (requestsUsed < requestCount && merged.size < HOT_MUSIC_MAX_ITEMS) {
      const cursor = pickNextHotMusicFallbackCursor()
      if (!cursor) break
      requestsUsed += 1

      try {
        const data = await fetchJsonWithTimeout(
          `/api/v1/music/search?keywords=${encodeURIComponent(cursor.keyword)}&page=${cursor.page}&page_size=${HOT_MUSIC_FALLBACK_PAGE_SIZE}`,
          HOT_MUSIC_FALLBACK_TIMEOUT_MS
        )
        const { items, total } = resolveListPayload(data)
        if (!items.length) {
          cursor.exhausted = true
          continue
        }

        items.forEach((song: any) => {
          if (merged.size >= HOT_MUSIC_MAX_ITEMS) return
          const item = toFallbackHotMusicItem(song, rank + 1)
          if (!item) return
          const key = buildHotMusicDedupeKey(item)
          if (merged.has(key)) return
          rank += 1
          item.rank = rank
          item.view_count = Math.max(0, 1000 - rank + 1)
          merged.set(key, item)
        })

        cursor.page += 1
        const reachedTotal = Number(total) > 0 && ((cursor.page - 1) * HOT_MUSIC_FALLBACK_PAGE_SIZE >= Number(total))
        const reachedSizeEnd = items.length < HOT_MUSIC_FALLBACK_PAGE_SIZE
        const reachedPageLimit = cursor.page > HOT_MUSIC_FALLBACK_MAX_PAGES
        if (reachedTotal || reachedSizeEnd || reachedPageLimit) {
          cursor.exhausted = true
        }
      } catch (error) {
        cursor.exhausted = true
        console.warn('加载热门配乐兜底失败:', cursor.keyword, error)
      }
    }

    douyinMusicAssets.value = Array.from(merged.values()).slice(0, HOT_MUSIC_MAX_ITEMS)
  } finally {
    hotMusicFallbackLoading.value = false
  }
}

const loadDouyinMusic = async () => {
  loadingDouyinMusic.value = true
  resetHotMusicFallbackCursors()
  try {
    let parsedItems: AudioListItem[] = []
    let parsedUpdatedAt: string | null = null
    let loaded = false

    for (const sourceUrl of DOUYIN_MUSIC_SOURCE_CANDIDATES) {
      try {
        const content = await fetchTextWithTimeout(sourceUrl, DOUYIN_MUSIC_FETCH_TIMEOUT_MS)
        const parsed = parseDouyinMusic(content)
        parsedItems = parsed.items
        parsedUpdatedAt = parsed.updatedAt || null
        loaded = true
        if (parsed.items.length > 0) {
          break
        }
      } catch (error) {
        console.warn('加载抖音音乐榜源失败:', sourceUrl, error)
      }
    }

    douyinMusicAssets.value = parsedItems
    if (parsedUpdatedAt) {
      douyinMusicUpdatedAt.value = parsedUpdatedAt
    } else if (parsedItems.length > 0) {
      const latest = parsedItems.reduce((max, item) => {
        if (!item.updatedAt) return max
        return item.updatedAt > max ? item.updatedAt : max
      }, '')
      douyinMusicUpdatedAt.value = latest || null
    }

    if (!loaded || parsedItems.length === 0) {
      await loadFallbackHotMusic()
    } else if (parsedItems.length < HOT_MUSIC_MAX_ITEMS) {
      await loadFallbackHotMusic(parsedItems)
    }
  } catch (error) {
    console.error('加载抖音音乐榜失败:', error)
    await loadFallbackHotMusic()
  } finally {
    loadingDouyinMusic.value = false
  }
}

const getAudioHotScore = (asset: AudioListItem) => {
  if (asset.source === 'sfx' && asset.rank && asset.rank > 0) {
    return 1_000_000 - asset.rank
  }
  if (asset.view_count && asset.view_count > 0) {
    return asset.view_count + (asset.isFavorite ? 1000 : 0)
  }
  if (asset.rank) {
    return 1000 - asset.rank
  }
  return asset.isFavorite ? 1000 : 0
}

const getAudioGroup = (asset: AudioListItem) => {
  const category = asset.category || ''
  const tagText = (asset.tags || []).map(t => t.name).join(' ')
  if (/音效|配音/.test(category) || /音效|配音/.test(tagText)) return 'sfx'
  if (/音乐|配乐|片头|片尾/.test(category) || /音乐|配乐|片头|片尾/.test(tagText)) return 'music'
  return 'music'
}

const isDouyinHot = (asset: AudioListItem) => {
  if (asset.source === 'douyin') return true
  const text = `${asset.name || ''} ${asset.description || ''} ${asset.category || ''} ${(asset.tags || []).map(t => t.name).join(' ')}`.toLowerCase()
  return text.includes('抖音') || text.includes('douyin') || text.includes('tiktok')
}

const audioList = computed<AudioListItem[]>(() => {
  if (audioMode.value === 'sfx') {
    return sfxAssets.value
  }
  const localItems = audioAssets.value.map((asset) => ({
    id: `asset-${asset.id}`,
    name: asset.name,
    url: asset.url,
    category: asset.category,
    duration: asset.duration,
    view_count: asset.view_count,
    description: asset.description,
    tags: (asset.tags || []).map(tag => ({ name: tag.name })),
    source: 'asset' as const,
    assetId: asset.id,
    isFavorite: asset.is_favorite
  }))

  const neteaseItems = audioSearch.value.trim() ? neteaseSearchResults.value : []
  return [...localItems, ...douyinMusicAssets.value, ...neteaseItems]
})

const audioCategoryOptions = computed(() => {
  if (audioMode.value === 'sfx') {
    return []
  }
  const categories = new Set<string>()
  audioList.value.forEach(asset => {
    if (getAudioGroup(asset) !== audioMode.value) return
    if (asset.category) categories.add(asset.category)
  })
  return [
    { label: $t('video.soundMusicAll'), value: 'all' },
    ...Array.from(categories).sort().map(category => ({ label: category, value: category }))
  ]
})

const audioSearchLoading = computed(() => {
  if (!audioSearch.value.trim()) {
    return false
  }
  return audioMode.value === 'sfx' ? loadingSfx.value : loadingNeteaseSearch.value
})

const audioListLoading = computed(() => {
  if (audioMode.value === 'sfx') {
    return loadingSfx.value
  }
  if (audioSearch.value.trim()) {
    return false
  }
  return loadingAudioAssets.value || loadingDouyinMusic.value
})

const showAudioListOverlay = computed(() => {
  if (audioSearch.value.trim()) {
    return false
  }
  return audioListLoading.value
})

const filteredAudioAssets = computed(() => {
  const query = audioSearch.value.trim().toLowerCase()
  if (audioMode.value === 'music' && query) {
    return neteaseSearchResults.value
  }
  const assets = audioList.value
    .filter(asset => asset.url || asset.source === 'netease')
    .filter(asset => audioMode.value === 'sfx' || getAudioGroup(asset) === audioMode.value)
    .filter(asset => audioMode.value !== 'sfx' ? (audioCategory.value === 'all' || asset.category === audioCategory.value) : true)
    .filter(asset => {
      if (!query) return true
      const tagText = (asset.tags || []).map(t => t.name).join(' ')
      const haystack = `${asset.name || ''} ${asset.description || ''} ${asset.category || ''} ${tagText}`.toLowerCase()
      return haystack.includes(query)
    })
    .sort((a, b) => getAudioHotScore(b) - getAudioHotScore(a))

  if (audioMode.value === 'music' && !query && audioHotOnly.value) {
    const douyinHot = assets.filter(asset => isDouyinHot(asset))
    if (douyinHot.length > 0) {
      return douyinHot.slice(0, HOT_MUSIC_MAX_ITEMS)
    }
    return assets.slice(0, HOT_MUSIC_MAX_ITEMS)
  }

  return assets
})

const isHotMusicLazyMode = computed(() => {
  return audioMode.value === 'music' && !audioSearch.value.trim()
})

const isSfxLazyMode = computed(() => {
  return audioMode.value === 'sfx' && !audioSearch.value.trim()
})

const isAudioLazyMode = computed(() => {
  return isHotMusicLazyMode.value || isSfxLazyMode.value
})

const isAudioScrollLoading = computed(() => {
  return isAudioLazyMode.value && (hotMusicScrollLoading.value || hotMusicFallbackLoading.value || sfxLoadingMore.value)
})

const showAudioLazyTip = computed(() => {
  if (!isAudioLazyMode.value || isAudioScrollLoading.value) return false
  if (isSfxLazyMode.value) {
    return sfxHasMore.value
  }
  return pagedAudioAssets.value.length < filteredAudioAssets.value.length || canLoadMoreHotMusicFromSource()
})

const pagedAudioAssets = computed(() => {
  const assets = filteredAudioAssets.value
  if (isHotMusicLazyMode.value) {
    return assets.slice(0, Math.min(hotMusicVisibleCount.value, assets.length))
  }
  return assets
})

const getScrollContainer = (el: HTMLElement | null): HTMLElement | null => {
  let current = el?.parentElement || null
  while (current) {
    const style = window.getComputedStyle(current)
    const overflowY = style.overflowY
    const canScroll = (overflowY === 'auto' || overflowY === 'scroll') && current.scrollHeight > current.clientHeight
    if (canScroll) {
      return current
    }
    current = current.parentElement
  }
  return null
}

const scrollVideoGenerationToBottom = async () => {
  await nextTick()
  window.requestAnimationFrame(() => {
    const container = getScrollContainer(videoGenerationBottomRef.value)
    if (!container) return

    container.scrollTo({
      top: container.scrollHeight,
      behavior: 'smooth'
    })
  })
}

const growHotMusicVisibleCount = () => {
  if (!isHotMusicLazyMode.value) return
  if (hotMusicVisibleCount.value >= filteredAudioAssets.value.length) return
  hotMusicVisibleCount.value = Math.min(
    filteredAudioAssets.value.length,
    hotMusicVisibleCount.value + HOT_MUSIC_LAZY_BATCH_SIZE
  )
}

const isNearBottom = (container: HTMLElement) => {
  const remain = container.scrollHeight - container.scrollTop - container.clientHeight
  return remain <= 120
}

const canLoadMoreHotMusicFromSource = () => {
  return hasMoreHotMusicFallbackPages() && douyinMusicAssets.value.length < HOT_MUSIC_MAX_ITEMS
}

const tryLoadMoreHotMusicWhileScrolling = async () => {
  if (!isAudioLazyMode.value || !hotMusicScrollContainer) return
  if (hotMusicScrollLoading.value) return

  hotMusicScrollLoading.value = true
  try {
    let guard = 0
    while (guard < 12 && hotMusicScrollContainer && isNearBottom(hotMusicScrollContainer)) {
      if (isSfxLazyMode.value) {
        if (sfxLoadingMore.value || !sfxHasMore.value) {
          break
        }

        const before = sfxAssets.value.length
        await loadSfx(true)
        guard += 1
        await nextTick()

        if (sfxAssets.value.length <= before) {
          break
        }
        continue
      }

      if (hotMusicVisibleCount.value < filteredAudioAssets.value.length) {
        growHotMusicVisibleCount()
        guard += 1
        await nextTick()
        continue
      }

      if (!canLoadMoreHotMusicFromSource()) {
        break
      }

      const before = filteredAudioAssets.value.length
      await loadFallbackHotMusic([], 1)
      guard += 1
      await nextTick()
      if (filteredAudioAssets.value.length <= before) {
        if (!canLoadMoreHotMusicFromSource()) {
          break
        }
        continue
      }
    }
  } finally {
    hotMusicScrollLoading.value = false
  }
}

const handleHotMusicScroll = () => {
  if (!hotMusicScrollContainer) return
  if (hotMusicScrollRAF) {
    window.cancelAnimationFrame(hotMusicScrollRAF)
  }
  hotMusicScrollRAF = window.requestAnimationFrame(() => {
    hotMusicScrollRAF = 0
    void tryLoadMoreHotMusicWhileScrolling()
  })
}

const unbindHotMusicScroll = () => {
  if (hotMusicScrollContainer && hotMusicScrollBound) {
    hotMusicScrollContainer.removeEventListener('scroll', handleHotMusicScroll)
  }
  hotMusicScrollContainer = null
  hotMusicScrollBound = false
  hotMusicScrollLoading.value = false
  if (hotMusicScrollRAF) {
    window.cancelAnimationFrame(hotMusicScrollRAF)
    hotMusicScrollRAF = 0
  }
}

const bindHotMusicScroll = async () => {
  unbindHotMusicScroll()
  if (!isAudioLazyMode.value || activeTab.value !== 'audio') return

  await nextTick()
  hotMusicScrollContainer = getScrollContainer(audioListRef.value)
  if (!hotMusicScrollContainer) return

  hotMusicScrollContainer.addEventListener('scroll', handleHotMusicScroll, { passive: true })
  hotMusicScrollBound = true
  void tryLoadMoreHotMusicWhileScrolling()
}

const resetHotMusicVisibleCount = () => {
  hotMusicVisibleCount.value = HOT_MUSIC_LAZY_BATCH_SIZE
}

watch(audioMode, () => {
  resetHotMusicVisibleCount()
  audioCategory.value = 'all'
  audioSearchPage.value = 1
  if (audioMode.value === 'sfx') {
    if (neteaseSearchAbortController) {
      neteaseSearchAbortController.abort()
      neteaseSearchAbortController = null
    }
    neteaseSearchResults.value = []
    neteaseSearchTotal.value = 0
    loadSfx()
    void bindHotMusicScroll()
    return
  }

  if (audioSearch.value.trim()) {
    searchNeteaseSongs(audioSearch.value)
  } else if (douyinMusicAssets.value.length === 0) {
    loadDouyinMusic()
  }
  void bindHotMusicScroll()
})

watch(audioCategoryOptions, (options) => {
  if (!options.some(option => option.value === audioCategory.value)) {
    audioCategory.value = 'all'
  }
})

watch(sfxCategory, () => {
  if (audioMode.value === 'sfx' && !audioSearch.value.trim()) {
    loadSfx()
    void bindHotMusicScroll()
  }
})

let audioSearchTimer: number | null = null
let neteaseSearchAbortController: AbortController | null = null
let hotMusicScrollContainer: HTMLElement | null = null
let hotMusicScrollBound = false
let hotMusicScrollRAF = 0
let hotMusicFallbackCursors: HotMusicFallbackCursor[] = []
let hotMusicFallbackCursorIndex = 0
const hotMusicScrollLoading = ref(false)
const hotMusicFallbackLoading = ref(false)
watch(audioSearch, (value) => {
  resetHotMusicVisibleCount()
  const query = value.trim()
  audioSearchPage.value = 1
  if (audioSearchTimer) {
    window.clearTimeout(audioSearchTimer)
  }

  if (audioMode.value === 'sfx') {
    if (neteaseSearchAbortController) {
      neteaseSearchAbortController.abort()
      neteaseSearchAbortController = null
    }
    neteaseSearchResults.value = []
    neteaseSearchTotal.value = 0
    audioSearchTimer = window.setTimeout(() => {
      loadSfx()
    }, 220)
    return
  }

  if (!query || audioMode.value !== 'music') {
    neteaseSearchResults.value = []
    neteaseSearchTotal.value = 0
    if (neteaseSearchAbortController) {
      neteaseSearchAbortController.abort()
      neteaseSearchAbortController = null
    }
    return
  }
  audioSearchTimer = window.setTimeout(() => {
    searchNeteaseSongs(query)
  }, 260)
})

watch([audioSearchPage, audioSearchPageSize], () => {
  if (!audioSearch.value.trim() || audioMode.value !== 'music') return
  searchNeteaseSongs(audioSearch.value)
})

const handleAudioSearchPageChange = () => {
  if (!audioSearch.value.trim() || audioMode.value !== 'music') return
  neteaseSearchResults.value = []
  searchNeteaseSongs(audioSearch.value)
}

const stopAudioPreview = (resetLoading = true) => {
  if (previewAudioPlayer.value) {
    previewAudioPlayer.value.pause()
    previewAudioPlayer.value.currentTime = 0
  }
  previewAudioPlayer.value = null
  previewingAudioId.value = null
  if (resetLoading) {
    previewLoadingAudioId.value = null
  }
}

watch(activeTab, (tab) => {
  if (tab !== 'audio') {
    unbindHotMusicScroll()
    stopAudioPreview()
    return
  }
  if (audioMode.value === 'sfx') {
    loadSfx()
  } else if (douyinMusicAssets.value.length === 0) {
    loadDouyinMusic()
  }
  void bindHotMusicScroll()
})

watch(
  [isAudioLazyMode, () => filteredAudioAssets.value.length, activeTab, isHotMusicLazyMode],
  async ([lazyMode, , , hotLazyMode]) => {
    if (!lazyMode || activeTab.value !== 'audio') {
      unbindHotMusicScroll()
      return
    }
    if (hotLazyMode) {
      if (hotMusicVisibleCount.value > filteredAudioAssets.value.length) {
        hotMusicVisibleCount.value = filteredAudioAssets.value.length
      }
      if (hotMusicVisibleCount.value < HOT_MUSIC_LAZY_BATCH_SIZE) {
        hotMusicVisibleCount.value = HOT_MUSIC_LAZY_BATCH_SIZE
      }
    }
    await bindHotMusicScroll()
  }
)

const buildAssetResolverUrl = (asset: AudioListItem) => {
  const source = (asset.source || '').toLowerCase()
  if (!source || source === 'netease' || source === 'douyin' || source === 'asset' || source === 'sfx') {
    return ''
  }
  return `/api/v1/music/stream?source=${encodeURIComponent(source)}&id=${encodeURIComponent(asset.sourceId || '')}&mid=${encodeURIComponent(asset.sourceMid || '')}&hash=${encodeURIComponent(asset.sourceHash || '')}&content_id=${encodeURIComponent(asset.sourceContentId || '')}&title=${encodeURIComponent(asset.name || '')}&artist=${encodeURIComponent(asset.artist || '')}`
}

const buildAssetCandidateUrls = (asset: AudioListItem) => {
  const candidates: string[] = []
  if (asset.source === 'netease' && asset.sourceId) {
    candidates.push(buildNeteaseStreamUrl(asset.sourceId))
  }
  const resolverUrl = buildAssetResolverUrl(asset)
  if (resolverUrl) candidates.push(resolverUrl)
  if (asset.sourceSongUrl) {
    candidates.push(`/api/v1/music/stream?url=${encodeURIComponent(asset.sourceSongUrl)}`)
  }
  if (asset.url) {
    candidates.push(asset.url)
  }
  return [...new Set(candidates.filter(Boolean))]
}

const probeAudioUrl = async (url: string) => {
  try {
    const controller = new AbortController()
    const timer = window.setTimeout(() => controller.abort(), MUSIC_PREVIEW_PROBE_TIMEOUT_MS)
    const response = await fetch(url, {
      method: 'GET',
      headers: { Range: 'bytes=0-2048' },
      cache: 'no-store',
      signal: controller.signal
    })
    window.clearTimeout(timer)
    if (!response.ok) return false

    const contentType = (response.headers.get('content-type') || '').toLowerCase()
    const contentLength = Number(response.headers.get('content-length') || 0)
    if (contentType.includes('application/json') || contentType.includes('text/html') || contentType.includes('text/plain')) {
      return false
    }
    const isBackendMusicProxy = url.includes('/api/v1/music/stream') || url.includes('/api/v1/music/netease/stream')
    const chunk = new Uint8Array(await response.arrayBuffer())
    if (isBackendMusicProxy && contentType.includes('audio/wav')) {
      // 后端兜底静音 wav：RIFF 头 + 大量 0 字节，直接判定为不可播放候选。
      const payload = chunk.slice(44)
      const sample = payload.length > 0 ? payload : chunk
      const zeroCount = sample.reduce((count, b) => count + (b === 0 ? 1 : 0), 0)
      const zeroRatio = sample.length > 0 ? zeroCount / sample.length : 1
      if ((contentLength > 0 && contentLength <= 90000) || zeroRatio > 0.98) {
        return false
      }
    }
    if (contentType.startsWith('audio/') || contentType.includes('octet-stream')) {
      return true
    }
    if (chunk.length < 4) return false
    const header = Array.from(chunk.slice(0, 4)).map(b => String.fromCharCode(b)).join('')
    return header === 'RIFF' || header === 'ID3' || header === 'OggS' || header === 'fLaC'
  } catch {
    return false
  }
}

const resolvePlayableAudioUrl = async (asset: AudioListItem) => {
  const cacheKey = asset.id || `${asset.name}-${asset.artist || ''}`
  const cached = audioPlayableUrlCache.get(cacheKey)
  if (cached) return cached

  const candidates = buildAssetCandidateUrls(asset)
  for (const candidate of candidates) {
    const normalized = resolveAudioUrl(candidate)
    if (!normalized) continue
    if (await probeAudioUrl(normalized)) {
      audioPlayableUrlCache.set(cacheKey, normalized)
      audioPreviewCache.set(cacheKey, normalized)
      asset.url = normalized
      return normalized
    }
  }

  const fallbackCandidates = await resolveSearchPreviewCandidates(asset)
  for (const fallback of fallbackCandidates) {
    const normalizedFallback = resolveAudioUrl(fallback)
    if (!normalizedFallback) continue
    if (await probeAudioUrl(normalizedFallback)) {
      audioPlayableUrlCache.set(cacheKey, normalizedFallback)
      audioPreviewCache.set(cacheKey, normalizedFallback)
      asset.url = normalizedFallback
      return normalizedFallback
    }
  }
  const fallbackUrl = resolveAudioUrl(candidates[0]) || resolveAudioUrl(fallbackCandidates[0])
  if (fallbackUrl) {
    audioPlayableUrlCache.set(cacheKey, fallbackUrl)
    audioPreviewCache.set(cacheKey, fallbackUrl)
    asset.url = fallbackUrl
    return fallbackUrl
  }
  return ''
}

const getAudioDurationFromUrl = (url: string): Promise<number | undefined> => {
  return new Promise((resolve) => {
    const audio = document.createElement('audio')
    let timeoutId: number | null = null
    const cleanup = () => {
      audio.onloadedmetadata = null
      audio.onerror = null
      if (timeoutId) {
        window.clearTimeout(timeoutId)
      }
      audio.src = ''
      audio.remove()
    }

    timeoutId = window.setTimeout(() => {
      cleanup()
      resolve(undefined)
    }, AUDIO_METADATA_TIMEOUT_MS)

    audio.preload = 'metadata'
    audio.onloadedmetadata = () => {
      const duration = Number(audio.duration)
      cleanup()
      if (Number.isFinite(duration) && duration > 0) {
        resolve(duration)
      } else {
        resolve(undefined)
      }
    }
    audio.onerror = () => {
      cleanup()
      resolve(undefined)
    }
    audio.src = url
  })
}

const fillMissingDurations = async (assets: AudioListItem[]) => {
  const tasks = assets.map(async (asset) => {
    if (asset.duration && asset.duration > 0) return
    const cacheKey = asset.id || `${asset.name}-${asset.artist || ''}`
    if (!cacheKey) return

    if (audioDurationCache.has(cacheKey)) {
      asset.duration = audioDurationCache.get(cacheKey)
      return
    }
    if (audioDurationLoading.has(cacheKey)) return

    audioDurationLoading.add(cacheKey)
    try {
      const playableUrl = await resolvePlayableAudioUrl(asset)
      if (!playableUrl) return
      const duration = await getAudioDurationFromUrl(playableUrl)
      if (duration && duration > 0) {
        const seconds = Math.round(duration)
        audioDurationCache.set(cacheKey, seconds)
        asset.duration = seconds
      }
    } finally {
      audioDurationLoading.delete(cacheKey)
    }
  })

  await Promise.allSettled(tasks)
}

const toggleAudioPreview = async (asset: AudioListItem) => {
  if (previewLoadingAudioId.value === asset.id) {
    return
  }

  if (previewingAudioId.value === asset.id && !previewLoadingAudioId.value) {
    stopAudioPreview()
    return
  }

  const requestId = ++previewRequestSeq
  previewLoadingAudioId.value = asset.id

  try {
    const url = await resolvePlayableAudioUrl(asset)
    if (requestId !== previewRequestSeq) return

    if (!url) {
      unsupportedPreviewAudioIds.value.add(asset.id)
      ElMessage.warning('该音乐暂不支持试听，已自动尝试多个音源')
      previewingAudioId.value = null
      return
    }

    unsupportedPreviewAudioIds.value.delete(asset.id)
    stopAudioPreview(false)
    const audio = new Audio(url)
    previewAudioPlayer.value = audio
    previewingAudioId.value = asset.id

    try {
      await audio.play()
      audio.onended = () => {
        if (previewingAudioId.value === asset.id) {
          previewingAudioId.value = null
        }
      }
      return
    } catch (err) {
      console.warn('音频播放失败:', err)
    }

    audioPlayableUrlCache.delete(asset.id)
    const retryUrl = await resolvePlayableAudioUrl(asset)
    if (requestId !== previewRequestSeq) return

    if (retryUrl && retryUrl !== url) {
      const retryAudio = new Audio(retryUrl)
      previewAudioPlayer.value = retryAudio
      previewingAudioId.value = asset.id
      try {
        await retryAudio.play()
        retryAudio.onended = () => {
          if (previewingAudioId.value === asset.id) {
            previewingAudioId.value = null
          }
        }
        return
      } catch (retryErr) {
        console.warn('音频重试失败:', retryErr)
      }
    }

    unsupportedPreviewAudioIds.value.add(asset.id)
    ElMessage.warning('该音乐暂不支持试听')
    previewingAudioId.value = null
  } finally {
    if (requestId === previewRequestSeq && previewLoadingAudioId.value === asset.id) {
      previewLoadingAudioId.value = null
    }
  }
}

const addAudioToTimeline = async (asset: AudioListItem) => {
  const url = await resolvePlayableAudioUrl(asset)
  if (!url) {
    ElMessage.error('音频素材不可用')
    return
  }
  const editor = timelineEditorRef.value as any
  if (!editor?.addAudioClipFromAsset) {
    ElMessage.warning('时间线未就绪，无法添加音频')
    return
  }
  const assetId = asset.assetId ?? asset.id
  await editor.addAudioClipFromAsset({
    id: assetId,
    url,
    duration: asset.duration,
    name: asset.name
  })
}

// 当前模型能力
const currentModelCapability = computed<VideoModelCapability | null>(() => {
  if (!selectedVideoModel.value) return null
  const matched = videoModelCapabilities.value.find(m => m.id === selectedVideoModel.value)
  if (matched) return matched
  return getVideoModelCapability(selectedVideoModel.value)
})

// 当前模型支持的参考图模式
const availableReferenceModes = computed(() => {
  const capability = currentModelCapability.value
  if (!capability) return []

  const modes: Array<{ value: string, label: string, description?: string }> = []

  if (capability.supportTextOnly) {
    modes.push({ value: 'none', label: '纯文本', description: '不使用参考图' })
  }
  if (capability.supportSingleImage) {
    modes.push({ value: 'single', label: '单图', description: '使用单张参考图' })
  }
  if (capability.supportFirstLastFrame) {
    modes.push({ value: 'first_last', label: '首尾帧', description: '使用首帧和尾帧' })
  }
  if (capability.supportMultipleImages) {
    modes.push({ value: 'multiple', label: '多图', description: `最多${capability.maxImages}张` })
  }

  return modes
})

const pickDefaultReferenceMode = (modes: Array<{ value: string }>) => {
  if (!modes.length) return ''

  const preferredOrder = ['single', 'first_last', 'multiple', 'none']
  for (const mode of preferredOrder) {
    if (modes.some(m => m.value === mode)) {
      return mode
    }
  }

  return modes[0].value
}

// 帧提示词存储key生成函数
const getPromptStorageKey = (storyboardId: number | string | undefined, frameType: FrameType) => {
  if (!storyboardId) return null
  return `frame_prompt_${storyboardId}_${frameType}`
}

const isCharacterSelected = (charId: number) => {
  return selectedCharacters.value.includes(charId)
}

const toggleCharacter = (charId: number) => {
  const index = selectedCharacters.value.indexOf(charId)
  if (index > -1) {
    selectedCharacters.value.splice(index, 1)
  } else {
    selectedCharacters.value.push(charId)
  }
}

const currentStoryboard = computed(() => {
  if (!currentStoryboardId.value) return null
  return storyboards.value.find(s => String(s.id) === String(currentStoryboardId.value)) || null
})

// 获取上一个镜头
const previousStoryboard = computed(() => {
  if (!currentStoryboardId.value || storyboards.value.length < 2) return null
  const currentIndex = storyboards.value.findIndex(s => String(s.id) === String(currentStoryboardId.value))
  if (currentIndex <= 0) return null
  return storyboards.value[currentIndex - 1]
})

// 上一个镜头的尾帧图片列表（支持多个）
const previousStoryboardLastFrames = ref<any[]>([])

// 加载上一个镜头的尾帧
const loadPreviousStoryboardLastFrame = async () => {
  if (!previousStoryboard.value) {
    previousStoryboardLastFrames.value = []
    return
  }
  try {
    const result = await imageAPI.listImages({
      storyboard_id: Number(previousStoryboard.value.id),
      frame_type: 'last',
      page: 1,
      page_size: 10
    })
    const images = result.items || []
    previousStoryboardLastFrames.value = images.filter((img: any) => img.status === 'completed' && img.image_url)
  } catch (error) {
    console.error('加载上一镜头尾帧失败:', error)
    previousStoryboardLastFrames.value = []
  }
}

// 选择上一镜头尾帧作为首帧参考
const selectPreviousLastFrame = (img: any) => {
  // 检查是否已选中，已选中则取消
  const currentIndex = selectedImagesForVideo.value.indexOf(img.id)
  if (currentIndex > -1) {
    selectedImagesForVideo.value.splice(currentIndex, 1)
    ElMessage.success('已取消首帧参考')
    return
  }

  // 参考handleImageSelect的逻辑，根据模式处理
  if (!selectedReferenceMode.value || selectedReferenceMode.value === 'single') {
    // 单图模式或未选模式：直接替换
    selectedImagesForVideo.value = [img.id]
  } else if (selectedReferenceMode.value === 'first_last') {
    // 首尾帧模式：作为首帧参考
    selectedImagesForVideo.value = [img.id]
  } else if (selectedReferenceMode.value === 'multiple') {
    // 多图模式：添加到列表
    const capability = currentModelCapability.value
    if (capability && selectedImagesForVideo.value.length >= capability.maxImages) {
      ElMessage.warning(`最多只能选择${capability.maxImages}张图片`)
      return
    }
    selectedImagesForVideo.value.push(img.id)
  }
  ElMessage.success('已添加为首帧参考')
}

// 监听帧类型切换，从存储中加载或清空
watch(selectedFrameType, (newType) => {
  // 切换帧类型时，停止之前的轮询，避免旧结果覆盖新帧类型
  stopPolling()

  if (!currentStoryboard.value) {
    currentFramePrompt.value = ''
    generatedImages.value = []
    return
  }

  // 设置切换标志，防止watch(currentFramePrompt)错误保存
  isSwitchingFrameType.value = true

  // 从 framePrompts 对象中加载该帧类型的提示词
  currentFramePrompt.value = framePrompts.value[newType] || ''

  // 从 sessionStorage 中加载该帧类型之前的提示词（如果framePrompts中没有）
  if (!currentFramePrompt.value) {
    const storageKey = `frame_prompt_${currentStoryboard.value.id}_${newType}`
    const stored = sessionStorage.getItem(storageKey)
    if (stored) {
      currentFramePrompt.value = stored
      framePrompts.value[newType] = stored
    }
  }

  // 重新加载该帧类型的图片
  loadStoryboardImages(Number(currentStoryboard.value.id), newType)

  // 重置切换标志
  setTimeout(() => {
    isSwitchingFrameType.value = false
  }, 0)
})

// 监听当前分镜切换，重置提示词
watch(currentStoryboard, async (newStoryboard) => {
  if (!newStoryboard) {
    currentFramePrompt.value = ''
    generatedImages.value = []
    generatedVideos.value = []
    videoReferenceImages.value = []
    previousStoryboardLastFrames.value = []
    return
  }

  // 设置切换标志
  isSwitchingFrameType.value = true

  // 加载当前帧类型的提示词
  const storageKey = getPromptStorageKey(newStoryboard.id, selectedFrameType.value)
  if (storageKey) {
    const stored = sessionStorage.getItem(storageKey)
    currentFramePrompt.value = stored || ''
  } else {
    currentFramePrompt.value = ''
  }

  // 重置切换标志
  setTimeout(() => {
    isSwitchingFrameType.value = false
  }, 0)

  // 加载该分镜的图片列表（根据当前选择的帧类型）
  await loadStoryboardImages(Number(newStoryboard.id), selectedFrameType.value)

  // 加载视频参考图片（所有帧类型）
  await loadVideoReferenceImages(Number(newStoryboard.id))

  // 加载该分镜的视频列表
  await loadStoryboardVideos(Number(newStoryboard.id))

  // 加载上一镜头的尾帧
  await loadPreviousStoryboardLastFrame()
})

// 监听提示词变化，自动保存到sessionStorage
watch(currentFramePrompt, (newPrompt) => {
  // 如果正在切换帧类型或分镜，不要保存（避免错误保存到新帧类型）
  if (isSwitchingFrameType.value) return
  if (!currentStoryboard.value) return

  const storageKey = getPromptStorageKey(currentStoryboard.value.id, selectedFrameType.value)
  if (storageKey) {
    if (newPrompt) {
      sessionStorage.setItem(storageKey, newPrompt)
    } else {
      sessionStorage.removeItem(storageKey)
    }
  }
})

// 监听视频模型切换，清空已选图片和参考图模式
watch(selectedVideoModel, async () => {
  selectedImagesForVideo.value = []
  selectedLastImageForVideo.value = null
  await nextTick()
  selectedReferenceMode.value = pickDefaultReferenceMode(availableReferenceModes.value)
})

watch(availableReferenceModes, (modes) => {
  if (!modes.length) {
    selectedReferenceMode.value = ''
    return
  }

  const current = selectedReferenceMode.value
  const isValid = modes.some(mode => mode.value === current)
  if (!current || !isValid) {
    selectedReferenceMode.value = pickDefaultReferenceMode(modes)
  }
}, { immediate: true })

// 监听镜头切换，自动更新视频时长
watch(currentStoryboard, (newStoryboard) => {
  if (newStoryboard?.duration) {
    // 如果镜头有duration字段，使用镜头的时长
    videoDuration.value = Math.round(newStoryboard.duration)
  } else {
    // 否则使用默认值5秒
    videoDuration.value = 5
  }
})

// 监听参考图模式切换，清空已选图片
watch(selectedReferenceMode, () => {
  selectedImagesForVideo.value = []
  selectedLastImageForVideo.value = null
})

// 当前分镜的角色列表
const currentStoryboardCharacters = computed(() => {
  if (!currentStoryboard.value?.characters) return []

  // currentStoryboard.characters 是角色对象数组
  if (Array.isArray(currentStoryboard.value.characters) && currentStoryboard.value.characters.length > 0) {
    const firstItem = currentStoryboard.value.characters[0]
    // 如果是对象数组（包含id和name），直接返回
    if (typeof firstItem === 'object' && firstItem.id) {
      return currentStoryboard.value.characters
    }
    // 如果是ID数组，从characters中查找匹配的角色
    if (typeof firstItem === 'number') {
      return characters.value.filter(c => currentStoryboard.value.characters.includes(c.id))
    }
  }

  return []
})

// 可选择的角色列表
const availableCharacters = computed(() => {
  return characters.value || []
})

// 检查角色是否已在当前镜头中
const isCharacterInCurrentShot = (charId: number) => {
  if (!currentStoryboard.value?.characters) return false

  if (Array.isArray(currentStoryboard.value.characters) && currentStoryboard.value.characters.length > 0) {
    const firstItem = currentStoryboard.value.characters[0]
    if (typeof firstItem === 'object' && firstItem.id) {
      return currentStoryboard.value.characters.some(c => c.id === charId)
    }
    if (typeof firstItem === 'number') {
      return currentStoryboard.value.characters.includes(charId)
    }
  }

  return false
}

// 切换角色在镜头中的状态
const showCharacterImage = (char: any) => {
  previewCharacter.value = char
  showCharacterImagePreview.value = true
}

// 展示场景大图
const showSceneImage = () => {
  if (currentStoryboard.value?.background?.image_url) {
    showSceneImagePreview.value = true
  }
}

// 保存分镜字段
const saveStoryboardField = async (fieldName: string) => {
  if (!currentStoryboard.value) return
  try {
    const updateData: any = {}
    updateData[fieldName] = currentStoryboard.value[fieldName]

    await dramaAPI.updateStoryboard(currentStoryboard.value.id.toString(), updateData)
  } catch (error: any) {
    ElMessage.error('保存失败: ' + (error.message || '未知错误'))
  }
}

// 提取帧提示词
const extractFramePrompt = async () => {
  if (!currentStoryboard.value) return

  // 记录点击时的帧类型，避免切换tab后提示词显示错位
  const targetFrameType = selectedFrameType.value

  generatingPrompt.value = true
  try {
    const params: any = { frame_type: targetFrameType }
    if (targetFrameType === 'panel') {
      params.panel_count = panelCount.value
    }

    const result = await generateFramePrompt(Number(currentStoryboard.value.id), params)

    // 根据记录的帧类型提取prompt，确保更新到正确的位置
    let extractedPrompt = ''
    if (result.single_frame) {
      extractedPrompt = result.single_frame.prompt
    } else if (result.multi_frame && result.multi_frame.frames) {
      // 多帧情况，将所有帧的prompt合并
      extractedPrompt = result.multi_frame.frames
        .map((frame: any, index: number) => `${frame.description}: ${frame.prompt}`)
        .join('\n\n')
    }

    // 只在当前仍然选中该帧类型时才更新显示
    if (selectedFrameType.value === targetFrameType) {
      currentFramePrompt.value = extractedPrompt
    }

    // 存储到对应帧类型的提示词中
    framePrompts.value[targetFrameType] = extractedPrompt

    ElMessage.success(`${getFrameTypeLabel(targetFrameType)}提示词提取成功`)
  } catch (error: any) {
    ElMessage.error('提取失败: ' + (error.message || '未知错误'))
  } finally {
    generatingPrompt.value = false
  }
}

// 获取帧类型的中文标签
const getFrameTypeLabel = (frameType: string): string => {
  const labels: Record<string, string> = {
    key: '关键帧',
    first: '首帧',
    last: '尾帧',
    panel: '分镜版'
  }
  return labels[frameType] || frameType
}

// 加载分镜的图片列表
const loadStoryboardImages = async (storyboardId: number, frameType?: string) => {
  loadingImages.value = true
  try {
    const params: any = {
      storyboard_id: storyboardId,
      page: 1,
      page_size: 50
    }
    // 如果指定了帧类型，添加过滤
    if (frameType) {
      params.frame_type = frameType
    }
    const result = await imageAPI.listImages(params)
    generatedImages.value = result.items || []

    // 如果有进行中的任务，启动轮询
    const hasPendingOrProcessing = generatedImages.value.some(
      img => img.status === 'pending' || img.status === 'processing'
    )
    if (hasPendingOrProcessing) {
      startPolling()
    }
  } catch (error: any) {
    console.error('加载图片列表失败:', error)
  } finally {
    loadingImages.value = false
  }
}

// 启动状态轮询
const startPolling = () => {
  if (pollingTimer) return

  // 记录开始轮询时的帧类型
  pollingFrameType = selectedFrameType.value

  pollingTimer = setInterval(async () => {
    if (!currentStoryboard.value) {
      stopPolling()
      return
    }

    // 如果帧类型已切换，停止轮询（防止更新到错误的帧类型）
    if (selectedFrameType.value !== pollingFrameType) {
      stopPolling()
      return
    }

    try {
      const params: any = {
        storyboard_id: Number(currentStoryboard.value.id),
        page: 1,
        page_size: 50
      }
      // 使用轮询开始时记录的帧类型
      if (pollingFrameType) {
        params.frame_type = pollingFrameType
      }
      const result = await imageAPI.listImages(params)

      // 再次检查帧类型是否仍然匹配，避免竞态条件
      if (selectedFrameType.value === pollingFrameType) {
        generatedImages.value = result.items || []
      }

      // 如果没有进行中的任务，停止轮询并刷新视频参考图片
      const hasPendingOrProcessing = (result.items || []).some(
        (img: any) => img.status === 'pending' || img.status === 'processing'
      )
      if (!hasPendingOrProcessing) {
        stopPolling()
        // 刷新视频参考图片列表
        if (currentStoryboard.value) {
          loadVideoReferenceImages(Number(currentStoryboard.value.id))
        }
      }
    } catch (error) {
      console.error('轮询图片状态失败:', error)
    }
  }, 3000) // 每3秒轮询一次
}

// 停止轮询
const stopPolling = () => {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
  pollingFrameType = null
}

// 生成图片
const generateFrameImage = async () => {
  if (!currentStoryboard.value || !currentFramePrompt.value) return

  generatingImage.value = true
  try {
    // 收集参考图片URL
    const referenceImages: string[] = []

    // 1. 添加场景图片（从background字段获取）
    if (currentStoryboard.value.background?.image_url) {
      referenceImages.push(currentStoryboard.value.background.image_url)
    }

    // 2. 添加当前镜头登场的角色图片
    const storyboardCharacters = currentStoryboardCharacters.value
    if (storyboardCharacters && storyboardCharacters.length > 0) {
      storyboardCharacters.forEach((char: any) => {
        if (char.image_url) {
          referenceImages.push(char.image_url)
        }
      })
    }

    const result = await imageAPI.generateImage({
      drama_id: dramaId.toString(),
      prompt: currentFramePrompt.value,
      storyboard_id: Number(currentStoryboard.value.id),
      image_type: 'storyboard',
      frame_type: selectedFrameType.value,
      reference_images: referenceImages.length > 0 ? referenceImages : undefined
    })

    generatedImages.value.unshift(result)

    ElMessage.success('图片生成任务已提交')

    // 启动轮询
    startPolling()
  } catch (error: any) {
    ElMessage.error('生成失败: ' + (error.message || '未知错误'))
  } finally {
    generatingImage.value = false
  }
}

// 获取状态标签类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, any> = {
    pending: 'info',
    processing: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return statusMap[status] || 'info'
}

const normalizeVideo = (video: VideoGeneration) => {
  const normalized: VideoGeneration = { ...video }
  const derivedUrl = normalized.minio_url || normalized.local_path || normalized.video_url || ''
  const statusText = String(normalized.status || '')
  if (statusText === 'success') {
    normalized.status = 'completed'
  }
  if (derivedUrl) {
    normalized.video_url = derivedUrl
  }
  if (derivedUrl && (normalized.status === 'failed' || !normalized.status)) {
    normalized.status = 'completed'
  } else if (!normalized.status) {
    normalized.status = 'processing'
  }
  return normalized
}

// 播放视频
const playVideo = (video: VideoGeneration) => {
  previewVideo.value = video
  showVideoPreview.value = true
}

const openVideoUrl = (url?: string | null) => {
  if (!url) return
  const resolved = resolveVideoUrl(url)
  if (!resolved) return
  window.open(resolved, '_blank')
}

const normalizeVideoAssetCompareURL = (raw?: string | null) => {
  const fixed = fixMediaUrl(raw || '')
  if (!fixed) return ''
  if (fixed.startsWith('data:') || fixed.startsWith('blob:')) return fixed
  return fixed.replace(/^https?:\/\/[^/]+/i, '')
}

const isVideoInAssetLibrary = (video: VideoGeneration) => {
  if (!video) return false

  const videoID = Number(video.id)
  const videoURL = normalizeVideoAssetCompareURL(video.video_url || video.local_path || video.minio_url || '')

  return videoAssets.value.some((asset: any) => {
    if (asset?.type !== 'video') return false

    const assetVideoGenID = Number(asset.video_gen_id)
    if (!Number.isNaN(assetVideoGenID) && assetVideoGenID > 0 && assetVideoGenID === videoID) {
      return true
    }

    const assetURL = normalizeVideoAssetCompareURL(asset.url || asset.local_path || '')
    return !!assetURL && !!videoURL && assetURL === videoURL
  })
}

// 添加视频到素材库
const addVideoToAssets = async (video: VideoGeneration) => {
  if (video.status !== 'completed' || !video.video_url) {
    ElMessage.warning('只能添加已完成的视频到素材库')
    return
  }

  if (isVideoInAssetLibrary(video)) {
    ElMessage.warning('该视频已在素材库中，请勿重复添加')
    return
  }

  addingToAssets.value.add(video.id)

  try {
    // 直接新增到素材库，允许同一分镜保留多个视频版本
    await assetAPI.importFromVideo(video.id)
    ElMessage.success('已添加到素材库')

    // 重新加载素材库列表
    await loadVideoAssets()
  } catch (error: any) {
    ElMessage.error(error.message || '添加失败')
  } finally {
    addingToAssets.value.delete(video.id)
  }
}

// 获取状态中文文本
const getStatusText = (status: string) => {
  const statusTextMap: Record<string, string> = {
    pending: '等待中',
    processing: '生成中',
    completed: '已完成',
    failed: '失败'
  }
  return statusTextMap[status] || status
}

// 获取帧类型中文文本
const getFrameTypeText = (frameType?: string) => {
  if (!frameType) return ''
  const frameTypeMap: Record<string, string> = {
    first: '首帧',
    key: '关键帧',
    last: '尾帧',
    panel: '分镜板',
    action: '动作序列'
  }
  return frameTypeMap[frameType] || frameType
}

// 获取分镜缩略图
const getStoryboardThumbnail = (storyboard: any) => {
  // 优先使用composed_image
  if (storyboard.composed_image) {
    return storyboard.composed_image
  }

  // 如果没有composed_image，从image_url字段获取
  if (storyboard.image_url) {
    return storyboard.image_url
  }

  return null
}

// 处理图片选择（根据模型能力）
const handleImageSelect = (imageId: number) => {
  if (!selectedReferenceMode.value) {
    ElMessage.warning('请先选择参考图模式')
    return
  }

  if (!currentModelCapability.value) {
    ElMessage.warning('请先选择视频生成模型')
    return
  }

  const capability = currentModelCapability.value
  const currentIndex = selectedImagesForVideo.value.indexOf(imageId)

  // 已选中，则取消选择
  if (currentIndex > -1) {
    selectedImagesForVideo.value.splice(currentIndex, 1)
    return
  }

  // 获取当前点击的图片对象
  const clickedImage = videoReferenceImages.value.find(img => img.id === imageId)
  if (!clickedImage) return

  // 根据选择的参考图模式处理
  switch (selectedReferenceMode.value) {
    case 'single':
      // 单图模式：只能选1张，直接替换
      selectedImagesForVideo.value = [imageId]
      break

    case 'first_last':
      // 首尾帧模式：根据图片类型分别处理
      const frameType = clickedImage.frame_type

      if (frameType === 'first' || frameType === 'panel' || frameType === 'key') {
        // 首帧：直接替换
        selectedImagesForVideo.value = [imageId]
      } else if (frameType === 'last') {
        // 尾帧：设置到单独的变量
        selectedLastImageForVideo.value = imageId
      } else {
        ElMessage.warning('首尾帧模式下，请选择首帧或尾帧类型的图片')
      }
      break

    case 'multiple':
      // 多图模式：检查是否超出最大数量
      if (selectedImagesForVideo.value.length >= capability.maxImages) {
        ElMessage.warning(`最多只能选择${capability.maxImages}张图片`)
        return
      }
      selectedImagesForVideo.value.push(imageId)
      break

    default:
      ElMessage.warning('未知的参考图模式')
  }
}

// 预览图片
const previewImage = (url: string) => {
  // 使用Element Plus的图片预览
  const viewer = document.createElement('div')
  viewer.innerHTML = `
    <div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 9999; background: rgba(0,0,0,0.8); display: flex; align-items: center; justify-content: center;" onclick="this.remove()">
      <img src="${url}" style="max-width: 90vw; max-height: 90vh; object-fit: contain;" onclick="event.stopPropagation();" />
    </div>
  `
  document.body.appendChild(viewer)
}

// 获取已选图片对象列表
const selectedImageObjects = computed(() => {
  return selectedImagesForVideo.value
    .map(id => videoReferenceImages.value.find(img => img.id === id))
    .filter(img => img && img.image_url)
})

// 首尾帧模式：获取首帧图片
const firstFrameSlotImage = computed(() => {
  if (selectedImagesForVideo.value.length === 0) return null
  const firstImageId = selectedImagesForVideo.value[0]
  // 同时搜索当前镜头图片和上一镜头尾帧
  return videoReferenceImages.value.find(img => img.id === firstImageId) 
    || previousStoryboardLastFrames.value.find(img => img.id === firstImageId)
})

// 首尾帧模式：获取尾帧图片
const lastFrameSlotImage = computed(() => {
  if (!selectedLastImageForVideo.value) return null
  // 同时搜索当前镜头图片和上一镜头尾帧
  return videoReferenceImages.value.find(img => img.id === selectedLastImageForVideo.value)
    || previousStoryboardLastFrames.value.find(img => img.id === selectedLastImageForVideo.value)
})

// 移除已选择的图片
const removeSelectedImage = (imageId: number) => {
  // 检查是否是尾帧
  if (selectedLastImageForVideo.value === imageId) {
    selectedLastImageForVideo.value = null
    return
  }

  // 检查是否是首帧或其他图片
  const index = selectedImagesForVideo.value.indexOf(imageId)
  if (index > -1) {
    selectedImagesForVideo.value.splice(index, 1)
  }
}

// 生成视频
const generateVideo = async () => {
  if (!selectedVideoModel.value) {
    ElMessage.warning('请先选择视频生成模型')
    return
  }

  if (!currentStoryboard.value) {
    ElMessage.warning('请先选择分镜')
    return
  }

  // 检查参考图模式
  if (selectedReferenceMode.value !== 'none' && selectedImagesForVideo.value.length === 0) {
    ElMessage.warning('请选择参考图片')
    return
  }

  // 获取第一张选中的图片（仅在需要图片的模式下）
  let selectedImage = null
  if (selectedReferenceMode.value !== 'none' && selectedImagesForVideo.value.length > 0) {
    // 同时搜索当前镜头图片和上一镜头尾帧
    selectedImage = videoReferenceImages.value.find(img => img.id === selectedImagesForVideo.value[0])
      || previousStoryboardLastFrames.value.find(img => img.id === selectedImagesForVideo.value[0])
    if (!selectedImage || !selectedImage.image_url) {
      ElMessage.error('请选择有效的参考图片')
      return
    }
  }

  generatingVideo.value = true
  try {
    // 构建请求参数
    const promptText = (currentStoryboard.value.video_prompt
      || currentStoryboard.value.action
      || currentStoryboard.value.description
      || currentStoryboard.value.title
      || '').trim() || '生成视频'

    const referenceMode = selectedReferenceMode.value || (selectedImage ? 'single' : 'none')
    const requestParams: any = {
      drama_id: dramaId.toString(),
      storyboard_id: Number(currentStoryboard.value.id),
      prompt: promptText,
      duration: videoDuration.value,
      reference_mode: referenceMode
    }

    // 根据参考图模式设置参数
    switch (referenceMode) {
      case 'single':
        // 单图模式
        requestParams.image_gen_id = selectedImage.id
        break

      case 'first_last':
        // 首尾帧模式（同时搜索当前镜头图片和上一镜头尾帧）
        const firstImage = videoReferenceImages.value.find(img => img.id === selectedImagesForVideo.value[0])
          || previousStoryboardLastFrames.value.find(img => img.id === selectedImagesForVideo.value[0])
        const lastImage = videoReferenceImages.value.find(img => img.id === selectedLastImageForVideo.value)
          || previousStoryboardLastFrames.value.find(img => img.id === selectedLastImageForVideo.value)

        if (firstImage?.image_url) {
          requestParams.first_frame_url = toBackendMediaUrl(firstImage.image_url)
        }
        if (lastImage?.image_url) {
          requestParams.last_frame_url = toBackendMediaUrl(lastImage.image_url)
        }
        break

      case 'multiple':
        // 多图模式
        const selectedImages = selectedImagesForVideo.value
          .map(id => videoReferenceImages.value.find(img => img.id === id))
          .filter(img => img?.image_url)
          .map(img => toBackendMediaUrl(img!.image_url))
        requestParams.reference_image_urls = selectedImages
        break

      case 'none':
        // 无参考图模式
        break
    }

    const result = await videoAPI.generateVideo(requestParams)

    generatedVideos.value.unshift(normalizeVideo(result))
    await scrollVideoGenerationToBottom()
    ElMessage.success('视频生成任务已提交')

    // 启动视频轮询
    startVideoPolling()
  } catch (error: any) {
    ElMessage.error('生成失败: ' + (error.message || '未知错误'))
  } finally {
    generatingVideo.value = false
  }
}

// 加载分镜的视频参考图片（所有帧类型）
const loadVideoReferenceImages = async (storyboardId: number) => {
  try {
    const result = await imageAPI.listImages({
      storyboard_id: storyboardId,
      page: 1,
      page_size: 100
    })
    videoReferenceImages.value = result.items || []
  } catch (error: any) {
    console.error('加载视频参考图片失败:', error)
  }
}

// 加载分镜的视频列表
const loadStoryboardVideos = async (storyboardId: number) => {
  loadingVideos.value = true
  try {
    const result = await videoAPI.listVideos({
      storyboard_id: storyboardId.toString(),
      page: 1,
      page_size: 50
    })
    generatedVideos.value = (result.items || []).map(normalizeVideo)

    // 如果有进行中的任务，启动轮询
    const hasPendingOrProcessing = generatedVideos.value.some(
      v => v.status === 'pending' || v.status === 'processing'
    )
    if (hasPendingOrProcessing) {
      startVideoPolling()
    }
  } catch (error: any) {
    console.error('加载视频列表失败:', error)
  } finally {
    loadingVideos.value = false
  }
}

// 启动视频状态轮询
const startVideoPolling = () => {
  if (videoPollingTimer) return

  videoPollingTimer = setInterval(async () => {
    if (!currentStoryboard.value) {
      stopVideoPolling()
      return
    }

    try {
      const result = await videoAPI.listVideos({
        storyboard_id: currentStoryboard.value.id.toString(),
        page: 1,
        page_size: 50
      })
      generatedVideos.value = (result.items || []).map(normalizeVideo)

      // 如果没有进行中的任务，停止轮询
      const hasPendingOrProcessing = generatedVideos.value.some(
        v => v.status === 'pending' || v.status === 'processing'
      )
      if (!hasPendingOrProcessing) {
        stopVideoPolling()
      }
    } catch (error) {
      console.error('轮询视频状态失败:', error)
    }
  }, 5000) // 每5秒轮询一次
}

// 停止视频轮询
const stopVideoPolling = () => {
  if (videoPollingTimer) {
    clearInterval(videoPollingTimer)
    videoPollingTimer = null
  }
}

const toggleCharacterInShot = async (charId: number) => {
  if (!currentStoryboard.value) return

  // 初始化characters数组
  if (!currentStoryboard.value.characters) {
    currentStoryboard.value.characters = []
  }

  const char = characters.value.find(c => c.id === charId)
  if (!char) return

  // 检查是否已存在
  const existIndex = currentStoryboard.value.characters.findIndex(c =>
    typeof c === 'object' ? c.id === charId : c === charId
  )

  if (existIndex > -1) {
    // 移除角色
    currentStoryboard.value.characters.splice(existIndex, 1)
  } else {
    // 添加角色（作为对象）
    currentStoryboard.value.characters.push(char)
  }

  // 保存到后端
  try {
    const characterIds = currentStoryboard.value.characters.map(c =>
      typeof c === 'object' ? c.id : c
    )

    await dramaAPI.updateStoryboard(currentStoryboard.value.id.toString(), {
      character_ids: characterIds
    })

    if (existIndex > -1) {
      ElMessage.success(`已移除角色: ${char.name}`)
    } else {
      ElMessage.success(`已添加角色: ${char.name}`)
      showCharacterSelector.value = false
    }
  } catch (error: any) {
    ElMessage.error('保存失败: ' + (error.message || '未知错误'))
    // 回滚操作
    if (existIndex > -1) {
      currentStoryboard.value.characters.push(char)
    } else {
      currentStoryboard.value.characters.splice(currentStoryboard.value.characters.length - 1, 1)
    }
  }
}

const removeCharacterFromShot = async (charId: number) => {
  if (!currentStoryboard.value) return

  // 初始化characters数组
  if (!currentStoryboard.value.characters) {
    currentStoryboard.value.characters = []
  }

  const char = characters.value.find(c => c.id === charId)
  if (!char) return

  // 检查是否已存在
  const existIndex = currentStoryboard.value.characters.findIndex(c =>
    typeof c === 'object' ? c.id === charId : c === charId
  )

  if (existIndex > -1) {
    // 移除角色
    currentStoryboard.value.characters.splice(existIndex, 1)
  }

  // 保存到后端
  try {
    const characterIds = currentStoryboard.value.characters.map(c =>
      typeof c === 'object' ? c.id : c
    )

    await dramaAPI.updateStoryboard(currentStoryboard.value.id.toString(), {
      character_ids: characterIds
    })

    ElMessage.success(`已移除角色: ${char.name}`)
  } catch (error: any) {
    ElMessage.error('保存失败: ' + (error.message || '未知错误'))
    // 回滚操作
    currentStoryboard.value.characters.push(char)
  }
}

const loadData = async () => {
  try {
    // 加载剧集信息
    const dramaRes = await dramaAPI.get(dramaId.toString())
    drama.value = dramaRes

    // 找到当前章节
    const ep = dramaRes.episodes?.find(e => e.episode_number === episodeNumber)
    if (!ep) {
      ElMessage.error('章节不存在')
      router.back()
      return
    }

    episode.value = ep
    episodeId.value = ep.id

    // 加载分镜列表
    const storyboardsRes = await dramaAPI.getStoryboards(ep.id.toString())

    // API返回格式: {storyboards: [...], total: number}
    storyboards.value = storyboardsRes?.storyboards || []

    // 默认选中第一个分镜
    if (storyboards.value.length > 0 && !currentStoryboardId.value) {
      currentStoryboardId.value = storyboards.value[0].id
    }

    // 加载角色列表
    characters.value = dramaRes.characters || []

    // 加载可用场景列表（从后台获取场景/背景列表）
    try {
      const scenes = await dramaAPI.getBackgrounds(ep.id.toString())
      availableScenes.value = scenes || []
    } catch (sceneError: any) {
      console.warn('加载场景列表失败，回退到剧集数据:', sceneError)
      availableScenes.value = dramaRes.scenes || []
    }

    // 加载视频素材库
    await loadVideoAssets()

    // 加载音频素材库
    await loadAudioAssets()
    await loadDouyinMusic()

  } catch (error: any) {
    ElMessage.error('加载数据失败: ' + (error.message || '未知错误'))
  }
}

const selectScene = async (sceneId: number) => {
  if (!currentStoryboard.value) return

  try {
    // TODO: 调用API更新分镜的scene_id
    await dramaAPI.updateScene(currentStoryboard.value.id.toString(), {
      scene_id: sceneId
    })

    // 重新加载数据
    await loadData()
    showSceneSelector.value = false
    ElMessage.success('场景关联成功')
  } catch (error: any) {
    ElMessage.error(error.message || '场景关联失败')
  }
}

const selectStoryboard = (id: string | number) => {
  currentStoryboardId.value = String(id)
}

const refreshCurrentStoryboardMedia = async () => {
  if (!currentStoryboard.value) return
  const storyboardId = Number(currentStoryboard.value.id)
  await Promise.all([
    loadVideoReferenceImages(storyboardId),
    loadStoryboardVideos(storyboardId)
  ])
}

const handleTimelineSelect = (sceneId: number) => {
  selectStoryboard(String(sceneId))
}

const handleAddStoryboard = async () => {
  ElMessage.info('添加分镜功能开发中')
}

const togglePlay = () => {
  if (currentPlayState.value === 'playing') {
    currentPlayState.value = 'paused'
  } else {
    currentPlayState.value = 'playing'
  }
}

const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const zoomIn = () => {
  ElMessage.info('时间线缩放功能开发中')
}

const zoomOut = () => {
  ElMessage.info('时间线缩放功能开发中')
}

const uploadInputRef = ref<HTMLInputElement | null>(null)
const uploadingImage = ref(false)

const generateImage = async () => {
  if (!currentStoryboard.value) return

  try {
    ElMessage.info('图片生成功能开发中')
  } catch (error: any) {
    ElMessage.error(error.message || '生成失败')
  }
}

const uploadImage = () => {
  uploadInputRef.value?.click()
}

const handleUploadImage = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  if (!currentStoryboard.value) {
    ElMessage.warning('请先选择分镜')
    input.value = ''
    return
  }

  uploadingImage.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)

    const uploadResp = await fetch('/api/v1/upload/image', {
      method: 'POST',
      body: formData
    })
    const uploadJson = await uploadResp.json()
    if (!uploadResp.ok || uploadJson?.success === false) {
      throw new Error(uploadJson?.error?.message || '上传失败')
    }
    const imageUrl = uploadJson?.data?.url || uploadJson?.url
    if (!imageUrl) {
      throw new Error('上传失败：未返回图片地址')
    }

    const record = await imageAPI.createImageRecord({
      drama_id: dramaId.toString(),
      storyboard_id: Number(currentStoryboard.value.id),
      frame_type: selectedFrameType.value,
      image_type: 'storyboard',
      image_url: imageUrl,
      prompt: currentFramePrompt.value || '手动上传'
    })

    generatedImages.value.unshift(record)
    videoReferenceImages.value.unshift(record)
    if (record.frame_type) {
      selectedVideoFrameType.value = record.frame_type as FrameType
    }
    ElMessage.success('上传成功')
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
  } finally {
    uploadingImage.value = false
    input.value = ''
  }
}

const goBack = () => {
  router.replace({
    name: 'EpisodeWorkflowNew',
    params: { id: dramaId, episodeNumber }
  })
}

const parseDistributionHashtags = (input: string): string[] => {
  if (!input.trim()) {
    return []
  }

  const unique = new Set<string>()
  input
    .split(/[\s,，]+/)
    .map(item => item.trim().replace(/^#+/, ''))
    .filter(Boolean)
    .forEach((tag) => {
      if (unique.size < 20) {
        unique.add(tag)
      }
    })

  return Array.from(unique)
}

const getPlatformLabel = (platform: string) => {
  const matched = distributionPlatforms.find(item => item.value === platform)
  return matched?.label || platform
}

const getDistributionBinding = (platform: DistributionPlatform) => {
  return socialBindings.value[platform]
}

const isDistributionPlatformBound = (platform: DistributionPlatform) => {
  return !!getDistributionBinding(platform)
}

const getDistributionBindingLabel = (platform: DistributionPlatform) => {
  const binding = getDistributionBinding(platform)
  if (!binding) {
    return ''
  }
  return binding.display_name?.trim() || binding.account_identifier
}

const syncSelectedDistributionPlatforms = () => {
  const boundPlatforms = distributionPlatforms
    .map(item => item.value)
    .filter(platform => isDistributionPlatformBound(platform))

  if (!distributionForm.value.platforms.length) {
    distributionForm.value.platforms = [...boundPlatforms]
    return
  }

  distributionForm.value.platforms = distributionForm.value.platforms.filter(platform =>
    boundPlatforms.includes(platform as ManagedDistributionPlatform)
  )
}

const loadSocialBindings = async () => {
  loadingSocialBindings.value = true
  try {
    const bindings = await socialBindingAPI.listBindings()
    socialBindings.value = bindings.reduce<Record<string, SocialAccountBinding>>((acc, item) => {
      acc[item.platform] = item
      return acc
    }, {})
    syncSelectedDistributionPlatforms()
  } catch (error) {
    console.error('加载社交平台绑定失败:', error)
    socialBindings.value = {}
    syncSelectedDistributionPlatforms()
  } finally {
    loadingSocialBindings.value = false
  }
}

const toggleDistributionPlatform = (platform: DistributionPlatform, checked: boolean) => {
  if (!isDistributionPlatformBound(platform)) {
    return
  }

  const current = new Set(distributionForm.value.platforms)
  if (checked) {
    current.add(platform)
  } else {
    current.delete(platform)
  }
  distributionForm.value.platforms = Array.from(current)
}

const openBindingPrompt = async (platform: { value: ManagedDistributionPlatform; label: string; placeholder: string }) => {
  bindingPlatformLoading.value = {
    ...bindingPlatformLoading.value,
    [platform.value]: true
  }

  try {
    const existingBinding = getDistributionBinding(platform.value)
    const actionText = existingBinding ? '更新绑定' : '绑定账号'
    const { value } = await ElMessageBox.prompt(
      `请输入 ${platform.label} 的账号标识，绑定后仅当前身份可用于分发。`,
      `${platform.label}${actionText}`,
      {
        confirmButtonText: '保存',
        cancelButtonText: '取消',
        inputValue: existingBinding?.account_identifier || '',
        inputPlaceholder: platform.placeholder,
        inputPattern: /\S+/,
        inputErrorMessage: '请输入有效的账号标识'
      }
    )

    const binding = await socialBindingAPI.upsertBinding(platform.value, {
      account_identifier: value.trim()
    })
    socialBindings.value = {
      ...socialBindings.value,
      [platform.value]: binding
    }
    if (!distributionForm.value.platforms.includes(platform.value)) {
      distributionForm.value.platforms = [...distributionForm.value.platforms, platform.value]
    }
    ElMessage.success(`${platform.label} 已绑定到当前身份`)
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error.message || '绑定失败')
    }
  } finally {
    bindingPlatformLoading.value = {
      ...bindingPlatformLoading.value,
      [platform.value]: false
    }
  }
}

const getDistributionStatusText = (status: VideoDistributionStatus) => {
  switch (status) {
    case 'pending':
      return '待分发'
    case 'processing':
      return '分发中'
    case 'published':
      return '已发布'
    case 'failed':
      return '失败'
    default:
      return status
  }
}

const getDistributionStatusType = (status: VideoDistributionStatus) => {
  switch (status) {
    case 'published':
      return 'success'
    case 'failed':
      return 'danger'
    case 'pending':
    case 'processing':
      return 'warning'
    default:
      return 'info'
  }
}

const getMergeDistributionSummary = (mergeId: number): VideoDistribution[] => {
  const distributions = mergeDistributions.value[mergeId] || []
  if (!distributions.length) {
    return []
  }

  const latestByPlatform = new Map<string, VideoDistribution>()
  distributions.forEach((item) => {
    if (!latestByPlatform.has(item.platform)) {
      latestByPlatform.set(item.platform, item)
    }
  })

  const order: DistributionPlatform[] = ['discord', 'reddit', 'pinterest', 'tiktok', 'youtube', 'instagram', 'x']
  return Array.from(latestByPlatform.values()).sort(
    (a, b) => order.indexOf(a.platform) - order.indexOf(b.platform)
  )
}

const resetDistributionForm = () => {
  distributionForm.value = {
    mergeId: 0,
    platforms: [],
    title: '',
    description: '',
    hashtagsText: ''
  }
}

const stopDistributionPolling = () => {
  if (distributionPollingTimer) {
    clearInterval(distributionPollingTimer)
    distributionPollingTimer = null
  }
}

const loadMergeDistributions = async (merges: VideoMerge[]) => {
  const completed = merges.filter((merge) => merge.status === 'completed')
  if (completed.length === 0) {
    mergeDistributions.value = {}
    stopDistributionPolling()
    return
  }

  const entries = await Promise.all(
    completed.map(async (merge) => {
      try {
        const distributions = await videoMergeAPI.listDistributions(merge.id)
        return [merge.id, distributions] as const
      } catch (error) {
        console.error('加载分发记录失败:', error)
        return [merge.id, [] as VideoDistribution[]] as const
      }
    })
  )

  const nextMap: Record<number, VideoDistribution[]> = {}
  let hasProcessing = false
  entries.forEach(([mergeId, distributions]) => {
    nextMap[mergeId] = distributions
    if (distributions.some(item => item.status === 'pending' || item.status === 'processing')) {
      hasProcessing = true
    }
  })
  mergeDistributions.value = nextMap

  if (hasProcessing) {
    startDistributionPolling()
  } else {
    stopDistributionPolling()
  }
}

const startDistributionPolling = () => {
  if (distributionPollingTimer) return

  distributionPollingTimer = setInterval(async () => {
    const merges = videoMerges.value.filter(merge => merge.status === 'completed')
    if (!merges.length) {
      stopDistributionPolling()
      return
    }

    await loadMergeDistributions(videoMerges.value)
  }, 3000)
}

const getDefaultDistributionTitle = (merge: VideoMerge) => {
  const dramaTitle = drama.value?.title?.trim()
  if (dramaTitle) {
    return dramaTitle
  }

  const mergeTitle = (merge.title || '').trim()
  if (!mergeTitle) {
    return '短剧'
  }

  return mergeTitle.replace(/\s*[-·•|｜]\s*第[\d一二三四五六七八九十百千万]+集\s*$/u, '').trim() || mergeTitle
}

const openDistributionDialog = (merge: VideoMerge) => {
  distributionTargetMerge.value = merge
  distributionDialogVisible.value = true
}

const submitDistribution = async () => {
  if (!distributionForm.value.mergeId) {
    ElMessage.warning('请选择需要分发的视频')
    return
  }

  if (!distributionForm.value.platforms.length) {
    ElMessage.warning('请至少选择一个分发平台')
    return
  }

  submittingDistribution.value = true
  try {
    const distributions = await videoMergeAPI.distributeVideo(distributionForm.value.mergeId, {
      platforms: distributionForm.value.platforms,
      title: distributionForm.value.title.trim(),
      description: distributionForm.value.description.trim(),
      hashtags: parseDistributionHashtags(distributionForm.value.hashtagsText)
    })

    mergeDistributions.value = {
      ...mergeDistributions.value,
      [distributionForm.value.mergeId]: distributions
    }

    ElMessage.success('分发任务已提交')
    distributionDialogVisible.value = false
    distributionTargetMerge.value = null
    resetDistributionForm()
    startDistributionPolling()
  } catch (error: any) {
    ElMessage.error(error.message || '分发失败')
  } finally {
    submittingDistribution.value = false
  }
}

const openDistributionRecord = (distribution: VideoDistribution) => {
  if (!distribution.published_url) {
    ElMessage.warning('该分发记录暂无可打开的发布链接')
    return
  }
  window.open(distribution.published_url, '_blank')
}

// 加载视频合成列表
const loadVideoMerges = async () => {
  if (!episodeId.value) {
    mergeListInitialized.value = true
    return
  }

  try {
    loadingMerges.value = true
    const result = await videoMergeAPI.listMerges({
      episode_id: episodeId.value.toString(),
      page: 1,
      page_size: 20
    })
    videoMerges.value = result.merges
    await loadMergeDistributions(result.merges)

    // 检查是否有进行中的任务
    const hasProcessingTasks = result.merges.some(
      (merge: any) => merge.status === 'pending' || merge.status === 'processing'
    )

    if (hasProcessingTasks) {
      startMergePolling()
    } else {
      stopMergePolling()
    }
  } catch (error: any) {
    console.error('加载视频合成列表失败:', error)
    ElMessage.error('加载视频合成列表失败')
  } finally {
    loadingMerges.value = false
    mergeListInitialized.value = true
  }
}

// 启动视频合成列表轮询
const startMergePolling = () => {
  if (mergePollingTimer) return

  mergePollingTimer = setInterval(async () => {
    if (!episodeId.value) {
      stopMergePolling()
      return
    }

    try {
      const result = await videoMergeAPI.listMerges({
        episode_id: episodeId.value.toString(),
        page: 1,
        page_size: 20
      })
      videoMerges.value = result.merges
      await loadMergeDistributions(result.merges)

      // 检查是否还有进行中的任务
      const hasProcessingTasks = result.merges.some(
        (merge: any) => merge.status === 'pending' || merge.status === 'processing'
      )

      if (!hasProcessingTasks) {
        stopMergePolling()
      }
    } catch (error) {
    }
  }, 3000) // 每3秒轮询一次
}

// 停止视频合成列表轮询
const stopMergePolling = () => {
  if (mergePollingTimer) {
    clearInterval(mergePollingTimer)
    mergePollingTimer = null
  }
}

// 处理视频合成完成事件
const handleMergeCompleted = async (mergeId: number) => {
  // 刷新视频合成列表
  await loadVideoMerges()
  // 切换到视频合成标签页
  activeTab.value = 'merges'
}

// 下载视频
const downloadVideo = async (url: string, title: string) => {
  let loadingMsg: { close: () => void } | null = null

  try {
    const resolvedUrl = resolveVideoUrl(url)
    if (!resolvedUrl) {
      ElMessage.warning('暂无可下载视频')
      return
    }

    loadingMsg = ElMessage.info({
      message: '正在准备下载...',
      duration: 0
    })

    // 使用fetch获取视频blob
    const response = await fetch(resolvedUrl)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const blob = await response.blob()
    const blobUrl = window.URL.createObjectURL(blob)

    // 创建下载链接
    const link = document.createElement('a')
    link.href = blobUrl
    link.download = `${title}.mp4`
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()

    // 清理
    setTimeout(() => {
      document.body.removeChild(link)
      window.URL.revokeObjectURL(blobUrl)
    }, 100)

    ElMessage.success('视频下载已开始')
  } catch (error) {
    console.error('下载视频失败:', error)
    ElMessage.error('视频下载失败，请稍后重试')
  } finally {
    loadingMsg?.close()
  }
}

// 预览合成视频
const previewMergedVideo = (url?: string | null) => {
  const resolvedUrl = resolveVideoUrl(url)
  if (!resolvedUrl) {
    ElMessage.warning('暂无可预览视频')
    return
  }

  window.open(resolvedUrl, '_blank', 'noopener,noreferrer')
}

// 删除视频合成记录
const deleteMerge = async (mergeId: number) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除此合成记录吗？此操作不可恢复。',
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await videoMergeAPI.deleteMerge(mergeId)
    ElMessage.success('删除成功')
    // 刷新列表
    await loadVideoMerges()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

// 格式化日期时间
const formatDateTime = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`

  // 超过7天显示完整日期
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')
  return `${month}-${day} ${hour}:${minute}`
}

onMounted(async () => {
  await loadData()
  await loadVideoModels()
  await loadVideoMerges()
})

onActivated(async () => {
  await loadData()
  await refreshCurrentStoryboardMedia()
})

// 组件卸载时停止轮询
onBeforeUnmount(() => {
  stopPolling()
  stopVideoPolling()
  stopMergePolling()
  stopDistributionPolling()
  unbindHotMusicScroll()
  stopAudioPreview()
})
</script>

<style scoped lang="scss">
// 镜头列表项样式
.storyboard-item {
  padding: 8px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
  border: 1px solid var(--border-primary);
  margin-bottom: 8px;
  display: flex;
  gap: 10px;
  align-items: center;
  background: var(--bg-card);

  &:hover {
    background: var(--bg-card-hover);
    border-color: var(--border-secondary);
  }

  &.active {
    background: var(--accent);
    border-color: var(--accent);

    .shot-number,
    .shot-title {
      color: var(--text-inverse) !important;
    }

    .shot-duration {
      background: rgba(255, 255, 255, 0.2);
      color: var(--text-inverse);
    }
  }

  .shot-thumbnail {
    width: 80px;
    height: 50px;
    border-radius: 4px;
    overflow: hidden;
    background: var(--bg-secondary);
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .shot-content {
    flex: 1;
    min-width: 0;

    .shot-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 4px;

      .shot-number {
        font-size: 11px;
        color: var(--text-secondary);
        font-weight: 500;
      }

      .shot-duration {
        font-size: 11px;
        color: var(--text-secondary);
        background: var(--bg-secondary);
        padding: 2px 6px;
        border-radius: 3px;
      }
    }

    .shot-title {
      font-size: 13px;
      color: var(--text-primary);
      font-weight: 500;
      line-height: 1.3;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

// 视频合成列表样式
.merges-list {
  padding: 16px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  background: var(--bg-secondary);

  .merge-items {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .merge-item {
    position: relative;
    background: var(--bg-card);
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--border-primary);

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
      border-color: var(--accent);
    }

    .status-indicator {
      position: absolute;
      left: 0;
      top: 0;
      bottom: 0;
      width: 4px;
      transition: all 0.3s;
    }

    &.merge-status-completed .status-indicator {
      background: linear-gradient(to bottom, #67c23a, #85ce61);
    }

    &.merge-status-processing .status-indicator {
      background: linear-gradient(to bottom, #e6a23c, #f0c78a);
      animation: pulse 2s ease-in-out infinite;
    }

    &.merge-status-failed .status-indicator {
      background: linear-gradient(to bottom, #f56c6c, #f89898);
    }

    &.merge-status-pending .status-indicator {
      background: linear-gradient(to bottom, #909399, #b1b3b8);
    }

    .merge-content {
      padding: 20px 24px;
      padding-left: 28px;
    }

    .merge-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      padding-bottom: 14px;
      border-bottom: 1px solid var(--border-primary);

      .title-section {
        display: flex;
        align-items: center;
        gap: 12px;
        flex: 1;

        .title-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 38px;
          height: 38px;
          border-radius: 10px;
          background: var(--bg-secondary);
          color: var(--text-secondary);
          transition: all 0.3s;
        }

        .merge-title {
          margin: 0;
          font-size: 15px;
          font-weight: 500;
          color: var(--text-secondary);
          line-height: 1.4;
        }
      }

      :deep(.el-tag) {
        font-weight: 500;
        padding: 4px 12px;
        font-size: 12px;
      }
    }

    &.merge-status-completed .title-icon {
      background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
      color: #fff;
    }

    &.merge-status-processing .title-icon {
      background: linear-gradient(135deg, #e6a23c 0%, #f0c78a 100%);
      color: #fff;
    }

    &.merge-status-failed .title-icon {
      background: linear-gradient(135deg, #f56c6c 0%, #f89898 100%);
      color: #fff;
    }

    .merge-details {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
      margin-bottom: 16px;

      .detail-item {
        display: flex;
        gap: 10px;
        padding: 12px 14px;
        background: var(--bg-secondary);
        border-radius: 8px;
        border: 1px solid var(--border-primary);
        transition: all 0.3s;

        &:hover {
          border-color: var(--accent);
          transform: translateY(-1px);
        }

        .detail-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 28px;
          height: 28px;
          border-radius: 6px;
          background: var(--bg-card);
          color: var(--accent);
          flex-shrink: 0;
        }

        .detail-content {
          flex: 1;
          min-width: 0;

          .detail-label {
            font-size: 11px;
            color: var(--text-muted);
            margin-bottom: 3px;
            font-weight: 500;
          }

          .detail-value {
            font-size: 13px;
            color: var(--text-primary);
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }
        }
      }
    }

    .merge-error {
      margin-bottom: 12px;

      :deep(.el-alert) {
        border-radius: 8px;
        border: none;
        padding: 8px 12px;
        font-size: 12px;
      }
    }

    .merge-actions {
      display: flex;
      gap: 8px;
      margin-top: 12px;

      :deep(.el-button) {
        flex: 1;
        max-width: 160px;
        font-weight: 500;
        padding: 8px 15px;
        font-size: 13px;
      }
    }
  }
}

// 旋转动画
@keyframes rotating {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

.rotating {
  animation: rotating 2s linear infinite;
}

// 脉冲动画
@keyframes pulse {

  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.6;
  }
}

// 白色主题样式
.shot-editor-new {
  padding: 16px;
  height: 100%;
  overflow-y: auto;
  // background: #fff;

  .section-label {
    font-size: 12px;
    color: #666;
    margin-bottom: 8px;
  }

  // 场景预览
  .scene-section {
    margin-bottom: 20px;
  }

  .scene-preview {
    width: 100%;
    height: 80px;
    border-radius: 6px;
    overflow: hidden;
    position: relative;
    background: #f5f5f5;
    border: 1px solid var(--border-primary);

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .scene-info {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      padding: 6px 8px;
      background: linear-gradient(to top, rgba(0, 0, 0, 0.7), transparent);
      font-size: 11px;
      color: #fff;

      .scene-id {
        font-size: 10px;
        color: #e0e0e0;
        margin-top: 2px;
      }
    }
  }

  .scene-preview-empty {
    width: 100%;
    height: 80px;
    border-radius: 6px;
    border: 1px dashed var(--border-secondary);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: var(--bg-soft);

    .el-icon {
      font-size: 32px !important;
      color: var(--text-muted);
    }

    div {
      font-size: 11px;
      color: var(--text-muted);
    }
  }

  // 角色列表
  .cast-section {
    margin-bottom: 20px;
  }

  .cast-list {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 8px;

    .cast-item {
      position: relative;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 4px;
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        .cast-avatar {
          border-color: #409eff;
        }

        .cast-remove {
          opacity: 1;
          visibility: visible;
        }
      }

      &.active {
        .cast-avatar {
          border-color: #409eff;
          background: #409eff;
        }
      }

      .cast-avatar {
        width: 36px;
        height: 36px;
        border-radius: 50%;
        border: 2px solid #e0e0e0;
        overflow: hidden;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #f5f5f5;
        font-size: 14px;
        font-weight: 500;
        color: #666;
        transition: all 0.2s;

        img {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }
      }

      .cast-name {
        font-size: 10px;
        color: #666;
        max-width: 36px;
        text-align: center;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .cast-remove {
        position: absolute;
        top: -3px;
        right: -3px;
        width: 16px;
        height: 16px;
        border-radius: 50%;
        background: #f56c6c;
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        transition: all 0.2s;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        z-index: 10;
        opacity: 0;
        visibility: hidden;
        font-size: 12px;

        &:hover {
          background: #f23030;
          transform: scale(1.1);
        }
      }
    }

    .cast-empty {
      width: 100%;
      text-align: center;
      padding: 15px;
      color: var(--text-muted);
      font-size: 11px;
    }
  }

  // 视效设置
  .settings-section {
    margin-bottom: 16px;

    .settings-grid {
      display: grid;
      grid-template-columns: 1fr 1fr 1fr;
      gap: 10px;

      .setting-item {
        label {
          display: block;
          font-size: 11px;
          color: var(--text-secondary);
          margin-bottom: 6px;
        }
      }
    }

    .audio-controls {
      margin-top: 8px;
    }
  }

  // 叙事内容
  .narrative-section {
    margin-bottom: 14px;
  }

  .dialogue-section {
    margin-bottom: 14px;
  }
}

.sound-music-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .sound-music-header {
    display: flex;
    flex-direction: column;
    gap: 4px;

    .sound-music-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--text-primary);
    }

    .sound-music-subtitle {
      font-size: 12px;
      color: var(--text-muted);
    }

    .sound-music-meta {
      font-size: 11px;
      color: var(--text-muted);
    }
  }

  .sound-music-filters {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;

    .audio-search {
      width: 200px;
    }

    .audio-search-loading {
      color: var(--text-muted);
      animation: rotating 1s linear infinite;
    }

    .audio-category {
      width: 140px;
    }

    .sfx-categories {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-wrap: wrap;
    }

    .sfx-ai-generate {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-wrap: wrap;
    }

    .sfx-ai-input {
      width: 220px;
    }
  }

    .audio-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
      position: relative;
    }

    .audio-search-status {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 6px 10px;
      border-radius: 6px;
      background: rgba(59, 130, 246, 0.08);
      color: var(--text-secondary);
      font-size: 12px;
      align-self: flex-start;
    }

  .audio-pagination {
    display: flex;
    justify-content: center;
    padding: 8px 0 4px;
    align-items: center;
    gap: 12px;
  }

  .audio-lazy-tip {
    font-size: 12px;
    color: var(--text-muted);
    text-align: center;
    padding: 6px 0 2px;
  }

  .audio-scroll-loading {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    align-self: center;
    padding: 6px 12px;
    border-radius: 6px;
    background: rgba(59, 130, 246, 0.08);
    color: var(--text-secondary);
    font-size: 12px;
  }

  .audio-total {
    font-size: 12px;
    color: var(--text-muted);
  }

  .audio-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .audio-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 12px;
    border-radius: 10px;
    border: 1px solid var(--border-primary);
    background: var(--bg-secondary);
    transition: all 0.2s;

    &:hover {
      border-color: var(--accent);
      box-shadow: var(--shadow-md);
      transform: translateY(-1px);
    }

    &.previewing,
    &.preview-loading {
      border-color: var(--accent);
      background: rgba(64, 158, 255, 0.08);
      box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.18), var(--shadow-md);
    }

    .audio-card-main {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .audio-icon {
      width: 44px;
      height: 44px;
      border-radius: 12px;
      background: rgba(124, 58, 237, 0.12);
      color: #7c3aed;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      flex-shrink: 0;
      overflow: hidden;
    }

    .audio-cover {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .audio-info {
      flex: 1;
      min-width: 0;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .audio-name {
      font-size: 13px;
      font-weight: 600;
      color: var(--text-primary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .audio-artist {
      font-size: 12px;
      color: var(--text-muted);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .audio-meta {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 11px;
      color: var(--text-muted);
      flex-wrap: wrap;
    }

    .sfx-categories {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .audio-duration,
    .audio-views {
      font-size: 11px;
      color: var(--text-muted);
    }

    .audio-actions {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }
  }
}

// 场景选择对话框样式
.scene-selector-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  max-height: 500px;
  overflow-y: auto;

  .scene-card {
    border: 2px solid var(--border-primary);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: var(--shadow-md);
    }

    &.selected {
      border-color: var(--accent);
      background: var(--accent-light);
    }

    .scene-image {
      width: 100%;
      height: 150px;
      background: var(--bg-secondary);
      display: flex;
      align-items: center;
      justify-content: center;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
    }

    .scene-info {
      padding: 12px;
      background: var(--bg-card);

      .scene-location {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 4px;
      }

      .scene-time {
        font-size: 12px;
        color: var(--text-muted);
      }
    }
  }

  .empty-scenes {
    grid-column: 1 / -1;
    padding: 40px 0;
  }
}

// 更新section-label样式以支持按钮
.section-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

// 角色选择对话框样式
.character-selector-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  max-height: 500px;
  overflow-y: auto;

  .character-card {
    position: relative;
    border: 2px solid var(--border-primary);
    border-radius: 8px;
    padding: 16px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;

    &:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: var(--shadow-md);
    }

    &.selected {
      border-color: var(--accent);
      background: var(--accent-light);
    }

    .character-avatar-large {
      width: 80px;
      height: 80px;
      border-radius: 50%;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--bg-secondary);
      font-size: 32px;
      font-weight: 600;
      color: var(--accent);

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
    }

    .character-info {
      text-align: center;

      .character-name {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 4px;
      }

      .character-role {
        font-size: 12px;
        color: var(--text-muted);
      }
    }

    .character-check {
      position: absolute;
      top: 8px;
      right: 8px;
    }
  }

  .empty-characters {
    grid-column: 1 / -1;
    padding: 40px 0;
  }
}

// 角色大图预览样式
.character-image-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;

  img {
    max-width: 100%;
    max-height: 500px;
    border-radius: 8px;
    object-fit: contain;
  }
}

// 场景大图预览样式
.scene-image-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 450px;
  background: var(--bg-secondary);
  border-radius: 8px;

  img {
    max-width: 100%;
    max-height: 600px;
    border-radius: 8px;
    object-fit: contain;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
}

// 设置部分样式
.settings-section {
  margin-bottom: 20px;

  .section-label {
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 12px;
  }

  .settings-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;

    .setting-item {
      label {
        display: block;
        font-size: 11px;
        color: var(--text-secondary);
        margin-bottom: 6px;
      }
    }
  }

  .audio-controls {
    :deep(.el-textarea__inner) {
      background: var(--bg-card);
      border-color: var(--border-primary);
      color: var(--text-primary);

      &::placeholder {
        color: var(--text-muted);
      }
    }

    :deep(.el-select) {
      width: 100%;
    }

    :deep(.el-slider__runway) {
      background: #e4e7ed;
    }

    :deep(.el-slider__bar) {
      background: #409eff;
    }

    :deep(.el-slider__button) {
      border-color: #409eff;
    }
  }
}

.professional-editor {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  color: var(--text-primary);

  .editor-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    background: var(--bg-card);
    border-bottom: 1px solid var(--border-primary);

    .toolbar-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .back-btn {
        color: var(--text-secondary);

        &:hover {
          color: var(--accent);
        }
      }

      .episode-title {
        font-size: 14px;
        color: var(--text-primary);
      }
    }

    .toolbar-right {
      display: flex;
      gap: 8px;
    }
  }

  .editor-main {
    flex: 1;
    display: flex;
    overflow: hidden;
    height: calc(100vh - 60px);

    .storyboard-panel {
      width: 280px;
      background: var(--bg-card);
      border-right: 1px solid var(--border-primary);
      display: flex;
      flex-direction: column;

      .panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px;
        border-bottom: 1px solid var(--border-primary);

        h3 {
          margin: 0;
          font-size: 16px;
          font-weight: 500;
        }
      }

      .storyboard-list {
        flex: 1;
        overflow-y: auto;
        padding: 8px;
        scrollbar-width: thin;
        scrollbar-color: var(--accent) var(--bg-secondary);

        &::-webkit-scrollbar {
          width: 8px;
        }

        &::-webkit-scrollbar-track {
          background: var(--bg-secondary);
          border-radius: 8px;
        }

        &::-webkit-scrollbar-thumb {
          background: var(--accent);
          border-radius: 8px;
        }

        .storyboard-item {
          display: flex;
          flex-direction: column;
          padding: 12px;
          margin-bottom: 8px;
          background: var(--bg-secondary);
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.2s;

          &:hover {
            background: var(--bg-card-hover);
          }

          &.active {
            background: var(--accent-light);
            border-left: 3px solid var(--accent);

            .shot-content {

              .shot-number,
              .shot-title {
                color: var(--accent) !important;
              }

              .shot-action {
                color: var(--text-primary) !important;
              }

              .shot-duration {
                background: var(--accent-light);
                color: var(--accent);
              }
            }
          }

          .shot-content {
            width: 100%;

            .shot-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 6px;
              gap: 8px;

              .shot-title-row {
                display: flex;
                align-items: baseline;
                gap: 8px;
                flex: 1;
                min-width: 0;

                .shot-number {
                  font-size: 12px;
                  font-weight: 600;
                  color: var(--text-secondary);
                  flex-shrink: 0;
                }

                .shot-title {
                  font-size: 13px;
                  font-weight: 500;
                  color: var(--text-primary);
                  overflow: hidden;
                  text-overflow: ellipsis;
                  white-space: nowrap;
                }
              }

              .shot-duration {
                font-size: 11px;
                color: var(--text-muted);
                background: var(--bg-card-hover);
                padding: 2px 8px;
                border-radius: 4px;
                flex-shrink: 0;
              }
            }

            .shot-action {
              font-size: 11px;
              color: var(--text-secondary);
              line-height: 1.5;
              overflow: hidden;
              text-overflow: ellipsis;
              display: -webkit-box;
              -webkit-line-clamp: 2;
              -webkit-box-orient: vertical;
            }
          }
        }
      }
    }

    .timeline-area {
      flex: 1;
      display: flex;
      flex-direction: column;
      background: var(--bg-secondary);
      overflow: hidden;

      .empty-timeline {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }

    .edit-panel {
      width: 520px;
      background: var(--bg-card);
      border-left: 1px solid var(--border-primary);
      overflow: hidden;
      flex-shrink: 0;

      .edit-tabs {
        height: 100%;

        :deep(.el-tabs__header) {
          margin: 0;
          background: var(--bg-secondary);
          padding: 0 16px;
          border-bottom: 1px solid var(--border-primary);
        }

        :deep(.el-tabs__content) {
          height: calc(100% - 55px);
          overflow-y: auto;
        }

        .tab-content {
          padding: 16px;
        }

        .sound-music-tab-content {
          padding-top: 0px;
        }

        .scene-editor,
        .shot-editor {
          .el-form-item {
            margin-bottom: 16px;
          }
        }
      }
    }
  }
}

// 通用参数行样式
.param-row {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;

  &:last-child {
    margin-bottom: 0;
  }
}

.param-label {
  min-width: 50px;
  font-size: 12px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

// 图片生成界面样式
.image-generation-section {
  .frame-type-selector {
    margin-bottom: 20px;

    .section-label {
      font-size: 13px;
      color: var(--text-primary);
      font-weight: 500;
      margin-bottom: 12px;
    }

    :deep(.el-radio-group) {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }

    .panel-count-input {
      width: 80px;
    }
  }

  .prompt-section {
    margin-bottom: 20px;

    .section-label {
      font-size: 13px;
      color: var(--text-primary);
      font-weight: 500;
      margin-bottom: 12px;
      display: flex;
      align-items: center;
    }

    :deep(.el-textarea__inner) {
      font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
      font-size: 12px;
      line-height: 1.6;
    }
  }

  .generation-controls {
    margin-bottom: 20px;
    display: flex;
    gap: 10px;
  }

  .generation-result {
    .section-label {
      font-size: 13px;
      color: var(--text-primary);
      font-weight: 600;
      margin-bottom: 12px;
      display: flex;
      align-items: center;
      gap: 6px;

      &::before {
        content: '';
        width: 3px;
        height: 14px;
        background: linear-gradient(to bottom, var(--accent), var(--accent-hover));
        border-radius: 2px;
      }
    }

    .image-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
      gap: 10px;

      &.image-grid--generated {
        grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
        gap: 12px;

        .image-item {
          aspect-ratio: 4 / 3;
          padding: 6px;
          background: linear-gradient(135deg, #f7f8fa 0%, #eef1f5 100%);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .image-item.has-image::before {
          content: '';
          position: absolute;
          inset: 0;
          background-image: var(--image-url);
          background-size: cover;
          background-position: center;
          filter: blur(12px) saturate(1.05);
          transform: scale(1.08);
          opacity: 0.5;
        }

        .image-placeholder {
          aspect-ratio: 4 / 3;
        }

        :deep(.el-image) {
          width: 100%;
          height: 100%;
          position: relative;
          z-index: 1;
          aspect-ratio: auto;
        }

        :deep(.el-image__inner) {
          object-fit: contain;
          background: transparent;
        }
      }

      .image-item {
        position: relative;
        border-radius: 8px;
        overflow: hidden;
        background: var(--bg-card);
        border: 1px solid var(--border-primary);
        transition: all 0.2s ease;
        cursor: pointer;
        box-shadow: var(--shadow-sm);

        &:hover {
          transform: translateY(-2px);
          box-shadow: var(--shadow-md);
          border-color: var(--accent);
        }

        :deep(.el-image) {
          width: 100%;
          aspect-ratio: 16 / 9;
          background: transparent;
          display: block;
          position: relative;
          z-index: 1;
        }

        :deep(.el-image__inner) {
          width: 100%;
          height: 100%;
          object-fit: cover;
          display: block;
        }

        .image-placeholder {
          width: 100%;
          aspect-ratio: 16 / 9;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 8px;
          background: var(--bg-secondary);
          color: var(--text-muted);
          position: relative;
          overflow: hidden;

          &::before {
            content: '';
            position: absolute;
            width: 200%;
            height: 200%;
            background: linear-gradient(45deg,
                transparent 30%,
                rgba(255, 255, 255, 0.3) 50%,
                transparent 70%);
            animation: shimmer 2s infinite;
            top: -50%;
            left: -50%;
          }

          .el-icon {
            position: relative;
            z-index: 1;
            font-size: 24px !important;
          }

          p {
            margin: 0;
            font-size: 11px;
            font-weight: 500;
            position: relative;
            z-index: 1;
          }
        }

        .image-info {
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          padding: 6px 8px;
          background: linear-gradient(to top, rgba(0, 0, 0, 0.75), rgba(0, 0, 0, 0.2) 70%, transparent);
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 4px;
          z-index: 2;

          :deep(.el-tag) {
            backdrop-filter: blur(8px);
            font-size: 10px;
            height: 20px;
            padding: 0 6px;
          }

          .frame-type-tag {
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 10px;
            font-weight: 500;
            background: rgba(255, 255, 255, 0.25);
            color: white;
            backdrop-filter: blur(8px);
            border: 1px solid rgba(255, 255, 255, 0.3);
            text-transform: uppercase;
            letter-spacing: 0.3px;
          }
        }
      }
    }

    @keyframes shimmer {
      0% {
        transform: translateX(-100%) translateY(-100%) rotate(45deg);
      }

      100% {
        transform: translateX(100%) translateY(100%) rotate(45deg);
      }
    }
  }



  .panel-count-label {
    margin-left: 5px;
    font-size: 12px;
    color: var(--text-muted);
  }

  .model-tags {
    font-size: 12px;
    color: var(--text-muted);
  }

  .mode-description {
    font-size: 12px;
    color: var(--text-muted);
  }



}

// 视频生成样式
.video-generation-section {
  .section-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 12px;
    padding-left: 8px;
    border-left: 3px solid var(--accent);
  }

  // 视频生成结果样式
  .generation-result {
    margin-top: 24px;

    .section-label {
      font-size: 13px;
      color: var(--text-primary);
      font-weight: 600;
      margin-bottom: 12px;
      display: flex;
      align-items: center;
      gap: 6px;

      // &::before {
      //   content: '';
      //   width: 3px;
      //   height: 14px;
      //   background: linear-gradient(to bottom, #409eff, #66b1ff);
      //   border-radius: 2px;
      // }
    }

    .image-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
      gap: 10px;

      .image-item {
        position: relative;
        border-radius: 8px;
        overflow: hidden;
        background: var(--bg-card);
        border: 1px solid var(--border-primary);
        transition: all 0.2s ease;
        cursor: pointer;
        box-shadow: var(--shadow-card);

        &:hover {
          transform: translateY(-2px);
          box-shadow: var(--shadow-card-hover);
          border-color: var(--accent);
        }

        .image-placeholder {
          width: 100%;
          aspect-ratio: 16 / 9;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 8px;
          background: linear-gradient(135deg, var(--bg-soft) 0%, var(--bg-secondary) 100%);
          color: var(--text-muted);
          position: relative;
          overflow: hidden;

          &::before {
            content: '';
            position: absolute;
            width: 200%;
            height: 200%;
            background: linear-gradient(45deg,
                transparent 30%,
                rgba(255, 255, 255, 0.3) 50%,
                transparent 70%);
            animation: shimmer 2s infinite;
            top: -50%;
            left: -50%;
          }

          .el-icon {
            position: relative;
            z-index: 1;
            font-size: 24px !important;
          }

          p {
            margin: 0;
            font-size: 11px;
            font-weight: 500;
            position: relative;
            z-index: 1;
          }
        }

        .image-info {
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          padding: 6px 8px;
          background: linear-gradient(to top, rgba(0, 0, 0, 0.75), rgba(0, 0, 0, 0.2) 70%, transparent);
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 4px;

          :deep(.el-tag) {
            backdrop-filter: blur(8px);
            font-size: 10px;
            height: 20px;
            padding: 0 6px;
          }

          .frame-type-tag {
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 10px;
            font-weight: 500;
            background: rgba(255, 255, 255, 0.25);
            color: white;
            backdrop-filter: blur(8px);
            border: 1px solid rgba(255, 255, 255, 0.3);
            text-transform: uppercase;
            letter-spacing: 0.3px;
          }
        }

        // 视频缩略图特殊样式
        &.video-item .video-thumbnail {
          position: relative;
          width: 100%;
          aspect-ratio: 16 / 9;
          overflow: hidden;
          cursor: pointer;

          video {
            width: 100%;
            height: 100%;
            object-fit: cover;
            display: block;
            pointer-events: none;
          }

          .play-overlay {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            background: rgba(0, 0, 0, 0.3);
            opacity: 0;
            transition: opacity 0.2s ease;

            .el-icon {
              filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.3));
            }
          }

          &:hover .play-overlay {
            opacity: 1;
          }
        }
      }
    }
  }

  .reference-mode-title {
    margin-bottom: 12px;
    font-size: 13px;
    color: var(--text-primary);
    font-weight: 500;
  }

  .frame-label {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--text-muted);
  }

  .slot-hint {
    margin-top: 8px;
    font-size: 12px;
    color: var(--text-muted);
  }

  .image-slot {
    position: relative;
    width: 140px;
    height: 90px;
    border: 2px dashed var(--border-primary);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    background: var(--bg-card);
    // display: flex;
    // align-items: center;
    // justify-content: center;

    &:hover {
      border-color: var(--accent);
    }
  }

  .video-params-section {
    margin-bottom: 16px;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px solid var(--border-primary);
  }

  .image-slots-container {
    padding: 12px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px dashed var(--border-primary);
  }

  .image-slot {
    position: relative;
    width: 140px;
    height: 90px;
    border: 2px dashed var(--border-primary);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.3s;
    background: var(--bg-card);

    &:hover {
      border-color: var(--accent);
      box-shadow: var(--shadow-md);
    }

    &.image-slot-small {
      width: 80px;
      height: 52px;
    }
  }

  .image-slot-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
  }

  .image-slot-index {
    margin-top: 4px;
    font-size: 10px;
    color: var(--text-muted);
  }

  .image-slot-remove {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 24px;
    height: 24px;
    background: rgba(0, 0, 0, 0.6);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background: rgba(255, 73, 73, 0.9);
      transform: scale(1.1);
    }
  }
  .reference-images-section {
    margin-top: 12px;


    .frame-type-buttons {
      margin-bottom: 12px;
      text-align: center;

      :deep(.el-radio-group) {
        display: inline-flex;
        flex-wrap: wrap;
        gap: 0;
      }

      :deep(.el-radio-button) {
        overflow: hidden;

        &:first-child .el-radio-button__inner {
          border-radius: 6px 0 0 6px;
        }

        &:last-child .el-radio-button__inner {
          border-radius: 0 6px 6px 0;
        }
      }

      :deep(.el-radio-button__inner) {
        padding: 6px 12px;
        font-size: 12px;
        font-weight: 500;
        border-color: var(--border-primary);
        transition: all 0.2s;

        &:hover {
          // color: var(--accent);
          border-color: var(--accent);
        }
      }

      :deep(.el-radio-button.is-active .el-radio-button__inner) {
        background: var(--accent);
        border-color: var(--accent);
        box-shadow: 0 2px 6px rgba(14, 165, 233, 0.25);
      }
    }

    .frame-type-content {
      padding: 4px 10px;
      background: var(--bg-card);
      border-radius: 8px;
      border: 1px solid var(--border-primary);
      min-height: 160px;
    }

    .image-scroll-container {
      max-height: 220px;
      overflow-y: auto;
      overflow-x: hidden;
      padding-right: 4px;

      &::-webkit-scrollbar {
        width: 6px;
      }

      &::-webkit-scrollbar-track {
        background: #f1f1f1;
        border-radius: 3px;
      }

      &::-webkit-scrollbar-thumb {
        background: #c1c1c1;
        border-radius: 3px;

        &:hover {
          background: #a8a8a8;
        }
      }
    }

    .previous-frame-section {
      margin-bottom: 12px;
      padding: 8px;
      background: var(--bg-secondary);
      border: 1px solid var(--border-primary);
      border-radius: 6px;

      .hint-text {
        color: var(--text-muted);
        font-size: 11px;
      }
    }

    .reference-grid {
      display: grid !important;
      grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)) !important;
      gap: 10px !important;

      .reference-item {
        margin-top: 0;
        position: relative;
        border-radius: 6px;
        overflow: hidden;
        cursor: pointer;
        border: 2px solid transparent;
        transition: all 0.2s ease;
        width: 100% !important;
        max-width: none !important;
        aspect-ratio: 4 / 3;
        background: var(--bg-card);
        display: flex;
        align-items: center;
        justify-content: center;

        &:hover {
          transform: translateY(-4px) scale(1.02);
          box-shadow: var(--shadow-lg);
          border-color: var(--accent);
        }

        &.selected {
          border-color: var(--accent);
          box-shadow: var(--shadow-glow);
        }

        :deep(.reference-image) {
          width: 100%;
          height: 100%;
          display: block;
          pointer-events: none;
        }

        :deep(.reference-image .el-image__inner) {
          width: 100%;
          height: 100%;
          object-fit: cover;
          display: block;
          transition: transform 0.3s;
        }

        &:hover :deep(.reference-image .el-image__inner) {
          transform: scale(1.05);
        }

        .reference-label {
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          padding: 4px 8px;
          background: linear-gradient(to top, rgba(0, 0, 0, 0.7), transparent);
          color: white;
          font-size: 10px;
          text-align: center;
        }
      }
    }
  }

  .generation-controls {
    margin-top: 40px;
    padding-top: 0;
    text-align: center;

    :deep(.el-button) {
      padding: 12px 32px;
      font-size: 14px;
      font-weight: 500;
      border-radius: 8px;
      transition: all 0.3s;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
      }
    }
  }

  @keyframes shimmer {
    0% {
      transform: translateX(-100%) translateY(-100%) rotate(45deg);
    }

    100% {
      transform: translateX(100%) translateY(100%) rotate(45deg);
    }
  }
}

// 视频合成列表样式
.merges-list {
  min-height: 300px;

  .merge-loading-state {
    min-height: 300px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: var(--text-muted);

    .merge-loading-text {
      font-size: 14px;
      font-weight: 500;
    }
  }

  .merge-items {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .merge-item {
    position: relative;
    background: var(--bg-card);
    border-radius: 12px;
    overflow: hidden;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--border-primary);
    box-shadow: var(--shadow-sm);

    &:hover {
      transform: translateY(-2px);
      box-shadow: var(--shadow-md);
      border-color: var(--accent-light);
    }

    // 状态指示条
    .status-indicator {
      position: absolute;
      left: 0;
      top: 0;
      bottom: 0;
      width: 4px;
      transition: all 0.3s ease;
    }

    &.merge-status-pending .status-indicator {
      background: linear-gradient(to bottom, #909399, #b1b3b8);
    }

    &.merge-status-processing .status-indicator {
      background: linear-gradient(to bottom, #e6a23c, #f0c78a);
      animation: pulse 2s ease-in-out infinite;
    }

    &.merge-status-completed .status-indicator {
      background: linear-gradient(to bottom, #67c23a, #95d475);
    }

    &.merge-status-failed .status-indicator {
      background: linear-gradient(to bottom, #f56c6c, #f89898);
    }

    .merge-content {
      padding: 20px 20px 20px 24px;
    }

    .merge-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 16px;
      gap: 12px;

      .title-section {
        display: flex;
        align-items: center;
        gap: 12px;
        flex: 1;
        min-width: 0;

        .title-icon {
          color: #409eff;
          flex-shrink: 0;

          &.rotating {
            animation: rotate 1.5s linear infinite;
          }
        }

        .merge-title {
          margin: 0;
          font-size: 16px;
          font-weight: 600;
          color: var(--text-secondary);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }

      :deep(.el-tag) {
        flex-shrink: 0;
        font-weight: 500;
        letter-spacing: 0.3px;
      }
    }

    .merge-details {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 16px;
      margin-bottom: 16px;
      padding: 16px;
      background: var(--bg-secondary);
      border-radius: 8px;
      border: 1px solid var(--border-primary);

      .detail-item {
        display: flex;
        align-items: flex-start;
        gap: 10px;

        .detail-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 32px;
          height: 32px;
          background: var(--bg-card);
          border-radius: 8px;
          color: var(--accent);
          flex-shrink: 0;
          box-shadow: var(--shadow-xs);
        }

        .detail-content {
          flex: 1;
          min-width: 0;

          .detail-label {
            font-size: 12px;
            color: var(--text-muted);
            margin-bottom: 4px;
            font-weight: 500;
          }

          .detail-value {
            font-size: 14px;
            color: var(--text-primary);
            font-weight: 500;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }
        }
      }
    }

    .merge-error {
      margin-bottom: 16px;

      :deep(.el-alert) {
        border-radius: 8px;
        border-left: 4px solid #f56c6c;
      }
    }

    .merge-actions {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
      padding-top: 16px;
      border-top: 1px solid var(--border-primary);

      &.single-action {
        grid-template-columns: minmax(0, 1fr);
      }

      :deep(.el-button) {
        width: 100%;
        margin: 0;
        font-weight: 500;
        transition: all 0.3s ease;

        &:hover {
          transform: translateY(-1px);
        }

        &.el-button--primary {
          box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);

          &:hover {
            box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
          }
        }

        &.preview-button {
          background: linear-gradient(135deg, #f2fbff 0%, #dff2ff 100%);
          border-color: #b8def6;
          color: #1d7eb8;
          box-shadow: 0 2px 8px rgba(29, 126, 184, 0.14);

          &:hover {
            background: linear-gradient(135deg, #ebf9ff 0%, #cfeeff 100%);
            border-color: #95cfee;
            color: #136b9f;
            box-shadow: 0 4px 12px rgba(29, 126, 184, 0.2);
          }
        }
      }
    }
  }

  @keyframes pulse {

    0%,
    100% {
      opacity: 1;
    }

    50% {
      opacity: 0.6;
    }
  }

  @keyframes rotate {
    from {
      transform: rotate(0deg);
    }

    to {
      transform: rotate(360deg);
    }
  }
}

.video-meta {
  margin-top: 16px;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  background: var(--bg-secondary);
}

.distribution-summary {
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: start;
  column-gap: 12px;
  row-gap: 8px;
  margin-bottom: 12px;
}

.distribution-summary-label {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 28px;
}

.distribution-summary-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 10px;
}

:deep(.distribution-tag.el-tag) {
  width: 100%;
  min-width: 0;
  justify-content: center;
  margin: 0;
}

.distribution-tag.is-link {
  cursor: pointer;
}

.distribution-form {
  :deep(.el-form-item) {
    margin-bottom: 16px;
  }

  .distribution-platform-list {
    flex: 1 0 100%;
    width: 100%;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
  }

  .distribution-platform-item {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    justify-content: flex-start;
    gap: 6px;
    min-height: 96px;
    padding: 12px 12px 8px;
    border-radius: 16px;
    border: 1px solid var(--border-primary);
    background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
    transition: all 0.2s ease;

    &.is-bound {
      border-color: rgba(38, 181, 123, 0.28);
      background: linear-gradient(135deg, rgba(38, 181, 123, 0.08) 0%, rgba(38, 181, 123, 0.02) 100%);
    }
  }

  .distribution-platform-main {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    cursor: pointer;
    width: 100%;
    margin-bottom: 2px;
  }

  .distribution-platform-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    color: #fff;
    flex-shrink: 0;
    box-shadow: 0 10px 20px rgba(15, 23, 42, 0.12);

    svg {
      width: 18px;
      height: 18px;
    }

    &.is-discord {
      background: #5865f2;
    }

    &.is-reddit {
      background: #ff4500;
      box-shadow: none;

      svg {
        width: 22px;
        height: 22px;
      }
    }

    &.is-pinterest {
      background: #e60023;
    }
  }

  .distribution-platform-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .distribution-platform-side {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    justify-content: flex-start;
    gap: 4px;
  }

  :deep(.distribution-binding-tag.el-tag) {
    width: 100%;
    justify-content: center;
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :deep(.binding-action-button.el-button) {
    width: 100%;
    margin: 0;
    border-radius: 999px;
    border: 1px solid rgba(32, 157, 235, 0.18);
    background: rgba(32, 157, 235, 0.12);
    color: #1387d6;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.2px;
    transition: all 0.2s ease;

    &:hover {
      border-color: rgba(19, 135, 214, 0.28);
      background: rgba(32, 157, 235, 0.18);
      color: #0f6cae;
    }
  }

  .distribution-target {
    margin-top: 0;
    font-size: 12px;
    color: var(--text-muted);
  }

  @media (max-width: 640px) {
    .distribution-platform-list {
      grid-template-columns: repeat(1, minmax(0, 1fr));
    }
  }
}

.distribution-dialog {
  margin: 0 auto !important;
  max-width: calc(100vw - 32px);
}

.distribution-dialog :deep(.el-dialog__header) {
  padding-bottom: 8px;
}

.distribution-dialog :deep(.el-dialog__body) {
  max-height: none;
  overflow: visible;
  padding: 16px 24px 12px !important;
}

.distribution-dialog :deep(.el-dialog__footer) {
  position: static;
  padding-top: 12px;
}
</style>
<style>
.distribution-dialog-overlay {
  overflow-y: hidden !important;
}

.video-prompt-box {
  margin-bottom: 10px;
  padding: 10px 12px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border: 1px solid var(--border-primary);
  font-size: 13px;
  line-height: 1.55;
  color: var(--text-secondary);
  word-break: break-word;
  min-height: 140px;
  max-height: 200px;
  overflow-y: auto;
}
</style>
