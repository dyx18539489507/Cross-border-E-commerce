/**
 * 模块说明：数字丝路音乐与音效接口封装。
 * 业务场景：营销视频剪辑需要搜索配乐、试听代理流，并生成或检索短音效。
 * 核心职责：复用既有音乐代理与 SFX 能力，给跨境营销剪辑页提供稳定的前端调用入口。
 */
import request from '@/utils/request'

export interface MusicSearchItem {
  id?: string | number
  mid?: string
  hash?: string
  content_id?: string
  title?: string
  name?: string
  artist?: string
  singer?: string
  source?: string
  duration?: number | string
  song_url?: string
  url?: string
}

export interface MusicSearchResponse {
  items?: MusicSearchItem[]
  songs?: MusicSearchItem[]
  total?: number
}

export interface SfxItem {
  id: string
  name: string
  url: string
  category: string
  duration?: number
  rank?: number
  view_count?: number
}

export const musicAPI = {
  searchMusic(params: { keywords: string; page?: number; page_size?: number }) {
    return request.get<MusicSearchResponse>('/music/search', { params })
  },

  listSfx(params?: { keywords?: string; category?: string; page?: number; limit?: number }) {
    return request.get<{ items: SfxItem[]; total?: number; has_more?: boolean }>('/sfx', { params })
  },

  generateSfx(data: { prompt: string; count?: number }) {
    return request.post<{ items: SfxItem[] }>('/sfx/generate', data)
  }
}
