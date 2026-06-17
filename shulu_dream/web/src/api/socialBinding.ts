import request from '../utils/request'

export type SocialBindingPlatform = 'discord' | 'reddit' | 'pinterest'

export interface SocialAccountBinding {
  id: number
  platform: SocialBindingPlatform
  account_identifier: string
  display_name?: string
  created_at: string
  updated_at: string
}

export const socialBindingAPI = {
  async listBindings(): Promise<SocialAccountBinding[]> {
    const response = await request.get<{ bindings: SocialAccountBinding[] }>('/social-bindings')
    return response.bindings || []
  },

  async upsertBinding(
    platform: SocialBindingPlatform,
    data: { account_identifier: string; display_name?: string }
  ): Promise<SocialAccountBinding> {
    const response = await request.put<{ binding: SocialAccountBinding }>(`/social-bindings/${platform}`, data)
    return response.binding
  }
}
