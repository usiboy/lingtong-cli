# 接口索引与调用方式

## 已有共享能力

表格、视图、透视表、仪表盘及连接器、模型、下载接口统一维护在 CLI 的 `internal/openapi/spec/openapi.json`，本 Skill 不复制这些接口及数据结构。通过已安装 CLI 的 `schema` 查询接口摘要，通过专用命令或 `service --help` 查询可用参数；需要完整 DTO 时查看共享规范的对应 `paths` 和 `definitions`。

| 能力 | 使用指引 | 参数查询入口 |
| --- | --- | --- |
| 表格、字段、记录、视图、分组 | [表格 Skill](../../lingtong-table/SKILL.md) | `lingtong-cli table --help`；`schema module basicdata`；`schema module tableView` |
| 透视表 | [透视表 Skill](../../lingtong-pivot-table/SKILL.md) | `lingtong-cli table pivot --help` |
| 仪表盘 | [仪表盘技能](../../lingtong-dashboard/SKILL.md) | `lingtong-cli service dashboard --help`；`schema module dashboard` |
| 连接器、模型等业务取数 | 按需求使用对应业务 Skill | `lingtong-cli schema search '关键词'` |
| 没有专用命令的接口 | [service Skill](../../lingtong-service/SKILL.md) | `lingtong-cli service <模块> <操作> --help` |

例如：

```bash
lingtong-cli schema path /basicdata/save
lingtong-cli schema path /dashboard/save
lingtong-cli service dashboard query --help
```

## 本模块补充接口

共享规范尚未收录的接口，其方法、参数和关键返回字段直接维护在对应章节：

- [资源与文件夹](resources.md)：绑定应用、资源列表、文件夹增删改、菜单移动排序、图标替换。
- [页面接口与回传](page-api.md)：页面增删改查、草稿源码、保存、HTML 加载、发布和历史。
- [角色与资源权限](permissions.md)：资源授权、角色及成员管理，包含完整 Web 路径与调用限制。

资源和页面的 JSON 接口使用通用 CLI 业务路径；权限章节中的 UPMS 接口尚未确认 CLI 网关映射；草稿 HTML 由浏览器加载。接口已记录不代表已经通过 CLI 实测。

## 选择调用方式

优先使用专用命令，其次使用已收录在共享规范中的 `service` 命令；本模块补充的 JSON 对象接口使用 `api`。

- `api GET` 用 `--params`；POST/PUT 的 JSON 对象请求体用 `--data`。
- 当前通用 `api` 不转发 POST/PUT/DELETE 的 `--params`，需要 query 时将参数 URL 编码后放入路径。
- 通用 `api --data` 不支持顶层数组；使用已支持的专用命令或 `service --body`。multipart 导入需要支持表单的调用能力，不能当作 JSON 提交。
- 不将 UPMS Web 路径猜成 CLI 业务路径；授权和能力边界见权限章节。

接口有差异时修正其所属章节或共享规范；同一接口不再另建一份参数字典。
