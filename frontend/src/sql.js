function stripIdentifier(value) {
  return String(value || '').trim().replace(/^`|`$/g, '')
}

function splitComma(input) {
  const parts = []
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
    if (ch === '"' || ch === "'") {
      quote = ch
      current += ch
      continue
    }
    if (ch === '(') depth++
    if (ch === ')') depth--
    if (ch === ',' && depth === 0) {
      parts.push(current.trim())
      current = ''
    } else {
      current += ch
    }
  }
  if (current.trim()) parts.push(current.trim())
  return parts
}

function parseLiteral(raw) {
  const value = raw.trim()
  if (/^null$/i.test(value)) return null
  if (/^true$/i.test(value)) return true
  if (/^false$/i.test(value)) return false
  if (/^-?\d+(?:\.\d+)?$/.test(value)) return Number(value)
  const quoted = value.match(/^(['"])([\s\S]*)\1$/)
  if (quoted) return quoted[2].replace(/\\(['"\\])/g, '$1')
  throw new Error(`Unsupported SQL value: ${value}. Use quoted strings, numbers, true, false or null.`)
}

function normaliseField(field, baseAlias) {
  let path = stripIdentifier(field)
  if (baseAlias && path.startsWith(baseAlias + '.')) path = path.slice(baseAlias.length + 1)
  return path
}

function parseWhere(whereText, baseAlias) {
  if (!whereText) return null
  const conditions = whereText.split(/\s+AND\s+/i).map(v => v.trim()).filter(Boolean)
  const match = {}
  for (const condition of conditions) {
    const m = condition.match(/^([A-Za-z_][\w.$]*(?:\.[A-Za-z_][\w$]*)*)\s*(=|!=|>=|<=|>|<)\s*(.+)$/i)
    if (!m) throw new Error(`Unsupported WHERE condition: ${condition}. V1.1 SQL mode currently supports AND with =, !=, >, >=, < and <=.`)
    const field = normaliseField(m[1], baseAlias)
    const op = m[2]
    const value = parseLiteral(m[3])
    if (op === '=') match[field] = value
    else {
      const mongoOp = { '!=': '$ne', '>': '$gt', '>=': '$gte', '<': '$lt', '<=': '$lte' }[op]
      match[field] = { ...(match[field] && typeof match[field] === 'object' ? match[field] : {}), [mongoOp]: value }
    }
  }
  return match
}

function parseSource(source) {
  const bits = source.trim().split(/\s+/)
  if (bits.length > 3) throw new Error(`Invalid FROM source: ${source}`)
  let qualified = stripIdentifier(bits[0])
  let alias = ''
  if (bits.length === 2) alias = /^AS$/i.test(bits[1]) ? '' : stripIdentifier(bits[1])
  if (bits.length === 3) {
    if (!/^AS$/i.test(bits[1])) throw new Error(`Invalid FROM alias: ${source}`)
    alias = stripIdentifier(bits[2])
  }
  let database = ''
  let collection = qualified
  const dot = qualified.indexOf('.')
  if (dot > 0) {
    database = qualified.slice(0, dot)
    collection = qualified.slice(dot + 1)
  }
  if (!/^[A-Za-z_][\w.-]*$/.test(collection)) throw new Error(`Invalid collection name: ${collection}`)
  return { database, collection, alias }
}

function projectionFor(selectText, baseAlias) {
  const fields = splitComma(selectText)
  if (fields.length === 1 && fields[0] === '*') return null
  const project = { _id: 0 }
  for (const item of fields) {
    const m = item.match(/^([A-Za-z_][\w.$]*(?:\.[A-Za-z_][\w$]*)*|\*)\s*(?:AS\s+([A-Za-z_][\w$]*))?$/i)
    if (!m) throw new Error(`Unsupported SELECT expression: ${item}. Use field names, optional AS aliases, or *.`)
    if (m[1] === '*') throw new Error('SELECT * cannot be mixed with named fields in SQL mode.')
    const path = normaliseField(m[1], baseAlias)
    const output = m[2] || path.replace(/\./g, '_')
    project[output] = '$' + path
  }
  return project
}

function parseJoins(joinText, baseAlias) {
  const stages = []
  if (!joinText?.trim()) return stages
  const joinRegex = /\b(LEFT\s+JOIN|INNER\s+JOIN|JOIN)\s+([A-Za-z_][\w.-]*)(?:\s+(?:AS\s+)?([A-Za-z_][\w$]*))?\s+ON\s+([A-Za-z_][\w$]*(?:\.[A-Za-z_][\w$]*)+)\s*=\s*([A-Za-z_][\w$]*(?:\.[A-Za-z_][\w$]*)+)/ig
  let consumed = 0
  let match
  while ((match = joinRegex.exec(joinText)) !== null) {
    if (joinText.slice(consumed, match.index).trim()) throw new Error(`Unsupported JOIN syntax near: ${joinText.slice(consumed, match.index).trim()}`)
    consumed = joinRegex.lastIndex
    const joinType = match[1].toUpperCase()
    const from = stripIdentifier(match[2])
    const alias = stripIdentifier(match[3] || from)
    const leftRaw = match[4]
    const rightRaw = match[5]
    const leftParts = leftRaw.split('.')
    const rightParts = rightRaw.split('.')
    if (leftParts.length < 2 || rightParts.length < 2) throw new Error('JOIN fields must be qualified, e.g. developers.team_id = teams._id.')

    let localField
    let foreignField
    if (rightParts[0] === alias || rightParts[0] === from) {
      localField = normaliseField(leftRaw, baseAlias)
      foreignField = rightParts.slice(1).join('.')
    } else if (leftParts[0] === alias || leftParts[0] === from) {
      localField = normaliseField(rightRaw, baseAlias)
      foreignField = leftParts.slice(1).join('.')
    } else {
      throw new Error(`JOIN must compare the base collection with ${from}.`)
    }
    stages.push({ $lookup: { from, localField, foreignField, as: alias } })
    stages.push({ $unwind: { path: '$' + alias, preserveNullAndEmptyArrays: joinType.startsWith('LEFT') } })
  }
  if (joinText.slice(consumed).trim()) throw new Error(`Unsupported JOIN syntax near: ${joinText.slice(consumed).trim()}`)
  return stages
}

export function translateSQL(sql, defaultCollection = '') {
  const source = String(sql || '').trim().replace(/;\s*$/, '')
  if (!source) throw new Error('Enter a SELECT query.')
  const selectMatch = source.match(/^SELECT\s+([\s\S]+?)\s+FROM\s+([\s\S]+)$/i)
  if (!selectMatch) throw new Error('SQL mode currently supports SELECT queries only.')

  const selectText = selectMatch[1].trim()
  let tail = selectMatch[2].trim()

  const clauseMatch = tail.match(/\s+(LEFT\s+JOIN|INNER\s+JOIN|JOIN|WHERE|ORDER\s+BY|LIMIT)\s+/i)
  let fromText = tail
  let clauses = ''
  if (clauseMatch) {
    fromText = tail.slice(0, clauseMatch.index).trim()
    clauses = tail.slice(clauseMatch.index).trim()
  }
  const from = parseSource(fromText || defaultCollection)
  if (!from.collection) from.collection = defaultCollection
  if (!from.collection) throw new Error('SQL query must name a collection after FROM.')
  const baseAlias = from.alias || from.collection

  let limit = null
  const limitMatch = clauses.match(/(?:^|\s)LIMIT\s+(\d+)\s*$/i)
  if (limitMatch) {
    limit = Math.min(Math.max(Number(limitMatch[1]), 1), 500)
    clauses = clauses.slice(0, limitMatch.index).trim()
  }

  let orderText = ''
  const orderMatch = clauses.match(/(?:^|\s)ORDER\s+BY\s+([\s\S]+)$/i)
  if (orderMatch) {
    orderText = orderMatch[1].trim()
    clauses = clauses.slice(0, orderMatch.index).trim()
  }

  let whereText = ''
  const whereMatch = clauses.match(/(?:^|\s)WHERE\s+([\s\S]+)$/i)
  if (whereMatch) {
    whereText = whereMatch[1].trim()
    clauses = clauses.slice(0, whereMatch.index).trim()
  }

  const stages = []
  stages.push(...parseJoins(clauses, baseAlias))
  const matchDoc = parseWhere(whereText, baseAlias)
  if (matchDoc && Object.keys(matchDoc).length) stages.push({ $match: matchDoc })

  const project = projectionFor(selectText, baseAlias)
  if (project) stages.push({ $project: project })

  if (orderText) {
    const sort = {}
    for (const item of splitComma(orderText)) {
      const m = item.match(/^([A-Za-z_][\w.$]*(?:\.[A-Za-z_][\w$]*)*)(?:\s+(ASC|DESC))?$/i)
      if (!m) throw new Error(`Unsupported ORDER BY expression: ${item}`)
      sort[normaliseField(m[1], baseAlias)] = String(m[2] || 'ASC').toUpperCase() === 'DESC' ? -1 : 1
    }
    stages.push({ $sort: sort })
  }
  if (limit !== null) stages.push({ $limit: limit })

  return { database: from.database, collection: from.collection, pipeline: stages }
}
