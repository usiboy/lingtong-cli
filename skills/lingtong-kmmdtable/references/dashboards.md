# 仪表盘入口

仪表盘组件、数据源、布局、筛选、示例与校验统一由 [lingtong-dashboard](../../lingtong-dashboard/SKILL.md) 维护。接口和配置 DTO 统一维护在 CLI 共享规范，通过 [接口索引](api-catalog.md) 查询。

仪表盘仍是当前应用下以 id 定位的 dashboard 资源。本模块处理菜单组织、文件夹、排序、图标及资源权限；创建/编辑仪表盘内容时沿用同一 appId 和已确认的 dashboardId，转入仪表盘技能。表格与透视配置分别复用对应技能。

接口调用与保存后验证见 [仪表盘接口流程](../../lingtong-dashboard/references/api-workflow.md)，公共定义使用 CLI 共享 OpenAPI，避免在这里维护第二份 DTO。

## 不保存配置直接查询

请求的所有后端字段、切片器、指标卡、图表及透视表配置详见 [直接查询参数](direct-query.md)。

使用 `POST /dashboard/query/direct`，请求包含 `appId`、与保存接口相同的 `config`，以及可选的 `slicers`（格式与 `/dashboard/query` 相同）。租户从登录用户获取；接口不会创建仪表盘或关联视图，响应与仪表盘查询一致。请根据数据源的实际字段和组件类型组装配置：

```bash
lingtong-cli api POST /dashboard/query/direct --data '{"appId":942,"config":{"globalFilterList":[],"tableConfigList":[{"id":"chart1","title":"订单数","type":"chart","displayType":"barCharts","tableId":1568,"dataConfig":{"dimensions":[],"measures":[{"id":"m1","type":"count","disabled":false}]}}]}}' --envelope
```

示例查询表格 1568 的记录数；请将 `appId` 和 `tableId` 替换为实际值。字段类型配置由服务端生成，无需提交或保存。与现有仪表盘查询一样，基础表格和透视表视图组件不由此接口直接返回行数据；需要通过相应表格接口查询，透视表可使用 [临时配置查询](../../lingtong-pivot-table/SKILL.md)。单个图表失败时现有查询流程会跳过该图表，请检查响应及服务日志。
