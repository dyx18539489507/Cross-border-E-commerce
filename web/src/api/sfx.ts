/**
 * 模块说明：数字丝路音效接口封装。
 * 业务场景：营销短视频剪辑台需要检索成交、转场、强调等音效，也可以按提示词生成轻量音效素材。
 * 核心职责：统一调用 /sfx 相关接口，避免编辑器组件直接拼接后端 URL。
 */
import request from '@/utils/request'

export interface SoundEffectItem {
  id: string
  name: string
  url: string
  category: string
  duration?: number
  rank?: number
  view_count?: number
  source?: string
  title?: string
  audio_url?: string
  file_url?: string
  file_path?: string
  preview_url?: string
  artist?: string
  user?: string
  cover?: string
  image?: string
  description?: string
}

export interface SoundEffectListParams {
  keywords?: string
  category?: string
  page?: number
  limit?: number
}

export interface SoundEffectListResponse {
  items: SoundEffectItem[]
  page?: number
  limit?: number
  total?: number
  has_more?: boolean
  warnings?: string[]
}

export interface GenerateSoundEffectRequest {
  prompt: string
  count?: number
}

export const sfxAPI = {
  getSoundEffects(params?: SoundEffectListParams) {
    return request.get<SoundEffectListResponse>('/sfx', { params })
  },

  searchSoundEffects(params: SoundEffectListParams) {
    return request.get<SoundEffectListResponse>('/sfx', { params })
  },

  getSoundEffectDetail(id: string) {
    return request.get<SoundEffectItem>('/sfx', { params: { keywords: id, limit: 1 } })
  },

  generateSoundEffect(data: GenerateSoundEffectRequest) {
    return request.post<{ items: SoundEffectItem[] }>('/sfx/generate', data)
  }
}
