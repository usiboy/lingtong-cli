---
name: lingtong-service
version: 1.0.0
description: "绫通 CLI 自动生成命令层 (service)：从 OpenAPI 规范自动暴露全部 ~487 个 API 为命令行,覆盖 basicdata/scene/factory/meta/noco 等所有模块。当用户需要调用没有专用命令的长尾 API、或想按模块浏览全部可用接口时触发。"
---

# lingtong-cli 自动生成命令层 (service)

`service` 是从平台 OpenAPI 规范(Swagger 2.0,515 个 operation)自动生成的命令层,把几乎全部 API 暴露为命令行,无需为每个接口手写命令。

## 何时用 service,何时用专用命令

- **优先用专用命令**(`connector info`、`scene list`、`workflow execute` 等):高频操作,参数校验更强、交互更友好。
- **用 `service`**:调用没有专用命令包装的长尾接口,或想按模块浏览全部可用 API。

二者**共享同一套输出契约**:`--envelope` 信封、类型化退出码、`--jq` 过滤、`--omit-null` 全部通用(详见 [[lingtong-shared]])。

## 命令结构

```
lingtong-cli service <模块> <操作> [--参数 ...]
```

模块 = API 路径一级前缀(`scene`、`basicdata`、`factory`、`meta`、`noco`、`task` …);操作名由剩余路径派生(`/basicdata/record/listNew` → `record-listnew`)。

```bash
# 浏览所有模块
lingtong-cli service --help

# 浏览某模块下所有操作
lingtong-cli service scene --help

# 查看某操作的参数(自动从规范推导,含必填标记和中文描述)
lingtong-cli service account connector-info --help
```

## 参数规则(自动推导)

- **query 参数** → 同名 flag,按类型生成(`integer`→`--page 1`,`boolean`→`--active`,`array`→`--ids a,b`)。
- **path 参数**(如 `/meta/model/{connector}/get`)→ 必填 flag,运行时替换进路径:`--connector push_state`。
- **body 参数** → 两种方式:
  - `--body '<JSON>'`:整体传 JSON(复杂结构推荐)。
  - 当 body schema 有具体字段时,也可用 `--<字段名>` 逐个传,CLI 自动组装成 JSON body。
- 必填参数缺失 → 退出码 **2**(validation)。

## 示例

```bash
# GET + query
lingtong-cli service account connector-info --connector kmerp

# GET + path 参数 + jq
lingtong-cli service meta model-connector-get --connector push_state --jq '.result'

# POST + 整体 body
lingtong-cli service basicdata save --body '{"name":"test","modelId":123}'

# POST + 逐字段
lingtong-cli service basicdata save --name test --modelId 123

# 配合信封,供 AI Agent 解析
lingtong-cli service factory http-list --connector kmerp --envelope --jq '.data'
```

## 规范来源(零配置 + 可覆盖)

规范在**编译期内嵌进二进制**,因此 `service` 开箱即用,无需任何配置。

需要使用更新的规范时,按以下优先级覆盖(先命中先用):

1. 环境变量 `LINGTONG_OPENAPI=/path/to/openapi.json`
2. `~/.lingtong-cli/openapi.json`
3. 内嵌副本(默认)

```bash
# 临时使用最新规范
LINGTONG_OPENAPI=./latest-openapi.json lingtong-cli service --help

# 检查当前生效的规范来源与 operation 数量
lingtong-cli doctor   # 看 "Service Spec" 一行
```

## 限制

- 带通配符的路径(如 `/noco/datasource/app/{id}/**`)无法表示为固定命令,**自动跳过**。
- 因命名冲突,极少数 operation 会被去重丢弃(515 → ~487 条命令)。
- 自动生成命令侧重"可发现性 + 基础校验",复杂业务校验仍以专用命令为准。
