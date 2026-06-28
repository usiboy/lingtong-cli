---
name: lingtong-factory
version: 1.0.3
description: "绫通连接器工厂 (factory)：自定义连接器定义、HTTP 接口、执行模型、前后置脚本和调试运行。当用户需要管理连接器工厂、创建 HTTP 方法、补齐模型、写脚本、测试运行接口时触发。关键词：factory、连接器工厂、http 接口、executeModel、Api-H0、script、脚本。"
metadata:
  requires:
    bins: ["lingtong-cli"]
  cliHelp: "lingtong-cli factory --help"
---

# lingtong-factory

连接器工厂(`/factory/*`)用于定义自定义连接器，包含工厂本体、HTTP 接口、DOC 模型、EXECUTE 模型、环境参数和前后置脚本。

## 先读结论

- 不要把 `factory http run` 的简化示例当成可靠运行方式。当前后端 `/factory/http/run` 从 `Api-*` 请求头解析运行参数。
- 创建 JSON HTTP 接口时必须同时写 `requestBody.jsonSchema`、`response.jsonSchema` 和 `response.responseExamples[].responseExample`。
- `/factory/http/save` 保存 DOC 模型。运行前还要调用 `service factory http-executemodel-save` 保存 EXECUTE 模型。
- 新建工厂后系统会自动创建全局参数 `URL`。接口 `path` 推荐保存相对路径，通过不同环境的 `URL` 切换域名。
- 脚本里请求头注入使用 `pm.request.headers.upsert('Header-Name', String(value))`，不要用对象参数形式。
- token、secret、Cookie、`Server-Token` 只能在本地请求中使用，不要写入文档、日志、示例输出或总结。

## 工厂本体

```bash
lingtong-cli factory list --page 1 --page-size 20
lingtong-cli factory get --id <factory-id>
lingtong-cli factory save --data '<factory-json>'
lingtong-cli factory delete --id <factory-id> --yes
```

`factory save` 通过 `--data` 是否包含 `id` 区分创建和更新。删除属于破坏性操作，必须显式加 `--yes`。

## HTTP 接口

```bash
lingtong-cli factory http list --factory-id <factory-id>
lingtong-cli factory http get --id <http-id>
lingtong-cli factory http delete --id <http-id> --yes
```

HTTP 接口创建和模型保存通常需要自动生成层：

```bash
lingtong-cli service factory http-save --body '<doc-json>' --envelope --format json
lingtong-cli service factory http-executemodel-save --body '<execute-json>' --envelope --format json
```

未包装的工厂长尾接口优先使用 `lingtong-cli service factory ...` 或 `lingtong-cli api METHOD /path`。

## JSON 接口模型

对 JSON 请求体，不要只写 `requestBody.parameters`。只写 `parameters` 会导致界面和模型元数据缺少请求体字段。

最小正确结构：

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

模型规则：

- `requestBody.type` 使用后端枚举值 `application/json`，不要写 `JSON`。
- 请求体字段写入 `requestBody.jsonSchema`，后端保存为 `REQUEST_BODY` 模型。
- 响应字段写入 `response.jsonSchema`，后端保存为 `RESPONSE` 模型。
- 响应示例字段名是 `responseExample`，不是 `body`。
- `responseExamples` 实际必填，缺失会触发数据库 `response_examples` 非空错误。
- DOC JSON 和 EXECUTE JSON 字段应保持一致，避免界面展示和实际运行不一致。

## URL 和环境参数

新建工厂后，后端会自动创建系统全局参数 `URL`：

```text
name=URL
title=服务（前置URL）
system=1
applyConnector=1
```

推荐做法：

- 接口 `path` 保存相对路径，例如 `/ierp/api/login.do`。
- 不要重复创建 `baseUrl`、`host`、`domain` 这类前置 URL 参数。
- 不同环境域名写入 `URL` 的环境值，例如 `test`、`formal`。
- 认证中间变量如 `appToken`、`accessToken` 建议 `applyConnector=0`。
- token 占位值用空字符串，不要用 `null`，避免脚本持久化新增同名字段时报重复。

保存环境参数示例：

```bash
lingtong-cli service factory params-env-save --body '[
  {"connectorFactoryId":70,"globalId":173,"env":"test","name":"URL","title":"服务（前置URL）","value":"https://example.com","applyConnector":1}
]' --envelope --format json
```

删除全局参数注意：`params-global-delete` 的 OpenAPI 可能把 `id` 标成 body，但后端实际是普通请求参数。若自动生成命令返回“请传递Id”，可用：

```bash
lingtong-cli api POST '/factory/params/global/delete?id=<id>' --envelope
```

## 运行 HTTP 接口

可靠流程：

```bash
lingtong-cli service factory http-save --body '<doc-json>' --envelope --format json
lingtong-cli service factory http-executemodel-save --body '<execute-json>' --envelope --format json
```

然后调用后端 `/factory/http/run`，运行参数必须放在请求头：

| 请求头 | 含义 | 示例 |
|---|---|---|
| `Api-H0` | 目标接口请求头和 Cookie | `Content-Type=application/json,accessToken=<redacted>` |
| `Api-O0` | 方法和超时 | `method=POST,timeout=30000` |
| `Api-U` | 目标完整 URL，可带 query string | `https://example.com/api/items?page=1` |
| `Api-S` | 工厂运行系统参数 | `factoryHttpDocId=526,factoryHttpEnv=test` |
| `Api-PreScript` | 前置脚本数组 JSON | `[{"type":"customScript","enabled":true,"extendScriptId":44,"data":""}]` |
| `Api-PostScript` | 后置脚本数组 JSON | `[{"type":"customScript","enabled":true,"extendScriptId":45,"data":""}]` |

`HttpRunAnalyzer.analysisValue` 的解析规则很脆弱：

- `Api-H0`、`Api-O0`、`Api-S` 默认用逗号分隔键值对。
- `Api-U` 的 query string 用 `&` 分隔键值对。
- `Cookie` 值会再用分号分隔。
- 每个片段再按 `=` 拆分，且必须恰好拆成两段。
- 值里如果包含 `,`、`=`、`&`、`;`，必须先 URL 编码。
- 典型风险值包括 base64 token、复杂 Cookie、带逗号的 `User-Agent`、签名串和回调 URL。

如果当前 CLI 命令没有暴露自定义请求头能力，不要用旧的简化 run 示例硬试。应使用支持传原始 header 的代理入口，或先提出 CLI 改造，让 `http-run` 显式支持 `Api-*` headers 和请求 body。

## 前后置脚本

HTTP 工厂脚本运行在 ECMAScript 5.1/Nashorn 环境，使用 Postman 风格 `pm` 对象。

环境变量：

```javascript
pm.environment.has('accessToken');
pm.environment.get('accessToken');
pm.environment.set('accessToken', token);
```

请求头注入：

```javascript
var accessToken = pm.environment.get('accessToken');
if (accessToken !== null && accessToken !== undefined && String(accessToken) !== '') {
  pm.request.headers.upsert('accessToken', String(accessToken));
}
```

不要使用对象参数形式，例如 `upsert({ key, value })`。`LtAppPmRequestHeader` 公开的 `upsert` 是双参数形式。对象形式可能不会真正写入 header，最终导致业务接口返回未授权。

后置脚本解析响应：

```javascript
var json = pm.response.json();
if (json && json.data && json.data.access_token) {
  pm.environment.set('accessToken', String(json.data.access_token));
}
```

环境变量持久化规则：

- 后端会比较脚本执行前后的 `pm.environment`。
- 新增变量会创建全局参数，默认 `system=0`、`applyConnector=0`。
- 已有变量会按当前运行环境更新值。
- 删除变量会按名称删除全局参数，谨慎使用。

## 鉴权和登录态边界

- CLI 常用的 `apk` token 属于网关密钥模式。
- 网关 `AuthSecretFilter` 校验 Redis `auth_secret:<token>`，适合 `/gw/ai/proxy` 这类 API 调用。
- `apk` token 不会自动转换成浏览器 `Server-Token` Cookie。
- `Server-Token` 是普通登录或 `/admin/upms/ltapp/login/token` 等登录响应生成的 JWT Cookie。
- `/admin/upms/ltapp/login/token` 消费用户中心 `TJ_TOKEN` 一次性缓存，不能直接传 `apk` token。
- 文件上传、部分 Web 页面和 SSE 调试接口可能只接受浏览器 `Server-Token`，此时不要假设 CLI token 可用。

## 图标和文件上传

- 工厂 `icon` 字段不适合直接写 base64 data URI，容易触发数据库字段过长。
- 推荐先通过 Web 上传接口拿到图片 URL，再把 URL 写入工厂 `icon`。
- 上传接口通常需要浏览器 `Server-Token` Cookie。
- 如果只有 CLI `apk` token，先无图标创建工厂，后续由浏览器登录态补图标 URL。

## WAF 和外部系统排查

外部系统如果返回 HTML、418、403 或“访问被拦截”，先判断是否是 WAF，而不是业务认证失败。

排查规则：

- 对比直连和平台转发结果，确认是否都被拦截。
- 尝试补充浏览器风格 `User-Agent`、`Origin`、`Referer`。
- 如果补充请求头后直连成功，也要把这些 header 同步写入 DOC 和 EXECUTE 模型。
- 只输出状态码、错误码、token 是否存在和 token 长度，不输出 token 明文。

## 安全规则

- 删除工厂、HTTP 接口、脚本必须显式传 `--yes`。
- 创建、更新、运行前确认当前 profile、租户、工厂 ID、接口 ID 和环境名。
- 调试认证链路时只输出 `hasValue`、`valueLength`、`success`、`errorCode`、`message` 等摘要。
- 不要把 token、secret、Cookie、`Server-Token` 写入技能文档、项目文档、提交信息或终端总结。
- 写入或发布平台资源前优先使用 `--dry-run`，如果命令支持。

## 与其他技能

| 需求 | 推荐入口 |
|---|---|
| 管理自定义连接器定义和 HTTP 方法 | `lingtong-factory` |
| 管理连接器账号授权 | `lingtong-connector` |
| 调用已发布连接器方法 | `lingtong-connector` |
| 调用未封装 factory API | `lingtong-service` |
| 编写脚本节点或字段函数 | `lingtong-script` |
| 编排工作流 | `lingtong-workflow` |
