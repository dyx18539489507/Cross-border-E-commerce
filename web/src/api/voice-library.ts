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
  is_custom?: boolean
  status?: string
  resource_id?: string
  source_audio_url?: string
  last_error?: string
}

export const voiceLibraryAPI = {
  list: () => request.get<VoiceLibraryItem[]>('/voice-library'),

  createCustom: (audioFile: File, name?: string) => {
    const formData = new FormData()
    formData.append('audio', audioFile)
    if (name?.trim()) {
      formData.append('name', name.trim())
    }

    return request.post<VoiceLibraryItem>('/voice-library/custom', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  getCustomStatus: (id: string) => request.get<VoiceLibraryItem>(`/voice-library/custom/${id}/status`)
}
