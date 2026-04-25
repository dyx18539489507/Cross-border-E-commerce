<template>
  <!-- Create Drama Dialog / 创建短剧弹窗 -->
  <el-dialog
    v-model="visible"
    :title="$t('drama.createNew')"
    width="640px"
    align-center
    append-to-body
    :close-on-click-modal="false"
    modal-class="create-dialog-overlay"
    class="create-dialog dialog-form-safe"
    @closed="handleClosed"
  >
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
    :title="t('compliance.dialogTitle')"
    width="1240px"
    top="0"
    append-to-body
    :close-on-click-modal="false"
    modal-class="compliance-dialog-overlay"
    class="compliance-dialog"
  >
    <div class="compliance-meta-row">
      <div class="compliance-meta-left">
        <span>{{ t('compliance.checkedAt') }}: {{ complianceCheckedAt }}</span>
      </div>
      <button
        class="pdf-export-btn"
        type="button"
        @click="handleExportCompliancePdf"
      >
        <el-icon><Download /></el-icon>
        {{ t('compliance.exportPdf') }}
      </button>
    </div>

    <el-alert
      v-if="isComplianceBlocked"
      type="error"
      :closable="false"
      show-icon
      :title="t('compliance.blockedTitle')"
      class="compliance-alert"
    />
    <el-alert
      v-else-if="isOrangeRisk"
      type="warning"
      :closable="false"
      show-icon
      :title="t('compliance.orangeTitle')"
      class="compliance-alert"
    />

    <div class="compliance-main-grid">
      <section class="risk-score-card">
        <h3 class="section-title">{{ t('compliance.score') }}</h3>
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
          <h3 class="section-title">{{ t('compliance.details') }}</h3>
          <span class="pending-count">{{ t('compliance.itemsPending', { count: complianceIssueItems.length }) }}</span>
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
      <h3 class="section-title">{{ t('compliance.rectification') }}</h3>
      <ul class="rectification-list">
        <li
          v-for="(item, index) in rectificationList"
          :key="`${item}-${index}`"
        >
          {{ item }}
        </li>
      </ul>
      <div v-if="complianceCategories.length" class="category-row">
        <span class="category-label">{{ t('compliance.suggestedCategories') }}:</span>
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
          {{ complianceCanProceed ? $t("common.next") : t("compliance.editAction") }}
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

const { t, locale } = useI18n();
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
    level_label: t("compliance.riskLevel.green"),
    summary: t("compliance.noResult"),
    non_compliance_points: [],
    rectification_suggestions: [],
    suggested_categories: [],
  };
});

const complianceRiskMeta = computed(() =>
  getComplianceRiskMeta(currentCompliance.value, locale.value),
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
  return [t("compliance.completeProductInfo")];
});

const complianceCategories = computed(
  () => currentCompliance.value.suggested_categories || [],
);
const complianceCategorySeparator = computed(() =>
  locale.value.startsWith("zh") ? "、" : ", "
);

const normalizeText = (value: string) =>
  value.toLowerCase().replace(/\s+/g, "");

const inferIssueLevel = (text: string): ComplianceRiskLevel => {
  const raw = normalizeText(text);
  if (/(禁售|违法|武器|毒|侵权|走私|伪造|医疗|药品|处方|高危|禁止|banned|illegal|weapon|toxic|infring|smuggl|counterfeit|medical|drug|prescription|highrisk|blocked|prohibited)/.test(raw)) {
    return "red";
  }
  if (
    /(缺少|未提供|认证|隐私|数据|gdpr|appi|ukca|ce|pse|违规|不符合|不合规|高风险|missing|notprovided|certif|privacy|data|violation|noncompliant|highrisk)/.test(
      raw,
    )
  ) {
    return "orange";
  }
  if (/(敏感|绝对化|夸大|中风险|建议|优化|提示|sensitive|absolute|exaggerat|mediumrisk|suggest|optimi|notice|warning)/.test(raw)) {
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
        title: currentCompliance.value.summary || t("compliance.noIssue"),
        suggestion: suggestions[0] || t("compliance.manualReviewSuggestion"),
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
        t("compliance.completeQualifications"),
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
  if (level === "red") return t("compliance.riskLevel.red");
  if (level === "orange") return t("compliance.riskLevel.orange");
  if (level === "yellow") return t("compliance.riskLevel.yellow");
  return t("compliance.riskLevel.green");
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
    { required: true, message: t("validation.projectNameRequired"), trigger: "blur" },
    { min: 1, max: 50, message: t("validation.projectNameLength"), trigger: "blur" },
  ],
  description: [
    { required: true, message: t("validation.projectDescRequired"), trigger: "blur" },
    {
      min: 1,
      max: 500,
      message: t("validation.projectDescLength"),
      trigger: "blur",
    },
  ],
  target_country: [
    {
      type: "array",
      required: true,
      min: 1,
      message: t("validation.targetCountryRequired"),
      trigger: "change",
    },
  ],
  material_composition: [
    { max: 200, message: t("validation.materialLength"), trigger: "blur" },
  ],
  marketing_selling_points: [
    { max: 200, message: t("validation.marketingLength"), trigger: "blur" },
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
  const normalized = normalizeComplianceResult(rawCompliance, locale.value);
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
  ElMessage.info(t("compliance.editHint"));
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
    ElMessage.warning(t("compliance.blockedRetry"));
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
      ElMessage.warning(t("compliance.orangeContinue"));
    } else {
      ElMessage.success(t("compliance.created"));
    }
    router.push(`/dramas/${dramaId}`);
  } catch (error: any) {
    if (error?.code === "COMPLIANCE_PRECHECK_REQUIRED") {
      pendingComplianceToken.value = "";
      complianceDialogVisible.value = false;
      ElMessage.warning(error.message || t("compliance.precheckExpired"));
      return;
    }
    const opened = openComplianceDialog(error?.details?.compliance);
    if (opened) {
      ElMessage.warning(error.message || t("compliance.riskChanged"));
      return;
    }
    ElMessage.error(error.message || t("compliance.createFailed"));
  } finally {
    complianceSubmitting.value = false;
  }
};

const handleExportCompliancePdf = () => {
  const report = currentCompliance.value;
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");
  if (!ctx) {
    ElMessage.error(t("compliance.exportFailed"));
    return;
  }

  const fontFamily = locale.value.startsWith("zh")
    ? '"PingFang SC", "Microsoft YaHei", "Segoe UI", Arial, sans-serif'
    : '"Inter", "Segoe UI", "Helvetica Neue", Arial, sans-serif';
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
    report.summary || t("compliance.noCategory"),
    contentWidth,
  );
  const issueGroups = (
    complianceIssueItems.value.length
      ? complianceIssueItems.value
      : [
          {
            level: report.level,
            title: t("compliance.noIssue"),
            suggestion: t("compliance.manualReviewSuggestion"),
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
    ? complianceCategories.value.join(complianceCategorySeparator.value)
    : t("compliance.noCategory");
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
  ctx.fillText(t("compliance.reportTitle"), marginX, y);

  y += 62;
  ctx.fillStyle = "#5b6b80";
  ctx.font = metaFont;
  ctx.fillText(`${t("compliance.checkedAt")}: ${complianceCheckedAt.value}`, marginX, y);
  y += 36;
  ctx.fillText(`${t("compliance.labels.score")}: ${report.score}`, marginX, y);
  y += 36;
  ctx.fillText(`${t("compliance.labels.level")}: ${getRiskLevelLabel(report.level)}`, marginX, y);

  y += 30;
  ctx.fillStyle = "#111827";
  ctx.font = headingFont;
  ctx.fillText(t("compliance.resultTitle"), marginX, y);
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
  ctx.fillText(t("compliance.details"), marginX, y);
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
  ctx.fillText(t("compliance.rectification"), marginX, y);
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
  ctx.fillText(t("compliance.suggestedCategories"), marginX, y);
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
  downloadBlob(pdfBlob, `${t("compliance.defaultFileName")}_${fileTime}.pdf`);
  ElMessage.success(t("compliance.pdfDownloaded"));
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
      ElMessage.error(t("compliance.missingResult"));
      return;
    }

    if (!isComplianceBlocked.value && !pendingComplianceToken.value) {
      pendingCreatePayload.value = null;
      complianceDialogVisible.value = false;
      ElMessage.error(t("compliance.missingToken"));
      return;
    }

    if (isComplianceBlocked.value) {
      ElMessage.warning(t("compliance.blockedRetry"));
      return;
    }

    if (isOrangeRisk.value) {
      ElMessage.warning(t("compliance.orangeContinue"));
    }
  } catch (error: any) {
    pendingCreatePayload.value = null;
    pendingComplianceToken.value = "";
    const opened = openComplianceDialog(error?.details?.compliance);
    if (opened) {
      ElMessage.warning(error.message || t("compliance.blockedRetry"));
      return;
    }
    ElMessage.error(error.message || t("compliance.createFailed"));
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
:global(.create-dialog-overlay) {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px 16px;
  overflow: hidden;
  overscroll-behavior: contain;
  background:
    radial-gradient(circle at 12% 18%, rgba(78, 180, 255, 0.12), transparent 24%),
    radial-gradient(circle at 88% 14%, rgba(255, 156, 84, 0.1), transparent 22%),
    radial-gradient(circle at 50% 100%, rgba(135, 102, 255, 0.08), transparent 28%),
    rgba(255, 255, 255, 0.06);
  backdrop-filter: blur(8px) saturate(1.04);
}

:global(.compliance-dialog-overlay) {
  position: fixed;
  inset: 0;
  z-index: 2001;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  overscroll-behavior: contain;
  background:
    radial-gradient(circle at 14% 18%, rgba(78, 180, 255, 0.12), transparent 24%),
    radial-gradient(circle at 86% 14%, rgba(255, 156, 84, 0.1), transparent 22%),
    radial-gradient(circle at 50% 100%, rgba(135, 102, 255, 0.09), transparent 28%),
    rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px) saturate(1.04);
}

:global(.create-dialog-overlay .el-overlay-dialog) {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

:global(.compliance-dialog-overlay .el-overlay-dialog) {
  position: relative;
  z-index: 1;
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
  border: 1px solid rgba(208, 220, 242, 0.78);
  border-radius: 32px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 251, 255, 0.93)),
    linear-gradient(135deg, rgba(109, 170, 255, 0.06), rgba(255, 147, 92, 0.05));
  box-shadow:
    0 34px 80px -44px rgba(26, 57, 106, 0.45),
    0 22px 42px -30px rgba(60, 117, 204, 0.24);
  backdrop-filter: blur(26px);
  overflow: hidden;
}

.create-dialog :deep(.el-dialog__header) {
  flex-shrink: 0;
  padding: 1.5rem 1.75rem 1.125rem;
  border-bottom: 1px solid rgba(220, 229, 245, 0.86);
  margin-right: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.68), rgba(255, 255, 255, 0.2));
}

.create-dialog :deep(.el-dialog__title) {
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: #1e3978;
}

.create-dialog :deep(.el-dialog__headerbtn) {
  top: 20px;
  right: 20px;
  width: 36px;
  height: 36px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.78);
  color: #6a7faa;
  box-shadow: 0 14px 28px -24px rgba(31, 67, 123, 0.45);
  transition: all 0.2s ease;
}

.create-dialog :deep(.el-dialog__headerbtn:hover) {
  background: rgba(255, 255, 255, 0.96);
  color: #1f4f9a;
  transform: translateY(-1px);
}

.create-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 1.5rem 1.75rem 1.25rem;
}

.create-dialog :deep(.el-dialog__footer) {
  flex-shrink: 0;
  padding: 1rem 1.75rem 1.75rem;
  border-top: 1px solid rgba(220, 229, 245, 0.86);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.2), rgba(248, 250, 255, 0.74));
}

.dialog-desc {
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
  color: #6f84ad;
}

.create-form :deep(.el-form-item) {
  margin-bottom: 1.35rem;
}

.create-form :deep(.el-form-item__label) {
  font-weight: 700;
  letter-spacing: -0.01em;
  color: #2d4a84;
  margin-bottom: 0.5rem;
}

.create-form :deep(.el-input__wrapper),
.create-form :deep(.el-textarea__inner),
.create-form :deep(.el-select__wrapper) {
  background: rgba(255, 255, 255, 0.82);
  border-radius: 18px;
  box-shadow:
    0 0 0 1px rgba(212, 224, 244, 0.96) inset,
    0 18px 32px -30px rgba(65, 111, 192, 0.28);
  transition: all 0.2s ease;
}

.create-form :deep(.el-input__wrapper:hover),
.create-form :deep(.el-textarea__inner:hover),
.create-form :deep(.el-select__wrapper:hover) {
  box-shadow:
    0 0 0 1px rgba(120, 173, 255, 0.9) inset,
    0 24px 40px -34px rgba(65, 111, 192, 0.34);
}

.create-form :deep(.el-input__wrapper.is-focus),
.create-form :deep(.el-textarea__inner:focus),
.create-form :deep(.el-select__wrapper.is-focused) {
  box-shadow:
    0 0 0 2px rgba(82, 147, 255, 0.72) inset,
    0 26px 44px -36px rgba(72, 118, 214, 0.4);
}

.create-form :deep(.el-input__inner),
.create-form :deep(.el-textarea__inner),
.create-form :deep(.el-select__selected-item) {
  color: #274176;
}

.create-form :deep(.el-input__wrapper),
.create-form :deep(.el-select__wrapper) {
  min-height: 52px;
}

.create-form :deep(.el-textarea__inner) {
  padding: 14px 16px;
  line-height: 1.6;
}

.country-select :deep(.el-select__placeholder) {
  color: #9eaec8;
}

.country-select.has-value :deep(.el-select__placeholder) {
  color: #274176;
}

.create-form :deep(.el-input__inner::placeholder),
.create-form :deep(.el-textarea__inner::placeholder) {
  color: #9eaec8;
}

.create-form :deep(.el-input__count) {
  color: #95a6c3;
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
  min-width: 116px;
  min-height: 50px;
  border-radius: 18px;
  font-weight: 700;
  box-shadow: 0 18px 34px -30px rgba(34, 62, 109, 0.32);
}

.dialog-footer .el-button:not(.el-button--primary) {
  border-color: rgba(214, 224, 242, 0.9);
  background: rgba(255, 255, 255, 0.8);
  color: #38588f;
}

.dialog-footer .el-button:not(.el-button--primary):hover {
  border-color: rgba(140, 183, 255, 0.86);
  background: rgba(255, 255, 255, 0.96);
  color: #234f97;
}

.dialog-footer .el-button--primary {
  border: none;
  background: linear-gradient(135deg, #ff7a1a, #ff9a3d);
  color: #fff;
}

.dialog-footer .el-button--primary:hover {
  background: linear-gradient(135deg, #ff8a2c, #ffad55);
}

.dialog-footer .el-button .el-icon {
  margin-right: 2px;
}

:global(.dark) :global(.create-dialog-overlay) {
  background:
    radial-gradient(circle at 12% 18%, rgba(78, 180, 255, 0.12), transparent 24%),
    radial-gradient(circle at 88% 14%, rgba(255, 156, 84, 0.1), transparent 22%),
    radial-gradient(circle at 50% 100%, rgba(135, 102, 255, 0.09), transparent 28%),
    rgba(6, 12, 24, 0.18);
}

:global(.dark) :deep(.el-dialog.create-dialog) {
  border-color: rgba(73, 96, 136, 0.72);
  background:
    linear-gradient(180deg, rgba(14, 25, 46, 0.96), rgba(18, 31, 57, 0.92)),
    linear-gradient(135deg, rgba(109, 170, 255, 0.06), rgba(255, 147, 92, 0.05));
  box-shadow:
    0 34px 80px -44px rgba(0, 0, 0, 0.75),
    0 22px 42px -30px rgba(40, 79, 141, 0.34);
}

:global(.dark) .create-dialog :deep(.el-dialog__header),
:global(.dark) .create-dialog :deep(.el-dialog__footer) {
  border-color: rgba(66, 89, 128, 0.72);
}

:global(.dark) .create-dialog :deep(.el-dialog__title),
:global(.dark) .create-form :deep(.el-form-item__label),
:global(.dark) .create-form :deep(.el-input__inner),
:global(.dark) .create-form :deep(.el-textarea__inner),
:global(.dark) .create-form :deep(.el-select__selected-item) {
  color: #e8efff;
}

:global(.dark) .create-dialog :deep(.el-dialog__headerbtn) {
  background: rgba(18, 31, 57, 0.8);
  color: #a9bce3;
}

:global(.dark) .create-form :deep(.el-input__wrapper),
:global(.dark) .create-form :deep(.el-textarea__inner),
:global(.dark) .create-form :deep(.el-select__wrapper) {
  background: rgba(16, 28, 50, 0.86);
  box-shadow:
    0 0 0 1px rgba(66, 89, 128, 0.88) inset,
    0 18px 32px -30px rgba(0, 0, 0, 0.48);
}

:global(.dark) .create-form :deep(.el-input__inner::placeholder),
:global(.dark) .create-form :deep(.el-textarea__inner::placeholder),
:global(.dark) .country-select :deep(.el-select__placeholder),
:global(.dark) .create-form :deep(.el-input__count) {
  color: #8195b9;
}

:deep(.el-dialog.compliance-dialog) {
  width: min(1480px, calc(100vw - 24px)) !important;
  max-width: calc(100vw - 24px);
  height: auto;
  max-height: calc(100vh - 60px);
  margin: 0 !important;
  display: flex;
  flex-direction: column;
  border: 1px solid rgba(208, 220, 242, 0.82);
  border-radius: 32px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 251, 255, 0.93)),
    linear-gradient(135deg, rgba(109, 170, 255, 0.06), rgba(255, 147, 92, 0.05));
  box-shadow:
    0 34px 80px -44px rgba(26, 57, 106, 0.45),
    0 22px 42px -30px rgba(60, 117, 204, 0.24);
  backdrop-filter: blur(26px);
  overflow: hidden;
}

.compliance-dialog :deep(.el-dialog__header) {
  flex-shrink: 0;
  padding: 1.5rem 1.75rem 1.125rem;
  border-bottom: 1px solid rgba(220, 229, 245, 0.86);
  margin-right: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.68), rgba(255, 255, 255, 0.2));
}

.compliance-dialog :deep(.el-dialog__title) {
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: #1e3978;
}

.compliance-dialog :deep(.el-dialog__headerbtn) {
  top: 20px;
  right: 20px;
  width: 36px;
  height: 36px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.78);
  color: #6a7faa;
  box-shadow: 0 14px 28px -24px rgba(31, 67, 123, 0.45);
  transition: all 0.2s ease;
}

.compliance-dialog :deep(.el-dialog__headerbtn:hover) {
  background: rgba(255, 255, 255, 0.96);
  color: #1f4f9a;
  transform: translateY(-1px);
}

.compliance-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 1.25rem 1.75rem 1.5rem;
  background: transparent;
}

.compliance-dialog :deep(.el-dialog__footer) {
  flex-shrink: 0;
  padding: 1rem 1.75rem 1.75rem;
  border-top: 1px solid rgba(220, 229, 245, 0.86);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.2), rgba(248, 250, 255, 0.74));
}

.compliance-meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
  color: #6f84ad;
  font-size: 0.96rem;
}

.compliance-meta-left {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
}

.pdf-export-btn {
  padding: 11px 16px;
  border: 1px solid rgba(214, 224, 242, 0.9);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.78);
  color: #355e9f;
  font-size: 0.95rem;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 18px 34px -30px rgba(34, 62, 109, 0.26);
  transition: all 0.2s ease;
  cursor: pointer;
}

.pdf-export-btn:hover {
  border-color: rgba(140, 183, 255, 0.86);
  background: rgba(255, 255, 255, 0.94);
  color: #1f4f9a;
  transform: translateY(-1px);
}

.compliance-alert {
  margin-bottom: 14px;
}

.compliance-alert :deep(.el-alert) {
  border-radius: 20px;
  border: 1px solid rgba(214, 224, 242, 0.82);
  box-shadow: 0 20px 36px -30px rgba(34, 62, 109, 0.22);
}

.compliance-alert :deep(.el-alert--error.is-light) {
  background: linear-gradient(135deg, rgba(255, 240, 242, 0.92), rgba(255, 247, 248, 0.88));
}

.compliance-alert :deep(.el-alert--warning.is-light) {
  background: linear-gradient(135deg, rgba(255, 246, 234, 0.92), rgba(255, 250, 242, 0.88));
}

.compliance-main-grid {
  display: grid;
  grid-template-columns: 340px minmax(0, 1fr);
  gap: 18px;
}

.risk-score-card,
.risk-details-card,
.rectification-card {
  border: 1px solid rgba(214, 224, 242, 0.88);
  border-radius: 26px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.86), rgba(247, 250, 255, 0.78)),
    radial-gradient(circle at top right, rgba(85, 104, 239, 0.08), transparent 28%);
  box-shadow: 0 28px 50px -40px rgba(34, 62, 109, 0.26);
  backdrop-filter: blur(16px);
}

.risk-score-card {
  padding: 24px 22px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.section-title {
  margin: 0;
  font-size: 1.34rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: #1f3d7a;
}

.risk-ring {
  width: 136px;
  height: 136px;
  margin-top: 18px;
  border-radius: 50%;
  background: conic-gradient(
    var(--risk-color) var(--risk-angle),
    rgba(224, 233, 247, 0.9) var(--risk-angle)
  );
  display: grid;
  place-items: center;
  box-shadow: inset 0 0 0 1px rgba(214, 224, 242, 0.8);
}

.risk-ring-inner {
  width: 102px;
  height: 102px;
  border-radius: 50%;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(247, 250, 255, 0.92));
  box-shadow: inset 0 0 0 1px rgba(220, 229, 245, 0.84);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.risk-score-value {
  font-size: 34px;
  line-height: 1;
  font-weight: 800;
  color: #18366d;
}

.risk-score-level {
  margin-top: 8px;
  font-size: 14px;
  font-weight: 800;
}

.risk-summary-text {
  margin: 16px 0 0;
  color: #5d79a3;
  font-size: 0.98rem;
  line-height: 1.8;
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
  padding: 18px 20px;
  border-bottom: 1px solid rgba(220, 229, 245, 0.82);
}

.pending-count {
  color: #ef6d1f;
  background: rgba(255, 150, 82, 0.16);
  border-radius: 999px;
  padding: 7px 14px;
  font-size: 12px;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(255, 150, 82, 0.18);
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
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(220, 229, 245, 0.78);
}

.risk-item:last-child {
  border-bottom: none;
}

.risk-item-dot {
  margin-top: 8px;
  width: 11px;
  height: 11px;
  border-radius: 50%;
  flex: 0 0 auto;
  box-shadow: 0 0 0 5px rgba(255, 255, 255, 0.72);
}

.risk-item-dot--red {
  background: #e85d54;
}

.risk-item-dot--orange {
  background: #f08a32;
}

.risk-item-dot--yellow {
  background: #e7ad27;
}

.risk-item-dot--green {
  background: #26b96a;
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
  font-size: 1.08rem;
  font-weight: 800;
  color: #1f3d7a;
  line-height: 1.6;
}

.risk-item-desc {
  margin: 8px 0 0;
  color: #5d79a3;
  font-size: 0.96rem;
  line-height: 1.75;
}

.risk-level-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 7px 14px;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.24);
}

.risk-level-chip--red {
  color: #c84646;
  background: rgba(232, 93, 84, 0.14);
}

.risk-level-chip--orange {
  color: #ef6d1f;
  background: rgba(255, 150, 82, 0.16);
}

.risk-level-chip--yellow {
  color: #cc8c06;
  background: rgba(231, 173, 39, 0.16);
}

.risk-level-chip--green {
  color: #179b53;
  background: rgba(38, 185, 106, 0.15);
}

.rectification-card {
  margin-top: 18px;
  padding: 20px 22px;
}

.rectification-list {
  margin: 14px 0 0;
  padding-left: 20px;
  color: #5d79a3;
  font-size: 0.95rem;
  line-height: 1.8;
  columns: 2;
  column-gap: 30px;
}

.rectification-list li {
  break-inside: avoid;
  margin-bottom: 10px;
}

.category-row {
  margin-top: 18px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.category-label {
  font-size: 0.94rem;
  color: #6f84ad;
  font-weight: 700;
  flex: 0 0 auto;
}

.category-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.category-tag {
  padding: 8px 14px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(52, 183, 232, 0.12), rgba(139, 92, 246, 0.1));
  color: #485ce4;
  font-size: 0.9rem;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(140, 183, 255, 0.26);
}

.compliance-footer {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.compliance-footer .el-button {
  min-width: 120px;
  min-height: 52px;
  border-radius: 18px;
  font-weight: 700;
  box-shadow: 0 18px 34px -30px rgba(34, 62, 109, 0.32);
}

.footer-secondary-btn {
  border-color: rgba(214, 224, 242, 0.9);
  color: #38588f;
  background: rgba(255, 255, 255, 0.8);
}

.footer-secondary-btn:hover {
  border-color: rgba(140, 183, 255, 0.86);
  color: #234f97;
  background: rgba(255, 255, 255, 0.96);
}

.footer-primary-btn {
  border: none;
  background: linear-gradient(135deg, #ff7a1a, #ff9a3d);
  box-shadow: 0 24px 40px -28px rgba(255, 108, 30, 0.46);
}

.footer-primary-btn:hover {
  background: linear-gradient(135deg, #ff8a2c, #ffad55);
  box-shadow: 0 26px 44px -30px rgba(255, 108, 30, 0.52);
  color: #fff;
}

:global(.dark) :global(.compliance-dialog-overlay) {
  background:
    radial-gradient(circle at 14% 18%, rgba(78, 180, 255, 0.12), transparent 24%),
    radial-gradient(circle at 86% 14%, rgba(255, 156, 84, 0.1), transparent 22%),
    radial-gradient(circle at 50% 100%, rgba(135, 102, 255, 0.09), transparent 28%),
    rgba(6, 12, 24, 0.18);
}

:global(.dark) :deep(.el-dialog.compliance-dialog) {
  border-color: rgba(73, 96, 136, 0.72);
  background:
    linear-gradient(180deg, rgba(14, 25, 46, 0.96), rgba(18, 31, 57, 0.92)),
    linear-gradient(135deg, rgba(109, 170, 255, 0.06), rgba(255, 147, 92, 0.05));
  box-shadow:
    0 34px 80px -44px rgba(0, 0, 0, 0.75),
    0 22px 42px -30px rgba(40, 79, 141, 0.34);
}

:global(.dark) .compliance-dialog :deep(.el-dialog__header),
:global(.dark) .compliance-dialog :deep(.el-dialog__footer),
:global(.dark) .risk-details-header {
  border-color: rgba(66, 89, 128, 0.72);
}

:global(.dark) .compliance-dialog :deep(.el-dialog__title),
:global(.dark) .section-title,
:global(.dark) .risk-score-value,
:global(.dark) .risk-item-title {
  color: #e8efff;
}

:global(.dark) .compliance-dialog :deep(.el-dialog__headerbtn) {
  background: rgba(18, 31, 57, 0.8);
  color: #a9bce3;
}

:global(.dark) .compliance-meta-row,
:global(.dark) .category-label,
:global(.dark) .risk-summary-text,
:global(.dark) .risk-item-desc,
:global(.dark) .rectification-list {
  color: #98acd2;
}

:global(.dark) .pdf-export-btn,
:global(.dark) .footer-secondary-btn {
  border-color: rgba(66, 89, 128, 0.82);
  background: rgba(16, 28, 50, 0.82);
  color: #d7e4ff;
}

:global(.dark) .risk-score-card,
:global(.dark) .risk-details-card,
:global(.dark) .rectification-card,
:global(.dark) .risk-ring-inner {
  border-color: rgba(66, 89, 128, 0.82);
  background:
    linear-gradient(180deg, rgba(16, 28, 50, 0.9), rgba(13, 22, 40, 0.84)),
    radial-gradient(circle at top right, rgba(85, 104, 239, 0.14), transparent 28%);
}

:global(.dark) .risk-item {
  border-color: rgba(66, 89, 128, 0.58);
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
