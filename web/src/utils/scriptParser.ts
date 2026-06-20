import type { ParseScriptRequest, ParseScriptResult, ParsedCharacter, ParsedEpisode } from '@/types/generation'

const episodeMarkerRE = /(?:^|\n)\s*(?:第\s*(\d+)\s*[集章]|内容\s*(\d+)|Episode\s*(\d+))[\s:：.-]*(.*)/gi
const sceneMarkerRE = /(?:^|\n)\s*(?:场景|镜头|Scene)\s*(\d+)?[\s:：.-]*(.*)/gi
const dialogueRE = /^\s*([\u4e00-\u9fa5A-Za-z][\u4e00-\u9fa5A-Za-z0-9 _-]{1,20})\s*[：:]\s*(.+)$/

const trimText = (value: string, fallback = '') => value.trim() || fallback

const splitByMarkers = (content: string, marker: RegExp) => {
  const matches = [...content.matchAll(marker)]
  if (matches.length === 0) return []

  return matches.map((match, index) => {
    const start = (match.index || 0) + match[0].length
    const end = matches[index + 1]?.index ?? content.length
    return {
      match,
      body: content.slice(start, end).trim()
    }
  })
}

const parseScenes = (episodeContent: string) => {
  const sceneBlocks = splitByMarkers(episodeContent, sceneMarkerRE)
  const blocks = sceneBlocks.length > 0
    ? sceneBlocks
    : [{ match: null as RegExpMatchArray | null, body: episodeContent.trim() }]

  return blocks.map((block, index) => {
    const marker = block.match
    const markerTitle = marker ? trimText(marker[2] || '', `营销场景 ${index + 1}`) : `营销场景 ${index + 1}`
    const lines = block.body.split(/\n+/).map((line) => line.trim()).filter(Boolean)
    const dialogueLines = lines.filter((line) => dialogueRE.test(line))
    const description = lines.filter((line) => !dialogueRE.test(line)).join('\n')
    const characters = dialogueLines
      .map((line) => line.match(dialogueRE)?.[1]?.trim())
      .filter(Boolean)

    return {
      storyboard_number: Number(marker?.[1]) || index + 1,
      title: markerTitle,
      location: markerTitle,
      time: '',
      characters: Array.from(new Set(characters)).join('、'),
      dialogue: dialogueLines.join('\n'),
      description: trimText(description, block.body.slice(0, 240))
    }
  })
}

const parseCharacters = (content: string): ParsedCharacter[] => {
  const names = new Set<string>()
  content.split(/\n+/).forEach((line) => {
    const match = line.match(dialogueRE)
    if (match?.[1]) {
      names.add(match[1].trim())
    }
  })

  return Array.from(names).slice(0, 20).map((name, index) => ({
    name,
    role: index === 0 ? 'main' : 'supporting',
    description: '从营销脚本口播或对话中识别出的数字人表达角色',
    personality: '清晰、可信、适合跨境商品讲解'
  }))
}

export const parseMarketingScript = (request: ParseScriptRequest): ParseScriptResult => {
  const content = request.script_content.trim()
  if (!content) {
    return {
      episodes: [],
      characters: [],
      summary: ''
    }
  }

  const episodeBlocks = request.auto_split === false ? [] : splitByMarkers(content, episodeMarkerRE)
  const blocks = episodeBlocks.length > 0
    ? episodeBlocks
    : [{ match: null as RegExpMatchArray | null, body: content }]

  const episodes: ParsedEpisode[] = blocks.map((block, index) => {
    const marker = block.match
    const episodeNumber = Number(marker?.[1] || marker?.[2] || marker?.[3]) || index + 1
    const title = trimText(marker?.[4] || '', `营销内容 ${episodeNumber}`)
    const scenes = parseScenes(block.body)

    return {
      episode_number: episodeNumber,
      title,
      description: block.body.slice(0, 280),
      script_content: block.body,
      duration: Math.max(15, scenes.length * 6),
      scenes
    }
  })

  return {
    episodes,
    characters: parseCharacters(content),
    summary: content.slice(0, 500)
  }
}
