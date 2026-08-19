import test from 'node:test'
import assert from 'node:assert/strict'
import { validateDocumentJSON } from './utils.js'

test('valid document JSON', () => {
  assert.equal(validateDocumentJSON('{"name":"Chris","active":true}'), '')
})

test('invalid JSON is rejected', () => {
  assert.match(validateDocumentJSON('{'), /^Invalid JSON:/)
})

test('non-object JSON roots are rejected', () => {
  assert.match(validateDocumentJSON('[]'), /JSON object/)
  assert.match(validateDocumentJSON('null'), /JSON object/)
})
