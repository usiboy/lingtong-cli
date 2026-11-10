# 仪表盘入口

仪表盘组件、数据源、布局、筛选、示例与校验统一由 [lingtong-dashboard](../../lingtong-dashboard/SKILL.md) 维护。

仪表盘仍是当前应用下以 id 定位的 dashboard 资源。本模块处理菜单组织、文件夹、排序、图标及资源权限；创建/编辑仪表盘内容时沿用同一 appId 和已确认的 dashboardId，转入仪表盘技能。表格与透视配置分别复用对应技能。

接口调用与保存后验证见 [仪表盘接口流程](../../lingtong-dashboard/references/api-workflow.md)，公共定义使用 CLI 共享 OpenAPI，避免在这里维护第二份 DTO。
