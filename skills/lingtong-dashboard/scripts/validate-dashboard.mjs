#!/usr/bin/env node
// Offline validation only. No network, writes, coercion, or migration of input.
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export const COMPONENTS = Object.freeze({
  metricCard: 'chart', barCharts: 'chart', HBarCharts: 'chart', lineCharts: 'chart',
  areaCharts: 'chart', pieCharts: 'chart', basicTable: 'view', pivotTable: 'view', slicer: 'other',
});
const AGGREGATIONS = ['count', 'sum', 'avg', 'std', 'unique', 'valueCount', 'max', 'min', 'median', 'percent', 'percentRank'];
const METRIC_AGGREGATIONS = AGGREGATIONS.filter(v => !['percent', 'percentRank'].includes(v));
const FORMATS = ['numberThousands', 'number', 'percent', 'cny', 'usd'];
const THEMES = ['default', 'forest', 'ocean', 'sunset', 'pastel', 'jewel', 'berry', 'blues', 'teal', 'emerald'];
const TIME_RANGES = ['today', 'yesterday', 'thisWeek', 'thisWeekToDate', 'lastWeek', 'thisMonth', 'thisMonthToDate', 'lastMonth', 'thisQuarter', 'thisQuarterToDate', 'lastQuarter', 'thisYear', 'thisYearToDate', 'lastYear', 'recentDays', 'recentMonths', 'custom'];
const FILTER_OPS = {
  text: ['eq', 'neq', 'ct', 'nct', 'empty', 'notempty'],
  number: ['eq', 'neq', 'gt', 'gte', 'lt', 'lte', 'empty', 'notempty'],
  date: ['eq', 'lt', 'gt', 'empty', 'notempty'],
  table: ['empty', 'notempty'], password: ['empty', 'notempty'],
};
const own = (o, k) => Object.prototype.hasOwnProperty.call(o ?? {}, k);
const object = v => v !== null && typeof v === 'object' && !Array.isArray(v);
const text = v => typeof v === 'string' && v.trim().length > 0;
const blank = v => v === null || v === undefined || v === '';
const positiveId = v => (Number.isSafeInteger(v) && v > 0) || (typeof v === 'string' && /^[1-9]\d*$/.test(v));
const numeric = v => (typeof v === 'number' || text(v)) && Number.isFinite(Number(v));
const positiveInt = v => numeric(v) && Number.isSafeInteger(Number(v)) && Number(v) > 0;
const date = v => {
  if (typeof v !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(v)) return false;
  const d = new Date(`${v}T00:00:00Z`);
  return Number.isFinite(d.getTime()) && d.toISOString().slice(0, 10) === v;
};
const columnType = c => {
  const t = c?.type === 'function' ? c.functionFieldType : c?.type;
  if (['number', 'integer', 'double'].includes(t)) return 'number';
  if (['date', 'createDate', 'updateDate'].includes(t)) return 'date';
  if (['table', 'password', 'multiSelect'].includes(t)) return t;
  return 'text';
};

/** Input: a dashboard config or {appId,id?,name?,config}; options are documented in validation.md. */
export function validateDashboard(input, { schemas, pivotConfigs, requireBindings = false } = {}) {
  const errors = [], warnings = [], counts = {};
  const error = (path, message) => errors.push({ path, message });
  const warn = (path, message) => warnings.push({ path, message });
  const record = (v, p) => {
    if (!object(v)) { error(p, '必须是 JSON 对象'); return {}; }
    return v;
  };
  const array = (v, p) => {
    if (!Array.isArray(v)) { error(p, '必须是数组'); return []; }
    return v;
  };
  const enumeration = (v, values, p, optional = false) => {
    if (optional && v === undefined) return;
    if (!values.includes(v)) error(p, `应为 ${values.join(' | ')}`);
  };
  const bools = (v, keys, p) => keys.forEach(k => {
    if (v[k] !== undefined && typeof v[k] !== 'boolean') error(`${p}.${k}`, '必须是布尔值');
  });
  const id = (v, p) => { if (!positiveId(v)) error(p, '必须是正整数 ID（大整数使用十进制字符串）'); };
  const unique = (items, p, key = 'id') => {
    const seen = new Set();
    items.forEach((v, i) => {
      if (!object(v)) { error(`${p}[${i}]`, '必须是对象'); return; }
      if (!text(v[key])) error(`${p}[${i}].${key}`, '必须是非空字符串');
      else if (seen.has(v[key])) error(`${p}[${i}].${key}`, 'ID 重复');
      seen.add(v[key]);
    });
  };
  const raw = record(input, '$');
  const wrapped = own(raw, 'config');
  const root = wrapped ? '$.config' : '$';
  const cfg = wrapped ? record(raw.config, root) : raw;
  if (wrapped) {
    id(raw.appId, '$.appId');
    if (raw.id !== undefined) id(raw.id, '$.id');
    if (raw.name !== undefined && !text(raw.name)) error('$.name', '名称不能为空');
  }
  const charts = array(cfg.tableConfigList, `${root}.tableConfigList`);
  const layouts = array(cfg.layoutList, `${root}.layoutList`);
  unique(charts, `${root}.tableConfigList`);
  unique(layouts, `${root}.layoutList`, 'i');
  const tables = new Map(), views = new Map();
  if (schemas !== undefined) {
    const meta = record(schemas, '$schemas');
    array(meta.tables, '$schemas.tables').forEach((v, i) => {
      const p = `$schemas.tables[${i}]`, t = record(v, p);
      id(t.id, `${p}.id`); id(t.schemaId, `${p}.schemaId`);
      if (tables.has(String(t.id))) error(`${p}.id`, '表格 ID 重复');
      const columns = array(t.columnsSchema, `${p}.columnsSchema`);
      const columnIds = new Set(), keys = new Set();
      columns.forEach((v, j) => {
        const c = record(v, `${p}.columnsSchema[${j}]`);
        id(c.id, `${p}.columnsSchema[${j}].id`);
        if (!text(c.key) || keys.has(c.key)) error(`${p}.columnsSchema[${j}].key`, '字段 key 缺失或重复');
        if (columnIds.has(String(c.id))) error(`${p}.columnsSchema[${j}].id`, '字段 ID 重复');
        if (!text(c.type)) error(`${p}.columnsSchema[${j}].type`, '需要字段类型');
        if (c.type === 'function' && !text(c.functionFieldType)) error(`${p}.columnsSchema[${j}].functionFieldType`, '公式字段需要结果类型');
        keys.add(c.key); columnIds.add(String(c.id));
      });
      tables.set(String(t.id), { ...t, columnsSchema: columns.filter(object) });
    });
    if (meta.views !== undefined) array(meta.views, '$schemas.views').forEach((v, i) => {
      const p = `$schemas.views[${i}]`, view = record(v, p);
      id(view.id, `${p}.id`); id(view.tableId, `${p}.tableId`);
      if (views.has(String(view.id))) error(`${p}.id`, '视图 ID 重复');
      if (!tables.has(String(view.tableId))) error(`${p}.tableId`, '视图所属表格不在元数据中');
      enumeration(view.type, ['basic', 'group', 'pivot'], `${p}.type`);
      views.set(String(view.id), view);
    });
  } else warn('$schemas', '未提供字段元数据；只检查配置结构及引用，不能证明表格和字段真实存在');
  const table = (tableId, p) => {
    id(tableId, p);
    const t = tables.get(String(tableId));
    if (schemas !== undefined && !t) error(p, '表格不在提供的元数据中');
    return t;
  };
  const field = (tableId, key, p, allowed) => {
    if (!text(key)) { error(p, '需要真实字段 key'); return; }
    const t = tables.get(String(tableId));
    if (!t) return;
    const c = t.columnsSchema.find(c => c.key === key);
    if (!c) error(p, '字段 key 不属于当前数据表');
    else if (allowed && !allowed.includes(columnType(c))) error(p, `字段类型应为 ${allowed.join(' | ')}`);
    return c;
  };
  const range = (value, p) => {
    const r = record(value, p);
    if (!date(r.start) || !date(r.end) || r.start > r.end) error(p, '需要有效且顺序正确的 YYYY-MM-DD 起止日期');
  };
  const timeRange = (value, amount, customRange, p) => {
    enumeration(value, TIME_RANGES, `${p}.timeRange`);
    if (['recentDays', 'recentMonths'].includes(value) && !positiveInt(amount)) error(`${p}.amount`, '最近时间范围数量必须为正整数');
    if (value === 'custom') range(Array.isArray(customRange) ? { start: customRange[0], end: customRange[1] } : customRange, `${p}.customRange`);
  };
  const filters = (list, p, expectedTable) => array(list, p).forEach((v, i) => {
    const f = record(v, `${p}[${i}]`), q = `${p}[${i}]`;
    id(f.id, `${q}.id`); id(f.columnId, `${q}.columnId`); id(f.schemaId, `${q}.schemaId`);
    const t = table(f.tableId, `${q}.tableId`);
    if (expectedTable !== undefined && String(f.tableId) !== String(expectedTable)) error(`${q}.tableId`, '过滤条件必须属于该数据源');
    if (t && String(f.schemaId) !== String(t.schemaId)) error(`${q}.schemaId`, 'schemaId 与数据表不匹配');
    const c = t?.columnsSchema.find(c => String(c.id) === String(f.columnId));
    if (t && !c) error(`${q}.columnId`, '字段 ID 不属于当前 Schema');
    if (c?.type === 'multiSelect') error(`${q}.columnId`, '当前过滤协议不支持多选集合字段');
    enumeration(f.comparator, c ? (FILTER_OPS[columnType(c)] ?? []) : [...new Set(Object.values(FILTER_OPS).flat())], `${q}.comparator`);
    if (typeof f.value !== 'string') error(`${q}.value`, '保存协议要求字符串；数组需按前端协议以逗号连接');
    if (!['empty', 'notempty'].includes(f.comparator) && !text(f.value)) error(`${q}.value`, '非空判断条件必须填写过滤值');
  });
  filters(cfg.globalFilterList, `${root}.globalFilterList`);

  const metricStyle = (value, p) => {
    if (value === undefined) return;
    const s = record(value, p);
    enumeration(s.numberFormat, FORMATS, `${p}.numberFormat`, true);
    enumeration(s.fontSizeMode, ['auto', 'custom'], `${p}.fontSizeMode`, true);
    bools(s, ['showPresetDateRange', 'thousandSeparator'], p);
    if (s.description !== undefined && (typeof s.description !== 'string' || s.description.length > 200)) error(`${p}.description`, '描述须为不超过 200 字符的字符串');
    for (const key of ['backgroundColor', 'textColor']) if (s[key] !== undefined && (typeof s[key] !== 'string' || !/^#(?:[0-9a-f]{3}|[0-9a-f]{6})$/i.test(s[key]))) error(`${p}.${key}`, '颜色须为 #RGB 或 #RRGGBB');
    for (const key of ['backgroundOpacity', 'textOpacity']) if (s[key] !== undefined && (typeof s[key] !== 'number' || s[key] < 0 || s[key] > 100)) error(`${p}.${key}`, '透明度范围为 0～100');
    if (s.backgroundBlur !== undefined && (typeof s.backgroundBlur !== 'number' || s.backgroundBlur < 0)) error(`${p}.backgroundBlur`, '模糊值必须为非负数字');
    for (const key of ['labelFontSize', 'valueFontSize', 'compareLabelFontSize', 'compareValueFontSize']) if (s[key] !== undefined && !positiveInt(s[key])) error(`${p}.${key}`, '字号必须为正整数');
    if (s.precision !== undefined && (!Number.isInteger(s.precision) || s.precision < 0 || s.precision > 9)) error(`${p}.precision`, '小数位数范围为 0～9');
    if (s.unit !== undefined && typeof s.unit !== 'string') error(`${p}.unit`, '单位必须为字符串');
  };
  const measure = (v, p, tableId, pivot = false) => {
    const m = record(v, p);
    bools(m, ['disabled', 'expand'], p);
    if (m.disabled === true) return;
    enumeration(m.type, AGGREGATIONS, `${p}.type`);
    if (m.type !== 'count') field(tableId, m.field, `${p}.field`, ['sum', 'avg', 'std', 'median', 'percent', 'percentRank'].includes(m.type) ? ['number'] : ['max', 'min'].includes(m.type) ? ['number', 'date'] : undefined);
    if (['percent', 'percentRank'].includes(m.type)) {
      const values = array(m.config?.percents, `${p}.config.percents`);
      if (!values.length || values.some(n => !numeric(n) || (!pivot && (!Number.isInteger(Number(n)) || Number(n) <= 0)) || (m.type === 'percent' && (Number(n) < 0 || Number(n) > 100)))) error(`${p}.config.percents`, '百分数须为 0～100，图表编辑器还要求正整数；percentRank 为数值阈值');
    }
  };
  const dimension = (v, p, tableId, measures, pivot = false) => {
    const d = record(v, p);
    bools(d, ['disabled', 'expand'], p);
    if (d.disabled === true) return;
    enumeration(d.type, ['term', 'date_histogram', 'range', 'date_range'], `${p}.type`);
    const column = field(tableId, d.field, `${p}.field`, d.type === 'term' ? ['text'] : d.type === 'range' ? ['number'] : ['date']);
    if (d.type === 'term' && column && !['text', 'textarea', 'string', 'select'].includes(column.type) && !(pivot && column.type === 'function' && column.functionFieldType === 'text')) error(`${p}.field`, 'term 维度应使用文本/单选字段；只有透视表支持文本公式');
    const c = record(d.config, `${p}.config`);
    if (d.type === 'term') {
      if (!positiveInt(c.sortCount)) error(`${p}.config.sortCount`, '分类数量须为正整数');
      enumeration(c.sortOrder, ['asc', 'desc'], `${p}.config.sortOrder`);
      bools(c, ['groupOther'], `${p}.config`);
      if (!blank(c.sortByMeasureId) && !measures.some(m => object(m) && m.id === c.sortByMeasureId && !m.disabled)) error(`${p}.config.sortByMeasureId`, '排序度量不存在或已禁用');
    }
    if (d.type === 'date_histogram' && (typeof c.miniInterval !== 'string' || !/^[1-9]\d*[smhdwMY]$/.test(c.miniInterval))) error(`${p}.config.miniInterval`, '时间间隔须为 1d、1M 等；m=分钟，M=月，Y=年');
    if (d.type === 'date_histogram' && text(c.miniIntervalTitle)) {
      let interval = c.miniIntervalTitle.replace(/分钟/g, 'm').replace(/小时/g, 'h').replace(/秒/g, 's').replace(/天/g, 'd').replace(/周/g, 'w').replace(/月/g, 'M').replace(/年/g, 'Y').replace(/\s/g, '');
      if (/^[smhdwMY]$/.test(interval)) interval = `1${interval}`;
      if (interval !== c.miniInterval) error(`${p}.config.miniIntervalTitle`, '时间间隔文案与 miniInterval 不一致');
    }
    if (d.type === 'range') {
      const rs = array(c.rangeValueList, `${p}.config.rangeValueList`);
      if (!rs.length) error(`${p}.config.rangeValueList`, '至少填写一个数值范围');
      rs.forEach((v, i) => {
        const r = record(v, `${p}.config.rangeValueList[${i}]`);
        const minEmpty = blank(r.min), maxEmpty = blank(r.max);
        if ((!pivot && (minEmpty || maxEmpty)) || (minEmpty && maxEmpty) || (!minEmpty && !numeric(r.min)) || (!maxEmpty && !numeric(r.max)) || (!minEmpty && !maxEmpty && (pivot ? Number(r.min) >= Number(r.max) : Number(r.min) > Number(r.max)))) error(`${p}.config.rangeValueList[${i}]`, '范围端点无效或顺序错误；透视范围允许单侧开放');
      });
    }
    if (d.type === 'date_range') {
      const key = pivot ? 'dateList' : 'dateRangeValueList';
      const rs = array(c[key], `${p}.config.${key}`);
      if (!rs.length) error(`${p}.config.${key}`, '至少填写一个日期范围');
      rs.forEach((v, i) => {
        const r = record(v, `${p}.config.${key}[${i}]`);
        if (pivot && (blank(r.start) || blank(r.end))) {
          if (!date(r.start) && !date(r.end)) error(`${p}.config.${key}[${i}]`, '开放范围至少需要一个有效日期端点');
        } else range(r, `${p}.config.${key}[${i}]`);
      });
    }
  };
  const activeTables = new Set();
  charts.filter(object).forEach(c => {
    if (c.displayType === 'slicer') return;
    if (c.displayType === 'metricCard') {
      const d = object(c.dataConfig) ? c.dataConfig : {};
      const sources = own(d, 'multiMetricEnabled') ? (Array.isArray(d.sources) ? d.sources : []) : (object(d.source) ? [d.source] : []);
      (d.multiMetricEnabled === true ? sources : sources.slice(0, 1)).filter(object).forEach(s => activeTables.add(String(s.tableId)));
    } else if (['chart', 'view'].includes(c.type)) activeTables.add(String(c.tableId));
  });

  charts.forEach((value, i) => {
    const p = `${root}.tableConfigList[${i}]`, c = record(value, p);
    enumeration(c.displayType, Object.keys(COMPONENTS), `${p}.displayType`);
    if (own(COMPONENTS, c.displayType) && c.type !== COMPONENTS[c.displayType]) error(`${p}.type`, `该组件 type 必须为 ${COMPONENTS[c.displayType]}`);
    counts[c.displayType] = (counts[c.displayType] ?? 0) + 1;
    const d = record(c.dataConfig, `${p}.dataConfig`), dp = `${p}.dataConfig`;
    if (c.title !== undefined && typeof c.title !== 'string') error(`${p}.title`, '标题必须为字符串');
    if (c.displayType === 'metricCard') {
      const modern = own(d, 'multiMetricEnabled');
      bools(d, ['multiMetricEnabled'], dp);
      const sources = modern ? array(d.sources, `${dp}.sources`) : [record(d.source, `${dp}.source`)];
      if (!modern) warn(dp, '历史 source 结构；新建请显式使用 multiMetricEnabled 与 sources');
      if (modern && own(d, 'source')) error(dp, 'source 与 sources 不应混用');
      unique(sources, modern ? `${dp}.sources` : `${dp}.legacySources`);
      if (sources.length < 1 || sources.length > 20) error(`${dp}.sources`, '指标数须为 1～20');
      sources.forEach((v, n) => {
        const s = record(v, `${dp}.sources[${n}]`), sp = modern ? `${dp}.sources[${n}]` : `${dp}.source`;
        if (d.multiMetricEnabled === false && n > 0) {
          warn(sp, '单指标模式下保留的非活动来源，仅检查唯一 ID；切换多指标后需重新校验其配置');
          return;
        }
        table(s.tableId, `${sp}.tableId`);
        filters(s.tableFilters, `${sp}.tableFilters`, s.tableId);
        const a = record(s.aggregation, `${sp}.aggregation`);
        enumeration(a.type, METRIC_AGGREGATIONS, `${sp}.aggregation.type`);
        if (a.type !== 'count') field(s.tableId, a.field, `${sp}.aggregation.field`, ['number']);
        if (typeof a.alias !== 'string') error(`${sp}.aggregation.alias`, '指标别名必须为字符串（允许空字符串）');
        const compare = record(s.compareConfig, `${sp}.compareConfig`);
        if (typeof compare.enabled !== 'boolean') error(`${sp}.compareConfig.enabled`, '必须明确同环比是否启用');
        if (compare.enabled && d.multiMetricEnabled) warn(`${sp}.compareConfig.enabled`, '多指标模式不计算同环比；该配置只在单指标模式生效');
        if (compare.enabled && !d.multiMetricEnabled && n === 0) {
          field(s.tableId, compare.dateField, `${sp}.compareConfig.dateField`, ['date']);
          timeRange(compare.timeRangeType, compare.amount, compare.customRange, `${sp}.compareConfig`);
          enumeration(compare.compareType, ['period', 'week', 'month', 'year'], `${sp}.compareConfig.compareType`);
          enumeration(compare.resultType, ['rate', 'diff', 'value'], `${sp}.compareConfig.resultType`);
        }
        metricStyle(modern ? s.styleConfig : d.styleConfig, `${sp}.styleConfig`);
      });
      if (modern) {
        const m = record(d.multiMetricConfig, `${dp}.multiMetricConfig`), mp = `${dp}.multiMetricConfig`;
        enumeration(m.displayMode, ['number', 'bar', 'area', 'horizontalBar'], `${mp}.displayMode`);
        enumeration(m.sortOrder, ['asc', 'desc', 'sequence'], `${mp}.sortOrder`);
        enumeration(m.align, ['left', 'center', 'right'], `${mp}.align`);
        const pct = record(m.percentage, `${mp}.percentage`);
        enumeration(pct.mode, ['none', 'total', 'custom'], `${mp}.percentage.mode`);
        if (d.multiMetricEnabled === true && pct.mode === 'custom' && !sources.some(s => object(s) && s.id === pct.denominatorSourceId)) error(`${mp}.percentage.denominatorSourceId`, '分母指标必须来自本卡片 sources');
        metricStyle(record(m.chartFormat, `${mp}.chartFormat`), `${mp}.chartFormat`);
      }
      return;
    }
    table(c.tableId, `${p}.tableId`);
    filters(c.tableFilters, `${p}.tableFilters`, c.tableId);
    const s = record(c.styleConfig, `${p}.styleConfig`);
    bools(s, ['toolTip', 'showValueOnChart', 'showXAxis', 'isRing', 'showOperator', 'showPresetDateRange'], `${p}.styleConfig`);
    enumeration(s.legendPosition, ['top', 'left', 'right', 'bottom', 'hidden'], `${p}.styleConfig.legendPosition`, true);
    enumeration(s.colorTheme, THEMES, `${p}.styleConfig.colorTheme`, true);
    if (c.displayType === 'slicer') {
      for (const key of ['measures', 'dimensions']) if (array(d[key], `${dp}.${key}`).length) error(`${dp}.${key}`, '切片器不使用图表度量或维度');
      if (!activeTables.has(String(c.tableId))) error(`${p}.tableId`, '切片器的数据表必须被至少一个有效数据组件引用（单指标只计算第一个 source）');
      enumeration(d.fieldType, ['text', 'number', 'date'], `${dp}.fieldType`);
      field(c.tableId, d.field, `${dp}.field`, [d.fieldType]);
      enumeration(d.style, d.fieldType === 'text' ? ['dropdown', 'tile', 'search'] : [d.fieldType], `${dp}.style`);
      enumeration(d.selectionMode, ['single', 'multiple'], `${dp}.selectionMode`);
      if (d.style === 'search' && d.selectionMode !== 'single') error(`${dp}.selectionMode`, '搜索切片器仅支持单选');
      enumeration(d.sortOrder, ['asc', 'desc'], `${dp}.sortOrder`);
      enumeration(d.operator, d.fieldType === 'text' ? ['eq'] : d.fieldType === 'date' ? ['eq', 'gt', 'lt', 'between'] : ['eq', 'ne', 'gt', 'gte', 'lt', 'lte', 'between'], `${dp}.operator`);
      if (d.optionLimit !== 10) error(`${dp}.optionLimit`, '当前前端固定为 10');
      if (Array.isArray(c.tableFilters) && c.tableFilters.length) error(`${p}.tableFilters`, '切片器必须为空数组，过滤值使用 dataConfig');
      for (const flag of ['showOperator', 'showPresetDateRange']) if (own(d, flag)) warn(`${dp}.${flag}`, '旧位置字段；保存时迁移到组件 styleConfig');
      for (const key of ['defaultValue', 'value']) {
        const v = d[key], q = `${dp}.${key}`;
        if (v === null) continue;
        if (v === undefined) { error(q, '必须显式填写，未选择时为 null'); continue; }
        if (d.fieldType === 'date') {
          if (d.operator === 'between') {
            if (Array.isArray(v)) {
              if (v.length !== 2) error(q, '历史日期区间需要两个端点');
              range({ start: v[0], end: v[1] }, q);
            } else timeRange(v, d.amount, d.customRange, q);
          } else if (!date(v)) error(q, '日期比较值须为 YYYY-MM-DD');
        } else if (d.fieldType === 'number') {
          if (d.operator === 'between') {
            if (!Array.isArray(v) || v.length !== 2 || v.some(n => typeof n !== 'number' || !Number.isFinite(n)) || v[0] > v[1]) error(q, '数字区间必须是两个有序的数字');
          } else if (typeof v !== 'number' || !Number.isFinite(v)) error(q, '数字比较值必须为数字');
        } else if (d.selectionMode === 'multiple') {
          if (!Array.isArray(v) || !v.length || v.some(x => !text(x) && typeof x !== 'number')) error(q, '多选值应为非空字符串/数字数组，清空使用 null');
        } else if (!text(v) && typeof v !== 'number') error(q, '单选值应为字符串或数字');
      }
      return;
    }
    if (c.type === 'view') {
      for (const key of ['measures', 'dimensions']) if (d[key] !== undefined && array(d[key], `${dp}.${key}`).length) error(`${dp}.${key}`, '表格统计配置通过视图/透视接口保存，不能写入组件图表字段');
      bools(d, ['allowAddData', 'allowEditData', 'allowEditSchema', 'allowFilter', 'allowExport'], dp);
      if (blank(c.tableViewId)) {
        (requireBindings ? error : warn)(`${p}.tableViewId`, '等待 /dashboard/save 返回的视图 ID；取得后才能设置透视配置');
      } else {
        id(c.tableViewId, `${p}.tableViewId`);
        const v = views.get(String(c.tableViewId));
        if (schemas?.views !== undefined && !v) error(`${p}.tableViewId`, '视图不在提供的元数据中');
        if (v && (String(v.tableId) !== String(c.tableId) || (c.displayType === 'pivotTable' ? v.type !== 'pivot' : !['basic', 'group'].includes(v.type)))) error(`${p}.tableViewId`, '视图所属表格或类型与组件不匹配');
      }
      return;
    }
    const ms = array(d.measures, `${dp}.measures`), ds = array(d.dimensions, `${dp}.dimensions`);
    unique(ms, `${dp}.measures`); unique(ds, `${dp}.dimensions`);
    if (!ms.some(m => object(m) && !m.disabled) || ms.length > (c.displayType === 'pieCharts' ? 1 : 3)) error(`${dp}.measures`, c.displayType === 'pieCharts' ? '饼图须有且仅有 1 个启用度量' : '须有 1～3 个度量且至少启用 1 个');
    const usedModes = new Set();
    ms.forEach((v, n) => {
      measure(v, `${dp}.measures[${n}]`, c.tableId);
      if (!object(v) || v.styleConfig === undefined || c.displayType === 'pieCharts') return;
      const st = record(v.styleConfig, `${dp}.measures[${n}].styleConfig`), sp = `${dp}.measures[${n}].styleConfig`;
      enumeration(st.chartType, ['bar', 'line', 'area'], `${sp}.chartType`, true);
      enumeration(st.seriesStack, ['stacked', 'normal'], `${sp}.seriesStack`, true);
      enumeration(st.scaleType, ['line', 'log'], `${sp}.scaleType`, true);
      enumeration(st.valueAxisPosition, c.displayType === 'HBarCharts' ? ['top', 'bottom'] : ['left', 'right'], `${sp}.valueAxisPosition`, true);
      enumeration(st.align, ['horizontal', 'vertical', 'oblique'], `${sp}.align`, true);
    });
    ds.forEach((v, n) => {
      dimension(v, `${dp}.dimensions[${n}]`, c.tableId, ms);
      if (!object(v)) return;
      enumeration(v.displayModel, ['normal', 'split', 'cross'], `${dp}.dimensions[${n}].displayModel`);
      if (c.displayType !== 'pieCharts' && usedModes.has(v.displayModel)) error(`${dp}.dimensions[${n}].displayModel`, '普通图表每种维度角色只能有一个');
      usedModes.add(v.displayModel);
      if (v.displayModel === 'split') enumeration(v.splitModel, ['row', 'column'], `${dp}.dimensions[${n}].splitModel`);
      if (c.displayType === 'pieCharts' && v.displayModel === 'split' && ds.slice(0, n).some(d => object(d) && d.displayModel === 'split')) error(`${dp}.dimensions[${n}].displayModel`, '饼图最多一个拆分图表维度；cross 可用于多层嵌套');
    });
    if (c.displayType !== 'pieCharts' && ds.length > 3) error(`${dp}.dimensions`, '普通图表最多 3 个维度角色');
  });

  const chartIds = new Set(charts.filter(object).map(c => c.id));
  const layoutIds = new Set(layouts.filter(object).map(l => l.i));
  charts.filter(object).forEach(c => { if (!layoutIds.has(c.id)) error(`${root}.layoutList`, `组件 ${c.id} 缺少布局`); });
  layouts.forEach((v, i) => {
    const p = `${root}.layoutList[${i}]`, l = record(v, p);
    if (!chartIds.has(l.i)) error(`${p}.i`, '布局没有对应组件');
    for (const key of ['x', 'y', 'w', 'h']) if (!Number.isSafeInteger(l[key]) || l[key] < (['w', 'h'].includes(key) ? 1 : 0)) error(`${p}.${key}`, '位置为非负整数，尺寸为正整数');
    if (l.x + l.w > 24) error(p, '布局超出 24 列栅格');
    for (const axis of ['W', 'H']) {
      for (const bound of ['min', 'max']) {
        const key = `${bound}${axis}`;
        if (l[key] !== undefined && (!Number.isSafeInteger(l[key]) || l[key] < 1 || (bound === 'min' ? l[axis.toLowerCase()] < l[key] : l[axis.toLowerCase()] > l[key]))) error(`${p}.${key}`, '尺寸必须满足有效的 min/max 限制');
      }
    }
    layouts.slice(0, i).filter(object).forEach(other => {
      if (l.x < other.x + other.w && l.x + l.w > other.x && l.y < other.y + other.h && l.y + l.h > other.y) error(p, `布局与 ${other.i} 重叠`);
    });
  });

  if (pivotConfigs !== undefined) array(pivotConfigs, '$pivotConfigs').forEach((value, i) => {
    const p = `$pivotConfigs[${i}]`, payload = record(value, p);
    table(payload.tableId, `${p}.tableId`); id(payload.viewId, `${p}.viewId`);
    if (String(payload.businessId) !== String(payload.tableId) || payload.businessType !== 2) error(p, '透视配置要求 businessId=tableId，businessType=2');
    if (!charts.some(c => object(c) && c.displayType === 'pivotTable' && String(c.tableId) === String(payload.tableId) && String(c.tableViewId) === String(payload.viewId))) error(p, '透视配置必须绑定当前仪表盘已保存的透视组件');
    const pc = record(payload.config, `${p}.config`), pp = `${p}.config`;
    const ms = array(pc.measures, `${pp}.measures`), ds = array(pc.dimensions, `${pp}.dimensions`);
    unique([...ms, ...ds], `${pp}.fields`);
    if (!ms.some(m => object(m) && !m.disabled)) error(`${pp}.measures`, '至少启用一个度量');
    ms.forEach((m, n) => measure(m, `${pp}.measures[${n}]`, payload.tableId, true));
    ds.forEach((d, n) => dimension(d, `${pp}.dimensions[${n}]`, payload.tableId, ms, true));
    const a = record(pc.advanced, `${pp}.advanced`), o = record(pc.other, `${pp}.other`);
    bools(a, ['openDimensionSplit', 'openStatistic'], `${pp}.advanced`);
    enumeration(a.displayMode, ['normal', 'split', 'cross'], `${pp}.advanced.displayMode`);
    if (a.openDimensionSplit && !ds.some(d => object(d) && !d.disabled && d.id === a.colDimensionId)) error(`${pp}.advanced.colDimensionId`, '拆分维度不存在或被禁用');
    if (a.openDimensionSplit && a.displayMode === 'normal') error(`${pp}.advanced.displayMode`, '启用拆分后使用 split 或 cross');
    array(a.totalMeasures, `${pp}.advanced.totalMeasures`).forEach((v, n) => {
      const t = record(v, `${pp}.advanced.totalMeasures[${n}]`);
      if (!t.disabled && a.openStatistic) {
        if (!ms.some(m => object(m) && !m.disabled && m.id === t.id)) error(`${pp}.advanced.totalMeasures[${n}].id`, '合计引用无效度量');
        enumeration(t.type, ['sum', 'valueCount', 'max', 'min', 'median', 'avg'], `${pp}.advanced.totalMeasures[${n}].type`);
      }
    });
    bools(o, ['showSummary'], `${pp}.other`);
    if (!positiveInt(o.maxRowsPerPage)) error(`${pp}.other.maxRowsPerPage`, '每页行数须为正整数');
    array(o.otherExtractConfigVOList, `${pp}.other.otherExtractConfigVOList`).forEach((v, n) => {
      const s = record(v, `${pp}.other.otherExtractConfigVOList[${n}]`), sp = `${pp}.other.otherExtractConfigVOList[${n}]`;
      if (s.disabled || !o.showSummary) return;
      enumeration(s.type, ['measure', 'dimension'], `${sp}.type`);
      const target = (s.type === 'measure' ? ms : ds).find(f => object(f) && f.id === s.id && !f.disabled);
      if (!target || s.field !== target.field || (a.openDimensionSplit && s.type === 'dimension' && s.id === a.colDimensionId)) error(sp, '汇总引用无效，field/type/id 必须与启用的统计字段一致');
      const col = tables.get(String(payload.tableId))?.columnsSchema.find(c => c.key === s.field);
      const allowed = blank(s.field) || columnType(col) === 'number' ? ['count', 'sum', 'avg', 'max', 'min', 'cardinality'] : columnType(col) === 'date' ? ['count', 'min', 'max'] : ['count', 'cardinality'];
      enumeration(s.summaryType, col || blank(s.field) ? allowed : ['count', 'sum', 'avg', 'max', 'min', 'cardinality'], `${sp}.summaryType`);
    });
  });
  return { ok: errors.length === 0, errors, warnings, summary: { components: charts.length, counts, metadataChecked: schemas !== undefined, pivotConfigsChecked: pivotConfigs !== undefined } };
}

function main(args) {
  if (args.includes('--help')) {
    process.stdout.write('Usage: node validate-dashboard.mjs CONFIG.json [--schemas metadata.json] [--pivot-configs pivot-save.json] [--require-bindings]\nInput: config or {appId,id?,name?,config}. JSON output; exit 0=valid, 2=invalid/input error. Offline; no writes.\n');
    return;
  }
  try {
    let filename;
    const options = {};
    for (let i = 0; i < args.length; i++) {
      const arg = args[i];
      if (arg === '--require-bindings') options.requireBindings = true;
      else if (['--schemas', '--pivot-configs'].includes(arg)) {
        if (!args[i + 1] || args[i + 1].startsWith('--')) throw new Error(`${arg} 需要 JSON 文件路径`);
        options[arg === '--schemas' ? 'schemas' : 'pivotConfigs'] = JSON.parse(readFileSync(args[++i], 'utf8'));
      } else if (arg.startsWith('--') || filename) throw new Error(`未知参数 ${arg}`);
      else filename = arg;
    }
    if (!filename) throw new Error('缺少配置文件；使用 --help 查看用法');
    const result = validateDashboard(JSON.parse(readFileSync(filename, 'utf8')), options);
    process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
    process.exitCode = result.ok ? 0 : 2;
  } catch (e) {
    process.stdout.write(`${JSON.stringify({ ok: false, errors: [{ path: '$input', message: e.message }], warnings: [] }, null, 2)}\n`);
    process.exitCode = 2;
  }
}
if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) main(process.argv.slice(2));
