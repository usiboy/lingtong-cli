# 全局过滤、组件过滤与切片器

## 静态过滤

全局 config.globalFilterList、普通组件 tableFilters、指标 sources[].tableFilters 都使用扁平数组。每项：

```json
{"id":9001,"tableId":1001,"schemaId":2001,"columnId":5,"comparator":"eq","value":"华东"}
```

id 是条件 ID；tableId/schemaId/columnId 从真实元数据取得。保存 value 为字符串，前端多值按逗号连接；不能把编辑器的列标题/类型或按表分组的结构直接当成持久化格式。组件/指标过滤必须属于该数据源表。全局条件按表作用，不会自动把不同表上的同名字段关联起来。

| 字段类型 | comparator |
| --- | --- |
| 文本（含普通单选等） | eq / neq / ct / nct / empty / notempty |
| 数字 | eq / neq / gt / gte / lt / lte / empty / notempty |
| 日期 | eq / lt / gt / empty / notempty |
| 子表、密码 | empty / notempty |

empty/notempty 的 value 可为空字符串，其他条件要填写值。公式字段按 functionFieldType 选择。多选集合字段 multiSelect 当前不在过滤契约内。条件叠加和最终结果以实际查询为准。

## 切片器公共结构

`type=other, displayType=slicer`，tableFilters=[]。数据源必须已被图表、表格或活动指标源引用；单指标仅 sources[0] 算活动源。切片器不额外创建业务表。

`dataConfig`：空 measures/dimensions、field（字段 key）、fieldType、style、selectionMode、sortOrder、operator、defaultValue、value、amount、customRange、optionLimit。

- defaultValue 为默认值，value 为当前值；清空显式为 null；新建设置默认时通常让两者一致。当前前端会通过保存流程持久化筛选变化，不能假定只影响本地会话。
- sortOrder=asc/desc；optionLimit 当前固定 10。
- 顶层 styleConfig.showOperator（默认 true）、showPresetDateRange（默认 false）。日期范围展示开关不改变查询；数字 UI 仍展示比较器，不承诺该开关对所有样式都生效。
- 旧 dataConfig.showOperator/showPresetDateRange 仅用于兼容读取；保存移到顶层 styleConfig 并移除旧位置，两处都有时顶层优先。

| fieldType | style | operator | 值 |
| --- | --- | --- | --- |
| text | dropdown / tile | eq | single 为字符串/数字；multiple 为数组 |
| text | search | eq | 仅 single |
| number | number | eq/ne/gt/gte/lt/lte/between | 比较为数字；between 为 `[最小值,最大值]` |
| date | date | eq/gt/lt | YYYY-MM-DD 字符串 |
| date | date | between | 时间预设或 custom，详见下方 |

注意数字切片器“不等于”为 **ne**，静态过滤为 **neq**。多选切片器是选择多个文本值，不代表支持 multiSelect 集合字段。字段类型按公式结果类型映射，排除多选集合、子表和密码。

## 日期范围

value/defaultValue 保存相对预设：today/yesterday、thisWeek/lastWeek、thisMonth/lastMonth、thisQuarter/lastQuarter、thisYear/lastYear、thisWeekToDate/thisMonthToDate/thisQuarterToDate/thisYearToDate、recentDays/recentMonths/custom。

recentDays/recentMonths 需要正整数 amount；custom 需要 customRange={start,end}。历史 `[start,end]` 数组可读取，新建用 custom + customRange。一个切片器共享 amount/customRange，不要给默认值和当前值表达两个不同的自定义范围。

统一使用北京时间 UTC+8，范围包含起止日期：

- 自然周从周一到周日；月、季度、年覆盖完整自然周期。
- “至今”从周期首日到今天结束。
- 最近 N 天从今天减 N−1 天到今天结束。
- 最近 N 个月从今天向前 N 个月的同日到今天结束，无同日取月末，不额外加一天。
- 自定义范围必须是有效有序日期；带时间字段用 `[开始日00:00,结束日次日00:00)` 理解边界。

保存相对预设，不把它永久展开成今日的绝对日期；服务端查询时解释。日期文案刷新不意味着数据定时刷新，零点/恢复页面及月末结果需要联调验证。完整样例见 slicer-variants.json。
