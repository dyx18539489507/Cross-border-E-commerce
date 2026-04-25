import request from '../utils/request'
import type {
  CreateDistributionRequest,
  DistributionJob,
  DistributionTarget,
  DistributionTargetsView,
  UploadPostProfile
} from '@/types/distribution'

export interface UploadPostConnectLinkResponse {
  profile: UploadPostProfile
  access_url: string
}

export interface UpsertDiscordTargetRequest {
  webhookUrl: string
  name?: string
  isDefault?: boolean
}

export interface UpsertRedditTargetRequest {
  subreddit: string
  flairId?: string
}

export const distributionAPI = {
  async listTargets(): Promise<DistributionTargetsView> {
    const response = await request.get<{ targets: DistributionTargetsView }>('/distributions/targets')
    return response.targets
  },

  async ensureUploadPostProfile(): Promise<UploadPostProfile> {
    const response = await request.post<{ profile: UploadPostProfile }>('/distributions/upload-post/profile/ensure')
    return response.profile
  },

  async syncUploadPostProfile(): Promise<UploadPostProfile> {
    const response = await request.post<{ profile: UploadPostProfile }>('/distributions/upload-post/sync')
    return response.profile
  },

  async generateUploadPostConnectLink(): Promise<UploadPostConnectLinkResponse> {
    return request.post<UploadPostConnectLinkResponse>('/distributions/upload-post/connect-link')
  },

  async listPinterestBoards(): Promise<DistributionTarget[]> {
    const response = await request.get<{ boards: DistributionTarget[] }>('/distributions/pinterest/boards')
    return response.boards || []
  },

  async setDefaultTarget(targetId: number): Promise<DistributionTarget> {
    const response = await request.put<{ target: DistributionTarget }>(`/distributions/targets/${targetId}/default`)
    return response.target
  },

  async saveRedditDefaultTarget(payload: UpsertRedditTargetRequest): Promise<DistributionTarget> {
    const response = await request.put<{ target: DistributionTarget }>('/distributions/targets/reddit/default', payload)
    return response.target
  },

  async upsertDiscordTarget(payload: UpsertDiscordTargetRequest): Promise<DistributionTarget> {
    const response = await request.post<{ target: DistributionTarget }>('/distributions/targets/discord', payload)
    return response.target
  },

  async deleteTarget(targetId: number): Promise<void> {
    await request.delete(`/distributions/targets/${targetId}`)
  },

  async listJobs(params?: { page?: number; page_size?: number }): Promise<{ jobs: DistributionJob[]; total: number }> {
    const response = await request.get<{ jobs: DistributionJob[]; total: number }>('/distributions', { params })
    return {
      jobs: response.jobs || [],
      total: response.total || 0
    }
  },

  async createDistribution(payload: CreateDistributionRequest): Promise<DistributionJob> {
    const response = await request.post<{ job: DistributionJob }>('/distributions', payload)
    return response.job
  },

  async getJob(jobId: number): Promise<DistributionJob> {
    const response = await request.get<{ job: DistributionJob }>(`/distributions/${jobId}`)
    return response.job
  },

  async retryJob(jobId: number): Promise<DistributionJob> {
    const response = await request.post<{ job: DistributionJob }>(`/distributions/${jobId}/retry`)
    return response.job
  }
}
