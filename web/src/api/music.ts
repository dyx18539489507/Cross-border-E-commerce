/**
 * 模块说明：数字丝路音乐与音效接口封装。
 * 业务场景：营销视频剪辑需要搜索配乐、试听代理流，并生成或检索短音效。
 * 核心职责：复用既有音乐代理与 SFX 能力，给跨境营销剪辑页提供稳定的前端调用入口。
 */
import request, { buildAPIURL } from '@/utils/request'

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

export interface MusicStreamParams {
  source?: string
  id?: string | number
  mid?: string
  hash?: string
  content_id?: string
  title?: string
  artist?: string
  url?: string
}

const buildQuery = (params: object) => {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    query.set(key, String(value))
  })
  return query.toString()
}

export const musicAPI = {
  getMusicList(params: { keywords?: string; page?: number; page_size?: number } = {}) {
    return request.get<MusicSearchResponse>('/music/search', { params })
  },

  searchMusic(params: { keywords: string; page?: number; page_size?: number }) {
    return request.get<MusicSearchResponse>('/music/search', { params })
  },

  getMusicDetail(params: { keywords: string; page?: number; page_size?: number }) {
    return request.get<MusicSearchResponse>('/music/search', { params })
  },

  importMusic(data: Record<string, unknown>) {
    return request.post('/assets', {
      ...data,
      type: 'audio',
      category: data.category || '营销配乐'
    })
  },

  deleteMusic(assetId: number | string) {
    return request.delete(`/assets/${assetId}`)
  },

  getNeteaseStreamUrl(id: string | number) {
    return `${buildAPIURL('/music/netease/stream')}?id=${encodeURIComponent(String(id))}`
  },

  getMusicStreamUrl(params: MusicStreamParams) {
    const query = buildQuery(params)
    return query ? `${buildAPIURL('/music/stream')}?${query}` : ''
  },

  isMusicStreamUrl(url?: string | null) {
    return !!url && (
      url.startsWith('/api/v1/music/stream') ||
      url.startsWith('/api/v1/music/netease/stream') ||
      url.startsWith(buildAPIURL('/music/stream')) ||
      url.startsWith(buildAPIURL('/music/netease/stream'))
    )
  }
}
