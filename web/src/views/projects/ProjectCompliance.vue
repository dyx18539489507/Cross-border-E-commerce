<template>
  <SilkRoadWorkspace title="项目合规分析"><ProjectStageNav :project-id="projectId" />
    <el-skeleton v-if="loading" :rows="7" animated />
    <template v-else-if="project">
      <section class="page-head"><div><span>COMPLIANCE INTELLIGENCE</span><h1>{{ project.product_name }} · 目标市场合规</h1><p>结合商品资料、宣传卖点和目标市场规则识别营销风险。</p></div><el-button type="primary" :loading="analyzing" @click="analyze">重新分析</el-button></section>
      <div class="compliance-grid">
        <article class="score-card"><el-progress type="dashboard" :percentage="result?.score ?? project.compliance_score ?? 0" :color="scoreColor" :width="176" /><h2>{{ result?.level_label || levelLabel }}</h2><p>{{ result?.summary || '点击重新分析，生成最新市场合规结论。' }}</p></article>
        <article><h2>风险关注点</h2><ul v-if="result?.non_compliance_points?.length"><li v-for="item in result.non_compliance_points" :key="item">{{ item }}</li></ul><el-empty v-else description="暂未发现明确风险项" :image-size="72" /></article>
        <article><h2>整改与优化建议</h2><ol v-if="result?.rectification_suggestions?.length"><li v-for="item in result.rectification_suggestions" :key="item">{{ item }}</li></ol><el-empty v-else description="运行分析后显示优化建议" :image-size="72" /></article>
      </div>
    </template>
  </SilkRoadWorkspace>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'; import { ElMessage } from 'element-plus'; import { useRoute } from 'vue-router'; import SilkRoadWorkspace from '@/layouts/SilkRoadWorkspace.vue'; import ProjectStageNav from '@/components/projects/ProjectStageNav.vue'; import { projectAPI } from '@/api/project'; import { complianceAPI } from '@/api/compliance'; import type { ComplianceResult, MarketingProject } from '@/types/project'
const route=useRoute(); const projectId=String(route.params.id); const project=ref<MarketingProject>(); const result=ref<ComplianceResult>(); const loading=ref(false); const analyzing=ref(false)
const levelLabel=computed(()=>({green:'低风险',yellow:'中风险',red:'高风险',pending:'待分析'}[project.value?.compliance_status||'pending'])); const scoreColor=computed(()=>result.value?.level==='red'?'#e85d54':result.value?.level==='yellow'?'#f0a020':'#22b36f')
const analyze=async()=>{if(!project.value)return; analyzing.value=true;try{const data=await complianceAPI.analyze({product_name:project.value.product_name,product_description:project.value.product_description||project.value.product_name,target_markets:project.value.target_markets,product_selling_points:project.value.product_selling_points,material_composition:project.value.material_composition,compliance_focus:project.value.compliance_focus});result.value=data.compliance;ElMessage.success('合规分析已更新')}catch(e:any){ElMessage.error(e?.message||'合规分析失败')}finally{analyzing.value=false}}
onMounted(async()=>{loading.value=true;try{project.value=await projectAPI.get(projectId)}finally{loading.value=false}})
</script>
<style scoped>
.page-head{display:flex;align-items:end;justify-content:space-between;margin-bottom:22px}.page-head span{font-size:12px;letter-spacing:.16em;color:var(--accent);font-weight:800}.page-head h1{font-size:29px;margin:8px 0}.page-head p{color:var(--text-secondary)}.compliance-grid{display:grid;grid-template-columns:320px 1fr;gap:18px}.compliance-grid article{padding:25px;border:1px solid var(--border-primary);border-radius:20px;background:var(--bg-card)}.score-card{grid-row:span 2;text-align:center}.score-card h2{margin:12px}.score-card p{color:var(--text-secondary);line-height:1.7}.compliance-grid ul,.compliance-grid ol{padding:14px 0 0 22px;color:var(--text-secondary)}.compliance-grid li{padding:7px;line-height:1.6}@media(max-width:800px){.compliance-grid{grid-template-columns:1fr}.score-card{grid-row:auto}}
</style>
