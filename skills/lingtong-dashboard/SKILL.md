---
name: lingtong-dashboard
description: "通过 lingtong-cli 创建和修改绫通仪表盘，配置指标卡、柱状图、条形图、折线图、面积图、饼图、基础表格、透视表和切片器，处理数据绑定、筛选联动、布局及配置校验。用于经营看板、仪表盘组件和已有看板调整；应用菜单组织交给 lingtong-kmmdtable。"
---

# 绫通仪表盘

将业务需求转换成平台仪表盘配置，用 CLI 保存与回读；页面由平台已有 dashboard 渲染器展示。支持九类组件，不需要生成 React 页面。

## 工作流程

1. 按 [lingtong-shared](../lingtong-shared/SKILL.md) 检查授权，复用用户指定的应用、仪表盘；未指定应用时由 [lingtong-kmmdtable](../lingtong-kmmdtable/SKILL.md) 查询当前绑定应用。编辑已有仪表盘先读取最新详情。
2. 按 [组件选型](references/components.md) 确定组件、业务统计口径和筛选。通过 [表格技能](../lingtong-table/SKILL.md) 查询真实表格、Schema、字段 key/ID；指标时间依据、聚合字段或业务口径无法推断时再向用户澄清。
3. 按 [配置与布局](references/config-and-layout.md) 生成或合并配置。只加载本任务涉及的组件参考；示例 ID 和数据名称均为虚构，使用前替换全部表格、Schema、字段、视图、应用和仪表盘引用。
4. 使用 [离线校验](references/validation.md)，优先同时提供真实元数据。先修复 errors；warnings 中“视图未绑定”可在首次保存后解决。校验不访问网络、不改写配置，也不证明线上查询成功。
5. 按 [接口流程](references/api-workflow.md) 保存，检查业务 success 和信封 ok，回读服务端配置。表格组件取回 tableViewId 后再配置透视统计。查询数据并核对各组件结果、时间边界和浏览器显示。
6. 汇报 appId、dashboardId、实际变更及验证结果。结果不明时先回读，不重复创建；保存过期快照可能覆盖其他编辑，提交前合并最新配置。

用户只要求分析、预览或导出时，交付相应配置即可。用户已授权创建或修改时按授权范围执行，无需重复确认。仪表盘保存直接生效，没有自定义页面的草稿/发布步骤。

## 按任务读取

| 任务 | 参考 |
| --- | --- |
| 选组件、查看示例覆盖 | [components.md](references/components.md) |
| 配置对象、ID、24 列布局、局部修改 | [config-and-layout.md](references/config-and-layout.md) |
| 五种统计图、维度/度量、拆分/交叉/组合图 | [charts.md](references/charts.md) |
| 单/多指标、同环比、格式、占比 | [metric-card.md](references/metric-card.md) |
| 基础表格、透视绑定与第二阶段统计配置 | [tables-and-pivot.md](references/tables-and-pivot.md) |
| 全局/局部过滤、文本/数字/日期切片器 | [filters-and-slicers.md](references/filters-and-slicers.md) |
| 真实数据元数据、校验命令及范围 | [validation.md](references/validation.md) |
| CLI 请求、保存、回读、查询与兼容差异 | [api-workflow.md](references/api-workflow.md) |

权限、文件夹、菜单顺序和资源管理交给 [lingtong-kmmdtable](../lingtong-kmmdtable/SKILL.md)；已有表格/透视表能力复用对应技能。公共接口定义以 CLI 内嵌 OpenAPI 为单一来源，本技能补充前端配置语义、示例与校验。
