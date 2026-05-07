export interface NotificationItem {
  id: number
  type: string
  title: string
  content: string
  path?: string
  metadata?: Record<string, unknown>
  read_at?: string | null
  dismissed_at?: string | null
  created_at: string
  updated_at: string
}

export interface NotificationListResponse {
  items: NotificationItem[]
  unreadCount: number
}

export interface NotificationUnreadResponse {
  unreadCount: number
}

export interface CreateNotificationPayload {
  type?: string
  title: string
  content: string
  path?: string
  metadata?: Record<string, unknown>
}

export interface NotificationStreamEvent {
  type: 'snapshot' | 'created' | 'updated' | 'read_all' | 'dismissed' | string
  unreadCount: number
  notification?: NotificationItem
}
