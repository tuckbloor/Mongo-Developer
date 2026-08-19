import { describe, expect, it } from 'vitest'
import { translateSQL } from './sql'

describe('translateSQL', () => {
  it('translates select all', () => {
    expect(translateSQL('SELECT * FROM developers')).toEqual({ database:'', collection:'developers', pipeline:[] })
  })

  it('translates where, order and limit', () => {
    const out = translateSQL("SELECT name, age FROM developers d WHERE d.active = true AND d.age >= 18 ORDER BY d.name DESC LIMIT 25")
    expect(out.collection).toBe('developers')
    expect(out.pipeline).toEqual([
      { $match: { active:true, age:{ $gte:18 } } },
      { $project: { _id:0, name:'$name', age:'$age' } },
      { $sort: { name:-1 } },
      { $limit:25 },
    ])
  })

  it('translates a left join to lookup and unwind', () => {
    const out = translateSQL('SELECT d.name, t.name AS team FROM developers d LEFT JOIN teams t ON d.team_id = t._id')
    expect(out.pipeline[0]).toEqual({ $lookup:{ from:'teams', localField:'team_id', foreignField:'_id', as:'t' } })
    expect(out.pipeline[1]).toEqual({ $unwind:{ path:'$t', preserveNullAndEmptyArrays:true } })
    expect(out.pipeline[2]).toEqual({ $project:{ _id:0, name:'$name', team:'$t.name' } })
  })

  it('accepts a database-qualified collection', () => {
    const out = translateSQL('SELECT * FROM test.developers')
    expect(out.database).toBe('test')
    expect(out.collection).toBe('developers')
  })
})
