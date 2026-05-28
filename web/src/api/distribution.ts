/**
 * 模块说明：数字丝路多平台分发接口封装。
 * 业务场景：生成的图文/视频素材需要发布到 Pinterest、Reddit、Discord 等外部渠道。
 * 核心职责：管理分发账号、目标、任务创建、任务查询和失败重试，隐藏后端返回包装结构。
 */
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
  /**
   * 功能：获取当前设备身份下的分发账号和目标。
   * 参数：无。
   * 返回：Upload-Post profile 与各平台目标列表，用于判断能否提交分发。
   */
  async listTargets(): Promise<DistributionTargetsView> {
    const response = await request.get<{ targets: DistributionTargetsView }>('/distributions/targets')
    return response.targets
  },

  /**
   * 功能：确保当前设备有 Upload-Post profile。
   * 参数：无。
   * 返回：UploadPostProfile，后续 Pinterest/Reddit 授权链接都依赖该 profile。
   */
  async ensureUploadPostProfile(): Promise<UploadPostProfile> {
    const response = await request.post<{ profile: UploadPostProfile }>('/distributions/upload-post/profile/ensure')
    return response.profile
  },

  /**
   * 功能：同步 Upload-Post 授权状态。
   * 参数：无。
   * 返回：最新 profile，包含已连接平台和最近同步时间。
   */
  async syncUploadPostProfile(): Promise<UploadPostProfile> {
    const response = await request.post<{ profile: UploadPostProfile }>('/distributions/upload-post/sync')
    return response.profile
  },

  /**
   * 功能：生成外部平台授权连接页。
   * 参数：无。
   * 返回：profile 与 access_url，前端打开 access_url 让用户完成 Pinterest/Reddit 授权。
   */
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

  /**
   * 功能：创建一次多平台分发任务。
   * 参数：payload 包含内容类型、来源、媒体地址、目标平台、平台参数和发布模式。
   * 返回：DistributionJob，包含每个平台的执行结果初始状态。
   */
  async createDistribution(payload: CreateDistributionRequest): Promise<DistributionJob> {
    const response = await request.post<{ job: DistributionJob }>('/distributions', payload)
    return response.job
  },

  /**
   * 功能：查询单个分发任务最新状态。
   * 参数：jobId 为后端分发任务 ID。
   * 返回：DistributionJob，前端轮询它来展示各平台成功、失败或待处理状态。
   */
  async getJob(jobId: number): Promise<DistributionJob> {
    const response = await request.get<{ job: DistributionJob }>(`/distributions/${jobId}`)
    return response.job
  },

  async retryJob(jobId: number): Promise<DistributionJob> {
    const response = await request.post<{ job: DistributionJob }>(`/distributions/${jobId}/retry`)
    return response.job
  }
}
