import request from '@/utils/request'

export interface VoiceLibraryItem {
  id: string
  voice_type: string
  name: string
  avatar?: string
  gender?: string
  age?: string
  trial_url?: string
  categories?: string[]
}

export const voiceLibraryAPI = {
  // `request` already has baseURL `/api/v1`, so only provide the resource path here.
  list: () => request.get<VoiceLibraryItem[]>('/voice-library')
}
