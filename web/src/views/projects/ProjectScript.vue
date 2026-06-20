<template>
  <SilkRoadWorkspace title="营销脚本与分镜">
    <template #actions><el-button :loading="saving" @click="save">保存脚本</el-button><el-button type="primary" :disabled="!activeVersion" :loading="generating" @click="generateShots">生成营销分镜</el-button></template>
    <ProjectStageNav :project-id="projectId" />
    <section class="script-head"><div><span>MARKETING CONTENT</span><h1>{{ data?.project_name || '营销内容工作区' }}</h1><p>用目标市场语言组织商品卖点、口播、字幕和镜头节奏。</p></div><el-select v-if="versions.length" v-model="activeId" placeholder="选择内容版本"><el-option v-for="item in versions" :key="item.id" :label="`版本 ${item.version_number} · ${item.title}`" :value="item.id" /></el-select></section>
    <el-skeleton v-if="loading" :rows="9" animated />
    <el-empty v-else-if="!activeVersion" description="暂无营销内容版本"><el-button type="primary" @click="createVersion">新建营销脚本</el-button></el-empty>
    <template v-else>
      <section class="editor-card"><el-input v-model="activeVersion.title" placeholder="内容版本标题" /><el-input v-model="activeVersion.script_content" type="textarea" :rows="14" placeholder="编写商品开场钩子、卖点证明、使用场景、信任信息和行动引导..." /><div class="script-meta"><span>建议结构：开场吸引 → 商品痛点 → 卖点演示 → 合规证明 → 行动引导</span><strong>{{ activeVersion.script_content?.length || 0 }} 字</strong></div></section>
      <section class="shots-card"><header><div><h2>营销视频分镜</h2><p>每个镜头围绕商品表达和目标市场适配展开。</p></div><el-button @click="addShot">添加镜头</el-button></header>
        <el-table :data="activeVersion.shots || []" empty-text="生成或手动添加第一个营销镜头">
          <el-table-column label="镜头序号" width="92"><template #default="scope">{{ scope.row.shot_number || scope.$index + 1 }}</template></el-table-column>
          <el-table-column label="画面描述" min-width="180"><template #default="scope"><el-input v-model="scope.row.visual" type="textarea" :rows="2" placeholder="商品使用场景与视觉构图" /></template></el-table-column>
          <el-table-column label="商品卖点" min-width="150"><template #default="scope"><el-input v-model="scope.row.selling_point" placeholder="本镜头传达的卖点" /></template></el-table-column>
          <el-table-column label="口播文案 / 字幕" min-width="220"><template #default="scope"><el-input v-model="scope.row.voiceover" placeholder="口播文案" /><el-input v-model="scope.row.subtitle" class="sub-input" placeholder="本地化字幕" /></template></el-table-column>
          <el-table-column label="数字人动作" min-width="140"><template #default="scope"><el-input v-model="scope.row.digital_human_action" placeholder="手势、表情、动作" /></template></el-table-column>
          <el-table-column label="素材来源" min-width="130"><template #default="scope"><el-input v-model="scope.row.source" placeholder="商品图 / AI生成" /></template></el-table-column>
          <el-table-column label="市场适配说明" min-width="170"><template #default="scope"><el-input v-model="scope.row.market_adaptation" placeholder="文化、语言或平台适配" /></template></el-table-column>
        </el-table>
      </section>
    </template>
  </SilkRoadWorkspace>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'; import { ElMessage } from 'element-plus'; import { useRoute } from 'vue-router'; import SilkRoadWorkspace from '@/layouts/SilkRoadWorkspace.vue'; import ProjectStageNav from '@/components/projects/ProjectStageNav.vue'; import { marketingContentAPI } from '@/api/marketing-content'; import type { MarketingContentVersion, ProjectScriptResponse } from '@/types/project'
const route=useRoute(); const projectId=String(route.params.id); const data=ref<ProjectScriptResponse>(); const versions=ref<MarketingContentVersion[]>([]); const activeId=ref<number>(); const loading=ref(false); const saving=ref(false); const generating=ref(false)
const activeVersion=computed(()=>versions.value.find(v=>v.id===activeId.value)); const createVersion=()=>{const next={id:Date.now(),project_id:Number(projectId),version_number:versions.value.length+1,title:`营销内容版本 ${versions.value.length+1}`,script_content:'',status:'draft',shots:[]} as MarketingContentVersion;versions.value.push(next);activeId.value=next.id}
const addShot=()=>{if(!activeVersion.value)return;activeVersion.value.shots ||= [];activeVersion.value.shots.push({id:Date.now(),shot_number:activeVersion.value.shots.length+1,visual:'',selling_point:'',voiceover:'',subtitle:'',digital_human_action:'',source:'',market_adaptation:''})}
const save=async()=>{saving.value=true;try{await marketingContentAPI.saveScript(projectId,versions.value);ElMessage.success('营销脚本与分镜已保存')}catch(e:any){ElMessage.error(e?.message||'保存失败')}finally{saving.value=false}}
const generateShots=async()=>{if(!activeVersion.value)return;generating.value=true;try{await marketingContentAPI.generateShotPlan(activeVersion.value.id);ElMessage.success('营销分镜生成任务已提交');await load()}catch(e:any){ElMessage.error(e?.message||'分镜生成失败')}finally{generating.value=false}}
const load=async()=>{loading.value=true;try{data.value=await marketingContentAPI.getScript(projectId);versions.value=data.value.content_versions||[];if(!versions.value.length)createVersion();else activeId.value=versions.value[0].id}catch(e:any){ElMessage.error(e?.message||'脚本加载失败')}finally{loading.value=false}};onMounted(load)
</script>
<style scoped>
.script-head{display:flex;justify-content:space-between;align-items:end;margin-bottom:20px}.script-head span{font-size:12px;letter-spacing:.16em;color:var(--accent);font-weight:800}.script-head h1{font-size:28px;margin:8px 0}.script-head p,.shots-card header p{color:var(--text-secondary)}.script-head .el-select{width:280px}.editor-card,.shots-card{padding:22px;border:1px solid var(--border-primary);border-radius:20px;background:var(--bg-card);margin-bottom:18px}.editor-card>.el-input:first-child{margin-bottom:12px}.script-meta{display:flex;justify-content:space-between;margin-top:12px;font-size:13px;color:var(--text-muted)}.shots-card header{display:flex;align-items:center;justify-content:space-between;margin-bottom:16px}.shots-card h2{margin-bottom:5px}.sub-input{margin-top:7px}@media(max-width:700px){.script-head{align-items:stretch;flex-direction:column;gap:14px}.script-head .el-select{width:100%}}
</style>
