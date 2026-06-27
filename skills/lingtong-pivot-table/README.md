# lingtong-pivot-table 技能

`lingtong-pivot-table` 专注绫通表格的数据透视表能力，包括透视表视图创建、配置读取/保存、维度/度量配置和聚合查询。

## 何时使用

- 用户明确提到透视表、数据透视表、pivot、维度、度量、聚合
- 需要把表格数据按字段分组、计数、求和、平均、去重
- 需要保存或调整透视表配置
- 需要查询透视表展示数据

## 前置条件

透视表依赖表格和视图：

```bash
# 先确认表格 schema
lingtong-cli table schema query --basic-data-id 1568

# 创建 pivot 类型视图
lingtong-cli table view save --schema-id 2367 --name 透视表 --type pivot
```

## 快速开始

```bash
# 获取当前配置
lingtong-cli table pivot config --table-id 1568 --view-id 621

# 保存配置
lingtong-cli table pivot config-save --table-id 1568 --view-id 621 \
  --business-id 1568 --business-type 2 \
  --config '{"dimensions":[{"id":"dim1","field":"data1568.1","alias":"单据类型","type":"term","disabled":false}],"measures":[{"id":"m1","type":"count","disabled":false}],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"}}'

# 查询结果
lingtong-cli table pivot query --table-id 1568 --view-id 621
```

## 命令清单

| 目标 | 命令 |
|------|------|
| 查询透视表数据 | `lingtong-cli table pivot query --table-id <id> --view-id <id>` |
| 获取透视表配置 | `lingtong-cli table pivot config --table-id <id> --view-id <id>` |
| 保存透视表配置 | `lingtong-cli table pivot config-save --table-id <id> --view-id <id> --business-id <id> --business-type 2 --config '<json>'` |

## 配置核心结构

```json
{
  "dimensions": [
    {"id": "dim1", "field": "data1568.1", "alias": "单据类型", "type": "term", "disabled": false}
  ],
  "measures": [
    {"id": "m1", "type": "count", "disabled": false}
  ],
  "advanced": {
    "openDimensionSplit": false,
    "openStatistic": true,
    "displayMode": "normal"
  }
}
```

字段名通常使用 `data<tableId>.<fieldId>`，例如 `data1568.1`。

## 维度与度量

| 类型 | 用途 |
|------|------|
| `term` | 按文本/枚举字段分组 |
| `date_histogram` | 按日期周期分组 |
| `range` | 按数值范围分组 |
| `date_range` | 按日期范围分组 |
| `count` | 记录数 |
| `valueCount` | 非空值数量 |
| `sum` / `avg` / `min` / `max` | 数值或日期聚合 |
| `unique` | 去重计数 |

## 与 lingtong-table 的关系

- 表格、Schema、普通视图、记录查询使用 `lingtong-table`。
- 透视表配置和聚合查询使用本技能。
- 创建透视表视图依赖 `table view save --type pivot`。

## 相关文档

- [SKILL.md](SKILL.md) - 详细配置说明
- `lingtong-table` - 表格与视图基础能力
