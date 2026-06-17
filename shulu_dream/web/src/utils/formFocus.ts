const NAVIGABLE_SELECTOR = [
  'input:not([type="hidden"]):not([disabled])',
  'textarea:not([disabled])',
  '[contenteditable="true"]'
].join(', ')

const BLOCKED_CONTAINER_SELECTOR = [
  '.el-select',
  '.el-cascader',
  '.el-date-editor',
  '.el-time-picker',
  '.el-autocomplete',
  '.el-mention'
].join(', ')

const isVisible = (element: HTMLElement) => {
  if (!element.isConnected) return false
  return !!(element.offsetWidth || element.offsetHeight || element.getClientRects().length)
}

const getFocusableFields = (root: HTMLElement) => {
  return Array.from(root.querySelectorAll<HTMLElement>(NAVIGABLE_SELECTOR)).filter(
    (element) => isVisible(element) && element.tabIndex !== -1
  )
}

export const handleFormEnterNavigation = (event: KeyboardEvent) => {
  if (event.key !== 'Enter' || event.isComposing) return
  if ((event.target as HTMLElement | null)?.tagName === 'TEXTAREA') return

  const target = event.target as HTMLElement | null
  const container = event.currentTarget as HTMLElement | null
  if (!target || !container) return
  if (target.closest(BLOCKED_CONTAINER_SELECTOR)) return

  const fields = getFocusableFields(container)
  if (!fields.length) return

  const currentIndex = fields.findIndex(
    (field) => field === target || field.contains(target) || target.contains(field)
  )
  if (currentIndex === -1) return

  const nextIndex = event.shiftKey ? currentIndex - 1 : currentIndex + 1
  const nextField = fields[nextIndex]
  if (!nextField) return

  event.preventDefault()
  nextField.focus()
  if (nextField instanceof HTMLInputElement || nextField instanceof HTMLTextAreaElement) {
    nextField.select()
  }
}
