---
name: lingtong-factory
version: 1.0.0
description: "绫通连接器工厂 (factory)：自定义连接器定义、其 HTTP 接口(方法)与脚本的增删改查。当用户需要管理连接器工厂、HTTP 接口、脚本,或测试运行接口时触发。关键词：factory、连接器工厂、http 接口、script、脚本。"
metadata:
  requires:
    bins: ["lingtong-cli"]
  cliHelp: "lingtong-cli factory --help"
---

# lingtong-factory

连接器工厂(`/factory/*`)的手写命令组。一个工厂定义一个自定义连接器,包含若干 **HTTP 接口(方法)** 和 **脚本**。

## 工厂本体

```bash
lingtong-cli factory list                       # 列出工厂(分页)
lingtong-cli factory list --page 2 --page-size 50
lingtong-cli factory get --id 12                # 工厂详情
lingtong-cli factory save --data '{"name":"my-connector","connector":"mycon"}'  # 新建
lingtong-cli factory save --data '{"id":12,"name":"renamed"}'                    # 更新(带 id)
lingtong-cli factory delete --id 12 --yes       # 删除(破坏性,需 --yes)
```

`save` 通过 `--data` 是否含 `id` 区分新建/更新。

## HTTP 接口(方法)

```bash
lingtong-cli factory http list --factory-id 12  # 列出工厂的 HTTP 接口
lingtong-cli factory http get --id 88           # 接口详情
lingtong-cli factory http run --data '{"id":88,"params":{"page":1}}'  # 运行/测试接口
lingtong-cli factory http delete --id 88 --yes  # 删除接口(需 --yes)
```

## 脚本

```bash
lingtong-cli factory script list --factory-id 12  # 列出脚本
lingtong-cli factory script get --id 55           # 脚本详情
lingtong-cli factory script delete --id 55 --yes  # 删除脚本(需 --yes)
```

## 规则

- **破坏性操作**(`delete`)默认拒绝执行,必须显式加 `--yes`,否则以**退出码 10**(需要确认)失败。
- 所有命令共享统一输出契约:`--envelope`、类型化退出码、`--jq`、`--omit-null`(见 [[lingtong-shared]])。
- 缺必填参数 → 退出码 **2**;`--data` JSON 非法 → 退出码 **2**。
- 未包装的长尾工厂接口(模型/参数/标签等)可用自动生成层 [[lingtong-service]]:`lingtong-cli service factory ...`。

## 典型流程

```bash
# 1. 找到工厂
lingtong-cli factory list --jq '.result.list[] | {id, name}'

# 2. 查看其 HTTP 接口
lingtong-cli factory http list --factory-id 12 --jq '.result.list[] | {id, name}'

# 3. 测试运行某接口
lingtong-cli factory http run --data '{"id":88,"params":{}}' --envelope --jq '.ok'
```
