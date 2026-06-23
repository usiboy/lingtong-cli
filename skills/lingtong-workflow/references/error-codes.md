# 错误码说明

本文档列出绫通工作流常见错误码及解决方案。

## 工作流错误

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| `1120001` | 工作流不存在 | 检查 workflowId 是否正确 |
| `1120002` | 工作流未发布 | 执行 `workflow publish` |
| `1120003` | 节点执行错误 | 检查节点配置和脚本 |
| `1120004` | 参数验证失败 | 检查 Start 节点参数 |
| `1120005` | 工作流执行超时 | 检查连接器响应时间 |

## 节点错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `NullPointerException` @ `buildFlowSource` | 节点图未持久化（多因用 `/workflow/update/basic` 推 DSL，或发布快照 `nodeDtoList` 为空） | 用 `/workflow/update`（content 为字符串）保存后再发布；详见 [dsl-specification.md 持久化与保存](dsl-specification.md) |
| `断言配置不允许为null` | w_script 节点缺少 `assertConfig` | 每个脚本节点 `data` 加 `assertConfig:{"assertType":"throwException"}` |
| `节点执行失败:XxxResponse(total=null, list=null...)` | ①手动包了 `requestBody` 导致双重包裹 ②pageSize 过小 ③确无数据 | 传【扁平】参数（AppInvoker 自动包裹，勿再包）；pageSize 用 ≥20（建议100）；查询脚本 `try/catch` 仅对 `total=null` 降级；先用 `connector invoke` 单点验证 |
| `页数为空或不符合规定` | pageSize 过小/缺失（如 5） | pageSize 用 ≥20（建议 100） |
| `connector,不能为空` | 连接器节点缺少 connector 配置 | 添加 connector 字段 |
| `method 不能为空` | 方法名为空 | 检查方法名参数 |
| `授权信息不存在` | authAccount 名称错误 | 检查账户名称 |
| `context 信息不完整` | 脚本缺少 _env | 添加 `context.put("_env", env)` |
| `args is not defined` | 脚本引用未定义变量 | 使用 `context.get("key")` |

## 脚本错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `SyntaxError` | ES6+ 语法 | 改用 ES5.1 语法 |
| `ReferenceError` | 变量未定义 | 检查变量名和 inputVariables |
| `TypeError` | 类型错误 | 检查数据类型 |

## DSL 验证错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `缺少开始节点` | 没有 w_start | 添加 w_start 节点 |
| `缺少结束节点` | 没有 w_end | 添加 w_end 节点 |
| `节点类型无效` | 使用了不支持的类型 | 检查 12 种合法类型 |
| `边引用不存在的节点` | source/target 错误 | 检查节点 ID |

## API 错误

| HTTP 状态码 | 说明 | 解决方案 |
|-------------|------|----------|
| 400 | 请求参数错误 | 检查参数格式 |
| 401 | 未授权 | 检查 token |
| 403 | 权限不足 | 检查用户权限 |
| 404 | 资源不存在 | 检查 ID 是否正确 |
| 500 | 服务器错误 | 联系管理员 |

## 连接器错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `连接器调用失败` | 外部系统错误 | 检查外部系统状态 |
| `接口不存在` | method 错误 | 检查方法名 |
| `请求参数错误` | requestBody 格式错误 | 检查参数格式 |
