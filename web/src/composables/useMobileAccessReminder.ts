import { onBeforeUnmount, onMounted, ref } from 'vue'

import {
  hasDismissedMobileAccessReminder,
  isMobileOrTabletAccess,
  markMobileAccessReminderDismissed
} from '@/utils/mobileAccess'

export const useMobileAccessReminder = () => {
  const visible = ref(false)

  const openReminder = () => {
    if (hasDismissedMobileAccessReminder()) return
    if (!isMobileOrTabletAccess()) return
    visible.value = true
  }

  const dismissReminder = () => {
    markMobileAccessReminderDismissed()
    visible.value = false
  }

  const handleWindowLoad = () => {
    window.requestAnimationFrame(openReminder)
  }

  onMounted(() => {
    if (hasDismissedMobileAccessReminder()) return

    if (document.readyState === 'complete') {
      handleWindowLoad()
      return
    }

    window.addEventListener('load', handleWindowLoad, { once: true })
  })

  onBeforeUnmount(() => {
    window.removeEventListener('load', handleWindowLoad)
  })

  return {
    visible,
    dismissReminder
  }
}
