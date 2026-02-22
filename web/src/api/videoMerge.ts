import request from '../utils/request'

export interface SceneClip {
  scene_id: string
  video_url: string
  start_time: number
  end_time: number
  duration: number
  order: number
}

export interface AudioClip {
  audio_url: string
  start_time: number
  end_time: number
  duration: number
  position: number
  volume?: number
  order?: number
  title?: string
}

export interface MergeVideoRequest {
  episode_id: string
  drama_id: string
  title: string
  scenes: SceneClip[]
  audio_clips?: AudioClip[]
  provider?: string
  model?: string
}

export interface VideoMerge {
  id: number
  episode_id: string
  drama_id: string
  title: string
  provider: string
  model?: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  scenes: SceneClip[]
  audio_clips?: AudioClip[]
  merged_url?: string
  duration?: number
  task_id?: string
  error_msg?: string
  created_at: string
  completed_at?: string
}

export type DistributionPlatform = 'tiktok' | 'youtube' | 'instagram' | 'x'
export type VideoDistributionStatus = 'pending' | 'processing' | 'published' | 'failed'

export interface DistributeVideoRequest {
  platforms: DistributionPlatform[]
  title?: string
  description?: string
  hashtags?: string[]
}

export interface VideoDistribution {
  id: number
  merge_id: number
  episode_id: number
  drama_id: number
  platform: DistributionPlatform
  title?: string
  description?: string
  hashtags?: string[]
  source_url: string
  status: VideoDistributionStatus
  message?: string
  published_url?: string
  error_msg?: string
  started_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export const videoMergeAPI = {
  async mergeVideos(data: MergeVideoRequest): Promise<VideoMerge> {
    const response = await request.post<{ merge: VideoMerge }>('/video-merges', data)
    return response.merge
  },

  async getMerge(mergeId: number): Promise<VideoMerge> {
    const response = await request.get<{ merge: VideoMerge }>(`/video-merges/${mergeId}`)
    return response.merge
  },

  async listMerges(params: {
    episode_id?: string
    status?: string
    page?: number
    page_size?: number
  }): Promise<{ merges: VideoMerge[]; total: number }> {
    const response = await request.get<{ merges: VideoMerge[]; total: number }>('/video-merges', { params })
    return {
      merges: response.merges || [],
      total: response.total || 0
    }
  },

  async deleteMerge(mergeId: number): Promise<void> {
    await request.delete(`/video-merges/${mergeId}`)
  },

  async distributeVideo(mergeId: number, data: DistributeVideoRequest): Promise<VideoDistribution[]> {
    const response = await request.post<{ distributions: VideoDistribution[] }>(`/video-merges/${mergeId}/distribute`, data)
    return response.distributions || []
  },

  async listDistributions(mergeId: number): Promise<VideoDistribution[]> {
    const response = await request.get<{ distributions: VideoDistribution[] }>(`/video-merges/${mergeId}/distributions`)
    return response.distributions || []
  }
}
