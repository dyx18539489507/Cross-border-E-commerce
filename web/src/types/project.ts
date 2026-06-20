export interface MarketingContentVersion {
  id: number
  project_id?: number
  version_number: number
  title: string
  script_content?: string
  description?: string
  duration?: number
  status?: string
  video_url?: string
  shots?: MarketingShot[]
  created_at?: string
  updated_at?: string
}

export interface MarketingShot {
  id: number
  shot_number: number
  title?: string
  visual?: string
  selling_point?: string
  voiceover?: string
  subtitle?: string
  digital_human_action?: string
  source?: string
  market_adaptation?: string
  duration?: number
}

export interface DigitalPresenter {
  id: number
  name: string
  role?: string
  description?: string
  appearance?: string
  personality?: string
  voice_style?: string
  image_url?: string
  created_at?: string
  updated_at?: string
}

export interface MarketingProject {
  id: number
  project_name: string
  product_name: string
  product_description?: string
  target_markets: string[]
  target_language?: string
  product_selling_points?: string
  material_composition?: string
  marketing_style?: string
  platform_channels: string[]
  product_image?: string
  compliance_focus?: string
  compliance_score?: number
  compliance_status?: string
  status: string
  thumbnail?: string
  content_versions?: MarketingContentVersion[]
  digital_presenters?: DigitalPresenter[]
  marketing_scenes?: unknown[]
  created_at: string
  updated_at: string
}

export interface ProjectListResponse {
  items: MarketingProject[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface CreateProjectInput {
  product_name: string
  project_name?: string
  product_description: string
  target_markets: string[]
  target_language?: string
  product_selling_points?: string
  material_composition?: string
  marketing_style?: string
  platform_channels?: string[]
  product_image?: string
  compliance_focus?: string
  compliance_token?: string
}

export interface ComplianceResult {
  score: number
  level: string
  level_label?: string
  summary: string
  non_compliance_points: string[]
  rectification_suggestions: string[]
  suggested_categories?: string[]
}

export interface ComplianceCheckResponse {
  compliance: ComplianceResult
  compliance_token: string
}

export interface ProjectScriptResponse {
  project_id: number
  project_name: string
  content_versions: MarketingContentVersion[]
}

export interface ProjectAssetsResponse {
  project_id: number
  items: Array<Record<string, any>>
}

export interface ProjectTasksResponse {
  project_id: number
  images: Array<Record<string, any>>
  videos: Array<Record<string, any>>
  tasks: Array<Record<string, any>>
}
