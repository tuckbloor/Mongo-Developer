export function parseEJSON(value) {
  try { return JSON.parse(value) } catch { return { _raw: value } }
}

export function shortValue(v) {
  if (v === null) return 'null'
  if (v === undefined) return ''
  if (typeof v === 'object') {
    if (v.$oid) return v.$oid
    if (v.$date) return typeof v.$date === 'object' ? String(v.$date.$numberLong || '') : String(v.$date)
    const s = JSON.stringify(v)
    return s.length > 80 ? s.slice(0, 77) + '…' : s
  }
  return String(v)
}

export function classNameFor(collection) {
  let s = (collection || '').trim()
  if (s.toLowerCase().endsWith('ies')) s = s.slice(0, -3) + 'y'
  else if (s.toLowerCase().endsWith('s') && s.length > 1) s = s.slice(0, -1)
  return s.split(/[^A-Za-z0-9]+/).filter(Boolean).map(x => x[0].toUpperCase() + x.slice(1)).join('') || 'MongoModel'
}

export function idFilter(doc) {
  if (!doc || doc._id === undefined) throw new Error('Document does not contain _id')
  return JSON.stringify({ _id: doc._id })
}

export function validateDocumentJSON(text) {
  try {
    const value = JSON.parse(String(text ?? ''))
    if (!value || Array.isArray(value) || typeof value !== 'object') {
      return 'Document JSON must be a JSON object, for example { \"name\": \"Chris\" }.'
    }
    return ''
  } catch (e) {
    return 'Invalid JSON: ' + String(e?.message || e)
  }
}
