# 仪表盘与透视表临时查询参数

两个接口均使用 `lingtong-cli api POST <路径> --data '<JSON>' --envelope`。租户和用户取自登录态，查询时不会保存配置、创建视图或读取已保存的仪表盘、透视表及视图筛选配置。服务端仍会读取数据源表结构及 ES mapping；字段类型配置由服务端生成。示例中的应用、表和字段必须替换为当前租户实际存在的值。

## 仪表盘：`POST /dashboard/query/direct`

| 路径 | 类型 | 必填 | 查询含义 |
| --- | --- | --- | --- |
| `appId` | 整数 | 是 | 应用 ID；与保存仪表盘时的 `appId` 相同。 |
| `config` | 对象 | 是 | 本次仪表盘配置。|
| `config.tableConfigList` | 数组 | 查询组件时是 | 组件列表；不传或为空时返回空列表。每个元素的字段见下文。 |
| `config.globalFilterList` | 数组 | 否 | 全局筛选条件。按筛选项的 `tableId` 作用于相同数据源的组件；结构见“筛选条件”。 |
| `slicers` | 数组 | 否 | 本次切片器值，格式 `[ {"slicerId":"s1","value":"李莹"} ]`。|
| `slicers[].slicerId` | 字符串 | 是 | 对应切片器组件 `id`；不可重复，必须存在于当前配置。 |
| `slicers[].value` | 任意 JSON 值 | 否 | 覆盖该切片器的本次值；`null`、空字符串或空数组表示全部。若未提供该切片器项，先使用配置中的 `value`，再使用 `defaultValue`。 |

`id` 是保存接口中的仪表盘 ID，此接口不需要。`config.layoutList`、组件或度量的 `styleConfig`、`expand`、`splitModel` 等是布局或外观字段，查询计算不需要。`esFieldTypeConfig` 由服务端依据当前表结构和 ES mapping 生成，无需由 CLI 提供；不要复用其他表的该配置。

### 组件：`config.tableConfigList[]`

| 字段 | 类型 | 查询含义 |
| --- | --- | --- |
| `id` | 字符串 | 组件标识；结果通过此值对应组件，切片器必须有唯一非空 ID。 |
| `title` | 字符串 | 组件标题，随结果返回。 |
| `type` | 字符串 | 组件类型：图表使用 `chart`，切片器使用 `other`，视图使用 `view`。缺失时查询会跳过该组件；切片器校验时可自动补 `other`。 |
| `displayType` | 字符串 | 数据处理类型：`barCharts` 柱状图、`HBarCharts` 条形图、`lineCharts` 折线图、`areaCharts` 面积图、`pieCharts` 饼图、`metricCard` 指标卡、`slicer` 切片器。 |
| `tableId` | 整数 | 数据源表的业务 ID；图表和切片器需要。指标卡数据源在 `dataConfig.sources[].tableId`。 |
| `tableFilters` | 数组 | 仅作用于本组件的筛选条件，格式见“筛选条件”。 |
| `dataConfig` | 对象 | 组件的数据配置，结构按 `displayType` 区分，见下文。 |
| `tableViewId` | 整数或 null | 保存后的关联视图 ID；临时图表查询不依靠该字段，不必传。 |

`basicTable` 和 `pivotTable` 是视图组件：现有仪表盘查询没有对应数据处理器，会跳过，不能仅凭上面线上仪表盘配置得到视图行数据。基础表数据使用表格查询；透视表需要完整透视配置，调用本页的透视表临时查询。仪表盘中的视图组件 `dataConfig.measures/dimensions` 为空时，不能据此还原真实透视配置。单个图表异常会被现有流程记录并跳过，检查返回的组件 ID 是否齐全。

### 普通图表：`dataConfig`

适用于柱状、条形、折线、面积和饼图。`measures` 和 `dimensions` 是数组，字段结构见下文“维度与度量”。柱状、条形、折线和面积图应传 `dimensions`，即使没有维度也传 `[]`；其普通模式使用首个启用的普通维度，并可使用拆分和交叉维度。饼图使用交叉维度和可选拆分维度，**只使用第一个启用的度量**；线上样本的饼图维度使用 `displayModel: "cross"`。每个启用项均需传 `disabled: false`。

### 切片器：`dataConfig`

| 字段 | 类型 | 查询含义 |
| --- | --- | --- |
| `field` | 字符串 | 数据源表中的字段 key、selfKey 或列 ID；切片器必须提供，服务端按表结构校验并解析。 |
| `fieldType` | 字符串 | `text`、`number`、`date`；可省略，由表结构补全。传入时必须与实际类型一致。 |
| `style` | 字符串 | 文本支持 `dropdown`（默认）、`tile`、`search`；数值为 `number`，日期为 `date`。`search` 使用匹配查询，其他文本样式使用精确选择。 |
| `selectionMode` | 字符串 | 文本的 `single`（默认）或 `multiple`；多选值为数组。`search`、数值和日期会归一化为单选。 |
| `operator` | 字符串 | 数值支持 `eq`、`ne`、`gt`、`gte`、`lt`、`lte`、`between`；日期支持 `eq`、`gt`、`lt`、`between`；默认 `eq`。文本按样式和选择方式确定运算符。 |
| `sortOrder` | 字符串 | 文本枚举选项排序 `asc`（默认）或 `desc`。 |
| `optionLimit` | 整数 | 枚举选项上限，默认 100，范围 1～1000。 |
| `defaultValue` | JSON 值 | 没有运行时值时使用的默认值；类型随字段及运算符变化。 |
| `value` | JSON 值 | 配置中的当前值；没有 `slicers` 覆盖时优先于 `defaultValue`。 |
| `amount` | 整数 | 日期快捷范围 `recentDays`/`recentMonths` 的数量，默认分别为 7 天/3 个月。 |
| `customRange.start`、`customRange.end` | 字符串 | 日期快捷范围为 `custom` 时的起止日期。 |

数值或日期 `between` 使用含两个元素的数组。日期值支持 `today`、`yesterday`、`thisWeek`、`lastWeek`、`thisMonth`、`lastMonth`、`thisQuarter`、`lastQuarter`、`thisYear`、`lastYear`、`thisWeekToDate`、`thisMonthToDate`、`thisQuarterToDate`、`thisYearToDate`、`recentDays`、`recentMonths`、`recentHalfYear`、`custom`；也可传日期或日期时间。仅日期的区间结束日按次日零点的左闭右开边界处理。运行时值也可传对象，例如 `{"value":"recentDays","amount":14}` 或 `{"value":"custom","customRange":{"start":"2026-08-14","end":"2026-09-12"}}`；对象中的数量和自定义区间优先于组件配置。

### 指标卡：`dataConfig`

| 字段 | 类型 | 查询含义 |
| --- | --- | --- |
| `multiMetricEnabled` | 布尔值 | 新结构标记；`false` 只查询 `sources[0]`，`true` 按指标分别查询，最多 20 个。字段不存在时使用旧结构 `source`，因此显式 `false` 不能省略。 |
| `sources` | 数组 | 新结构指标配置，1～20 项；多指标模式每项 `id` 必填且不可重复。 |
| `source` | 对象 | 旧结构单指标数据源，仅当 `multiMetricEnabled` 字段不存在时使用；字段与 `sources[]` 相同。 |
| `sources[].id` | 字符串 | 指标标识，用于结果及多指标匹配；多指标模式必填。 |
| `sources[].tableId` | 整数 | 当前指标的数据源表 ID；指标卡可跨表配置多个指标。 |
| `sources[].aggregation.type` | 字符串 | 聚合方式，例如 `count`、`sum`、`avg`、`min`、`max`、`unique`；`count` 不需要字段。 |
| `sources[].aggregation.field` | 字符串 | 聚合字段 key；除 `count` 外按聚合类型提供。 |
| `sources[].aggregation.alias` | 字符串 | 聚合结果别名。 |
| `sources[].tableFilters` | 数组 | 当前指标的筛选条件，结构见“筛选条件”。 |
| `sources[].compareConfig` | 对象 | 单指标的对比配置；多指标模式不执行同环比。 |
| `compareConfig.enabled` | 布尔值 | 是否按时间范围查询当前值并计算对比值。 |
| `compareConfig.dateField` | 字符串 | 启用对比时用于限定时间的日期字段 key。 |
| `compareConfig.timeRangeType` | 字符串 | 当前时间范围，支持上节日期快捷范围中的各值，包括 `custom`。 |
| `compareConfig.amount` | 整数 | `recentDays`、`recentMonths` 数量；默认分别为 7、3。 |
| `compareConfig.customRange.start`、`.end` | 字符串 | 自定义起止日期，例如 `2026-08-14`；服务端使用日期部分。 |
| `compareConfig.compareType` | 字符串 | `week`、`month`、`year` 分别将当前区间前移一周、一月、一年；`period` 使用紧邻的同长度区间；`custom` 使用 `customRange`。 |
| `compareConfig.resultType` | 字符串 | `value` 返回对比值；`diff` 返回当前值减对比值；`rate` 返回差值除以对比值的百分数，对比值为零时返回 null。 |
| `compareConfig.label` | 字符串 | 展示文案；不参与查询或计算。 |
| `multiMetricConfig.percentage.mode`、`.denominatorSourceId` | 字符串 | 面板百分比配置；`custom` 时分母 ID 须匹配 `sources[].id`，否则服务端归一化为 `none`。不改变聚合查询。 |

`multiMetricConfig` 中的 `displayMode`、`sortOrder`、`align`、`chartFormat` 及各类 `styleConfig` 是展示配置，不参与服务端聚合。服务端会在组件或各指标中生成 `esFieldTypeConfig`，无需手动构造。

### 筛选条件

适用于 `globalFilterList[]`、组件 `tableFilters[]`、指标卡 `sources[].tableFilters[]` 和透视表请求的 `filters[]`。

| 字段 | 类型 | 查询含义 |
| --- | --- | --- |
| `columnId` | 字符串 | 筛选字段 key 或列 ID，转换为底层查询字段。 |
| `comparator` | 字符串 | 运算符，支持 `EQ`、`NE`、`NEQ`、`GT`、`GTE`、`LT`、`LTE`、`EXISTS`、`MATCH`、`EMPTY`、`NOT_EMPTY`、`NOTEMPTY`、`CT`、`NCT`；大小写不敏感。`CT`/`NCT` 分别映射为包含/不包含列表，并按英文逗号拆分 `value`；不要在此字符串协议中直接使用 `IN`、`NOT_IN` 或 `RANGE`，它们不能通过当前转换得到所需的列表或区间。 |
| `value` | 字符串 | 比较值；无值型运算符可不传。具体解析取决于运算符和字段类型。 |
| `tableId` | 整数 | 全局筛选必须传，用于匹配组件或指标的数据源；组件筛选不传时作用于当前表，传入时须与该组件表匹配。透视表直接查询只使用字段、运算符和值。 |
| `id` | 整数 | 视图筛选项标识；本次查询不使用，可省略。 |

## 透视表：`POST /api/pivot-table/query/direct`

| 路径 | 类型 | 必填 | 查询含义 |
| --- | --- | --- | --- |
| `businessId` | 整数 | 是 | 数据源业务 ID；查询基础资料时为表 ID，例如 2966。 |
| `businessType` | 整数 | 是 | 业务类型编码；`2` 为基础资料，`1` 为场景。请使用实际数据源类型。 |
| `filters` | 数组 | 否 | 此次筛选，结构见上节；不读取视图筛选。 |
| `config` | 对象 | 是 | 透视配置，仅本次使用。 |
| `config.dimensions` | 数组 | 否 | 分组维度，见“维度与度量”；普通无分组查询可省略。拆分/交叉模式按配置引用维度。 |
| `config.measures` | 数组 | 是 | 度量列表，不可为空；每项需显式 `disabled`。 |
| `config.advanced` | 对象 | 是 | 显示模式及拆分配置。 |
| `config.advanced.displayMode` | 字符串 | 建议传 | `normal` 普通、`split` 拆分、`cross` 交叉；空或未知值按 `normal`。 |
| `config.advanced.openDimensionSplit` | 布尔值 | 拆分/交叉时是 | 为 `false` 时按普通模式处理；使用 `split`/`cross` 应传 `true`。 |
| `config.advanced.openStatistic` | 布尔值 | 交叉汇总时是 | 交叉模式为 `true` 时读取 `totalMeasures`。 |
| `config.advanced.colDimensionId` | 字符串 | 拆分/交叉时是 | 引用 `dimensions[].id`；拆分模式作为拆分维度，交叉模式作为列维度。 |
| `config.advanced.totalMeasures` | 数组 | 交叉统计开启时是 | 要汇总的度量项，使用匹配 `measures[].id` 的 `id` 和显式 `disabled`。 |
| `config.advanced.rowDimensions` | 字符串数组 | 否 | 当前转换逻辑不读取该字段；交叉模式的行维度取除列维度外启用的维度。 |
| `config.other` | 对象 | 是 | 其他查询配置。 |
| `config.other.showSummary` | 布尔值 | 是 | 是否计算汇总；`true` 时使用下面的汇总项。 |
| `config.other.otherExtractConfigVOList` | 数组 | 否 | 汇总项，`showSummary` 为 `true` 时生效。 |
| `config.other.otherExtractConfigVOList[].id` | 字符串 | 汇总项必填 | 关联维度或度量的 `id`。 |
| `config.other.otherExtractConfigVOList[].type` | 字符串 | 汇总项必填 | 汇总目标：`dimension` 或 `measure`。 |
| `config.other.otherExtractConfigVOList[].summaryType` | 字符串 | 汇总项必填 | 汇总计算类型，例如 `count`、`sum`、`avg`、`min`、`max`、`cardinality`（去重计数）。 |
| `config.other.otherExtractConfigVOList[].disabled` | 布尔值 | 汇总项必填 | `false` 启用，`true` 禁用。 |
| `config.other.maxRowsPerPage`、`showSummary` 以外的展示属性 | 各类型 | 否 | 当前直接查询转换未读取 `maxRowsPerPage`，不要依赖它限制返回行数。 |

`config.id`、`tenantId`、`viewId`、`tableId`、`businessId`、`businessType`、配置名称和创建更新信息不用于确定本次数据源；实际业务 ID、类型和租户使用外层参数及登录态。请求中的 `config.viewId` 不触发视图筛选读取。

### 将线上透视配置改为临时查询

`/api/pivot-table/config` 的响应外层 `success`、`msg`、`code`、`businessCode`、`clueId` 是响应信封，不是查询参数。取其 `result` 中的 `businessId`、`businessType` 放到新请求外层，取 `result.dimensions`、`result.measures`、`result.advanced` 和 `result.other` 放到外层 `config` 中。直接查询不会使用 `result.viewId` 从数据库补充筛选；原视图筛选不属于这份响应，如需筛选需由调用方明确传入 `filters`。

对于示例中的 `result.id`、`tenantId`、`viewId`、`tableId`、`configName`、`configDesc`、创建更新时间和人员，以及 `esFieldTypeConfig`，此查询不需要传。`enabled` 是返回的派生状态，不是配置开关；使用 `disabled`，建议去掉 `enabled`。`expand`、`styleConfig`、`dashboardAlign`、`splitModel`、空值字段也可省略。保留维度 ID、度量 ID 以及排序和汇总对它们的引用。

以下是以上线返回配置为基础的等效查询请求，省略了这些非查询字段。保存为 `pivot-query.json` 后运行 `lingtong-cli api POST /api/pivot-table/query/direct --data "$(cat pivot-query.json)" --envelope`；实际使用时应确认当前登录租户中的数据源和字段可用。

```json
{
  "businessId": 2966,
  "businessType": 2,
  "config": {
    "dimensions": [{
      "id": "pd_group",
      "disabled": false,
      "field": "zhiku_group",
      "alias": "运营小组",
      "type": "term",
      "config": {
        "sortByMeasureId": "pm_sales",
        "sortOrder": "desc",
        "sortCount": 20,
        "groupOther": false,
        "extendedBounds": false
      }
    }],
    "measures": [
      {"id": "pm_sales", "disabled": false, "field": "sales_amount", "alias": "销售金额（元）", "type": "sum"},
      {"id": "pm_items", "disabled": false, "field": "sales_items", "alias": "销售件数", "type": "sum"},
      {"id": "pm_profit", "disabled": false, "field": "gross_profit", "alias": "商品毛利（元）", "type": "sum"}
    ],
    "advanced": {
      "openDimensionSplit": false,
      "openStatistic": false,
      "displayMode": "normal",
      "totalMeasures": []
    },
    "other": {
      "showSummary": true,
      "otherExtractConfigVOList": [
        {"id": "pm_sales", "type": "measure", "summaryType": "sum", "disabled": false},
        {"id": "pm_items", "type": "measure", "summaryType": "sum", "disabled": false},
        {"id": "pm_profit", "type": "measure", "summaryType": "sum", "disabled": false}
      ]
    }
  }
}
```

要复现特定筛选结果，在请求顶层补充例如 `"filters":[{"columnId":"zhiku_group","comparator":"EQ","value":"李莹"}]`。此示例没有设置筛选，不能保证与携带视图筛选或仪表盘切片器的查询结果一致。线上配置中的 `maxRowsPerPage: 50` 目前不会被这个接口的查询转换使用，示例因此省略。

## 维度与度量

仪表盘普通图表的 `dataConfig.dimensions/measures` 和透视表的 `config.dimensions/measures` 共用以下结构。字段名推荐使用表 Schema 的字段 key，例如线上配置的 `zhiku_group`、`listing_time`、`sales_amount`；按实际 Schema 核对。底层可将字段 key 映射为 ES 字段名，不能假定所有字段都必须手写 `data{tableId}.{fieldId}`。

| 路径 | 类型 | 查询含义 |
| --- | --- | --- |
| `dimensions[].id` | 字符串 | 维度标识；透视表高级配置通过此值引用。 |
| `dimensions[].field` | 字符串 | 分组字段 key。 |
| `dimensions[].alias` | 字符串 | 结果中的维度别名。 |
| `dimensions[].title` | 字符串 | 图表维度标题，可选。 |
| `dimensions[].type` | 字符串 | `term` 分组、`date_histogram` 日期分组、`range` 数值区间、`date_range` 日期区间。 |
| `dimensions[].disabled` | 布尔值 | 必须显式设置；`false` 启用，`true` 跳过。 |
| `dimensions[].displayModel` | 字符串 | 仪表盘图表维度角色：`normal` 行维度、`split` 拆分维度、`cross` 交叉维度；缺失按普通维度处理。不是透视表的 `advanced.displayMode`。 |
| `dimensions[].config.sortByMeasureId` | 字符串 | 按指定 `measures[].id` 对分组排序；不传默认按分组键排序。 |
| `dimensions[].config.sortOrder` | 字符串 | 分组排序方向 `asc` 或 `desc`。 |
| `dimensions[].config.sortCount` | 整数 | 返回的分组桶数量上限。 |
| `dimensions[].config.groupOther` | 布尔值 | 是否合并未展示的分组为其他。 |
| `dimensions[].config.miniInterval` | 字符串 | 日期分组间隔，例如 `1d`、`1w`、`1M`；服务端转换为 ES 查询间隔。请求字段是 `miniInterval`，不是 `interval`。 |
| `dimensions[].config.extendedBounds` | 布尔值 | 日期分组是否自动补空桶；默认 `false`。 |
| `dimensions[].config.rangeValueList[].min`、`.max` | 整数 | 数值区间上下界，转换为 ES 的 `from`、`to`。 |
| `dimensions[].config.dateList[].start`、`.end` | 字符串 | 日期区间上下界，转换为 ES 的 `from`、`to`。 |
| `dimensions[].config.miniIntervalTitle` | 字符串 | 前端间隔文案，不参与查询。 |
| `measures[].id` | 字符串 | 度量标识；供维度排序和透视表汇总引用，应保持唯一。 |
| `measures[].field` | 字符串 | 聚合字段 key；`count` 可省略，其他聚合按类型提供。 |
| `measures[].type` | 字符串 | `count` 记录数、`valueCount` 非空值数、`sum`、`avg`、`min`、`max`、`unique` 去重数、`median` 中位数、`percent` 百分位、`std` 标准差。 |
| `measures[].alias` | 字符串 | 结果列别名；缺省时使用底层默认列名。 |
| `measures[].title` | 字符串 | 图表度量标题；未提供别名时用于图表标题。 |
| `measures[].disabled` | 布尔值 | 必须显式设置；`false` 启用，`true` 跳过。 |
| `measures[].config.percents` | 数字数组 | `percent` 聚合的百分位列表，例如 `[50,95]`。 |

底层存在 `percentRank` 聚合，但当前请求 DTO 的度量配置没有传递它所需的 `values`，不要按上述接口声称它可直接使用。图表字段 `styleConfig`、`expand`、`splitModel`、度量 `dashboardAlign` 不参与这些查询计算。


## 参考线上配置的仪表盘请求

下面保留了线上样例中的切片器、销售金额指标卡、运营小组柱状图中参与查询的字段，去掉布局、颜色和服务端生成的字段类型配置。保存为 `dashboard-query.json` 后执行 `lingtong-cli api POST /dashboard/query/direct --data "$(cat dashboard-query.json)" --envelope`；使用前请确认表 2966 和字段在当前租户存在。

```json
{
  "appId": 925,
  "config": {
    "globalFilterList": [],
    "tableConfigList": [
      {
        "id": "demo61_slicer_group",
        "title": "运营小组筛选",
        "type": "other",
        "displayType": "slicer",
        "tableId": 2966,
        "tableFilters": [],
        "dataConfig": {
          "field": "zhiku_group",
          "fieldType": "text",
          "style": "tile",
          "selectionMode": "single",
          "sortOrder": "asc",
          "optionLimit": 10,
          "value": "李莹"
        }
      },
      {
        "id": "demo61_metric_sales",
        "title": "销售金额合计",
        "type": "chart",
        "displayType": "metricCard",
        "dataConfig": {
          "multiMetricEnabled": false,
          "sources": [{
            "id": "demo61_sales_source",
            "tableId": 2966,
            "aggregation": {"field": "sales_amount", "alias": "销售金额合计", "type": "sum"},
            "tableFilters": [],
            "compareConfig": {"enabled": false}
          }]
        }
      },
      {
        "id": "demo61_bar_group_sales",
        "title": "运营小组销售金额",
        "type": "chart",
        "displayType": "barCharts",
        "tableId": 2966,
        "tableFilters": [],
        "dataConfig": {
          "measures": [{"id": "m_group_sales", "field": "sales_amount", "alias": "销售金额（元）", "type": "sum", "disabled": false}],
          "dimensions": [{
            "id": "d_operating_group", "field": "zhiku_group", "alias": "运营小组", "type": "term",
            "displayModel": "normal", "disabled": false,
            "config": {"sortOrder": "desc", "sortCount": 10, "groupOther": false, "sortByMeasureId": "m_group_sales"}
          }]
        }
      }
    ]
  },
  "slicers": [{"slicerId": "demo61_slicer_group", "value": "李莹"}]
}
```

切片器值会作用于同一数据源的指标卡和图表。对其他类型的图表，将 `displayType`、维度角色和聚合配置按前面的字段说明调整；饼图使用 `pieCharts` 和 `cross` 维度。该示例不需要 `id: 61`，也不需要先保存仪表盘。
