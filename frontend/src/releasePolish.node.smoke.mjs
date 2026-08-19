import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const frontend = path.resolve(here, '..')
const root = path.resolve(frontend, '..')
const app = fs.readFileSync(path.join(here, 'App.vue'), 'utf8')

test('visible release version is v1.5.0', () => {
  assert.match(app, /<span class="version">v1\.5\.0<\/span>/)
})

test('saved query controls and persistence are wired', () => {
  assert.match(app, /Saved Queries/)
  assert.match(app, /mongo-developer-saved-queries/)
  assert.match(app, /saveCurrentQuery/)
  assert.match(app, /loadSavedQuery/)
})

test('explain and index advice are wired', () => {
  assert.match(app, /Explain Query/)
  assert.match(app, /ExplainFind/)
  assert.match(app, /ExplainAggregation/)
  assert.match(app, /Suggested index/)
  assert.match(app, /Review in Create Index/)
})

test('diagnostics and reconnect are wired', () => {
  assert.match(app, /GetDiagnostics/)
  assert.match(app, /Copy Diagnostics/)
  assert.match(app, /reconnectActive/)
  assert.match(app, /GetConnectionHealth/)
})

test('unsaved query or document guard is wired', () => {
  assert.match(app, /hasUnsavedWork/)
  assert.match(app, /Discard unsaved changes/)
  assert.match(app, /guardUnsaved\(`open \$\{db\}\.\$\{coll\}`/)
})

test('SQL Run remains runnable after mode switches', () => {
  assert.doesNotMatch(app, /:disabled="[^\"]*sqlDirty[^\"]*"/)
  assert.match(app, /queryMode==='sql'\?\(sqlDirty\?'Run Changes':'Run'\)/)
})

test('JOIN example button stays removed while query engines remain', () => {
  assert.doesNotMatch(app, />JOIN Example</i)
  assert.match(fs.readFileSync(path.join(here, 'sql.js'), 'utf8'), /JOIN/i)
})


test('read-only aggregation write stages are blocked front and back', () => {
  assert.match(app, /read-only; \$out and \$merge aggregation stages are disabled/)
  const backend = fs.readFileSync(path.join(root, 'app.go'), 'utf8')
  assert.match(backend, /AggregationWriteStage/)
  assert.match(backend, /requireWritable/)
})

test('Windows metadata JSON is BOM-free and versioned', () => {
  const infoPath = path.join(root, 'build', 'windows', 'info.json')
  const bytes = fs.readFileSync(infoPath)
  assert.equal(bytes[0], 123)
  const info = JSON.parse(bytes.toString('utf8'))
  assert.equal(info.fixed.file_version, '1.5.0.0')
  assert.equal(info.info['0409'].ProductVersion, '1.5.0')
})
