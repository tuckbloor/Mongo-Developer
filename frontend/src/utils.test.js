import { describe, expect, it } from 'vitest'
import { classNameFor, idFilter, validateDocumentJSON } from './utils'

describe('utils', () => {
  it('builds Laravel model names', () => {
    expect(classNameFor('players')).toBe('Player')
    expect(classNameFor('categories')).toBe('Category')
  })

  it('keeps extended JSON ids', () => {
    expect(idFilter({_id: {$oid: 'abc'}})).toBe('{"_id":{"$oid":"abc"}}')
  })

  it('validates document JSON before save', () => {
    expect(validateDocumentJSON('{"name":"Chris"}')).toBe('')
    expect(validateDocumentJSON('{')).toContain('Invalid JSON')
    expect(validateDocumentJSON('[]')).toContain('JSON object')
  })
})
