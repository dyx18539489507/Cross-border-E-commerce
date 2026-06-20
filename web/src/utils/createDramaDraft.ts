/**
 * 模块说明：商品录入到后端项目/合规接口的临时适配草稿。
 * 业务场景：数字丝路商品录入完成后，仍需复用现有 /projects 创建和合规预检接口。
 * 核心职责：把 ProductEntryDraft 派生出的 CreateDramaRequest 暂存到会话中，并在合规页读取。
 */
import type { CreateDramaRequest } from "@/types/drama";
import { buildCreateDramaPayload } from "@/utils/compliance";

const CREATE_DRAMA_DRAFT_KEY = "drama:create:draft";

const normalizeDraft = (value: unknown): CreateDramaRequest | null => {
  if (!value || typeof value !== "object") {
    return null;
  }

  const raw = value as Partial<CreateDramaRequest>;
  return {
    title: typeof raw.title === "string" ? raw.title.trim() : "",
    description:
      typeof raw.description === "string" ? raw.description.trim() : "",
    target_country: Array.isArray(raw.target_country)
      ? raw.target_country
          .map((item) => String(item).trim())
          .filter((item) => item.length > 0)
      : [],
    material_composition:
      typeof raw.material_composition === "string"
        ? raw.material_composition.trim()
        : "",
    marketing_selling_points:
      typeof raw.marketing_selling_points === "string"
        ? raw.marketing_selling_points.trim()
        : "",
    genre: typeof raw.genre === "string" ? raw.genre : undefined,
    tags: typeof raw.tags === "string" ? raw.tags : undefined,
  };
};

/**
 * 功能：保存可提交给合规预检和项目创建的请求草稿。
 * 参数：form 为商品录入流程适配出的 CreateDramaRequest。
 * 返回：无返回值；写入 sessionStorage，供合规页继续创建后续脚本/分镜流程。
 */
export const saveCreateDramaDraft = (form: CreateDramaRequest) => {
  if (typeof window === "undefined") {
    return;
  }

  const payload = buildCreateDramaPayload(form);
  window.sessionStorage.setItem(
    CREATE_DRAMA_DRAFT_KEY,
    JSON.stringify(payload),
  );
};

/**
 * 功能：读取并移除一次性创建草稿。
 * 参数：无。
 * 返回：CreateDramaRequest 或 null；用于真正提交创建后避免旧商品资料被重复使用。
 */
export const consumeCreateDramaDraft = (): CreateDramaRequest | null => {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.sessionStorage.getItem(CREATE_DRAMA_DRAFT_KEY);
  if (!raw) {
    return null;
  }

  window.sessionStorage.removeItem(CREATE_DRAMA_DRAFT_KEY);

  try {
    return normalizeDraft(JSON.parse(raw));
  } catch {
    return null;
  }
};

/**
 * 功能：只查看当前创建草稿而不消费。
 * 参数：无。
 * 返回：CreateDramaRequest 或 null；合规页需要先展示和预检，再由创建动作决定是否清理。
 */
export const peekCreateDramaDraft = (): CreateDramaRequest | null => {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.sessionStorage.getItem(CREATE_DRAMA_DRAFT_KEY);
  if (!raw) {
    return null;
  }

  try {
    return normalizeDraft(JSON.parse(raw));
  } catch {
    return null;
  }
};

export const clearCreateDramaDraft = () => {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.removeItem(CREATE_DRAMA_DRAFT_KEY);
};
