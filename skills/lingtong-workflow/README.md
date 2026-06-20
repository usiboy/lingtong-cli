# lingtong-workflow 技能

绫通工作流全生命周期管理技能，提供工作流设计、创建、验证、测试、发布、监控的完整能力。

## 目录结构

```
lingtong-workflow/
├── SKILL.md                    # 主技能文档
├── README.md                   # 本文件
└── references/                 # 参考文档
    ├── node-details.md         # 12种节点详细配置
    ├── dsl-specification.md    # DSL 完整规范
    ├── script-guide.md         # 脚本编写指南（ES5.1）
    ├── connector-invoke.md     # 通用连接器调用模式
    ├── template-guide.md       # 模板使用指南
    ├── testing-guide.md        # 测试指南
    ├── error-codes.md          # 错误码说明
    └── examples/               # 示例文档
        ├── simple-workflow.md  # 简单工作流
        ├── connector-api.md    # 连接器API化
        └── data-pipeline.md    # 数据管道
```

## 快速开始

### 1. 创建工作流

```bash
# 使用模板
lingtong-cli workflow create --name "我的流程" --template simple

# 使用 DSL 文件
lingtong-cli workflow create --name "自定义" --dsl-file workflow.json
```

### 2. 验证 DSL

```bash
lingtong-cli workflow validate --dsl-file workflow.json
```

### 3. 测试执行

```bash
lingtong-cli workflow test run --workflow-id <id> --params '{"key":"value"}'
```

### 4. 发布

```bash
lingtong-cli workflow publish --workflow-id <id> --version "v1.0.0"
```

## 核心功能

- **节点编排**: 支持 12 种节点类型
- **DSL 验证**: 自动检查 DSL 格式和逻辑
- **模板管理**: 内置常用工作流模板
- **测试执行**: 支持参数化测试
- **版本管理**: 发布、回滚、版本历史
- **Open API**: 启用/禁用 API 访问

## 文档导航

| 文档 | 说明 |
|------|------|
| [SKILL.md](SKILL.md) | 主技能文档，包含快速路由和命令参考 |
| [node-details.md](references/node-details.md) | 12种节点详细配置 |
| [dsl-specification.md](references/dsl-specification.md) | DSL 完整规范 |
| [script-guide.md](references/script-guide.md) | 脚本编写指南 |
| [connector-invoke.md](references/connector-invoke.md) | 通用连接器调用 |
| [template-guide.md](references/template-guide.md) | 模板使用指南 |
| [testing-guide.md](references/testing-guide.md) | 测试指南 |
| [error-codes.md](references/error-codes.md) | 错误码说明 |

## 依赖

- `lingtong-cli`: 绫通命令行工具

## 相关技能

- `lingtong-connector`: 连接器管理
- `lingtong-scene`: 场景管理
- `lingtong-table`: 表格操作

## 维护说明

- 节点配置以 `references/node-details.md` 为准
- DSL 格式以 `references/dsl-specification.md` 为准
- 脚本语法限制：ES5.1（Nashorn 引擎）
