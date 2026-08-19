import assert from 'node:assert/strict'
import test from 'node:test'
import { SQL_EDITOR_MIN_HEIGHT, sqlEditorAutoHeight } from './editorSizing.js'

test('short SQL uses the compact five-line height', () => {
  assert.equal(sqlEditorAutoHeight('SELECT *\nFROM developers\nLIMIT 50;', 1000), SQL_EDITOR_MIN_HEIGHT)
})

test('long SQL grows automatically and stays capped', () => {
  const shortHeight = sqlEditorAutoHeight('SELECT *\nFROM developers\nLIMIT 50;', 1000)
  const longHeight = sqlEditorAutoHeight(Array.from({ length: 12 }, (_, i) => `line ${i + 1}`).join('\n'), 1000)
  assert.ok(longHeight > shortHeight)
  const hugeHeight = sqlEditorAutoHeight(Array.from({ length: 100 }, (_, i) => `line ${i + 1}`).join('\n'), 1200)
  assert.ok(hugeHeight <= 480)
})
