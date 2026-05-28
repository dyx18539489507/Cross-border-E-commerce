/**
 * 模块说明：丝路 Agent 前端请求封装。
 * 业务场景：首页、过渡页和结果页需要把商品文本、图片、目标市场等信息交给后端 Agent。
 * 核心职责：统一调用 Agent 识别与方案生成接口，并用类型约束返回给页面展示的业务结构。
 */
import type { AgentInput, AgentResult } from '@/types/agent'
import request from '@/utils/request'

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
  createdAt: string
  updatedAt: string
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
  /**
   * 功能：生成完整的丝路 Agent 出海营销方案。
   * 参数：data 为商品、图片、目标市场、平台、人群、卖点等上下文。
   * 返回：AgentResult，包含识别信息、合规、本地化、脚本、数字人和投放建议。
   */
  generate(data: AgentInput) {
    return request.post<AgentResult>('/agent/generate', data)
  },
  listHistory(limit = 20) {
    return request.get<{ items: AgentHistoryItem[] }>('/agent/history', { params: { limit } })
  },
  getHistory(id: number) {
    return request.get<{ item: AgentHistoryItem }>(`/agent/history/${id}`)
  }
}
