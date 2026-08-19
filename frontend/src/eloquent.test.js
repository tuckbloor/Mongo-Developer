import { describe, expect, it } from 'vitest'
import { translateEloquent } from './eloquent'

describe('translateEloquent', () => {
  const models = [{className:'Developer', collection:'developers'}]

  it('translates a basic model query', () => {
    expect(translateEloquent("Developer::where('active', true)->orderBy('name')->limit(25)->get();", {models, defaultCollection:'developers'})).toMatchObject({
      collection:'developers',
      pipeline:[{$match:{active:true}}, {$sort:{name:1}}, {$limit:25}],
    })
  })

  it('supports operator where and first', () => {
    expect(translateEloquent("Developer::query()->where('age', '>=', 18)->first();", {models}).pipeline).toEqual([
      {$match:{age:{$gte:18}}}, {$limit:1},
    ])
  })

  it('supports whereIn, select and count', () => {
    expect(translateEloquent("Developer::whereIn('role', ['Developer','Lead'])->select('role')->count();", {models}).pipeline).toEqual([
      {$match:{role:{$in:['Developer','Lead']}}}, {$project:{_id:1,role:1}}, {$count:'count'},
    ])
  })

  it('rejects arbitrary PHP', () => {
    expect(() => translateEloquent("Developer::raw(function () { return 1; });", {models})).toThrow()
  })
})
