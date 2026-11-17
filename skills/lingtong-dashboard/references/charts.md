# 柱、条、线、面积与饼图

## 公共结构

五类组件使用 `type=chart`、`tableId`、`tableFilters`、`dataConfig.measures[]`、`dataConfig.dimensions[]`、顶层 `styleConfig`。先选真实表格和字段。

### 度量 measures

| 字段 | 配置 |
| --- | --- |
| id | 本组件内唯一字符串 |
| field | Schema 列 key；count 可为 `""` |
| type | count、sum、avg、std、unique、valueCount、max、min、median、percent、percentRank |
| alias / title | 业务别名/展示名；可用 alias 表达口径 |
| disabled / expand | 是否禁用/编辑器折叠状态；新建显式 false |
| config.percents | percent/percentRank 的参数数组 |
| styleConfig.chartType | bar / line / area；用于组合图 |
| styleConfig.valueAxisPosition | 纵向图 left/right；HBarCharts bottom/top |
| styleConfig.seriesStack | normal / stacked |
| styleConfig.scaleType | line / log（line 在此表示线性刻度） |
| styleConfig.align | horizontal / vertical / oblique |

普通图最多 3 个度量，饼图为 1 个；至少启用一个。sum/avg/std/median/percent/percentRank 使用数值字段，max/min 可用数值或日期，unique/valueCount 根据实际字段和统计口径选择。图表编辑器的 percent/percentRank 输入要求正整数，percent 还应在 0～100 范围内；不要把 percentRank 的阈值误认为百分比。

### 维度 dimensions

公共字段为 `id,field,type,alias,disabled,expand,displayModel,splitModel,config`。

| type | 字段类型 | config |
| --- | --- | --- |
| term | 文本/多行文本/字符串/单选 | sortCount 正整数；sortOrder=asc/desc；sortByMeasureId 为空或启用的度量 ID；groupOther 布尔 |
| date_histogram | 日期、创建/更新时间、日期公式 | miniInterval，如 1d/1M；miniIntervalTitle，如 1天/1月 |
| range | 数值及数值公式 | rangeValueList: `[{min:"0",max:"100"}]`；图表需要完整两端，min<=max |
| date_range | 日期及日期公式 | dateRangeValueList: `[{start:"2026-01-01",end:"2026-01-31"}]` |

`miniInterval` 单位为 s/m/h/d/w/M/Y，m=分钟、M=月。图表 term 选择器不提供公式文本或多选集合字段；透视 term 另支持文本公式。时间说明保持与实际间隔一致。

`displayModel`：normal 表示主分类轴，split 表示多个小图，cross 表示拆分序列。普通图每种角色至多一个，合计最多 3 个维度；split 使用 `splitModel=row|column`。无维度时可按全表度量汇总，不要强制伪造维度。排序引用使用度量 id，不能用 field 或名称。

### 饼图

只使用一个度量，忽略笛卡尔图度量样式。`isRing=true` 为环形。split 按维度生成多个饼图，最多一个；cross 可重复用于多层嵌套，不能把普通图的“三维度上限”套到饼图。已有 normal 维度可保留。见 advanced-charts.json 中 nested_pie。

## 图表面板

| styleConfig 字段 | 值 |
| --- | --- |
| legendPosition | top / left / right / bottom / hidden |
| toolTip | 是否显示提示 |
| showValueOnChart | 是否显示值 |
| showXAxis | 是否显示主轴 |
| colorTheme | default / forest / ocean / sunset / pastel / jewel / berry / blues / teal / emerald |
| isRing | 饼图环形开关 |

默认参考各最小示例。组合图保留真实 displayType（例如 barCharts），然后分别设置度量 chartType 和轴位置。log 刻度是否适合取决于真实结果，离线配置不能判断零值/负值数据。

返回数据按组件 id 对应；笛卡尔图关注 `data.charts`、`data.displayModes`、series 的 measureId，饼图关注 `data.pies`。后端负责聚合，前端根据返回的 NORMAL/SPLIT/CROSS 模式渲染；不要把大写返回值写入小写配置 displayModel。
