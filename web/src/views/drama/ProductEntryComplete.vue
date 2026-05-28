<!--
/**
 * 模块说明：数字丝路商品录入完成页。
 * 业务场景：用户完成商品资料录入后，需要进入合规分析或回到工作台。
 * 核心职责：把最终商品草稿同步为后端创建/合规接口可识别的请求草稿，并承接下一步跳转。
 */
-->
<template>
  <ProductEntryFlowShell active-step="complete">
    <section class="product-entry-card product-entry-card--complete">
      <div class="complete-panel">
        <div class="complete-panel__icon" aria-hidden="true">
          <span></span>
        </div>

        <h2>商品信息已提交</h2>
        <p>我们正在为您进行智能合规检测和内容生成，预计需要2-3分钟</p>

        <div class="complete-actions">
          <button type="button" class="complete-button complete-button--analysis" @click="goCompliance">
            <span>查看合规分析</span>
            <img :src="arrowRightIcon" alt="" />
          </button>
          <button type="button" class="complete-button complete-button--ghost" @click="goWorkbench">返回工作台</button>
        </div>
      </div>
    </section>
  </ProductEntryFlowShell>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ProductEntryFlowShell from '@/components/product-entry/ProductEntryFlowShell.vue'
import arrowRightIcon from '@/assets/figma/product-entry/arrow-right.svg'
import { saveCreateDramaDraft } from '@/utils/createDramaDraft'
import { buildCreateDramaDraftFromProductEntry } from '@/utils/productEntryDraft'

const router = useRouter()

const goCompliance = () => {
  const createDraft = buildCreateDramaDraftFromProductEntry()
  if (createDraft) {
    // 完成页再次同步创建草稿，是为了用户刷新或直接点击合规分析时仍能拿到完整商品上下文。
    saveCreateDramaDraft(createDraft)
  }

  router.push('/compliance')
}

const goWorkbench = () => {
  router.push('/dramas')
}

onMounted(() => {
  const createDraft = buildCreateDramaDraftFromProductEntry()
  if (createDraft) {
    saveCreateDramaDraft(createDraft)
  }
})
</script>

<style scoped>
.product-entry-card {
  width: 960px;
  min-height: 418px;
  margin: 36px auto 0;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #ffffff;
  box-shadow:
    0 10px 15px 0 rgba(0, 0, 0, 0.1),
    0 4px 6px 0 rgba(0, 0, 0, 0.1);
}

.complete-panel {
  min-height: 352px;
  padding: 81px 33px 65px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.complete-panel__icon {
  width: 80px;
  height: 80px;
  border-radius: 999px;
  background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.complete-panel__icon span {
  width: 36px;
  height: 36px;
  border: 3px solid #ffffff;
  border-radius: 999px;
  position: relative;
}

.complete-panel__icon span::after {
  content: '';
  position: absolute;
  left: 9px;
  top: 10px;
  width: 12px;
  height: 7px;
  border-bottom: 3px solid #ffffff;
  border-left: 3px solid #ffffff;
  transform: rotate(-45deg);
}

.complete-panel h2 {
  margin: 24px 0 0;
  color: #0a2463;
  font-family: 'Urbanist', 'Noto Sans SC', 'PingFang SC', sans-serif;
  font-size: 24px;
  font-weight: 700;
  line-height: 32px;
}

.complete-panel p {
  max-width: 560px;
  margin: 16px 0 0;
  color: #45556c;
  font-size: 16px;
  line-height: 24px;
}

.complete-actions {
  margin-top: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.complete-button {
  height: 48px;
  border: none;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
  cursor: pointer;
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.complete-button img {
  width: 20px;
  height: 20px;
  display: block;
}

.complete-button:focus-visible {
  outline: none;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.09);
}

.complete-button--analysis {
  width: 172px;
  color: #ffffff;
  background: linear-gradient(90deg, #f97316 0%, #fb923c 100%);
}

.complete-button--analysis:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(249, 115, 22, 0.24);
}

.complete-button--ghost {
  width: 128px;
  color: #0a2463;
  background: #f1f5f9;
}

.complete-button--ghost:hover {
  background: #e8eef6;
}

.complete-button:active {
  transform: translateY(0);
}

@media (max-width: 1120px) {
  .product-entry-card {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .complete-panel {
    padding: 56px 20px 48px;
  }

  .complete-actions {
    width: 100%;
    flex-direction: column;
  }

  .complete-button,
  .complete-button--analysis,
  .complete-button--ghost {
    width: 100%;
  }
}
</style>
