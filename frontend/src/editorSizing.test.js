import { describe, expect, it } from 'vitest'
import { SQL_EDITOR_MIN_HEIGHT, sqlEditorAutoHeight } from './editorSizing'

describe('SQL editor auto sizing', () => {
  it('keeps short SQL at the five-line default', () => {
    expect(sqlEditorAutoHeight('SELECT *\nFROM developers\nLIMIT 50;', 1000)).toBe(SQL_EDITOR_MIN_HEIGHT)
  })

  it('grows for longer edited SQL', () => {
    const shortHeight = sqlEditorAutoHeight('SELECT *\nFROM developers\nLIMIT 50;', 1000)
    const longHeight = sqlEditorAutoHeight(Array.from({ length: 12 }, (_, i) => `line ${i + 1}`).join('\n'), 1000)
    expect(longHeight).toBeGreaterThan(shortHeight)
  })

  it('caps itself so results still have room', () => {
    const manyLines = Array.from({ length: 100 }, (_, i) => `line ${i + 1}`).join('\n')
    expect(sqlEditorAutoHeight(manyLines, 1200)).toBeLessThanOrEqual(480)
  })
})
