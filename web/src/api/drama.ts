/**
 * 模块说明：项目与合规接口请求封装。
 * 业务场景：数字丝路商品录入完成后，前端复用现有项目接口完成合规预检、创建项目和后续脚本/分镜流程。
 * 核心职责：封装 /dramas 相关请求；本次注释只解释数字丝路用到的合规与创建入口，旧短剧接口不展开说明。
 */
import type {
    CheckDramaComplianceResponse,
    CreateDramaRequest,
    CreateDramaResponse,
    Drama,
    DramaListQuery,
    DramaStats,
    UpdateDramaRequest
} from '../types/drama'
import request from '../utils/request'

export const dramaAPI = {
  list(params?: DramaListQuery) {
    return request.get<{
      items: Drama[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/dramas', { params })
  },

  /**
   * 功能：创建数字丝路商品项目并进入后续脚本/分镜流程。
   * 参数：data 为商品录入适配出的 CreateDramaRequest，可携带 compliance_token 证明已完成同内容预检。
   * 返回：CreateDramaResponse，包含创建后的项目记录和后端保存的合规结果。
   */
  create(data: CreateDramaRequest) {
    return request.post<CreateDramaResponse>('/dramas', data)
  },

  /**
   * 功能：对商品录入数据进行合规预检。
   * 参数：data 包含商品标题、描述、目标国家、材质和营销卖点。
   * 返回：合规评分、风险等级、整改建议和短期 compliance_token，供后续创建接口校验。
   */
  checkCompliance(data: CreateDramaRequest) {
    return request.post<CheckDramaComplianceResponse>('/dramas/compliance-check', data)
  },

  get(id: string) {
    return request.get<Drama>(`/dramas/${id}`)
  },

  update(id: string, data: UpdateDramaRequest) {
    return request.put<Drama>(`/dramas/${id}`, data)
  },

  delete(id: string) {
    return request.delete(`/dramas/${id}`)
  },

  getStats() {
    return request.get<DramaStats>('/dramas/stats')
  },

  saveOutline(id: string, data: { title: string; summary: string; genre?: string; tags?: string[] }) {
    return request.put(`/dramas/${id}/outline`, data)
  },

  getCharacters(dramaId: string) {
    return request.get(`/dramas/${dramaId}/characters`)
  },

  saveCharacters(id: string, data: any[], episodeId?: string) {
    return request.put(`/dramas/${id}/characters`, { 
      characters: data,
      episode_id: episodeId ? parseInt(episodeId) : undefined
    })
  },

  saveEpisodes(id: string, data: any[]) {
    return request.put(`/dramas/${id}/episodes`, { episodes: data })
  },

  saveProgress(id: string, data: { current_step: string; step_data?: any }) {
    return request.put(`/dramas/${id}/progress`, data)
  },

  generateStoryboard(episodeId: string, model?: string) {
    return request.post(`/episodes/${episodeId}/storyboards`, { model })
  },

  getBackgrounds(episodeId: string) {
    return request.get(`/images/episode/${episodeId}/backgrounds`)
  },

  extractBackgrounds(episodeId: string, model?: string) {
    return request.post<{ task_id: string; status: string; message: string }>(`/images/episode/${episodeId}/backgrounds/extract`, { model })
  },

  batchGenerateBackgrounds(episodeId: string) {
    return request.post(`/images/episode/${episodeId}/batch`)
  },

  generateSingleBackground(backgroundId: number, dramaId: string, prompt: string) {
    return request.post('/images', {
      background_id: backgroundId,
      drama_id: dramaId,
      prompt: prompt
    })
  },

  getStoryboards(episodeId: string) {
    return request.get(`/episodes/${episodeId}/storyboards`)
  },

  updateStoryboard(storyboardId: string, data: any) {
    return request.put(`/storyboards/${storyboardId}`, data)
  },

  deleteStoryboard(storyboardId: string) {
    return request.delete(`/storyboards/${storyboardId}`)
  },

  updateScene(sceneId: string, data: { 
    background_id?: string; 
    characters?: string[];
    location?: string;
    time?: string;
    action?: string;
    dialogue?: string;
    description?: string;
    duration?: number;
    scene_id?: number;
  }) {
    return request.put(`/scenes/${sceneId}`, data)
  },

  generateSceneImage(data: { scene_id: number; prompt?: string }) {
    return request.post<{ image_generation: { id: number } }>('/scenes/generate-image', data)
  },

  updateScenePrompt(sceneId: string, prompt: string) {
    return request.put(`/scenes/${sceneId}/prompt`, { prompt })
  },

  deleteScene(sceneId: string) {
    return request.delete(`/scenes/${sceneId}`)
  },

  // 完成集数制作（触发视频合成）
  finalizeEpisode(episodeId: string, timelineData?: any) {
    return request.post(`/episodes/${episodeId}/finalize`, timelineData || {})
  }
}
