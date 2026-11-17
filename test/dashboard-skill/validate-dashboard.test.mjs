import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync, mkdtempSync, writeFileSync, rmSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { spawnSync } from 'node:child_process';
import { validateDashboard, COMPONENTS } from '../../skills/lingtong-dashboard/scripts/validate-dashboard.mjs';

const skill = new URL('../../skills/lingtong-dashboard/', import.meta.url);
const examples = new URL('assets/examples/', skill);
const load = name => JSON.parse(readFileSync(new URL(`${name}.json`, examples), 'utf8'));
const schemas = load('metadata');
const options = { schemas, pivotConfigs: load('pivot-configs'), requireBindings: true };
const run = (edit, opts = options) => { const v = load('operating-overview'); edit(v.config, v); return validateDashboard(v, opts); };
const component = (c, type) => c.tableConfigList.find(v => v.displayType === type);
const fails = (result, path) => { assert.equal(result.ok, false, JSON.stringify(result)); assert.ok(result.errors.some(e => e.path.includes(path)), JSON.stringify(result)); };

for (const name of readdirSync(examples).filter(n => n.endsWith('.json') && !['metadata.json', 'pivot-configs.json'].includes(n))) {
  test(`example ${name}`, () => {
    const input = JSON.parse(readFileSync(new URL(name, examples), 'utf8'));
    const result = validateDashboard(input, name === 'operating-overview.json' ? options : { schemas });
    assert.equal(result.ok, true, JSON.stringify(result.errors));
  });
}
test('business overview covers all nine registered components and metadata', () => {
  const result = validateDashboard(load('operating-overview'), options);
  assert.deepEqual(Object.keys(result.summary.counts).sort(), Object.keys(COMPONENTS).sort());
  assert.equal(result.summary.components, 12);
  assert.deepEqual(result.warnings, []);
});
test('standalone config and optional metadata report their validation scope', () => {
  const result = validateDashboard(load('bar-chart').config);
  assert.equal(result.ok, true);
  assert.equal(result.summary.metadataChecked, false);
  assert.ok(result.warnings.some(w => w.path === '$schemas'));
});
test('each component rejects its wrong category', () => {
  for (const type of Object.keys(COMPONENTS)) fails(run(c => { component(c, type).type = 'invalid'; }), '.type');
});
test('unknown display types and duplicate IDs cannot silently render', () => {
  fails(run(c => { c.tableConfigList[0].displayType = 'combinedChart'; }), '.displayType');
  fails(run(c => { c.tableConfigList[1].id = c.tableConfigList[0].id; }), '.id');
  fails(run(c => { c.layoutList[1].i = c.layoutList[0].i; }), '.i');
  fails(run(c => { c.tableConfigList[0].id = 123; }), '.id');
});
test('layout checks mapping, overlap, bounds, integer size and min/max', () => {
  fails(run(c => { c.layoutList.pop(); }), '.layoutList');
  fails(run(c => { c.layoutList[0].i = 'missing'; }), '.i');
  fails(run(c => { c.layoutList[0].w = 25; }), '.layoutList');
  fails(run(c => { c.layoutList[0].h = 0; }), '.h');
  fails(run(c => { c.layoutList[0].x = -1; }), '.x');
  fails(run(c => { c.layoutList[0].y = 1.5; }), '.y');
  fails(run(c => { c.layoutList[1].x = 0; }), '.layoutList');
  fails(run(c => { c.layoutList[0].minW = 7; }), '.minW');
  fails(run(c => { c.layoutList[0].maxH = 4; }), '.maxH');
});
test('all five charts validate active measures and field types', () => {
  for (const type of ['barCharts', 'HBarCharts', 'lineCharts', 'areaCharts', 'pieCharts']) {
    fails(run(c => { component(c, type).dataConfig.measures = []; }), '.measures');
    fails(run(c => { component(c, type).dataConfig.measures[0].field = 'data2001.1'; }), '.field');
    fails(run(c => { component(c, type).dataConfig.measures[0].field = 'invented'; }), '.field');
    fails(run(c => { component(c, type).dataConfig.measures[0].disabled = true; }), '.measures');
  }
});
test('pie has one measure; cartesian charts have three and role limits', () => {
  fails(run(c => { const d=component(c,'pieCharts').dataConfig; d.measures.push({...d.measures[0],id:'m2'}); }), '.measures');
  fails(run(c => { const d=component(c,'barCharts').dataConfig; d.measures=Array.from({length:4},(_,i)=>({...d.measures[0],id:`m${i}`})); }), '.measures');
  fails(run(c => { const d=component(c,'barCharts').dataConfig;d.dimensions.push({...d.dimensions[0],id:'d2'}); }), '.displayModel');
  fails(run(c => { component(c,'HBarCharts').dataConfig.measures[0].styleConfig.valueAxisPosition='right'; }), '.valueAxisPosition');
  assert.equal(run(c => { component(c,'barCharts').styleConfig.legendPosition='hidden'; }).ok,true);
});
test('dimension types, sort references, time intervals and ranges', () => {
  fails(run(c => { component(c,'barCharts').dataConfig.dimensions[0].config.sortByMeasureId='missing'; }), '.sortByMeasureId');
  fails(run(c => { component(c,'barCharts').dataConfig.dimensions[0].config.sortCount=0; }), '.sortCount');
  fails(run(c => { component(c,'lineCharts').dataConfig.dimensions[0].config.miniInterval='0d'; }), '.miniInterval');
  fails(run(c => { component(c,'lineCharts').dataConfig.dimensions[0].field='data2001.2'; }), '.field');
  for (const config of [{rangeValueList:[{min:'9',max:'1'}]},{rangeValueList:[{min:'',max:'1'}]}]) {
    fails(run(c => {const d=component(c,'barCharts').dataConfig.dimensions[0];Object.assign(d,{type:'range',field:'data2001.2',config});}),'.rangeValueList');
  }
  fails(run(c => {const d=component(c,'lineCharts').dataConfig.dimensions[0];Object.assign(d,{type:'date_range',config:{dateRangeValueList:[{start:'2026-02-30',end:'2026-03-01'}]}});}),'.dateRangeValueList');
  fails(run(c => {const m=component(c,'barCharts').dataConfig.measures[0];m.type='percent';m.config={percents:[101]};}),'.percents');
});
test('new metric format requires explicit switch, source identity, numeric fields and count limits', () => {
  fails(run(c => { delete c.tableConfigList[0].dataConfig.multiMetricEnabled; }), '.source');
  fails(run(c => { c.tableConfigList[0].dataConfig.sources=[]; }), '.sources');
  fails(run(c => { const d=c.tableConfigList[0].dataConfig;d.sources=Array.from({length:21},(_,i)=>({...d.sources[0],id:`source${i}`})); }), '.sources');
  fails(run(c => { const d=c.tableConfigList[1].dataConfig;d.sources[1].id=d.sources[0].id; }), '.id');
  fails(run(c => { c.tableConfigList[0].dataConfig.sources[0].aggregation.field='data2001.3'; }), '.field');
  fails(run(c => { c.tableConfigList[0].dataConfig.source={}; }), '.dataConfig');
});
test('metric comparisons, display formats and percentages', () => {
  fails(run(c => { c.tableConfigList[0].dataConfig.sources[0].compareConfig.dateField='data2001.2'; }), '.dateField');
  fails(run(c => { const comp=c.tableConfigList[0].dataConfig.sources[0].compareConfig;comp.timeRangeType='recentDays';comp.amount=0; }), '.amount');
  fails(run(c => { const comp=c.tableConfigList[0].dataConfig.sources[0].compareConfig;comp.timeRangeType='custom';comp.customRange={start:'2026-02-30',end:'2026-03-01'}; }), '.customRange');
  fails(run(c => { c.tableConfigList[1].dataConfig.multiMetricConfig.percentage.denominatorSourceId='missing'; }), '.denominatorSourceId');
  fails(run(c => { c.tableConfigList[0].dataConfig.sources[0].styleConfig.precision=10; }), '.precision');
  fails(run(c => { c.tableConfigList[0].dataConfig.sources[0].styleConfig.description='x'.repeat(201); }), '.description');
  const result=run(c => { c.tableConfigList[1].dataConfig.sources[0].compareConfig.enabled=true; });
  assert.equal(result.ok,true); assert.ok(result.warnings.some(w=>w.path.includes('compareConfig')));
});
test('slicer must reference an active source, including multi-source tables', () => {
  fails(run(c => { component(c,'slicer').tableId=9999; }), '.tableId');
  const edit=(c)=>{const s=component(c,'slicer');s.tableId=1002;Object.assign(s.dataConfig,{field:'data2002.1',fieldType:'number',style:'number',operator:'eq',value:null,defaultValue:null});};
  assert.equal(run(edit).ok,true);
  fails(run(c=>{edit(c);c.tableConfigList[1].dataConfig.multiMetricEnabled=false;}),'.tableId');
});
test('slicer styles, selection modes and static-filter operator distinction', () => {
  fails(run(c=>{component(c,'slicer').dataConfig.optionLimit=100;}),'.optionLimit');
  fails(run(c=>{component(c,'slicer').dataConfig.field='data2001.7';}),'.field');
  fails(run(c=>{Object.assign(component(c,'slicer').dataConfig,{style:'search',selectionMode:'multiple'});}),'.selectionMode');
  fails(run(c=>{Object.assign(component(c,'slicer').dataConfig,{selectionMode:'multiple',value:'shop'});}),'.value');
  fails(run(c=>{c.tableConfigList.find(v=>v.id==='slicer_amount').dataConfig.operator='neq';}),'.operator');
  fails(run(c=>{component(c,'slicer').tableFilters=[structuredClone(c.globalFilterList[0])];}),'.tableFilters');
});
test('number and date slicers reject incomplete, inverted or impossible ranges', () => {
  for(const value of [[5,null],[10,1],['1','2'],[]]) fails(run(c=>{c.tableConfigList.find(v=>v.id==='slicer_amount').dataConfig.value=value;}),'.value');
  fails(run(c=>{const d=c.tableConfigList.find(v=>v.id==='slicer_date').dataConfig;d.value='recentDays';d.amount=1.5;}),'.amount');
  fails(run(c=>{const d=c.tableConfigList.find(v=>v.id==='slicer_date').dataConfig;d.value='custom';d.customRange={start:'2026-11-10',end:'2026-11-01'};}),'.customRange');
  fails(run(c=>{const d=c.tableConfigList.find(v=>v.id==='slicer_date').dataConfig;d.operator='eq';d.value=d.defaultValue='2026-02-30';}),'.value');
  assert.equal(run(c=>{const d=c.tableConfigList.find(v=>v.id==='slicer_date').dataConfig;d.value=['2024-02-29','2024-03-01'];}).ok,true);
});
test('global and per-source filters validate schema identity and persisted format', () => {
  fails(run(c=>{c.globalFilterList[0].schemaId=1001;}),'.schemaId');
  fails(run(c=>{c.globalFilterList[0].columnId=999;}),'.columnId');
  fails(run(c=>{c.globalFilterList[0].value=['a','b'];}),'.value');
  fails(run(c=>{c.globalFilterList[0].comparator='ne';}),'.comparator');
  fails(run(c=>{c.globalFilterList[0].columnId=7;}),'.columnId');
  fails(run(c=>{c.tableConfigList[1].dataConfig.sources[1].tableFilters=[c.globalFilterList[0]];}),'.tableId');
  assert.equal(run(c=>{Object.assign(c.globalFilterList[0],{comparator:'empty',value:''});}).ok,true);
});
test('views have separate pre-save and post-save validation', () => {
  const input=load('pivot-table');
  const before=validateDashboard(input,{schemas});assert.equal(before.ok,true);assert.ok(before.warnings.some(w=>w.path.includes('tableViewId')));
  fails(validateDashboard(input,{schemas,requireBindings:true}),'.tableViewId');
  fails(run(c=>{component(c,'basicTable').tableViewId='8002';}),'.tableViewId');
  fails(run(c=>{component(c,'pivotTable').tableViewId='8001';}),'.tableViewId');
  fails(run(c=>{component(c,'basicTable').dataConfig.allowEditData='false';}),'.allowEditData');
});
test('pivot statistics use bound views, distinct date keys and valid totals/summary', () => {
  const execute=edit=>{const pivotConfigs=load('pivot-configs');edit(pivotConfigs[0]);return validateDashboard(load('operating-overview'),{schemas,pivotConfigs});};
  fails(execute(p=>{p.viewId='9999';}),'$pivotConfigs');
  fails(execute(p=>{p.businessId=2001;}),'$pivotConfigs');
  fails(execute(p=>{p.config.advanced.colDimensionId='missing';}),'.colDimensionId');
  fails(execute(p=>{p.config.advanced.totalMeasures[0].type='count';}),'.type');
  fails(execute(p=>{p.config.other.otherExtractConfigVOList[0].field='data2001.1';}),'.otherExtractConfigVOList');
  fails(execute(p=>{p.config.other.maxRowsPerPage=0;}),'.maxRowsPerPage');
  fails(execute(p=>{Object.assign(p.config.dimensions[0],{type:'date_range',field:'data2001.3',config:{dateRangeValueList:[{start:'2026-01-01',end:'2026-01-31'}]}});}),'.dateList');
  assert.equal(execute(p=>{Object.assign(p.config.dimensions[0],{type:'range',field:'data2001.2',config:{rangeValueList:[{min:'',max:100}]}});}).ok,true);
});
test('unknown extension fields survive and validation never mutates input', () => {
  const input=load('operating-overview');input.config.futureFeature={keep:true};
  const snapshot=JSON.stringify(input);const meta=JSON.stringify(options);
  assert.equal(validateDashboard(input,options).ok,true);
  assert.equal(JSON.stringify(input),snapshot);assert.equal(JSON.stringify(options),meta);
});
test('single metrics preserve inactive sources and validate them when re-enabled', () => {
  const input=load('metric-card');
  const d=input.config.tableConfigList[0].dataConfig;
  const inactive=structuredClone(d.sources[0]);
  inactive.id='saved_for_later';
  inactive.tableId='';
  inactive.aggregation.field='unfinished';
  d.sources.push(inactive);
  const before=JSON.stringify(input);
  const single=validateDashboard(input,{schemas});
  assert.equal(single.ok,true,JSON.stringify(single.errors));
  assert.ok(single.warnings.some(w=>w.path.endsWith('.sources[1]')));
  assert.equal(JSON.stringify(input),before);
  d.multiMetricEnabled=true;
  fails(validateDashboard(input,{schemas}),'.sources[1].tableId');
  d.multiMetricEnabled=false;
  d.sources[1].id=d.sources[0].id;
  fails(validateDashboard(input,{schemas}),'.id');
});
test('custom percentage denominator is required only in active multi-metric mode', () => {
  const input=load('metric-card');
  const d=input.config.tableConfigList[0].dataConfig;
  d.multiMetricConfig.percentage={mode:'custom',denominatorSourceId:null};
  assert.equal(validateDashboard(input,{schemas}).ok,true);
  d.multiMetricEnabled=true;
  fails(validateDashboard(input,{schemas}),'.denominatorSourceId');
});
test('malformed nested JSON reports errors rather than throwing', () => {
  for (const value of [null,[],1,'text',false,{}, {config:null}]) assert.doesNotThrow(()=>{assert.equal(validateDashboard(value).ok,false);});
  for(const edit of [c=>{c.tableConfigList=[null];},c=>{c.layoutList=[null];},c=>{c.globalFilterList=[null];},c=>{c.tableConfigList[0].dataConfig.sources=[null];},c=>{component(c,'barCharts').dataConfig.measures=[null];},c=>{component(c,'barCharts').dataConfig.dimensions=[null];},c=>{c.tableConfigList[0].dataConfig=null;}]) assert.doesNotThrow(()=>{assert.equal(run(edit).ok,false);});
});
test('metadata cannot invent fields or duplicate identities',()=>{
 const meta=structuredClone(schemas);meta.tables[0].columnsSchema.push({...meta.tables[0].columnsSchema[0]});
 fails(validateDashboard(load('operating-overview'),{schemas:meta}),'$schemas.tables');
});
test('CLI uses JSON output and exit status 2 for file, parse, validation and argument errors', () => {
  const dir=mkdtempSync(join(tmpdir(),'dashboard-validator-test-'));
  const script=fileURLToPath(new URL('scripts/validate-dashboard.mjs',skill));
  try {
    const valid=join(dir,'valid.json'),invalid=join(dir,'invalid.json'),broken=join(dir,'broken.json');
    writeFileSync(valid,JSON.stringify(load('bar-chart')));writeFileSync(invalid,'{}');writeFileSync(broken,'{bad');
    for(const args of [[],[invalid],[broken],[join(dir,'missing.json')],[valid,'--bad'],[valid,'--schemas']]) {
      const result=spawnSync(process.execPath,[script,...args],{encoding:'utf8'});
      assert.equal(result.status,2,result.stderr);assert.equal(JSON.parse(result.stdout).ok,false);
    }
    const result=spawnSync(process.execPath,[script,valid],{encoding:'utf8'});assert.equal(result.status,0,result.stderr);assert.equal(JSON.parse(result.stdout).ok,true);
  } finally {rmSync(dir,{recursive:true,force:true});}
});
