import request from '@/utils/request'
import type { AnalyticsSummary } from '@/types/analytics'

export const analyticsAPI = {
  summary() {
    return request.get<AnalyticsSummary>('/analytics/summary')
  }
}
