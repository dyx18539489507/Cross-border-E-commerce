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
  ProfessionalEditor: () => import('../views/drama/ProfessionalEditor.vue')
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
  { key: 'analytics', matcher: /^\/analytics\/?(?:[?#].*)?$/, loader: viewLoaders.DataAnalysis }
]

const runRouteLoader = (key: string, loader: RouteLoader) => {
  if (preloadedRouteKeys.has(key)) {
    return
  }
  preloadedRouteKeys.add(key)
  void loader().catch(() => {
    preloadedRouteKeys.delete(key)
  })
}

export const preloadRouteByPath = (path?: string) => {
  if (!path) {
    return
  }
  const target = workspacePreloadEntries.find((entry) => entry.matcher.test(path))
  if (target) {
    runRouteLoader(target.key, target.loader)
  }
}

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
