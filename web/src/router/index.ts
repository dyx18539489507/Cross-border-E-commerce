import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/home/HomeLanding.vue')
  },
  {
    path: '/products',
    name: 'Product',
    component: () => import('../views/home/ProductLanding.vue')
  },
  {
    path: '/pricing',
    redirect: '/'
  },
  {
    path: '/about',
    name: 'About',
    component: () => import('../views/home/AboutLanding.vue')
  },
  {
    path: '/dramas',
    name: 'DramaList',
    component: () => import('../views/drama/WorkbenchDashboard.vue')
  },
  {
    path: '/dramas/create',
    name: 'DramaCreate',
    component: () => import('../views/drama/DramaCreate.vue')
  },
  {
    path: '/compliance',
    name: 'ComplianceAnalysis',
    component: () => import('../views/drama/ComplianceAnalysis.vue')
  },
  {
    path: '/analytics',
    name: 'DataAnalysis',
    component: () => import('../views/drama/DataAnalysis.vue')
  },
  {
    path: '/workspace/script',
    name: 'WorkspaceScript',
    component: () => import('../views/script/ScriptEdit.vue')
  },
  {
    path: '/workspace/content',
    name: 'WorkspaceContent',
    component: () => import('../views/generation/ImageGeneration.vue')
  },
  {
    path: '/workspace/timeline',
    name: 'WorkspaceTimeline',
    component: () => import('../views/editor/TimelineEditor.vue')
  },
  {
    path: '/dramas/:id/script',
    name: 'DramaScriptStage',
    component: () => import('../views/script/ScriptEdit.vue')
  },
  {
    path: '/dramas/:id',
    name: 'DramaManagement',
    component: () => import('../views/drama/DramaManagement.vue')
  },
  {
    path: '/dramas/:id/episode/:episodeNumber',
    name: 'EpisodeWorkflowNew',
    component: () => import('../views/drama/EpisodeWorkflow.vue')
  },
  {
    path: '/dramas/:id/characters',
    name: 'CharacterExtraction',
    component: () => import('../views/workflow/CharacterExtraction.vue')
  },
  {
    path: '/dramas/:id/images/characters',
    name: 'CharacterImages',
    component: () => import('../views/workflow/CharacterImages.vue')
  },
  {
    path: '/dramas/:id/settings',
    name: 'DramaSettings',
    component: () => import('../views/workflow/DramaSettings.vue')
  },
  {
    path: '/episodes/:id/edit',
    name: 'ScriptEdit',
    component: () => import('../views/script/ScriptEdit.vue')
  },
  {
    path: '/episodes/:id/storyboard',
    name: 'StoryboardEdit',
    component: () => import('../views/storyboard/StoryboardEdit.vue')
  },
  {
    path: '/episodes/:id/generate',
    name: 'Generation',
    component: () => import('../views/generation/ImageGeneration.vue')
  },
  {
    path: '/timeline/:id',
    name: 'TimelineEditor',
    component: () => import('../views/editor/TimelineEditor.vue')
  },
  {
    path: '/dramas/:dramaId/episode/:episodeNumber/professional',
    name: 'ProfessionalEditor',
    component: () => import('../views/drama/ProfessionalEditor.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior() {
    return { top: 0, left: 0 }
  }
})

// 开源版本 - 无需认证

export default router
