export interface WorkbenchOverview {
  pendingProducts: number
  complianceCompleted: number
  videosGenerated: number
  coveredMarkets: number
}

export interface WorkbenchOverviewTrend {
  pendingProducts: number
  complianceCompleted: number
  videosGenerated: number
  coveredMarkets: number
}

export interface WorkbenchMetricPoint {
  label: string
  value: number
}

export interface WorkbenchRecentTask {
  title: string
  market: string
  status: string
  meta: string
  tone: 'done' | 'progress' | 'pending' | string
  path?: string
}

export interface WorkbenchSummary {
  overview: WorkbenchOverview
  trends: WorkbenchOverviewTrend
  weeklyActivity: WorkbenchMetricPoint[]
  conversionTrend: WorkbenchMetricPoint[]
  recentTasks: WorkbenchRecentTask[]
}
