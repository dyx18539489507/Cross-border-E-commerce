/**
 * 模块说明：媒体代理与二进制素材请求封装。
 * 业务场景：剪辑台需要通过后端代理试听外部音乐、下载成片、探测音频可播放性。
 * 核心职责：集中处理 /media/proxy 和音乐流等二进制请求，组件只关心业务结果。
 */
import request, { buildAPIURL } from '@/utils/request'

export interface MediaRequestOptions {
  timeoutMs?: number
  signal?: AbortSignal
}

const API_PREFIX = '/api/v1'

const stripApiPrefix = (url: string) => {
  if (url.startsWith(API_PREFIX)) {
    return url.slice(API_PREFIX.length) || '/'
  }
  return url
}

const isAbsoluteHttpUrl = (url: string) => /^https?:\/\//i.test(url)

const toRequestUrl = (url: string) => {
  if (isAbsoluteHttpUrl(url)) {
    return '/media/proxy'
  }
  return stripApiPrefix(url)
}

const toRequestParams = (url: string) => {
  if (isAbsoluteHttpUrl(url)) {
    return { url }
  }
  return undefined
}

export const mediaAPI = {
  getMediaProxyUrl(url: string) {
    return `${buildAPIURL('/media/proxy')}?url=${encodeURIComponent(url)}`
  },

  isMediaProxyUrl(url?: string | null) {
    if (!url) return false
    if (url.startsWith(`${API_PREFIX}/media/proxy`) || url.startsWith(buildAPIURL('/media/proxy'))) return true
    try {
      const parsed = new URL(url, window.location.origin)
      const proxy = new URL(buildAPIURL('/media/proxy'), window.location.origin)
      return parsed.origin === proxy.origin && parsed.pathname === proxy.pathname
    } catch {
      return false
    }
  },

  getMediaList(params?: Record<string, unknown>) {
    return request.get('/assets', { params })
  },

  getMediaDetail(id: number | string) {
    return request.get(`/assets/${id}`)
  },

  deleteMedia(id: number | string) {
    return request.delete(`/assets/${id}`)
  },

  async fetchText(url: string, options: MediaRequestOptions = {}) {
    return request.get<string>(toRequestUrl(url), {
      params: toRequestParams(url),
      responseType: 'text',
      timeout: options.timeoutMs,
      signal: options.signal
    })
  },

  async fetchJSON<T = any>(url: string, options: MediaRequestOptions = {}) {
    return request.get<T>(toRequestUrl(url), {
      params: toRequestParams(url),
      timeout: options.timeoutMs,
      signal: options.signal
    })
  },

  async fetchBlob(url: string, options: MediaRequestOptions = {}) {
    return request.get<Blob>(toRequestUrl(url), {
      params: toRequestParams(url),
      responseType: 'blob',
      timeout: options.timeoutMs,
      signal: options.signal
    })
  },

  async probeAudioUrl(url: string, timeoutMs: number) {
    const controller = new AbortController()
    const timer = window.setTimeout(() => controller.abort(), timeoutMs)
    try {
      const response = await request.get<ArrayBuffer>(toRequestUrl(url), {
        params: toRequestParams(url),
        responseType: 'arraybuffer',
        headers: { Range: 'bytes=0-2048' },
        timeout: timeoutMs,
        signal: controller.signal
      })
      return response
    } finally {
      window.clearTimeout(timer)
    }
  }
}
