function splitTopLevel(input, delimiter = ',') {
  const out = []
  let current = ''
  let quote = null
  let depth = 0
  for (let i = 0; i < input.length; i++) {
    const ch = input[i]
    if (quote) {
      current += ch
      if (ch === quote && input[i - 1] !== '\\') quote = null
      continue
    }
    if (ch === "'" || ch === '"') {
      quote = ch
      current += ch
      continue
    }
    if (ch === '(' || ch === '[') depth++
    if (ch === ')' || ch === ']') depth--
    if (ch === delimiter && depth === 0) {
      out.push(current.trim())
      current = ''
    } else {
      current += ch
    }
  }
  if (current.trim()) out.push(current.trim())
  return out
}

function parseValue(raw) {
  const value = String(raw || '').trim()
  if (/^null$/i.test(value)) return null
  if (/^true$/i.test(value)) return true
  if (/^false$/i.test(value)) return false
  if (/^-?\d+(?:\.\d+)?$/.test(value)) return Number(value)
  const quoted = value.match(/^(['"])([\s\S]*)\1$/)
  if (quoted) return quoted[2].replace(/\\(['"\\])/g, '$1')
  if (value.startsWith('[') && value.endsWith(']')) {
    const body = value.slice(1, -1).trim()
    return body ? splitTopLevel(body).map(parseValue) : []
  }
  throw new Error(`Unsupported Eloquent value: ${value}. Use strings, numbers, booleans, null, or simple arrays.`)
}

function parseArgs(raw) {
  const body = String(raw || '').trim()
  return body ? splitTopLevel(body).map(v => v.trim()) : []
}

function condition(fieldRaw, opRaw, valueRaw) {
  const field = parseValue(fieldRaw)
  if (typeof field !== 'string' || !field) throw new Error('Eloquent field names must be quoted strings.')
  const op = String(opRaw || '=').toLowerCase()
  const value = parseValue(valueRaw)
  if (op === '=' || op === '==') return { [field]: value }
  const mongo = { '!=':'$ne', '<>':'$ne', '>':'$gt', '>=':'$gte', '<':'$lt', '<=':'$lte' }[op]
  if (!mongo) throw new Error(`Unsupported Eloquent operator: ${op}`)
  return { [field]: { [mongo]: value } }
}

function mergeMatch(target, piece) {
  for (const [field, value] of Object.entries(piece)) {
    if (target[field] === undefined) target[field] = value
    else {
      const existing = target[field]
      if (existing && typeof existing === 'object' && !Array.isArray(existing) && value && typeof value === 'object' && !Array.isArray(value)) {
        target[field] = { ...existing, ...value }
      } else {
        throw new Error(`Multiple incompatible where() clauses for ${field} are not supported in Eloquent mode yet.`)
      }
    }
  }
}

function parseChain(source) {
  const trimmed = String(source || '').trim().replace(/;\s*$/, '')
  const first = trimmed.match(/^\\?([A-Za-z_][\w\\]*)::([A-Za-z_][\w]*)\s*\(([^)]*)\)([\s\S]*)$/)
  if (!first) throw new Error('Start with an Eloquent model query, for example Developer::query()->limit(50)->get().')
  const classPath = first[1]
  const modelClass = classPath.split('\\').pop()
  const calls = [{ name:first[2], args:first[3] }]
  let rest = first[4].trim()
  const re = /^->\s*([A-Za-z_][\w]*)\s*\(/ 
  while (rest) {
    const head = rest.match(re)
    if (!head) throw new Error(`Unsupported Eloquent syntax near: ${rest.slice(0, 60)}`)
    let i = head[0].length
    let quote = null
    let depth = 1
    let args = ''
    for (; i < rest.length; i++) {
      const ch = rest[i]
      if (quote) {
        args += ch
        if (ch === quote && rest[i - 1] !== '\\') quote = null
        continue
      }
      if (ch === "'" || ch === '"') { quote = ch; args += ch; continue }
      if (ch === '(') { depth++; args += ch; continue }
      if (ch === ')') {
        depth--
        if (depth === 0) { i++; break }
        args += ch
        continue
      }
      args += ch
    }
    if (depth !== 0) throw new Error(`Unclosed ${head[1]}() call.`)
    calls.push({ name:head[1], args })
    rest = rest.slice(i).trim()
  }
  return { modelClass, calls }
}

function collectionFor(modelClass, models, defaultCollection) {
  const list = Array.isArray(models) ? models : []
  const found = list.find(m => String(m.className || '').toLowerCase() === modelClass.toLowerCase())
  if (found?.collection) return { collection:found.collection, model:found }
  if (defaultCollection) return { collection:defaultCollection, model:found || null }
  throw new Error(`Could not map Laravel model ${modelClass} to a MongoDB collection.`)
}

export function translateEloquent(source, options = {}) {
  const { modelClass, calls } = parseChain(source)
  const mapped = collectionFor(modelClass, options.models, options.defaultCollection)
  const match = {}
  const stages = []
  let sort = null
  let skip = null
  let limit = null
  let project = null
  let terminal = 'get'

  for (const call of calls) {
    const name = call.name.toLowerCase()
    const args = parseArgs(call.args)
    if (name === 'query') {
      if (args.length) throw new Error('query() does not take arguments here.')
      continue
    }
    if (name === 'where') {
      if (args.length === 2) mergeMatch(match, condition(args[0], '=', args[1]))
      else if (args.length === 3) mergeMatch(match, condition(args[0], parseValue(args[1]), args[2]))
      else throw new Error('where() expects 2 or 3 arguments.')
      continue
    }
    if (name === 'wherein' || name === 'wherenotin') {
      if (args.length !== 2) throw new Error(`${call.name}() expects field and array.`)
      const field = parseValue(args[0]); const values = parseValue(args[1])
      if (typeof field !== 'string' || !Array.isArray(values)) throw new Error(`${call.name}() expects a quoted field and simple array.`)
      mergeMatch(match, { [field]: { [name === 'wherein' ? '$in' : '$nin']: values } })
      continue
    }
    if (name === 'wherenull' || name === 'wherenotnull') {
      if (args.length !== 1) throw new Error(`${call.name}() expects one field.`)
      const field = parseValue(args[0]); if (typeof field !== 'string') throw new Error(`${call.name}() field must be quoted.`)
      mergeMatch(match, name === 'wherenull' ? { [field]: null } : { [field]: { $ne:null } })
      continue
    }
    if (name === 'orderby') {
      if (args.length < 1 || args.length > 2) throw new Error('orderBy() expects field and optional direction.')
      const field = parseValue(args[0]); const dir = args[1] ? String(parseValue(args[1])).toLowerCase() : 'asc'
      if (typeof field !== 'string' || !['asc','desc'].includes(dir)) throw new Error('orderBy() direction must be asc or desc.')
      sort = { ...(sort || {}), [field]: dir === 'desc' ? -1 : 1 }
      continue
    }
    if (name === 'latest' || name === 'oldest') {
      const field = args.length ? parseValue(args[0]) : 'created_at'
      if (typeof field !== 'string') throw new Error(`${call.name}() field must be quoted.`)
      sort = { ...(sort || {}), [field]: name === 'latest' ? -1 : 1 }
      continue
    }
    if (name === 'limit' || name === 'take') {
      if (args.length !== 1) throw new Error(`${call.name}() expects one number.`)
      limit = Math.min(Math.max(Number(parseValue(args[0])), 1), 500)
      continue
    }
    if (name === 'skip' || name === 'offset') {
      if (args.length !== 1) throw new Error(`${call.name}() expects one number.`)
      skip = Math.max(Number(parseValue(args[0])), 0)
      continue
    }
    if (name === 'select') {
      const fields = args.length === 1 && String(args[0]).trim().startsWith('[') ? parseValue(args[0]) : args.map(parseValue)
      if (!Array.isArray(fields) || fields.some(v => typeof v !== 'string')) throw new Error('select() expects quoted field names or an array of field names.')
      project = { _id:1 }
      for (const field of fields) project[field] = 1
      continue
    }
    if (['get','first','count'].includes(name)) {
      terminal = name
      if (args.length) throw new Error(`${call.name}() does not take arguments in this preview.`)
      continue
    }
    throw new Error(`Eloquent method ${call.name}() is not supported yet. Supported: query, where, whereIn, whereNotIn, whereNull, whereNotNull, orderBy, latest, oldest, limit/take, skip/offset, select, get, first, count.`)
  }

  if (Object.keys(match).length) stages.push({ $match:match })
  if (project) stages.push({ $project:project })
  if (sort) stages.push({ $sort:sort })
  if (skip !== null) stages.push({ $skip:skip })
  if (terminal === 'first') stages.push({ $limit:1 })
  else if (limit !== null) stages.push({ $limit:limit })
  if (terminal === 'count') stages.push({ $count:'count' })

  return { database:'', collection:mapped.collection, pipeline:stages, modelClass, model:mapped.model, terminal }
}
