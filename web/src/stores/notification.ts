import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { notificationAPI } from '@/api/notification'
import type { CreateNotificationPayload, NotificationItem, NotificationStreamEvent } from '@/types/notification'

export const useNotificationStore = defineStore('notification', () => {
  const items = ref<NotificationItem[]>([])
  const unreadCount = ref(0)
  const loading = ref(false)
  const connected = ref(false)
  const errorMessage = ref('')

  let source: EventSource | null = null
  let loadPromise: Promise<void> | null = null
  let lastLocalMutationAt = 0

  const visibleItems = computed(() => items.value.filter((item) => !item.dismissed_at))

  const load = async () => {
    if (loadPromise) {
      return loadPromise
    }

    loading.value = true
    errorMessage.value = ''
    const startedAt = Date.now()

    loadPromise = (async () => {
      try {
        const response = await notificationAPI.list()
        if (startedAt >= lastLocalMutationAt) {
          items.value = response.items || []
          unreadCount.value = Number(response.unreadCount || 0)
        }
      } catch {
        errorMessage.value = ''
      } finally {
        loading.value = false
        loadPromise = null
      }
    })()

    return loadPromise
  }

  const create = async (payload: CreateNotificationPayload) => {
    const created = await notificationAPI.create(payload)
    mergeNotification(created, true)
    unreadCount.value += created.read_at ? 0 : 1
    return created
  }

  const markRead = async (id: number) => {
    const existing = items.value.find((item) => item.id === id)
    if (existing && !existing.read_at) {
      lastLocalMutationAt = Date.now()
      const now = new Date().toISOString()
      items.value = items.value.map((item) => (item.id === id ? { ...item, read_at: now } : item))
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }

    const updated = await notificationAPI.markRead(id)
    mergeNotification(updated)
    return updated
  }

  const markAllRead = async () => {
    lastLocalMutationAt = Date.now()
    const now = new Date().toISOString()

    items.value = items.value.map((item) => ({
      ...item,
      read_at: item.read_at || now
    }))
    unreadCount.value = 0

    try {
      await notificationAPI.markAllRead()
    } catch {
      errorMessage.value = ''
    }
  }

  const dismiss = async (id: number) => {
    lastLocalMutationAt = Date.now()
    const previousItems = items.value
    items.value = items.value.filter((item) => item.id !== id)
    unreadCount.value = items.value.filter((item) => !item.read_at && !item.dismissed_at).length

    try {
      await notificationAPI.dismiss(id)
      await refreshUnreadCount()
    } catch (error: any) {
      errorMessage.value = error?.message || '忽略通知失败'
      items.value = previousItems
      await refreshUnreadCount()
      throw error
    }
  }

  const refreshUnreadCount = async () => {
    try {
      const response = await notificationAPI.unreadCount()
      unreadCount.value = Number(response.unreadCount || 0)
    } catch {
      // Keep the current count if the lightweight refresh fails.
    }
  }

  const connect = () => {
    if (source || typeof window === 'undefined' || typeof EventSource === 'undefined') {
      return
    }

    source = new EventSource(notificationAPI.streamURL())
    source.addEventListener('open', () => {
      connected.value = true
    })
    source.addEventListener('error', () => {
      connected.value = false
    })
    source.addEventListener('snapshot', handleStreamMessage)
    source.addEventListener('notification', handleStreamMessage)
  }

  const disconnect = () => {
    if (!source) return
    source.close()
    source = null
    connected.value = false
  }

  const handleStreamMessage = (event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data) as NotificationStreamEvent
      unreadCount.value = Number(data.unreadCount || 0)

      if (data.notification) {
        mergeNotification(data.notification, data.type === 'created')
      }
      if (data.type === 'read_all') {
        items.value = items.value.map((item) => ({
          ...item,
          read_at: item.read_at || new Date().toISOString()
        }))
      }
      if (data.type === 'dismissed') {
        void load()
      }
    } catch {
      // Ignore malformed SSE frames.
    }
  }

  const mergeNotification = (notification: NotificationItem, prepend = false) => {
    const index = items.value.findIndex((item) => item.id === notification.id)
    if (index >= 0) {
      items.value.splice(index, 1, notification)
      return
    }

    if (prepend) {
      items.value.unshift(notification)
    } else {
      items.value.push(notification)
    }
  }

  return {
    items,
    visibleItems,
    unreadCount,
    loading,
    connected,
    errorMessage,
    load,
    create,
    markRead,
    markAllRead,
    dismiss,
    connect,
    disconnect
  }
})
