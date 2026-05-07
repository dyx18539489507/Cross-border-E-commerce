<template>
  <div ref="rootRef" class="notification-bell">
    <button
      type="button"
      class="notification-bell__button"
      aria-label="消息提醒"
      :aria-expanded="open"
      aria-haspopup="dialog"
      @click="toggle"
    >
      <img :src="bellIcon" alt="" />
      <span v-if="store.unreadCount > 0" class="notification-bell__dot">
        {{ store.unreadCount > 99 ? '99+' : store.unreadCount }}
      </span>
    </button>

    <section v-if="open" class="notification-bell__popover" aria-label="消息通知">
      <div class="notification-bell__head">
        <div>
          <strong>消息通知</strong>
          <span>{{ store.unreadCount }} 条未读消息</span>
        </div>
        <button
          type="button"
          class="notification-bell__link"
          :disabled="store.unreadCount === 0 || markingAllRead"
          @click.stop="markAllRead"
        >
          {{ markingAllRead ? '处理中' : '全部已读' }}
        </button>
      </div>

      <div class="notification-bell__list">
        <article
          v-for="notice in store.visibleItems"
          :key="notice.id"
          class="notification-bell__item"
          :class="{ 'notification-bell__item--unread': !notice.read_at }"
        >
          <span class="notification-bell__status" aria-hidden="true"></span>
          <div class="notification-bell__body">
            <div class="notification-bell__title-row">
              <strong>{{ notice.title }}</strong>
              <span>{{ formatTime(notice.created_at) }}</span>
            </div>
            <p>{{ notice.content }}</p>
            <div class="notification-bell__actions">
              <button type="button" @click="openNotification(notice)">查看</button>
              <button type="button" @click="dismissNotification(notice.id)">忽略</button>
            </div>
          </div>
        </article>

        <p v-if="store.loading && store.visibleItems.length === 0" class="notification-bell__empty">正在加载消息...</p>
        <p v-else-if="store.visibleItems.length === 0" class="notification-bell__empty">
          暂无新的消息
        </p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import bellIcon from '@/assets/workbench/bell.svg'
import { useNotificationStore } from '@/stores/notification'
import type { NotificationItem } from '@/types/notification'

const store = useNotificationStore()
const router = useRouter()
const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)
const markingAllRead = ref(false)

const toggle = async () => {
  open.value = !open.value
  if (open.value) {
    await store.load()
  }
}

const markAllRead = async () => {
  if (store.unreadCount === 0 || markingAllRead.value) return
  markingAllRead.value = true
  try {
    await store.markAllRead()
  } catch {
    // Store exposes a user-facing error message inside the popover.
  } finally {
    markingAllRead.value = false
  }
}

const openNotification = async (notice: NotificationItem) => {
  if (!notice.read_at) {
    await store.markRead(notice.id)
  }
  open.value = false
  if (notice.path) {
    router.push(notice.path)
  }
}

const dismissNotification = async (id: number) => {
  await store.dismiss(id).catch(() => {})
}

const handleDocumentClick = (event: MouseEvent) => {
  if (!open.value || !rootRef.value) return
  if (!rootRef.value.contains(event.target as Node)) {
    open.value = false
  }
}

const formatTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const diff = Date.now() - date.getTime()
  if (diff < 60_000) return '刚刚'
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} 分钟前`
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} 小时前`
  return `${Math.floor(diff / 86_400_000)} 天前`
}

onMounted(() => {
  store.load()
  store.connect()
  document.addEventListener('click', handleDocumentClick)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
  store.disconnect()
})
</script>

<style scoped>
.notification-bell {
  position: relative;
}

.notification-bell__button {
  position: relative;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 12px;
  background: transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 180ms ease;
}

.notification-bell__button:hover {
  background: rgba(241, 245, 249, 0.92);
}

.notification-bell__button img {
  width: 20px;
  height: 20px;
  display: block;
}

.notification-bell__dot {
  position: absolute;
  top: 2px;
  left: 22px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: #f97316;
  border: 2px solid #ffffff;
  color: #ffffff;
  font-size: 10px;
  font-weight: 700;
  line-height: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 16px rgba(249, 115, 22, 0.22);
}

.notification-bell__popover {
  position: absolute;
  top: calc(100% + 14px);
  right: 0;
  z-index: 40;
  width: 340px;
  border: 1px solid rgba(226, 232, 240, 0.96);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.98);
  box-shadow:
    0 24px 48px rgba(15, 23, 42, 0.14),
    0 8px 18px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}

.notification-bell__popover::before {
  content: '';
  position: absolute;
  top: -7px;
  right: 18px;
  width: 14px;
  height: 14px;
  background: #ffffff;
  border-top: 1px solid rgba(226, 232, 240, 0.96);
  border-left: 1px solid rgba(226, 232, 240, 0.96);
  transform: rotate(45deg);
}

.notification-bell__head {
  position: relative;
  z-index: 1;
  padding: 18px 18px 14px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid #eef2f7;
}

.notification-bell__head div {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.notification-bell__head strong {
  color: #0a2463;
  font-size: 16px;
  font-weight: 700;
  line-height: 22px;
}

.notification-bell__head span {
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
}

.notification-bell__link,
.notification-bell__actions button {
  border: none;
  padding: 0;
  background: transparent;
  color: #2563eb;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
  cursor: pointer;
  white-space: nowrap;
}

.notification-bell__link:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.notification-bell__list {
  position: relative;
  z-index: 1;
  max-height: 360px;
  overflow-y: auto;
}

.notification-bell__item {
  display: grid;
  grid-template-columns: 8px 1fr;
  gap: 10px;
  padding: 14px 18px;
  border-bottom: 1px solid #f1f5f9;
  background: #ffffff;
}

.notification-bell__item:last-child {
  border-bottom: none;
}

.notification-bell__item--unread {
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.08) 0%, rgba(124, 58, 237, 0.06) 100%);
}

.notification-bell__status {
  width: 8px;
  height: 8px;
  margin-top: 7px;
  border-radius: 999px;
  background: #cbd5e1;
}

.notification-bell__item--unread .notification-bell__status {
  background: #f97316;
}

.notification-bell__body {
  min-width: 0;
}

.notification-bell__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.notification-bell__title-row strong {
  min-width: 0;
  color: #0a2463;
  font-size: 14px;
  font-weight: 700;
  line-height: 20px;
}

.notification-bell__title-row span {
  color: #90a1b9;
  font-size: 12px;
  line-height: 18px;
  white-space: nowrap;
}

.notification-bell__item p {
  margin: 4px 0 0;
  color: #45556c;
  font-size: 13px;
  line-height: 20px;
}

.notification-bell__actions {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.notification-bell__actions button:last-child {
  color: #64748b;
}

.notification-bell__link:hover,
.notification-bell__actions button:hover {
  color: #0a2463;
}

.notification-bell__empty {
  margin: 0;
  padding: 28px 18px;
  color: #64748b;
  font-size: 14px;
  line-height: 22px;
  text-align: center;
}

.notification-bell__empty--error {
  color: #b42318;
}

@media (max-width: 640px) {
  .notification-bell {
    position: static;
  }

  .notification-bell__popover {
    position: fixed;
    top: 78px;
    right: 16px;
    left: 16px;
    width: auto;
  }

  .notification-bell__popover::before {
    right: 22px;
  }
}
</style>
