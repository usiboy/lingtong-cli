# lingtong-factory 技能

`lingtong-factory` 用于管理连接器工厂，包括自定义连接器定义、HTTP 接口、DOC/EXECUTE 模型、环境参数和前后置脚本。它面向连接器开发和调试，不用于普通连接器账号授权。

## 何时使用

- 查询、创建或更新自定义连接器工厂。
- 创建 HTTP 方法并补齐请求体、响应体和响应示例模型。
- 保存 EXECUTE 模型并调试运行 HTTP 接口。
- 编写认证链路脚本，将 token 写入环境变量并注入业务接口请求头。
- 排查工厂运行中的 WAF、登录态、Cookie、`Api-*` 请求头问题。

## 快速开始

```bash
lingtong-cli factory list --page 1 --page-size 20 --envelope --format json
lingtong-cli factory get --id <factory-id> --envelope --format json
lingtong-cli factory http list --factory-id <factory-id> --envelope --format json
lingtong-cli factory http get --id <http-id> --envelope --format json
```

创建或更新模型：

```bash
lingtong-cli service factory http-save --body '<doc-json>' --envelope --format json
lingtong-cli service factory http-executemodel-save --body '<execute-json>' --envelope --format json
```

不要把简化的 HTTP run 命令当成可靠调试入口。当前后端 `/factory/http/run` 从 `Api-*` 请求头解析运行参数，详见 `SKILL.md`。

## 命令清单

| 目标 | 命令 |
|---|---|
| 列出工厂 | `lingtong-cli factory list --page 1 --page-size 20` |
| 查询工厂 | `lingtong-cli factory get --id <factory-id>` |
| 创建/更新工厂 | `lingtong-cli factory save --data '<json>'` |
| 删除工厂 | `lingtong-cli factory delete --id <factory-id> --yes` |
| 列出 HTTP 接口 | `lingtong-cli factory http list --factory-id <factory-id>` |
| 查询 HTTP 接口 | `lingtong-cli factory http get --id <http-id>` |
| 保存 DOC 模型 | `lingtong-cli service factory http-save --body '<doc-json>'` |
| 保存 EXECUTE 模型 | `lingtong-cli service factory http-executemodel-save --body '<execute-json>'` |
| 删除 HTTP 接口 | `lingtong-cli factory http delete --id <http-id> --yes` |
| 列出脚本 | `lingtong-cli factory script list --factory-id <factory-id>` |
| 查询脚本 | `lingtong-cli factory script get --id <script-id>` |
| 删除脚本 | `lingtong-cli factory script delete --id <script-id> --yes` |

## JSON 接口模型要点

创建 JSON HTTP 接口时必须写完整模型：

```json
{
  "connectorFactoryId": 12,
  "name": "获取商品",
  "interfaceType": "query",
  "method": "POST",
  "path": "/api/items",
  "headers": [
    {"name":"Content-Type","title":"Content-Type","type":"string","required":true,"enabled":true,"value":"application/json"}
  ],
  "requestBody": {
    "type": "application/json",
    "parameters": [],
    "body": "{\"pageIndex\":1,\"pageSize\":1}",
    "jsonSchema": {
      "type": "object",
      "required": [],
      "primaryKey": [],
      "properties": {
        "pageIndex": {"title":"页码","type":"integer"},
        "pageSize": {"title":"每页数量","type":"integer"}
      }
    }
  },
  "response": {
    "contentFormat": "JSON",
    "statusCode": "200",
    "statusSettingCondition": [],
    "responseExamples": [
      {"name":"默认示例","responseExample":"{\"success\":true}","orderId":1}
    ],
    "jsonSchema": {
      "type": "object",
      "required": [],
      "primaryKey": [],
      "properties": {
        "success": {"title":"是否成功","type":"boolean"},
        "message": {"title":"消息","type":"string"},
        "data": {"title":"返回数据","type":"object","properties":{}}
      }
    }
  }
}
```

关键规则：

- `requestBody.type` 使用 `application/json`，不要使用 `JSON`。
- JSON 请求字段写入 `requestBody.jsonSchema`，不要只写 `requestBody.parameters`。
- 响应字段写入 `response.jsonSchema`。
- 响应示例字段名是 `responseExample`，不是 `body`。
- `responseExamples` 实际必填，缺失会触发后端 `response_examples` 非空错误。
- `/factory/http/save` 保存 DOC 模型，运行前还要保存 EXECUTE 模型。

## URL 和环境参数

后端会为每个工厂自动创建系统内置全局参数 `URL`，标题为 `服务（前置URL）`。通常不要再新增 `baseUrl`、`host`、`domain` 这类重复参数。

推荐做法：

- HTTP 接口 `path` 保存相对路径，例如 `/ierp/api/login.do`。
- 通过 `service factory params-env-save` 为 `test`、`formal` 等环境保存 `URL` 值。
- 认证中间变量如 `appToken`、`accessToken` 使用 `applyConnector=0`。
- token 占位值使用空字符串，不要使用 `null`。

示例：

```bash
lingtong-cli service factory params-env-save --body '[
  {"connectorFactoryId":70,"globalId":173,"env":"test","name":"URL","title":"服务（前置URL）","value":"https://example.com","applyConnector":1}
]' --envelope --format json
```

如果 `params-global-delete --body ...` 返回“请传递Id”，说明该接口的 OpenAPI 与后端参数绑定不一致，可改用：

```bash
lingtong-cli api POST '/factory/params/global/delete?id=<id>' --envelope
```

## 运行参数和 Api Headers

当前后端运行接口从请求头解析：

| 请求头 | 含义 |
|---|---|
| `Api-H0` | 目标接口请求头和 Cookie |
| `Api-O0` | 方法和超时 |
| `Api-U` | 目标完整 URL，可带 query string |
| `Api-S` | `factoryHttpDocId` 和 `factoryHttpEnv` |
| `Api-PreScript` | 前置脚本数组 JSON |
| `Api-PostScript` | 后置脚本数组 JSON |

编码规则：

- `Api-H0`、`Api-O0`、`Api-S` 使用逗号分隔键值对。
- `Api-U` 的 query string 使用 `&` 分隔键值对。
- `Cookie` 值会用分号分隔。
- 键值对按 `=` 拆分且必须恰好两段。
- 值里包含 `,`、`=`、`&`、`;` 时必须 URL 编码。

如果 CLI 版本没有暴露自定义请求头能力，不要用简化 run 命令硬试。应使用支持传原始 header 的代理入口，或先改造 CLI 的 `http-run` 命令。

## 前后置脚本 token 链路

推荐写法：

```javascript
var accessToken = pm.environment.get('accessToken');
if (accessToken !== null && accessToken !== undefined && String(accessToken) !== '') {
  pm.request.headers.upsert('accessToken', String(accessToken));
}
```

不要用对象参数形式注入请求头，例如 `upsert({ key, value })`。原因：`LtAppPmRequestHeader` 公开的 `upsert` 是双参数形式。对象形式可能不会写入 header。

常见模式：

- 认证接口后置脚本解析响应 token。
- 后置脚本执行 `pm.environment.set('accessToken', token)`。
- 业务接口前置脚本读取环境变量并写入请求头。
- 后端 `EnvironmentUpdater` 会把新增或更新的环境变量持久化到当前工厂环境。

## 鉴权和 Web 登录态

- CLI `apk` token 走网关 `AuthSecretFilter` 和 Redis `auth_secret:<token>`，适合 API 调用。
- `apk` token 不会转换成浏览器 `Server-Token` Cookie。
- `Server-Token` 由普通登录或 `/admin/upms/ltapp/login/token` 生成。
- `/admin/upms/ltapp/login/token` 依赖 `TJ_TOKEN` 一次性缓存，不能直接传 `apk` token。
- 上传图片、部分 Web 页面和 SSE 调试可能需要浏览器 `Server-Token`。

## WAF 和外部系统

如果外部系统返回 HTML、418、403 或“访问被拦截”，先按 WAF 排查：

- 对比直连和平台转发结果。
- 补充浏览器风格 `User-Agent`、`Origin`、`Referer`。
- 直连补 header 成功后，同步写入 DOC 和 EXECUTE 模型。
- 日志和总结只输出状态摘要，不输出 token 或 secret。

## 安全规则

- 删除工厂、HTTP 接口、脚本都必须显式传 `--yes`。
- `factory save`、`http-save`、`http-executemodel-save` 和运行接口都会影响平台资源，执行前确认 profile、租户、工厂 ID、接口 ID 和环境。
- 不要把密钥、token、Cookie、`Server-Token` 写入文档或仓库中的示例 JSON。
- 调试输出只保留 `hasValue`、`valueLength`、状态码、错误码和简短消息。

## 与其他命令的关系

| 需求 | 推荐入口 |
|---|---|
| 管理自定义连接器定义 | `lingtong-factory` |
| 管理连接器授权账号 | `lingtong-connector` |
| 调用已发布连接器方法 | `lingtong-connector` |
| 调用未封装的长尾 factory API | `lingtong-service` |
| 编写前置/后置脚本 | `lingtong-script` |

## 相关文档

- [SKILL.md](SKILL.md) - Agent 使用指令
- `lingtong-connector` - 连接器账号和接口调用
- `lingtong-script` - 脚本编写规范
