<template>
  <SilkRoadWorkspace title="创建跨境营销项目">
    <section class="create-shell">
      <header><span class="eyebrow">NEW MARKETING PROJECT</span><h1>从商品信息开始</h1><p>填写目标市场与营销要求，系统将在创建前完成合规预检。</p></header>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="create-form">
        <div class="form-grid">
          <el-form-item label="商品名称" prop="product_name"><el-input v-model="form.product_name" placeholder="例如：便携式智能翻译耳机" /></el-form-item>
          <el-form-item label="目标市场" prop="target_markets"><el-select v-model="form.target_markets" multiple filterable placeholder="选择一个或多个市场"><el-option v-for="market in markets" :key="market.value" :label="market.label" :value="market.value" /></el-select></el-form-item>
          <el-form-item label="目标语言" prop="target_language"><el-select v-model="form.target_language" placeholder="选择内容语言"><el-option v-for="lang in languages" :key="lang" :label="lang" :value="lang" /></el-select></el-form-item>
          <el-form-item label="平台渠道" prop="platform_channels"><el-select v-model="form.platform_channels" multiple placeholder="选择主要分发渠道"><el-option v-for="platform in platforms" :key="platform" :label="platform" :value="platform" /></el-select></el-form-item>
        </div>
        <el-form-item label="商品介绍" prop="product_description"><el-input v-model="form.product_description" type="textarea" :rows="3" placeholder="说明用途、规格、适用人群和使用方式" /></el-form-item>
        <el-form-item label="商品卖点" prop="product_selling_points"><el-input v-model="form.product_selling_points" type="textarea" :rows="3" placeholder="填写希望重点传达的差异化优势，避免绝对化承诺" /></el-form-item>
        <div class="form-grid">
          <el-form-item label="营销风格"><el-select v-model="form.marketing_style"><el-option v-for="style in styles" :key="style" :label="style" :value="style" /></el-select></el-form-item>
          <el-form-item label="商品材质 / 成分"><el-input v-model="form.material_composition" placeholder="用于目标市场合规判断" /></el-form-item>
        </div>
        <el-form-item label="商品图片"><el-upload drag :auto-upload="false" :limit="1" accept="image/png,image/jpeg,image/webp" :on-change="handleImage"><el-icon><UploadFilled /></el-icon><div>拖入商品图片，或点击选择文件</div><template #tip>支持 JPG、PNG、WebP，图片将用于场景图和视频素材生成。</template></el-upload></el-form-item>
        <el-form-item label="合规关注点"><el-input v-model="form.compliance_focus" type="textarea" :rows="2" placeholder="例如：功效表述、环保宣称、儿童用品要求、平台广告政策" /></el-form-item>
        <el-alert v-if="compliance" :type="compliance.level === 'red' ? 'error' : compliance.level === 'yellow' ? 'warning' : 'success'" :closable="false" show-icon><template #title>合规评分 {{ compliance.score }}：{{ compliance.summary }}</template></el-alert>
        <div class="form-actions"><el-button @click="router.push('/projects')">取消</el-button><el-button :loading="checking" @click="runCompliance">{{ compliance ? '重新分析合规' : '先做合规分析' }}</el-button><el-button type="primary" :loading="creating" :disabled="!complianceToken || compliance?.level === 'red'" @click="createProject">创建营销项目</el-button></div>
      </el-form>
    </section>
  </SilkRoadWorkspace>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules, type UploadFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import SilkRoadWorkspace from '@/layouts/SilkRoadWorkspace.vue'
import { complianceAPI } from '@/api/compliance'
import { projectAPI } from '@/api/project'
import type { ComplianceResult, CreateProjectInput } from '@/types/project'

const router = useRouter(); const formRef = ref<FormInstance>(); const checking = ref(false); const creating = ref(false); const complianceToken = ref(''); const compliance = ref<ComplianceResult>()
const form = reactive<CreateProjectInput>({ product_name:'', product_description:'', target_markets:[], target_language:'英语', product_selling_points:'', material_composition:'', marketing_style:'真实产品演示', platform_channels:[], product_image:'', compliance_focus:'' })
const rules: FormRules = { product_name:[{required:true,message:'请输入商品名称',trigger:'blur'}], product_description:[{required:true,message:'请输入商品介绍',trigger:'blur'}], target_markets:[{required:true,message:'请选择目标市场',trigger:'change'}], target_language:[{required:true,message:'请选择目标语言',trigger:'change'}], product_selling_points:[{required:true,message:'请填写商品卖点',trigger:'blur'}], platform_channels:[{required:true,message:'请选择平台渠道',trigger:'change'}] }
const markets=[{label:'美国',value:'US'},{label:'英国',value:'GB'},{label:'德国',value:'DE'},{label:'法国',value:'FR'},{label:'日本',value:'JP'},{label:'新加坡',value:'SG'},{label:'中东',value:'AE'}]
const languages=['英语','德语','法语','日语','阿拉伯语','西班牙语']; const platforms=['TikTok','Instagram','YouTube','Amazon','Shopee','独立站']; const styles=['真实产品演示','达人口播','生活方式','科技质感','极简高级','本地化测评']
const validate = async () => { try { await formRef.value?.validate(); return true } catch { ElMessage.warning('请先完善必填信息'); return false } }
const handleImage = (file: UploadFile) => { if (!file.raw) return; const reader=new FileReader(); reader.onload=()=>{form.product_image=String(reader.result||'')}; reader.readAsDataURL(file.raw) }
const runCompliance = async () => { if (!(await validate())) return; checking.value=true; try { const result=await complianceAPI.analyze(form); compliance.value=result.compliance; complianceToken.value=result.compliance_token; ElMessage.success('目标市场合规分析已完成') } catch(e:any){ ElMessage.error(e?.message||'合规分析失败') } finally { checking.value=false } }
const createProject = async () => { if (!complianceToken.value) return; creating.value=true; try { const result=await projectAPI.create({...form,compliance_token:complianceToken.value}); ElMessage.success('营销项目创建成功'); router.push(`/projects/${result.project.id}`) } catch(e:any){ ElMessage.error(e?.message||'项目创建失败') } finally { creating.value=false } }
</script>

<style scoped>
.create-shell{max-width:980px;margin:0 auto}.create-shell>header{text-align:center;margin-bottom:28px}.create-shell h1{font-size:34px;margin:8px}.create-shell header p{color:var(--text-secondary)}.eyebrow{font-size:12px;letter-spacing:.18em;color:var(--accent);font-weight:800}.create-form{padding:30px;border:1px solid var(--border-primary);border-radius:24px;background:var(--bg-card);box-shadow:var(--shadow-card)}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:0 22px}.create-form :deep(.el-select){width:100%}.create-form :deep(.el-upload){width:100%}.create-form :deep(.el-upload-dragger){width:100%;background:var(--bg-muted)}.form-actions{display:flex;justify-content:flex-end;gap:10px;margin-top:24px}@media(max-width:700px){.form-grid{grid-template-columns:1fr}.create-form{padding:20px}.form-actions{flex-wrap:wrap}}
</style>
