# 角色、成员与资源访问权限

功能权限控制允许执行的操作；资源访问权限控制可见的应用菜单。资源是否可见以资源接口列表为准，不自行附加过滤。

## 功能权限

| 操作 | 权限码 |
| --- | --- |
| 文件夹创建/编辑/删除 | `kmMDTable:folder:create` / `kmMDTable:folder:edit` / `kmMDTable:folder:delete` |
| 页面创建/编辑/删除/发布 | `kmMDTable:page:create` / `kmMDTable:page:edit` / `kmMDTable:page:delete` / `kmMDTable:page:publish` |
| 表格创建/编辑/删除 | `kmMDTable:table:create` / `kmMDTable:table:edit` / `kmMDTable:table:delete` |
| 仪表盘创建/编辑/删除 | `kmMDTable:dashboard:create` / `kmMDTable:dashboard:edit` / `kmMDTable:dashboard:delete` |
| 数据更新/删除 | `kmMDTable:updateData` / `kmMDTable:deleteData` |
| 表格设置 | `kmMDTable:setting` |
| 角色列表/创建/编辑/删除 | `kmMDTable:setting:roleList` / `kmMDTable:setting:create` / `kmMDTable:setting:edit` / `kmMDTable:setting:delete` |
| 成员列表/编辑 | `kmMDTable:setting:memberList` / `kmMDTable:setting:editMember` |

改名、换图标、拖动排序按对应资源编辑权限控制。进入快麦多维表格表格时不使用旧表格弹窗的独立查看/编辑/删除/加密字段权限开关；功能操作权限仍有效。

页面草稿与版本面板还受作者/管理员身份限制；发布不授予页面或其数据源访问权限，详见 [页面展示](custom-page.md#草稿预览与发布展示)。

## 角色与成员接口

以下是宿主的完整 Web 路径和实际参数。尚未确认这些 UPMS 接口的 CLI 网关映射，不能自行去掉前缀当成 `/role/*` 或 `/user/*` 调用。接到权限写入任务时，使用已确认支持这些契约的工具；没有对应能力时说明缺失接口，不声称已完成授权。

| 操作 | 方法与 Web 路径 | 参数 / 结果 |
| --- | --- | --- |
| 角色列表 | `GET /admin/upms/ltapp/role/list` | 无参数，`result` 为列表 |
| 新建角色 | `POST /admin/upms/ltapp/role/add` | body `{name,menuIds}` |
| 功能权限列表 | `GET /admin/upms/ltapp/role/scoped/menu/list` | query `topMenuNames=快麦多维表格` |
| 角色功能权限 | `GET /admin/upms/ltapp/role/scoped/info/get` | query `roleId,topMenuNames`；返回 `name,manage,menuIds` |
| 修改角色 | `PUT /admin/upms/ltapp/role/scoped/modify/` | query `topMenuNames=快麦多维表格`；body `{id,name,menuIds}` |
| 删除角色 | `DELETE /admin/upms/ltapp/role/deleted` | query `roleId` |
| 角色成员 | `GET /admin/upms/ltapp/user/list/byRole` | query `roleId` |
| 成员列表 | `GET /admin/upms/ltapp/user/list` | query `pageNum,pageSize`；结果在 `result.rows` |
| 成员详情 | `GET /admin/upms/ltapp/user/editUserPre` | query `userId`；返回 `roles,roleIds,sysUserDto` |
| 修改成员角色 | `PUT /admin/upms/ltapp/user/modify` | body `{id,roleIds,status,topMenuNames}`；沿用详情中的 `status`，范围为 `快麦多维表格` |

角色列表的每项包含 `id`、`name`；功能权限列表包含 `id`、`pId`、`name`、`perms`，其中 `perms` 是权限码。角色和成员接口的 ID 使用整数，`menuIds`、`roleIds` 是整数数组。修改角色先读取并保留未修改的授权；`menuIds` 传菜单 ID，不传权限码。新角色先完成创建，再以返回的角色 ID 配置资源权限。

## 资源访问权限

查询：`GET /app/resource/role/privilege/list`，query `{appId,roleId}`。`result` 为按菜单顺序排列的资源数组，每项含 `id,resourceType,parentFolderId,scene,sceneId,permit` 等。

更新：`PUT /admin/upms/ltapp/dataprivilege/update`，body 为**变更数组**。每项结构：

```json
{
  "busiId": "123",
  "roleId": "8",
  "scene": "使用列表返回的 scene",
  "sceneList": [{"sceneId": null, "permit": 1}]
}
```

`busiId` 为资源 ID 的字符串；`roleId` 也按字符串提交。`scene`、`sceneId` 从资源权限列表取得，缺少 `sceneId` 时传 `null`；`permit` 为 `1`（允许）或 `0`（不允许）。只提交相对原列表变化的资源。当前通用 CLI 的数组请求体能力未就绪，不将这个数组序列化成普通字符串冒充提交。

目录规则：

- 勾选文件夹作用于当前子资源；资源移入不自动获权，移出不自动撤权。
- 父目录可以仅作为已授权子资源的导航，不因此授权兄弟资源。
- “目录独立可见”仅授予文件夹本身；取消该标记不撤销子资源权限。空文件夹也能独立授权。
- 管理员资源权限也可修改，不预设全选或禁用；角色功能权限的管理员规则仍按系统返回处理。

权限结果以服务端响应为准。资源权限更新、角色保存和成员角色更新是独立操作，分别确认结果。
