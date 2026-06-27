# lingtong-shared 技能

`lingtong-shared` 是所有 lingtong-cli 技能的共享入口，负责配置、认证、安全规则、输出契约、Schema 自省、更新检查和 Skills 安装说明。

## 何时使用

- 首次使用 `lingtong-cli`
- 配置平台 Host、切换环境或检查当前配置
- 登录、切换、删除多身份 Token
- 需要机器可读输出、jq 过滤或错误信封
- 需要浏览 OpenAPI、检查版本、安装 Agent Skills
- 遇到认证、权限、配置或连通性问题

## 快速开始

```bash
# 配置 Host
lingtong-cli config init --host https://your-lingtong-host.com

# 使用环境变量登录，避免 Token 出现在 shell history
export LINGTONG_API_TOKEN=apk-xxx
lingtong-cli auth login --from-env --label default

# 检查状态
lingtong-cli auth status
lingtong-cli doctor
```

## 配置与多环境

| 目标 | 命令 |
|------|------|
| 查看当前配置 | `lingtong-cli config show` |
| 添加环境 Profile | `lingtong-cli config profile add --name dev --host https://dev.example.com --token apk-xxx` |
| 列出 Profile | `lingtong-cli config profile list` |
| 切换 Profile | `lingtong-cli config profile use prod` |
| 临时使用 Profile | `lingtong-cli --profile prod scene list` |
| 删除 Profile | `lingtong-cli config profile remove --name dev --yes` |

每个 Profile 可以有独立 Host、Brand 和 Profile 级 Token。运行时 `LINGTONG_TOKEN` 会覆盖配置中的 Token。

## 多身份认证

| 目标 | 命令 |
|------|------|
| 登录并命名身份 | `lingtong-cli auth login --from-env --label company-a` |
| 列出身份 | `lingtong-cli auth list` |
| 切换默认身份 | `lingtong-cli auth use company-a` |
| 临时使用身份 | `lingtong-cli --auth company-a app list` |
| 删除身份 | `lingtong-cli auth remove --label company-a` |
| 登出全部身份 | `lingtong-cli auth logout --all` |

Token 存储在 OS Keychain 中，配置文件只保存身份 label、Token 后缀和租户缓存信息。

## 输出契约

```bash
# 结构化错误信封，适合 Agent 判定 .ok / .error.type
lingtong-cli scene list --envelope

# jq 过滤输出
lingtong-cli schema list --jq '.totalOperations'

# 命令级格式化
lingtong-cli connector info --connector kmerp --format pretty
```

全局标志：`--profile`、`--auth`、`--envelope`、`--jq/-q`、`--omit-null`。

## Schema、更新与 Skills

```bash
# 浏览 OpenAPI
lingtong-cli schema list
lingtong-cli schema module scene
lingtong-cli schema path /scene/list
lingtong-cli schema search 场景

# 检查更新
lingtong-cli update --check

# 安装 Skills
lingtong-cli skills install
lingtong-cli skills status
lingtong-cli skills list
```

## 安全规则

- 不输出 Token 明文。
- 写入、删除、发布、回滚前先确认目标和环境。
- 删除类命令必须使用 `--yes`，导入/导出/回滚优先使用 `--dry-run`。
- Agent 自动化调用优先加 `--envelope`，不要只依赖人类可读文本。

## 相关技能

| 技能 | 说明 |
|------|------|
| `lingtong-service` | 调用长尾 OpenAPI |
| `lingtong-connector` | 连接器与账号授权 |
| `lingtong-table` | 表格与基础资料 |
| `lingtong-workflow` | 工作流编排与执行 |

## 维护说明

- 本 README 负责共享能力导航。
- Agent 执行细则以 [SKILL.md](SKILL.md) 为准。
- 新增全局标志、认证模式或输出契约时，需要同步更新主 README 和所有相关技能示例。
