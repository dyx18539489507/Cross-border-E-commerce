import request from '@/utils/request'
import type {
  ComplianceCheckResponse,
  CreateProjectInput,
  MarketingContentVersion,
  MarketingProject,
  ProjectAssetsResponse,
  ProjectListResponse,
  ProjectScriptResponse,
  ProjectTasksResponse
} from '@/types/project'

type LegacyProject = Record<string, any>

const parseList = (value: unknown): string[] => {
  if (Array.isArray(value)) return value.map(String).filter(Boolean)
  if (typeof value !== 'string') return []
  const trimmed = value.trim()
  if (!trimmed) return []
  try {
    const parsed = JSON.parse(trimmed)
    if (Array.isArray(parsed)) return parsed.map(String).filter(Boolean)
  } catch {
    // Existing records may store comma-separated values.
  }
  return trimmed.split(',').map((item) => item.trim()).filter(Boolean)
}

const readMetadata = (value: unknown): Record<string, any> => {
  if (value && typeof value === 'object' && !Array.isArray(value)) return value as Record<string, any>
  if (typeof value === 'string' && value.trim()) {
    try { return JSON.parse(value) } catch { return {} }
  }
  return {}
}

const mapShot = (item: LegacyProject, index: number) => ({
  id: Number(item.id ?? index + 1),
  shot_number: Number(item.shot_number ?? item.storyboard_number ?? index + 1),
  title: item.title,
  visual: item.visual ?? item.description ?? '',
  selling_point: item.selling_point ?? item.result ?? '',
  voiceover: item.voiceover ?? item.dialogue ?? '',
  subtitle: item.subtitle ?? item.title ?? '',
  digital_human_action: item.digital_human_action ?? item.action ?? '',
  source: item.source ?? item.image_prompt ?? '',
  market_adaptation: item.market_adaptation ?? item.atmosphere ?? '',
  duration: Number(item.duration ?? 5)
})

const mapContentVersion = (item: LegacyProject): MarketingContentVersion => ({
  id: Number(item.id),
  project_id: Number(item.project_id ?? item.drama_id) || undefined,
  version_number: Number(item.version_number ?? item.episode_number ?? 1),
  title: String(item.title || `营销内容版本 ${item.version_number ?? item.episode_number ?? 1}`),
  script_content: item.script_content,
  description: item.description,
  duration: item.duration,
  status: item.status,
  video_url: item.video_url,
  shots: (item.shots ?? item.storyboards ?? []).map(mapShot),
  created_at: item.created_at,
  updated_at: item.updated_at
})

const mapProject = (item: LegacyProject): MarketingProject => {
  const metadata = readMetadata(item.metadata)
  return {
    id: Number(item.id),
    project_name: String(item.project_name ?? item.title ?? ''),
    product_name: String(item.product_name ?? item.title ?? ''),
    product_description: item.product_description ?? item.description ?? '',
    target_markets: parseList(item.target_markets ?? item.target_country),
    target_language: item.target_language ?? metadata.target_language ?? '',
    product_selling_points: item.product_selling_points ?? item.marketing_selling_points ?? '',
    material_composition: item.material_composition ?? '',
    marketing_style: item.marketing_style ?? item.style ?? '',
    platform_channels: parseList(item.platform_channels ?? metadata.platform_channels),
    product_image: item.product_image ?? item.thumbnail ?? '',
    compliance_focus: item.compliance_focus ?? metadata.compliance_focus ?? '',
    compliance_score: Number(item.compliance_score ?? 0),
    compliance_status: item.compliance_status ?? item.compliance_level ?? 'pending',
    status: String(item.status ?? 'draft'),
    thumbnail: item.thumbnail,
    content_versions: (item.content_versions ?? item.episodes ?? []).map(mapContentVersion),
    digital_presenters: item.digital_presenters ?? item.characters ?? [],
    marketing_scenes: item.marketing_scenes ?? item.scenes ?? [],
    created_at: String(item.created_at ?? ''),
    updated_at: String(item.updated_at ?? '')
  }
}

const toBackendPayload = (input: CreateProjectInput) => ({
  title: input.project_name || input.product_name,
  description: input.product_description,
  target_country: input.target_markets,
  material_composition: input.material_composition || '',
  marketing_selling_points: input.product_selling_points || '',
  compliance_token: input.compliance_token || '',
  genre: input.marketing_style || 'product-marketing',
  tags: JSON.stringify({
    target_language: input.target_language,
    platform_channels: input.platform_channels,
    product_image: input.product_image,
    compliance_focus: input.compliance_focus
  })
})

export const projectAPI = {
  async list(params?: { page?: number; page_size?: number; status?: string; keyword?: string }) {
    const data = await request.get<{ items: LegacyProject[]; pagination: ProjectListResponse['pagination'] }>('/projects', { params })
    return { ...data, items: (data.items || []).map(mapProject) } as ProjectListResponse
  },
  async get(id: string | number) {
    return mapProject(await request.get<LegacyProject>(`/projects/${id}`))
  },
  async create(input: CreateProjectInput) {
    const data = await request.post<any>('/projects', toBackendPayload(input))
    return { ...data, project: mapProject(data.project ?? data.drama ?? data) }
  },
  async update(id: string | number, input: Partial<CreateProjectInput> & { status?: string }) {
    const data = await request.put<LegacyProject>(`/projects/${id}`, toBackendPayload({
      product_name: input.product_name || input.project_name || '营销项目',
      product_description: input.product_description || '',
      target_markets: input.target_markets || [],
      ...input
    } as CreateProjectInput))
    return mapProject(data)
  },
  remove(id: string | number) {
    return request.delete(`/projects/${id}`)
  },
  checkCompliance(input: CreateProjectInput) {
    return request.post<ComplianceCheckResponse>('/projects/compliance-check', toBackendPayload(input))
  },
  async getScript(id: string | number) {
    const data = await request.get<any>(`/projects/${id}/script`)
    return { ...data, content_versions: (data.content_versions ?? data.episodes ?? []).map(mapContentVersion) } as ProjectScriptResponse
  },
  saveScript(id: string | number, content_versions: Array<Partial<MarketingContentVersion>>) {
    return request.put<ProjectScriptResponse>(`/projects/${id}/script`, { content_versions })
  },
  getAssets(id: string | number) {
    return request.get<ProjectAssetsResponse>(`/projects/${id}/assets`)
  },
  getTasks(id: string | number) {
    return request.get<ProjectTasksResponse>(`/projects/${id}/tasks`)
  },
  getTimeline(id: string | number) {
    return request.get<{ project_id: number; timeline: Record<string, any> | null }>(`/projects/${id}/timeline`)
  },
  saveTimeline(id: string | number, timeline: Record<string, any>) {
    return request.put(`/projects/${id}/timeline`, timeline)
  }
}
