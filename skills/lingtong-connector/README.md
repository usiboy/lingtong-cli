# lingtong-connector 技能

`lingtong-connector` 用于连接器查询、类目与模型探索、授权账户管理、连接器方法 Schema 查询和通用连接器方法调用。

## 何时使用

- 查询连接器详情、分类、接口模型
- 列出、创建、验证连接器授权账户
- 判断工作流、场景或连接器是否缺少授权
- 查询某个连接器可调用的方法列表
- 查询方法的请求/响应字段结构
- 直接调用连接器方法并查看结果

## 快速开始

```bash
# 查询连接器详情
lingtong-cli connector info --connector kmerp

# 查账号与方法
lingtong-cli connector account list --connector kmerp --env test
lingtong-cli connector methods --connector kmerp --app-id 165 --auth-account-id 296

# 查方法字段并调用
lingtong-cli connector schema --connector kmerp --method erp.warehouse.list.query --auth-account-id 296
lingtong-cli connector invoke --connector kmerp --method erp.warehouse.list.query --auth-account "广州力人服饰" --env test --body '{}'
```

## 命令清单

| 目标 | 命令 |
|------|------|
| 查询连接器详情 | `lingtong-cli connector info --connector <name>` |
| 列出账户 | `lingtong-cli connector list [--app-id <id>]` |
| 查询类目 | `lingtong-cli connector category list --connector <name>` |
| 列出授权账户 | `lingtong-cli connector account list [--connector <name>] [--env test|prod]` |
| 创建授权账户 | `lingtong-cli connector account create --connector <name> --name <name> --data '<json>'` |
| 验证授权账户 | `lingtong-cli connector account verify --connector <name> --account-id <id>` |
| 检查授权状态 | `lingtong-cli connector check-auth --workflow-id <id>` |
| 列出方法 | `lingtong-cli connector methods --connector <name> --app-id <id> --auth-account-id <id>` |
| 查询字段 Schema | `lingtong-cli connector schema --connector <name> --method <method> --auth-account-id <id>` |
| 调用方法 | `lingtong-cli connector invoke --connector <name> --method <method> --auth-account <account-name> --body '<json>'` |
| 查看旧本地缓存 | `lingtong-cli connector cache list` |
| 清理旧本地缓存 | `lingtong-cli connector cache clear [--key <env>]` |

## 通用 invoke 机制

`connector invoke` 使用通用代理调用 `/ltapp/connector/invoke` 对应的控制器，无需创建工作流。代理请求路径为 `/gw/connector/invoke`。

`--auth-account` 映射为 `authName`，内置连接器可省略；`--body` 映射为 `requestBody`，默认 `{}`，传业务参数即可。环境默认 `prod`。

完整保留接口响应：检查 `result.receiptStatus`，从 `result.responseBody` 读取业务数据。外层成功不代表业务成功。

`--force` 保留为弃用的空操作。`connector cache list/clear` 仅用于旧本地缓存，不删除平台工作流，新调用不使用缓存。

## 常见排错

| 现象 | 处理 |
|------|------|
| `method not found` | 先运行 `connector methods` 确认 method 名称 |
| 账号不可用 | 运行 `connector account verify` |
| 返回字段不了解 | 运行 `connector schema` 或 `model interface list` |
| invoke 结果异常 | 检查 `result.receiptStatus` 和 `result.responseBody`；结果未知时先核实 |
| 权限错误 | `auth status` 检查 Token，必要时切换 `--profile` 或 `--auth` |

## 相关技能

| 技能 | 说明 |
|------|------|
| `lingtong-model` | 深入查询模型元数据 |
| `lingtong-workflow` | 工作流编排与 AppInvoker DSL |
| `lingtong-factory` | 自定义连接器工厂 |
| `lingtong-script` | AppInvoker 脚本模式 |

## 维护说明

- Agent 执行细则以 [SKILL.md](SKILL.md) 为准。
- 新增连接器命令时同步更新主 README 和本 README。
