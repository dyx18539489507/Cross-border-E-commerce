/**
 * 模块说明：丝路 Agent 前端请求封装。
 * 业务场景：首页、过渡页和结果页需要把商品文本、图片、目标市场等信息交给后端 Agent。
 * 核心职责：统一调用 Agent 识别与方案生成接口，并用类型约束返回给页面展示的业务结构。
 */
import type {
  AgentAnalyzeInput,
  AgentFollowUpInput,
  AgentInput,
  AgentResult,
  CreateProjectFromAgentResponse,
  WorkflowResult
} from '@/types/agent'
import request, { buildAPIURL, getClientDeviceID } from '@/utils/request'

export interface AgentHistoryItem {
  id: number
  requestId: string
  productName: string
  category: string
  targetMarket: string
  targetPlatform: string
  targetAudience: string
  rawPrompt: string
  status: string
  model: string
  input: AgentInput
  result?: AgentResult
  workflow?: WorkflowResult
  createdAt: string
  updatedAt: string
}

export type ProductInfoExtractRequest = AgentInput
export type ProductInfoExtractResponse = AgentInput
export type MarketingSolutionRequest = AgentInput
export type MarketingSolutionResponse = AgentResult
export type AgentWorkflowRequest = AgentInput & { workflow?: boolean }
export type AgentWorkflowResponse = WorkflowResult
export type AgentAnalysisRequest = AgentAnalyzeInput
export type AgentFollowUpRequest = AgentFollowUpInput
export type AgentHistoryListResponse = { items: AgentHistoryItem[] }
export type AgentHistoryDetailResponse = { item: AgentHistoryItem }
export type CreateProjectFromAgentRequest = { result?: AgentResult; workflow?: WorkflowResult }

const postAgentStream = (path: string, data: unknown, signal?: AbortSignal) => {
  return fetch(buildAPIURL(path), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Device-ID': getClientDeviceID()
    },
    body: JSON.stringify(data),
    signal
  })
}

export const agentAPI = {
  /**
   * 功能：把用户输入提交给后端做轻量结构化识别。
   * 参数：data 包含原始描述、商品字段、目标市场和可选图片 Data URL。
   * 返回：归一后的 AgentInput，用于前端在正式生成前补齐可识别字段。
   */
  extract(data: AgentInput) {
    return request.post<AgentInput>('/agent/extract', data)
  },
  extractProductInfo(data: ProductInfoExtractRequest) {
    return request.post<ProductInfoExtractResponse>('/agent/extract', data)
  },
  /**
   * 功能：生成完整的丝路 Agent 出海营销方案。
   * 参数：data 为商品、图片、目标市场、平台、人群、卖点等上下文。
   * 返回：AgentResult，包含识别信息、合规、本地化、脚本、数字人和投放建议。
   */
  generate(data: AgentInput) {
    return request.post<AgentResult>('/agent/generate', data)
  },
  generateMarketingSolution(data: MarketingSolutionRequest) {
    return request.post<MarketingSolutionResponse>('/agent/generate', data)
  },
  generateWorkflow(data: AgentWorkflowRequest) {
    return request.post<WorkflowResult>('/agent/workflow', { ...data, workflow: true })
  },
  workflow(data: AgentWorkflowRequest) {
    return request.post<WorkflowResult>('/agent/workflow', { ...data, workflow: true })
  },
  runAgentWorkflow(data: AgentWorkflowRequest) {
    return request.post<AgentWorkflowResponse>('/agent/workflow', { ...data, workflow: true })
  },
  analyze(data: AgentAnalyzeInput) {
    return request.post('/agent/analyze', data)
  },
  analyzeProduct(data: AgentAnalysisRequest) {
    return request.post('/agent/analyze', data)
  },
  analyzeProductStream(data: AgentAnalysisRequest, signal?: AbortSignal) {
    return postAgentStream('/agent/analyze', data, signal)
  },
  followUp(data: AgentFollowUpInput) {
    return request.post('/agent/follow-up', data)
  },
  sendFollowUp(data: AgentFollowUpRequest) {
    return request.post('/agent/follow-up', data)
  },
  sendFollowUpStream(data: AgentFollowUpRequest, signal?: AbortSignal) {
    return postAgentStream('/agent/follow-up', data, signal)
  },
  listHistory(params: { limit?: number; page?: number } | number = { limit: 20 }) {
    const normalizedParams = typeof params === 'number' ? { limit: params } : params
    return request.get<AgentHistoryListResponse>('/agent/history', { params: normalizedParams })
  },
  getAgentHistory(params: { limit?: number; page?: number } | number = { limit: 20 }) {
    const normalizedParams = typeof params === 'number' ? { limit: params } : params
    return request.get<AgentHistoryListResponse>('/agent/history', { params: normalizedParams })
  },
  getHistory(id: number | string) {
    return request.get<AgentHistoryDetailResponse>(`/agent/history/${id}`)
  },
  getAgentHistoryDetail(id: number | string) {
    return request.get<AgentHistoryDetailResponse>(`/agent/history/${id}`)
  },
  createProjectFromAgent(id?: number | string, data?: CreateProjectFromAgentRequest) {
    if (id !== undefined && id !== null && String(id).trim() !== '') {
      return request.post<CreateProjectFromAgentResponse>(`/agent/${id}/create-project`, data || {})
    }
    return request.post<CreateProjectFromAgentResponse>('/agent/create-project', data || {})
  }
}
