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

export const clearCreateDramaDraft = () => {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.removeItem(CREATE_DRAMA_DRAFT_KEY);
};
