# 配置、ID 与布局

## 保存模型

`POST /dashboard/save` 的新增/修改请求使用 `{appId, id?, name?, config}`。config 是对象，不是 JSON 字符串：

```json
{
  "globalFilterList": [],
  "layoutList": [],
  "tableConfigList": []
}
```

`layoutList` 与 `tableConfigList` 一一对应。每个组件使用唯一非空字符串 `id`，布局 `i` 必须相同；保留服务端已有 ID。`type`/`displayType` 见组件表。普通组件含 `tableId`、`tableFilters`、`dataConfig`、`styleConfig`；指标卡的持久化数据源位于 `dataConfig.sources[]`，不依赖顶层 tableId。

## ID 对照

| 名称 | 含义/取得方式 |
| --- | --- |
| appId | 目标应用；不要为缺省目标新建应用 |
| dashboardId / 请求 id | 仪表盘主键；详情与查询使用 id |
| 组件 id / 布局 i | 客户端生成且相互对应的字符串 |
| tableId / basicDataId | 真实业务表格主键 |
| schemaId | 该表结构 ID，不能直接用 tableId 替代 |
| columnId | Schema 中列的数字 ID，用于过滤条件 |
| field | Schema 列的 key，用于度量、维度、切片器、日期依据；例如示例 data2001.2 |
| tableViewId / viewId | 服务端创建/返回的视图 ID，用于嵌入表格和透视统计 |
| measure.id / dimension.id / source.id | 组件内部唯一字符串，排序、拆分、结果匹配和占比使用它 |

不要从展示名猜字段 key，不要拼接 `data${schemaId}.${columnId}` 作为真实字段，直接采用 Schema 返回值。数字 ID 使用正整数；超出 JavaScript 安全整数范围时使用十进制字符串并确认服务端兼容，不能先转 Number 丢失精度。

## 布局

固定 24 列，rowHeight=30 像素，另有布局间距，不能直接把像素当作 w/h。`x/y` 为非负整数，`w/h` 为正整数，`x+w<=24`；不重叠；遵守条目已有的 minW/minH/maxW/maxH。

新增默认 minW=4，普通组件 minH=5，切片器 minH=3。经营概览示例采用顶部筛选、指标、图表、明细和透视的排布。布局过小即使 JSON 有效也可能无法展示，需要在浏览器验证。

## 局部修改与兼容

- 改布局：读取最新详情 → 按 i 合并指定条目的 x/y/w/h → 保存最新完整 config，保留筛选、组件、未知扩展字段与其他布局限制。
- 改组件：按 id 合并，不用一个最小示例替换整张仪表盘。切换表格时清理旧字段/过滤/视图引用，回读新的 tableViewId。
- 删组件：同时删布局；确认切片器的表格仍有活动数据组件引用；删指标源时修正自定义分母。校验器会检查最终引用关系。
- 重命名：只提交 `{id,appId,name}`，不发送空 config。
- 指标卡新旧结构和切片器旧样式字段按对应章节处理。读取不会要求全量迁移，编辑相关部分时再规范化。

校验器不擅自删除未知字段，也不会自动修复或迁移。它只接收完整 config 或含 config 的保存对象，单纯重命名请求无需用它校验。
