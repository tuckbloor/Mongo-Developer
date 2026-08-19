import test from 'node:test'
import assert from 'node:assert/strict'
import { translateEloquent } from './eloquent.js'
const models=[{className:'Developer',collection:'developers'}]
test('Eloquent basic query',()=>assert.deepEqual(translateEloquent("Developer::where('active', true)->orderBy('name')->limit(25)->get();",{models}).pipeline,[{$match:{active:true}},{$sort:{name:1}},{$limit:25}]))
test('Eloquent first',()=>assert.deepEqual(translateEloquent("Developer::where('age','>=',18)->first();",{models}).pipeline,[{$match:{age:{$gte:18}}},{$limit:1}]))
test('Eloquent whereIn select count',()=>assert.deepEqual(translateEloquent("Developer::whereIn('role',['Developer','Lead'])->select('role')->count();",{models}).pipeline,[{$match:{role:{$in:['Developer','Lead']}}},{$project:{_id:1,role:1}},{$count:'count'}]))
test('Eloquent rejects arbitrary PHP',()=>assert.throws(()=>translateEloquent("Developer::raw(function () { return 1; });",{models})))
