import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync, existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join, dirname, resolve } from 'node:path';
import { COMPONENTS } from '../../skills/lingtong-dashboard/scripts/validate-dashboard.mjs';
const root=fileURLToPath(new URL('../../',import.meta.url));
const skill=join(root,'skills/lingtong-dashboard');
function* files(dir){for(const ent of readdirSync(dir,{withFileTypes:true})){const path=join(dir,ent.name);if(ent.isDirectory())yield* files(path);else yield path;}}

test('every local skill reference resolves in the distributable skills tree',()=>{
 for(const file of files(skill)){
  if(!file.endsWith('.md'))continue;
  const md=readFileSync(file,'utf8');
  for(const match of md.matchAll(/\[[^\]]+\]\(([^)]+)\)/g)){
   const target=match[1].split('#')[0];if(!target||/^[a-z]+:/i.test(target))continue;
   assert.equal(existsSync(resolve(dirname(file),target)),true,`${file}: ${target}`);
  }
 }
});
test('embedded API contract covers all renderers and scoped dashboard filter metadata',()=>{
 const spec=JSON.parse(readFileSync(join(root,'internal/openapi/spec/openapi.json'),'utf8'));
 const defs=spec.definitions;
 assert.deepEqual([...defs.LtDashBoardTableBaseConfig.properties.displayType.enum].sort(),Object.keys(COMPONENTS).sort());
 assert.equal(defs.LtDashboardConfig.properties.globalFilterList.items.$ref,'#/definitions/LtDashboardFilterCondition');
 assert.equal(defs.LtDashBoardTableBaseConfig.properties.tableFilters.items.$ref,'#/definitions/LtDashboardFilterCondition');
 assert.ok(defs.LtDashboardFilterCondition.properties.schemaId);
 assert.ok(defs.LtDashboardFilterCondition.properties.tableId);
 const operation=spec.paths['/dashboard/query'];assert.equal(operation.get,undefined);assert.ok(operation.post);
 assert.equal(operation.post.parameters[0].in,'body');assert.deepEqual(operation.post.parameters[0].schema.required,['id']);
 const inspect=value=>{if(!value||typeof value!=='object')return;if(value.$ref?.startsWith('#/definitions/'))assert.ok(defs[value.$ref.slice(14)],value.$ref);for(const child of Object.values(value))inspect(child);};
 for(const [name,value] of Object.entries(defs))if(name.startsWith('LtDashboard')||name==='LtDashBoardTableBaseConfig')inspect(value);
});
