/**
 * 模块说明：音频处理接口封装。
 * 业务场景：营销视频剪辑中需要从视频素材提取口播、背景声或可复用音频轨道。
 * 核心职责：统一通过 request 调用后端音频提取接口，避免页面绕过拦截器。
 */
import request from '@/utils/request'

export interface ExtractAudioRequest {
  video_url: string
}

export interface ExtractAudioResponse {
  audio_url: string
  duration: number
}

export interface BatchExtractAudioRequest {
  video_urls: string[]
}

export interface BatchExtractAudioResponse {
  results: ExtractAudioResponse[]
  total: number
}

export const audioAPI = {
  extractAudio(videoUrl: string) {
    return request.post<ExtractAudioResponse>('/audio/extract', { video_url: videoUrl })
  },

  batchExtractAudio(videoUrls: string[]) {
    return request.post<BatchExtractAudioResponse>('/audio/extract/batch', { video_urls: videoUrls })
  }
}
