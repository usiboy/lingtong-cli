---
name: lingtong-pivot-table
version: 1.0.0
description: "绫通表格数据透视表：透视表查询、配置管理、维度/度量配置、显示模式。当用户需要操作绫通表格的数据透视表、配置透视表维度度量、查询透视表数据时触发。关键词：pivot、透视表、数据透视表、维度、度量、聚合。"
related_skills:
  - lingtong-table
---

# lingtong-pivot-table 技能

## 概述

本技能指导你如何通过 `lingtong-cli` 管理绫通表格的数据透视表。数据透视表基于 ES 聚合 API 实现，支持多种维度类型和度量计算。

## 前置条件

1. 已配置平台 Host：
```bash
lingtong-cli config init --host https://your-lingtong-host.com
```

2. 已创建透视表视图（type=pivot）：
```bash
lingtong-cli table view save --schema-id <id> --name "透视表名称" --type pivot
```

## 核心命令

### 查询透视表数据

```bash
lingtong-cli table pivot query --table-id <id> --view-id <id>
```

**参数说明**：
- `--table-id`: 表格 ID（必填）
- `--view-id`: 视图 ID（必填）

**响应结构**：
```json
{
  "success": true,
  "result": {
    "data": {
      "displayMode": "normal",
      "columns": [
        {"key": "维度别名", "type": "dimension", "dataType": "text"},
        {"key": "度量别名", "type": "measure", "dataType": "number"}
      ],
      "rows": [
        {"维度别名": "值1", "度量别名": 100}
      ],
      "footerData": {"度量别名": {"type": "count", "value": 100, "text": "记录总数100"}}
    },
    "pagination": {"currentPage": 1, "pageSize": 50, "total": 100},
    "summary": {"summaryData": {"度量别名": 1000}}
  }
}
```

### 获取透视表配置

```bash
lingtong-cli table pivot config --table-id <id> --view-id <id>
```

### 保存透视表配置

```bash
lingtong-cli table pivot config-save --table-id <id> --view-id <id> \
  --business-id <id> --business-type <n> --config <json>
```

**参数说明**：
- `--table-id`: 表格 ID（必填）
- `--view-id`: 视图 ID（必填）
- `--business-id`: 业务 ID（必填，通常与 table-id 相同）
- `--business-type`: 业务类型（默认 2，2=场景）
- `--config`: 配置 JSON（必填）

## 配置结构详解

### 完整配置示例

```json
{
  "dimensions": [
    {
      "id": "dim1",
      "field": "data1568.1",
      "alias": "单据类型",
      "type": "term",
      "disabled": false,
      "config": {
        "sortCount": 10,
        "sortOrder": "desc",
        "groupOther": true
      }
    }
  ],
  "measures": [
    {
      "id": "m1",
      "type": "count",
      "disabled": false
    },
    {
      "id": "m2",
      "field": "data1568.17",
      "alias": "数量合计",
      "type": "sum",
      "disabled": false
    }
  ],
  "advanced": {
    "openDimensionSplit": false,
    "openStatistic": true,
    "displayMode": "normal"
  },
  "other": {
    "maxRowsPerPage": 50,
    "showSummary": true,
    "otherExtractConfigVOList": [
      {"id": "m1", "type": "measure", "summaryType": "count", "disabled": false},
      {"id": "m2", "type": "measure", "summaryType": "sum", "disabled": false}
    ]
  }
}
```

### 维度配置 (dimensions)

维度用于分组数据，支持以下类型：

| 类型 | 说明 | 适用字段 | config 配置 |
|------|------|----------|-------------|
| `term` | 分组（按字段值分组） | text/keyword | sortCount, sortOrder, groupOther |
| `date_histogram` | 日期分组 | date | interval, extendedBounds |
| `range` | 数值范围分组 | number | rangeValueList |
| `date_range` | 日期范围分组 | date | dateList |

**term 维度配置**：
```json
{
  "id": "dim1",
  "field": "data1568.1",
  "alias": "单据类型",
  "type": "term",
  "disabled": false,
  "config": {
    "sortCount": 10,
    "sortOrder": "desc",
    "groupOther": true
  }
}
```

- `sortCount`: 返回的桶数量（默认 100）
- `sortOrder`: 排序方向（asc/desc）
- `groupOther`: 是否将超出 sortCount 的桶归类为"其他"

**date_histogram 维度配置**：
```json
{
  "id": "dim1",
  "field": "data1568.22",
  "alias": "操作时间",
  "type": "date_histogram",
  "disabled": false,
  "config": {
    "interval": "1M",
    "extendedBounds": false
  }
}
```

- `interval`: 时间间隔（1d=天，1w=周，1M=月，1y=年）
- `extendedBounds`: 是否自动填充空桶

**range 维度配置**：
```json
{
  "id": "dim1",
  "field": "data1568.17",
  "alias": "数量范围",
  "type": "range",
  "disabled": false,
  "config": {
    "rangeValueList": [
      {"min": -999999, "max": 0},
      {"min": 0, "max": 10},
      {"min": 10, "max": 100},
      {"min": 100, "max": 999999}
    ]
  }
}
```

**date_range 维度配置**：
```json
{
  "id": "dim1",
  "field": "data1568.22",
  "alias": "时间范围",
  "type": "date_range",
  "disabled": false,
  "config": {
    "dateList": [
      {"start": "2024-01-01", "end": "2024-06-30"},
      {"start": "2024-07-01", "end": "2024-12-31"}
    ]
  }
}
```

### 度量配置 (measures)

度量用于聚合计算，支持以下类型：

| 类型 | 说明 | field 要求 |
|------|------|-----------|
| `count` | 计数（统计文档数） | 可选 |
| `valueCount` | 值计数（非空值数量） | 必填 |
| `sum` | 求和 | 必填（数值字段） |
| `avg` | 平均值 | 必填（数值字段） |
| `min` | 最小值 | 必填（数值/日期字段） |
| `max` | 最大值 | 必填（数值/日期字段） |
| `unique` | 去重计数 | 必填 |

**度量配置示例**：
```json
{
  "id": "m1",
  "type": "count",
  "disabled": false
}
```

```json
{
  "id": "m2",
  "field": "data1568.17",
  "alias": "数量合计",
  "type": "sum",
  "disabled": false
}
```

### 高级配置 (advanced)

| 字段 | 说明 | 可选值 |
|------|------|--------|
| `displayMode` | 显示模式 | normal（普通表格）/ split（拆分表）/ cross（交叉表） |
| `openDimensionSplit` | 开启维度拆分 | true/false |
| `openStatistic` | 开启维度度量统计 | true/false |

**显示模式说明**：
- `normal`: 普通表格模式，维度作为行，度量作为列
- `split`: 拆分表模式，按指定维度拆分成多个表格
- `cross`: 交叉表模式，行维度和列维度交叉展示

### 其他配置 (other)

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `maxRowsPerPage` | 每页最大行数 | 50 |
| `showSummary` | 是否显示汇总行 | true |
| `otherExtractConfigVOList` | 汇总计算配置 | - |

**汇总计算配置**：
```json
{
  "id": "m1",
  "type": "measure",
  "summaryType": "count",
  "disabled": false
}
```

- `summaryType`: 计算类型（count/sum/avg/max/min/unique）

## 使用示例

### 示例 1: 基础透视表（单维度 + 计数）

```bash
lingtong-cli table pivot config-save --table-id 1568 --view-id 694 \
  --business-id 1568 --business-type 2 \
  --config '{"dimensions":[{"id":"dim1","field":"data1568.1","alias":"单据类型","type":"term","disabled":false}],"measures":[{"id":"m1","type":"count","disabled":false}],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"},"other":{"maxRowsPerPage":50,"showSummary":true,"otherExtractConfigVOList":[{"id":"m1","type":"measure","summaryType":"count","disabled":false}]}}'
```

### 示例 2: 多维度 + 多度量

```bash
lingtong-cli table pivot config-save --table-id 1568 --view-id 694 \
  --business-id 1568 --business-type 2 \
  --config '{"dimensions":[{"id":"dim1","field":"data1568.1","alias":"单据类型","type":"term","disabled":false},{"id":"dim2","field":"data1568.18","alias":"仓库","type":"term","disabled":false}],"measures":[{"id":"m1","type":"count","disabled":false},{"id":"m2","field":"data1568.17","alias":"数量合计","type":"sum","disabled":false}],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"},"other":{"maxRowsPerPage":50,"showSummary":true,"otherExtractConfigVOList":[{"id":"m1","type":"measure","summaryType":"count","disabled":false},{"id":"m2","type":"measure","summaryType":"sum","disabled":false}]}}'
```

### 示例 3: 日期分组

```bash
lingtong-cli table pivot config-save --table-id 1568 --view-id 694 \
  --business-id 1568 --business-type 2 \
  --config '{"dimensions":[{"id":"dim1","field":"data1568.22","alias":"操作时间","type":"date_histogram","disabled":false,"config":{"interval":"1M"}}],"measures":[{"id":"m1","type":"count","disabled":false}],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"},"other":{"maxRowsPerPage":50,"showSummary":true,"otherExtractConfigVOList":[{"id":"m1","type":"measure","summaryType":"count","disabled":false}]}}'
```

### 示例 4: 数值范围分组

```bash
lingtong-cli table pivot config-save --table-id 1568 --view-id 694 \
  --business-id 1568 --business-type 2 \
  --config '{"dimensions":[{"id":"dim1","field":"data1568.17","alias":"数量范围","type":"range","disabled":false,"config":{"rangeValueList":[{"min":-999999,"max":0},{"min":0,"max":10},{"min":10,"max":100},{"min":100,"max":999999}]}}],"measures":[{"id":"m1","type":"count","disabled":false}],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"},"other":{"maxRowsPerPage":50,"showSummary":true,"otherExtractConfigVOList":[{"id":"m1","type":"measure","summaryType":"count","disabled":false}]}}'
```

### 示例 5: 带排序和"其他"分组

```bash
lingtong-cli table pivot config-save --table-id 1568 --view-id 694 \
  --business-id 1568 --business-type 2 \
  --config '{"dimensions":[{"id":"dim1","field":"data1568.1","alias":"单据类型","type":"term","disabled":false,"config":{"sortCount":5,"sortOrder":"desc","groupOther":true}}],"measures":[{"id":"m1","type":"count","disabled":false}],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"},"other":{"maxRowsPerPage":50,"showSummary":true,"otherExtractConfigVOList":[{"id":"m1","type":"measure","summaryType":"count","disabled":false}]}}'
```

## 字段 ID 获取

透视表配置中的 `field` 字段需要使用 ES 字段名格式：`data{tableId}.{fieldId}`

例如：
- 表格 ID 1568，字段 ID 1 → `data1568.1`
- 表格 ID 1568，字段 ID 17 → `data1568.17`

获取字段 ID 的方法：
```bash
lingtong-cli table schema query --basic-data-id 1568 | jq '.result.columnsSchema[] | {id, key, title, type}'
```

## 注意事项

1. **维度类型选择**：
   - 文本字段使用 `term` 类型
   - 日期字段使用 `date_histogram` 或 `date_range` 类型
   - 数值字段使用 `range` 类型

2. **度量类型选择**：
   - `count` 不需要指定 field
   - `sum/avg/min/max` 必须指定数值字段
   - `unique` 可以指定任意字段

3. **显示模式**：
   - `normal` 模式最常用，适合大多数场景
   - `split` 和 `cross` 模式需要额外配置

4. **配置保存后**：
   - 配置会自动验证字段是否存在
   - 验证失败会返回错误信息
   - 验证通过后才能查询

## 相关技能

- [lingtong-table](../lingtong-table/SKILL.md) - 绫通表格基础操作
