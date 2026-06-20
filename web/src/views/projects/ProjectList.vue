<template>
  <SilkRoadWorkspace title="跨境营销项目">
    <template #actions><el-button type="primary" @click="router.push('/projects/create')"><el-icon><Plus /></el-icon>创建营销项目</el-button></template>
    <section class="hero-row">
      <div><span class="eyebrow">PROJECT PORTFOLIO</span><h1>营销项目</h1><p>集中管理商品、目标市场、营销方案、生成任务与合规状态。</p></div>
      <el-input v-model="keyword" clearable placeholder="搜索商品或目标市场" :prefix-icon="Search" @keyup.enter="loadProjects" />
    </section>
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
    <div v-loading="loading" class="project-grid">
      <article v-for="project in filteredProjects" :key="project.id" class="project-card" @click="router.push(`/projects/${project.id}`)">
        <div class="project-card__visual">
          <img v-if="project.product_image || project.thumbnail" :src="project.product_image || project.thumbnail" :alt="project.product_name" />
          <el-icon v-else><Goods /></el-icon>
          <span :class="`risk risk--${project.compliance_status}`">{{ complianceText(project.compliance_status) }}</span>
        </div>
        <div class="project-card__body">
          <div class="project-card__title"><h2>{{ project.product_name }}</h2><el-tag effect="plain">{{ statusText(project.status) }}</el-tag></div>
          <p>{{ project.product_description || '等待补充商品介绍与营销目标' }}</p>
          <dl><div><dt>目标市场</dt><dd>{{ project.target_markets.join('、') || '待设置' }}</dd></div><div><dt>内容版本</dt><dd>{{ project.content_versions?.length || 0 }}</dd></div></dl>
        </div>
      </article>
    </div>
    <el-empty v-if="!loading && !filteredProjects.length" description="暂无跨境营销项目">
      <el-button type="primary" @click="router.push('/projects/create')">创建第一个营销项目</el-button>
      <el-button @click="router.push('/agent')">交给丝路 Agent</el-button>
    </el-empty>
  </SilkRoadWorkspace>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Goods, Plus, Search } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import SilkRoadWorkspace from '@/layouts/SilkRoadWorkspace.vue'
import { projectAPI } from '@/api/project'
import type { MarketingProject } from '@/types/project'

const router = useRouter()
const projects = ref<MarketingProject[]>([])
const keyword = ref('')
const loading = ref(false)
const error = ref('')
const filteredProjects = computed(() => {
  const term = keyword.value.trim().toLowerCase()
  if (!term) return projects.value
  return projects.value.filter((item) => [item.product_name, item.project_name, ...item.target_markets].join(' ').toLowerCase().includes(term))
})
const complianceText = (status?: string) => ({ green: '合规通过', yellow: '需优化', red: '高风险', pending: '待分析' }[status || 'pending'] || '待分析')
const statusText = (status: string) => ({ draft: '方案草稿', planning: '策划中', production: '生成中', completed: '已完成', archived: '已归档' }[status] || status)
const loadProjects = async () => {
  loading.value = true; error.value = ''
  try { projects.value = (await projectAPI.list({ page_size: 100 })).items } catch (e: any) { error.value = e?.message || '营销项目加载失败，请稍后重试。' } finally { loading.value = false }
}
onMounted(loadProjects)
</script>

<style scoped>
.hero-row{display:flex;justify-content:space-between;align-items:end;gap:28px;margin-bottom:26px}.hero-row h1{font-size:34px;margin:7px 0}.hero-row p{color:var(--text-secondary)}.hero-row .el-input{width:320px}.eyebrow{font-size:12px;letter-spacing:.18em;color:var(--accent);font-weight:800}.project-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(310px,1fr));gap:20px;min-height:160px}.project-card{overflow:hidden;border:1px solid var(--border-primary);border-radius:22px;background:var(--bg-card);box-shadow:var(--shadow-card);cursor:pointer;transition:.2s}.project-card:hover{transform:translateY(-4px);box-shadow:var(--shadow-card-hover)}.project-card__visual{height:156px;display:grid;place-items:center;position:relative;background:linear-gradient(135deg,#dbeafe,#ede9fe)}.project-card__visual img{width:100%;height:100%;object-fit:cover}.project-card__visual>.el-icon{font-size:58px;color:#5570c9}.risk{position:absolute;top:14px;right:14px;padding:5px 10px;border-radius:99px;background:#fff;font-size:12px;font-weight:800}.risk--green{color:#16885c}.risk--yellow{color:#b66d00}.risk--red{color:#d14343}.project-card__body{padding:20px}.project-card__title{display:flex;align-items:center;justify-content:space-between;gap:12px}.project-card h2{font-size:19px}.project-card p{margin:12px 0 18px;color:var(--text-secondary);height:42px;overflow:hidden}.project-card dl{display:grid;grid-template-columns:1fr 1fr;gap:12px}.project-card dl div{padding:10px;background:var(--bg-muted);border-radius:12px}.project-card dt{font-size:12px;color:var(--text-muted)}.project-card dd{margin-top:5px;font-weight:700}@media(max-width:700px){.hero-row{align-items:stretch;flex-direction:column}.hero-row .el-input{width:100%}}
</style>
