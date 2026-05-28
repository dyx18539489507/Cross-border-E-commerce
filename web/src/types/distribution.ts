/**
 * 模块说明：数字丝路分发业务类型定义。
 * 业务场景：内容创作完成后，需要把图文或视频分发到 Pinterest、Reddit、Discord 并追踪结果。
 * 核心职责：定义平台、目标、任务、执行结果和创建请求的前后端数据合同。
 */
export type DistributionPlatform = 'pinterest' | 'reddit' | 'discord'
export type DistributionContentType = 'text' | 'image' | 'video'
export type DistributionPublishMode = 'immediate' | 'schedule'
export type DistributionJobStatus =
  | 'pending'
  | 'scheduled'
  | 'processing'
  | 'completed'
  | 'partially_failed'
  | 'failed'
export type DistributionResultStatus = 'pending' | 'scheduled' | 'processing' | 'success' | 'failed'
export type DistributionTargetStatus = 'pending' | 'active' | 'needs_rebind' | 'disabled'

export interface UploadPostProfile {
  id: number
  username: string
  status: 'pending' | 'active' | 'error'
  connected_platforms?: string[]
  profile_snapshot?: Record<string, any>
  last_sync_at?: string
  created_at: string
  updated_at: string
}

// DistributionTarget 表示一个可发布目的地，例如 Pinterest board、Reddit subreddit 或 Discord webhook。
export interface DistributionTarget {
  id: number
  platform: DistributionPlatform
  target_type: string
  identifier: string
  name?: string
  status: DistributionTargetStatus
  is_default: boolean
  config?: Record<string, any>
  last_validated_at?: string
  last_sync_at?: string
  created_at: string
  updated_at: string
}

export interface DistributionTargetsView {
  uploadPostProfile?: UploadPostProfile
  targets: DistributionTarget[]
}

export interface DistributionResult {
  id: number
  job_id: number
  platform: DistributionPlatform
  target_id?: number
  content_type: DistributionContentType
  status: DistributionResultStatus
  target_snapshot?: Record<string, any>
  request_snapshot?: Record<string, any>
  response_snapshot?: Record<string, any>
  request_id?: string
  job_id_external?: string
  message_id?: string
  published_url?: string
  error_msg?: string
  attempt_count: number
  next_retry_at?: string
  started_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
  target?: DistributionTarget
}

export interface DistributionJob {
  id: number
  source_type: string
  source_ref?: string
  content_type: DistributionContentType
  title?: string
  body?: string
  media_url?: string
  selected_platforms?: DistributionPlatform[]
  platform_options?: Record<string, any>
  publish_mode: DistributionPublishMode
  scheduled_at?: string
  status: DistributionJobStatus
  request_snapshot?: Record<string, any>
  error_msg?: string
  created_at: string
  updated_at: string
  completed_at?: string
  results: DistributionResult[]
}

// CreateDistributionRequest 是前端提交分发任务的业务请求，平台特有参数放在 platformOptions 中隔离。
export interface CreateDistributionRequest {
  sourceType?: string
  sourceRef?: string
  contentType: DistributionContentType
  title: string
  body?: string
  mediaUrl?: string
  mediaRef?: string
  selectedPlatforms: DistributionPlatform[]
  platformOptions?: {
    reddit?: {
      subreddit?: string
      flairId?: string
      firstComment?: string
    }
    pinterest?: {
      boardId?: string
    }
    discord?: {
      targetId?: number
    }
  }
  publishMode?: DistributionPublishMode
  scheduledAt?: string
}
