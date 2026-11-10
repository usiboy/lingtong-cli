# 单指标与多指标卡

## 新建格式

`type=chart, displayType=metricCard`。使用 `dataConfig.multiMetricEnabled`（显式布尔）、`sources`（1～20 个）、`multiMetricConfig`。**是否存在 multiMetricEnabled 字段**决定前端使用新格式；仅写 sources 而省略开关会被当成旧格式。

sources 内每项：

| 字段 | 说明 |
| --- | --- |
| id | 卡片内唯一字符串；后端结果 sourceId 与它对应 |
| tableId | 真实数据表；不同指标允许不同表 |
| tableFilters | 属于该指标数据表的过滤条件 |
| aggregation | `{type,field,alias}` |
| compareConfig | 同环比设置 |
| styleConfig | 该指标的显示样式 |

aggregation.type 为 count/sum/max/min/median/avg/std/unique/valueCount。count 统计记录，field=null；其他方式按当前指标卡编辑器使用**数字字段**，含数字公式。不要照搬普通图表允许的日期 max/min 或文本 unique。

单指标模式只查询/显示 sources[0]；其余源可以留存以切回多指标，但不是切片器的活动来源。校验时保留这些非活动源，检查其 ID 唯一性，不因它们尚未完成数据绑定而阻止保存；切回多指标后再完整校验全部来源和占比分母。多指标模式查询全部源，当前不计算同环比。

## 同环比

`compareConfig`：

- enabled：是否启用。
- dateField：日期字段 key。
- timeRangeType：today/yesterday、thisWeek/lastWeek、thisMonth/lastMonth、thisQuarter/lastQuarter、thisYear/lastYear；以及 thisWeekToDate/thisMonthToDate/thisQuarterToDate/thisYearToDate、recentDays/recentMonths/custom。
- amount：recentDays/recentMonths 的正整数数量，其他场景 null。
- customRange：custom 时 `{start,end}`，格式 YYYY-MM-DD；其他场景 null。读取时兼容历史 `[start,end]`。
- compareType：period（环比）、week（月/年之外的周同比）、month、year。旧 custom 类型不作为新建可选值，编辑器将它显示为 period，需确认业务口径再迁移。
- resultType：rate（差异率）、diff（差值）、value（原始值）；label 为业务展示名称。

北京时间日历边界见 [筛选与日期](filters-and-slicers.md)。当前值百分比格式会将原值乘 100；同环比 rate 返回值前端直接加 `%`，不要再次乘 100。

## 多指标整体配置

`multiMetricConfig`：

| 字段 | 值/含义 |
| --- | --- |
| displayMode | number / bar / area / horizontalBar |
| sortOrder | sequence（数组顺序）/ asc / desc |
| align | left / center / right；数字模式生效 |
| percentage.mode | none / total / custom |
| percentage.denominatorSourceId | custom 时取本卡片某个 source.id；其他模式 null |
| chartFormat | `{precision,numberFormat}`，控制图形模式数值格式 |

total 为占全部指标合计，custom 为占指定指标；分母自身不展示其占比。分母为零或数据缺失时显示 `--`。新建显式用 sequence，历史缺省值为 desc。多指标结果按 sourceId 匹配，不能靠数组位置匹配。

## 每个指标的 styleConfig

- description：最多 200 字符；showPresetDateRange：显示同环比日期范围。
- backgroundColor/textColor：十六进制颜色；backgroundOpacity/textOpacity：0～100；backgroundBlur：非负值。
- fontSizeMode：auto/custom；labelFontSize/valueFontSize/compareLabelFontSize/compareValueFontSize 为正整数字号。
- precision：0～9；numberFormat：numberThousands/number/percent/cny/usd。
- unit：单位文本；thousandSeparator：布尔。

完整默认值在 metric-card.json；跨表占比在 multi-metric-card.json。

## 历史兼容与保存

旧结构为 `dataConfig.source` + `dataConfig.styleConfig`。只读时保留；编辑迁移时将 source 转为 sources[0]，将旧样式和标题语义移到该 source，补齐开关和 multiMetricConfig，然后移除旧 source/styleConfig，避免两套同时存在。保存新结构仅需组件 id/title/type/displayType/dataConfig，不要强求普通图的顶层 tableId、measures、dimensions。

保存响应可能规范化来源/样式，回读以服务端结果为准。单指标查询 `data.currentValue/resultValue/resultType`；多指标查询 `data.list[]` 中的 sourceId/currentValue/status/errorMessage。HTTP 成功不代表每个指标都计算成功。
