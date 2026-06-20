export interface WorkbenchOverview {
  totalProjects: number
  pendingProducts: number
  complianceCompleted: number
  imagesGenerated: number
  videosGenerated: number
  processingTasks: number
  coveredMarkets: number
}

export interface WorkbenchOverviewTrend {
  totalProjects: number
  pendingProducts: number
  complianceCompleted: number
  imagesGenerated: number
  videosGenerated: number
  processingTasks: number
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
