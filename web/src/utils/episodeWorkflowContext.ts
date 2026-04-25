export interface EpisodeWorkflowContext {
  dramaId: string
  episodeId: string
  episodeNumber: number
}

const STORAGE_PREFIX = 'drama:episode-workflow:context:'

const normalizeString = (value: unknown): string => String(value || '').trim()

const normalizeNumber = (value: unknown): number => {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
}

const getStorageKey = (episodeId: string) => `${STORAGE_PREFIX}${episodeId}`

export const saveEpisodeWorkflowContext = (
  context: Partial<EpisodeWorkflowContext> & { episodeId: string }
) => {
  if (typeof window === 'undefined') return

  const episodeId = normalizeString(context.episodeId)
  if (!episodeId) return

  const existing = getEpisodeWorkflowContext(episodeId)

  const merged: EpisodeWorkflowContext = {
    dramaId: normalizeString(context.dramaId || existing?.dramaId),
    episodeId,
    episodeNumber: normalizeNumber(context.episodeNumber || existing?.episodeNumber)
  }

  if (!merged.dramaId || !merged.episodeNumber) return

  window.sessionStorage.setItem(getStorageKey(episodeId), JSON.stringify(merged))
}

export const getEpisodeWorkflowContext = (episodeId?: string | number | null): EpisodeWorkflowContext | null => {
  if (typeof window === 'undefined') return null

  const normalizedEpisodeId = normalizeString(episodeId)
  if (!normalizedEpisodeId) return null

  const raw = window.sessionStorage.getItem(getStorageKey(normalizedEpisodeId))
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as Partial<EpisodeWorkflowContext>
    const context: EpisodeWorkflowContext = {
      dramaId: normalizeString(parsed.dramaId),
      episodeId: normalizedEpisodeId,
      episodeNumber: normalizeNumber(parsed.episodeNumber)
    }

    if (!context.dramaId || !context.episodeNumber) {
      window.sessionStorage.removeItem(getStorageKey(normalizedEpisodeId))
      return null
    }

    return context
  } catch {
    window.sessionStorage.removeItem(getStorageKey(normalizedEpisodeId))
    return null
  }
}

export const resolveEpisodeWorkflowContext = (input: {
  episodeId?: unknown
  dramaId?: unknown
  episodeNumber?: unknown
}): EpisodeWorkflowContext | null => {
  const episodeId = normalizeString(input.episodeId)
  if (!episodeId) return null

  const directDramaId = normalizeString(input.dramaId)
  const directEpisodeNumber = normalizeNumber(input.episodeNumber)

  if (directDramaId && directEpisodeNumber) {
    return {
      dramaId: directDramaId,
      episodeId,
      episodeNumber: directEpisodeNumber
    }
  }

  return getEpisodeWorkflowContext(episodeId)
}

const buildQuery = (context: EpisodeWorkflowContext) => {
  const query = new URLSearchParams({
    dramaId: context.dramaId,
    episodeNumber: String(context.episodeNumber)
  })
  return query.toString()
}

export const buildEpisodeStagePath = (
  stage: 'script' | 'generation' | 'timeline',
  context: EpisodeWorkflowContext
) => {
  const query = buildQuery(context)

  if (stage === 'script') {
    return `/episodes/${context.episodeId}/edit?${query}`
  }

  if (stage === 'generation') {
    return `/episodes/${context.episodeId}/generate?${query}`
  }

  return `/timeline/${context.episodeId}?${query}`
}

export const buildProfessionalEditorPath = (context: EpisodeWorkflowContext) =>
  `/dramas/${context.dramaId}/episode/${context.episodeNumber}/professional`
