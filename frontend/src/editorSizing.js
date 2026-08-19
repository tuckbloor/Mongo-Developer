export const SQL_EDITOR_MIN_LINES = 5
export const SQL_EDITOR_MIN_HEIGHT = 123
export const SQL_EDITOR_MAX_HEIGHT = 480

export function sqlEditorAutoHeight(sql, viewportHeight = 900) {
  const lineCount = String(sql ?? '').split(/\r?\n/).length
  const visibleLines = Math.max(SQL_EDITOR_MIN_LINES, lineCount)
  const lineHeight = 19
  const verticalPaddingAndBorder = 28
  const contentHeight = verticalPaddingAndBorder + visibleLines * lineHeight
  const viewportMax = Math.max(180, Math.min(SQL_EDITOR_MAX_HEIGHT, Number(viewportHeight || 900) - 430))

  return Math.max(
    SQL_EDITOR_MIN_HEIGHT,
    Math.min(viewportMax, contentHeight),
  )
}
