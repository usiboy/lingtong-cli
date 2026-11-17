# 自定义页面开发与草稿回传

页面资源类型为 `page`，资源的 `id` 就是 `pageId`。开发前按主 Skill 完成 CLI 授权。

## 数据流程与页面开发

1. 按需求使用 CLI 的表格、连接器、连接器工厂、场景、模型等模块及对应业务 Skill 查找接口，获取真实数据，确认参数、字段和分页范围。
2. 平台提供公共后端服务，Agent 只需交付一个 UTF-8 HTML 文件。技术栈不限，JS、CSS 和所需框架依赖全部内联，不依赖外部 JS/CSS 或本机文件。
3. HTML 只实现业务内容，不重复实现宿主菜单和页面管理功能。适配桌面与移动端，检查主要交互、数据展示及加载、空数据、失败状态。

基于已有表格生成页面时，先核对同一应用中的 `basicDataId`、Schema 字段 ID 与类型，以及需要使用的 `viewId`。页面按这张表的真实字段和查询范围动态取数；ERP 数据写入表格与页面读取表格是两个步骤，页面刷新不自动触发 ERP 重新拉取或同步。

用户要求先预览时，先提供预览，按其确认顺序回传；已明确要求回传时直接完成授权范围内的操作。默认只保存草稿，发布需用户明确要求。

## 页面运行时请求约定

Agent 开发取数和保存草稿通过 CLI，页面运行时通过浏览器业务接口取数。

- 使用当前部署环境的平台 Web 业务接口，例如 `/admin/ltapp/...`；不使用 `/gw` 网关，也不把 `/gw` 作为失败重试或备用前缀。
- 登录态由浏览器携带当前 Cookie。页面不从 Cookie、`localStorage` 或 `sessionStorage` 中读取、缓存 Token，不手动拼接 `Authorization: Bearer ...` 或设置 `Cookie` 请求头，也不将 CLI Token 写入 HTML。
- 使用 Fetch 时设置 `credentials: 'include'`；使用 Axios 时设置 `withCredentials: true`。

下面的 `businessUrl` 是按当前部署环境确认的 Web 业务接口地址：

```js
const response = await fetch(businessUrl, {
  method: 'GET',
  headers: { Accept: 'application/json' },
  credentials: 'include'
});
```

查看者需要有效的浏览器登录态及页面、数据源相应权限；CLI 授权或页面发布不替查看者授予表格数据权限。

## 新建页面

1. 从用户上下文或绑定应用查询取得目标 `appId`，开发并验证 HTML。
2. 调用 `POST /page/create`，传 `appId`、`name`、`parentFolderId`、`icon`。默认根目录为 `parentFolderId: null`，默认图标为 `i-page`；名称按应用避重，未指定时可使用接口默认命名。
3. 创建成功后，从 **`result.id`** 取得并记录 `pageId`。
4. 调用 `POST /page/draft/save`，传该 `pageId`、首次基准版本 `baseVersionNo: 0` 和完整代码 `htmlContent`。创建成功但保存失败时，继续使用已有 `pageId`，不重复创建。

## 更新已有页面

1. 用户提供了 `pageId`，或从 `通过 AI 创建（pageId: xxx）` 进入时，直接更新这个页面，不调用创建接口。
2. 调用 `GET /page/get`，查询参数为 **`id`**，确认页面归属与草稿状态。
3. 已有草稿时调用 `GET /page/draft/source`，查询参数为 **`pageId`**，读取 `htmlContent` 和 **`versionNo`**，在原内容上按需求修改。确认从未保存过草稿时，基准版本为 `0`；读取失败不能当作没有草稿。
4. 调用 `POST /page/draft/save`，传原 `pageId`、**编辑前读取的 `versionNo` 作为 `baseVersionNo`**、完整 `htmlContent`。版本冲突时重新读取并合并修改，不仅替换版本号强行覆盖。

## 回传与结果

按 [页面接口与回传](page-api.md#保存草稿) 的示例保存代码，并完成其中的 [保存结果校验](page-api.md#保存结果)。

完成后告知本地 HTML 文件、`appId`、`pageId`、实际草稿/发布版本、保存结果及查看方式；仅保存草稿时说明是否需要切换到“最新草稿”，不把保存成功报告为普通查看者已能看到新内容。尚未成功的步骤如实说明。

## 草稿预览与发布展示

当前宿主仅向页面作者（资源 `isAuthor`）或管理员（用户 `manage`）开放草稿及版本面板；编辑还需 `kmMDTable:page:edit`，发布还需 `kmMDTable:page:publish`。这些是前端展示条件，实际操作仍以服务端权限校验为准。

| 页面状态 | 作者或管理员进入/刷新页面 | 普通查看者 |
| --- | --- | --- |
| 尚无草稿、也未发布 | 空页面；有编辑权限时显示“通过 AI 创建”入口 | 显示“页面尚未发布” |
| 已保存草稿、从未发布 | 默认预览草稿 | 显示“页面尚未发布” |
| 已有发布版本，可能另有新草稿 | 默认展示最新发布版本；在“版本/发布”面板选择未发布草稿或历史发布版本 | 展示最新发布版本 |

保存新草稿不替换已发布内容。处于发布版本预览时，新草稿回传后仍显示所选发布版本；需要切换到“最新草稿”核对。页面刷新重新按上表选择版本，回传本身不保证浏览器立即刷新；可重新进入页面或使用刷新按钮更新状态。

前端将 `draftVersionNo > 0` 且与 `publishedVersionNo` 不同视为有未发布草稿；发布后即使保留草稿元数据，也不能据此认定还有待发布修改。分享入口仅加载最新发布版本，并检查登录态和访问权限，不展示草稿，也不是匿名公开访问。
