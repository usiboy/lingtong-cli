# lingtong-factory 技能

`lingtong-factory` 用于管理连接器工厂，包括自定义连接器定义、HTTP 接口和脚本。它面向连接器开发和调试，不用于普通连接器账号授权。

## 何时使用

- 查询或创建自定义连接器工厂
- 查看工厂下的 HTTP 接口列表
- 测试运行某个 HTTP 接口
- 查看或删除工厂脚本
- 需要通过 `/factory/*` 专用命令而不是长尾 `service factory` 调用

## 快速开始

```bash
# 找到工厂
lingtong-cli factory list --page 1 --page-size 20
lingtong-cli factory get --id 12

# 查看和测试 HTTP 接口
lingtong-cli factory http list --factory-id 12
lingtong-cli factory http run --data '{"id":88,"params":{"page":1}}'
```

## 命令清单

| 目标 | 命令 |
|------|------|
| 列出工厂 | `lingtong-cli factory list --page 1 --page-size 20` |
| 查询工厂 | `lingtong-cli factory get --id <factory-id>` |
| 创建/更新工厂 | `lingtong-cli factory save --data '<json>'` |
| 删除工厂 | `lingtong-cli factory delete --id <factory-id> --yes` |
| 列出 HTTP 接口 | `lingtong-cli factory http list --factory-id <factory-id>` |
| 查询 HTTP 接口 | `lingtong-cli factory http get --id <http-id>` |
| 运行 HTTP 接口 | `lingtong-cli factory http run --data '<json>'` |
| 删除 HTTP 接口 | `lingtong-cli factory http delete --id <http-id> --yes` |
| 列出脚本 | `lingtong-cli factory script list --factory-id <factory-id>` |
| 查询脚本 | `lingtong-cli factory script get --id <script-id>` |
| 删除脚本 | `lingtong-cli factory script delete --id <script-id> --yes` |

## 数据示例

```bash
# 创建或更新工厂，带 id 表示更新，不带 id 表示创建
lingtong-cli factory save --data '{"name":"my-connector","connector":"mycon"}'
lingtong-cli factory save --data '{"id":12,"name":"renamed"}'

# 测试接口
lingtong-cli factory http run --data '{"id":88,"params":{"page":1,"pageSize":20}}'
```

## 安全规则

- 删除工厂、HTTP 接口、脚本都必须显式传 `--yes`。
- `factory save` 和 `factory http run` 会写入或调用平台资源，执行前确认当前 `--profile`、`--auth`。
- 不要把密钥写入文档或仓库中的示例 JSON。

## 与其他命令的关系

| 需求 | 推荐入口 |
|------|----------|
| 管理自定义连接器定义 | `factory` |
| 管理连接器授权账号 | `connector account` |
| 调用已发布连接器方法 | `connector invoke` |
| 调用未封装的长尾 factory API | `service factory ...` |
| 编写前置/后置脚本 | `lingtong-script` |

## 相关文档

- [SKILL.md](SKILL.md) - Agent 使用指令
- `lingtong-connector` - 连接器账号和接口调用
- `lingtong-script` - 脚本编写规范
