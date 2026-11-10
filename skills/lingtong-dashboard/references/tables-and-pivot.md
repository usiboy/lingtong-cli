# 基础表格与透视表

## 组件保存

共同结构：`type=view`，displayType=basicTable/pivotTable，tableId 为业务表 ID，tableFilters 为该表过滤条件。dataConfig 新建包含空 measures/dimensions；统计配置不会从这两个数组读取。

| 组件 | dataConfig 开关 | 默认 |
| --- | --- | --- |
| basicTable | allowAddData / allowEditData / allowEditSchema | 全 false |
| pivotTable | allowFilter / allowExport | 全 false |

按用户要求开启操作能力，并遵循已有平台权限。开关不等于授予权限。快麦来源编辑透视统计同时需要仪表盘编辑和表格编辑权限。

前端通过主应用表格页面嵌入视图，不把表格数据直接塞到仪表盘图表查询结果中。

## 两阶段流程

1. 新增时提交绑定了真实 tableId 的组件，可省略 tableViewId；不要编造视图 ID。
2. 回读 dashboard/save 响应或 dashboard/get，从该组件获取服务端 tableViewId。基础表格也要等待该 ID 才能展示。
3. 透视表通过已有 [lingtong-pivot-table](../../lingtong-pivot-table/SKILL.md) 读取/保存视图统计。请求为 `{tableId,viewId,businessId:tableId,businessType:2,config}`。
4. 回读透视配置并查询透视数据；验证 dashboardId 过滤上下文、视图归属和浏览器展示。更换 tableId 后不要复用旧视图绑定。

最小 pivot-table.json 示范第一步；operating-overview.json 与 pivot-configs.json 示范取得视图 ID 后的绑定。两者 ID 均是假数据，不能直接用于线上。

## 透视统计配置

维度/度量的 id、field、type、alias、disabled、config 与透视表技能一致。以下为仪表盘编辑器的关键差异：

- measures 至少启用一个，无普通图表的 3 个上限。
- dimensions 使用 term/date_histogram/range/date_range；term 支持文本公式；range 允许一侧空值，但有两个端点时 min<max。
- 日期范围字段名是 `config.dateList`，普通图表是 dateRangeValueList，不能混用。
- percent 参数为 0～100 数值，percentRank 参数为数值阈值，透视编辑器不要求正整数。
- advanced.openDimensionSplit 开启时，colDimensionId 必须引用启用维度，displayMode 为 split 或 cross；未拆分为 normal。
- advanced.openStatistic 控制交叉合计；totalMeasures 用度量 ID，支持 sum/valueCount/max/min/median/avg。
- other.maxRowsPerPage 为正整数；showSummary 控制汇总；otherExtractConfigVOList 项为 `{id,field,type,summaryType,disabled}`。
- 汇总 type=measure/dimension，id 指向该统计字段，field 必须一致；拆分维度不能再作为汇总行字段。summaryType 根据类型取 count/sum/avg/max/min/cardinality；日期用 count/min/max，文本用 count/cardinality。

使用 `--pivot-configs` 校验上述第二阶段请求及其组件绑定。未知视图/权限/查询能力仍需真实接口验证。
