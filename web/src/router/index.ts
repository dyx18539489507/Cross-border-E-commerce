/**
 * 模块说明：数字丝路前端路由与工作区页面预加载。
 * 业务场景：用户从首页进入丝路 Agent、商品录入、合规分析、内容生产和数据分析等跨境电商工作流。
 * 核心职责：维护数字丝路页面路径、按访问意图提前加载关键页面，并保留旧工作台路径的兼容跳转。
 */
import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

type RouteLoader = () => Promise<unknown>

const viewLoaders = {
  HomeLanding: () => import('../views/home/HomeLanding.vue'),
  ProductLanding: () => import('../views/home/ProductLanding.vue'),
  AboutLanding: () => import('../views/home/AboutLanding.vue'),
  AgentTransitionPage: () => import('../views/agent/AgentTransitionPage.vue'),
  AgentResultPage: () => import('../views/agent/AgentResultPage.vue'),
  WorkbenchDashboard: () => import('../views/drama/WorkbenchDashboard.vue'),
  DramaCreate: () => import('../views/drama/DramaCreate.vue'),
  ProductEntryMarket: () => import('../views/drama/ProductEntryMarket.vue'),
  ProductEntryDetails: () => import('../views/drama/ProductEntryDetails.vue'),
  ProductEntryComplete: () => import('../views/drama/ProductEntryComplete.vue'),
  ComplianceAnalysis: () => import('../views/drama/ComplianceAnalysis.vue'),
  DataAnalysis: () => import('../views/drama/DataAnalysis.vue'),
  ScriptEdit: () => import('../views/script/ScriptEdit.vue'),
  ImageGeneration: () => import('../views/generation/ImageGeneration.vue'),
  TimelineEditor: () => import('../views/editor/TimelineEditor.vue'),
  DramaManagement: () => import('../views/drama/DramaManagement.vue'),
  EpisodeWorkflow: () => import('../views/drama/EpisodeWorkflow.vue'),
  CharacterExtraction: () => import('../views/workflow/CharacterExtraction.vue'),
  CharacterImages: () => import('../views/workflow/CharacterImages.vue'),
  DramaSettings: () => import('../views/workflow/DramaSettings.vue'),
  StoryboardEdit: () => import('../views/storyboard/StoryboardEdit.vue'),
  ProfessionalEditor: () => import('../views/drama/ProfessionalEditor.vue'),
  AIConfig: () => import('../views/settings/AIConfig.vue'),
  SystemSettings: () => import('../views/settings/SystemSettings.vue')
} satisfies Record<string, RouteLoader>

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: viewLoaders.HomeLanding
  },
  {
    path: '/products',
    name: 'Product',
    component: viewLoaders.ProductLanding
  },
  {
    path: '/pricing',
    redirect: '/'
  },
  {
    path: '/about',
    name: 'About',
    component: viewLoaders.AboutLanding
  },
  // 丝路 Agent 使用“过渡页 + 结果页”两段式路由，让流式分析过程和最终结构化方案分开承载。
  {
    path: '/agent/transition',
    name: 'AgentTransition',
    component: viewLoaders.AgentTransitionPage
  },
  {
    path: '/agent/result',
    name: 'AgentResult',
    component: viewLoaders.AgentResultPage
  },
  {
    path: '/agent',
    name: 'AgentEntry',
    redirect: '/agent/transition'
  },
  // 商品录入仍复用历史 /dramas/create 入口，但业务语义已经转为数字丝路的商品基础信息第一步。
  {
    path: '/dramas',
    name: 'DramaList',
    component: viewLoaders.WorkbenchDashboard
  },
  {
    path: '/dramas/create',
    name: 'DramaCreate',
    component: viewLoaders.DramaCreate
  },
  // 商品录入后续步骤拆成独立路由，便于每一步把草稿落到 sessionStorage 后再进入下一步。
  {
    path: '/product-entry/market',
    name: 'ProductEntryMarket',
    component: viewLoaders.ProductEntryMarket
  },
  {
    path: '/product-entry/details',
    name: 'ProductEntryDetails',
    component: viewLoaders.ProductEntryDetails
  },
  {
    path: '/product-entry/complete',
    name: 'ProductEntryComplete',
    component: viewLoaders.ProductEntryComplete
  },
  {
    path: '/compliance',
    name: 'ComplianceAnalysis',
    component: viewLoaders.ComplianceAnalysis
  },
  {
    path: '/analytics',
    name: 'DataAnalysis',
    component: viewLoaders.DataAnalysis
  },
  {
    path: '/settings',
    name: 'SystemSettings',
    component: viewLoaders.SystemSettings
  },
  {
    path: '/settings/ai',
    name: 'AIConfig',
    component: viewLoaders.AIConfig
  },
  {
    path: '/workspace/script',
    name: 'WorkspaceScript',
    component: viewLoaders.ScriptEdit
  },
  {
    path: '/workspace/content',
    name: 'WorkspaceContent',
    component: viewLoaders.ImageGeneration
  },
  {
    path: '/workspace/timeline',
    name: 'WorkspaceTimeline',
    component: viewLoaders.TimelineEditor
  },
  {
    path: '/dramas/:id/script',
    name: 'DramaScriptStage',
    component: viewLoaders.ScriptEdit
  },
  {
    path: '/dramas/:id',
    name: 'DramaManagement',
    component: viewLoaders.DramaManagement
  },
  {
    path: '/dramas/:id/episode/:episodeNumber',
    name: 'EpisodeWorkflowNew',
    component: viewLoaders.EpisodeWorkflow
  },
  {
    path: '/dramas/:id/characters',
    name: 'CharacterExtraction',
    component: viewLoaders.CharacterExtraction
  },
  {
    path: '/dramas/:id/images/characters',
    name: 'CharacterImages',
    component: viewLoaders.CharacterImages
  },
  {
    path: '/dramas/:id/settings',
    name: 'DramaSettings',
    component: viewLoaders.DramaSettings
  },
  {
    path: '/episodes/:id/edit',
    name: 'ScriptEdit',
    component: viewLoaders.ScriptEdit
  },
  {
    path: '/episodes/:id/storyboard',
    name: 'StoryboardEdit',
    component: viewLoaders.StoryboardEdit
  },
  {
    path: '/episodes/:id/generate',
    name: 'Generation',
    component: viewLoaders.ImageGeneration
  },
  {
    path: '/timeline/:id',
    name: 'TimelineEditor',
    component: viewLoaders.TimelineEditor
  },
  {
    path: '/dramas/:dramaId/episode/:episodeNumber/professional',
    name: 'ProfessionalEditor',
    component: viewLoaders.ProfessionalEditor
  }
]

const preloadedRouteKeys = new Set<string>()
// 工作流页面体积较大，用户悬停导航或浏览器空闲时先加载，减少从商品录入跳到合规/创作页面的等待感。
const workspacePreloadEntries: Array<{ key: string; matcher: RegExp; loader: RouteLoader }> = [
  { key: 'workbench', matcher: /^\/dramas\/?(?:[?#].*)?$/, loader: viewLoaders.WorkbenchDashboard },
  { key: 'product-entry-basic', matcher: /^\/dramas\/create\/?(?:[?#].*)?$/, loader: viewLoaders.DramaCreate },
  { key: 'product-entry-market', matcher: /^\/product-entry\/market\/?(?:[?#].*)?$/, loader: viewLoaders.ProductEntryMarket },
  { key: 'product-entry-details', matcher: /^\/product-entry\/details\/?(?:[?#].*)?$/, loader: viewLoaders.ProductEntryDetails },
  { key: 'product-entry-complete', matcher: /^\/product-entry\/complete\/?(?:[?#].*)?$/, loader: viewLoaders.ProductEntryComplete },
  { key: 'compliance', matcher: /^\/compliance\/?(?:[?#].*)?$/, loader: viewLoaders.ComplianceAnalysis },
  { key: 'workspace-script', matcher: /^\/workspace\/script\/?(?:[?#].*)?$/, loader: viewLoaders.ScriptEdit },
  { key: 'episode-script', matcher: /^\/episodes\/[^/]+\/edit\/?(?:[?#].*)?$/, loader: viewLoaders.ScriptEdit },
  { key: 'workspace-content', matcher: /^\/workspace\/content\/?(?:[?#].*)?$/, loader: viewLoaders.ImageGeneration },
  { key: 'episode-content', matcher: /^\/episodes\/[^/]+\/generate\/?(?:[?#].*)?$/, loader: viewLoaders.ImageGeneration },
  { key: 'workspace-timeline', matcher: /^\/workspace\/timeline\/?(?:[?#].*)?$/, loader: viewLoaders.TimelineEditor },
  { key: 'episode-timeline', matcher: /^\/timeline\/[^/]+\/?(?:[?#].*)?$/, loader: viewLoaders.TimelineEditor },
  { key: 'analytics', matcher: /^\/analytics\/?(?:[?#].*)?$/, loader: viewLoaders.DataAnalysis },
  { key: 'settings', matcher: /^\/settings\/?(?:[?#].*)?$/, loader: viewLoaders.SystemSettings },
  { key: 'settings-ai', matcher: /^\/settings\/ai\/?(?:[?#].*)?$/, loader: viewLoaders.AIConfig }
]

const runRouteLoader = (key: string, loader: RouteLoader) => {
  if (preloadedRouteKeys.has(key)) {
    return
  }
  // 同一路由只触发一次预加载，失败后释放 key，避免临时网络错误永久阻塞后续重试。
  preloadedRouteKeys.add(key)
  void loader().catch(() => {
    preloadedRouteKeys.delete(key)
  })
}

/**
 * 功能：按目标路径预加载数字丝路工作流页面。
 * 参数：path 为即将访问的前端路径，可能来自导航悬停、聚焦或按钮点击前置动作。
 * 返回：无返回值；命中预加载表时异步触发对应视图模块加载。
 */
export const preloadRouteByPath = (path?: string) => {
  if (!path) {
    return
  }
  const target = workspacePreloadEntries.find((entry) => entry.matcher.test(path))
  if (target) {
    runRouteLoader(target.key, target.loader)
  }
}

/**
 * 功能：批量预加载数字丝路工作区核心页面。
 * 参数：无。
 * 返回：无返回值；用于浏览器空闲时提前准备商品录入、合规和内容生产页面。
 */
export const preloadWorkspaceRoutes = () => {
  workspacePreloadEntries.forEach((entry) => {
    runRouteLoader(entry.key, entry.loader)
  })
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to) {
    if (to.hash) {
      return {
        el: to.hash,
        top: 80,
        behavior: 'smooth'
      }
    }

    return { top: 0, left: 0 }
  }
})

if (typeof window !== 'undefined') {
  const preload = () => preloadWorkspaceRoutes()
  if (typeof window.requestIdleCallback === 'function') {
    window.requestIdleCallback(preload, { timeout: 2000 })
  } else {
    window.setTimeout(preload, 800)
  }
}

// 开源版本 - 无需认证

export default router
