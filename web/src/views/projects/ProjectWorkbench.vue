<template>
  <SilkRoadWorkspace :title="project?.product_name || '项目工作台'">
    <template #actions><el-button @click="loadProject">刷新项目</el-button></template>
    <ProjectStageNav :project-id="projectId" />
    <el-skeleton v-if="loading" :rows="8" animated />
    <el-result v-else-if="error" icon="error" title="项目加载失败" :sub-title="error"><template #extra><el-button @click="loadProject">重试</el-button></template></el-result>
    <template v-else-if="project">
      <section class="project-hero">
        <div><span class="eyebrow">MARKETING WORKBENCH</span><h1>{{ project.product_name }}</h1><p>{{ project.product_description || '暂无商品介绍' }}</p><div class="tags"><el-tag v-for="market in project.target_markets" :key="market">{{ market }}</el-tag><el-tag type="success">{{ complianceLabel }}</el-tag></div></div>
        <div class="score"><strong>{{ project.compliance_score || 0 }}</strong><span>合规评分</span></div>
      </section>
      <section class="facts">
        <article><span>商品卖点</span><strong>{{ project.product_selling_points || '待完善' }}</strong></article>
        <article><span>目标语言</span><strong>{{ project.target_language || '待设置' }}</strong></article>
        <article><span>营销风格</span><strong>{{ project.marketing_style || '待设置' }}</strong></article>
        <article><span>平台渠道</span><strong>{{ project.platform_channels.join('、') || '待设置' }}</strong></article>
      </section>
      <section class="workflow-grid">
        <button v-for="stage in stages" :key="stage.path" @click="router.push(stage.path)"><span class="stage-icon"><el-icon><component :is="stage.icon" /></el-icon></span><div><h2>{{ stage.title }}</h2><p>{{ stage.description }}</p></div><el-icon><ArrowRight /></el-icon></button>
      </section>
    </template>
  </SilkRoadWorkspace>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRight, DataAnalysis, Document, Film, FolderOpened, MagicStick, Picture, User } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import SilkRoadWorkspace from '@/layouts/SilkRoadWorkspace.vue'
import ProjectStageNav from '@/components/projects/ProjectStageNav.vue'
import { projectAPI } from '@/api/project'
import type { MarketingProject } from '@/types/project'

const route=useRoute(); const router=useRouter(); const projectId=String(route.params.id); const project=ref<MarketingProject>(); const loading=ref(false); const error=ref('')
const complianceLabel=computed(()=>({green:'合规通过',yellow:'需要优化',red:'高风险',pending:'待分析'}[project.value?.compliance_status||'pending']))
const base=`/projects/${projectId}`
const stages=[
  {title:'合规分析',description:'检查目标市场规则、风险表述与整改建议',path:`${base}/compliance`,icon:DataAnalysis},
  {title:'营销脚本与分镜',description:'编辑口播文案、卖点节奏和营销视频镜头',path:`${base}/script`,icon:Document},
  {title:'图片与视频素材',description:'管理商品图、广告图、视频和音频素材',path:`${base}/assets`,icon:Picture},
  {title:'数字人讲解',description:'配置商品讲解人、口播音色与动作',path:`/digital-human?projectId=${projectId}`,icon:User},
  {title:'营销视频剪辑',description:'在时间线上编排视频、字幕、音乐与音效',path:`${base}/editor`,icon:Film},
  {title:'生成任务与分发',description:'跟踪内容生成、导出和平台分发进度',path:`${base}/tasks`,icon:MagicStick},
  {title:'数据反馈',description:'查看营销内容生产和渠道反馈指标',path:`/analytics?projectId=${projectId}`,icon:FolderOpened}
]
const loadProject=async()=>{loading.value=true;error.value='';try{project.value=await projectAPI.get(projectId)}catch(e:any){error.value=e?.message||'无法读取营销项目'}finally{loading.value=false}}
onMounted(loadProject)
</script>

<style scoped>
.project-hero{display:flex;justify-content:space-between;align-items:center;gap:28px;padding:30px;border-radius:24px;background:linear-gradient(125deg,#102a5c,#3d55c6);color:#fff;box-shadow:var(--shadow-lg)}.eyebrow{font-size:12px;letter-spacing:.17em;color:#93c5fd}.project-hero h1{font-size:34px;margin:8px 0}.project-hero p{max-width:760px;color:rgba(255,255,255,.72)}.tags{display:flex;gap:8px;margin-top:18px}.score{width:120px;height:120px;flex:0 0 120px;border:1px solid rgba(255,255,255,.25);border-radius:50%;display:grid;place-content:center;text-align:center;background:rgba(255,255,255,.1)}.score strong{font-size:38px}.score span{font-size:12px;opacity:.7}.facts{display:grid;grid-template-columns:repeat(4,1fr);gap:14px;margin:20px 0}.facts article{padding:18px;border:1px solid var(--border-primary);background:var(--bg-card);border-radius:16px}.facts span{display:block;font-size:12px;color:var(--text-muted);margin-bottom:8px}.workflow-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:16px}.workflow-grid button{display:grid;grid-template-columns:auto 1fr auto;align-items:center;gap:16px;text-align:left;padding:22px;border:1px solid var(--border-primary);border-radius:18px;background:var(--bg-card);color:var(--text-primary);cursor:pointer}.workflow-grid button:hover{border-color:var(--accent);transform:translateY(-2px)}.stage-icon{width:48px;height:48px;border-radius:14px;display:grid;place-items:center;background:var(--accent-light);color:var(--accent);font-size:22px}.workflow-grid h2{font-size:17px;margin-bottom:6px}.workflow-grid p{color:var(--text-secondary)}@media(max-width:900px){.facts{grid-template-columns:repeat(2,1fr)}.workflow-grid{grid-template-columns:1fr}}@media(max-width:600px){.project-hero{align-items:flex-start}.score{display:none}.facts{grid-template-columns:1fr}}
</style>
