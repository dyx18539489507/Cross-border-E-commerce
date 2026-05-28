export type AnalyticsMetricIcon = 'eye' | 'spark' | 'cart' | 'coin'
export type AnalyticsInsightIcon = 'growth' | 'audience' | 'video'
export type AnalyticsTone = 'blue' | 'purple' | 'orange' | 'green'

export interface AnalyticsMetricCard {
  icon: AnalyticsMetricIcon
  label: string
  value: string
  trend: string
  tone: AnalyticsTone
}

export interface AnalyticsTrendSeries {
  name: string
  color: string
  width: number
  opacity: number
  data: number[]
}

export interface AnalyticsMarketShare {
  name: string
  value: string
  color: string
}

export interface AnalyticsVideoBar {
  name: string
  views: number
  conversions: number
}

export interface AnalyticsInsight {
  icon: AnalyticsInsightIcon
  tone: AnalyticsTone
  title: string
  description: string
}

export interface AnalyticsRecommendation {
  title: string
  description: string
}

export interface AnalyticsSummary {
  metricCards: AnalyticsMetricCard[]
  trendXAxis: string[]
  trendSeries: AnalyticsTrendSeries[]
  marketShare: AnalyticsMarketShare[]
  videoBars: AnalyticsVideoBar[]
  insights: AnalyticsInsight[]
  recommendations: AnalyticsRecommendation[]
}
