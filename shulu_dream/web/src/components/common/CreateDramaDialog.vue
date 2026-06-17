<template>
  <!-- Create Drama Dialog / 创建短剧弹窗 -->
  <el-dialog
    v-model="visible"
    :title="$t('drama.createNew')"
    width="640px"
    align-center
    :close-on-click-modal="false"
    modal-class="create-dialog-overlay"
    class="create-dialog dialog-form-safe"
    @closed="handleClosed"
  >
    <div class="dialog-desc">{{ $t("drama.createDesc") }}</div>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      class="create-form long-form form-enter-flow"
      @submit.prevent="handleSubmit"
      @keydown.enter="handleFormEnterNavigation"
    >
      <el-form-item :label="$t('drama.projectName')" prop="title" required>
        <el-input
          v-model="form.title"
          :placeholder="$t('drama.projectNamePlaceholder')"
          size="large"
          maxlength="50"
          show-word-limit
        />
      </el-form-item>

      <el-form-item
        :label="$t('drama.projectDesc')"
        prop="description"
        required
      >
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          :placeholder="$t('drama.projectDescPlaceholder')"
          maxlength="500"
          show-word-limit
          resize="none"
        />
      </el-form-item>

      <el-form-item
        :label="$t('drama.targetCountry')"
        prop="target_country"
        required
      >
        <el-select
          v-model="form.target_country"
          size="large"
          multiple
          filterable
          :reserve-keyword="false"
          :filter-method="handleCountryFilter"
          @change="handleCountryChange"
          @visible-change="handleCountryVisibleChange"
          :placeholder="$t('drama.targetCountryPlaceholder')"
          :class="[
            'country-select',
            { 'has-value': (form.target_country?.length || 0) > 0 },
          ]"
        >
          <el-option
            v-for="country in filteredCountries"
            :key="country.code"
            :label="country.label"
            :value="country.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        :label="$t('drama.materialComposition')"
        prop="material_composition"
      >
        <el-input
          v-model="form.material_composition"
          type="textarea"
          :rows="2"
          :placeholder="$t('drama.materialCompositionPlaceholder')"
          maxlength="200"
          show-word-limit
          resize="none"
        />
      </el-form-item>

      <el-form-item
        :label="$t('drama.marketingSellingPoints')"
        prop="marketing_selling_points"
      >
        <el-input
          v-model="form.marketing_selling_points"
          type="textarea"
          :rows="2"
          :placeholder="$t('drama.marketingSellingPointsPlaceholder')"
          maxlength="200"
          show-word-limit
          resize="none"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button size="large" @click="handleClose">
          {{ $t("common.cancel") }}
        </el-button>
        <el-button
          type="primary"
          size="large"
          :loading="loading"
          @click="handleSubmit"
        >
          <el-icon v-if="!loading"><ArrowRight /></el-icon>
          {{ $t("common.next") }}
        </el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog
    v-model="complianceDialogVisible"
    title="合规校验详情"
    width="1240px"
    top="0"
    append-to-body
    :close-on-click-modal="false"
    modal-class="compliance-dialog-overlay"
    class="compliance-dialog"
  >
    <div class="compliance-meta-row">
      <div class="compliance-meta-left">
        <span>校验时间：{{ complianceCheckedAt }}</span>
      </div>
      <button
        class="pdf-export-btn"
        type="button"
        @click="handleExportCompliancePdf"
      >
        <el-icon><Download /></el-icon>
        导出PDF报告
      </button>
    </div>

    <el-alert
      v-if="isComplianceBlocked"
      type="error"
      :closable="false"
      show-icon
      title="当前评分 >= 80，禁止进入下一步，请先按整改建议完善信息。"
      class="compliance-alert"
    />
    <el-alert
      v-else-if="isOrangeRisk"
      type="warning"
      :closable="false"
      show-icon
      title="当前评分为橙色风险（60-79），可继续下一步，但建议先处理高风险项。"
      class="compliance-alert"
    />

    <div class="compliance-main-grid">
      <section class="risk-score-card">
        <h3 class="section-title">综合风险评分</h3>
        <div class="risk-ring" :style="scoreRingStyle">
          <div class="risk-ring-inner">
            <div class="risk-score-value">{{ currentCompliance.score }}</div>
            <div
              class="risk-score-level"
              :style="{ color: complianceRiskMeta.color }"
            >
              {{ complianceRiskMeta.badge }}
            </div>
          </div>
        </div>
        <p class="risk-summary-text">{{ currentCompliance.summary }}</p>
      </section>

      <section class="risk-details-card">
        <header class="risk-details-header">
          <h3 class="section-title">不合规明细</h3>
          <span class="pending-count"
            >{{ complianceIssueItems.length }} 项待处理</span
          >
        </header>
        <div
          ref="riskItemListRef"
          class="risk-item-list"
          :class="{ 'risk-item-list--scrollable': riskItemListScrollable }"
          :style="riskItemListStyle"
        >
          <article
            v-for="(item, index) in complianceIssueItems"
            :key="`${item.title}-${index}`"
            class="risk-item"
          >
            <span
              class="risk-item-dot"
              :class="`risk-item-dot--${item.level}`"
            />
            <div class="risk-item-body">
              <div class="risk-item-title-row">
                <p class="risk-item-title">{{ item.title }}</p>
                <span
                  class="risk-level-chip"
                  :class="`risk-level-chip--${item.level}`"
                >
                  {{ getRiskLevelLabel(item.level) }}
                </span>
              </div>
              <p class="risk-item-desc">{{ item.suggestion }}</p>
            </div>
          </article>
        </div>
      </section>
    </div>

    <section class="rectification-card">
      <h3 class="section-title">整改建议</h3>
      <ul class="rectification-list">
        <li
          v-for="(item, index) in rectificationList"
          :key="`${item}-${index}`"
        >
          {{ item }}
        </li>
      </ul>
      <div v-if="complianceCategories.length" class="category-row">
        <span class="category-label">建议类目：</span>
        <div class="category-tags">
          <span
            v-for="(item, index) in complianceCategories"
            :key="`${item}-${index}`"
            class="category-tag"
          >
            {{ item }}
          </span>
        </div>
      </div>
    </section>

    <template #footer>
      <div class="compliance-footer">
        <el-button
          size="large"
          class="footer-secondary-btn"
          :disabled="complianceSubmitting"
          @click="handleComplianceCancel"
        >
          {{ $t("common.cancel") }}
        </el-button>
        <el-button
          type="primary"
          size="large"
          class="footer-primary-btn"
          :loading="complianceSubmitting"
          @click="handleCompliancePrimaryAction"
        >
          {{ complianceCanProceed ? $t("common.next") : "去修改" }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";
import { ArrowRight, Download } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import { dramaAPI } from "@/api/drama";
import { ALL_COUNTRIES } from "@/constants/countries";
import type {
  ComplianceResult,
  ComplianceRiskLevel,
  CreateDramaRequest,
} from "@/types/drama";
import {
  buildCreateDramaPayload,
  getComplianceRiskMeta,
  normalizeComplianceResult,
} from "@/utils/compliance";
import { saveCreateDramaDraft } from "@/utils/createDramaDraft";
import { handleFormEnterNavigation } from "@/utils/formFocus";

interface ComplianceIssueItem {
  level: ComplianceRiskLevel;
  title: string;
  suggestion: string;
}

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  created: [id: string];
}>();

const { t } = useI18n();
const router = useRouter();
const formRef = ref<FormInstance>();
const loading = ref(false);
const complianceSubmitting = ref(false);
const countryKeyword = ref("");
const complianceDialogVisible = ref(false);
const complianceData = ref<ComplianceResult | null>(null);
const complianceDramaId = ref("");
const complianceCheckedAt = ref("");
const riskItemListRef = ref<HTMLDivElement>();
const riskItemListMaxHeight = ref("");
const pendingCreatePayload = ref<CreateDramaRequest | null>(null);
const pendingComplianceToken = ref("");

const visible = ref(props.modelValue);
watch(
  () => props.modelValue,
  (val) => {
    visible.value = val;
  },
);
watch(visible, (val) => {
  emit("update:modelValue", val);
});

watch(
  () => complianceDialogVisible.value,
  (open) => {
    if (open) {
      void updateRiskItemListHeight();
    }
  },
);

const form = reactive<CreateDramaRequest>({
  title: "",
  description: "",
  target_country: [],
  material_composition: "",
  marketing_selling_points: "",
});

const filteredCountries = computed(() => {
  const keyword = countryKeyword.value.trim().toLowerCase();
  if (!keyword) {
    return ALL_COUNTRIES;
  }
  return ALL_COUNTRIES.filter((country) =>
    country.searchText.includes(keyword),
  );
});

const currentCompliance = computed<ComplianceResult>(() => {
  if (complianceData.value) {
    return complianceData.value;
  }
  return {
    score: 0,
    level: "green",
    level_label: "低",
    summary: "暂无合规评估结果",
    non_compliance_points: [],
    rectification_suggestions: [],
    suggested_categories: [],
  };
});

const complianceRiskMeta = computed(() =>
  getComplianceRiskMeta(currentCompliance.value),
);

const isComplianceBlocked = computed(() => {
  return (
    currentCompliance.value.level === "red" ||
    currentCompliance.value.score >= 80
  );
});

const isOrangeRisk = computed(() => currentCompliance.value.level === "orange");

const complianceCanProceed = computed(() => {
  return (
    !isComplianceBlocked.value &&
    !!pendingCreatePayload.value &&
    !!pendingComplianceToken.value
  );
});

const scoreRingStyle = computed(() => {
  const score = Math.max(0, Math.min(currentCompliance.value.score, 100));
  return {
    "--risk-angle": `${(score / 100) * 360}deg`,
    "--risk-color": complianceRiskMeta.value.color,
  };
});

const rectificationList = computed(() => {
  const list = currentCompliance.value.rectification_suggestions || [];
  if (list.length > 0) {
    return list;
  }
  return ["请补充完整商品信息并重新进行合规校验。"];
});

const complianceCategories = computed(
  () => currentCompliance.value.suggested_categories || [],
);

const normalizeText = (value: string) =>
  value.toLowerCase().replace(/\s+/g, "");

const inferIssueLevel = (text: string): ComplianceRiskLevel => {
  const raw = normalizeText(text);
  if (/(禁售|违法|武器|毒|侵权|走私|伪造|医疗|药品|处方|高危|禁止)/.test(raw)) {
    return "red";
  }
  if (
    /(缺少|未提供|认证|隐私|数据|gdpr|appi|ukca|ce|pse|违规|不符合|不合规|高风险)/.test(
      raw,
    )
  ) {
    return "orange";
  }
  if (/(敏感|绝对化|夸大|中风险|建议|优化|提示)/.test(raw)) {
    return "yellow";
  }
  return currentCompliance.value.level;
};

const ISSUE_LEVEL_PRIORITY: Record<ComplianceRiskLevel, number> = {
  red: 0,
  orange: 1,
  yellow: 2,
  green: 3,
};

const complianceIssueItems = computed<ComplianceIssueItem[]>(() => {
  const points = currentCompliance.value.non_compliance_points || [];
  const suggestions = currentCompliance.value.rectification_suggestions || [];

  if (points.length === 0) {
    return [
      {
        level: currentCompliance.value.level,
        title:
          currentCompliance.value.summary || "暂无明确不合规项，请人工复核。",
        suggestion: suggestions[0] || "请结合目标国家法规继续完善商品信息。",
      },
    ];
  }

  return points
    .map((title, index) => ({
      order: index,
      level: inferIssueLevel(title),
      title,
      suggestion:
        suggestions[index] ||
        suggestions[0] ||
        "请补充相关资质文件并重新校验。",
    }))
    .sort((a, b) => {
      const priorityDiff =
        ISSUE_LEVEL_PRIORITY[a.level] - ISSUE_LEVEL_PRIORITY[b.level];
      if (priorityDiff !== 0) return priorityDiff;
      return a.order - b.order;
    })
    .map(({ order, ...item }) => item);
});

const riskItemListScrollable = computed(
  () => complianceIssueItems.value.length > 3,
);

const riskItemListStyle = computed(() => {
  if (!riskItemListScrollable.value || !riskItemListMaxHeight.value) {
    return undefined;
  }

  return {
    maxHeight: riskItemListMaxHeight.value,
  };
});

const getRiskItemListMaxAllowedHeight = () => {
  if (typeof window === "undefined") {
    return 360;
  }

  return Math.max(210, Math.round(window.innerHeight * 0.34));
};

const updateRiskItemListHeight = async () => {
  await nextTick();

  const listEl = riskItemListRef.value;
  if (!listEl || !riskItemListScrollable.value) {
    riskItemListMaxHeight.value = "";
    return;
  }

  const items = Array.from(listEl.querySelectorAll<HTMLElement>(".risk-item"));
  if (items.length === 0) {
    riskItemListMaxHeight.value = "";
    return;
  }

  const twoItemHeight = items
    .slice(0, 2)
    .reduce((total, item) => total + item.offsetHeight, 0);
  const halfThirdItemHeight = items[2]
    ? Math.round(items[2].offsetHeight * 0.5)
    : 0;

  const maxAllowedHeight = getRiskItemListMaxAllowedHeight();
  const compressedHeight = Math.max(
    196,
    twoItemHeight + halfThirdItemHeight - 12,
  );
  riskItemListMaxHeight.value = `${Math.min(compressedHeight, maxAllowedHeight)}px`;
};

watch(
  () =>
    complianceIssueItems.value
      .map((item) => `${item.level}:${item.title}:${item.suggestion}`)
      .join("|"),
  () => {
    if (complianceDialogVisible.value) {
      void updateRiskItemListHeight();
    }
  },
);

const handleComplianceResize = () => {
  if (complianceDialogVisible.value) {
    void updateRiskItemListHeight();
  }
};

if (typeof window !== "undefined") {
  window.addEventListener("resize", handleComplianceResize);
}

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener("resize", handleComplianceResize);
  }
});

const getRiskLevelLabel = (level: ComplianceRiskLevel) => {
  if (level === "red") return "禁止";
  if (level === "orange") return "高风险";
  if (level === "yellow") return "中风险";
  return "低风险";
};

const handleCountryFilter = (keyword: string) => {
  countryKeyword.value = keyword;
};

const handleCountryVisibleChange = (open: boolean) => {
  if (!open) {
    countryKeyword.value = "";
  }
};

const handleCountryChange = () => {
  countryKeyword.value = "";
};

const rules: FormRules = {
  title: [
    { required: true, message: "请输入项目标题", trigger: "blur" },
    { min: 1, max: 50, message: "标题长度在 1 到 50 个字符", trigger: "blur" },
  ],
  description: [
    { required: true, message: "请输入项目描述", trigger: "blur" },
    {
      min: 1,
      max: 500,
      message: "描述长度在 1 到 500 个字符",
      trigger: "blur",
    },
  ],
  target_country: [
    {
      type: "array",
      required: true,
      min: 1,
      message: "请选择目标国家",
      trigger: "change",
    },
  ],
  material_composition: [
    { max: 200, message: "材质/成分长度不能超过 200 个字符", trigger: "blur" },
  ],
  marketing_selling_points: [
    { max: 200, message: "宣传卖点长度不能超过 200 个字符", trigger: "blur" },
  ],
};

const splitWrappedLines = (
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
): string[] => {
  const blocks = String(text ?? "")
    .replace(/\r\n/g, "\n")
    .split("\n");
  const lines: string[] = [];

  for (const block of blocks) {
    if (!block) {
      lines.push("");
      continue;
    }

    let line = "";
    for (const char of block) {
      const candidate = line + char;
      if (!line || ctx.measureText(candidate).width <= maxWidth) {
        line = candidate;
      } else {
        lines.push(line);
        line = char;
      }
    }
    if (line) {
      lines.push(line);
    }
  }

  return lines.length ? lines : [""];
};

const textToUint8 = (value: string): Uint8Array => {
  return new TextEncoder().encode(value);
};

const concatUint8Arrays = (parts: Uint8Array[]): Uint8Array => {
  const totalLength = parts.reduce((sum, item) => sum + item.length, 0);
  const merged = new Uint8Array(totalLength);
  let offset = 0;
  for (const item of parts) {
    merged.set(item, offset);
    offset += item.length;
  }
  return merged;
};

const buildPdfBlobFromJpegDataUrl = (
  jpegDataUrl: string,
  imageWidth: number,
  imageHeight: number,
): Blob => {
  const base64 = jpegDataUrl.split(",")[1] || "";
  const binary = atob(base64);
  const jpegBytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    jpegBytes[i] = binary.charCodeAt(i);
  }

  const pageWidth = 595.28;
  const pageHeight = Number(
    ((pageWidth * imageHeight) / imageWidth).toFixed(2),
  );
  const contentStream = `q\n${pageWidth.toFixed(2)} 0 0 ${pageHeight.toFixed(2)} 0 0 cm\n/Im0 Do\nQ\n`;

  const objects: Uint8Array[] = [];
  objects.push(
    textToUint8("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n"),
  );
  objects.push(
    textToUint8("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n"),
  );
  objects.push(
    textToUint8(
      `3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${pageWidth.toFixed(2)} ${pageHeight.toFixed(2)}] /Resources << /XObject << /Im0 4 0 R >> >> /Contents 5 0 R >>\nendobj\n`,
    ),
  );
  objects.push(
    concatUint8Arrays([
      textToUint8(
        `4 0 obj\n<< /Type /XObject /Subtype /Image /Width ${Math.round(imageWidth)} /Height ${Math.round(imageHeight)} /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length ${jpegBytes.length} >>\nstream\n`,
      ),
      jpegBytes,
      textToUint8("\nendstream\nendobj\n"),
    ]),
  );
  objects.push(
    textToUint8(
      `5 0 obj\n<< /Length ${textToUint8(contentStream).length} >>\nstream\n${contentStream}endstream\nendobj\n`,
    ),
  );

  const header = textToUint8("%PDF-1.4\n%\xFF\xFF\xFF\xFF\n");
  const offsets: number[] = [0];
  let currentOffset = header.length;
  for (const objectBytes of objects) {
    offsets.push(currentOffset);
    currentOffset += objectBytes.length;
  }

  let xref = "xref\n0 6\n0000000000 65535 f \n";
  for (let index = 1; index <= 5; index += 1) {
    xref += `${String(offsets[index]).padStart(10, "0")} 00000 n \n`;
  }

  const xrefBytes = textToUint8(xref);
  const trailer = textToUint8(
    `trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n${currentOffset}\n%%EOF`,
  );

  const pdfBytes = concatUint8Arrays([header, ...objects, xrefBytes, trailer]);
  const blobBytes = new Uint8Array(pdfBytes.byteLength);
  blobBytes.set(pdfBytes);
  return new Blob([blobBytes], { type: "application/pdf" });
};

const downloadBlob = (blob: Blob, filename: string) => {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.style.display = "none";
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(url);
};

const formatDateTime = (date: Date) => {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  const hour = `${date.getHours()}`.padStart(2, "0");
  const minute = `${date.getMinutes()}`.padStart(2, "0");
  const second = `${date.getSeconds()}`.padStart(2, "0");
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
};

const openComplianceDialog = (rawCompliance: unknown) => {
  const normalized = normalizeComplianceResult(rawCompliance);
  if (!normalized) {
    return false;
  }

  complianceData.value = normalized;
  complianceDramaId.value = "";
  complianceCheckedAt.value = formatDateTime(new Date());
  complianceDialogVisible.value = true;
  return true;
};

const handleComplianceCancel = () => {
  complianceDialogVisible.value = false;
};

const handleComplianceEdit = async () => {
  const draft = pendingCreatePayload.value || buildCreateDramaPayload(form);
  saveCreateDramaDraft(draft);
  complianceDialogVisible.value = false;
  visible.value = false;
  await router.push("/dramas/create");
  ElMessage.info("已保留当前内容，请在“创建新项目”页面修改后再提交。");
};

const handleCompliancePrimaryAction = async () => {
  if (complianceCanProceed.value) {
    await handleComplianceConfirm();
    return;
  }
  await handleComplianceEdit();
};

const handleComplianceConfirm = async () => {
  if (!complianceCanProceed.value || !pendingCreatePayload.value) {
    ElMessage.warning("风险评级为红色（>=80），请先根据整改建议完善后再提交。");
    return;
  }

  complianceSubmitting.value = true;
  try {
    const result = await dramaAPI.create({
      ...pendingCreatePayload.value,
      compliance_token: pendingComplianceToken.value,
    });
    const dramaId = String(result.drama.id);
    complianceDramaId.value = dramaId;
    complianceDialogVisible.value = false;
    visible.value = false;
    pendingCreatePayload.value = null;
    pendingComplianceToken.value = "";
    emit("created", dramaId);

    if (isOrangeRisk.value) {
      ElMessage.warning(
        "当前项目为橙色高风险，已进入下一步，请优先处理不合规项。",
      );
    } else {
      ElMessage.success("创建成功");
    }
    router.push(`/dramas/${dramaId}`);
  } catch (error: any) {
    if (error?.code === "COMPLIANCE_PRECHECK_REQUIRED") {
      pendingComplianceToken.value = "";
      complianceDialogVisible.value = false;
      ElMessage.warning(
        error.message || "合规预检已失效，请重新校验后再继续。",
      );
      return;
    }
    const opened = openComplianceDialog(error?.details?.compliance);
    if (opened) {
      ElMessage.warning(
        error.message || "风险评级发生变化，请根据最新整改建议处理后再继续。",
      );
      return;
    }
    ElMessage.error(error.message || "创建失败");
  } finally {
    complianceSubmitting.value = false;
  }
};

const handleExportCompliancePdf = () => {
  const report = currentCompliance.value;
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");
  if (!ctx) {
    ElMessage.error("导出失败，请重试");
    return;
  }

  const fontFamily =
    '"PingFang SC", "Microsoft YaHei", "Segoe UI", Arial, sans-serif';
  const canvasWidth = 1240;
  const marginX = 72;
  const contentWidth = canvasWidth - marginX * 2;
  const lineHeight = 34;

  const titleFont = `700 48px ${fontFamily}`;
  const headingFont = `700 30px ${fontFamily}`;
  const bodyFont = `500 24px ${fontFamily}`;
  const metaFont = `500 22px ${fontFamily}`;

  ctx.font = bodyFont;
  const summaryLines = splitWrappedLines(
    ctx,
    report.summary || "无",
    contentWidth,
  );
  const issueGroups = (
    complianceIssueItems.value.length
      ? complianceIssueItems.value
      : [
          {
            level: report.level,
            title: "暂无明显不合规项",
            suggestion: "建议继续人工复核目标国家法规要求。",
          },
        ]
  ).map((item) => {
    ctx.font = headingFont;
    const titleLines = splitWrappedLines(ctx, item.title, contentWidth - 52);
    ctx.font = bodyFont;
    const suggestionLines = splitWrappedLines(
      ctx,
      item.suggestion,
      contentWidth - 52,
    );
    return { ...item, titleLines, suggestionLines };
  });

  ctx.font = bodyFont;
  const rectificationGroups = rectificationList.value.map((item) =>
    splitWrappedLines(ctx, item, contentWidth - 32),
  );
  const categoryText = complianceCategories.value.length
    ? complianceCategories.value.join("、")
    : "无";
  const categoryLines = splitWrappedLines(
    ctx,
    categoryText,
    contentWidth - 120,
  );

  let totalHeight = 72;
  totalHeight += 76;
  totalHeight += 44 * 3;
  totalHeight += 20;
  totalHeight += 46 + summaryLines.length * lineHeight + 14;
  totalHeight += 46;
  totalHeight += issueGroups.reduce(
    (sum, item) =>
      sum +
      item.titleLines.length * lineHeight +
      item.suggestionLines.length * lineHeight +
      26,
    0,
  );
  totalHeight += 24;
  totalHeight += 46;
  totalHeight += rectificationGroups.reduce(
    (sum, lines) => sum + lines.length * lineHeight + 12,
    0,
  );
  totalHeight += Math.max(1, categoryLines.length) * lineHeight + 46;
  totalHeight += 72;

  canvas.width = canvasWidth;
  canvas.height = Math.max(1180, Math.ceil(totalHeight));
  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  let y = 84;
  ctx.fillStyle = "#1f2937";
  ctx.font = titleFont;
  ctx.fillText("合规校验报告", marginX, y);

  y += 62;
  ctx.fillStyle = "#5b6b80";
  ctx.font = metaFont;
  ctx.fillText(`校验时间：${complianceCheckedAt.value}`, marginX, y);
  y += 36;
  ctx.fillText(`风险评分：${report.score}`, marginX, y);
  y += 36;
  ctx.fillText(`风险等级：${getRiskLevelLabel(report.level)}`, marginX, y);

  y += 30;
  ctx.fillStyle = "#111827";
  ctx.font = headingFont;
  ctx.fillText("评估结论", marginX, y);
  y += 42;
  ctx.fillStyle = "#374151";
  ctx.font = bodyFont;
  for (const line of summaryLines) {
    ctx.fillText(line, marginX, y);
    y += lineHeight;
  }

  y += 10;
  ctx.fillStyle = "#111827";
  ctx.font = headingFont;
  ctx.fillText("不合规明细", marginX, y);
  y += 42;

  for (const item of issueGroups) {
    ctx.fillStyle = "#ef4444";
    ctx.beginPath();
    ctx.arc(marginX + 10, y - 8, 6, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = "#111827";
    ctx.font = headingFont;
    for (const line of item.titleLines) {
      ctx.fillText(line, marginX + 30, y);
      y += lineHeight;
    }

    ctx.fillStyle = "#4b5563";
    ctx.font = bodyFont;
    for (const line of item.suggestionLines) {
      ctx.fillText(line, marginX + 30, y);
      y += lineHeight;
    }
    y += 10;
  }

  y += 10;
  ctx.fillStyle = "#111827";
  ctx.font = headingFont;
  ctx.fillText("整改建议", marginX, y);
  y += 42;

  ctx.fillStyle = "#374151";
  ctx.font = bodyFont;
  for (const lines of rectificationGroups) {
    ctx.fillText("•", marginX, y);
    for (const line of lines) {
      ctx.fillText(line, marginX + 22, y);
      y += lineHeight;
    }
    y += 8;
  }

  y += 10;
  ctx.fillStyle = "#111827";
  ctx.font = headingFont;
  ctx.fillText("建议类目", marginX, y);
  y += 42;

  ctx.fillStyle = "#1d4ed8";
  ctx.font = bodyFont;
  for (const line of categoryLines) {
    ctx.fillText(line, marginX, y);
    y += lineHeight;
  }

  const jpegDataUrl = canvas.toDataURL("image/jpeg", 0.92);
  const pdfBlob = buildPdfBlobFromJpegDataUrl(
    jpegDataUrl,
    canvas.width,
    canvas.height,
  );
  const fallbackTime = formatDateTime(new Date()).replace(/[: ]/g, "-");
  const fileTime = (complianceCheckedAt.value || fallbackTime).replace(
    /[: ]/g,
    "-",
  );
  downloadBlob(pdfBlob, `合规校验报告_${fileTime}.pdf`);
  ElMessage.success("PDF报告已下载");
};

const resetForm = () => {
  form.title = "";
  form.description = "";
  form.target_country = [];
  form.material_composition = "";
  form.marketing_selling_points = "";
  pendingCreatePayload.value = null;
  pendingComplianceToken.value = "";
  complianceData.value = null;
  complianceDramaId.value = "";
  complianceCheckedAt.value = "";
  riskItemListMaxHeight.value = "";
  complianceSubmitting.value = false;
};

const handleClosed = () => {
  resetForm();
  countryKeyword.value = "";
  formRef.value?.clearValidate();
};

const handleClose = () => {
  visible.value = false;
};

const handleSubmit = async () => {
  if (!formRef.value) return;

  const valid = await formRef.value
    .validate()
    .then(() => true)
    .catch(() => false);
  if (!valid) return;

  loading.value = true;
  try {
    const payload = buildCreateDramaPayload(form);
    pendingCreatePayload.value = payload;
    const result = await dramaAPI.checkCompliance(payload);
    pendingComplianceToken.value = result.compliance_token || "";
    const opened = openComplianceDialog(result.compliance);

    if (!opened) {
      pendingCreatePayload.value = null;
      pendingComplianceToken.value = "";
      ElMessage.error("未获取到合规校验结果，请重试");
      return;
    }

    if (!isComplianceBlocked.value && !pendingComplianceToken.value) {
      pendingCreatePayload.value = null;
      complianceDialogVisible.value = false;
      ElMessage.error("未获取到合规预检凭证，请重试");
      return;
    }

    if (isComplianceBlocked.value) {
      ElMessage.warning(
        "风险评级为红色（>=80），禁止进入下一步，请先整改后重试。",
      );
      return;
    }

    if (isOrangeRisk.value) {
      ElMessage.warning(
        "当前风险为橙色，可继续下一步，但建议先优先处理高风险项。",
      );
    }
  } catch (error: any) {
    pendingCreatePayload.value = null;
    pendingComplianceToken.value = "";
    const opened = openComplianceDialog(error?.details?.compliance);
    if (opened) {
      ElMessage.warning(
        error.message || "风险评级为红色（>=80），请先整改后再提交。",
      );
      return;
    }
    ElMessage.error(error.message || "创建失败");
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
:global(.create-dialog-overlay) {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px 16px;
  overflow: hidden;
  overscroll-behavior: contain;
}

:global(.compliance-dialog-overlay) {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  overscroll-behavior: contain;
}

:global(.create-dialog-overlay .el-overlay-dialog) {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

:global(.compliance-dialog-overlay .el-overlay-dialog) {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  padding: 30px 8px;
}

:deep(.el-dialog.create-dialog) {
  width: min(640px, calc(100vw - 32px)) !important;
  max-width: calc(100vw - 32px);
  max-height: calc(100vh - 32px);
  margin: 0 !important;
  display: flex;
  flex-direction: column;
  border-radius: var(--radius-xl);
  overflow: hidden;
}

.create-dialog :deep(.el-dialog__header) {
  flex-shrink: 0;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-primary);
  margin-right: 0;
}

.create-dialog :deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.create-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 1.5rem;
}

.create-dialog :deep(.el-dialog__footer) {
  flex-shrink: 0;
  padding: 1rem 1.5rem 1.5rem;
  border-top: 1px solid var(--border-primary);
  background: var(--bg-card);
}

.dialog-desc {
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.create-form :deep(.el-form-item) {
  margin-bottom: 1.25rem;
}

.create-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.create-form :deep(.el-input__wrapper),
.create-form :deep(.el-textarea__inner),
.create-form :deep(.el-select__wrapper) {
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
  transition: all var(--transition-fast);
}

.create-form :deep(.el-input__wrapper:hover),
.create-form :deep(.el-textarea__inner:hover),
.create-form :deep(.el-select__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--border-secondary) inset;
}

.create-form :deep(.el-input__wrapper.is-focus),
.create-form :deep(.el-textarea__inner:focus),
.create-form :deep(.el-select__wrapper.is-focused) {
  box-shadow: 0 0 0 2px var(--accent) inset;
}

.create-form :deep(.el-input__inner),
.create-form :deep(.el-textarea__inner),
.create-form :deep(.el-select__selected-item) {
  color: var(--text-primary);
}

.country-select :deep(.el-select__placeholder) {
  color: #a8b5c6;
}

.country-select.has-value :deep(.el-select__placeholder) {
  color: var(--text-primary);
}

.create-form :deep(.el-input__inner::placeholder),
.create-form :deep(.el-textarea__inner::placeholder) {
  color: var(--text-muted);
}

.create-form :deep(.el-input__count) {
  color: var(--text-muted);
  background: transparent;
}

.create-form :deep(.el-form-item.is-error .el-input__wrapper),
.create-form :deep(.el-form-item.is-error .el-textarea__inner),
.create-form :deep(.el-form-item.is-error .el-select__wrapper) {
  box-shadow: 0 0 0 1px var(--el-color-danger) inset !important;
}

.create-form :deep(.el-form-item.is-error .el-input__wrapper.is-focus),
.create-form :deep(.el-form-item.is-error .el-select__wrapper.is-focused),
.create-form :deep(.el-form-item.is-error .el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px var(--el-color-danger) inset !important;
}

.country-select {
  width: 100%;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.dialog-footer .el-button {
  min-width: 100px;
}

:deep(.el-dialog.compliance-dialog) {
  width: min(1650px, calc(100vw - 16px)) !important;
  max-width: calc(100vw - 16px);
  height: auto;
  max-height: calc(100vh - 60px);
  margin: 0 !important;
  display: flex;
  flex-direction: column;
  border-radius: 22px;
  overflow: hidden;
}

.compliance-dialog :deep(.el-dialog__header) {
  flex-shrink: 0;
  padding: 14px 20px;
  border-bottom: 1px solid #e6ebf2;
  margin-right: 0;
}

.compliance-dialog :deep(.el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
  color: #17243a;
}

.compliance-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 14px 20px;
  background: #f7f9fc;
}

.compliance-dialog :deep(.el-dialog__footer) {
  flex-shrink: 0;
  padding: 16px 20px;
  border-top: 1px solid #e6ebf2;
  background: #fff;
}

.compliance-meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  color: #6f7f95;
  font-size: 14px;
}

.compliance-meta-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pdf-export-btn {
  border: none;
  background: transparent;
  color: #0ea5e9;
  font-size: 14px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.compliance-alert {
  margin-bottom: 10px;
}

.compliance-main-grid {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 12px;
}

.risk-score-card,
.risk-details-card,
.rectification-card {
  background: #ffffff;
  border: 1px solid #dce4ee;
  border-radius: 16px;
}

.risk-score-card {
  padding: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.section-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: #1d293f;
}

.risk-ring {
  width: 136px;
  height: 136px;
  margin-top: 12px;
  border-radius: 50%;
  background: conic-gradient(
    var(--risk-color) var(--risk-angle),
    #edf2f7 var(--risk-angle)
  );
  display: grid;
  place-items: center;
}

.risk-ring-inner {
  width: 102px;
  height: 102px;
  border-radius: 50%;
  background: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.risk-score-value {
  font-size: 34px;
  line-height: 1;
  font-weight: 700;
  color: #1d293f;
}

.risk-score-level {
  margin-top: 8px;
  font-size: 14px;
  font-weight: 700;
}

.risk-summary-text {
  margin: 10px 0 0;
  color: #667891;
  font-size: 14px;
  line-height: 1.55;
  text-align: center;
}

.risk-details-card {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.risk-details-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid #e8edf5;
}

.pending-count {
  color: #ef4444;
  background: #fff1f2;
  border-radius: 12px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
}

.risk-item-list {
  overflow: visible;
}

.risk-item-list--scrollable {
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-gutter: stable;
  overscroll-behavior: contain;
}

.risk-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid #edf2f7;
}

.risk-item:last-child {
  border-bottom: none;
}

.risk-item-dot {
  margin-top: 6px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex: 0 0 auto;
}

.risk-item-dot--red {
  background: #ef4444;
}

.risk-item-dot--orange {
  background: #f97316;
}

.risk-item-dot--yellow {
  background: #f59e0b;
}

.risk-item-dot--green {
  background: #22c55e;
}

.risk-item-body {
  flex: 1;
  min-width: 0;
}

.risk-item-title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.risk-item-title {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #1d293f;
  line-height: 1.45;
}

.risk-item-desc {
  margin: 6px 0 0;
  color: #6b7c92;
  font-size: 13px;
  line-height: 1.55;
}

.risk-level-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.risk-level-chip--red {
  color: #ef4444;
  background: #fff1f2;
}

.risk-level-chip--orange {
  color: #f97316;
  background: #fff7ed;
}

.risk-level-chip--yellow {
  color: #d97706;
  background: #fffbeb;
}

.risk-level-chip--green {
  color: #16a34a;
  background: #ecfdf3;
}

.rectification-card {
  margin-top: 12px;
  padding: 14px 16px;
}

.rectification-list {
  margin: 10px 0 0;
  padding-left: 20px;
  color: #2f3f55;
  font-size: 12px;
  line-height: 1.5;
  columns: 2;
  column-gap: 24px;
}

.rectification-list li {
  break-inside: avoid;
  margin-bottom: 6px;
}

.category-row {
  margin-top: 10px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.category-label {
  font-size: 13px;
  color: #5f7088;
  flex: 0 0 auto;
}

.category-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.category-tag {
  padding: 5px 10px;
  border-radius: 999px;
  background: #edf6ff;
  color: #0284c7;
  font-size: 12px;
}

.compliance-footer {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.compliance-footer .el-button {
  min-width: 116px;
  min-height: 44px;
  border-radius: 12px;
}

.footer-secondary-btn {
  border-color: #d3dce8;
  color: #4b5565;
  background: #fff;
}

.footer-secondary-btn:hover {
  border-color: #b6c4d7;
  color: #344153;
  background: #fff;
}

.footer-primary-btn {
  border: none;
  background: linear-gradient(135deg, var(--accent) 0%, #0284c7 100%);
  box-shadow: 0 8px 18px rgba(14, 165, 233, 0.28);
}

.footer-primary-btn:hover {
  background: linear-gradient(135deg, #38bdf8 0%, #0ea5e9 100%);
  box-shadow: 0 10px 22px rgba(14, 165, 233, 0.34);
  color: #fff;
}

@media (max-width: 1200px) {
  :deep(.el-dialog.compliance-dialog) {
    width: min(98vw, 1650px) !important;
  }

  .compliance-main-grid {
    grid-template-columns: 1fr;
  }

  .rectification-list {
    columns: 1;
  }
}

@media (max-height: 860px) and (min-width: 769px) {
  :global(.compliance-dialog-overlay .el-overlay-dialog) {
    padding: 20px 6px;
  }

  :deep(.el-dialog.compliance-dialog) {
    width: min(1600px, calc(100vw - 8px)) !important;
    max-width: calc(100vw - 8px);
    max-height: calc(100vh - 40px);
  }

  .compliance-dialog :deep(.el-dialog__header) {
    padding: 11px 16px;
  }

  .compliance-dialog :deep(.el-dialog__body) {
    padding: 11px 16px;
  }

  .compliance-dialog :deep(.el-dialog__footer) {
    padding: 12px 16px;
  }

  .section-title {
    font-size: 17px;
  }

  .compliance-main-grid {
    grid-template-columns: 280px minmax(0, 1fr);
    gap: 10px;
  }

  .risk-score-card {
    padding: 12px;
  }

  .risk-ring {
    width: 122px;
    height: 122px;
    margin-top: 8px;
  }

  .risk-ring-inner {
    width: 92px;
    height: 92px;
  }

  .risk-score-value {
    font-size: 30px;
  }

  .risk-summary-text {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.4;
  }

  .risk-item {
    padding: 9px 12px;
  }

  .risk-item-title {
    font-size: 13px;
  }

  .risk-item-desc,
  .rectification-list,
  .category-label {
    font-size: 12px;
    line-height: 1.4;
  }

  .rectification-card {
    margin-top: 6px;
    padding: 11px 12px;
  }
}

@media (max-width: 768px) {
  :global(.create-dialog-overlay) {
    padding: 20px 10px;
  }

  :global(.compliance-dialog-overlay .el-overlay-dialog) {
    padding: 20px 10px;
  }

  :deep(.el-dialog.create-dialog),
  :deep(.el-dialog.compliance-dialog) {
    max-width: calc(100vw - 20px);
    max-height: calc(100vh - 20px);
  }

  :deep(.el-dialog.compliance-dialog) {
    max-height: calc(100vh - 40px);
  }

  .create-dialog :deep(.el-dialog__body) {
    padding: 1rem;
  }

  .dialog-footer {
    flex-direction: column-reverse;
    align-items: stretch;
  }

  .dialog-footer .el-button {
    width: 100%;
    min-width: 0;
  }

  .compliance-dialog :deep(.el-dialog__header),
  .compliance-dialog :deep(.el-dialog__body),
  .compliance-dialog :deep(.el-dialog__footer) {
    padding: 14px;
  }

  .compliance-meta-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .compliance-footer {
    flex-direction: column-reverse;
    align-items: stretch;
  }

  .compliance-footer .el-button {
    width: 100%;
  }
}
</style>
