import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

type RouteLoader = () => Promise<unknown>
const views = {
  home: () => import('@/views/home/HomeLanding.vue'),
  product: () => import('@/views/home/ProductLanding.vue'),
  about: () => import('@/views/home/AboutLanding.vue'),
  dashboard: () => import('@/views/dashboard/Dashboard.vue'),
  agent: () => import('@/views/agent/AgentTransitionPage.vue'),
  agentResult: () => import('@/views/agent/AgentResultPage.vue'),
  agentHistory: () => import('@/views/silk-agent/AgentHistory.vue'),
  projects: () => import('@/views/projects/ProjectList.vue'),
  projectCreate: () => import('@/views/projects/ProjectCreate.vue'),
  project: () => import('@/views/projects/ProjectWorkbench.vue'),
  projectCompliance: () => import('@/views/projects/ProjectCompliance.vue'),
  projectScript: () => import('@/views/projects/ProjectScript.vue'),
  projectAssets: () => import('@/views/projects/ProjectAssets.vue'),
  projectTasks: () => import('@/views/projects/ProjectTasks.vue'),
  projectEditor: () => import('@/views/editor/MarketingVideoEditor.vue'),
  compliance: () => import('@/views/compliance/ComplianceAnalysis.vue'),
  image: () => import('@/views/media-generation/ImageGeneration.vue'),
  video: () => import('@/views/media-generation/VideoGeneration.vue'),
  music: () => import('@/views/media-generation/MusicLibrary.vue'),
  digitalHuman: () => import('@/views/digital-human/DigitalHumanDashboard.vue'),
  digitalHumanCreate: () => import('@/views/digital-human/DigitalHumanCreate.vue'),
  analytics: () => import('@/views/analytics/AnalyticsDashboard.vue'),
  settings: () => import('@/views/settings/SystemSettings.vue'),
  aiSettings: () => import('@/views/settings/AIConfig.vue')
} satisfies Record<string, RouteLoader>

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'Home', component: views.home },
  { path: '/products', name: 'Product', component: views.product },
  { path: '/about', name: 'About', component: views.about },
  { path: '/pricing', redirect: '/' },
  { path: '/dashboard', name: 'Dashboard', component: views.dashboard },
  { path: '/agent', name: 'AgentHome', component: views.agent },
  { path: '/agent/transition', redirect: '/agent' },
  { path: '/agent/result', name: 'AgentResult', component: views.agentResult },
  { path: '/agent/history', name: 'AgentHistory', component: views.agentHistory },
  { path: '/projects', name: 'ProjectList', component: views.projects },
  { path: '/projects/create', name: 'ProjectCreate', component: views.projectCreate },
  { path: '/projects/:id', name: 'ProjectWorkbench', component: views.project },
  { path: '/projects/:id/compliance', name: 'ProjectCompliance', component: views.projectCompliance },
  { path: '/projects/:id/script', name: 'ProjectScript', component: views.projectScript },
  { path: '/projects/:id/assets', name: 'ProjectAssets', component: views.projectAssets },
  { path: '/projects/:id/editor', name: 'MarketingVideoEditor', component: views.projectEditor },
  { path: '/projects/:id/tasks', name: 'ProjectTasks', component: views.projectTasks },
  { path: '/compliance', name: 'ComplianceCenter', component: views.compliance },
  { path: '/media/image', name: 'ImageGeneration', component: views.image },
  { path: '/media/video', name: 'VideoGeneration', component: views.video },
  { path: '/media/music', name: 'MusicLibrary', component: views.music },
  { path: '/digital-human', name: 'DigitalHumanCenter', component: views.digitalHuman },
  { path: '/digital-human/create', name: 'DigitalHumanCreate', component: views.digitalHumanCreate },
  { path: '/analytics', name: 'AnalyticsDashboard', component: views.analytics },
  { path: '/settings', name: 'SystemSettings', component: views.settings },
  { path: '/settings/ai', name: 'AIConfig', component: views.aiSettings },
  // Hidden compatibility paths. They never load legacy pages and are not used by menus.
  { path: '/dramas', redirect: '/projects' },
  { path: '/dramas/create', redirect: '/projects/create' },
  { path: '/dramas/:id', redirect: to => `/projects/${to.params.id}` },
  { path: '/workspace/script', redirect: to => to.query.projectId ? `/projects/${to.query.projectId}/script` : '/projects' },
  { path: '/workspace/content', redirect: to => ({ path: '/media/image', query: to.query }) },
  { path: '/workspace/timeline', redirect: to => to.query.projectId ? `/projects/${to.query.projectId}/editor` : '/projects' },
  { path: '/episodes/:id/edit', redirect: '/projects' },
  { path: '/episodes/:id/storyboard', redirect: '/projects' },
  { path: '/episodes/:id/generate', redirect: '/media/image' },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
]

export const preloadRouteByPath = (path?: string) => {
  if (!path) return
  const entry = Object.entries(views).find(([key]) => path.toLowerCase().includes(key.toLowerCase()))
  if (entry) void entry[1]().catch(() => undefined)
}
export const preloadWorkspaceRoutes = () => {
  ;[views.dashboard, views.agent, views.projects, views.compliance].forEach(loader => void loader().catch(() => undefined))
}

const router = createRouter({ history: createWebHistory(import.meta.env.BASE_URL), routes, scrollBehavior: () => ({ top: 0 }) })
export default router
