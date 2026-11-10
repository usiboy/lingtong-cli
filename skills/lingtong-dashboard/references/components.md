# 组件选型与示例

## 注册组件

`type` 是分类，`displayType` 是渲染器标识，大小写不可改变。

| displayType | type | 适用需求 | 核心配置 | 最小示例 |
| --- | --- | --- | --- | --- |
| metricCard | chart | 总额、订单量、目标完成率、同环比 | sources / aggregation / compareConfig | [metric-card.json](../assets/examples/metric-card.json) |
| barCharts | chart | 分类对比、组合图 | measures + dimensions；默认 bar | [bar-chart.json](../assets/examples/bar-chart.json) |
| HBarCharts | chart | 横向排名 | 默认 bar，度量轴 bottom/top | [horizontal-bar-chart.json](../assets/examples/horizontal-bar-chart.json) |
| lineCharts | chart | 时间趋势 | 默认 line，日期直方图 | [line-chart.json](../assets/examples/line-chart.json) |
| areaCharts | chart | 趋势与面积 | 默认 area，可堆叠 | [area-chart.json](../assets/examples/area-chart.json) |
| pieCharts | chart | 构成、环形、多层饼图 | 单度量、维度、isRing | [pie-chart.json](../assets/examples/pie-chart.json) |
| basicTable | view | 明细数据 | tableId / tableViewId / 操作开关 | [basic-table.json](../assets/examples/basic-table.json) |
| pivotTable | view | 维度交叉、合计与汇总 | 视图绑定后独立保存统计配置 | [pivot-table.json](../assets/examples/pivot-table.json) |
| slicer | other | 联动筛选 | field / fieldType / operator / value | [slicer.json](../assets/examples/slicer.json) |

切片器示例附带一个指标卡，因为不能单独绑定没有被数据组件使用的表。默认新增尺寸：指标卡 6×5、普通图表 8×8、表格 16×12、切片器 8×3；均为栅格单位。示例为方便阅读可采用不同有效布局。

## 组合与高级示例

- [operating-overview.json](../assets/examples/operating-overview.json)：12 个组件覆盖全部 9 类；单指标同环比、跨表多指标、5 种图表、明细、透视、3 种字段类型的切片器及全局过滤。含虚构的保存后视图 ID，用于演示第二阶段。
- [multi-metric-card.json](../assets/examples/multi-metric-card.json)：跨表指标、自定义分母、横向条形展示。
- [advanced-charts.json](../assets/examples/advanced-charts.json)：双轴柱线组合、堆叠、拆分+交叉、数字范围、日期范围、百分数、多层嵌套饼图。
- [slicer-variants.json](../assets/examples/slicer-variants.json)：下拉、平铺多选、搜索、数字区间、日期预设、自定义、最近 N 天及日期比较。它是配置样例集合，多切片器同时启用会叠加过滤，使用时只选需要的组件。
- [pivot-configs.json](../assets/examples/pivot-configs.json)：经营概览中透视卡片的第二阶段请求数组；每一项单独提交给透视配置接口，不能作为 dashboard/save 请求。
- [legacy-metric-card.json](../assets/examples/legacy-metric-card.json)：历史 source 格式的兼容读取示例，新建使用 sources。
- [metadata.json](../assets/examples/metadata.json)：以上示例的虚构表格/字段/视图元数据，仅供离线校验。实际使用必须从目标应用查询。

图表没有独立的 combinedChart 组件类型：组合图通过每个度量的 styleConfig.chartType 配置。全局过滤与演示模式也不是新增组件类型。
