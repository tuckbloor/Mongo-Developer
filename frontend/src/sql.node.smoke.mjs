import test from 'node:test'
import assert from 'node:assert/strict'
import { translateSQL } from './sql.js'

test('SELECT *', () => {
  assert.deepEqual(translateSQL('SELECT * FROM developers'), {database:'', collection:'developers', pipeline:[]})
})

test('WHERE ORDER BY LIMIT', () => {
  const out=translateSQL('SELECT name, age FROM developers d WHERE d.active = true AND d.age >= 18 ORDER BY d.name DESC LIMIT 25')
  assert.deepEqual(out.pipeline,[
    {$match:{active:true,age:{$gte:18}}},
    {$project:{_id:0,name:'$name',age:'$age'}},
    {$sort:{name:-1}},
    {$limit:25},
  ])
})

test('LEFT JOIN', () => {
  const out=translateSQL('SELECT d.name, t.name AS team FROM developers d LEFT JOIN teams t ON d.team_id = t._id')
  assert.deepEqual(out.pipeline,[
    {$lookup:{from:'teams',localField:'team_id',foreignField:'_id',as:'t'}},
    {$unwind:{path:'$t',preserveNullAndEmptyArrays:true}},
    {$project:{_id:0,name:'$name',team:'$t.name'}},
  ])
})

test('database-qualified source', () => {
  const out=translateSQL('SELECT * FROM test.developers')
  assert.equal(out.database,'test')
  assert.equal(out.collection,'developers')
})
