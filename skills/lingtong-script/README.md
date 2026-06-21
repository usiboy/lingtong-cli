# lingtong-script 技能

绫通脚本编写技能，覆盖工作流脚本节点、场景字段函数、连接器字段函数三种执行场景。

## 目录结构

```
lingtong-script/
├── SKILL.md                        # 主技能文档
├── README.md                       # 本文件
└── references/                     # 参考文档
    ├── inf-invoker-guide.md        # InfInvoker 表格函数完整指南
    ├── script-basics.md            # ES5.1 语法 + context API + 内置对象
    └── examples/                   # 示例文档
        └── table-operations.md     # 表格操作示例（基于出入库记录表）
```

## 核心能力

- **InfInvoker** — 绫通表格 CRUD 操作 + ES 8.15 聚合分析
- **AppInvoker** — 连接器接口调用
- **context API** — 脚本变量读写
- **ES5.1 语法规范** — Nashorn 引擎限制说明

## 适用场景

| 场景 | 说明 |
|------|------|
| 工作流脚本节点 | 工作流中的 w_script 节点 |
| 场景字段函数 | 场景配置中的字段映射函数 |
| 连接器字段函数 | 连接器节点中的参数预处理脚本 |

## 维护说明

- 基于后端源码（`LtInfInvoker.java`、`DefaultLtInfInvoker.java`）独立编写
- 不直接复制 lingtong-skill 文档，确保 lingtong-cli 可独立迭代
- 更新时同步检查后端接口变更
