import request, { buildAPIURL, getClientDeviceID } from '@/utils/request'
import type {
  CreateNotificationPayload,
  NotificationItem,
  NotificationListResponse,
  NotificationUnreadResponse
} from '@/types/notification'

const NOTIFICATION_REQUEST_TIMEOUT = 8000

export const notificationAPI = {
  list(limit = 20) {
    return request.get<NotificationListResponse>('/notifications', {
      params: { limit },
      timeout: NOTIFICATION_REQUEST_TIMEOUT
    })
  },

  unreadCount() {
    return request.get<NotificationUnreadResponse>('/notifications/unread-count', {
      timeout: NOTIFICATION_REQUEST_TIMEOUT
    })
  },

  create(data: CreateNotificationPayload) {
    return request.post<NotificationItem>('/notifications', data, {
      timeout: NOTIFICATION_REQUEST_TIMEOUT
    })
  },

  markRead(id: number) {
    return request.patch<NotificationItem>(`/notifications/${id}/read`, undefined, {
      timeout: NOTIFICATION_REQUEST_TIMEOUT
    })
  },

  markAllRead() {
    return request.patch<{ message: string }>('/notifications/read-all', undefined, {
      timeout: NOTIFICATION_REQUEST_TIMEOUT
    })
  },

  dismiss(id: number) {
    return request.delete<{ message: string }>(`/notifications/${id}`, {
      timeout: NOTIFICATION_REQUEST_TIMEOUT
    })
  },

  streamURL() {
    return `${buildAPIURL('/notifications/stream')}?device_id=${encodeURIComponent(getClientDeviceID())}`
  }
}
