/**
 * 模块说明：营销视频时间线接口封装。
 * 业务场景：剪辑台需要保存/读取时间线草稿，并提交合成或渲染任务。
 * 核心职责：先复用现有 episode finalize 与 video-merge 能力，对外提供 project/timeline 语义。
 */
import request from '@/utils/request'
import type { MergeVideoRequest, VideoMerge } from '@/api/videoMerge'

export interface TimelineClip {
  id?: string | number
  type: 'video' | 'image' | 'audio' | 'subtitle' | 'text'
  source_url?: string
  start_time: number
  end_time: number
  duration?: number
  track?: number
  order?: number
  metadata?: Record<string, unknown>
}

export interface TimelineData {
  project_id?: string | number
  episode_id?: string | number
  clips: TimelineClip[]
  audio_clips?: TimelineClip[]
  duration?: number
  metadata?: Record<string, unknown>
}

export const timelineAPI = {
  getTimeline(projectId: string | number) {
    return request.get(`/projects/${projectId}/timeline`)
  },

  saveTimeline(projectId: string | number, data: TimelineData) {
    return request.put(`/projects/${projectId}/timeline`, data)
  },

  exportTimeline(projectId: string | number, data?: TimelineData) {
    return request.post(`/projects/${projectId}/timeline/export`, data || {})
  },

  renderTimeline(episodeId: string | number, data: TimelineData) {
    return request.post(`/episodes/${episodeId}/finalize`, data)
  },

  mergeTimelineVideo(data: MergeVideoRequest) {
    return request.post<{ merge: VideoMerge }>('/video-merges', data)
  }
}
