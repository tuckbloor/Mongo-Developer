<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import {
  Database, Plug, Plus, RefreshCw, Search, Play, ChevronRight, ChevronDown, ChevronLeft,
  Table2, Waypoints, Layers3, Code2, ShieldCheck,
  Shield, Upload, Download, FilePlus2, Trash2, Pencil, FolderOpen, CheckCircle2,
  XCircle, LoaderCircle, KeyRound, Server, Circle, Save, X, Copy, Boxes, ListFilter
} from '@lucide/vue'
import { call } from './api'
import { classNameFor, idFilter, parseEJSON, shortValue, validateDocumentJSON } from './utils'
import { translateSQL } from './sql'
import { translateEloquent } from './eloquent'
import { sqlEditorAutoHeight } from './editorSizing'

const connections = ref([])
const connectionsWidth = ref(158)
const databasesWidth = ref(165)
const sidebarResize = reactive({ active:false, side:'', startX:0, startWidth:0 })
const searchText = ref('')
const searchInput = ref(null)
const activeConnection = ref(null)
const connected = reactive({})
const databases = ref([])
const collections = reactive({})
const openDatabases = reactive({})
const activeDatabase = ref('')
const activeCollection = ref('')
const docsRaw = ref([])
const docs = computed(() => docsRaw.value.map(parseEJSON))
const total = ref(0)
const queryMs = ref(0)
const skip = ref(0)
const limit = ref(50)
const query = ref('{}')
const sort = ref('{}')
const selectedRow = ref(-1)
const stats = ref({ documents: 0, indexes: 0, avgDocumentKb: 0, totalSizeMb: 0, storageSizeMb: 0 })
const indexes = ref([])
const aggregation = ref('[\n  { "$match": {} }\n]')
const aggregationResults = ref([])
const sqlQuery = ref('SELECT * FROM collection_name LIMIT 50')
const eloquentQuery = ref('Model::query()\n    ->limit(50)\n    ->get();')
const lastExecutedSQL = ref('')
const lastResultMode = ref('')
const queryMode = ref('find')
const resultView = ref('table')
const queryHistory = ref([])
const historySelection = ref('')
const editorHeight = ref(123)
const queryExpanded = ref(false)
const resultsExpanded = ref(false)
const tab = ref('Documents')
const busy = ref(false)
const error = ref('')
const notice = ref('')
const dummyReport = ref(null)
const connectionModal = ref(false)
const documentModal = ref(false)
const editingDocument = ref(false)
const documentText = ref('{\n  \n}')
const documentDrawer = ref(false)
const drawerTab = ref('fields')
const drawerDocument = ref(null)
const drawerOriginal = ref(null)
const drawerDocumentText = ref('')
const hiddenColumns = ref([])
const columnsMenuOpen = ref(false)
const columnWidths = reactive({})
const inlineEdit = reactive({ row:-1, field:'', value:'', originalType:'' })
const indexModal = ref(false)
const indexForm = reactive({ keys: '{ "field": 1 }', name: '', unique: false })
const createModal = ref(false)
const createMode = ref('database')
const createName = ref('')
const laravel = ref(null)
const setupHelp = ref('')
const modelName = ref('')
const laravelLinkEpoch = ref(0)
const confirmModal = ref(false)
const confirmTitle = ref('Confirm')
const confirmMessage = ref('')
const backupModal = ref(false)
const backups = ref([])
const backupDirectory = ref('')
const backupDatabase = ref('')
const savedQueriesModal = ref(false)
const savedQueries = ref([])
const savedQueryName = ref('')
const explainModal = ref(false)
const explainResult = ref(null)
const diagnosticsModal = ref(false)
const diagnostics = ref(null)
const connectionHealth = ref(null)
const editorInput = ref(null)
const dirtyModes = reactive({ find:false, aggregation:false, sql:false, eloquent:false })
const confirmDanger = ref(true)
const confirmActionLabel = ref('Confirm destructive action')
let confirmHandler = null
let resizeStartY = 0
let resizeStartHeight = 0

const form = reactive({ id:'', name:'', uri:'mongodb://localhost:27017/', host:'localhost', port:27017, username:'', password:'', authDb:'admin', tls:false, readOnly:false })

const activeProfile = computed(() => connections.value.find(c => c.id === activeConnection.value) || null)
const filteredConnections = computed(() => {
  const q=searchText.value.trim().toLowerCase(); if(!q)return connections.value
  return connections.value.filter(c => [c.name,c.uri,c.username].some(v=>String(v||'').toLowerCase().includes(q)))
})
const filteredDatabases = computed(() => {
  const q=searchText.value.trim().toLowerCase(); if(!q)return databases.value
  return databases.value.filter(db => db.name.toLowerCase().includes(q) || (collections[db.name]||[]).some(c=>c.name.toLowerCase().includes(q)))
})
function visibleCollections(db){ const q=searchText.value.trim().toLowerCase(); const list=collections[db]||[]; return q ? list.filter(c=>c.name.toLowerCase().includes(q)||db.toLowerCase().includes(q)) : list }
const readOnly = computed(() => !!activeProfile.value?.readOnly)
const dummyInstalled = computed(() => databases.value.some(db => db.name === 'dummy_test_data'))
const backupDatabases = computed(() => databases.value.filter(db => !['admin','config','local'].includes(String(db.name || '').toLowerCase())))
const bodyGridStyle = computed(() => ({ '--connections-width': `${connectionsWidth.value}px`, '--databases-width': `${databasesWidth.value}px` }))
const linkedLaravelModel = computed(() => {
  const list = Array.isArray(laravel.value?.models) ? laravel.value.models : []
  return list.find(m => String(m.collection || '').toLowerCase() === String(activeCollection.value || '').toLowerCase()) || null
})
const editorText = computed({
  get(){ return queryMode.value==='find' ? query.value : queryMode.value==='aggregation' ? aggregation.value : queryMode.value==='sql' ? sqlQuery.value : eloquentQuery.value },
  set(value){ if(queryMode.value==='find') query.value=value; else if(queryMode.value==='aggregation') aggregation.value=value; else if(queryMode.value==='sql') sqlQuery.value=value; else eloquentQuery.value=value }
})
const currentRawResults = computed(() => queryMode.value==='find' ? docsRaw.value : aggregationResults.value)
const currentDocs = computed(() => currentRawResults.value.map(parseEJSON))
const columns = computed(() => {
  const seen = []
  for (const d of currentDocs.value.slice(0, 40)) for (const k of Object.keys(d || {})) if (!seen.includes(k)) seen.push(k)
  return seen.slice(0, 20)
})
const displayColumns = computed(() => columns.value.filter(c => !hiddenColumns.value.includes(c)))
const sqlDirty = computed(() => queryMode.value==='sql' && sqlQuery.value.trim() !== lastExecutedSQL.value.trim())
const queryDirty = computed(() => !!dirtyModes[queryMode.value])
const drawerDirty = computed(() => {
  if(!documentDrawer.value || !drawerEditable.value || !drawerOriginal.value)return false
  if(drawerTab.value==='json') return drawerDocumentText.value.trim() !== JSON.stringify(drawerOriginal.value,null,2).trim()
  return JSON.stringify(drawerDocument.value) !== JSON.stringify(drawerOriginal.value)
})
const hasUnsavedWork = computed(() => Object.values(dirtyModes).some(Boolean) || drawerDirty.value || inlineEdit.row>=0 || (documentModal.value && documentText.value.trim() !== '{\n  \n}'))
const drawerEditable = computed(() => !!drawerOriginal.value && drawerOriginal.value._id !== undefined && queryMode.value!=='aggregation' && !readOnly.value)
const drawerFields = computed(() => Object.keys(drawerDocument.value || {}))
const pageStart = computed(() => queryMode.value==='find' && total.value ? skip.value + 1 : currentDocs.value.length ? 1 : 0)
const pageEnd = computed(() => queryMode.value==='find' ? Math.min(skip.value + docs.value.length, total.value) : currentDocs.value.length)
const resultTotal = computed(() => queryMode.value==='find' ? total.value : currentDocs.value.length)
const editorLabel = computed(() => queryMode.value==='find' ? 'Filter JSON / EJSON' : queryMode.value==='aggregation' ? 'Aggregation Pipeline' : queryMode.value==='sql' ? 'SQL Query' : 'Laravel Eloquent Preview')
const editorHelp = computed(() => {
  if(queryMode.value==='find') return 'Filter the selected collection. Use {} to return all documents.'
  if(queryMode.value==='aggregation') return 'Run a native MongoDB aggregation pipeline. $lookup is the MongoDB join stage.'
  if(queryMode.value==='sql') return 'SQL convenience mode. FROM names a collection (or database.collection); SELECT, WHERE, ORDER BY, LIMIT and straightforward JOINs are translated to aggregation.'
  return 'Laravel Eloquent preview. Common read chains are translated to MongoDB; arbitrary PHP is never executed.'
})
const editorError = computed(() => {
  try {
    const text = editorText.value.trim()
    if(queryMode.value==='find') {
      const value = JSON.parse(text || '{}')
      if(!value || Array.isArray(value) || typeof value!=='object') return 'Find mode requires a JSON object, for example { "active": true }.'
    } else if(queryMode.value==='aggregation') {
      const value = JSON.parse(text || '[]')
      if(!Array.isArray(value)) return 'Aggregation mode requires a JSON array of pipeline stages.'
      if(readOnly.value && value.some(stage=>stage&&typeof stage==='object'&&('$out' in stage||'$merge' in stage))) return 'This connection is read-only; $out and $merge aggregation stages are disabled.'
    } else if(queryMode.value==='sql') {
      translateSQL(text, activeCollection.value)
    } else {
      translateEloquent(text, { models:laravel.value?.models || [], defaultCollection:activeCollection.value })
    }
    if(queryMode.value==='find' && sort.value.trim()) {
      const sortValue = JSON.parse(sort.value)
      if(!sortValue || Array.isArray(sortValue) || typeof sortValue!=='object') return 'Sort must be a JSON object.'
    }
    return ''
  } catch(e) {
    return queryValidationMessage(e, editorText.value)
  }
})
const documentJsonError = computed(() => validateDocumentJSON(documentText.value))
const drawerJsonError = computed(() => drawerTab.value === 'json' ? validateDocumentJSON(drawerDocumentText.value) : '')
const canEditResult = computed(() => selectedRow.value>=0 && queryMode.value!=='aggregation' && !readOnly.value && currentDocs.value[selectedRow.value]?._id !== undefined)
const diagnosticsText = computed(() => {
  if(!diagnostics.value)return ''
  const d=diagnostics.value
  const lines=[
    `Mongo Developer ${d.version||'1.5.0'}`,
    `Checked: ${d.checkedAt||'—'}`,
    `Platform: ${d.os||'—'}/${d.arch||'—'}`,
    `Go: ${d.goVersion||'—'}`,
    `Connection: ${d.connectionName||'none'}`,
    `URI: ${d.connectionUri||'—'}`,
    `Status: ${d.connected?'connected':'not connected'}${d.pingMs>=0?` · ${d.pingMs} ms`:''}`,
    `MongoDB: ${d.mongoDbVersion||'—'}`,
    `Access: ${d.readOnly?'read only':'read/write'}`,
    `Laravel: ${laravel.value?.path||'not linked'}`,
    `Laravel MongoDB package: ${laravel.value?.mongoPackageInstalled?'installed':'not detected'}`,
    `Config: ${d.configDirectory||'—'}`,
    `Backups: ${d.backupDirectory||'—'}`,
    `WebView: ${navigator.userAgent}`,
  ]
  if(d.connectionError)lines.push(`Connection error: ${d.connectionError}`)
  return lines.join('\n')
})
const resultSummary = computed(() => {
  if(resultTotal.value===0) return '0 documents matched'
  if(queryMode.value==='find') return `${pageStart.value}–${pageEnd.value} of ${resultTotal.value}`
  return `${resultTotal.value} documents returned`
})

function queryValidationMessage(e,text){
  const message=String(e?.message||e)
  const match=message.match(/position\s+(\d+)/i)
  if(!match)return message
  const pos=Math.max(0,Number(match[1])||0);const before=String(text||'').slice(0,pos);const line=before.split('\n').length;const column=pos-before.lastIndexOf('\n')
  return `${message} · line ${line}, column ${column}`
}

function askConfirm(title, message, handler, options={}) {
  confirmTitle.value=title; confirmMessage.value=message; confirmHandler=handler; confirmDanger.value=options.danger!==false; confirmActionLabel.value=options.label || (confirmDanger.value?'Confirm destructive action':'Continue'); confirmModal.value=true
}
async function confirmYes() { const fn=confirmHandler; confirmModal.value=false; confirmHandler=null; if(fn) await fn() }
function flash(msg) { notice.value = msg; setTimeout(() => { if (notice.value === msg) notice.value = '' }, 3500) }
function fail(e) { error.value = String(e?.message || e); setTimeout(() => { if (error.value === String(e?.message || e)) error.value = '' }, 9000) }
async function copyError(){ if(!error.value)return; await navigator.clipboard.writeText(error.value); flash('Error copied') }
function markQueryDirty(){ dirtyModes[queryMode.value]=true }
function clearDirtyModes(){ for(const mode of Object.keys(dirtyModes))dirtyModes[mode]=false }
function discardUnsavedWork(){ clearDirtyModes(); documentDrawer.value=false; documentModal.value=false; cancelInlineEdit() }
function guardUnsaved(label, action){
  if(!hasUnsavedWork.value){ action(); return }
  askConfirm('Discard unsaved changes',`You have an edited query or document that has not been saved/run. Discard those changes and ${label}?`,async()=>{discardUnsavedWork();await action()},{danger:false,label:'Discard changes'})
}
async function run(fn) { busy.value = true; error.value = ''; try { return await fn() } catch (e) { fail(e); throw e } finally { busy.value = false } }
async function attempt(fn) { try { return { ok: true, value: await run(fn) } } catch { return { ok: false, value: null } } }

async function loadConnections() { connections.value = await call('GetConnections') || [] }
function isLocalProfile(profile) {
  const uri = String(profile?.uri || '').toLowerCase()
  return uri.includes('localhost') || uri.includes('127.0.0.1') || uri.includes('[::1]')
}
function dummySeedMarker(id) { return `mongo-developer-dummy-seeded:${id}` }
async function performSelectConnection(id) {
  connectionHealth.value=null
  await run(async () => {
    activeConnection.value = id; activeDatabase.value=''; activeCollection.value=''; docsRaw.value=[]; aggregationResults.value=[]; selectedRow.value=-1; connected[id]=false; dummyReport.value=null
    await call('Connect', id); connected[id] = true
    databases.value = await call('ListDatabases', id) || []

    const profile = connections.value.find(c => c.id === id)
    if (!profile?.readOnly) {
      try {
        const exists = databases.value.some(db => db.name === 'dummy_test_data')
        let firstLocalSeed = false
        try { firstLocalSeed = isLocalProfile(profile) && localStorage.getItem(dummySeedMarker(id)) !== '1' } catch {}

        if (firstLocalSeed) {
          const report = await call('EnsureDummyTestData', id)
          dummyReport.value = report
          databases.value = await call('ListDatabases', id) || []
          try { localStorage.setItem(dummySeedMarker(id), '1') } catch {}
          if (report?.created) flash(`Practice database installed: ${report.collections} collections, ${report.documents} documents`)
        } else if (exists) {
          dummyReport.value = await call('DummyTestDataStatus', id)
        }
      } catch (e) {
        console.warn('Could not auto-install dummy_test_data:', e)
      }
    }
  }).catch(()=>{})
  if(connected[id]) await refreshConnectionHealth()
}
function selectConnection(id){
  if(id===activeConnection.value)return
  guardUnsaved('switch connections',()=>performSelectConnection(id))
}
async function refreshConnectionHealth(){
  if(!activeConnection.value){connectionHealth.value=null;return}
  connectionHealth.value=await call('GetConnectionHealth',activeConnection.value).catch(e=>({connected:false,pingMs:0,mongoDbVersion:'',error:String(e?.message||e)}))
}
async function reconnectActive(){
  if(!activeConnection.value)return
  const r=await attempt(()=>call('Reconnect',activeConnection.value))
  if(!r.ok||!r.value){connected[activeConnection.value]=false;await refreshConnectionHealth();return}
  connected[activeConnection.value]=true;connectionHealth.value=r.value;await refreshDatabases();flash(`Reconnected${r.value.pingMs>=0?` · ${r.value.pingMs} ms`:''}`)
}
async function refreshDatabases() {
  if (!activeConnection.value) return
  databases.value = await run(() => call('ListDatabases', activeConnection.value)).catch(()=>databases.value)
}
async function installOrResetDummyData() {
  if (!activeConnection.value || readOnly.value) return
  const reset = async () => {
    const r = await attempt(() => call('ResetDummyTestData', activeConnection.value))
    if (!r.ok || !r.value) return
    dummyReport.value = r.value
    try { localStorage.setItem(dummySeedMarker(activeConnection.value), '1') } catch {}
    await refreshDatabases()
    openDatabases.dummy_test_data = true
    collections.dummy_test_data = await call('ListCollections', activeConnection.value, 'dummy_test_data').catch(() => []) || []
    flash(`dummy_test_data ready: ${r.value.collections} collections, ${r.value.documents} documents`)
  }
  if (dummyInstalled.value) {
    askConfirm(
      'Reset Dummy Test Data',
      'Rebuild dummy_test_data from the built-in practice dataset? Only dummy_test_data will be dropped and recreated. Every other database is left untouched.',
      reset,
    )
    return
  }
  await reset()
}
async function toggleDatabase(db) {
  if (!activeConnection.value) return
  openDatabases[db] = !openDatabases[db]
  if (openDatabases[db]) collections[db] = await run(() => call('ListCollections', activeConnection.value, db)).catch(()=>[]) || []
}
async function performSelectCollection(db, coll) {
  activeDatabase.value=db; activeCollection.value=coll; selectedRow.value=-1; tab.value='Documents'; skip.value=0
  queryMode.value='sql'; resultView.value='table'; hiddenColumns.value=[]; columnsMenuOpen.value=false
  query.value='{}'; sort.value='{}'; aggregation.value='[\n  { "$match": {} }\n]'; aggregationResults.value=[]; docsRaw.value=[]
  sqlQuery.value=`SELECT *\nFROM ${coll}\nLIMIT 50;`
  eloquentQuery.value=`${linkedLaravelModel.value?.className || classNameFor(coll)}::query()\n    ->limit(50)\n    ->get();`
  autoSizeCodeEditor()
  lastExecutedSQL.value=''
  lastResultMode.value=''
  await runWorkbench({remember:false})
  clearDirtyModes()
  await loadStats()
}
function selectCollection(db,coll){
  if(db===activeDatabase.value&&coll===activeCollection.value)return
  guardUnsaved(`open ${db}.${coll}`,()=>performSelectCollection(db,coll))
}
async function loadDocuments() {
  if (!activeCollection.value) return
  const result = await run(async() => {
    try{return await call('FindDocuments', {connectionId: activeConnection.value, database:activeDatabase.value, collection:activeCollection.value, filterJson:query.value || '{}', sortJson:sort.value || '{}', skip:skip.value, limit:limit.value})}
    catch(e){throw new Error(`Find query failed on ${activeDatabase.value}.${activeCollection.value}: ${String(e?.message||e)}`)}
  }).catch(()=>null)
  if (result) { docsRaw.value=result.documents || []; total.value=result.total || 0; queryMs.value=result.queryMs || 0 }
  return result
}
async function loadStats() {
  if (!activeCollection.value) return
  stats.value = await call('GetCollectionStats', activeConnection.value, activeDatabase.value, activeCollection.value).catch(()=>stats.value)
}
async function runQuery() { skip.value=0; const result=await loadDocuments(); if(result){lastResultMode.value='find';dirtyModes.find=false}; return result }
async function prevPage() { if(queryMode.value!=='find')return; skip.value=Math.max(0,skip.value-limit.value); await loadDocuments() }
async function nextPage() { if(queryMode.value==='find'&&skip.value+limit.value<total.value) { skip.value+=limit.value; await loadDocuments() } }

function setQueryMode(mode){
  queryMode.value=mode; selectedRow.value=-1; resultView.value='table'
  if(mode==='sql' && (!sqlQuery.value || /collection_name/i.test(sqlQuery.value))) {
    sqlQuery.value=`SELECT *\nFROM ${activeCollection.value || 'collection_name'}\nLIMIT 50;`
  }
  if(mode==='eloquent' && (!eloquentQuery.value || /^Model::/i.test(eloquentQuery.value))) {
    eloquentQuery.value=`${linkedLaravelModel.value?.className || classNameFor(activeCollection.value || 'model')}::query()\n    ->limit(50)\n    ->get();`
  }
  if(mode==='sql' || mode==='eloquent') autoSizeCodeEditor()
}
function formatEditor(){
  if(queryMode.value==='sql' || queryMode.value==='eloquent'){ editorText.value=editorText.value.trim().replace(/;?\s*$/, ';'); markQueryDirty(); autoSizeCodeEditor(); return }
  try { editorText.value=JSON.stringify(JSON.parse(editorText.value || (queryMode.value==='find'?'{}':'[]')),null,2); markQueryDirty() }
  catch(e){ fail('Cannot format until the JSON is valid: '+e.message) }
}
function clearEditor(){
  if(queryMode.value==='find') query.value='{}'
  else if(queryMode.value==='aggregation') aggregation.value='[\n  { "$match": {} }\n]'
  else if(queryMode.value==='sql') { sqlQuery.value=`SELECT *\nFROM ${activeCollection.value || 'collection_name'}\nLIMIT 50;`; autoSizeCodeEditor() }
  else { eloquentQuery.value=`${linkedLaravelModel.value?.className || classNameFor(activeCollection.value || 'model')}::query()\n    ->limit(50)\n    ->get();`; autoSizeCodeEditor() }
  markQueryDirty()
}
function useLookupExample(){
  queryMode.value='aggregation'
  aggregation.value='[\n  {\n    "$lookup": {\n      "from": "teams",\n      "localField": "team_id",\n      "foreignField": "_id",\n      "as": "team"\n    }\n  },\n  { "$unwind": { "path": "$team", "preserveNullAndEmptyArrays": true } }\n]'
  markQueryDirty()
}
function rememberQuery(durationMs=0,status='ok',errorMessage=''){
  const item={mode:queryMode.value,text:editorText.value,sort:sort.value,database:activeDatabase.value,collection:activeCollection.value,createdAt:new Date().toISOString(),durationMs:Math.max(0,Math.round(durationMs||0)),status,error:errorMessage||''}
  queryHistory.value=[item,...queryHistory.value.filter(h=>!(h.mode===item.mode&&h.text===item.text&&h.collection===item.collection&&h.database===item.database))].slice(0,50)
  try{ localStorage.setItem('mongo-developer-query-history',JSON.stringify(queryHistory.value)) }catch{}
}
function restoreHistory(){
  if(historySelection.value==='')return
  const item=queryHistory.value[Number(historySelection.value)]; if(!item)return
  queryMode.value=item.mode||'find'; editorText.value=item.text||'{}'; if(item.sort)sort.value=item.sort; historySelection.value=''
  dirtyModes[queryMode.value]=true
  if(queryMode.value==='sql' || queryMode.value==='eloquent') autoSizeCodeEditor()
}
async function runWorkbench(options={}){
  if(!activeCollection.value)return
  if(editorError.value){ fail(editorError.value); return }
  selectedRow.value=-1; cancelInlineEdit()
  const started=performance.now()
  if(queryMode.value==='find'){
    const result=await runQuery()
    if(result&&options.remember!==false)rememberQuery(result.queryMs||performance.now()-started,'ok','')
    return
  }
  let database=activeDatabase.value
  let collection=activeCollection.value
  let pipelineText=aggregation.value
  if(queryMode.value==='sql'){
    const translated=translateSQL(sqlQuery.value,activeCollection.value)
    database=translated.database||activeDatabase.value
    collection=translated.collection||activeCollection.value
    pipelineText=JSON.stringify(translated.pipeline)
  } else if(queryMode.value==='eloquent') {
    const translated=translateEloquent(eloquentQuery.value,{models:laravel.value?.models || [],defaultCollection:activeCollection.value})
    collection=translated.collection||activeCollection.value
    pipelineText=JSON.stringify(translated.pipeline)
  }
  const result=await run(async()=>{
    try{return await call('RunAggregation',activeConnection.value,database,collection,pipelineText)}
    catch(e){throw new Error(`${queryMode.value.toUpperCase()} query failed on ${database}.${collection}: ${String(e?.message||e)}`)}
  }).catch(()=>null)
  const duration=Math.max(0,Math.round(performance.now()-started))
  if(result!==null){
    aggregationResults.value=result || []
    lastResultMode.value=queryMode.value
    if(queryMode.value==='sql')lastExecutedSQL.value=sqlQuery.value
    dirtyModes[queryMode.value]=false
    if(options.remember!==false)rememberQuery(duration,'ok','')
  }
  queryMs.value=duration
}
async function refreshCurrentResults(){
  if(queryMode.value==='find'){ await loadDocuments(); return }
  if(queryMode.value==='sql' && sqlDirty.value && lastExecutedSQL.value){
    const edited=sqlQuery.value
    sqlQuery.value=lastExecutedSQL.value
    await runWorkbench({remember:false})
    sqlQuery.value=edited
      return
  }
  await runWorkbench({remember:false})
}
async function copyResults(){
  const text=JSON.stringify(currentDocs.value,null,2)
  await navigator.clipboard.writeText(text); flash(`Copied ${currentDocs.value.length} result${currentDocs.value.length===1?'':'s'}`)
}
function toggleQueryExpanded(){ queryExpanded.value=!queryExpanded.value; if(queryExpanded.value)resultsExpanded.value=false }
function toggleResultsExpanded(){ resultsExpanded.value=!resultsExpanded.value; if(resultsExpanded.value)queryExpanded.value=false }
function autoSizeCodeEditor(){
  if(!['sql','eloquent'].includes(queryMode.value) || queryExpanded.value || resultsExpanded.value)return
  editorHeight.value=sqlEditorAutoHeight(editorText.value, window.innerHeight)
}
function handleEditorInput(){
  markQueryDirty()
  if(queryMode.value==='sql' || queryMode.value==='eloquent') autoSizeCodeEditor()
}
function startEditorResize(event){
  if(queryExpanded.value||resultsExpanded.value)return
  event.preventDefault(); resizeStartY=event.clientY; resizeStartHeight=editorHeight.value
  window.addEventListener('mousemove',resizeEditor); window.addEventListener('mouseup',stopEditorResize)
}
function resizeEditor(event){
  const max=Math.max(180,Math.min(480,window.innerHeight-430))
  editorHeight.value=Math.max(123,Math.min(max,resizeStartHeight+(event.clientY-resizeStartY)))
}
function stopEditorResize(){
  window.removeEventListener('mousemove',resizeEditor); window.removeEventListener('mouseup',stopEditorResize)
  try{localStorage.setItem('mongo-developer-v122-editor-height',String(editorHeight.value))}catch{}
}

function newConnection() {
  Object.assign(form,{id:'',name:'',uri:'mongodb://localhost:27017/',host:'localhost',port:27017,username:'',password:'',authDb:'admin',tls:false,readOnly:false}); connectionModal.value=true
}
function editConnection(p) {
  Object.assign(form,{id:p.id,name:p.name,uri:p.uri,host:'',port:27017,username:p.username||'',password:'',authDb:p.authDb||'',tls:!!p.tls,readOnly:!!p.readOnly}); connectionModal.value=true
}
async function testConnection() { const r=await attempt(()=>call('TestConnection',{...form})); if(r.ok) flash('Connection test passed') }
async function saveConnection() {
  const r = await attempt(()=>call('SaveConnection',{...form}))
  if (!r.ok || !r.value) return
  connectionModal.value=false; await loadConnections(); flash('Connection saved')
}
function deleteConnection(p) {
  askConfirm('Delete saved connection', `Delete “${p.name}”? The MongoDB database itself will not be changed.`, async()=>{
    const r=await attempt(()=>call('DeleteConnection',p.id)); if(!r.ok)return; if(activeConnection.value===p.id){activeConnection.value=null;databases.value=[];connectionHealth.value=null}; await loadConnections(); flash('Saved connection deleted')
  })
}

function openCreate(mode) { createMode.value=mode; createName.value=''; createModal.value=true }
async function doCreate() {
  if (!createName.value.trim()) return
  const action = createMode.value==='database'
    ? ()=>call('CreateDatabase',activeConnection.value,createName.value.trim(),'initial')
    : ()=>call('CreateCollection',activeConnection.value,activeDatabase.value,createName.value.trim())
  const r=await attempt(action); if(!r.ok)return
  createModal.value=false; await refreshDatabases(); if(activeDatabase.value){collections[activeDatabase.value]=await call('ListCollections',activeConnection.value,activeDatabase.value)}; flash(`${createMode.value} created`)
}
function dropDatabase(db) {
  if(readOnly.value) return
  askConfirm('Drop database', `Permanently DROP database ${db}? All collections and documents in it will be deleted.`, async()=>{
    const r=await attempt(()=>call('DropDatabase',activeConnection.value,db)); if(!r.ok)return; if(activeDatabase.value===db){activeDatabase.value='';activeCollection.value='';docsRaw.value=[]}; delete collections[db]; delete openDatabases[db]; await refreshDatabases(); flash('Database dropped')
  })
}
function dropCollection() {
  if(readOnly.value||!activeCollection.value) return
  const db=activeDatabase.value, coll=activeCollection.value
  askConfirm('Drop collection', `Permanently DROP ${db}.${coll}? All documents in this collection will be deleted.`, async()=>{
    const r=await attempt(()=>call('DropCollection',activeConnection.value,db,coll)); if(!r.ok)return; activeCollection.value=''; docsRaw.value=[]; collections[db]=await call('ListCollections',activeConnection.value,db); flash('Collection dropped')
  })
}

function newDocument(){ editingDocument.value=false; documentText.value='{\n  \n}'; documentModal.value=true }
function closeDocumentModal(){
  const dirty=documentText.value.trim() !== '{\n  \n}'
  if(!documentModal.value||!dirty){documentModal.value=false;return}
  askConfirm('Discard unsaved document','Close this new document without saving it?',()=>{documentModal.value=false},{danger:false,label:'Discard document'})
}
function closeDocumentDrawer(){
  if(!documentDrawer.value){return}
  if(!drawerDirty.value){documentDrawer.value=false;return}
  askConfirm('Discard document changes','Close the document editor without saving your changes?',()=>{documentDrawer.value=false},{danger:false,label:'Discard changes'})
}
function openCreateCollectionFor(db){
  const action=()=>{activeDatabase.value=db;activeCollection.value='';docsRaw.value=[];aggregationResults.value=[];selectedRow.value=-1;openCreate('collection')}
  guardUnsaved(`create a collection in ${db}`,action)
}
function rowCanWrite(i){ return !readOnly.value && queryMode.value!=='aggregation' && currentDocs.value[i]?._id !== undefined }
function openDocumentDrawer(i,tabName='fields'){
  const d=currentDocs.value[i]; if(!d)return
  selectedRow.value=i
  drawerOriginal.value=JSON.parse(JSON.stringify(d))
  drawerDocument.value=JSON.parse(JSON.stringify(d))
  drawerDocumentText.value=JSON.stringify(d,null,2)
  drawerTab.value=tabName
  documentDrawer.value=true
}
function editDocument(){ if(selectedRow.value<0||!rowCanWrite(selectedRow.value))return; openDocumentDrawer(selectedRow.value,'fields') }
function setDrawerTab(name){
  if(name==='json')drawerDocumentText.value=JSON.stringify(drawerDocument.value,null,2)
  else if(drawerTab.value==='json'){
    try{drawerDocument.value=JSON.parse(drawerDocumentText.value)}catch(e){fail('Fix the JSON before switching to Fields: '+e.message);return}
  }
  drawerTab.value=name
}
async function saveDrawerDocument(){
  if(!drawerEditable.value)return
  let doc=drawerDocument.value
  if(drawerTab.value==='json'){
    const validationError=validateDocumentJSON(drawerDocumentText.value)
    if(validationError){fail(validationError);return}
    doc=JSON.parse(drawerDocumentText.value)
  }
  doc._id=JSON.parse(JSON.stringify(drawerOriginal.value._id))
  const r=await attempt(()=>call('ReplaceDocument',activeConnection.value,activeDatabase.value,activeCollection.value,idFilter(drawerOriginal.value),JSON.stringify(doc)))
  if(!r.ok)return
  documentDrawer.value=false; await refreshCurrentResults(); await loadStats(); flash('Document updated')
}
async function saveDocument(){
  const validationError=validateDocumentJSON(documentText.value)
  if(validationError){fail(validationError);return}
  const r=await attempt(()=>call('InsertDocument',activeConnection.value,activeDatabase.value,activeCollection.value,documentText.value))
  if(!r.ok)return
  documentModal.value=false; await refreshCurrentResults(); await loadStats(); flash('Document inserted')
}
function deleteDocument(){ if(selectedRow.value>=0)deleteDocumentAt(selectedRow.value) }
function deleteDocumentAt(i){
  if(!rowCanWrite(i))return; const d=currentDocs.value[i]
  askConfirm('Delete document','Permanently delete this document?',async()=>{
    const r=await attempt(()=>call('DeleteDocument',activeConnection.value,activeDatabase.value,activeCollection.value,idFilter(d))); if(!r.ok)return; selectedRow.value=-1; documentDrawer.value=false; await refreshCurrentResults(); await loadStats(); flash('Document deleted')
  })
}
async function duplicateDocument(i){
  if(!rowCanWrite(i))return
  const copy=JSON.parse(JSON.stringify(currentDocs.value[i])); delete copy._id
  const r=await attempt(()=>call('InsertDocument',activeConnection.value,activeDatabase.value,activeCollection.value,JSON.stringify(copy)))
  if(!r.ok)return; await refreshCurrentResults(); await loadStats(); flash('Document duplicated')
}
function isSimpleValue(value){ return value===null || ['string','number','boolean'].includes(typeof value) }
function beginCellEdit(i,field){
  const d=currentDocs.value[i]
  if(!rowCanWrite(i)||field==='_id')return
  if(!isSimpleValue(d?.[field])){ openDocumentDrawer(i,'json'); return }
  selectedRow.value=i
  inlineEdit.row=i; inlineEdit.field=field; inlineEdit.originalType=d[field]===null?'null':typeof d[field]
  inlineEdit.value=d[field]===null?'null':String(d[field])
}
function cancelInlineEdit(){ inlineEdit.row=-1; inlineEdit.field=''; inlineEdit.value=''; inlineEdit.originalType='' }
function coercedInlineValue(){
  const raw=inlineEdit.value
  if(inlineEdit.originalType==='number'){ const n=Number(raw); if(!Number.isFinite(n))throw new Error('Enter a valid number.'); return n }
  if(inlineEdit.originalType==='boolean'){ if(!/^(true|false)$/i.test(raw.trim()))throw new Error('Boolean values must be true or false.'); return raw.trim().toLowerCase()==='true' }
  if(inlineEdit.originalType==='null')return /^null$/i.test(raw.trim())?null:raw
  return raw
}
async function saveInlineCell(){
  if(inlineEdit.row<0)return
  const original=currentDocs.value[inlineEdit.row]; if(!original?._id)return
  let value; try{value=coercedInlineValue()}catch(e){fail(e);return}
  const replacement=JSON.parse(JSON.stringify(original)); replacement[inlineEdit.field]=value
  const r=await attempt(()=>call('ReplaceDocument',activeConnection.value,activeDatabase.value,activeCollection.value,idFilter(original),JSON.stringify(replacement)))
  if(!r.ok)return
  cancelInlineEdit(); await refreshCurrentResults(); flash('Cell updated')
}
function toggleColumn(field){ hiddenColumns.value=hiddenColumns.value.includes(field)?hiddenColumns.value.filter(c=>c!==field):[...hiddenColumns.value,field] }
function resetColumns(){ hiddenColumns.value=[]; columnsMenuOpen.value=false }
function startColumnResize(event,field){
  event.preventDefault(); event.stopPropagation()
  const startX=event.clientX; const start=columnWidths[field] || (field==='_id'?190:160)
  const move=e=>{columnWidths[field]=Math.max(90,Math.min(520,start+(e.clientX-startX)))}
  const up=()=>{window.removeEventListener('mousemove',move);window.removeEventListener('mouseup',up)}
  window.addEventListener('mousemove',move);window.addEventListener('mouseup',up)
}
async function openIndexes(){ tab.value='Indexes'; indexes.value=await run(()=>call('ListIndexes',activeConnection.value,activeDatabase.value,activeCollection.value)).catch(()=>[]) || [] }
async function createIndex(){
  const r=await attempt(()=>call('CreateIndex',activeConnection.value,activeDatabase.value,activeCollection.value,indexForm.keys,indexForm.name,indexForm.unique)); if(!r.ok)return; indexModal.value=false; await openIndexes(); flash('Index created')
}
async function importJSON(){ const r=await attempt(()=>call('ImportJSON',activeConnection.value,activeDatabase.value,activeCollection.value)); if(r.ok&&r.value!==null){flash(`Imported ${r.value} documents`);await loadDocuments();await loadStats()} }
async function exportJSON(){ const r=await attempt(()=>call('ExportJSON',activeConnection.value,activeDatabase.value,activeCollection.value)); if(r.ok&&r.value)flash(`Exported to ${r.value}`) }

function persistSavedQueries(){ try{localStorage.setItem('mongo-developer-saved-queries',JSON.stringify(savedQueries.value))}catch{} }
function openSaveQuery(){
  if(!activeCollection.value)return
  savedQueryName.value=`${activeCollection.value} ${queryMode.value}`
  savedQueriesModal.value=true
}
function saveCurrentQuery(){
  const name=savedQueryName.value.trim(); if(!name){fail('Give the saved query a name.');return}
  const item={id:`q-${Date.now()}-${Math.random().toString(16).slice(2)}`,name,mode:queryMode.value,text:editorText.value,sort:sort.value,database:activeDatabase.value,collection:activeCollection.value,createdAt:new Date().toISOString()}
  savedQueries.value=[item,...savedQueries.value].slice(0,100);persistSavedQueries();savedQueriesModal.value=false;flash(`Saved query: ${name}`)
}
function loadSavedQuery(item){
  const apply=async()=>{
    if(item.database&&item.collection&&(item.database!==activeDatabase.value||item.collection!==activeCollection.value))await performSelectCollection(item.database,item.collection)
    queryMode.value=item.mode||'find';editorText.value=item.text||'{}';sort.value=item.sort||'{}';dirtyModes[queryMode.value]=true;savedQueriesModal.value=false
    if(queryMode.value==='sql'||queryMode.value==='eloquent')autoSizeCodeEditor()
    flash(`Loaded saved query: ${item.name}`)
  }
  guardUnsaved(`load saved query “${item.name}”`,apply)
}
function deleteSavedQuery(item){
  savedQueries.value=savedQueries.value.filter(q=>q.id!==item.id);persistSavedQueries();flash('Saved query deleted')
}
async function explainCurrentQuery(){
  if(!activeConnection.value||!activeCollection.value)return
  if(editorError.value){fail(editorError.value);return}
  let result=null
  if(queryMode.value==='find'){
    result=await run(()=>call('ExplainFind',activeConnection.value,activeDatabase.value,activeCollection.value,query.value||'{}',sort.value||'{}')).catch(()=>null)
  }else{
    let database=activeDatabase.value,collection=activeCollection.value,pipelineText=aggregation.value
    if(queryMode.value==='sql'){
      const translated=translateSQL(sqlQuery.value,activeCollection.value);database=translated.database||activeDatabase.value;collection=translated.collection||activeCollection.value;pipelineText=JSON.stringify(translated.pipeline)
    }else if(queryMode.value==='eloquent'){
      const translated=translateEloquent(eloquentQuery.value,{models:laravel.value?.models||[],defaultCollection:activeCollection.value});collection=translated.collection||activeCollection.value;pipelineText=JSON.stringify(translated.pipeline)
    }
    result=await run(()=>call('ExplainAggregation',activeConnection.value,database,collection,pipelineText)).catch(()=>null)
  }
  if(result){result.mode=queryMode.value;explainResult.value=result;explainModal.value=true}
}
async function copySuggestedIndex(){if(!explainResult.value?.suggestedIndex)return;await navigator.clipboard.writeText(explainResult.value.suggestedIndex);flash('Suggested index copied')}
function prepareSuggestedIndex(){
  if(!explainResult.value?.suggestedIndex)return
  indexForm.keys=explainResult.value.suggestedIndex;indexForm.name='';indexForm.unique=false;explainModal.value=false;indexModal.value=true
}
async function openDiagnostics(){diagnosticsModal.value=true;diagnostics.value=await call('GetDiagnostics',activeConnection.value||'').catch(e=>({version:'1.5.0',connectionError:String(e?.message||e)}))}
async function copyDiagnostics(){if(!diagnosticsText.value)return;await navigator.clipboard.writeText(diagnosticsText.value);flash('Diagnostics copied')}

function backupSize(bytes){
  const n=Number(bytes||0); if(n<1024)return `${n} B`; if(n<1024*1024)return `${(n/1024).toFixed(1)} KB`; if(n<1024*1024*1024)return `${(n/1024/1024).toFixed(1)} MB`; return `${(n/1024/1024/1024).toFixed(2)} GB`
}
function backupDate(value){ try{return new Date(value).toLocaleString()}catch{return value||'—'} }
async function loadBackups(){ backups.value=await call('ListBackups').catch(()=>[]) || []; backupDirectory.value=await call('GetBackupDirectory').catch(()=>backupDirectory.value) || backupDirectory.value }
async function openBackupManager(){
  backupModal.value=true
  if(activeDatabase.value && !['admin','config','local'].includes(activeDatabase.value.toLowerCase())) backupDatabase.value=activeDatabase.value
  else if(!backupDatabase.value && backupDatabases.value.length) backupDatabase.value=backupDatabases.value[0].name
  await loadBackups()
}
async function createBackup(scope){
  if(!activeConnection.value){ fail('Select and connect to MongoDB before creating a backup.'); return }
  const database=scope==='database'?backupDatabase.value:''
  if(scope==='database'&&!database){ fail('Choose a database to back up.'); return }
  const r=await attempt(()=>call('CreateBackup',activeConnection.value,scope,database))
  if(!r.ok||!r.value)return
  await loadBackups()
  if(r.value.downloadPath) flash(`Backup created and copy saved to ${r.value.downloadPath}`)
  else if(r.value.downloadError) flash('Backup saved safely inside Mongo Developer. The extra Save As copy could not be created.')
  else flash('Backup saved inside Mongo Developer. Save As copy was cancelled.')
}
async function downloadBackup(item){
  const r=await attempt(()=>call('DownloadBackup',item.fileName)); if(r.ok&&r.value)flash(`Backup copy saved to ${r.value}`)
}
function deleteBackup(item){
  askConfirm('Delete Backup ZIP',`Delete ${item.fileName} from Mongo Developer's stored backups? This only deletes the backup ZIP; it does not touch any MongoDB database.`,async()=>{
    const r=await attempt(()=>call('DeleteBackup',item.fileName)); if(!r.ok)return; await loadBackups(); flash('Backup ZIP deleted')
  })
}
function restoreStoredBackup(item){
  if(!activeConnection.value){ fail('Select the MongoDB connection you want to restore into first.'); return }
  if(readOnly.value){ fail('This connection is read-only; restore is disabled.'); return }
  const scope=item.scope==='all'?`${item.databaseCount} databases`:`database ${item.database||''}`
  askConfirm('Restore Backup',`Restore ${item.fileName} into ${activeProfile.value?.name||'the active connection'}? Mongo Developer will DROP and replace the ${scope} contained in this backup. Databases not contained in the ZIP are left untouched.`,async()=>{
    const r=await attempt(()=>call('RestoreBackup',activeConnection.value,item.fileName)); if(!r.ok||!r.value)return; await finishRestore(r.value)
  })
}
function restoreExternalBackup(){
  if(!activeConnection.value){ fail('Select the MongoDB connection you want to restore into first.'); return }
  if(readOnly.value){ fail('This connection is read-only; restore is disabled.'); return }
  askConfirm('Restore Backup ZIP',`Choose a Mongo Developer backup ZIP to restore into ${activeProfile.value?.name||'the active connection'}. Databases contained in that ZIP will be dropped and replaced; unrelated databases are left untouched.`,async()=>{
    const r=await attempt(()=>call('RestoreBackupFromFile',activeConnection.value)); if(!r.ok||!r.value?.fileName)return; await finishRestore(r.value)
  })
}
async function finishRestore(report){
  activeDatabase.value=''; activeCollection.value=''; docsRaw.value=[]; aggregationResults.value=[]; selectedRow.value=-1
  await refreshDatabases(); await loadBackups()
  flash(`Restore complete: ${report.databaseCount} database(s), ${report.collectionCount} collections, ${report.documentCount} documents`)
}

async function linkLaravel(){
  const selected=await run(()=>call('SelectLaravelProject')).catch(()=>null)
  if(!selected?.path)return
  laravel.value=selected
  laravelLinkEpoch.value++
  try{localStorage.setItem('mongo-developer-laravel-project', selected.path)}catch{}
  eloquentQuery.value=`${linkedLaravelModel.value?.className || classNameFor(activeCollection.value || 'model')}::query()
    ->limit(50)
    ->get();`
}
function unlinkLaravel(){
  if(!laravel.value) return
  laravel.value=null
  laravelLinkEpoch.value++
  setupHelp.value=''
  modelName.value=''
  try{localStorage.removeItem('mongo-developer-laravel-project')}catch{}
  flash('Laravel project unlinked. The project folder and model files were not changed.')
}
async function refreshLaravel(){ if(laravel.value?.path){ laravel.value=await call('InspectLaravelProject',laravel.value.path); laravelLinkEpoch.value++ } }
async function generateModel(){
  if(!laravel.value?.path||!activeCollection.value)return
  const sample=currentRawResults.value[0] || '{}'; const name=modelName.value||classNameFor(activeCollection.value)
  const r=await attempt(()=>call('GenerateLaravelModel',laravel.value.path,activeCollection.value,name,sample)); if(r.ok&&r.value){ await refreshLaravel(); flash(`Generated ${r.value}`) }
}
async function showLaravelHelp(){ if(!activeConnection.value||!activeDatabase.value)return; setupHelp.value=await call('LaravelSetupHelp',activeConnection.value,activeDatabase.value) }
async function copyHelp(){ await navigator.clipboard.writeText(setupHelp.value); flash('Copied Laravel setup to clipboard') }

function clampSidebar(side, width){
  const min=side==='connections'?145:150
  const max=side==='connections'?330:360
  return Math.max(min,Math.min(max,width))
}
function startSidebarResize(event,side){
  event.preventDefault()
  sidebarResize.active=true; sidebarResize.side=side; sidebarResize.startX=event.clientX
  sidebarResize.startWidth=side==='connections'?connectionsWidth.value:databasesWidth.value
  window.addEventListener('mousemove',resizeSidebar); window.addEventListener('mouseup',stopSidebarResize)
}
function resizeSidebar(event){
  if(!sidebarResize.active)return
  const next=clampSidebar(sidebarResize.side,sidebarResize.startWidth+(event.clientX-sidebarResize.startX))
  if(sidebarResize.side==='connections')connectionsWidth.value=next; else databasesWidth.value=next
}
function stopSidebarResize(){
  if(sidebarResize.active){
    try{localStorage.setItem('mongo-developer-connections-width',String(connectionsWidth.value));localStorage.setItem('mongo-developer-databases-width',String(databasesWidth.value))}catch{}
  }
  sidebarResize.active=false; sidebarResize.side=''
  window.removeEventListener('mousemove',resizeSidebar); window.removeEventListener('mouseup',stopSidebarResize)
}
function resetSidebarWidth(side){
  if(side==='connections')connectionsWidth.value=158; else databasesWidth.value=165
  stopSidebarResize()
}

function onGlobalKey(e){
  const mod=e.ctrlKey||e.metaKey
  if(mod&&e.key.toLowerCase()==='k'){e.preventDefault();searchInput.value?.focus();return}
  if(mod&&e.key==='Enter'&&activeCollection.value&&tab.value==='Documents'){e.preventDefault();runWorkbench();return}
  if(mod&&e.shiftKey&&e.key.toLowerCase()==='f'&&activeCollection.value&&tab.value==='Documents'){e.preventDefault();formatEditor();return}
  if(mod&&e.key.toLowerCase()==='l'&&activeCollection.value&&tab.value==='Documents'){e.preventDefault();editorInput.value?.focus();return}
  if(mod&&e.key.toLowerCase()==='s'){e.preventDefault();if(documentDrawer.value&&drawerEditable.value){saveDrawerDocument();return}if(documentModal.value){saveDocument();return}if(activeCollection.value&&tab.value==='Documents'){openSaveQuery();return}}
  if(e.key==='Escape'){
    if(savedQueriesModal.value){savedQueriesModal.value=false;return}if(explainModal.value){explainModal.value=false;return}if(diagnosticsModal.value){diagnosticsModal.value=false;return}if(documentModal.value){closeDocumentModal();return}if(documentDrawer.value){closeDocumentDrawer();return}if(queryExpanded.value||resultsExpanded.value){queryExpanded.value=false;resultsExpanded.value=false;return}
  }
}
function onBeforeUnload(e){if(!hasUnsavedWork.value)return;e.preventDefault();e.returnValue=''}
onMounted(async()=>{
  loadConnections(); window.addEventListener('keydown',onGlobalKey);window.addEventListener('beforeunload',onBeforeUnload)
  try{
    const cw=Number(localStorage.getItem('mongo-developer-connections-width')); if(cw)connectionsWidth.value=clampSidebar('connections',cw)
    const dw=Number(localStorage.getItem('mongo-developer-databases-width')); if(dw)databasesWidth.value=clampSidebar('databases',dw)
    const savedHeight=Number(localStorage.getItem('mongo-developer-v122-editor-height')); if(savedHeight>=123)editorHeight.value=savedHeight
    const savedHistory=JSON.parse(localStorage.getItem('mongo-developer-query-history')||'[]'); if(Array.isArray(savedHistory))queryHistory.value=savedHistory.slice(0,50)
    const saved=JSON.parse(localStorage.getItem('mongo-developer-saved-queries')||'[]'); if(Array.isArray(saved))savedQueries.value=saved.slice(0,100)
    const savedLaravel=localStorage.getItem('mongo-developer-laravel-project')
    if(savedLaravel){ laravel.value=await call('InspectLaravelProject',savedLaravel).catch(()=>null); if(!laravel.value)localStorage.removeItem('mongo-developer-laravel-project') }
  }catch{}
})
onBeforeUnmount(()=>{ window.removeEventListener('keydown',onGlobalKey);window.removeEventListener('beforeunload',onBeforeUnload); stopEditorResize(); stopSidebarResize() })
</script>

<template>
<div class="app-shell">
  <header class="topbar">
    <div class="brand"><Database :size="19"/><strong>Mongo Developer</strong><span class="version">v1.5.0</span></div>
    <div class="global-search"><Search :size="16"/><input ref="searchInput" v-model="searchText" placeholder="Filter connections / databases (Ctrl+K)"></div>
    <div class="top-actions">
      <button class="btn backup-top" @click="openBackupManager"><Save :size="16"/> Backup & Restore</button>
      <button class="btn primary" @click="newConnection"><Plus :size="16"/> New Connection</button>
      <button class="btn" @click="refreshDatabases" :disabled="!activeConnection"><RefreshCw :size="16"/> Refresh</button>
      <button class="btn accent" @click="runWorkbench" :disabled="!activeCollection||tab!=='Documents'||!!editorError"><Play :size="16"/> {{queryMode==='sql'?(sqlDirty?'Run Changes':'Run Query'):'Run Query'}} <span class="shortcut-note">Ctrl+Enter</span></button>
    </div>
  </header>

  <div class="body-grid" :style="bodyGridStyle">
    <nav class="rail">
      <button class="rail-item active rail-button" @click="tab='Documents'"><Plug/><span>Connections</span></button>
      <button class="rail-item rail-button" @click="tab='Documents';setQueryMode('find')" :disabled="!activeCollection"><Table2/><span>Documents</span></button>
      <button class="rail-item rail-button" @click="tab='Documents';setQueryMode('aggregation')" :disabled="!activeCollection"><Waypoints/><span>Query</span></button>
      <button class="rail-item rail-button" @click="tab='Laravel'" :disabled="!activeCollection"><Code2/><span>Laravel</span></button>
      <div class="rail-spacer"></div>
      <button class="rail-item rail-button" @click="openDiagnostics"><ShieldCheck/><span>Diagnostics</span></button>
      <div class="rail-item rail-static"><ShieldCheck/><span>Safe CRUD</span></div>
    </nav>

    <aside class="connections-pane">
      <div class="pane-title"><span>Connections</span><div class="pane-actions"><button class="icon-btn" @click="newConnection" title="New connection"><Plus/></button></div></div><div class="pane-resizer" title="Drag to resize. Double-click to reset compact width." @mousedown="startSidebarResize($event,'connections')" @dblclick="resetSidebarWidth('connections')"><span></span></div>
      <div v-if="!connections.length" class="empty compact">No saved connections yet.<button class="link" @click="newConnection">Add one</button></div>
      <div v-for="c in filteredConnections" :key="c.id" class="connection-card" :class="{selected:c.id===activeConnection}" @click="selectConnection(c.id)">
        <Server :size="18"/>
        <div class="grow"><strong>{{c.name}}</strong><small v-if="c.readOnly" class="status read"><Shield :size="12"/> Read Only</small><small v-else-if="connected[c.id]" class="status ok"><CheckCircle2 :size="12"/> Connected</small><small v-else class="status"><Circle :size="10"/> Saved</small></div>
        <button class="icon-btn mini" @click.stop="editConnection(c)" title="Edit"><Pencil/></button>
        <button class="icon-btn mini danger-text" @click.stop="deleteConnection(c)" title="Delete"><Trash2/></button>
      </div>
    </aside>

    <aside class="database-pane">
      <div class="pane-title"><span>Databases</span><div class="pane-actions"><button class="icon-btn" @click="refreshDatabases" title="Refresh databases"><RefreshCw/></button><button class="icon-btn" @click="openCreate('database')" :disabled="!activeConnection||readOnly" title="Create database"><Plus/></button></div></div><div class="pane-resizer" title="Drag to resize. Double-click to reset compact width." @mousedown="startSidebarResize($event,'databases')" @dblclick="resetSidebarWidth('databases')"><span></span></div>
      <div class="search-row"><Search :size="14"/><input v-model="searchText" placeholder="Filter loaded databases / collections"></div>
      <div v-if="activeConnection" class="dummy-seed-card">
        <div class="dummy-seed-copy">
          <strong><Database :size="14"/> Practice Data</strong>
          <small v-if="dummyInstalled">dummy_test_data is installed<span v-if="dummyReport?.collections"> · {{dummyReport.collections}} collections · {{dummyReport.documents}} docs</span></small>
          <small v-else>Install the built-in resettable practice database.</small>
        </div>
        <button class="btn small" @click="installOrResetDummyData" :disabled="readOnly" :title="readOnly?'Read-only connections cannot install practice data':''">
          <RefreshCw :size="14"/> {{dummyInstalled?'Reset':'Install'}}
        </button>
      </div>
      <div v-if="!activeConnection" class="empty compact">Select a connection.</div>
      <div v-for="db in filteredDatabases" :key="db.name" class="db-node">
        <div class="db-line"><button class="db-row" :class="{selected:activeDatabase===db.name&&!activeCollection}" @click="toggleDatabase(db.name)"><component :is="openDatabases[db.name]?ChevronDown:ChevronRight" :size="15"/><Database :size="15"/><span>{{db.name}}</span></button><button class="icon-btn mini danger-text db-drop" @click="dropDatabase(db.name)" :disabled="readOnly" title="Drop database"><Trash2/></button></div>
        <div v-if="openDatabases[db.name]" class="collections">
          <button v-for="c in visibleCollections(db.name)" :key="c.name" class="collection-row" :class="{selected:activeDatabase===db.name&&activeCollection===c.name}" @click="selectCollection(db.name,c.name)"><Table2 :size="14"/><span>{{c.name}}</span></button>
          <button class="collection-row add-row" @click="openCreateCollectionFor(db.name)" :disabled="readOnly"><Plus :size="14"/> New collection</button>
        </div>
      </div>
    </aside>

    <main class="workspace">
      <div class="context-bar"><span>Connection: <b>{{activeProfile?.name||'—'}}</b></span><span v-if="connectionHealth?.connected" class="health-pill"><CheckCircle2/> {{connectionHealth.pingMs}} ms<span v-if="connectionHealth.mongoDbVersion"> · MongoDB {{connectionHealth.mongoDbVersion}}</span></span><button v-else-if="activeConnection" class="btn tiny reconnect-btn" @click="reconnectActive"><RefreshCw/> Reconnect</button><ChevronRight/><span>Database: <b>{{activeDatabase||'—'}}</b></span><ChevronRight/><span>Collection: <b>{{activeCollection||'—'}}</b></span><span v-if="readOnly" class="readonly-pill"><Shield/> READ ONLY</span></div>
      <div class="tabs" v-if="activeCollection">
        <button v-for="t in ['Documents','Schema','Indexes','Laravel']" :key="t" :class="{active:tab===t}" @click="t==='Indexes'?openIndexes():tab=t">{{t}}</button>
      </div>

      <section v-if="!activeCollection" class="welcome">
        <Database :size="48"/><h1>Mongo Developer</h1><p>Connect to MongoDB, browse databases and collections, edit documents, run native filters, aggregations or SQL-style queries, and bridge collections into Laravel.</p><button class="btn primary big" @click="newConnection"><Plus/> Add your first connection</button>
      </section>

      <section v-else-if="tab==='Documents'" class="documents-view" :class="{'query-expanded':queryExpanded,'results-expanded':resultsExpanded}">
        <div v-if="!resultsExpanded" class="query-workbench">
          <div class="workbench-toolbar">
            <div class="mode-switch" role="tablist" aria-label="Query mode">
              <button :class="{active:queryMode==='find'}" @click="setQueryMode('find')">Find JSON</button>
              <button :class="{active:queryMode==='aggregation'}" @click="setQueryMode('aggregation')">Aggregation</button>
              <button :class="{active:queryMode==='sql'}" @click="setQueryMode('sql')">SQL</button><button :class="{active:queryMode==='eloquent'}" @click="setQueryMode('eloquent')">Eloquent</button>
            </div>
            <span class="workbench-label">{{editorLabel}}</span>
            <span v-if="queryMode==='sql'&&!sqlDirty&&lastExecutedSQL" class="auto-loaded-badge"><CheckCircle2/> Auto-loaded</span>
            <span v-else-if="queryMode==='sql'&&sqlDirty" class="edited-badge"><Pencil/> Query edited</span>
            <select class="history-select" v-model="historySelection" @change="restoreHistory" title="Query history">
              <option value="">History…</option>
              <option v-for="(h,i) in queryHistory" :key="h.createdAt+i" :value="String(i)">{{h.mode.toUpperCase()}} · {{h.database}}.{{h.collection}} · {{h.durationMs??'—'}} ms · {{h.text.replace(/\s+/g,' ').slice(0,42)}}</option>
            </select>
            <button class="btn small" @click="openSaveQuery"><Save/> Save Query</button>
            <button class="btn small" @click="savedQueriesModal=true"><ListFilter/> Saved {{savedQueries.length?`(${savedQueries.length})`:``}}</button>
            <button class="btn small" @click="formatEditor">Format</button>
            <button class="btn small" @click="clearEditor">Reset</button>
            <button class="btn small" v-if="queryMode==='aggregation'" @click="useLookupExample">$lookup example</button>
            <button class="btn small" v-if="queryMode==='eloquent'" @click="clearEditor">Model example</button>
            <button class="btn small" @click="toggleQueryExpanded">{{queryExpanded?'Restore':'Expand'}}</button>
            <button class="btn small accent" @click="explainCurrentQuery" :disabled="!!editorError"><Search/> Explain</button>
            <button class="btn primary small" @click="runWorkbench" :disabled="!!editorError"><Play/> {{queryMode==='sql'?(sqlDirty?'Run Changes':'Run'):'Run'}}</button>
          </div>
          <div class="editor-help"><span>{{queryMode==='sql'?'Collection clicks run automatically. The SQL shown below is the query that produced the results; edit it directly only when you want to change the query.':editorHelp}}</span><span class="keyboard-hint">Ctrl+Enter to run changes</span></div>

          <div class="editor-shell" :class="{invalid:!!editorError}" :style="{height:queryExpanded?'calc(100vh - 205px)':editorHeight+'px'}">
            <div class="line-gutter">1</div>
            <textarea ref="editorInput" v-model="editorText" spellcheck="false" :aria-label="editorLabel" @input="handleEditorInput" @keydown.ctrl.enter.prevent="runWorkbench"></textarea>
          </div>
          <div class="editor-footer">
            <span v-if="editorError" class="validation-error"><XCircle/> {{editorError}}</span>
            <span v-else class="validation-ok"><CheckCircle2/> Valid {{queryMode==='sql'?'SQL':queryMode==='eloquent'?'Eloquent preview':'JSON'}}</span>
            <span v-if="queryMode==='sql'&&!sqlDirty&&lastExecutedSQL&&lastResultMode==='sql'" class="query-state-copy">Results match the SQL shown above.</span>
            <span v-else-if="queryMode==='sql'&&!sqlDirty&&lastResultMode&&lastResultMode!=='sql'" class="query-state-copy edited">Results are from {{lastResultMode}} mode — press Run to execute the SQL shown above.</span>
            <span v-if="queryMode==='sql'&&sqlDirty" class="query-state-copy edited">SQL changed — results stay as they are until Run Changes.</span><span v-else-if="queryMode!=='sql'&&queryDirty" class="query-state-copy edited">Query edited — press Run when ready.</span>
            <label v-if="queryMode==='find'" class="sort-workbench">Sort <input v-model="sort" placeholder='{ "name": 1 }' @input="markQueryDirty" @keyup.enter="runWorkbench"></label>
          </div>
        </div>

        <div v-if="!queryExpanded&&!resultsExpanded" class="vertical-splitter" title="Drag to resize query and results" @mousedown="startEditorResize"><span></span></div>

        <div v-if="!queryExpanded" class="results-workbench">
          <div class="results-toolbar">
            <div class="crud-actions">
              <button class="btn small" @click="newDocument" :disabled="readOnly"><FilePlus2/> New Document</button>
              <button class="btn small" @click="editDocument" :disabled="!canEditResult"><Pencil/> Edit Selected</button>
              <button class="btn small danger" @click="deleteDocument" :disabled="!canEditResult"><Trash2/> Delete Selected</button>
            </div>
            <div class="result-actions">
              <span class="result-meta">{{resultSummary}} · {{queryMs}} ms</span>
              <button class="btn small" @click="refreshCurrentResults"><RefreshCw/> Refresh</button>
              <label v-if="queryMode==='find'" class="limit-control">Page <select v-model.number="limit" @change="runQuery"><option :value="25">25</option><option :value="50">50</option><option :value="100">100</option><option :value="250">250</option></select></label>
              <div class="columns-control">
                <button class="btn small" @click="columnsMenuOpen=!columnsMenuOpen"><Table2/> Columns</button>
                <div v-if="columnsMenuOpen" class="columns-menu">
                  <strong>Visible columns</strong>
                  <label v-for="c in columns" :key="c"><input type="checkbox" :checked="!hiddenColumns.includes(c)" @change="toggleColumn(c)"> {{c}}</label>
                  <button class="btn small full" @click="resetColumns">Show all</button>
                </div>
              </div>
              <button class="view-toggle" :class="{active:resultView==='table'}" @click="resultView='table'">Table</button>
              <button class="view-toggle" :class="{active:resultView==='json'}" @click="resultView='json'">JSON</button>
              <button class="btn small" @click="copyResults" :disabled="!currentDocs.length"><Copy/> Copy</button>
              <button class="btn small" @click="toggleResultsExpanded">{{resultsExpanded?'Restore':'Expand'}}</button>
              <template v-if="queryMode==='find'"><button class="icon-btn" @click="prevPage" :disabled="skip===0" title="Previous page">‹</button><button class="icon-btn" @click="nextPage" :disabled="skip+limit>=total" title="Next page">›</button></template>
            </div>
          </div>
          <div v-if="resultView==='table'" class="table-wrap result-scroll">
            <table v-if="currentDocs.length" class="developer-table">
              <thead><tr><th class="selector"></th><th v-for="c in displayColumns" :key="c" :class="{'pinned-id':c==='_id'}" :style="{width:(columnWidths[c]||(c==='_id'?190:160))+'px',minWidth:(columnWidths[c]||(c==='_id'?190:160))+'px'}"><span>{{c}}</span><i class="column-resizer" @mousedown="startColumnResize($event,c)"></i></th><th class="actions-column">Actions</th></tr></thead>
              <tbody>
                <tr v-for="(d,i) in currentDocs" :key="i" :class="{selected:selectedRow===i}" @click="selectedRow=i">
                  <td class="selector"><span class="row-dot" :class="{on:selectedRow===i}"></span></td>
                  <td v-for="c in displayColumns" :key="c" :class="{'pinned-id':c==='_id','cell-editable':rowCanWrite(i)&&c!=='_id'}" :title="shortValue(d[c])" @dblclick.stop="beginCellEdit(i,c)">
                    <div v-if="inlineEdit.row===i&&inlineEdit.field===c" class="inline-cell-editor" @click.stop>
                      <input v-model="inlineEdit.value" autofocus @keyup.enter="saveInlineCell" @keyup.esc="cancelInlineEdit">
                      <button title="Save" @click="saveInlineCell"><Save/></button><button title="Cancel" @click="cancelInlineEdit"><X/></button>
                    </div>
                    <template v-else><span>{{shortValue(d[c])}}</span><small v-if="d[c]&&typeof d[c]==='object'" class="cell-kind">{{Array.isArray(d[c])?'Array':'Object'}}</small></template>
                  </td>
                  <td class="row-actions-cell" @click.stop>
                    <div class="row-actions">
                      <button title="View JSON" @click="openDocumentDrawer(i,'json')"><Code2/></button>
                      <button title="Edit document" @click="openDocumentDrawer(i,'fields')" :disabled="!rowCanWrite(i)"><Pencil/></button>
                      <button title="Duplicate" @click="duplicateDocument(i)" :disabled="!rowCanWrite(i)"><Copy/></button>
                      <button class="danger-text" title="Delete" @click="deleteDocumentAt(i)" :disabled="!rowCanWrite(i)"><Trash2/></button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-else class="empty result-empty"><Search :size="28"/><strong>No documents matched this query</strong><span v-if="queryMode==='find'">Try <code>{}</code> to return every document in {{activeCollection}}.</span><span v-else>Run a different pipeline/query or check the selected collection.</span></div>
          </div>
          <div v-else class="json-results result-scroll">
            <div v-if="currentRawResults.length" class="json-result-list"><pre v-for="(raw,i) in currentRawResults" :key="i" class="json-card"><span class="json-result-number">{{i+1}}</span>{{JSON.stringify(parseEJSON(raw),null,2)}}</pre></div>
            <div v-else class="empty result-empty"><Code2 :size="28"/><strong>No JSON results to display</strong><span>Run a query first. Empty results are no longer shown as a blank editor.</span></div>
          </div>
          <div class="statusbar"><span>{{currentDocs.length}} documents loaded</span><span v-if="queryMode==='sql'">Collection click auto-runs SELECT *</span><span>Query mode: {{queryMode}}</span><span v-if="readOnly"><Shield :size="13"/> Read only</span><span v-else><ShieldCheck :size="13"/> Writes enabled</span></div>
        </div>
      </section>

      <section v-else-if="tab==='Schema'" class="schema-pane">
        <h2>Observed schema</h2><p>Inferred from the currently loaded result set. MongoDB remains schemaless; this view is descriptive, not a validator.</p>
        <table class="schema-table"><thead><tr><th>Field</th><th>Observed type</th><th>Presence</th></tr></thead><tbody><tr v-for="c in columns" :key="c"><td>{{c}}</td><td>{{typeof currentDocs.find(d=>d[c]!==undefined)?.[c]}}</td><td>{{currentDocs.filter(d=>d[c]!==undefined).length}} / {{currentDocs.length}}</td></tr></tbody></table>
      </section>

      <section v-else-if="tab==='Indexes'" class="tool-pane"><div class="tool-heading"><div><h2>Indexes</h2><p>Indexes reported by MongoDB for this collection.</p></div><button class="btn primary" @click="indexModal=true" :disabled="readOnly"><Plus/> Create Index</button></div><pre v-for="(idx,i) in indexes" :key="i" class="json-card">{{JSON.stringify(parseEJSON(idx),null,2)}}</pre></section>

      <section v-else-if="tab==='Laravel'" class="tool-pane"><div class="tool-heading"><div><h2>Laravel Bridge</h2><p>Map this MongoDB collection to Laravel models, preview Eloquent queries, and generate current-style MongoDB Laravel models.</p></div><button class="btn primary" @click="linkLaravel"><FolderOpen/> {{laravel?'Change Project':'Link Project'}}</button></div>
        <div v-if="laravel" :key="'laravel-main-'+laravelLinkEpoch" class="laravel-grid">
          <div class="info-card"><h3>Project checks</h3><p><CheckCircle2 v-if="laravel.isLaravel"/><XCircle v-else/> Laravel {{laravel.laravelVersion||''}}</p><p><CheckCircle2 v-if="laravel.mongoPackageInstalled"/><XCircle v-else/> mongodb/laravel-mongodb {{laravel.mongoPackageVersion}}</p><p><CheckCircle2 v-if="laravel.mongoConfigured"/><XCircle v-else/> MongoDB connection configured</p><p><CheckCircle2 v-if="laravel.models?.length"/><XCircle v-else/> {{laravel.models?.length||0}} model(s) discovered</p><p class="muted project-path">{{laravel.path}}</p><div class="button-row"><button class="btn" @click="refreshLaravel"><RefreshCw/> Rescan Project</button><button class="btn danger" @click="unlinkLaravel"><XCircle/> Unlink Project</button></div></div>
          <div class="info-card"><h3>Current collection → model</h3><template v-if="linkedLaravelModel"><p class="model-map"><CheckCircle2/> <strong>{{linkedLaravelModel.namespace?linkedLaravelModel.namespace+'\\':''}}{{linkedLaravelModel.className}}</strong></p><p>Collection: <code>{{linkedLaravelModel.collection}}</code></p><p>Connection: <code>{{linkedLaravelModel.connection||'Laravel default'}}</code></p><p><CheckCircle2 v-if="linkedLaravelModel.mongoModel"/><XCircle v-else/> {{linkedLaravelModel.mongoModel?'MongoDB Eloquent model':'Not using MongoDB Eloquent base'}}</p><p v-if="linkedLaravelModel.usesLegacyField" class="warning-copy">Uses legacy <code>$collection</code>; current integration uses <code>$table</code>.</p><p class="muted project-path">{{linkedLaravelModel.path}}</p><div v-if="linkedLaravelModel.relationships?.length" class="relation-list"><strong>Relationships</strong><span v-for="r in linkedLaravelModel.relationships" :key="r">{{r}}</span></div><button class="btn accent" @click="tab='Documents';setQueryMode('eloquent')"><Code2/> Query with Eloquent</button></template><template v-else><p class="muted">No linked Laravel model currently maps to <strong>{{activeCollection}}</strong>.</p><label>Class name<input v-model="modelName" :placeholder="classNameFor(activeCollection)"></label><button class="btn accent" @click="generateModel" :disabled="!laravel.mongoPackageInstalled"><Code2/> Generate {{modelName||classNameFor(activeCollection)}}.php</button></template><button class="btn" @click="showLaravelHelp"><Copy/> Setup instructions</button></div>
        </div>
        <div v-else class="info-card empty-laravel"><h3>No Laravel project linked</h3><p>Link a Laravel project to discover models and enable model-aware Eloquent previews.</p><button class="btn primary" @click="linkLaravel"><FolderOpen/> Link Laravel Project</button></div>
        <pre v-if="setupHelp" class="setup-help">{{setupHelp}}<button class="btn small" @click="copyHelp"><Copy/> Copy</button></pre>
      </section>
    </main>

    <aside class="right-pane" v-if="activeCollection">
      <div class="side-card"><h3><Boxes/> Collection Stats</h3><dl><dt>Documents</dt><dd>{{stats.documents}}</dd><dt>Indexes</dt><dd>{{stats.indexes}}</dd><dt>Avg Document</dt><dd>{{stats.avgDocumentKb?.toFixed?.(2)||0}} KB</dd><dt>Total Size</dt><dd>{{stats.totalSizeMb?.toFixed?.(2)||0}} MB</dd><dt>Storage Size</dt><dd>{{stats.storageSizeMb?.toFixed?.(2)||0}} MB</dd></dl></div>
      <div class="side-card"><h3><Layers3/> Quick Actions</h3><button @click="newDocument" :disabled="readOnly"><Plus/> New Document</button><button @click="importJSON" :disabled="readOnly"><Upload/> Import JSON</button><button @click="exportJSON"><Download/> Export JSON</button><button @click="indexModal=true" :disabled="readOnly"><ListFilter/> Create Index</button><button class="danger-text" @click="dropCollection" :disabled="readOnly"><Trash2/> Drop Collection</button></div>
      <div class="side-card"><h3><Waypoints/> Query Workbench</h3><button @click="tab='Documents';setQueryMode('find')"><Search/> Find JSON</button><button @click="tab='Documents';setQueryMode('aggregation')"><Waypoints/> Aggregation</button><button @click="tab='Documents';setQueryMode('sql')"><Code2/> SQL Mode</button><button @click="tab='Documents';setQueryMode('eloquent')"><Code2/> Eloquent Mode</button></div>
      <div class="side-card" :key="'laravel-side-'+laravelLinkEpoch"><h3><Code2/> Laravel Bridge</h3><template v-if="laravel"><p class="check"><CheckCircle2/> Project linked</p><p class="check"><component :is="linkedLaravelModel?CheckCircle2:XCircle"/> {{linkedLaravelModel?linkedLaravelModel.className+' ↔ '+activeCollection:'No model for '+activeCollection}}</p><p class="check"><component :is="laravel.mongoPackageInstalled?CheckCircle2:XCircle"/> MongoDB Laravel package</p><button class="btn accent full" @click="tab='Laravel'"><Code2/> Laravel Details</button><button class="btn full" @click="tab='Documents';setQueryMode('eloquent')"><Code2/> Eloquent Query</button><button class="btn danger full unlink-project" @click="unlinkLaravel"><XCircle/> Unlink Project</button></template><template v-else><p class="muted sidebar-empty-copy">No Laravel project linked.</p><button class="btn full" @click="linkLaravel"><FolderOpen/> Link Laravel Project</button></template></div>
    </aside>
  </div>

  <div v-if="documentDrawer" class="document-drawer-backdrop" @click.self="closeDocumentDrawer">
    <aside class="document-drawer">
      <div class="drawer-head"><div><small>{{activeDatabase}}.{{activeCollection}}</small><h2>{{drawerEditable?'Edit Document':'View Document'}}</h2></div><button class="icon-btn" @click="closeDocumentDrawer"><X/></button></div>
      <div class="drawer-tabs"><button :class="{active:drawerTab==='fields'}" @click="setDrawerTab('fields')">Fields</button><button :class="{active:drawerTab==='json'}" @click="setDrawerTab('json')">JSON</button></div>
      <div v-if="drawerTab==='fields'" class="drawer-fields">
        <div v-for="field in drawerFields" :key="field" class="drawer-field">
          <label>{{field}}<small>{{drawerDocument?.[field]===null?'null':Array.isArray(drawerDocument?.[field])?'array':typeof drawerDocument?.[field]}}</small></label>
          <input v-if="drawerDocument?.[field]===null||typeof drawerDocument?.[field]==='string'" v-model="drawerDocument[field]" :disabled="field==='_id'||!drawerEditable"><input v-else-if="typeof drawerDocument?.[field]==='number'" type="number" v-model.number="drawerDocument[field]" :disabled="!drawerEditable">
          <select v-else-if="typeof drawerDocument?.[field]==='boolean'" v-model="drawerDocument[field]" :disabled="!drawerEditable"><option :value="true">true</option><option :value="false">false</option></select>
          <div v-else class="nested-field"><code>{{shortValue(drawerDocument?.[field])}}</code><button class="btn small" @click="setDrawerTab('json')">Edit in JSON</button></div>
        </div>
      </div>
      <div v-else class="drawer-json-wrap"><textarea v-model="drawerDocumentText" class="drawer-json" spellcheck="false" :readonly="!drawerEditable"></textarea><div v-if="drawerEditable" class="json-validation" :class="{invalid:!!drawerJsonError,valid:!drawerJsonError}"><XCircle v-if="drawerJsonError"/><CheckCircle2 v-else/>{{drawerJsonError || 'Valid JSON document'}}</div></div>
      <div class="drawer-foot"><span v-if="!drawerEditable" class="muted">This result cannot be edited because it is read-only, transformed, or does not include _id.</span><span class="spacer"></span><button class="btn" @click="closeDocumentDrawer">Close</button><button v-if="drawerEditable" class="btn primary" @click="saveDrawerDocument" :disabled="!!drawerJsonError"><Save/> Save Document</button></div>
    </aside>
  </div>

  <div v-if="busy" class="busy"><LoaderCircle class="spin"/> Working…</div>
  <div v-if="notice" class="toast success"><CheckCircle2/>{{notice}}</div>
  <div v-if="error" class="toast error"><XCircle/><span>{{error}}</span><button class="toast-copy" @click="copyError"><Copy/> Copy</button></div>

  <div v-if="connectionModal" class="modal-backdrop" @click.self="connectionModal=false"><div class="modal connection-form"><div class="modal-title"><h2>Add Connection</h2><button class="icon-btn" @click="connectionModal=false"><X/></button></div><div class="form-grid"><label>Connection Name<input v-model="form.name" placeholder="Development"></label><label class="span2">URI<input v-model="form.uri" placeholder="mongodb://localhost:27017/"></label><label>Host fallback<input v-model="form.host"></label><label>Port<input type="number" v-model.number="form.port"></label><label>Username<input v-model="form.username" autocomplete="off"></label><label>Password<input type="password" v-model="form.password" autocomplete="new-password" placeholder="Stored securely"></label><label>Auth DB<input v-model="form.authDb" placeholder="admin"></label><label class="check-label"><input type="checkbox" v-model="form.tls"> TLS / SSL</label><label class="check-label"><input type="checkbox" v-model="form.readOnly"> Read Only</label></div><div class="modal-actions"><button class="btn" @click="testConnection"><KeyRound/> Test Connection</button><span class="spacer"></span><button class="btn" @click="connectionModal=false">Cancel</button><button class="btn primary" @click="saveConnection"><Save/> Save</button></div></div></div>

  <div v-if="documentModal" class="modal-backdrop" @click.self="closeDocumentModal"><div class="modal editor-modal"><div class="modal-title"><h2>{{editingDocument?'Edit':'New'}} Document</h2><button class="icon-btn" @click="closeDocumentModal"><X/></button></div><textarea v-model="documentText" class="big-editor doc-editor" spellcheck="false"></textarea><div class="json-validation modal-json-validation" :class="{invalid:!!documentJsonError,valid:!documentJsonError}"><XCircle v-if="documentJsonError"/><CheckCircle2 v-else/>{{documentJsonError || 'Valid JSON document'}}</div><div class="modal-actions"><span class="spacer"></span><button class="btn" @click="closeDocumentModal">Cancel</button><button class="btn primary" @click="saveDocument" :disabled="!!documentJsonError"><Save/> Save Document</button></div></div></div>

  <div v-if="indexModal" class="modal-backdrop" @click.self="indexModal=false"><div class="modal small-modal"><div class="modal-title"><h2>Create Index</h2><button class="icon-btn" @click="indexModal=false"><X/></button></div><label>Keys EJSON<textarea v-model="indexForm.keys" class="mini-editor"></textarea></label><label>Index name (optional)<input v-model="indexForm.name"></label><label class="check-label"><input type="checkbox" v-model="indexForm.unique"> Unique index</label><div class="modal-actions"><span class="spacer"></span><button class="btn" @click="indexModal=false">Cancel</button><button class="btn primary" @click="createIndex">Create</button></div></div></div>

  <div v-if="createModal" class="modal-backdrop" @click.self="createModal=false"><div class="modal small-modal"><div class="modal-title"><h2>Create {{createMode}}</h2><button class="icon-btn" @click="createModal=false"><X/></button></div><label>Name<input v-model="createName" @keyup.enter="doCreate" autofocus></label><p v-if="createMode==='database'" class="muted">MongoDB creates databases lazily, so an initial collection named “initial” will be created.</p><div class="modal-actions"><span class="spacer"></span><button class="btn" @click="createModal=false">Cancel</button><button class="btn primary" @click="doCreate">Create</button></div></div></div>

  <div v-if="backupModal" class="modal-backdrop" @click.self="backupModal=false"><div class="modal backup-modal">
    <div class="modal-title"><div><h2>Backup & Restore</h2><small>Timestamped ZIP snapshots of user databases, BSON documents and indexes.</small></div><button class="icon-btn" @click="backupModal=false"><X/></button></div>
    <div class="backup-connection-strip"><Server/><div><strong>{{activeProfile?.name||'No active connection'}}</strong><small v-if="activeProfile">{{readOnly?'Backup only · restore disabled on read-only connection':'Backup and restore available'}}</small><small v-else>Select a connection to create or restore backups.</small></div><span class="spacer"></span><button class="btn" @click="loadBackups"><RefreshCw/> Refresh list</button></div>
    <div class="backup-create-grid">
      <section class="backup-card"><div><strong>Back up everything</strong><p>All user databases on this connection. MongoDB system databases <code>admin</code>, <code>config</code> and <code>local</code> are excluded.</p></div><button class="btn primary" @click="createBackup('all')" :disabled="!activeConnection"><Download/> Backup All</button></section>
      <section class="backup-card"><div><strong>Back up one database</strong><p>Create a ZIP snapshot of one database without touching the others.</p><select v-model="backupDatabase" :disabled="!activeConnection"><option value="" disabled>Choose database…</option><option v-for="db in backupDatabases" :key="db.name" :value="db.name">{{db.name}}</option></select></div><button class="btn primary" @click="createBackup('database')" :disabled="!activeConnection||!backupDatabase"><Download/> Backup Database</button></section>
    </div>
    <div class="backup-note"><ShieldCheck/><span>Every new backup is kept in Mongo Developer's local backup store with a timestamped filename. After creation, a Save As window also offers a second ZIP copy wherever you choose.<code v-if="backupDirectory">{{backupDirectory}}</code></span></div>
    <div class="backup-list-heading"><div><h3>Stored backups</h3><small>{{backups.length}} backup{{backups.length===1?'':'s'}} saved on this computer</small></div><button class="btn" @click="restoreExternalBackup" :disabled="!activeConnection||readOnly"><Upload/> Restore from ZIP…</button></div>
    <div class="backup-list" v-if="backups.length">
      <article v-for="item in backups" :key="item.fileName" class="backup-row">
        <div class="backup-main"><strong>{{item.fileName}}</strong><small>{{backupDate(item.createdAt)}} · {{item.scope==='all'?'All user databases':item.database}} · {{backupSize(item.sizeBytes)}}</small><span>{{item.databaseCount}} DB · {{item.collectionCount}} collections · {{item.documentCount}} documents · {{item.indexCount}} indexes<span v-if="item.sourceName"> · {{item.sourceName}}</span></span></div>
        <div class="backup-actions"><button class="btn small" @click="downloadBackup(item)"><Download/> Download Copy</button><button class="btn small accent" @click="restoreStoredBackup(item)" :disabled="!activeConnection||readOnly"><Upload/> Restore</button><button class="btn small danger" @click="deleteBackup(item)"><Trash2/> Delete ZIP</button></div>
      </article>
    </div>
    <div v-else class="backup-empty"><Database/><strong>No stored backups yet</strong><span>Create an all-database or single-database backup above.</span></div>
    <div class="modal-actions"><span class="spacer"></span><button class="btn" @click="backupModal=false">Close</button></div>
  </div></div>

  <div v-if="savedQueriesModal" class="modal-backdrop" @click.self="savedQueriesModal=false"><div class="modal saved-query-modal">
    <div class="modal-title"><div><h2>Saved Queries</h2><small>Keep useful SQL, Find JSON, Aggregation and Eloquent queries on this computer.</small></div><button class="icon-btn" @click="savedQueriesModal=false"><X/></button></div>
    <div class="saved-query-create" v-if="activeCollection"><input v-model="savedQueryName" placeholder="Query name" @keyup.enter="saveCurrentQuery"><button class="btn primary" @click="saveCurrentQuery"><Save/> Save Current Query</button></div>
    <div class="saved-query-list" v-if="savedQueries.length"><article v-for="item in savedQueries" :key="item.id" class="saved-query-row"><div><strong>{{item.name}}</strong><small>{{item.mode.toUpperCase()}} · {{item.database}}.{{item.collection}} · {{new Date(item.createdAt).toLocaleString()}}</small><code>{{item.text.replace(/\s+/g,' ').slice(0,120)}}</code></div><div class="saved-query-actions"><button class="btn small accent" @click="loadSavedQuery(item)"><Play/> Load</button><button class="btn small danger" @click="deleteSavedQuery(item)"><Trash2/> Delete</button></div></article></div>
    <div v-else class="backup-empty"><Save/><strong>No saved queries yet</strong><span>Save the current workbench query above.</span></div>
    <div class="modal-actions"><span class="spacer"></span><button class="btn" @click="savedQueriesModal=false">Close</button></div>
  </div></div>

  <div v-if="explainModal" class="modal-backdrop" @click.self="explainModal=false"><div class="modal explain-modal">
    <div class="modal-title"><div><h2>Explain Query</h2><small>{{explainResult?.database}}.{{explainResult?.collection}} · {{String(explainResult?.mode||'').toUpperCase()}}</small></div><button class="icon-btn" @click="explainModal=false"><X/></button></div>
    <div class="explain-summary" :class="{good:explainResult?.usesIndex,warn:explainResult?.collectionScan}"><component :is="explainResult?.usesIndex?CheckCircle2:XCircle"/><div><strong>{{explainResult?.summary}}</strong><small>Winning stage: {{explainResult?.stage||'UNKNOWN'}}<span v-if="explainResult?.indexName"> · {{explainResult.indexName}}</span></small></div></div>
    <div class="explain-stats"><div><span>Documents examined</span><strong>{{explainResult?.documentsExamined??0}}</strong></div><div><span>Keys examined</span><strong>{{explainResult?.keysExamined??0}}</strong></div><div><span>Returned</span><strong>{{explainResult?.returned??0}}</strong></div><div><span>Execution</span><strong>{{explainResult?.executionMs??0}} ms</strong></div></div>
    <div v-if="explainResult?.suggestedIndex" class="index-advice"><strong>Suggested index</strong><p>Consider this index if the query is frequent and selective. Review it before creating it.</p><code>{{explainResult.suggestedIndex}}</code><div class="button-row"><button class="btn" @click="copySuggestedIndex"><Copy/> Copy</button><button class="btn accent" @click="prepareSuggestedIndex"><Plus/> Review in Create Index</button></div></div>
    <div v-else-if="explainResult?.usesIndex" class="safe-banner"><ShieldCheck/> This plan is already using an index.</div>
    <div class="modal-actions"><span class="spacer"></span><button class="btn" @click="explainModal=false">Close</button></div>
  </div></div>

  <div v-if="diagnosticsModal" class="modal-backdrop" @click.self="diagnosticsModal=false"><div class="modal diagnostics-modal">
    <div class="modal-title"><div><h2>Diagnostics</h2><small>Release/support information with no stored password exposed.</small></div><button class="icon-btn" @click="diagnosticsModal=false"><X/></button></div>
    <div v-if="diagnostics" class="diagnostics-grid"><div><span>Version</span><strong>{{diagnostics.version}}</strong></div><div><span>Platform</span><strong>{{diagnostics.os}}/{{diagnostics.arch}}</strong></div><div><span>Connection</span><strong>{{diagnostics.connectionName||'None selected'}}</strong></div><div><span>Status</span><strong>{{diagnostics.connected?`Connected · ${diagnostics.pingMs} ms`:'Not connected'}}</strong></div><div><span>MongoDB</span><strong>{{diagnostics.mongoDbVersion||'—'}}</strong></div><div><span>Access</span><strong>{{diagnostics.readOnly?'Read only':'Read/write'}}</strong></div></div>
    <pre class="diagnostics-text">{{diagnosticsText}}</pre>
    <div class="modal-actions"><button class="btn" @click="openDiagnostics"><RefreshCw/> Refresh</button><span class="spacer"></span><button class="btn" @click="diagnosticsModal=false">Close</button><button class="btn primary" @click="copyDiagnostics"><Copy/> Copy Diagnostics</button></div>
  </div></div>

  <div v-if="confirmModal" class="modal-backdrop" @click.self="confirmModal=false"><div class="modal confirm-modal"><div class="modal-title"><h2>{{confirmTitle}}</h2><button class="icon-btn" @click="confirmModal=false"><X/></button></div><p class="confirm-copy">{{confirmMessage}}</p><div v-if="confirmDanger" class="danger-banner"><Shield/> Destructive action — this cannot be undone.</div><div v-else class="safe-banner"><ShieldCheck/> Unsaved work will be discarded; MongoDB data is not being deleted.</div><div class="modal-actions"><span class="spacer"></span><button class="btn" @click="confirmModal=false">Cancel</button><button class="btn" :class="confirmDanger?'danger-solid':'primary'" @click="confirmYes"><Trash2 v-if="confirmDanger"/><CheckCircle2 v-else/> {{confirmActionLabel}}</button></div></div></div>
</div>
</template>
