<template>
  <div class="silk-workspace">
    <aside class="silk-sidebar">
      <router-link class="silk-brand" to="/dashboard">
        <img src="/logo_circle.png" alt="" />
        <span><strong>数字丝路</strong><small>AI Agent 智能营销引擎</small></span>
      </router-link>
      <nav aria-label="数字丝路主导航">
        <router-link v-for="item in navigation" :key="item.path" :to="item.path">
          <el-icon><component :is="item.icon" /></el-icon><span>{{ item.label }}</span>
        </router-link>
      </nav>
      <div class="silk-sidebar__footer">
        <router-link to="/settings"><el-icon><Setting /></el-icon><span>系统设置</span></router-link>
        <LanguageSwitcher />
      </div>
    </aside>
    <div class="silk-body">
      <header class="silk-topbar">
        <div>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">数字丝路</el-breadcrumb-item>
            <el-breadcrumb-item>{{ title }}</el-breadcrumb-item>
          </el-breadcrumb>
          <strong>{{ title }}</strong>
        </div>
        <div class="silk-topbar__actions"><NotificationBell /><slot name="actions" /></div>
      </header>
      <main class="silk-main"><slot /></main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Aim, DataAnalysis, FolderOpened, MagicStick, Monitor, Picture, Setting, User } from '@element-plus/icons-vue'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import NotificationBell from '@/components/common/NotificationBell.vue'

defineProps<{ title: string }>()

const navigation = [
  { label: '工作台', path: '/dashboard', icon: Monitor },
  { label: '丝路 Agent', path: '/agent', icon: MagicStick },
  { label: '营销项目', path: '/projects', icon: FolderOpened },
  { label: '合规中心', path: '/compliance', icon: Aim },
  { label: '媒体生成', path: '/media/image', icon: Picture },
  { label: '数字人中心', path: '/digital-human', icon: User },
  { label: '数据分析', path: '/analytics', icon: DataAnalysis }
]
</script>

<style scoped>
.silk-workspace{min-height:100vh;display:grid;grid-template-columns:252px 1fr;background:var(--bg-primary)}
.silk-sidebar{position:sticky;top:0;height:100vh;padding:24px 18px;display:flex;flex-direction:column;background:#102a5c;color:#fff;box-shadow:12px 0 36px rgba(24,54,109,.12);z-index:20}
.silk-brand{display:flex;align-items:center;gap:12px;color:#fff;text-decoration:none;padding:4px 8px 28px}.silk-brand img{width:42px;height:42px}.silk-brand span{display:flex;flex-direction:column}.silk-brand strong{font-size:20px}.silk-brand small{margin-top:3px;opacity:.7;font-size:11px}
nav{display:flex;flex-direction:column;gap:7px}nav a,.silk-sidebar__footer a{display:flex;align-items:center;gap:12px;padding:13px 14px;color:rgba(255,255,255,.72);text-decoration:none;border-radius:13px;font-weight:600;transition:.2s}nav a:hover,nav a.router-link-active,.silk-sidebar__footer a:hover{color:#fff;background:rgba(255,255,255,.12)}
.silk-sidebar__footer{margin-top:auto;display:grid;gap:10px}.silk-body{min-width:0}.silk-topbar{height:88px;padding:16px 32px;display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid var(--border-primary);background:var(--bg-card);backdrop-filter:blur(18px);position:sticky;top:0;z-index:15}.silk-topbar>div:first-child{display:grid;gap:8px}.silk-topbar strong{font-size:19px}.silk-topbar__actions{display:flex;align-items:center;gap:12px}.silk-main{padding:30px 32px 56px;max-width:1500px;width:100%;margin:0 auto}
@media(max-width:900px){.silk-workspace{display:block}.silk-sidebar{position:relative;width:100%;height:auto;padding:14px 16px}.silk-brand{padding-bottom:12px}.silk-sidebar nav{flex-direction:row;overflow-x:auto}.silk-sidebar nav a{white-space:nowrap}.silk-sidebar__footer{display:none}.silk-topbar{height:72px;padding:12px 18px}.silk-main{padding:20px 16px 40px}}
</style>
