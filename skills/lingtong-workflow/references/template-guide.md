# 模板使用指南

绫通工作流提供内置模板，快速创建常见工作流模式。

## 可用模板

| 模板名 | 节点数 | 用途 | 适用场景 |
|--------|--------|------|----------|
| `simple` | 2 | 开始→结束 | 测试、学习 |
| `connector` | 3 | 开始→连接器→结束 | 简单 API 调用 |
| `order_sync` | 5 | 订单同步流程 | ERP 订单同步 |
| `approval` | 4 | 审批流程 | 条件审批 |
| `data_pipeline` | 5 | 数据管道流程 | 批量数据处理 |
| `api_wrapper` | 4 | API 包装流程 | 连接器接口暴露为 API |

## 命令

### 列出模板

```bash
lingtong-cli workflow template list
```

### 查看模板详情

```bash
lingtong-cli workflow template show --name order_sync
```

### 使用模板创建

```bash
# 使用模板生成 DSL 文件
lingtong-cli workflow template use --name approval --output approval.json

# 使用模板直接创建工作流
lingtong-cli workflow create --name "审批流程" --template approval
```

## 模板说明

### simple

最基础的开始→结束工作流，适合测试和学习。

```
w_start → w_end
```

### connector

简单的连接器调用工作流。

```
w_start → w_connector → w_end
```

### order_sync

订单同步流程，包含查询、转换、推送等步骤。

```
w_start → w_connector(查询) → w_script(转换) → w_connector(推送) → w_end
```

### approval

带条件判断的审批流程。

```
w_start → w_if(判断) → [true] w_connector(审批) → w_end
                     → [false] w_end
```

### data_pipeline

数据管道流程，支持批量数据同步。

```
w_start → w_modePipe → w_dataSplit → w_script → w_end
```

### api_wrapper

将连接器接口包装为独立 API。

```
w_start → w_connector → w_script(格式化) → w_end
```

## 自定义模板

如需自定义模板，可以：

1. 使用现有模板生成 DSL 文件
2. 修改 DSL 文件
3. 使用 `--dsl-file` 创建工作流

```bash
# 生成 DSL
lingtong-cli workflow template use --name connector --output base.json

# 编辑 DSL
vim base.json

# 创建
lingtong-cli workflow create --name "自定义" --dsl-file base.json
```
