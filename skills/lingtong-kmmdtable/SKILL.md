---
name: lingtong-kmmdtable
description: "通过 lingtong-cli 操作快麦多维表格（/kmMDTable）的模块入口。用于组织应用资源与嵌套文件夹、管理仪表盘、自定义页面开发与草稿回传、角色和资源权限；涉及表格数据、视图或透视表时复用对应 Skill。单纯操作表格记录时优先使用表格 Skill。"
---

# 快麦多维表格

同一应用下有 `table`、`dashboard`、`folder`、`page` 四类资源。本 Skill 维护应用内资源组织和组合操作；表格数据与透视表的具体用法由已有 Skill 维护。

## 授权与目标

按 [lingtong-shared](../lingtong-shared/SKILL.md) 完成 CLI 安装、配置和授权。先用 `lingtong-cli auth status` 检查状态，已有有效授权直接复用；需要登录时引导用户执行 `lingtong-cli auth login`。

从上下文取得目标应用；未指定时查询当前公司绑定应用，不另外创建集成应用。

```bash
lingtong-cli api GET /application/km/multi-table/id --envelope
lingtong-cli api GET /app/resource/list --params '{"appId":942}' --envelope
```

示例 ID 均需替换。资源按 `resourceType + id` 定位；表格资源 ID 是 `basicDataId`，另有字段结构 `schemaId` 和视图 `viewId`；页面资源 ID 就是 `pageId`。

## 按任务读取

| 用户需求 | 指引 |
| --- | --- |
| 查询资源、新建文件夹、嵌套移动、排序、改名、换图标、删除 | [资源与文件夹](references/resources.md) |
| 表格、字段、记录、视图、分组、导入导出与同步 | [lingtong-table](../lingtong-table/SKILL.md) |
| 透视表的维度、度量、聚合及显示配置 | [lingtong-pivot-table](../lingtong-pivot-table/SKILL.md) |
| 仪表盘组件、数据源和布局 | [仪表盘流程](references/dashboards.md) |
| 新建页面或为已有 pageId 开发内容 | [自定义页面开发](references/custom-page.md) |
| 页面管理、草稿回传、发布及历史 | [页面接口与回传](references/page-api.md) |
| 角色、成员、功能权限与资源访问权限 | [角色与资源权限](references/permissions.md) |
| 接口参数从哪里查询、使用哪种调用方式 | [接口索引](references/api-catalog.md) |

只加载任务需要的 Skill 或章节。表格与透视表组合进仪表盘或页面时，沿用同一个目标应用和已确认的资源 ID。

## 本模块的表格约定

手工建表沿用当前应用 `appId`，使用 `source: 1`、`type: 1`、`openHighMode: 1`、`openConnector: 0`；主键在创建时确定。创建、Schema 和数据操作读取表格 Skill，不在此重复维护参数说明。

需要调整表格菜单所属文件夹或顺序时，回到资源章节处理；不要将记录排序、视图排序与资源菜单的 `sortNum` 混用。

## 完成结果

资源可见性以接口返回列表为准；各项操作遵循对应权限和用户要求。页面开发、运行时请求及回传流程见 [自定义页面开发](references/custom-page.md)。

检查业务 `success`（信封输出同时检查 `ok`），写入后回读确认。结果不明确时先查状态，不重复创建。

完成后报告目标应用、资源类型和 ID、实际完成的变更。
