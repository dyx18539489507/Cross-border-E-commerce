export const MOBILE_ACCESS_REMINDER_MAX_WIDTH = 1024
export const MOBILE_ACCESS_REMINDER_SESSION_KEY = 'drama-mobile-access-reminder-dismissed'

const MOBILE_OR_TABLET_USER_AGENT_RE =
  /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini|Mobile|Tablet|Phone|PlayBook|Silk|Kindle|Nexus 7|Nexus 9|Nexus 10|SM-T|Tab/i
const IPAD_DESKTOP_USER_AGENT_RE = /Macintosh/i

interface MobileAccessDetectorOptions {
  userAgent?: string
  viewportWidth?: number
  maxTouchPoints?: number
}

export const isMobileOrTabletAccess = (options: MobileAccessDetectorOptions = {}) => {
  const userAgent = options.userAgent ?? (typeof navigator !== 'undefined' ? navigator.userAgent : '')
  const viewportWidth = options.viewportWidth ?? (typeof window !== 'undefined' ? window.innerWidth : Number.POSITIVE_INFINITY)
  const maxTouchPoints = options.maxTouchPoints ?? (typeof navigator !== 'undefined' ? navigator.maxTouchPoints || 0 : 0)

  const isCompactViewport = viewportWidth <= MOBILE_ACCESS_REMINDER_MAX_WIDTH
  const hasMobileOrTabletUserAgent = MOBILE_OR_TABLET_USER_AGENT_RE.test(userAgent)
  const isTabletDesktopUserAgent = IPAD_DESKTOP_USER_AGENT_RE.test(userAgent) && maxTouchPoints > 1

  return isCompactViewport && (hasMobileOrTabletUserAgent || isTabletDesktopUserAgent)
}

export const hasDismissedMobileAccessReminder = () => {
  if (typeof window === 'undefined') return false

  try {
    return window.sessionStorage.getItem(MOBILE_ACCESS_REMINDER_SESSION_KEY) === '1'
  } catch {
    return false
  }
}

export const markMobileAccessReminderDismissed = () => {
  if (typeof window === 'undefined') return

  try {
    window.sessionStorage.setItem(MOBILE_ACCESS_REMINDER_SESSION_KEY, '1')
  } catch {
    // Ignore storage failures and allow the reminder to show again next time.
  }
}
